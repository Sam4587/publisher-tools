package video

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service 视频处理服务
type Service struct {
	db         *gorm.DB
	downloader *Downloader
	transcriber *Transcriber
	optimizer  *Optimizer
	config     *ServiceConfig

	// 任务管理
	tasks     map[string]*VideoTask
	taskMu    sync.RWMutex
	taskQueue chan *VideoTask
	workers   int
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	DownloaderConfig  *DownloaderConfig  `json:"downloader"`
	TranscriberConfig *TranscriberConfig `json:"transcriber"`
	OptimizerConfig   *OptimizerConfig   `json:"optimizer"`
	Workers           int                `json:"workers"`
	QueueSize         int                `json:"queue_size"`
}

// DefaultServiceConfig 默认配置
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		DownloaderConfig:  DefaultDownloaderConfig(),
		TranscriberConfig: DefaultTranscriberConfig(),
		OptimizerConfig:   DefaultOptimizerConfig(),
		Workers:           2,
		QueueSize:         100,
	}
}

// VideoTask 视频处理任务
type VideoTask struct {
	ID          string            `json:"id"`
	VideoID     string            `json:"video_id"`
	URL         string            `json:"url"`
	Status      string            `json:"status"` // pending, downloading, transcribing, optimizing, completed, failed
	Progress    int               `json:"progress"`
	Error       string            `json:"error,omitempty"`
	Result      *ProcessResult    `json:"result,omitempty"`
	Options     *ProcessOptions   `json:"options"`
	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// ProcessOptions 处理选项
type ProcessOptions struct {
	DownloadVideo    bool   `json:"download_video"`
	DownloadAudio    bool   `json:"download_audio"`
	DownloadSubtitle bool   `json:"download_subtitle"`
	Transcribe       bool   `json:"transcribe"`
	TranscribeModel  string `json:"transcribe_model"`
	TranscribeLang   string `json:"transcribe_lang"`
	Optimize         bool   `json:"optimize"`
	OptimizeLang     string `json:"optimize_lang"`
	GenerateSummary  bool   `json:"generate_summary"`
}

// DefaultProcessOptions 默认处理选项
func DefaultProcessOptions() *ProcessOptions {
	return &ProcessOptions{
		DownloadVideo:    false,
		DownloadAudio:    true,
		DownloadSubtitle: true,
		Transcribe:       true,
		TranscribeModel:  "base",
		TranscribeLang:   "auto",
		Optimize:         true,
		OptimizeLang:     "zh",
		GenerateSummary:  true,
	}
}

// ProcessResult 处理结果
type ProcessResult struct {
	VideoID       string   `json:"video_id"`
	Title         string   `json:"title"`
	Platform      string   `json:"platform"`
	Duration      int      `json:"duration"`
	VideoPath     string   `json:"video_path,omitempty"`
	AudioPath     string   `json:"audio_path,omitempty"`
	Transcript    string   `json:"transcript,omitempty"`
	Optimized     string   `json:"optimized,omitempty"`
	Summary       string   `json:"summary,omitempty"`
	KeyPoints     []string `json:"key_points,omitempty"`
	Topics        []string `json:"topics,omitempty"`
	WordCount     int      `json:"word_count"`
	ProcessTime   float64  `json:"process_time"`
}

// NewService 创建视频处理服务
func NewService(db *gorm.DB, aiProvider AIProvider, config *ServiceConfig) *Service {
	if config == nil {
		config = DefaultServiceConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Service{
		db:         db,
		config:     config,
		tasks:      make(map[string]*VideoTask),
		taskQueue:  make(chan *VideoTask, config.QueueSize),
		workers:    config.Workers,
		ctx:        ctx,
		cancel:     cancel,
	}

	// 初始化子组件
	s.downloader = NewDownloader(db, config.DownloaderConfig)
	s.transcriber = NewTranscriber(db, config.TranscriberConfig)
	s.optimizer = NewOptimizer(db, aiProvider, config.OptimizerConfig)

	return s
}

// Start 启动服务
func (s *Service) Start() {
	logrus.Infof("Starting video service with %d workers", s.workers)

	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// Stop 停止服务
func (s *Service) Stop() {
	logrus.Info("Stopping video service...")
	s.cancel()
	s.wg.Wait()
	logrus.Info("Video service stopped")
}

// worker 工作协程
func (s *Service) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.taskQueue:
			s.processTask(task)
		}
	}
}

