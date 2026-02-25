package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"publisher-core/adapters"
	"publisher-core/ai"
	"publisher-core/ai/provider"
	"publisher-core/analytics"
	"publisher-core/analytics/collectors"
	"publisher-core/api"
	"publisher-core/hotspot"
	"publisher-core/hotspot/sources"
	"publisher-core/storage"
	"publisher-core/task"
	"publisher-core/task/handlers"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	port       int
	headless   bool
	cookieDir  string
	storageDir string
	dataDir    string
	baseURL    string
	debug      bool
)

func init() {
	flag.IntVar(&port, "port", 8080, "API server port")
	flag.BoolVar(&headless, "headless", true, "Browser headless mode")
	flag.StringVar(&cookieDir, "cookie-dir", "./cookies", "Cookie storage directory")
	flag.StringVar(&storageDir, "storage-dir", "./uploads", "File storage directory")
	flag.StringVar(&dataDir, "data-dir", "./data", "Data storage directory")
	flag.StringVar(&baseURL, "base-url", "", "File access base URL")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
}

func main() {
	flag.Parse()

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	setupLogger()

	store, err := storage.NewLocalStorage(storageDir, baseURL)
	if err != nil {
		logrus.Fatalf("Failed to create storage: %v", err)
	}

	taskMgr := task.NewTaskManager(task.NewMemoryStorage())

	factory := adapters.DefaultFactory()

	publishHandler := handlers.NewPublishHandler(factory)
	taskMgr.RegisterHandler("publish", publishHandler.Handle)

	publisherService := &PublisherService{
		factory: factory,
		taskMgr: taskMgr,
	}
	storageService := &StorageService{
		storage: store,
	}
	taskService := &TaskService{
		taskMgr: taskMgr,
	}

	// 初始化 AI 服务并从环境变量加载配置
	aiService := ai.NewServiceWithDefaults()
	setupAIProviders(aiService)
	aiAdapter := &AIServiceAdapter{service: aiService}

	server := api.NewServer(taskService, publisherService, storageService, aiAdapter)

	hotspotStorage, err := hotspot.NewJSONStorage(dataDir)
	if err != nil {
		logrus.Fatalf("Failed to create hotspot storage: %v", err)
	}
	hotspotService := hotspot.NewService(hotspotStorage)

	for _, src := range sources.CreateAllSources() {
		hotspotService.RegisterSource(src)
	}

	hotspotService.RegisterSource(sources.NewMockSource("mock", "Test Source"))

	// 注册热点监控API路由
	hotspotAPI := hotspot.NewAPIHandler(hotspotService)
	server.RegisterRoutes(hotspotAPI)

	// 注册存储API路由
	storageHandlers := api.NewStorageHandlers(store, storageDir)
	storageHandlers.RegisterRoutes(server.Router())

	analyticsStorage, err := analytics.NewJSONStorage(dataDir + "/analytics")
	if err != nil {
		logrus.Warnf("Failed to create analytics storage: %v", err)
	}
	analyticsService := analytics.NewService(analyticsStorage)

	analyticsService.RegisterCollector(collectors.NewDouyinCollector())
	analyticsService.RegisterCollector(collectors.NewXiaohongshuCollector())
	analyticsService.RegisterCollector(collectors.NewToutiaoCollector())

	go func() {
		addr := fmt.Sprintf(":%d", port)
		if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	logrus.Info("Publisher service started")
	logrus.Infof("API address: http://localhost:%d", port)
	logrus.Infof("Supported platforms: %v", factory.Platforms())
	logrus.Info("Hotspot service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		logrus.Errorf("Failed to shutdown server: %v", err)
	}

	select {
	case <-ctx.Done():
		logrus.Warn("Service shutdown timeout")
	default:
		logrus.Info("Service stopped")
	}
}

func setupLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// setupAIProviders 从环境变量加载 AI 提供商配置
func setupAIProviders(aiService *ai.Service) {
	// OpenRouter
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		aiService.RegisterProvider(provider.NewOpenRouterProvider(apiKey))
		logrus.Info("OpenRouter provider registered")
	}

	// Google
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		aiService.RegisterProvider(provider.NewGoogleProvider(apiKey))
		logrus.Info("Google provider registered")
	}

	// Groq
	if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
		aiService.RegisterProvider(provider.NewGroqProvider(apiKey))
		logrus.Info("Groq provider registered")
	}

	// DeepSeek
	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		baseURL := os.Getenv("DEEPSEEK_BASE_URL")
		p := provider.NewDeepSeekProviderWithBaseURL(apiKey, baseURL)
		aiService.RegisterProvider(p)
		logrus.Info("DeepSeek provider registered")
	}

	// Ollama
	if apiKey := os.Getenv("OLLAMA_API_KEY"); apiKey != "" {
		baseURL := os.Getenv("OLLAMA_BASE_URL")
		model := os.Getenv("OLLAMA_MODEL")
		p, err := provider.NewOllamaProvider(apiKey, baseURL, model)
		if err != nil {
			logrus.Warnf("Failed to create Ollama provider: %v", err)
		} else {
			aiService.RegisterProvider(p)
			logrus.Info("Ollama provider registered")
		}
	}

	// 设置默认提供商
	if defaultProvider := os.Getenv("AI_DEFAULT_PROVIDER"); defaultProvider != "" {
		aiService.SetPrimary(provider.ProviderType(defaultProvider))
	}
}

