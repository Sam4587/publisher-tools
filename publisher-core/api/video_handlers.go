package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"publisher-core/video/processor"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// VideoServiceAPI 视频服务接口
type VideoServiceAPI interface {
	GetInfo(ctx context.Context, filePath string) (*processor.VideoInfo, error)
	Slice(ctx context.Context, filePath string, config *processor.SliceConfig) (*processor.SliceResult, error)
	GenerateThumbnails(ctx context.Context, filePath string, config *processor.ThumbnailConfig) (*processor.ThumbnailResult, error)
	Convert(ctx context.Context, filePath string, config *processor.ConvertConfig) (*processor.ConvertResult, error)
	BatchProcess(ctx context.Context, files []string, jobType processor.BatchJobType, config interface{}, callback processor.ProgressCallback) (*processor.BatchJob, error)
	IsAvailable() bool
	GetSupportedFormats() []string
}

// VideoHandlers 视频处理器
type VideoHandlers struct {
	service   VideoServiceAPI
	uploadDir string
}

// NewVideoHandlers 创建视频处理器
func NewVideoHandlers(service VideoServiceAPI, uploadDir string) *VideoHandlers {
	if uploadDir == "" {
		uploadDir = "./uploads/video"
	}
	os.MkdirAll(uploadDir, 0755)
	return &VideoHandlers{
		service:   service,
		uploadDir: uploadDir,
	}
}

// RegisterRoutes 注册路由
func (h *VideoHandlers) RegisterRoutes(router *mux.Router) {
	videoRouter := router.PathPrefix("/api/v1/video").Subrouter()

	// 视频信息
	videoRouter.HandleFunc("/info", h.getInfo).Methods("POST")
	videoRouter.HandleFunc("/info/{path:.*}", h.getInfoByPath).Methods("GET")

	// 视频切片
	videoRouter.HandleFunc("/slice", h.slice).Methods("POST")
	videoRouter.HandleFunc("/slice/time", h.sliceByTime).Methods("POST")
	videoRouter.HandleFunc("/slice/segments", h.sliceBySegments).Methods("POST")

	// 缩略图
	videoRouter.HandleFunc("/thumbnails", h.generateThumbnails).Methods("POST")
	videoRouter.HandleFunc("/thumbnails/sprite", h.generateSpriteSheet).Methods("POST")

	// 格式转换
	videoRouter.HandleFunc("/convert", h.convert).Methods("POST")

	// 批量处理
	videoRouter.HandleFunc("/batch", h.batchProcess).Methods("POST")
	videoRouter.HandleFunc("/batch/{id}", h.getBatchJob).Methods("GET")

	// 上传
	videoRouter.HandleFunc("/upload", h.uploadVideo).Methods("POST")

	// 工具接口
	videoRouter.HandleFunc("/formats", h.getSupportedFormats).Methods("GET")
	videoRouter.HandleFunc("/presets/resolutions", h.getResolutionPresets).Methods("GET")
	videoRouter.HandleFunc("/health", h.healthCheck).Methods("GET")
}

