package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"publisher-core/account"
	"publisher-core/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 注意：此测试文件需要以下依赖：
// - github.com/gorilla/mux
// - gorm.io/gorm
// - gorm.io/driver/sqlite
// 运行测试前请确保已安装这些依赖

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&database.PlatformAccount{},
		&database.AccountUsageLog{},
		&database.AccountPool{},
		&database.AccountPoolMember{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// setupTestHandler 创建测试处理器
func setupTestHandler(t *testing.T) *AccountHandler {
	db := setupTestDB(t)

	// 创建账号服务
	encryptionConfig := &account.EncryptionConfig{
		EncryptionKey: "test-encryption-key-32-bytes-long",
	}
	accountService := account.NewAccountService(db, encryptionConfig)

	// 创建账号池管理器
	poolManager := account.NewPoolManager(db, accountService)

	return NewAccountHandler(accountService, poolManager)
}

func TestCheckLoginStatus(t *testing.T) {
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		platform   string
		wantStatus int
	}{
		{
			name:       "检查小红书登录状态",
			platform:   "xiaohongshu",
			wantStatus: http.StatusOK,
		},
		{
			name:       "缺少平台参数",
			platform:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/login/status?platform="+tt.platform, nil)
			w := httptest.NewRecorder()

			handler.CheckLoginStatus(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CheckLoginStatus() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// 验证响应格式
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}

			if _, ok := response["success"]; !ok {
				t.Error("Response missing 'success' field")
			}
		})
	}
}

func TestGetLoginQrcode(t *testing.T) {
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		platform   string
		wantStatus int
	}{
		{
			name:       "获取小红书登录二维码",
			platform:   "xiaohongshu",
			wantStatus: http.StatusOK,
		},
		{
			name:       "获取抖音登录二维码",
			platform:   "douyin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "缺少平台参数",
			platform:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/login/qrcode?platform="+tt.platform, nil)
			w := httptest.NewRecorder()

			handler.GetLoginQrcode(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetLoginQrcode() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				data := response["data"].(map[string]interface{})
				if data["platform"] != tt.platform {
					t.Errorf("Expected platform %s, got %v", tt.platform, data["platform"])
				}
				if _, ok := data["session_id"]; !ok {
					t.Error("Response missing 'session_id' field")
				}
			}
		})
	}
}

func TestLoginCallback(t *testing.T) {
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		request    map[string]interface{}
		wantStatus int
	}{
		{
			name: "成功的登录回调",
			request: map[string]interface{}{
				"platform":     "xiaohongshu",
				"session_id":   "test_session",
				"cookie_data":  "test_cookie_data",
				"account_name": "测试账号",
				"user_id":      "user_001",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "缺少平台参数",
			request: map[string]interface{}{
				"cookie_data": "test_cookie_data",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "缺少 Cookie 数据",
			request: map[string]interface{}{
				"platform": "xiaohongshu",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/login/callback", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.LoginCallback(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("LoginCallback() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				data := response["data"].(map[string]interface{})
				if _, ok := data["account_id"]; !ok {
					t.Error("Response missing 'account_id' field")
				}
			}
		})
	}
}

func TestCreateAccount(t *testing.T) {
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		request    map[string]interface{}
		wantStatus int
	}{
		{
			name: "创建账号成功",
			request: map[string]interface{}{
				"platform":     "douyin",
				"account_name": "抖音测试账号",
				"account_type": "personal",
				"cookie_data":  "test_cookie",
				"priority":     8,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "缺少必填字段",
			request: map[string]interface{}{
				"account_name": "测试账号",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateAccount(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateAccount() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestListAccounts(t *testing.T) {
	handler := setupTestHandler(t)

	// 先创建一些测试账号
	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(map[string]interface{}{
			"platform":     "xiaohongshu",
			"account_name": "测试账号",
			"cookie_data":  "test_cookie_" + string(rune('A'+i)),
		})
		req := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.CreateAccount(w, req)
	}

	tests := []struct {
		name       string
		platform   string
		wantStatus int
	}{
		{
			name:       "列出所有账号",
			platform:   "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "按平台筛选",
			platform:   "xiaohongshu",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/accounts"
			if tt.platform != "" {
				url += "?platform=" + tt.platform
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.ListAccounts(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ListAccounts() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}

			data := response["data"].(map[string]interface{})
			if _, ok := data["accounts"]; !ok {
				t.Error("Response missing 'accounts' field")
			}
		})
	}
}

func TestCreatePool(t *testing.T) {
	handler := setupTestHandler(t)

	request := map[string]interface{}{
		"name":        "测试账号池",
		"platform":    "xiaohongshu",
		"description": "用于测试的账号池",
		"strategy":    "round_robin",
		"max_size":    10,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/pools", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePool(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("CreatePool() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	if _, ok := data["pool_id"]; !ok {
		t.Error("Response missing 'pool_id' field")
	}
}
