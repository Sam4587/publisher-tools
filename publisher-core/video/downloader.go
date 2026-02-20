package video

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"publisher-core/database"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Downloader 视频下载器
type Downloader struct {
	db         *gorm.DB
	outputDir  string
	ytDlpPath  string
	proxy      string
	maxRetries int
}

// DownloaderConfig 下载器配置
type DownloaderConfig struct {
	OutputDir  string `json:"output_dir"`
	YtDlpPath  string `json:"yt_dlp_path"`
	Proxy      string `json:"proxy"`
	MaxRetries int    `json:"max_retries"`
}

// DefaultDownloaderConfig 默认配置
func DefaultDownloaderConfig() *DownloaderConfig {
	return &DownloaderConfig{
		OutputDir:  "./data/videos",
		YtDlpPath:  "yt-dlp",
		MaxRetries: 3,
	}
}

// NewDownloader 创建下载器
func NewDownloader(db *gorm.DB, config *DownloaderConfig) *Downloader {
	if config == nil {
		config = DefaultDownloaderConfig()
	}

	// 确保输出目录存在
	os.MkdirAll(config.OutputDir, 0755)

	return &Downloader{
		db:         db,
		outputDir:  config.OutputDir,
		ytDlpPath:  config.YtDlpPath,
		proxy:      config.Proxy,
		maxRetries: config.MaxRetries,
	}
}

// VideoInfo 视频信息
type VideoInfo struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Duration     int               `json:"duration"` // 秒
	Uploader     string            `json:"uploader"`
	UploadDate   string            `json:"upload_date"`
	ViewCount    int64             `json:"view_count"`
	LikeCount    int64             `json:"like_count"`
	Thumbnail    string            `json:"thumbnail"`
	Platform     string            `json:"platform"`
	URL          string            `json:"url"`
	Formats      []FormatInfo      `json:"formats"`
	Subtitles    map[string]string `json:"subtitles"`
	AutoCaptions map[string]string `json:"auto_captions"`
}

// FormatInfo 格式信息
type FormatInfo struct {
	FormatID   string `json:"format_id"`
	Ext        string `json:"ext"`
	Resolution string `json:"resolution"`
	FPS        int    `json:"fps"`
	VCodec     string `json:"vcodec"`
	ACodec     string `json:"acodec"`
	FileSize   int64  `json:"filesize"`
}

// DownloadResult 下载结果
type DownloadResult struct {
	VideoID    string `json:"video_id"`
	VideoPath  string `json:"video_path"`
	AudioPath  string `json:"audio_path"`
	Thumbnail  string `json:"thumbnail"`
	Duration   int    `json:"duration"`
	Title      string `json:"title"`
	Platform   string `json:"platform"`
}

// GetVideoInfo 获取视频信息
func (d *Downloader) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	args := []string{
		"--dump-json",
		"--no-download",
		"--no-warnings",
	}

	if d.proxy != "" {
		args = append(args, "--proxy", d.proxy)
	}

	args = append(args, url)

	cmd := exec.CommandContext(ctx, d.ytDlpPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("execute yt-dlp: %w", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("parse video info: %w", err)
	}

	// 检测平台
	info.Platform = detectPlatform(url)
	info.URL = url

	return &info, nil
}

