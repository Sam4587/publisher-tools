package processor

import (
	"context"
	"fmt"
	"time"
)

// VideoInfo 视频信息
type VideoInfo struct {
	FilePath     string            `json:"file_path"`
	FileName     string            `json:"file_name"`
	FileSize     int64             `json:"file_size"`
	Duration     float64           `json:"duration"`      // 时长(秒)
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	AspectRatio  string            `json:"aspect_ratio"`
	FPS          float64           `json:"fps"`
	BitRate      int64             `json:"bit_rate"`      // 比特率(bps)
	Codec        string            `json:"codec"`         // 视频编码
	AudioCodec   string            `json:"audio_codec"`   // 音频编码
	SampleRate   int               `json:"sample_rate"`   // 音频采样率
	Channels     int               `json:"channels"`      // 音频通道数
	Format       string            `json:"format"`        // 容器格式
	CreatedAt    time.Time         `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SliceConfig 切片配置
type SliceConfig struct {
	// 按时间切片
	StartTime    float64 `json:"start_time"`     // 开始时间(秒)
	EndTime      float64 `json:"end_time"`       // 结束时间(秒)
	Duration     float64 `json:"duration"`       // 切片时长(秒)

	// 按片段切片
	SegmentCount int     `json:"segment_count"`  // 分段数量
	SegmentDuration float64 `json:"segment_duration"` // 每段时长

	// 输出配置
	OutputDir    string  `json:"output_dir"`
	OutputFormat string  `json:"output_format"`  // mp4, webm, avi等
	OutputPrefix string  `json:"output_prefix"`  // 输出文件名前缀

	// 质量配置
	VideoBitrate int     `json:"video_bitrate"`  // 视频比特率
	AudioBitrate int     `json:"audio_bitrate"`  // 音频比特率
	Resolution   string  `json:"resolution"`     // 分辨率 1920x1080
	CRF          int     `json:"crf"`            // 恒定质量因子(0-51)
	Preset       string  `json:"preset"`         // 编码预设 ultrafast, fast, medium, slow
}

// DefaultSliceConfig 默认切片配置
func DefaultSliceConfig() *SliceConfig {
	return &SliceConfig{
		OutputFormat: "mp4",
		OutputDir:    "./data/slices",
		OutputPrefix: "slice",
		CRF:          23,
		Preset:       "medium",
	}
}

// SliceResult 切片结果
type SliceResult struct {
	SourceFile   string       `json:"source_file"`
	OutputDir    string       `json:"output_dir"`
	SliceCount   int          `json:"slice_count"`
	Slices       []SliceInfo  `json:"slices"`
	Duration     time.Duration `json:"duration"`
	TotalSize    int64        `json:"total_size"`
}

// SliceInfo 切片信息
type SliceInfo struct {
	Index     int     `json:"index"`
	FilePath  string  `json:"file_path"`
	FileName  string  `json:"file_name"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Duration  float64 `json:"duration"`
	Size      int64   `json:"size"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
}

// ThumbnailConfig 缩略图配置
type ThumbnailConfig struct {
	OutputDir     string  `json:"output_dir"`
	OutputFormat  string  `json:"output_format"`  // jpg, png, webp
	Width         int     `json:"width"`          // 宽度(高度自动计算)
	Height        int     `json:"height"`         // 高度(宽度自动计算)
	Quality       int     `json:"quality"`        // 图片质量(1-100)
	Timestamps    []float64 `json:"timestamps"`   // 指定时间点截图
	Interval      float64 `json:"interval"`       // 按间隔截图(秒)
	Count         int     `json:"count"`          // 截图数量(均匀分布)
	SpriteSheet   bool    `json:"sprite_sheet"`   // 是否生成雪碧图
	Columns       int     `json:"columns"`        // 雪碧图列数
	Rows          int     `json:"rows"`           // 雪碧图行数
}

// DefaultThumbnailConfig 默认缩略图配置
func DefaultThumbnailConfig() *ThumbnailConfig {
	return &ThumbnailConfig{
		OutputDir:    "./data/thumbnails",
		OutputFormat: "jpg",
		Quality:      85,
		Count:        5,
	}
}

// ThumbnailResult 缩略图结果
type ThumbnailResult struct {
	SourceFile    string           `json:"source_file"`
	OutputDir     string           `json:"output_dir"`
	Thumbnails    []ThumbnailInfo  `json:"thumbnails"`
	SpriteSheet   *SpriteSheetInfo `json:"sprite_sheet,omitempty"`
	Duration      time.Duration    `json:"duration"`
}

// ThumbnailInfo 缩略图信息
type ThumbnailInfo struct {
	Index     int     `json:"index"`
	FilePath  string  `json:"file_path"`
	FileName  string  `json:"file_name"`
	Timestamp float64 `json:"timestamp"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Size      int64   `json:"size"`
}

