package asr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ChunkProcessor 分片处理器
type ChunkProcessor struct {
	ffmpegPath   string
	chunkDuration float64  // 每个分片的时长(秒)
	maxChunkSize  int64    // 最大分片大小(字节)
	tempDir       string
	parallelJobs  int      // 并行处理数
}

// ChunkProcessorConfig 分片处理器配置
type ChunkProcessorConfig struct {
	FFmpegPath    string  `json:"ffmpeg_path"`
	ChunkDuration float64 `json:"chunk_duration"` // 默认300秒(5分钟)
	MaxChunkSize  int64   `json:"max_chunk_size"` // 默认50MB
	TempDir       string  `json:"temp_dir"`
	ParallelJobs  int     `json:"parallel_jobs"`  // 默认2
}

// DefaultChunkProcessorConfig 默认配置
func DefaultChunkProcessorConfig() *ChunkProcessorConfig {
	return &ChunkProcessorConfig{
		FFmpegPath:    "ffmpeg",
		ChunkDuration: 300, // 5分钟
		MaxChunkSize:  50 * 1024 * 1024, // 50MB
		TempDir:       "./data/temp/chunks",
		ParallelJobs:  2,
	}
}

// NewChunkProcessor 创建分片处理器
func NewChunkProcessor(config *ChunkProcessorConfig) *ChunkProcessor {
	if config == nil {
		config = DefaultChunkProcessorConfig()
	}

	os.MkdirAll(config.TempDir, 0755)

	return &ChunkProcessor{
		ffmpegPath:    config.FFmpegPath,
		chunkDuration: config.ChunkDuration,
		maxChunkSize:  config.MaxChunkSize,
		tempDir:       config.TempDir,
		parallelJobs:  config.ParallelJobs,
	}
}

