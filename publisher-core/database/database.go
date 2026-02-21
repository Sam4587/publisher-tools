package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// Config 数据库配置
type Config struct {
	// 数据库文件路径
	DBPath string
	// 是否开启调试模式
	Debug bool
	// 是否自动迁移
	AutoMigrate bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DBPath:      "./data/publisher.db",
		Debug:       false,
		AutoMigrate: true,
	}
}

// Init 初始化数据库
func Init(cfg *Config) (*gorm.DB, error) {
	var initErr error
	once.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}

		// 确保数据库目录存在
		dbDir := filepath.Dir(cfg.DBPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create database directory: %w", err)
			return
		}

		// 配置 GORM
		gormConfig := &gorm.Config{}
		if cfg.Debug {
			gormConfig.Logger = logger.Default.LogMode(logger.Info)
		} else {
			gormConfig.Logger = logger.Default.LogMode(logger.Silent)
		}

		// 打开数据库连接
		var err error
		db, err = gorm.Open(sqlite.Open(cfg.DBPath), gormConfig)
		if err != nil {
			initErr = fmt.Errorf("failed to connect database: %w", err)
			return
		}

		// 自动迁移
		if cfg.AutoMigrate {
			if err = autoMigrate(db); err != nil {
				initErr = fmt.Errorf("failed to migrate database: %w", err)
				return
			}
		}
	})

	return db, initErr
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// AI 服务配置
		&AIServiceConfig{},
		// 热点监控
		&Platform{},
		&Topic{},
		&RankHistory{},
		&CrawlRecord{},
		// 视频处理
		&Video{},
		&Transcript{},
		// 通知服务
		&NotificationChannel{},
		&NotificationTemplate{},
		// 任务管理
		&Task{},
		// Cookie 管理
		&Cookie{},
		// AI 历史记录
		&AIHistory{},
		// AI 提示词模板
		&PromptTemplate{},
		&PromptTemplateVersion{},
		&PromptTemplateABTest{},
		&PromptTemplateUsage{},
		// AI 成本追踪
		&AICostRecord{},
		&AIBudget{},
		&AICostAlert{},
		&AIModelPricing{},
		// 异步任务系统
		&AsyncTask{},
		&TaskQueue{},
		&TaskExecution{},
		&ScheduledTask{},
	)
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Reset 重置数据库（仅用于测试）
func Reset(cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 关闭现有连接
	if err := Close(); err != nil {
		return err
	}

	// 删除数据库文件
	if _, err := os.Stat(cfg.DBPath); err == nil {
		if err := os.Remove(cfg.DBPath); err != nil {
			return fmt.Errorf("failed to remove database file: %w", err)
		}
	}

	// 重置 once 以允许重新初始化
	once = sync.Once{}

	// 重新初始化
	_, err := Init(cfg)
	return err
}

// Transaction 执行事务
func Transaction(fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}

// Ping 测试数据库连接
func Ping() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
