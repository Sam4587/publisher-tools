package ai

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MockClient 模拟 AI 客户端
type MockClient struct {
	ShouldFail bool
	Response   string
}

func (m *MockClient) GenerateText(ctx context.Context, prompt string, opts ...Option) (string, error) {
	if m.ShouldFail {
		return "", &Error{Code: ErrRateLimit, Message: "Mock error"}
	}
	return m.Response, nil
}

func (m *MockClient) Close() error {
	return nil
}

// TestNewUnifiedService 测试创建统一服务
func TestNewUnifiedService(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)
	if service == nil {
		t.Fatal("NewUnifiedService() returned nil")
	}
}

// TestGetDefaultClient 测试获取默认客户端
func TestGetDefaultClient(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)

	// 测试获取默认客户端（如果没有配置，应该返回错误）
	client, err := service.GetDefaultClient("text")
	if err == nil {
		t.Error("Expected error when no default config exists")
	}
	if client != nil {
		t.Error("Expected nil client when no default config exists")
	}
}

// TestGenerateText 测试生成文本
func TestGenerateText(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)

	// 添加测试配置
	config := &AIServiceConfig{
		ServiceType: "text",
		Name:        "Test Provider",
		Provider:    "test",
		BaseURL:     "https://test.com",
		APIKey:      "test-key",
		Model:       "test-model",
		Endpoint:    "/chat/completions",
		Priority:    100,
		IsDefault:   true,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = service.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// 测试生成文本
	ctx := context.Background()
	result, err := service.GenerateText(ctx, "Test prompt")

	// 由于没有真实的客户端实现，这里会失败
	if err == nil {
		t.Error("Expected error when no client is available")
	}
	if result != "" {
		t.Error("Expected empty result when no client is available")
	}
}

// TestGetActiveConfigs 测试获取活跃配置
func TestGetActiveConfigs(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)

	// 添加测试配置
	configs := []*AIServiceConfig{
		{
			ServiceType: "text",
			Name:        "Provider 1",
			Provider:    "provider1",
			BaseURL:     "https://provider1.com",
			APIKey:      "key1",
			Model:       "model1",
			Endpoint:    "/chat",
			Priority:    100,
			IsActive:    true,
		},
		{
			ServiceType: "text",
			Name:        "Provider 2",
			Provider:    "provider2",
			BaseURL:     "https://provider2.com",
			APIKey:      "key2",
			Model:       "model2",
			Endpoint:    "/chat",
			Priority:    90,
			IsActive:    false, // 不活跃
		},
		{
			ServiceType: "image",
			Name:        "Image Provider",
			Provider:    "image",
			BaseURL:     "https://image.com",
			APIKey:      "key3",
			Model:       "image-model",
			Endpoint:    "/generate",
			Priority:    100,
			IsActive:    true,
		},
	}

	for _, config := range configs {
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()
		err := service.CreateConfig(config)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}
	}

	// 获取活跃的文本配置
	textConfigs, err := service.GetActiveConfigs("text")
	if err != nil {
		t.Fatalf("GetActiveConfigs() error = %v", err)
	}

	// 应该只返回 1 个活跃的文本配置
	if len(textConfigs) != 1 {
		t.Errorf("Expected 1 active text config, got %d", len(textConfigs))
	}

	// 验证配置是按优先级排序的
	if textConfigs[0].Provider != "provider1" {
		t.Error("Configs not sorted by priority")
	}

	// 获取活跃的图片配置
	imageConfigs, err := service.GetActiveConfigs("image")
	if err != nil {
		t.Fatalf("GetActiveConfigs() error = %v", err)
	}

	if len(imageConfigs) != 1 {
		t.Errorf("Expected 1 active image config, got %d", len(imageConfigs))
	}
}

// TestCreateConfig 测试创建配置
func TestCreateConfig(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)

	// 创建配置
	config := &AIServiceConfig{
		ServiceType: "text",
		Name:        "Test Provider",
		Provider:    "test",
		BaseURL:     "https://test.com",
		APIKey:      "test-key",
		Model:       "test-model",
		Endpoint:    "/chat/completions",
		Priority:    100,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = service.CreateConfig(config)
	if err != nil {
		t.Fatalf("CreateConfig() error = %v", err)
	}

	// 验证配置已创建
	if config.ID == 0 {
		t.Error("Config ID was not set")
	}

	// 从数据库读取配置
	var retrieved AIServiceConfig
	err = db.First(&retrieved, config.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve config: %v", err)
	}

	if retrieved.Name != config.Name {
		t.Error("Config name mismatch")
	}
}

// TestUpdateConfig 测试更新配置
func TestUpdateConfig(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)

	// 创建配置
	config := &AIServiceConfig{
		ServiceType: "text",
		Name:        "Test Provider",
		Provider:    "test",
		BaseURL:     "https://test.com",
		APIKey:      "test-key",
		Model:       "test-model",
		Endpoint:    "/chat/completions",
		Priority:    100,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = service.CreateConfig(config)
	if err != nil {
		t.Fatalf("CreateConfig() error = %v", err)
	}

	// 更新配置
	config.Name = "Updated Provider"
	config.Priority = 90
	config.UpdatedAt = time.Now()

	err = service.UpdateConfig(config)
	if err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}

	// 验证配置已更新
	var retrieved AIServiceConfig
	err = db.First(&retrieved, config.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve config: %v", err)
	}

	if retrieved.Name != "Updated Provider" {
		t.Error("Config name was not updated")
	}

	if retrieved.Priority != 90 {
		t.Error("Config priority was not updated")
	}
}

// TestDeleteConfig 测试删除配置
func TestDeleteConfig(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)

	// 创建配置
	config := &AIServiceConfig{
		ServiceType: "text",
		Name:        "Test Provider",
		Provider:    "test",
		BaseURL:     "https://test.com",
		APIKey:      "test-key",
		Model:       "test-model",
		Endpoint:    "/chat/completions",
		Priority:    100,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = service.CreateConfig(config)
	if err != nil {
		t.Fatalf("CreateConfig() error = %v", err)
	}

	// 删除配置
	err = service.DeleteConfig(config.ID)
	if err != nil {
		t.Fatalf("DeleteConfig() error = %v", err)
	}

	// 验证配置已删除
	var retrieved AIServiceConfig
	err = db.First(&retrieved, config.ID).Error
	if err == nil {
		t.Error("Config was not deleted")
	}
}

// TestGetConfigStats 测试获取配置统计
func TestGetConfigStats(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建服务
	service := NewUnifiedService(db)

	// 添加多个配置
	for i := 0; i < 5; i++ {
		config := &AIServiceConfig{
			ServiceType: "text",
			Name:        "Test Provider",
			Provider:    "test",
			BaseURL:     "https://test.com",
			APIKey:      "test-key",
			Model:       "test-model",
			Endpoint:    "/chat/completions",
			Priority:    100,
			IsActive:    i%2 == 0, // 部分活跃
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := service.CreateConfig(config)
		if err != nil {
			t.Fatalf("CreateConfig() error = %v", err)
		}
	}

	// 获取统计信息
	stats, err := service.GetConfigStats()
	if err != nil {
		t.Fatalf("GetConfigStats() error = %v", err)
	}

	if stats.Total != 5 {
		t.Errorf("Expected 5 total configs, got %d", stats.Total)
	}

	if stats.Active != 3 {
		t.Errorf("Expected 3 active configs, got %d", stats.Active)
	}

	if stats.Inactive != 2 {
		t.Errorf("Expected 2 inactive configs, got %d", stats.Inactive)
	}
}