// Download 下载视频
func (d *Downloader) Download(ctx context.Context, url string, opts *DownloadOptions) (*DownloadResult, error) {
	if opts == nil {
		opts = &DownloadOptions{}
	}

	// 获取视频信息
	info, err := d.GetVideoInfo(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("get video info: %w", err)
	}

	// 创建视频记录
	videoID := uuid.New().String()
	videoDir := filepath.Join(d.outputDir, videoID)
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		return nil, fmt.Errorf("create video directory: %w", err)
	}

	// 构建下载参数
	videoPath := filepath.Join(videoDir, "video.%(ext)s")
	audioPath := filepath.Join(videoDir, "audio.%(ext)s")
	thumbnailPath := filepath.Join(videoDir, "thumbnail.%(ext)s")

	args := []string{
		"--no-warnings",
		"--no-playlist",
	}

	// 下载视频
	if !opts.AudioOnly {
		args = append(args,
			"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
			"-o", videoPath,
		)
	}

	// 下载音频
	args = append(args,
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", audioPath,
	)

	// 下载缩略图
	args = append(args,
		"--write-thumbnail",
		"-o", thumbnailPath,
	)

	// 字幕
	if opts.DownloadSubtitles {
		args = append(args,
			"--write-subs",
			"--write-auto-subs",
			"--sub-lang", "zh-Hans,zh-Hant,en",
			"--sub-format", "srt",
		)
	}

	// 代理
	if d.proxy != "" {
		args = append(args, "--proxy", d.proxy)
	}

	args = append(args, url)

	// 执行下载
	var lastErr error
	for i := 0; i < d.maxRetries; i++ {
		cmd := exec.CommandContext(ctx, d.ytDlpPath, args...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			break
		}
		lastErr = fmt.Errorf("yt-dlp error: %s", string(output))
		logrus.Warnf("Download attempt %d failed: %v", i+1, lastErr)

		if i < d.maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	// 查找实际下载的文件
	actualVideoPath := d.findFile(videoDir, "video")
	actualAudioPath := d.findFile(videoDir, "audio")
	actualThumbnailPath := d.findFile(videoDir, "thumbnail")

	// 保存到数据库
	video := &database.Video{
		ID:        videoID,
		URL:       url,
		Platform:  info.Platform,
		Title:     info.Title,
		Duration:  info.Duration,
		Status:    "downloaded",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if d.db != nil {
		if err := d.db.Create(video).Error; err != nil {
			logrus.Warnf("Failed to save video record: %v", err)
		}
	}

	return &DownloadResult{
		VideoID:   videoID,
		VideoPath: actualVideoPath,
		AudioPath: actualAudioPath,
		Thumbnail: actualThumbnailPath,
		Duration:  info.Duration,
		Title:     info.Title,
		Platform:  info.Platform,
	}, nil
}

// DownloadOptions 下载选项
type DownloadOptions struct {
	AudioOnly         bool `json:"audio_only"`
	DownloadSubtitles bool `json:"download_subtitles"`
	MaxFileSize       int64 `json:"max_file_size"` // 0 表示不限制
}

// findFile 查找文件
func (d *Downloader) findFile(dir, prefix string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) {
			return filepath.Join(dir, f.Name())
		}
	}
	return ""
}

// detectPlatform 检测平台
func detectPlatform(url string) string {
	platforms := map[string]string{
		"youtube.com":     "youtube",
		"youtu.be":        "youtube",
		"bilibili.com":    "bilibili",
		"b23.tv":          "bilibili",
		"douyin.com":      "douyin",
		"v.douyin.com":    "douyin",
		"tiktok.com":      "tiktok",
		"vm.tiktok.com":   "tiktok",
		"twitter.com":     "twitter",
		"x.com":           "twitter",
		"vimeo.com":       "vimeo",
		"facebook.com":    "facebook",
		"instagram.com":   "instagram",
		"weibo.com":       "weibo",
		"weibo.cn":        "weibo",
		"youku.com":       "youku",
		"iqiyi.com":       "iqiyi",
		"v.qq.com":        "qq",
		"mgtv.com":        "mgtv",
		"acfun.cn":        "acfun",
		"zhihu.com":       "zhihu",
	}

	for domain, platform := range platforms {
		if strings.Contains(url, domain) {
			return platform
		}
	}

	return "unknown"
}

// IsYtDlpAvailable 检查 yt-dlp 是否可用
func (d *Downloader) IsYtDlpAvailable() bool {
	cmd := exec.Command(d.ytDlpPath, "--version")
	return cmd.Run() == nil
}

// GetSupportedPlatforms 获取支持的平台列表
func GetSupportedPlatforms() []string {
	return []string{
		"youtube",
		"bilibili",
		"douyin",
		"tiktok",
		"twitter",
		"vimeo",
		"facebook",
		"instagram",
		"weibo",
		"youku",
		"iqiyi",
		"qq",
		"mgtv",
		"acfun",
		"zhihu",
	}
}

// CleanupOldVideos 清理旧视频文件
func (d *Downloader) CleanupOldVideos(olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	var videos []database.Video

	if d.db == nil {
		return 0, nil
	}

	// 查找旧视频记录
	if err := d.db.Where("created_at < ? AND status = ?", cutoff, "completed").Find(&videos).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, video := range videos {
		videoDir := filepath.Join(d.outputDir, video.ID)
		if err := os.RemoveAll(videoDir); err != nil {
			logrus.Warnf("Failed to remove video directory %s: %v", videoDir, err)
			continue
		}

		// 删除数据库记录
		if err := d.db.Delete(&video).Error; err != nil {
			logrus.Warnf("Failed to delete video record %s: %v", video.ID, err)
			continue
		}

		count++
	}

	logrus.Infof("Cleaned up %d old videos", count)
	return count, nil
}