// ChunkInfo 分片信息
type ChunkInfo struct {
	Index     int     `json:"index"`
	Path      string  `json:"path"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Duration  float64 `json:"duration"`
	Size      int64   `json:"size"`
}

// SplitAudio 分割音频文件
func (p *ChunkProcessor) SplitAudio(ctx context.Context, audioPath string) ([]ChunkInfo, error) {
	// 获取音频信息
	info, err := p.getAudioInfo(audioPath)
	if err != nil {
		return nil, fmt.Errorf("get audio info: %w", err)
	}

	// 如果文件不大，不需要分片
	if info.Size < p.maxChunkSize && info.Duration < p.chunkDuration {
		return []ChunkInfo{{
			Index:     0,
			Path:      audioPath,
			StartTime: 0,
			EndTime:   info.Duration,
			Duration:  info.Duration,
			Size:      info.Size,
		}}, nil
	}

	logrus.Infof("Splitting large audio file: %s (duration: %.1fs, size: %dMB)",
		filepath.Base(audioPath), info.Duration, info.Size/1024/1024)

	// 计算分片数量
	numChunks := int(info.Duration/p.chunkDuration) + 1
	chunks := make([]ChunkInfo, 0, numChunks)

	// 创建临时目录
	sessionDir := filepath.Join(p.tempDir, fmt.Sprintf("chunk_%d", time.Now().UnixNano()))
	os.MkdirAll(sessionDir, 0755)
	defer os.RemoveAll(sessionDir)

	// 分割音频
	for i := 0; i < numChunks; i++ {
		startTime := float64(i) * p.chunkDuration
		endTime := startTime + p.chunkDuration
		if endTime > info.Duration {
			endTime = info.Duration
		}

		if startTime >= info.Duration {
			break
		}

		chunkPath := filepath.Join(sessionDir, fmt.Sprintf("chunk_%04d.mp3", i))

		// 使用ffmpeg分割
		args := []string{
			"-i", audioPath,
			"-ss", fmt.Sprintf("%.3f", startTime),
			"-t", fmt.Sprintf("%.3f", endTime-startTime),
			"-c", "copy",
			"-y",
			chunkPath,
		}

		cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			logrus.Warnf("Failed to create chunk %d: %v - %s", i, err, string(output))
			continue
		}

		// 检查分片文件
		chunkInfo, err := os.Stat(chunkPath)
		if err != nil {
			continue
		}

		chunks = append(chunks, ChunkInfo{
			Index:     i,
			Path:      chunkPath,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime - startTime,
			Size:      chunkInfo.Size(),
		})

		logrus.Debugf("Created chunk %d: %.1fs - %.1fs (%dMB)",
			i, startTime, endTime, chunkInfo.Size()/1024/1024)
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks created")
	}

	return chunks, nil
}

// AudioInfo 音频信息
type AudioInfo struct {
	Duration   float64
	Size       int64
	Format     string
	SampleRate int
	Channels   int
}

// getAudioInfo 获取音频信息
func (p *ChunkProcessor) getAudioInfo(audioPath string) (*AudioInfo, error) {
	// 获取文件大小
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return nil, err
	}

	// 使用ffprobe获取时长
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	}

	cmd := exec.Command("ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		// 如果ffprobe失败，使用默认值
		return &AudioInfo{
			Size: fileInfo.Size(),
		}, nil
	}

	var duration float64
	fmt.Sscanf(string(output), "%f", &duration)

	return &AudioInfo{
		Duration: duration,
		Size:     fileInfo.Size(),
	}, nil
}

// ProcessChunks 并行处理分片
func (p *ChunkProcessor) ProcessChunks(ctx context.Context, chunks []ChunkInfo, processor func(ctx context.Context, chunk ChunkInfo) (*RecognitionResult, error)) ([]*RecognitionResult, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks to process")
	}

	// 如果只有一个分片，直接处理
	if len(chunks) == 1 {
		result, err := processor(ctx, chunks[0])
		if err != nil {
			return nil, err
		}
		return []*RecognitionResult{result}, nil
	}

	// 并行处理
	results := make([]*RecognitionResult, len(chunks))
	errors := make([]error, len(chunks))

	var wg sync.WaitGroup
	sem := make(chan struct{}, p.parallelJobs)

	for i, chunk := range chunks {
		wg.Add(1)
		go func(idx int, c ChunkInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := processor(ctx, c)
			results[idx] = result
			errors[idx] = err
		}(i, chunk)
	}

	wg.Wait()

	// 检查错误
	var hasError bool
	for i, err := range errors {
		if err != nil {
			logrus.Warnf("Chunk %d processing failed: %v", i, err)
			hasError = true
		}
	}

	if hasError {
		// 如果有部分成功，继续合并
		successCount := 0
		for _, r := range results {
			if r != nil {
				successCount++
			}
		}
		if successCount == 0 {
			return nil, fmt.Errorf("all chunks failed")
		}
	}

	return results, nil
}

// MergeResults 合并识别结果
func (p *ChunkProcessor) MergeResults(chunks []ChunkInfo, results []*RecognitionResult) *RecognitionResult {
	if len(results) == 0 {
		return nil
	}

	// 过滤有效结果
	var validResults []*RecognitionResult
	var validChunks []ChunkInfo
	for i, r := range results {
		if r != nil {
			validResults = append(validResults, r)
			validChunks = append(validChunks, chunks[i])
		}
	}

	if len(validResults) == 0 {
		return nil
	}

	// 如果只有一个有效结果
	if len(validResults) == 1 {
		return validResults[0]
	}

	// 合并结果
	merged := &RecognitionResult{
		Provider: validResults[0].Provider,
		Language: validResults[0].Language,
	}

	var allTexts []string
	var totalDuration time.Duration
	var totalAudioDuration float64
	var totalWords int
	var totalConfidence float64
	var confidenceCount int

	for i, result := range validResults {
		chunk := validChunks[i]
		timeOffset := chunk.StartTime

		// 合并文本
		if result.Text != "" {
			allTexts = append(allTexts, result.Text)
		}

		// 合并片段（调整时间戳）
		for _, seg := range result.Segments {
			merged.Segments = append(merged.Segments, RecognitionSegment{
				ID:         len(merged.Segments),
				Start:      seg.Start + timeOffset,
				End:        seg.End + timeOffset,
				Text:       seg.Text,
				Confidence: seg.Confidence,
				Words:      p.adjustWordTimestamps(seg.Words, timeOffset),
			})

			if seg.Confidence > 0 {
				totalConfidence += seg.Confidence
				confidenceCount++
			}
		}

		totalDuration += result.Duration
		totalAudioDuration += chunk.Duration
		totalWords += result.WordCount
	}

	merged.Text = strings.Join(allTexts, " ")
	merged.Duration = totalDuration
	merged.AudioDuration = totalAudioDuration
	merged.WordCount = totalWords

	// 计算质量评分
	merged.QualityScore = CalculateQualityScore(merged)

	return merged
}

// adjustWordTimestamps 调整词级时间戳
func (p *ChunkProcessor) adjustWordTimestamps(words []Word, offset float64) []Word {
	if len(words) == 0 {
		return nil
	}

	adjusted := make([]Word, len(words))
	for i, w := range words {
		adjusted[i] = Word{
			Word:       w.Word,
			Start:      w.Start + offset,
			End:        w.End + offset,
			Confidence: w.Confidence,
		}
	}
	return adjusted
}

// RecognizeWithChunking 带分片的识别
func (s *Selector) RecognizeWithChunking(ctx context.Context, audioPath string, opts *RecognizeOptions) (*RecognitionResult, error) {
	if opts == nil {
		opts = DefaultRecognizeOptions()
	}

	// 创建分片处理器
	chunkProcessor := NewChunkProcessor(nil)

	// 分割音频
	chunks, err := chunkProcessor.SplitAudio(ctx, audioPath)
	if err != nil {
		return nil, fmt.Errorf("split audio: %w", err)
	}

	// 如果只有一个分片，使用普通识别
	if len(chunks) == 1 {
		return s.Recognize(ctx, audioPath, opts)
	}

	logrus.Infof("Processing %d audio chunks in parallel", len(chunks))

	// 并行处理分片
	results, err := chunkProcessor.ProcessChunks(ctx, chunks, func(ctx context.Context, chunk ChunkInfo) (*RecognitionResult, error) {
		return s.Recognize(ctx, chunk.Path, opts)
	})

	if err != nil {
		return nil, err
	}

	// 合并结果
	merged := chunkProcessor.MergeResults(chunks, results)
	if merged == nil {
		return nil, fmt.Errorf("failed to merge results")
	}

	// 添加元数据
	if merged.Metadata == nil {
		merged.Metadata = make(map[string]interface{})
	}
	merged.Metadata["chunk_count"] = len(chunks)
	merged.Metadata["chunked"] = true

	return merged, nil
}

// SortSegmentsByTime 按时间排序片段
func SortSegmentsByTime(segments []RecognitionSegment) []RecognitionSegment {
	sorted := make([]RecognitionSegment, len(segments))
	copy(sorted, segments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})
	return sorted
}