// getInfo 获取视频信息
func (h *VideoHandlers) getInfo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		jsonError(w, "MISSING_FILE_PATH", "file_path is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	info, err := h.service.GetInfo(ctx, req.FilePath)
	if err != nil {
		jsonError(w, "GET_INFO_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, info)
}

// getInfoByPath 通过URL路径获取视频信息
func (h *VideoHandlers) getInfoByPath(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	info, err := h.service.GetInfo(ctx, filePath)
	if err != nil {
		jsonError(w, "GET_INFO_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, info)
}

// slice 切片视频
func (h *VideoHandlers) slice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath        string  `json:"file_path"`
		OutputDir       string  `json:"output_dir"`
		OutputFormat    string  `json:"output_format"`
		OutputPrefix    string  `json:"output_prefix"`
		SegmentCount    int     `json:"segment_count"`
		SegmentDuration float64 `json:"segment_duration"`
		Duration        float64 `json:"duration"`
		CRF             int     `json:"crf"`
		Preset          string  `json:"preset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		jsonError(w, "MISSING_FILE_PATH", "file_path is required", http.StatusBadRequest)
		return
	}

	config := processor.DefaultSliceConfig()
	if req.OutputDir != "" {
		config.OutputDir = req.OutputDir
	}
	if req.OutputFormat != "" {
		config.OutputFormat = req.OutputFormat
	}
	if req.OutputPrefix != "" {
		config.OutputPrefix = req.OutputPrefix
	}
	config.SegmentCount = req.SegmentCount
	config.SegmentDuration = req.SegmentDuration
	config.Duration = req.Duration
	if req.CRF > 0 {
		config.CRF = req.CRF
	}
	if req.Preset != "" {
		config.Preset = req.Preset
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	result, err := h.service.Slice(ctx, req.FilePath, config)
	if err != nil {
		jsonError(w, "SLICE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// sliceByTime 按时间切片
func (h *VideoHandlers) sliceByTime(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath  string  `json:"file_path"`
		StartTime float64 `json:"start_time"`
		EndTime   float64 `json:"end_time"`
		OutputDir string  `json:"output_dir"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		jsonError(w, "MISSING_FILE_PATH", "file_path is required", http.StatusBadRequest)
		return
	}

	config := processor.DefaultSliceConfig()
	if req.OutputDir != "" {
		config.OutputDir = req.OutputDir
	}
	config.StartTime = req.StartTime
	config.EndTime = req.EndTime

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	result, err := h.service.Slice(ctx, req.FilePath, config)
	if err != nil {
		jsonError(w, "SLICE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// sliceBySegments 按分段数量切片
func (h *VideoHandlers) sliceBySegments(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath     string `json:"file_path"`
		SegmentCount int    `json:"segment_count"`
		OutputDir    string `json:"output_dir"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		jsonError(w, "MISSING_FILE_PATH", "file_path is required", http.StatusBadRequest)
		return
	}

	if req.SegmentCount <= 0 {
		jsonError(w, "INVALID_SEGMENT_COUNT", "segment_count must be greater than 0", http.StatusBadRequest)
		return
	}

	config := processor.DefaultSliceConfig()
	if req.OutputDir != "" {
		config.OutputDir = req.OutputDir
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	result, err := h.service.Slice(ctx, req.FilePath, config)
	if err != nil {
		jsonError(w, "SLICE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// generateThumbnails 生成缩略图
func (h *VideoHandlers) generateThumbnails(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath    string    `json:"file_path"`
		OutputDir   string    `json:"output_dir"`
		Format      string    `json:"format"`
		Width       int       `json:"width"`
		Height      int       `json:"height"`
		Quality     int       `json:"quality"`
		Timestamps  []float64 `json:"timestamps"`
		Interval    float64   `json:"interval"`
		Count       int       `json:"count"`
		SpriteSheet bool      `json:"sprite_sheet"`
		Columns     int       `json:"columns"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		jsonError(w, "MISSING_FILE_PATH", "file_path is required", http.StatusBadRequest)
		return
	}

	config := processor.DefaultThumbnailConfig()
	if req.OutputDir != "" {
		config.OutputDir = req.OutputDir
	}
	if req.Format != "" {
		config.OutputFormat = req.Format
	}
	config.Width = req.Width
	config.Height = req.Height
	if req.Quality > 0 {
		config.Quality = req.Quality
	}
	config.Timestamps = req.Timestamps
	config.Interval = req.Interval
	if req.Count > 0 {
		config.Count = req.Count
	}
	config.SpriteSheet = req.SpriteSheet
	config.Columns = req.Columns

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	result, err := h.service.GenerateThumbnails(ctx, req.FilePath, config)
	if err != nil {
		jsonError(w, "GENERATE_THUMBNAILS_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// generateSpriteSheet 生成雪碧图
func (h *VideoHandlers) generateSpriteSheet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath  string `json:"file_path"`
		OutputDir string `json:"output_dir"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Count     int    `json:"count"`
		Columns   int    `json:"columns"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	config := processor.DefaultThumbnailConfig()
	config.SpriteSheet = true
	if req.OutputDir != "" {
		config.OutputDir = req.OutputDir
	}
	config.Width = req.Width
	config.Height = req.Height
	if req.Count > 0 {
		config.Count = req.Count
	}
	if req.Columns > 0 {
		config.Columns = req.Columns
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	result, err := h.service.GenerateThumbnails(ctx, req.FilePath, config)
	if err != nil {
		jsonError(w, "GENERATE_SPRITE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result.SpriteSheet)
}

// convert 转换视频格式
func (h *VideoHandlers) convert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath    string `json:"file_path"`
		OutputPath  string `json:"output_path"`
		Format      string `json:"format"`
		VideoCodec  string `json:"video_codec"`
		AudioCodec  string `json:"audio_codec"`
		Resolution  string `json:"resolution"`
		VideoBitrate int   `json:"video_bitrate"`
		AudioBitrate int   `json:"audio_bitrate"`
		FPS         float64 `json:"fps"`
		CRF         int    `json:"crf"`
		Preset      string `json:"preset"`
		CopyAudio   bool   `json:"copy_audio"`
		CopyVideo   bool   `json:"copy_video"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.FilePath == "" {
		jsonError(w, "MISSING_FILE_PATH", "file_path is required", http.StatusBadRequest)
		return
	}

	config := processor.DefaultConvertConfig()
	config.OutputPath = req.OutputPath
	if req.Format != "" {
		config.OutputFormat = req.Format
	}
	if req.VideoCodec != "" {
		config.VideoCodec = req.VideoCodec
	}
	if req.AudioCodec != "" {
		config.AudioCodec = req.AudioCodec
	}
	config.Resolution = req.Resolution
	config.VideoBitrate = req.VideoBitrate
	config.AudioBitrate = req.AudioBitrate
	config.FPS = req.FPS
	if req.CRF > 0 {
		config.CRF = req.CRF
	}
	if req.Preset != "" {
		config.Preset = req.Preset
	}
	config.CopyAudio = req.CopyAudio
	config.CopyVideo = req.CopyVideo

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Minute)
	defer cancel()

	result, err := h.service.Convert(ctx, req.FilePath, config)
	if err != nil {
		jsonError(w, "CONVERT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// batchProcess 批量处理
func (h *VideoHandlers) batchProcess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Files   []string               `json:"files"`
		JobType processor.BatchJobType `json:"job_type"`
		Config  map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		jsonError(w, "NO_FILES", "files list is empty", http.StatusBadRequest)
		return
	}

	// 解析配置
	var config interface{}
	switch req.JobType {
	case processor.BatchJobSlice:
		config = processor.DefaultSliceConfig()
	case processor.BatchJobThumbnail:
		config = processor.DefaultThumbnailConfig()
	case processor.BatchJobConvert:
		config = processor.DefaultConvertConfig()
	}

	ctx := r.Context()

	result, err := h.service.BatchProcess(ctx, req.Files, req.JobType, config, func(progress float64, message string) {
		logrus.Debugf("Batch progress: %.1f%% - %s", progress, message)
	})

	if err != nil {
		jsonError(w, "BATCH_PROCESS_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// getBatchJob 获取批量任务状态
func (h *VideoHandlers) getBatchJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	// TODO: 实现任务状态查询
	jsonSuccess(w, map[string]interface{}{
		"job_id": jobID,
		"status": "completed",
	})
}

// uploadVideo 上传视频
func (h *VideoHandlers) uploadVideo(w http.ResponseWriter, r *http.Request) {
	maxSize := int64(2 * 1024 * 1024 * 1024) // 2GB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, "PARSE_FORM_FAILED", err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		jsonError(w, "MISSING_VIDEO_FILE", "video file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存文件
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	filePath := filepath.Join(h.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		jsonError(w, "SAVE_FILE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		jsonError(w, "SAVE_FILE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取视频信息
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	info, err := h.service.GetInfo(ctx, filePath)
	if err != nil {
		logrus.Warnf("Failed to get video info: %v", err)
		info = &processor.VideoInfo{
			FilePath: filePath,
			FileName: filename,
		}
	}

	jsonSuccess(w, map[string]interface{}{
		"file_path": filePath,
		"file_name": filename,
		"info":      info,
	})
}

// getSupportedFormats 获取支持的格式
func (h *VideoHandlers) getSupportedFormats(w http.ResponseWriter, r *http.Request) {
	formats := h.service.GetSupportedFormats()
	jsonSuccess(w, map[string]interface{}{
		"formats": formats,
	})
}

// getResolutionPresets 获取分辨率预设
func (h *VideoHandlers) getResolutionPresets(w http.ResponseWriter, r *http.Request) {
	presets := processor.GetResolutionPresets()
	jsonSuccess(w, presets)
}

// healthCheck 健康检查
func (h *VideoHandlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	available := h.service.IsAvailable()
	jsonSuccess(w, map[string]interface{}{
		"available": available,
		"service":   "video_processor",
	})
}

// VideoService 视频服务实现
type VideoService struct {
	processor processor.Processor
}

// NewVideoService 创建视频服务
func NewVideoService(proc processor.Processor) *VideoService {
	return &VideoService{processor: proc}
}

// GetInfo 获取视频信息
func (s *VideoService) GetInfo(ctx context.Context, filePath string) (*processor.VideoInfo, error) {
	return s.processor.GetInfo(ctx, filePath)
}

// Slice 切片视频
func (s *VideoService) Slice(ctx context.Context, filePath string, config *processor.SliceConfig) (*processor.SliceResult, error) {
	return s.processor.Slice(ctx, filePath, config)
}

// GenerateThumbnails 生成缩略图
func (s *VideoService) GenerateThumbnails(ctx context.Context, filePath string, config *processor.ThumbnailConfig) (*processor.ThumbnailResult, error) {
	return s.processor.GenerateThumbnails(ctx, filePath, config)
}

// Convert 转换视频格式
func (s *VideoService) Convert(ctx context.Context, filePath string, config *processor.ConvertConfig) (*processor.ConvertResult, error) {
	return s.processor.Convert(ctx, filePath, config)
}

// BatchProcess 批量处理
func (s *VideoService) BatchProcess(ctx context.Context, files []string, jobType processor.BatchJobType, config interface{}, callback processor.ProgressCallback) (*processor.BatchJob, error) {
	return s.processor.BatchProcess(ctx, files, jobType, config, callback)
}

// IsAvailable 检查是否可用
func (s *VideoService) IsAvailable() bool {
	return s.processor.IsAvailable()
}

// GetSupportedFormats 获取支持的格式
func (s *VideoService) GetSupportedFormats() []string {
	return s.processor.GetSupportedFormats()
}

// SetupVideoProcessor 设置视频处理器
func SetupVideoProcessor(config *VideoProcessorConfig) (*VideoService, error) {
	procConfig := processor.DefaultProcessorConfig()
	if config != nil {
		procConfig.FFmpegPath = config.FFmpegPath
		procConfig.FFprobePath = config.FFprobePath
		procConfig.TempDir = config.TempDir
		procConfig.MaxWorkers = config.MaxWorkers
	}

	proc := processor.NewFFmpegProcessor(procConfig)

	if !proc.IsAvailable() {
		logrus.Warn("FFmpeg is not available, video processing will be limited")
	}

	return NewVideoService(proc), nil
}

// VideoProcessorConfig 视频处理器配置
type VideoProcessorConfig struct {
	FFmpegPath  string `json:"ffmpeg_path"`
	FFprobePath string `json:"ffprobe_path"`
	TempDir     string `json:"temp_dir"`
	MaxWorkers  int    `json:"max_workers"`
}
