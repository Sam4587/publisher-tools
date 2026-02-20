package database

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

// TestInit 测试数据库初始化
func TestInit(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// 初始化数据库
	cfg := &Config{
		DBPath:      dbPath,
		Debug:       false,
		AutoMigrate: true,
	}

	db, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if db == nil {
		t.Fatal("Init() returned nil db")
	}

	// 验证数据库文件已创建
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// 验证表是否创建
	tables := []string{
		"ai_service_configs",
		"platforms",
		"topics",
		"rank_history",
		"crawl_records",
		"videos",
		"transcripts",
		"notification_channels",
		"notification_templates",
		"tasks",
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("Table %s was not created", table)
		}
	}

	// 清理
	Close()
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DBPath != "./data/publisher.db" {
		t.Errorf("Default DBPath = %v, want %v", cfg.DBPath, "./data/publisher.db")
	}

	if cfg.Debug != false {
		t.Errorf("Default Debug = %v, want %v", cfg.Debug, false)
	}

	if cfg.AutoMigrate != true {
		t.Errorf("Default AutoMigrate = %v, want %v", cfg.AutoMigrate, true)
	}
}

// TestGetDB 测试获取数据库实例
func TestGetDB(t *testing.T) {
	// 初始化数据库
	tmpDir := t.TempDir()
	cfg := &Config{
		DBPath:      filepath.Join(tmpDir, "test.db"),
		Debug:       false,
		AutoMigrate: false,
	}

	_, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 获取数据库实例
	db := GetDB()
	if db == nil {
		t.Error("GetDB() returned nil")
	}

	// 清理
	Close()
}

// TestClose 测试关闭数据库连接
func TestClose(t *testing.T) {
	// 初始化数据库
	tmpDir := t.TempDir()
	cfg := &Config{
		DBPath:      filepath.Join(tmpDir, "test.db"),
		Debug:       false,
		AutoMigrate: false,
	}

	_, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 关闭数据库
	err = Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 验证数据库已关闭
	db := GetDB()
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			err := sqlDB.Ping()
			if err == nil {
				t.Error("Database connection should be closed")
			}
		}
	}
}

// TestReset 测试重置数据库
func TestReset(t *testing.T) {
	// 初始化数据库
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	cfg := &Config{
		DBPath:      dbPath,
		Debug:       false,
		AutoMigrate: true,
	}

	_, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 插入一些测试数据
	db := GetDB()
	db.Exec("INSERT INTO platforms (id, name, is_active) VALUES ('test', 'Test Platform', 1)")

	// 重置数据库
	err = Reset(cfg)
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// 验证数据已清空
	var count int64
	db.Table("platforms").Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 records after reset, got %d", count)
	}

	// 清理
	Close()
}

// TestTransaction 测试事务
func TestTransaction(t *testing.T) {
	// 初始化数据库
	tmpDir := t.TempDir()
	cfg := &Config{
		DBPath:      filepath.Join(tmpDir, "test.db"),
		Debug:       false,
		AutoMigrate: true,
	}

	_, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 测试成功的事务
	err = Transaction(func(tx *gorm.DB) error {
		return tx.Exec("INSERT INTO platforms (id, name, is_active) VALUES ('test1', 'Test 1', 1)").Error
	})

	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}

	// 验证数据已插入
	var count int64
	GetDB().Table("platforms").Where("id = ?", "test1").Count(&count)
	if count != 1 {
		t.Error("Transaction did not commit")
	}

	// 测试失败的事务
	err = Transaction(func(tx *gorm.DB) error {
		tx.Exec("INSERT INTO platforms (id, name, is_active) VALUES ('test2', 'Test 2', 1)")
		return tx.Exec("INSERT INTO platforms (id, name, is_active) VALUES ('test2', 'Test 2', 1)").Error // 重复 ID
	})

	if err == nil {
		t.Error("Expected error for duplicate key")
	}

	// 验证数据未插入
	GetDB().Table("platforms").Where("id = ?", "test2").Count(&count)
	if count != 0 {
		t.Error("Rollback did not work")
	}

	// 清理
	Close()
}

// TestPing 测试数据库连接
func TestPing(t *testing.T) {
	// 初始化数据库
	tmpDir := t.TempDir()
	cfg := &Config{
		DBPath:      filepath.Join(tmpDir, "test.db"),
		Debug:       false,
		AutoMigrate: false,
	}

	_, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 测试 Ping
	err = Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// 清理
	Close()
}
