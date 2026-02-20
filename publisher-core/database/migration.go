package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MigrationConfig 迁移配置
type MigrationConfig struct {
	// JSON 数据目录
	DataDir string
	// 是否删除原 JSON 文件
	DeleteOriginal bool
	// 是否备份原 JSON 文件
	BackupOriginal bool
}

// DefaultMigrationConfig 返回默认迁移配置
func DefaultMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		DataDir:        "./data",
		DeleteOriginal: false,
		BackupOriginal: true,
	}
}

// MigrateFromJSON 从 JSON 文件迁移数据到 SQLite
func MigrateFromJSON(db *gorm.DB, cfg *MigrationConfig) error {
	if cfg == nil {
		cfg = DefaultMigrationConfig()
	}

	logrus.Info("Starting data migration from JSON to SQLite...")

	// 迁移热点话题数据
	if err := migrateHotspotTopics(db, cfg); err != nil {
		logrus.Warnf("Failed to migrate hotspot topics: %v", err)
	}

	// 迁移 AI 配置数据
	if err := migrateAIConfig(db, cfg); err != nil {
		logrus.Warnf("Failed to migrate AI config: %v", err)
	}

	// 迁移 Cookie 数据
	if err := migrateCookies(db, cfg); err != nil {
		logrus.Warnf("Failed to migrate cookies: %v", err)
	}

	logrus.Info("Data migration completed")
	return nil
}

// migrateHotspotTopics 迁移热点话题数据
func migrateHotspotTopics(db *gorm.DB, cfg *MigrationConfig) error {
	jsonPath := filepath.Join(cfg.DataDir, "hotspot_topics.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		logrus.Debug("No hotspot topics JSON file found, skipping")
		return nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read hotspot topics file: %w", err)
	}

	// 定义 JSON 话题结构（与原 hotspot.Topic 兼容）
	var jsonTopics []struct {
		ID          string    `json:"_id"`
		Title       string    `json:"title"`
		Description string    `json:"description,omitempty"`
		Category    string    `json:"category"`
		Heat        int       `json:"heat"`
		Trend       string    `json:"trend"`
		Source      string    `json:"source"`
		SourceID    string    `json:"sourceId,omitempty"`
		SourceURL   string    `json:"sourceUrl,omitempty"`
		OriginalURL string    `json:"originalUrl,omitempty"`
		Keywords    []string  `json:"keywords,omitempty"`
		Suitability int       `json:"suitability,omitempty"`
		PublishedAt time.Time `json:"publishedAt,omitempty"`
		CreatedAt   time.Time `json:"createdAt,omitempty"`
		UpdatedAt   time.Time `json:"updatedAt,omitempty"`
	}

	if err := json.Unmarshal(data, &jsonTopics); err != nil {
		return fmt.Errorf("parse hotspot topics: %w", err)
	}

	if len(jsonTopics) == 0 {
		logrus.Debug("No topics to migrate")
		return nil
	}

	// 转换为数据库模型
	var topics []Topic
	for _, jt := range jsonTopics {
		keywordsJSON, _ := json.Marshal(jt.Keywords)

		topic := Topic{
			ID:          jt.ID,
			Title:       jt.Title,
			Description: jt.Description,
			Category:    jt.Category,
			Heat:        jt.Heat,
			Trend:       jt.Trend,
			Source:      jt.Source,
			SourceID:    jt.SourceID,
			SourceURL:   jt.SourceURL,
			OriginalURL: jt.OriginalURL,
			Keywords:    string(keywordsJSON),
			Suitability: jt.Suitability,
			PublishedAt: jt.PublishedAt,
			CreatedAt:   jt.CreatedAt,
			UpdatedAt:   jt.UpdatedAt,
		}

		if topic.CreatedAt.IsZero() {
			topic.CreatedAt = time.Now()
		}
		if topic.UpdatedAt.IsZero() {
			topic.UpdatedAt = time.Now()
		}

		topics = append(topics, topic)
	}

	// 批量插入（忽略重复）
	result := db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(topics, 100)

	if result.Error != nil {
		return fmt.Errorf("insert topics: %w", result.Error)
	}

	logrus.Infof("Migrated %d topics from JSON", len(topics))

	// 备份原文件
	if cfg.BackupOriginal {
		backupPath := jsonPath + ".backup"
		if err := os.Rename(jsonPath, backupPath); err != nil {
			logrus.Warnf("Failed to backup original file: %v", err)
		} else {
			logrus.Infof("Backed up original file to %s", backupPath)
		}
	} else if cfg.DeleteOriginal {
		if err := os.Remove(jsonPath); err != nil {
			logrus.Warnf("Failed to delete original file: %v", err)
		}
	}

	return nil
}

