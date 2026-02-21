package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"publisher-core/video/asr"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ASRServiceAPI ASR服务接口
type ASRServiceAPI interface {
	Recognize(ctx context.Context, audioPath string, opts *asr.RecognizeOptions) (*asr.RecognitionResult, error)
	RecognizeWithChunking(ctx context.Context, audioPath string, opts *asr.RecognizeOptions) (*asr.RecognitionResult, error)
	RecognizeWithProvider(ctx context.Context, providerType asr.ProviderType, audioPath string, opts *asr.RecognizeOptions) (*asr.RecognitionResult, error)
	GetProviders() []asr.ProviderInfo
	GetStats() *asr.SelectorStats
	ClearCache() error
}

// ASRHandlers ASR处理器
type ASRHandlers struct {
	service    ASRServiceAPI
	uploadDir  string
}

// NewASRHandlers 创建ASR处理器
func NewASRHandlers(service ASRServiceAPI, uploadDir string) *ASRHandlers {
	if uploadDir == "" {
		uploadDir = "./uploads/audio"
	}
	os.MkdirAll(uploadDir, 0755)
	return &ASRHandlers{
		service:   service,
		uploadDir: uploadDir,
	}
}

// RegisterRoutes 注册路由
func (h *ASRHandlers) RegisterRoutes(router *mux.Router) {
	asrRouter := router.PathPrefix("/api/v1/asr").Subrouter()

	// 识别接口
	asrRouter.HandleFunc("/recognize", h.recognize).Methods("POST")
	asrRouter.HandleFunc("/recognize/upload", h.recognizeUpload).Methods("POST")
	asrRouter.HandleFunc("/recognize/chunked", h.recognizeChunked).Methods("POST")

	// 提供商管理
	asrRouter.HandleFunc("/providers", h.getProviders).Methods("GET")
	asrRouter.HandleFunc("/providers/{provider}/recognize", h.recognizeWithProvider).Methods("POST")

	// 统计和缓存
	asrRouter.HandleFunc("/stats", h.getStats).Methods("GET")
	asrRouter.HandleFunc("/cache/clear", h.clearCache).Methods("POST")

	// 支持的语言和模型
	asrRouter.HandleFunc("/languages", h.getLanguages).Methods("GET")
	asrRouter.HandleFunc("/models", h.getModels).Methods("GET")
}

// recognizeRequest 识别请求
type recognizeRequest struct {
	AudioPath       string  `json:"audio_path"`
	Language        string  `json:"language"`
	Model           string  `json:"model"`
	EnableTimestamps bool   `json:"enable_timestamps"`
	EnableWordLevel bool   `json:"enable_word_level"`
	EnableGPU       bool   `json:"enable_gpu"`
	Temperature     float64 `json:"temperature"`
	InitialPrompt   string  `json:"initial_prompt"`
	Timeout         int     `json:"timeout"` // 秒
}