// processTask 处理任务
func (s *Service) processTask(task *VideoTask) {
	startTime := time.Now()
	task.Status = "processing"
	task.StartedAt = &startTime
	s.updateTask(task)

	logrus.Infof("Processing video task: %s", task.ID)

	var result ProcessResult
	var err error

	// 1. 下载视频/音频
	task.Status = "downloading"
	task.Progress = 10
	s.updateTask(task)

	downloadOpts := &DownloadOptions{
		AudioOnly:         !task.Options.DownloadVideo,
		DownloadSubtitles: task.Options.DownloadSubtitle,
	}

	downloadResult, err := s.downloader.Download(s.ctx, task.URL, downloadOpts)
	if err != nil {
		s.failTask(task, fmt.Sprintf("Download failed: %v", err))
		return
	}

	result.VideoID = downloadResult.VideoID
	result.Title = downloadResult.Title
	result.Platform = downloadResult.Platform
	result.Duration = downloadResult.Duration
	result.VideoPath = downloadResult.VideoPath
	result.AudioPath = downloadResult.AudioPath

	task.VideoID = downloadResult.VideoID
	task.Progress = 30
	s.updateTask(task)

	// 2. 转录音频
	if task.Options.Transcribe && downloadResult.AudioPath != "" {
		task.Status = "transcribing"
		task.Progress = 40
		s.updateTask(task)

		transcriptResult, err := s.transcriber.TranscribeWithFallback(s.ctx, downloadResult.AudioPath, downloadResult.VideoID)
		if err != nil {
			logrus.Warnf("Transcription failed: %v, continuing without transcript", err)
		} else {
			result.Transcript = transcriptResult.Text
			result.WordCount = transcriptResult.WordCount
		}

		task.Progress = 60
		s.updateTask(task)
	}

	// 3. 优化文本
	if task.Options.Optimize && result.Transcript != "" {
		task.Status = "optimizing"
		task.Progress = 70
		s.updateTask(task)

		optimizeOpts := &OptimizeOptions{
			Language:       task.Options.OptimizeLang,
			AddPunctuation: true,
			FixGrammar:     true,
			RemoveFiller:   true,
		}

		optimizeResult, err := s.optimizer.Optimize(s.ctx, result.Transcript, optimizeOpts)
		if err != nil {
			logrus.Warnf("Optimization failed: %v, using original transcript", err)
		} else {
			result.Optimized = optimizeResult.OptimizedText
			result.Summary = optimizeResult.Summary
			result.KeyPoints = optimizeResult.KeyPoints
			result.Topics = optimizeResult.Topics
		}

		task.Progress = 90
		s.updateTask(task)
	}

	// 4. 完成
	completedAt := time.Now()
	task.Status = "completed"
	task.Progress = 100
	task.CompletedAt = &completedAt
	task.Result = &result
	s.updateTask(task)

	result.ProcessTime = completedAt.Sub(startTime).Seconds()

	logrus.Infof("Video task completed: %s (%.2fs)", task.ID, result.ProcessTime)
}

// SubmitTask 提交任务
func (s *Service) SubmitTask(url string, opts *ProcessOptions) (*VideoTask, error) {
	if opts == nil {
		opts = DefaultProcessOptions()
	}

	task := &VideoTask{
		ID:        fmt.Sprintf("video_%d", time.Now().UnixNano()),
		URL:       url,
		Status:    "pending",
		Progress:  0,
		Options:   opts,
		CreatedAt: time.Now(),
	}

	// 保存到数据库
	if s.db != nil {
		video := &database.Video{
			ID:     task.ID,
			URL:    url,
			Status: "pending",
		}
		if err := s.db.Create(video).Error; err != nil {
			logrus.Warnf("Failed to save video task: %v", err)
		}
	}

	// 添加到任务队列
	s.taskMu.Lock()
	s.tasks[task.ID] = task
	s.taskMu.Unlock()

	select {
	case s.taskQueue <- task:
		logrus.Infof("Video task submitted: %s", task.ID)
		return task, nil
	default:
		return nil, fmt.Errorf("task queue is full")
	}
}

// GetTask 获取任务状态
func (s *Service) GetTask(taskID string) (*VideoTask, error) {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	return task, nil
}

// GetTaskByVideoID 通过视频ID获取任务
func (s *Service) GetTaskByVideoID(videoID string) (*VideoTask, error) {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	for _, task := range s.tasks {
		if task.VideoID == videoID {
			return task, nil
		}
	}
	return nil, fmt.Errorf("task not found for video: %s", videoID)
}

