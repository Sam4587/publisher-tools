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

	aiService := ai.NewServiceWithDefaults()
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

func (s *PublisherService) Login(platform string) (interface{}, error) {
	pub, err := s.factory.Create(platform, nil)
	if err != nil {
		return nil, err
	}

	result, err := pub.Login(context.Background())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PublisherService) CheckLogin(platform string) (interface{}, error) {
	pub, err := s.factory.Create(platform, nil)
	if err != nil {
		return nil, err
	}

	loggedIn, err := pub.CheckLoginStatus(context.Background())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"platform":  platform,
		"logged_in": loggedIn,
	}, nil
}

func (s *PublisherService) Logout(platform string) (interface{}, error) {
	pub, err := s.factory.Create(platform, nil)
	if err != nil {
		return nil, err
	}

	err = pub.Logout(context.Background())
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

func (s *StorageService) Upload(file []byte, path string) (string, error) {
	if err := s.storage.Write(context.Background(), path, file); err != nil {
		return "", err
	}
	return s.storage.GetURL(context.Background(), path)
}

func (s *StorageService) Download(path string) ([]byte, error) {
	return s.storage.Read(context.Background(), path)
}

func (s *StorageService) List(prefix string) ([]string, error) {
	return s.storage.List(context.Background(), prefix)
}

func (s *StorageService) Delete(path string) error {
	return s.storage.Delete(context.Background(), path)
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

func (a *AIServiceAdapter) Generate(providerName string, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
	return a.service.Generate(context.Background(), opts)
}

func (a *AIServiceAdapter) GenerateStream(providerName string, opts *provider.GenerateOptions) (<-chan string, error) {
	return a.service.GenerateStream(context.Background(), opts)
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

func (a *AIServiceAdapter) GenerateContent(prompt string, options map[string]interface{}) (interface{}, error) {
	opts := &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
	}
	return a.service.Generate(context.Background(), opts)
}

func (a *AIServiceAdapter) OptimizeTitle(title string, platform string) (string, error) {
	opts := &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: "You are a title optimization expert."},
			{Role: provider.RoleUser, Content: fmt.Sprintf("Optimize this title for %s platform: %s", platform, title)},
		},
	}
	result, err := a.service.Generate(context.Background(), opts)
	if err != nil {
		return title, err
	}
	return result.Content, nil
}

func (a *AIServiceAdapter) AnalyzeContent(content string) (interface{}, error) {
	opts := &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: "You are a content analysis expert."},
			{Role: provider.RoleUser, Content: fmt.Sprintf("Analyze this content: %s", content)},
		},
	}
	return a.service.Generate(context.Background(), opts)
}