// recognize 识别音频
func (h *ASRHandlers) recognize(w http.ResponseWriter, r *http.Request) {
	var req recognizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.AudioPath == "" {
		jsonError(w, "MISSING_AUDIO_PATH", "audio_path is required", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(req.AudioPath); os.IsNotExist(err) {
		jsonError(w, "FILE_NOT_FOUND", "audio file not found", http.StatusNotFound)
		return
	}

	// 构建选项
	opts := h.buildOptions(&req)

	// 执行识别
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(req.Timeout)*time.Second)
	if req.Timeout == 0 {
		ctx, cancel = context.WithTimeout(r.Context(), 30*time.Minute)
	}
	defer cancel()

	result, err := h.service.Recognize(ctx, req.AudioPath, opts)
	if err != nil {
		jsonError(w, "RECOGNITION_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// recognizeUpload 上传并识别
func (h *ASRHandlers) recognizeUpload(w http.ResponseWriter, r *http.Request) {
	// 解析multipart表单
	maxSize := int64(500 * 1024 * 1024) // 500MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, "PARSE_FORM_FAILED", err.Error(), http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("audio")
	if err != nil {
		jsonError(w, "MISSING_AUDIO_FILE", "audio file is required", http.StatusBadRequest)
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

	if _, err := dst.ReadFrom(file); err != nil {
		os.Remove(filePath)
		jsonError(w, "SAVE_FILE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建选项
	req := &recognizeRequest{
		Language:         r.FormValue("language"),
		Model:            r.FormValue("model"),
		EnableTimestamps: r.FormValue("enable_timestamps") == "true",
		EnableWordLevel:  r.FormValue("enable_word_level") == "true",
		EnableGPU:        r.FormValue("enable_gpu") == "true",
		InitialPrompt:    r.FormValue("initial_prompt"),
	}
	fmt.Sscanf(r.FormValue("temperature"), "%f", &req.Temperature)
	fmt.Sscanf(r.FormValue("timeout"), "%d", &req.Timeout)

	opts := h.buildOptions(req)

	// 执行识别
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	result, err := h.service.Recognize(ctx, filePath, opts)
	if err != nil {
		os.Remove(filePath)
		jsonError(w, "RECOGNITION_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 可选：删除临时文件
	if r.FormValue("keep_file") != "true" {
		os.Remove(filePath)
	}

	jsonSuccess(w, result)
}

// recognizeChunked 分片识别
func (h *ASRHandlers) recognizeChunked(w http.ResponseWriter, r *http.Request) {
	var req recognizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.AudioPath == "" {
		jsonError(w, "MISSING_AUDIO_PATH", "audio_path is required", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(req.AudioPath); os.IsNotExist(err) {
		jsonError(w, "FILE_NOT_FOUND", "audio file not found", http.StatusNotFound)
		return
	}

	// 构建选项
	opts := h.buildOptions(&req)

	// 执行分片识别
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Minute)
	defer cancel()

	result, err := h.service.RecognizeWithChunking(ctx, req.AudioPath, opts)
	if err != nil {
		jsonError(w, "RECOGNITION_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// recognizeWithProvider 使用指定提供商识别
func (h *ASRHandlers) recognizeWithProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerType := asr.ProviderType(vars["provider"])

	var req recognizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.AudioPath == "" {
		jsonError(w, "MISSING_AUDIO_PATH", "audio_path is required", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(req.AudioPath); os.IsNotExist(err) {
		jsonError(w, "FILE_NOT_FOUND", "audio file not found", http.StatusNotFound)
		return
	}

	// 构建选项
	opts := h.buildOptions(&req)

	// 执行识别
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	result, err := h.service.RecognizeWithProvider(ctx, providerType, req.AudioPath, opts)
	if err != nil {
		jsonError(w, "RECOGNITION_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// getProviders 获取提供商列表
func (h *ASRHandlers) getProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.service.GetProviders()
	jsonSuccess(w, providers)
}

// getStats 获取统计信息
func (h *ASRHandlers) getStats(w http.ResponseWriter, r *http.Request) {
	stats := h.service.GetStats()
	jsonSuccess(w, stats)
}

// clearCache 清除缓存
func (h *ASRHandlers) clearCache(w http.ResponseWriter, r *http.Request) {
	if err := h.service.ClearCache(); err != nil {
		jsonError(w, "CLEAR_CACHE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	jsonSuccess(w, map[string]string{"message": "Cache cleared successfully"})
}

// getLanguages 获取支持的语言
func (h *ASRHandlers) getLanguages(w http.ResponseWriter, r *http.Request) {
	languages := asr.GetSupportedLanguages()
	jsonSuccess(w, languages)
}

// getModels 获取可用模型
func (h *ASRHandlers) getModels(w http.ResponseWriter, r *http.Request) {
	models := asr.GetAvailableModels()
	modelInfos := make([]map[string]interface{}, len(models))
	for i, model := range models {
		modelInfos[i] = map[string]interface{}{
			"name": model,
			"info": asr.GetModelInfo(model),
		}
	}
	jsonSuccess(w, modelInfos)
}

// buildOptions 构建识别选项
func (h *ASRHandlers) buildOptions(req *recognizeRequest) *asr.RecognizeOptions {
	opts := asr.DefaultRecognizeOptions()

	if req.Language != "" {
		opts.Language = req.Language
	}
	if req.Model != "" {
		opts.Model = req.Model
	}
	opts.EnableTimestamps = req.EnableTimestamps
	opts.EnableWordLevel = req.EnableWordLevel
	opts.EnableGPU = req.EnableGPU
	if req.Temperature > 0 {
		opts.Temperature = req.Temperature
	}
	if req.InitialPrompt != "" {
		opts.InitialPrompt = req.InitialPrompt
	}
	if req.Timeout > 0 {
		opts.Timeout = time.Duration(req.Timeout) * time.Second
	}

	return opts
}

// jsonError 返回JSON错误
func jsonError(w http.ResponseWriter, code string, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// ASRService ASR服务实现
type ASRService struct {
	selector *asr.Selector
}

// NewASRService 创建ASR服务
func NewASRService(selector *asr.Selector) *ASRService {
	return &ASRService{selector: selector}
}

// Recognize 执行识别
func (s *ASRService) Recognize(ctx context.Context, audioPath string, opts *asr.RecognizeOptions) (*asr.RecognitionResult, error) {
	return s.selector.Recognize(ctx, audioPath, opts)
}

// RecognizeWithChunking 分片识别
func (s *ASRService) RecognizeWithChunking(ctx context.Context, audioPath string, opts *asr.RecognizeOptions) (*asr.RecognitionResult, error) {
	return s.selector.RecognizeWithChunking(ctx, audioPath, opts)
}

// RecognizeWithProvider 使用指定提供商识别
func (s *ASRService) RecognizeWithProvider(ctx context.Context, providerType asr.ProviderType, audioPath string, opts *asr.RecognizeOptions) (*asr.RecognitionResult, error) {
	return s.selector.RecognizeWithProvider(ctx, providerType, audioPath, opts)
}

// GetProviders 获取提供商列表
func (s *ASRService) GetProviders() []asr.ProviderInfo {
	return s.selector.GetProviders()
}

// GetStats 获取统计信息
func (s *ASRService) GetStats() *asr.SelectorStats {
	return s.selector.GetStats()
}

// ClearCache 清除缓存
func (s *ASRService) ClearCache() error {
	return s.selector.ClearCache()
}

// SetupASR 设置ASR服务
func SetupASR(config *ASRConfig) (*ASRService, error) {
	// 创建选择器
	selector := asr.NewSelector(&asr.SelectorConfig{
		CacheEnabled: config.CacheEnabled,
		CacheDir:     config.CacheDir,
		CacheTTL:     config.CacheTTL,
	})

	// 注册BcutASR提供商
	if config.EnableBcutASR {
		bcutProvider := asr.NewBcutASRProvider(&asr.BcutASRConfig{
			APIKey:     config.BcutAPIKey,
			APIURL:     config.BcutAPIURL,
			Timeout:    config.BcutTimeout,
			MaxRetries: config.BcutMaxRetries,
			Priority:   1,
		})
		selector.RegisterProvider(bcutProvider)
		logrus.Info("BcutASR provider registered")
	}

	// 注册Whisper提供商
	if config.EnableWhisper {
		whisperProvider := asr.NewWhisperProvider(&asr.WhisperConfig{
			WhisperPath: config.WhisperPath,
			Model:       config.WhisperModel,
			OutputDir:   config.WhisperOutputDir,
			EnableGPU:   config.WhisperEnableGPU,
			Timeout:     config.WhisperTimeout,
			Priority:    2,
		})
		selector.RegisterProvider(whisperProvider)
		logrus.Info("Whisper provider registered")
	}

	return NewASRService(selector), nil
}

// ASRConfig ASR配置
type ASRConfig struct {
	// 缓存配置
	CacheEnabled bool          `json:"cache_enabled"`
	CacheDir     string        `json:"cache_dir"`
	CacheTTL     time.Duration `json:"cache_ttl"`

	// BcutASR配置
	EnableBcutASR  bool          `json:"enable_bcut_asr"`
	BcutAPIKey     string        `json:"bcut_api_key"`
	BcutAPIURL     string        `json:"bcut_api_url"`
	BcutTimeout    time.Duration `json:"bcut_timeout"`
	BcutMaxRetries int           `json:"bcut_max_retries"`

	// Whisper配置
	EnableWhisper     bool          `json:"enable_whisper"`
	WhisperPath       string        `json:"whisper_path"`
	WhisperModel      string        `json:"whisper_model"`
	WhisperOutputDir  string        `json:"whisper_output_dir"`
	WhisperEnableGPU  bool          `json:"whisper_enable_gpu"`
	WhisperTimeout    time.Duration `json:"whisper_timeout"`
}

// DefaultASRConfig 默认ASR配置
func DefaultASRConfig() *ASRConfig {
	return &ASRConfig{
		CacheEnabled:     true,
		CacheDir:         "./data/asr_cache",
		CacheTTL:         24 * time.Hour,
		EnableBcutASR:    true,
		BcutTimeout:      10 * time.Minute,
		BcutMaxRetries:   3,
		EnableWhisper:    true,
		WhisperPath:      "whisper",
		WhisperModel:     "base",
		WhisperOutputDir: "./data/transcripts",
		WhisperEnableGPU: false,
		WhisperTimeout:   30 * time.Minute,
	}
}