// SpriteSheetInfo 雪碧图信息
type SpriteSheetInfo struct {
	FilePath  string `json:"file_path"`
	FileName  string `json:"file_name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Columns   int    `json:"columns"`
	Rows      int    `json:"rows"`
	Count     int    `json:"count"`
	Size      int64  `json:"size"`
}

// ConvertConfig 转换配置
type ConvertConfig struct {
	OutputPath    string `json:"output_path"`
	OutputFormat  string `json:"output_format"`  // mp4, webm, avi, mov, mkv
	VideoCodec    string `json:"video_codec"`    // h264, h265, vp9, av1
	AudioCodec    string `json:"audio_codec"`    // aac, mp3, opus
	Resolution    string `json:"resolution"`     // 1920x1080, 1280x720
	VideoBitrate  int    `json:"video_bitrate"`  // 视频比特率
	AudioBitrate  int    `json:"audio_bitrate"`  // 音频比特率
	FPS           float64 `json:"fps"`           // 帧率
	CRF           int    `json:"crf"`            // 恒定质量因子
	Preset        string `json:"preset"`         // 编码预设
	CopyAudio     bool   `json:"copy_audio"`     // 直接复制音频流
	CopyVideo     bool   `json:"copy_video"`     // 直接复制视频流
	Overwrite     bool   `json:"overwrite"`      // 覆盖已存在文件
}

// DefaultConvertConfig 默认转换配置
func DefaultConvertConfig() *ConvertConfig {
	return &ConvertConfig{
		OutputFormat: "mp4",
		VideoCodec:   "libx264",
		AudioCodec:   "aac",
		CRF:          23,
		Preset:       "medium",
		Overwrite:    true,
	}
}

// ConvertResult 转换结果
type ConvertResult struct {
	SourceFile  string        `json:"source_file"`
	OutputFile  string        `json:"output_file"`
	SourceInfo  *VideoInfo    `json:"source_info"`
	OutputInfo  *VideoInfo    `json:"output_info"`
	Duration    time.Duration `json:"duration"`
	CompressionRatio float64  `json:"compression_ratio"` // 压缩比
}

// BatchJob 批量任务
type BatchJob struct {
	ID           string        `json:"id"`
	Type         BatchJobType  `json:"type"`
	Status       BatchJobStatus `json:"status"`
	TotalFiles   int           `json:"total_files"`
	Completed    int           `json:"completed"`
	Failed       int           `json:"failed"`
	Results      []interface{} `json:"results"`
	Errors       []BatchError  `json:"errors"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
}

// BatchJobType 批量任务类型
type BatchJobType string

const (
	BatchJobSlice      BatchJobType = "slice"
	BatchJobThumbnail  BatchJobType = "thumbnail"
	BatchJobConvert    BatchJobType = "convert"
	BatchJobInfo       BatchJobType = "info"
)

// BatchJobStatus 批量任务状态
type BatchJobStatus string