type PublisherService struct {
	factory *adapters.PublisherFactory
	taskMgr *task.TaskManager
}

func (s *PublisherService) GetPlatforms() []string {
	return s.factory.Platforms()
}

func (s *PublisherService) GetPlatformInfo(platform string) (interface{}, error) {
	pub, err := s.factory.Create(platform, nil)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"platform": pub.Platform(),
		"message":  fmt.Sprintf("Platform %s is ready", platform),
	}, nil
}

func (s *PublisherService) Login(ctx context.Context, platform string) (interface{}, error) {
	pub, err := s.factory.Create(platform, nil)
	if err != nil {
		return nil, err
	}

	result, err := pub.Login(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PublisherService) CheckLogin(ctx context.Context, platform string) (interface{}, error) {
	pub, err := s.factory.Create(platform, nil)
	if err != nil {
		return nil, err
	}

	loggedIn, err := pub.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"platform":  platform,
		"logged_in": loggedIn,
	}, nil
}

func (s *PublisherService) Logout(ctx context.Context, platform string) (interface{}, error) {
	pub, err := s.factory.Create(platform, nil)
	if err != nil {
		return nil, err
	}

	err = pub.Logout(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"platform": platform,
		"message":  "Logout success",
	}, nil
}

type StorageService struct {
	storage storage.Storage
}

func (s *StorageService) Upload(ctx context.Context, file []byte, path string) (string, error) {
	if err := s.storage.Write(ctx, path, file); err != nil {
		return "", err
	}
	return s.storage.GetURL(ctx, path)
}

func (s *StorageService) Download(ctx context.Context, path string) ([]byte, error) {
	return s.storage.Read(ctx, path)
}

func (s *StorageService) List(ctx context.Context, prefix string) ([]string, error) {
	return s.storage.List(ctx, prefix)
}

func (s *StorageService) Delete(ctx context.Context, path string) error {
	return s.storage.Delete(ctx, path)
}

type TaskService struct {
	taskMgr *task.TaskManager
}

func (s *TaskService) CreateTask(taskType string, platform string, payload map[string]interface{}) (interface{}, error) {
	t, err := s.taskMgr.CreateTask(taskType, platform, payload)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":         t.ID,
		"type":       t.Type,
		"status":     t.Status,
		"platform":   t.Platform,
		"payload":    t.Payload,
		"progress":   t.Progress,
		"created_at": t.CreatedAt,
	}, nil
}

func (s *TaskService) GetTask(taskID string) (interface{}, error) {
	return s.taskMgr.GetTask(taskID)
}

func (s *TaskService) ListTasks(status string, platform string, limit int) (interface{}, error) {
	filter := task.TaskFilter{
		Status:   task.TaskStatus(status),
		Platform: platform,
		Limit:    limit,
	}
	return s.taskMgr.ListTasks(filter)
}

func (s *TaskService) CancelTask(taskID string) error {
	return s.taskMgr.Cancel(taskID)
}

type AIServiceAdapter struct {
	service *ai.Service
}

func (a *AIServiceAdapter) Generate(ctx context.Context, providerName string, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
	return a.service.Generate(ctx, opts)
}

func (a *AIServiceAdapter) GenerateStream(ctx context.Context, providerName string, opts *provider.GenerateOptions) (<-chan string, error) {
	return a.service.GenerateStream(ctx, opts)
}

func (a *AIServiceAdapter) ListProviders() []string {
	providers := a.service.ListProviders()
	if len(providers) == 0 {
		return []string{"none"}
	}
	result := make([]string, len(providers))
	for i, p := range providers {
		result[i] = string(p)
	}
	return result
}

func (a *AIServiceAdapter) ListModels() map[string][]string {
	return a.service.ListModels()
}

func (a *AIServiceAdapter) GenerateContent(ctx context.Context, prompt string, options map[string]interface{}) (interface{}, error) {
	opts := &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
	}
	return a.service.Generate(ctx, opts)
}

func (a *AIServiceAdapter) OptimizeTitle(ctx context.Context, title string, platform string) (string, error) {
	opts := &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: "You are a title optimization expert."},
			{Role: provider.RoleUser, Content: fmt.Sprintf("Optimize this title for %s platform: %s", platform, title)},
		},
	}
	result, err := a.service.Generate(ctx, opts)
	if err != nil {
		return title, err
	}
	return result.Content, nil
}

func (a *AIServiceAdapter) AnalyzeContent(ctx context.Context, content string) (interface{}, error) {
	opts := &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: "You are a content analysis expert."},
			{Role: provider.RoleUser, Content: fmt.Sprintf("Analyze this content: %s", content)},
		},
	}
	return a.service.Generate(ctx, opts)
}