// migrateAIConfig 迁移 AI 配置数据
func migrateAIConfig(db *gorm.DB, cfg *MigrationConfig) error {
	jsonPath := filepath.Join(cfg.DataDir, "config", "ai.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		logrus.Debug("No AI config JSON file found, skipping")
		return nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read AI config file: %w", err)
	}

	// 定义 JSON 配置结构
	var jsonConfig struct {
		Primary   string `json:"primary"`
		Providers map[string]struct {
			APIKey   string `json:"api_key"`
			BaseURL  string `json:"base_url,omitempty"`
			Model    string `json:"default_model,omitempty"`
			Enabled  bool   `json:"enabled"`
			Priority int    `json:"priority"`
		} `json:"providers"`
	}

	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return fmt.Errorf("parse AI config: %w", err)
	}

	// 转换为数据库模型
	for name, pc := range jsonConfig.Providers {
		if !pc.Enabled || pc.APIKey == "" {
			continue
		}

		config := AIServiceConfig{
			ServiceType: "text",
			Provider:    name,
			Name:        fmt.Sprintf("%s (migrated)", name),
			BaseURL:     pc.BaseURL,
			APIKey:      pc.APIKey,
			Model:       pc.Model,
			Priority:    pc.Priority,
			IsDefault:   name == jsonConfig.Primary,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// 检查是否已存在
		var count int64
		db.Model(&AIServiceConfig{}).Where("provider = ? AND name = ?", config.Provider, config.Name).Count(&count)
		if count > 0 {
			continue
		}

		if err := db.Create(&config).Error; err != nil {
			logrus.Warnf("Failed to migrate AI config for %s: %v", name, err)
		}
	}

	logrus.Info("Migrated AI config from JSON")
	return nil
}

// migrateCookies 迁移 Cookie 数据
func migrateCookies(db *gorm.DB, cfg *MigrationConfig) error {
	cookiesDir := filepath.Join(cfg.DataDir, "..", "cookies")
	if _, err := os.Stat(cookiesDir); os.IsNotExist(err) {
		logrus.Debug("No cookies directory found, skipping")
		return nil
	}

	files, err := filepath.Glob(filepath.Join(cookiesDir, "*.json"))
	if err != nil {
		return fmt.Errorf("list cookie files: %w", err)
	}

	for _, file := range files {
		platform := filepath.Base(file)
		platform = platform[:len(platform)-len(filepath.Ext(platform))]

		data, err := os.ReadFile(file)
		if err != nil {
			logrus.Warnf("Failed to read cookie file %s: %v", file, err)
			continue
		}

		cookie := Cookie{
			Platform:   platform,
			CookieData: string(data),
			IsValid:    true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// 检查是否已存在
		var count int64
		db.Model(&Cookie{}).Where("platform = ?", cookie.Platform).Count(&count)
		if count > 0 {
			continue
		}

		if err := db.Create(&cookie).Error; err != nil {
			logrus.Warnf("Failed to migrate cookie for %s: %v", platform, err)
		}
	}

	logrus.Infof("Migrated cookies from %d files", len(files))
	return nil
}

// ExportToJSON 导出数据库数据到 JSON（用于备份）
func ExportToJSON(db *gorm.DB, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// 导出话题
	var topics []Topic
	if err := db.Find(&topics).Error; err != nil {
		return fmt.Errorf("query topics: %w", err)
	}
	if len(topics) > 0 {
		data, _ := json.MarshalIndent(topics, "", "  ")
		if err := os.WriteFile(filepath.Join(outputDir, "topics.json"), data, 0644); err != nil {
			return err
		}
	}

	// 导出 AI 配置
	var aiConfigs []AIServiceConfig
	if err := db.Find(&aiConfigs).Error; err != nil {
		return fmt.Errorf("query AI configs: %w", err)
	}
	if len(aiConfigs) > 0 {
		data, _ := json.MarshalIndent(aiConfigs, "", "  ")
		if err := os.WriteFile(filepath.Join(outputDir, "ai_configs.json"), data, 0644); err != nil {
			return err
		}
	}

	logrus.Infof("Exported data to %s", outputDir)
	return nil
}