const (
	BatchJobPending   BatchJobStatus = "pending"
	BatchJobRunning   BatchJobStatus = "running"
	BatchJobCompleted BatchJobStatus = "completed"
	BatchJobFailed    BatchJobStatus = "failed"
	BatchJobCancelled BatchJobStatus = "cancelled"
)

// BatchError 批量错误
type BatchError struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

// ProgressCallback 进度回调
type ProgressCallback func(progress float64, message string)

// Processor 视频处理器接口
type Processor interface {
	// GetInfo 获取视频信息
	GetInfo(ctx context.Context, filePath string) (*VideoInfo, error)

	// Slice 按配置切片视频
	Slice(ctx context.Context, filePath string, config *SliceConfig) (*SliceResult, error)

	// SliceByTime 按时间范围切片
	SliceByTime(ctx context.Context, filePath string, startTime, endTime float64, outputPath string) error

	// SliceBySegments 按分段数量切片
	SliceBySegments(ctx context.Context, filePath string, segmentCount int, config *SliceConfig) (*SliceResult, error)

	// GenerateThumbnails 生成缩略图
	GenerateThumbnails(ctx context.Context, filePath string, config *ThumbnailConfig) (*ThumbnailResult, error)

	// GenerateSpriteSheet 生成雪碧图
	GenerateSpriteSheet(ctx context.Context, filePath string, config *ThumbnailConfig) (*SpriteSheetInfo, error)

	// Convert 转换视频格式
	Convert(ctx context.Context, filePath string, config *ConvertConfig) (*ConvertResult, error)

	// BatchProcess 批量处理
	BatchProcess(ctx context.Context, files []string, jobType BatchJobType, config interface{}, callback ProgressCallback) (*BatchJob, error)

	// IsAvailable 检查FFmpeg是否可用
	IsAvailable() bool

	// GetSupportedFormats 获取支持的格式
	GetSupportedFormats() []string

	// GetSupportedCodecs 获取支持的编码
	GetSupportedCodecs() (videoCodecs, audioCodecs []string)
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	FFmpegPath    string `json:"ffmpeg_path"`
	FFprobePath   string `json:"ffprobe_path"`
	TempDir       string `json:"temp_dir"`
	MaxWorkers    int    `json:"max_workers"`    // 最大并行数
	Timeout       time.Duration `json:"timeout"`
}

// DefaultProcessorConfig 默认处理器配置
func DefaultProcessorConfig() *ProcessorConfig {
	return &ProcessorConfig{
		FFmpegPath:  "ffmpeg",
		FFprobePath: "ffprobe",
		TempDir:     "./data/temp/video",
		MaxWorkers:  4,
		Timeout:     30 * time.Minute,
	}
}

// ResolutionPreset 分辨率预设
type ResolutionPreset struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// GetResolutionPresets 获取分辨率预设
func GetResolutionPresets() []ResolutionPreset {
	return []ResolutionPreset{
		{Name: "4K", Width: 3840, Height: 2160},
		{Name: "1080p", Width: 1920, Height: 1080},
		{Name: "720p", Width: 1280, Height: 720},
		{Name: "480p", Width: 854, Height: 480},
		{Name: "360p", Width: 640, Height: 360},
		{Name: "240p", Width: 426, Height: 240},
	}
}

// ParseResolution 解析分辨率字符串
func ParseResolution(res string) (width, height int) {
	// 尝试预设名称
	for _, preset := range GetResolutionPresets() {
		if preset.Name == res {
			return preset.Width, preset.Height
		}
	}

	// 尝试 WxH 格式
	_, err := fmt.Sscanf(res, "%dx%d", &width, &height)
	if err == nil {
		return width, height
	}

	// 默认返回720p
	return 1280, 720
}

// FormatDuration 格式化时长
func FormatDuration(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
	}
	return fmt.Sprintf("%02d:%02d.%03d", minutes, secs, millis)
}

// FormatFileSize 格式化文件大小
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
