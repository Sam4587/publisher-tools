package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FFmpegProcessor FFmpeg视频处理器
type FFmpegProcessor struct {
	ffmpegPath  string
	ffprobePath string
	tempDir     string
	maxWorkers  int
	timeout     time.Duration
}

// NewFFmpegProcessor 创建FFmpeg处理器
func NewFFmpegProcessor(config *ProcessorConfig) *FFmpegProcessor {
	if config == nil {
		config = DefaultProcessorConfig()
	}

	os.MkdirAll(config.TempDir, 0755)

	return &FFmpegProcessor{
		ffmpegPath:  config.FFmpegPath,
		ffprobePath: config.FFprobePath,
		tempDir:     config.TempDir,
		maxWorkers:  config.MaxWorkers,
		timeout:     config.Timeout,
	}
}

// GetInfo 获取视频信息
func (p *FFmpegProcessor) GetInfo(ctx context.Context, filePath string) (*VideoInfo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// 使用ffprobe获取信息
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	cmd := exec.CommandContext(ctx, p.ffprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %w", err)
	}

	// 解析JSON输出
	var probeData struct {
		Format struct {
			Filename    string  `json:"filename"`
			Duration    string  `json:"duration"`
			Size        string  `json:"size"`
			BitRate     string  `json:"bit_rate"`
			FormatName  string  `json:"format_name"`
			Tags        map[string]string `json:"tags"`
		} `json:"format"`
		Streams []struct {
			Index      int    `json:"index"`
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			BitRate    string `json:"bit_rate"`
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
			Duration   string `json:"duration"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &probeData); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	info := &VideoInfo{
		FilePath:  filePath,
		FileName:  filepath.Base(filePath),
		Format:    probeData.Format.FormatName,
		Metadata:  probeData.Format.Tags,
	}

	// 解析文件大小
	if size, err := strconv.ParseInt(probeData.Format.Size, 10, 64); err == nil {
		info.FileSize = size
	}

	// 解析时长
	if duration, err := strconv.ParseFloat(probeData.Format.Duration, 64); err == nil {
		info.Duration = duration
	}

	// 解析比特率
	if bitrate, err := strconv.ParseInt(probeData.Format.BitRate, 10, 64); err == nil {
		info.BitRate = bitrate
	}

	// 解析流信息
	for _, stream := range probeData.Streams {
		switch stream.CodecType {
		case "video":
			info.Codec = stream.CodecName
			info.Width = stream.Width
			info.Height = stream.Height

			// 解析帧率
			if stream.RFrameRate != "" {
				parts := strings.Split(stream.RFrameRate, "/")
				if len(parts) == 2 {
					num, err1 := strconv.ParseFloat(parts[0], 64)
					den, err2 := strconv.ParseFloat(parts[1], 64)
					if err1 == nil && err2 == nil && den > 0 {
						info.FPS = num / den
					}
				}
			}

			// 计算宽高比
			if info.Width > 0 && info.Height > 0 {
				gcd := p.gcd(info.Width, info.Height)
				info.AspectRatio = fmt.Sprintf("%d:%d", info.Width/gcd, info.Height/gcd)
			}

		case "audio":
			info.AudioCodec = stream.CodecName
			if stream.SampleRate != "" {
				if rate, err := strconv.Atoi(stream.SampleRate); err == nil {
					info.SampleRate = rate
				}
			}
			info.Channels = stream.Channels
		}
	}

	// 获取文件修改时间
	if fileInfo, err := os.Stat(filePath); err == nil {
		info.CreatedAt = fileInfo.ModTime()
	}

	return info, nil
}

// Slice 按配置切片视频
func (p *FFmpegProcessor) Slice(ctx context.Context, filePath string, config *SliceConfig) (*SliceResult, error) {
	if config == nil {
		config = DefaultSliceConfig()
	}

	// 确保输出目录存在
	os.MkdirAll(config.OutputDir, 0755)

	// 获取视频信息
	info, err := p.GetInfo(ctx, filePath)
	if err != nil {
		return nil, err
	}

	result := &SliceResult{
		SourceFile: filePath,
		OutputDir:  config.OutputDir,
		Slices:     make([]SliceInfo, 0),
	}

	startTime := time.Now()

	// 计算切片数量和时长
	var slices []sliceRange
	if config.SegmentCount > 0 {
		// 按分段数量切片
		segmentDuration := info.Duration / float64(config.SegmentCount)
		for i := 0; i < config.SegmentCount; i++ {
			slices = append(slices, sliceRange{
				start: float64(i) * segmentDuration,
				end:   float64(i+1) * segmentDuration,
			})
		}
	} else if config.SegmentDuration > 0 {
		// 按每段时长切片
		count := int(info.Duration / config.SegmentDuration)
		for i := 0; i < count; i++ {
			slices = append(slices, sliceRange{
				start: float64(i) * config.SegmentDuration,
				end:   float64(i+1) * config.SegmentDuration,
			})
		}
		// 添加最后一段
		if float64(count)*config.SegmentDuration < info.Duration {
			slices = append(slices, sliceRange{
				start: float64(count) * config.SegmentDuration,
				end:   info.Duration,
			})
		}
	} else if config.StartTime >= 0 && config.EndTime > config.StartTime {
		// 按时间范围切片
		slices = append(slices, sliceRange{
			start: config.StartTime,
			end:   config.EndTime,
		})
	} else {
		// 默认按固定时长切片
		chunkDuration := config.Duration
		if chunkDuration <= 0 {
			chunkDuration = 300 // 默认5分钟
		}
		count := int(info.Duration / chunkDuration)
		for i := 0; i < count; i++ {
			slices = append(slices, sliceRange{
				start: float64(i) * chunkDuration,
				end:   float64(i+1) * chunkDuration,
			})
		}
		if float64(count)*chunkDuration < info.Duration {
			slices = append(slices, sliceRange{
				start: float64(count) * chunkDuration,
				end:   info.Duration,
			})
		}
	}

	// 执行切片
	for i, s := range slices {
		outputName := fmt.Sprintf("%s_%04d.%s", config.OutputPrefix, i+1, config.OutputFormat)
		outputPath := filepath.Join(config.OutputDir, outputName)

		err := p.SliceByTime(ctx, filePath, s.start, s.end, outputPath)
		if err != nil {
			logrus.Warnf("Failed to create slice %d: %v", i, err)
			continue
		}

		// 获取切片信息
		sliceInfo, err := p.GetInfo(ctx, outputPath)
		if err != nil {
			logrus.Warnf("Failed to get slice info %d: %v", i, err)
			continue
		}

		result.Slices = append(result.Slices, SliceInfo{
			Index:     i + 1,
			FilePath:  outputPath,
			FileName:  outputName,
			StartTime: s.start,
			EndTime:   s.end,
			Duration:  s.end - s.start,
			Size:      sliceInfo.FileSize,
			Width:     sliceInfo.Width,
			Height:    sliceInfo.Height,
		})
		result.TotalSize += sliceInfo.FileSize
	}

	result.SliceCount = len(result.Slices)
	result.Duration = time.Since(startTime)

	return result, nil
}

type sliceRange struct {
	start, end float64
}

// SliceByTime 按时间范围切片
func (p *FFmpegProcessor) SliceByTime(ctx context.Context, filePath string, startTime, endTime float64, outputPath string) error {
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-i", filePath,
		"-t", fmt.Sprintf("%.3f", endTime-startTime),
		"-c", "copy",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg slice error: %s - %w", string(output), err)
	}

	return nil
}

// SliceBySegments 按分段数量切片
func (p *FFmpegProcessor) SliceBySegments(ctx context.Context, filePath string, segmentCount int, config *SliceConfig) (*SliceResult, error) {
	if config == nil {
		config = DefaultSliceConfig()
	}
	config.SegmentCount = segmentCount
	return p.Slice(ctx, filePath, config)
}

// GenerateThumbnails 生成缩略图
func (p *FFmpegProcessor) GenerateThumbnails(ctx context.Context, filePath string, config *ThumbnailConfig) (*ThumbnailResult, error) {
	if config == nil {
		config = DefaultThumbnailConfig()
	}

	// 确保输出目录存在
	os.MkdirAll(config.OutputDir, 0755)

	// 获取视频信息
	info, err := p.GetInfo(ctx, filePath)
	if err != nil {
		return nil, err
	}

	result := &ThumbnailResult{
		SourceFile: filePath,
		OutputDir:  config.OutputDir,
		Thumbnails: make([]ThumbnailInfo, 0),
	}

	startTime := time.Now()

	// 计算截图时间点
	var timestamps []float64
	if len(config.Timestamps) > 0 {
		timestamps = config.Timestamps
	} else if config.Interval > 0 {
		for t := 0.0; t < info.Duration; t += config.Interval {
			timestamps = append(timestamps, t)
		}
	} else if config.Count > 0 {
		interval := info.Duration / float64(config.Count+1)
		for i := 1; i <= config.Count; i++ {
			timestamps = append(timestamps, float64(i)*interval)
		}
	} else {
		// 默认5张
		interval := info.Duration / 6
		for i := 1; i <= 5; i++ {
			timestamps = append(timestamps, float64(i)*interval)
		}
	}

	// 生成缩略图
	for i, ts := range timestamps {
		if ts >= info.Duration {
			continue
		}

		outputName := fmt.Sprintf("thumb_%04d.%s", i+1, config.OutputFormat)
		outputPath := filepath.Join(config.OutputDir, outputName)

		err := p.captureFrame(ctx, filePath, ts, outputPath, config)
		if err != nil {
			logrus.Warnf("Failed to capture frame at %.2fs: %v", ts, err)
			continue
		}

		// 获取文件信息
		fileInfo, _ := os.Stat(outputPath)

		result.Thumbnails = append(result.Thumbnails, ThumbnailInfo{
			Index:     i + 1,
			FilePath:  outputPath,
			FileName:  outputName,
			Timestamp: ts,
			Width:     config.Width,
			Height:    config.Height,
			Size:      fileInfo.Size(),
		})
	}

	// 生成雪碧图
	if config.SpriteSheet && len(result.Thumbnails) > 0 {
		spriteInfo, err := p.createSpriteSheet(ctx, result.Thumbnails, config)
		if err != nil {
			logrus.Warnf("Failed to create sprite sheet: %v", err)
		} else {
			result.SpriteSheet = spriteInfo
		}
	}

	result.Duration = time.Since(startTime)

	return result, nil
}

// captureFrame 捕获单帧
func (p *FFmpegProcessor) captureFrame(ctx context.Context, filePath string, timestamp float64, outputPath string, config *ThumbnailConfig) error {
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", filePath,
		"-vframes", "1",
	}

	// 设置尺寸
	if config.Width > 0 || config.Height > 0 {
		scale := ""
		if config.Width > 0 && config.Height > 0 {
			scale = fmt.Sprintf("scale=%d:%d", config.Width, config.Height)
		} else if config.Width > 0 {
			scale = fmt.Sprintf("scale=%d:-1", config.Width)
		} else {
			scale = fmt.Sprintf("scale=-1:%d", config.Height)
		}
		args = append(args, "-vf", scale)
	}

	// 设置质量
	if config.Quality > 0 && config.OutputFormat == "jpg" {
		args = append(args, "-q:v", strconv.Itoa(31-config.Quality*31/100))
	}

	args = append(args, outputPath)

	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg capture error: %s - %w", string(output), err)
	}

	return nil
}

// createSpriteSheet 创建雪碧图
func (p *FFmpegProcessor) createSpriteSheet(ctx context.Context, thumbnails []ThumbnailInfo, config *ThumbnailConfig) (*SpriteSheetInfo, error) {
	if len(thumbnails) == 0 {
		return nil, fmt.Errorf("no thumbnails")
	}

	columns := config.Columns
	if columns <= 0 {
		columns = 5
	}
	rows := (len(thumbnails) + columns - 1) / columns

	// 创建临时文件列表
	listFile := filepath.Join(p.tempDir, fmt.Sprintf("sprite_list_%d.txt", time.Now().UnixNano()))
	defer os.Remove(listFile)

	var listContent strings.Builder
	for _, t := range thumbnails {
		listContent.WriteString(fmt.Sprintf("file '%s'\n", t.FilePath))
	}
	os.WriteFile(listFile, []byte(listContent.String()), 0644)

	// 计算雪碧图尺寸
	tileWidth := config.Width
	tileHeight := config.Height
	if tileWidth <= 0 {
		tileWidth = 160
	}
	if tileHeight <= 0 {
		tileHeight = 90
	}

	outputName := fmt.Sprintf("sprite_%dx%d.%s", columns, rows, config.OutputFormat)
	outputPath := filepath.Join(config.OutputDir, outputName)

	args := []string{
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-vf", fmt.Sprintf("scale=%d:%d,tile=%dx%d", tileWidth, tileHeight, columns, rows),
		outputPath,
	}

	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg sprite error: %s - %w", string(output), err)
	}

	fileInfo, _ := os.Stat(outputPath)

	return &SpriteSheetInfo{
		FilePath: outputPath,
		FileName: outputName,
		Width:    tileWidth * columns,
		Height:   tileHeight * rows,
		Columns:  columns,
		Rows:     rows,
		Count:    len(thumbnails),
		Size:     fileInfo.Size(),
	}, nil
}

// GenerateSpriteSheet 生成雪碧图
func (p *FFmpegProcessor) GenerateSpriteSheet(ctx context.Context, filePath string, config *ThumbnailConfig) (*SpriteSheetInfo, error) {
	result, err := p.GenerateThumbnails(ctx, filePath, config)
	if err != nil {
		return nil, err
	}
	return result.SpriteSheet, nil
}

// Convert 转换视频格式
func (p *FFmpegProcessor) Convert(ctx context.Context, filePath string, config *ConvertConfig) (*ConvertResult, error) {
	if config == nil {
		config = DefaultConvertConfig()
	}

	// 获取源文件信息
	sourceInfo, err := p.GetInfo(ctx, filePath)
	if err != nil {
		return nil, err
	}

	// 确定输出路径
	outputPath := config.OutputPath
	if outputPath == "" {
		ext := filepath.Ext(filePath)
		base := strings.TrimSuffix(filepath.Base(filePath), ext)
		outputPath = fmt.Sprintf("%s_converted.%s", base, config.OutputFormat)
	}

	// 确保输出目录存在
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	startTime := time.Now()

	// 构建FFmpeg命令
	args := []string{"-y", "-i", filePath}

	// 视频编码
	if config.CopyVideo {
		args = append(args, "-c:v", "copy")
	} else if config.VideoCodec != "" {
		args = append(args, "-c:v", config.VideoCodec)
		if config.CRF > 0 {
			args = append(args, "-crf", strconv.Itoa(config.CRF))
		}
		if config.Preset != "" {
			args = append(args, "-preset", config.Preset)
		}
		if config.VideoBitrate > 0 {
			args = append(args, "-b:v", fmt.Sprintf("%dk", config.VideoBitrate/1000))
		}
	}

	// 音频编码
	if config.CopyAudio {
		args = append(args, "-c:a", "copy")
	} else if config.AudioCodec != "" {
		args = append(args, "-c:a", config.AudioCodec)
		if config.AudioBitrate > 0 {
			args = append(args, "-b:a", fmt.Sprintf("%dk", config.AudioBitrate/1000))
		}
	}

	// 分辨率
	if config.Resolution != "" {
		width, height := ParseResolution(config.Resolution)
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", width, height))
	}

	// 帧率
	if config.FPS > 0 {
		args = append(args, "-r", strconv.FormatFloat(config.FPS, 'f', -1, 64))
	}

	args = append(args, outputPath)

	// 执行转换
	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg convert error: %s - %w", string(output), err)
	}

	// 获取输出文件信息
	outputInfo, err := p.GetInfo(ctx, outputPath)
	if err != nil {
		return nil, err
	}

	result := &ConvertResult{
		SourceFile:  filePath,
		OutputFile:  outputPath,
		SourceInfo:  sourceInfo,
		OutputInfo:  outputInfo,
		Duration:    time.Since(startTime),
	}

	// 计算压缩比
	if sourceInfo.FileSize > 0 {
		result.CompressionRatio = float64(outputInfo.FileSize) / float64(sourceInfo.FileSize)
	}

	return result, nil
}

// BatchProcess 批量处理
func (p *FFmpegProcessor) BatchProcess(ctx context.Context, files []string, jobType BatchJobType, config interface{}, callback ProgressCallback) (*BatchJob, error) {
	job := &BatchJob{
		ID:         uuid.New().String(),
		Type:       jobType,
		Status:     BatchJobRunning,
		TotalFiles: len(files),
		StartTime:  time.Now(),
		Results:    make([]interface{}, 0),
		Errors:     make([]BatchError, 0),
	}

	// 使用工作池
	var wg sync.WaitGroup
	sem := make(chan struct{}, p.maxWorkers)
	var mu sync.Mutex

	for i, file := range files {
		wg.Add(1)
		go func(idx int, f string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var result interface{}
			var err error

			switch jobType {
			case BatchJobInfo:
				result, err = p.GetInfo(ctx, f)
			case BatchJobSlice:
				result, err = p.Slice(ctx, f, config.(*SliceConfig))
			case BatchJobThumbnail:
				result, err = p.GenerateThumbnails(ctx, f, config.(*ThumbnailConfig))
			case BatchJobConvert:
				result, err = p.Convert(ctx, f, config.(*ConvertConfig))
			}

			mu.Lock()
			if err != nil {
				job.Failed++
				job.Errors = append(job.Errors, BatchError{
					File:    f,
					Message: err.Error(),
				})
			} else {
				job.Completed++
				job.Results = append(job.Results, result)
			}
			mu.Unlock()

			// 进度回调
			if callback != nil {
				progress := float64(job.Completed+job.Failed) / float64(job.TotalFiles) * 100
				callback(progress, fmt.Sprintf("Processing %s (%d/%d)", filepath.Base(f), job.Completed+job.Failed, job.TotalFiles))
			}
		}(i, file)
	}

	wg.Wait()

	job.EndTime = time.Now()
	job.Duration = job.EndTime.Sub(job.StartTime)

	if job.Failed == job.TotalFiles {
		job.Status = BatchJobFailed
	} else {
		job.Status = BatchJobCompleted
	}

	return job, nil
}

// IsAvailable 检查FFmpeg是否可用
func (p *FFmpegProcessor) IsAvailable() bool {
	cmd := exec.Command(p.ffmpegPath, "-version")
	return cmd.Run() == nil
}

// GetSupportedFormats 获取支持的格式
func (p *FFmpegProcessor) GetSupportedFormats() []string {
	return []string{
		"mp4", "webm", "avi", "mov", "mkv", "flv", "wmv", "m4v",
		"mp3", "wav", "aac", "ogg", "flac", "m4a",
	}
}

// GetSupportedCodecs 获取支持的编码
func (p *FFmpegProcessor) GetSupportedCodecs() (videoCodecs, audioCodecs []string) {
	return []string{"h264", "h265", "vp9", "av1", "mpeg4", "mpeg2video"},
		[]string{"aac", "mp3", "opus", "vorbis", "flac", "pcm_s16le"}
}

// gcd 计算最大公约数
func (p *FFmpegProcessor) gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// ExtractAudio 提取音频
func (p *FFmpegProcessor) ExtractAudio(ctx context.Context, filePath string, outputPath string, audioCodec string) error {
	if audioCodec == "" {
		audioCodec = "aac"
	}

	args := []string{
		"-y",
		"-i", filePath,
		"-vn",
		"-c:a", audioCodec,
		outputPath,
	}

	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg extract audio error: %s - %w", string(output), err)
	}

	return nil
}

// ConcatVideos 合并视频
func (p *FFmpegProcessor) ConcatVideos(ctx context.Context, files []string, outputPath string) error {
	// 创建临时文件列表
	listFile := filepath.Join(p.tempDir, fmt.Sprintf("concat_%d.txt", time.Now().UnixNano()))
	defer os.Remove(listFile)

	var listContent strings.Builder
	for _, f := range files {
		absPath, _ := filepath.Abs(f)
		listContent.WriteString(fmt.Sprintf("file '%s'\n", absPath))
	}
	os.WriteFile(listFile, []byte(listContent.String()), 0644)

	args := []string{
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg concat error: %s - %w", string(output), err)
	}

	return nil
}

// AddWatermark 添加水印
func (p *FFmpegProcessor) AddWatermark(ctx context.Context, filePath string, watermarkPath string, outputPath string, position string) error {
	// 位置映射
	overlayMap := map[string]string{
		"top-left":      "10:10",
		"top-right":     "main_w-overlay_w-10:10",
		"bottom-left":   "10:main_h-overlay_h-10",
		"bottom-right":  "main_w-overlay_w-10:main_h-overlay_h-10",
		"center":        "(main_w-overlay_w)/2:(main_h-overlay_h)/2",
	}

	overlay := overlayMap[position]
	if overlay == "" {
		overlay = overlayMap["bottom-right"]
	}

	args := []string{
		"-y",
		"-i", filePath,
		"-i", watermarkPath,
		"-filter_complex", fmt.Sprintf("overlay=%s", overlay),
		"-c:a", "copy",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg watermark error: %s - %w", string(output), err)
	}

	return nil
}

// ParseTimeString 解析时间字符串
func ParseTimeString(timeStr string) float64 {
	// 支持格式: HH:MM:SS.mmm, MM:SS.mmm, SS.mmm
	re := regexp.MustCompile(`^(?:(\d+):)?(?:(\d+):)?(\d+(?:\.\d+)?)$`)
	matches := re.FindStringSubmatch(timeStr)

	if matches == nil {
		return 0
	}

	var hours, minutes float64
	seconds, _ := strconv.ParseFloat(matches[3], 64)

	if matches[2] != "" {
		minutes, _ = strconv.ParseFloat(matches[2], 64)
	}
	if matches[1] != "" {
		hours, _ = strconv.ParseFloat(matches[1], 64)
	}

	return hours*3600 + minutes*60 + seconds
}