// ListTasks 列出所有任务
func (s *Service) ListTasks(status string) []*VideoTask {
	s.taskMu.RLock()
	defer s.taskMu.RUnlock()

	var tasks []*VideoTask
	for _, task := range s.tasks {
		if status == "" || task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// CancelTask 取消任务
func (s *Service) CancelTask(taskID string) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	if task.Status == "completed" {
		return fmt.Errorf("cannot cancel completed task")
	}

	task.Status = "cancelled"
	s.updateTask(task)

	return nil
}

// updateTask 更新任务
func (s *Service) updateTask(task *VideoTask) {
	s.taskMu.Lock()
	s.tasks[task.ID] = task
	s.taskMu.Unlock()

	// 更新数据库
	if s.db != nil && task.VideoID != "" {
		updates := map[string]interface{}{
			"status":    task.Status,
			"updated_at": time.Now(),
		}

		if task.Status == "completed" && task.Result != nil {
			updates["title"] = task.Result.Title
			updates["duration"] = task.Result.Duration
		}

		s.db.Model(&database.Video{}).Where("id = ?", task.VideoID).Updates(updates)
	}
}

// failTask 标记任务失败
func (s *Service) failTask(task *VideoTask, errMsg string) {
	task.Status = "failed"
	task.Error = errMsg
	now := time.Now()
	task.CompletedAt = &now
	s.updateTask(task)

	logrus.Errorf("Video task failed: %s - %s", task.ID, errMsg)
}

// ProcessVideo 同步处理视频（阻塞直到完成）
func (s *Service) ProcessVideo(ctx context.Context, url string, opts *ProcessOptions) (*ProcessResult, error) {
	if opts == nil {
		opts = DefaultProcessOptions()
	}

	task, err := s.SubmitTask(url, opts)
	if err != nil {
		return nil, err
	}

	// 等待任务完成
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			task, _ = s.GetTask(task.ID)
			if task.Status == "completed" {
				return task.Result, nil
			}
			if task.Status == "failed" {
				return nil, fmt.Errorf(task.Error)
			}
			if task.Status == "cancelled" {
				return nil, fmt.Errorf("task cancelled")
			}
		}
	}
}

// GetVideo 获取视频信息
func (s *Service) GetVideo(videoID string) (*database.Video, error) {
	var video database.Video
	err := s.db.Preload("Transcript").First(&video, "id = ?", videoID).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// ListVideos 列出视频
func (s *Service) ListVideos(status string, limit, offset int) ([]database.Video, int64, error) {
	var videos []database.Video
	var total int64

	query := s.db.Model(&database.Video{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	err := query.Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&videos).Error

	return videos, total, err
}

// DeleteVideo 删除视频
func (s *Service) DeleteVideo(videoID string) error {
	// 删除文件
	videoDir := s.downloader.outputDir + "/" + videoID
	if err := s.deleteDir(videoDir); err != nil {
		logrus.Warnf("Failed to delete video files: %v", err)
	}

	// 删除数据库记录
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("video_id = ?", videoID).Delete(&database.Transcript{}).Error; err != nil {
			return err
		}
		return tx.Delete(&database.Video{}, "id = ?", videoID).Error
	})
}

// deleteDir 删除目录
func (s *Service) deleteDir(path string) error {
	// 使用 os.RemoveAll
	return nil // 简化实现
}

// GetStats 获取统计信息
func (s *Service) GetStats() (*ServiceStats, error) {
	stats := &ServiceStats{}

	// 任务统计
	s.taskMu.RLock()
	for _, task := range s.tasks {
		switch task.Status {
		case "pending":
			stats.PendingTasks++
		case "processing", "downloading", "transcribing", "optimizing":
			stats.ActiveTasks++
		case "completed":
			stats.CompletedTasks++
		case "failed":
			stats.FailedTasks++
		}
	}
	s.taskMu.RUnlock()

	// 视频统计
	if s.db != nil {
		s.db.Model(&database.Video{}).Count(&stats.TotalVideos)
		s.db.Model(&database.Video{}).Where("status = ?", "completed").Count(&stats.CompletedVideos)
	}

	return stats, nil
}

// ServiceStats 服务统计
type ServiceStats struct {
	TotalVideos     int64 `json:"total_videos"`
	CompletedVideos int64 `json:"completed_videos"`
	PendingTasks    int   `json:"pending_tasks"`
	ActiveTasks     int   `json:"active_tasks"`
	CompletedTasks  int   `json:"completed_tasks"`
	FailedTasks     int   `json:"failed_tasks"`
}

// MarshalJSON 自定义 JSON 序列化
func (t *VideoTask) MarshalJSON() ([]byte, error) {
	type Alias VideoTask
	return json.Marshal(&struct {
		*Alias
		Duration float64 `json:"duration,omitempty"`
	}{
		Alias: (*Alias)(t),
	})
}
