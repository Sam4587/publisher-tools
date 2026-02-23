package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"publisher-core/account"
	"publisher-core/database"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AccountHandler 账号管理 API 处理器
type AccountHandler struct {
	accountService *account.AccountService
	poolService    *account.PoolService
}

// NewAccountHandler 创建账号管理 API 处理器
func NewAccountHandler(accountService *account.AccountService, poolService *account.PoolService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		poolService:    poolService,
	}
}

// RegisterRoutes 注册路由
func (h *AccountHandler) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// 登录相关 API
	api.HandleFunc("/login/status", h.CheckLoginStatus).Methods("GET")
	api.HandleFunc("/login/qrcode", h.GetLoginQrcode).Methods("GET")
	api.HandleFunc("/login/callback", h.LoginCallback).Methods("POST")

	// 账号管理 API
	accounts := api.PathPrefix("/accounts").Subrouter()
	accounts.HandleFunc("", h.CreateAccount).Methods("POST")
	accounts.HandleFunc("", h.ListAccounts).Methods("GET")
	accounts.HandleFunc("/{id}", h.GetAccount).Methods("GET")
	accounts.HandleFunc("/{id}", h.UpdateAccount).Methods("PUT")
	accounts.HandleFunc("/{id}", h.DeleteAccount).Methods("DELETE")
	accounts.HandleFunc("/{id}/stats", h.GetAccountStats).Methods("GET")
	accounts.HandleFunc("/{id}/health-check", h.HealthCheck).Methods("POST")
	accounts.HandleFunc("/{id}/cookies", h.GetDecryptedCookies).Methods("GET")

	// 账号池管理 API
	pools := api.PathPrefix("/pools").Subrouter()
	pools.HandleFunc("", h.CreatePool).Methods("POST")
	pools.HandleFunc("", h.ListPools).Methods("GET")
	pools.HandleFunc("/{id}", h.GetPool).Methods("GET")
	pools.HandleFunc("/{id}", h.UpdatePool).Methods("PUT")
	pools.HandleFunc("/{id}", h.DeletePool).Methods("DELETE")
	pools.HandleFunc("/{id}/members", h.AddPoolMember).Methods("POST")
	pools.HandleFunc("/{id}/members/{accountId}", h.RemovePoolMember).Methods("DELETE")
	pools.HandleFunc("/{id}/select", h.SelectAccountFromPool).Methods("POST")

	// 批量操作
	api.HandleFunc("/accounts/batch-health-check", h.BatchHealthCheck).Methods("POST")
}

// =====================================================
// 登录相关 API
// =====================================================

// CheckLoginStatus 检查登录状态
// GET /api/v1/login/status?platform={platform}
func (h *AccountHandler) CheckLoginStatus(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		jsonError(w, "INVALID_REQUEST", "platform parameter is required", http.StatusBadRequest)
		return
	}

	// 验证平台名称
	if !validatePlatform(platform) {
		jsonError(w, "INVALID_PLATFORM", "unsupported platform: "+platform, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 查找该平台的活跃账号
	accounts, _, err := h.accountService.ListAccounts(ctx, &account.ListAccountsRequest{
		Platform: platform,
		Status:   string(database.AccountStatusActive),
		Limit:    10,
	})

	if err != nil {
		jsonError(w, "QUERY_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	if len(accounts) == 0 {
		jsonSuccess(w, map[string]interface{}{
			"is_logged_in": false,
			"platform":     platform,
			"message":      "No active account found for this platform",
		})
		return
	}

	// 检查第一个账号的健康状态
	account := accounts[0]
	err = h.accountService.HealthCheck(ctx, account.AccountID)

	response := map[string]interface{}{
		"is_logged_in": err == nil,
		"platform":     platform,
		"account_id":   account.AccountID,
		"account_name": account.AccountName,
		"last_check":   account.LastCheckAt,
	}

	if err != nil {
		response["message"] = "Account health check failed: " + err.Error()
		response["status"] = "unhealthy"
	} else {
		response["message"] = "Account is healthy and logged in"
		response["status"] = "healthy"
	}

	jsonSuccess(w, response)
}

// LoginQrcodeResponse 登录二维码响应
type LoginQrcodeResponse struct {
	Platform    string `json:"platform"`
	QrcodeURL   string `json:"qrcode_url"`
	QrcodeImage string `json:"qrcode_image,omitempty"` // Base64 编码的图片
	ExpiresIn   int    `json:"expires_in"`             // 过期时间（秒）
	SessionID   string `json:"session_id"`             // 会话ID，用于后续轮询
	Message     string `json:"message"`
}

// GetLoginQrcode 获取登录二维码
// GET /api/v1/login/qrcode?platform={platform}
func (h *AccountHandler) GetLoginQrcode(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		jsonError(w, "INVALID_REQUEST", "platform parameter is required", http.StatusBadRequest)
		return
	}

	// TODO: 这里需要集成实际的浏览器自动化服务来获取二维码
	// 参考 xiaohongshu-mcp 的实现，需要：
	// 1. 启动无头浏览器
	// 2. 访问平台登录页面
	// 3. 提取二维码图片
	// 4. 返回二维码和会话信息

	// 临时返回模拟数据
	sessionID := uuid.New().String()

	response := LoginQrcodeResponse{
		Platform:  platform,
		QrcodeURL: getPlatformLoginURL(platform),
		ExpiresIn: 300, // 5分钟
		SessionID: sessionID,
		Message:   "Please scan the QR code with your mobile app to login",
	}

	logrus.Infof("Generated login QR code for platform %s, session: %s", platform, sessionID)
	jsonSuccess(w, response)
}

// LoginCallback 登录回调
// POST /api/v1/login/callback
func (h *AccountHandler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Platform    string `json:"platform"`
		SessionID   string `json:"session_id"`
		CookieData  string `json:"cookie_data"`
		AccountName string `json:"account_name"`
		UserID      string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.Platform == "" || req.CookieData == "" {
		jsonError(w, "INVALID_REQUEST", "platform and cookie_data are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 创建新账号
	accountReq := &account.CreateAccountRequest{
		Platform:    req.Platform,
		AccountName: req.AccountName,
		AccountType: "personal",
		CookieData:  req.CookieData,
		Priority:    5,
		UserID:      req.UserID,
	}

	newAccount, err := h.accountService.CreateAccount(ctx, accountReq)
	if err != nil {
		jsonError(w, "CREATE_ACCOUNT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 执行健康检查（使用带超时的上下文）
	go func(accountID string) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Panic in health check goroutine for account %s: %v", accountID, r)
			}
		}()

		// 创建带超时的上下文，避免 goroutine 泄漏
		checkCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		time.Sleep(2 * time.Second)
		if err := h.accountService.HealthCheck(checkCtx, accountID); err != nil {
			logrus.Warnf("Health check failed for new account %s: %v", accountID, err)
		}
	}(newAccount.AccountID)

	logrus.Infof("Account created successfully: %s (%s)", newAccount.AccountID, req.Platform)
	jsonSuccess(w, map[string]interface{}{
		"account_id":   newAccount.AccountID,
		"platform":     newAccount.Platform,
		"account_name": newAccount.AccountName,
		"status":       newAccount.Status,
		"message":      "Account created and logged in successfully",
	})
}

// =====================================================
// 账号管理 API
// =====================================================

// CreateAccount 创建账号
// POST /api/v1/accounts
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req account.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.Platform == "" || req.CookieData == "" {
		jsonError(w, "INVALID_REQUEST", "platform and cookie_data are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	account, err := h.accountService.CreateAccount(ctx, &req)
	if err != nil {
		jsonError(w, "CREATE_ACCOUNT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, account)
}

// ListAccounts 列出账号
// GET /api/v1/accounts?platform={platform}&status={status}&limit={limit}&offset={offset}
func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	status := r.URL.Query().Get("status")
	userID := r.URL.Query().Get("user_id")
	projectID := r.URL.Query().Get("project_id")
	limit := parseIntParam(r.URL.Query().Get("limit"), 50)
	offset := parseIntParam(r.URL.Query().Get("offset"), 0)

	ctx := r.Context()
	accounts, total, err := h.accountService.ListAccounts(ctx, &account.ListAccountsRequest{
		Platform:  platform,
		Status:    status,
		UserID:    userID,
		ProjectID: projectID,
		Limit:     limit,
		Offset:    offset,
	})

	if err != nil {
		jsonError(w, "LIST_ACCOUNTS_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"accounts": accounts,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetAccount 获取账号详情
// GET /api/v1/accounts/{id}
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	ctx := r.Context()
	account, err := h.accountService.GetAccount(ctx, accountID)
	if err != nil {
		jsonError(w, "ACCOUNT_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	jsonSuccess(w, account)
}

// UpdateAccount 更新账号
// PUT /api/v1/accounts/{id}
func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	var req account.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.accountService.UpdateAccount(ctx, accountID, &req); err != nil {
		jsonError(w, "UPDATE_ACCOUNT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "Account updated successfully",
	})
}

// DeleteAccount 删除账号
// DELETE /api/v1/accounts/{id}
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	ctx := r.Context()
	if err := h.accountService.DeleteAccount(ctx, accountID); err != nil {
		jsonError(w, "DELETE_ACCOUNT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "Account deleted successfully",
	})
}

// GetAccountStats 获取账号统计
// GET /api/v1/accounts/{id}/stats
func (h *AccountHandler) GetAccountStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	ctx := r.Context()
	stats, err := h.accountService.GetAccountStats(ctx, accountID)
	if err != nil {
		jsonError(w, "GET_STATS_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, stats)
}

// HealthCheck 健康检查
// POST /api/v1/accounts/{id}/health-check
func (h *AccountHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	ctx := r.Context()
	if err := h.accountService.HealthCheck(ctx, accountID); err != nil {
		jsonSuccess(w, map[string]interface{}{
			"account_id": accountID,
			"healthy":    false,
			"error":      err.Error(),
		})
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"account_id": accountID,
		"healthy":    true,
		"message":    "Account is healthy",
	})
}

// GetDecryptedCookies 获取解密后的 Cookie
// GET /api/v1/accounts/{id}/cookies
// 注意：此接口返回敏感数据，生产环境应添加认证和授权检查
func (h *AccountHandler) GetDecryptedCookies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	// TODO: 添加认证和授权检查
	// 只有账号所有者或管理员才能访问此接口

	ctx := r.Context()
	cookies, err := h.accountService.GetDecryptedCookies(ctx, accountID)
	if err != nil {
		jsonError(w, "GET_COOKIES_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录访问日志（用于审计）
	logrus.Warnf("Sensitive cookie access: account_id=%s, remote_addr=%s", 
		accountID, r.RemoteAddr)

	jsonSuccess(w, map[string]string{
		"cookies": cookies,
		"warning": "This endpoint exposes sensitive data. Use with caution.",
	})
}

// BatchHealthCheck 批量健康检查
// POST /api/v1/accounts/batch-health-check?platform={platform}
func (h *AccountHandler) BatchHealthCheck(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")

	ctx := r.Context()
	successCount, failCount, err := h.accountService.BatchHealthCheck(ctx, platform)
	if err != nil {
		jsonError(w, "BATCH_HEALTH_CHECK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"success_count": successCount,
		"fail_count":    failCount,
		"total":         successCount + failCount,
	})
}

// =====================================================
// 账号池管理 API
// =====================================================

// CreatePool 创建账号池
// POST /api/v1/pools
func (h *AccountHandler) CreatePool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Platform    string `json:"platform"`
		Description string `json:"description"`
		Strategy    string `json:"strategy"`
		MaxSize     int    `json:"max_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	pool, err := h.poolService.CreatePool(ctx, &account.CreatePoolRequest{
		Name:        req.Name,
		Platform:    req.Platform,
		Description: req.Description,
		Strategy:    account.PoolStrategy(req.Strategy),
		MaxSize:     req.MaxSize,
	})
	if err != nil {
		jsonError(w, "CREATE_POOL_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, pool)
}

// ListPools 列出账号池
// GET /api/v1/pools?platform={platform}
func (h *AccountHandler) ListPools(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")

	ctx := r.Context()
	pools, _, err := h.poolService.ListPools(ctx, &account.ListPoolsRequest{
		Platform: platform,
	})
	if err != nil {
		jsonError(w, "LIST_POOLS_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, pools)
}

// GetPool 获取账号池详情
// GET /api/v1/pools/{id}
func (h *AccountHandler) GetPool(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]

	ctx := r.Context()
	pool, err := h.poolService.GetPool(ctx, poolID)
	if err != nil {
		jsonError(w, "POOL_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	jsonSuccess(w, pool)
}

// UpdatePool 更新账号池
// PUT /api/v1/pools/{id}
func (h *AccountHandler) UpdatePool(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Strategy    string `json:"strategy"`
		MaxSize     int    `json:"max_size"`
		IsActive    bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	isActive := req.IsActive
	if err := h.poolService.UpdatePool(ctx, poolID, &account.UpdatePoolRequest{
		Name:        req.Name,
		Description: req.Description,
		Strategy:    account.PoolStrategy(req.Strategy),
		MaxSize:     req.MaxSize,
		IsActive:    &isActive,
	}); err != nil {
		jsonError(w, "UPDATE_POOL_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "Pool updated successfully",
	})
}

// DeletePool 删除账号池
// DELETE /api/v1/pools/{id}
func (h *AccountHandler) DeletePool(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]

	ctx := r.Context()
	if err := h.poolService.DeletePool(ctx, poolID); err != nil {
		jsonError(w, "DELETE_POOL_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "Pool deleted successfully",
	})
}

// AddPoolMember 添加账号到池
// POST /api/v1/pools/{id}/members
func (h *AccountHandler) AddPoolMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]

	var req struct {
		AccountID string `json:"account_id"`
		Priority  int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.poolService.AddMember(ctx, poolID, req.AccountID, req.Priority); err != nil {
		jsonError(w, "ADD_MEMBER_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "Account added to pool successfully",
	})
}

// RemovePoolMember 从池中移除账号
// DELETE /api/v1/pools/{id}/members/{accountId}
func (h *AccountHandler) RemovePoolMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]
	accountID := vars["accountId"]

	ctx := r.Context()
	if err := h.poolService.RemoveMember(ctx, poolID, accountID); err != nil {
		jsonError(w, "REMOVE_MEMBER_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "Account removed from pool successfully",
	})
}

// SelectAccountFromPool 从池中选择账号
// POST /api/v1/pools/{id}/select
func (h *AccountHandler) SelectAccountFromPool(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	poolID := vars["id"]

	ctx := r.Context()
	selectedAccount, err := h.poolService.SelectAccountFromPool(ctx, poolID)
	if err != nil {
		jsonError(w, "SELECT_ACCOUNT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, selectedAccount)
}

// =====================================================
// 辅助函数
// =====================================================

// 输入验证常量
const (
	MaxRequestSize = 10 * 1024 * 1024 // 10MB
	MaxLimit       = 1000
	MaxPriority    = 10
)

// validatePlatform 验证平台名称
func validatePlatform(platform string) bool {
	validPlatforms := map[string]bool{
		"xiaohongshu": true,
		"douyin":      true,
		"toutiao":     true,
		"bilibili":    true,
	}
	return validPlatforms[platform]
}

// validateStrategy 验证负载均衡策略
func validateStrategy(strategy string) bool {
	validStrategies := map[string]bool{
		"round_robin": true,
		"random":      true,
		"priority":    true,
		"least_used":  true,
	}
	return validStrategies[strategy]
}

// validateUUID 验证 UUID 格式
func validateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// validatePriority 验证优先级
func validatePriority(priority int) bool {
	return priority >= 1 && priority <= MaxPriority
}

// getPlatformLoginURL 获取平台登录 URL
func getPlatformLoginURL(platform string) string {
	urls := map[string]string{
		"xiaohongshu": "https://www.xiaohongshu.com/explore",
		"douyin":      "https://www.douyin.com/",
		"toutiao":     "https://www.toutiao.com/",
		"bilibili":    "https://www.bilibili.com/",
	}

	if url, ok := urls[platform]; ok {
		return url
	}
	return ""
}

// parseIntParam 解析整数参数
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	
	// 使用 strconv.Atoi 更高效且正确
	val, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}
	
	// 添加边界检查，防止溢出
	if val < 0 {
		return 0
	}
	if val > 10000 { // 设置合理的最大值
		return 10000
	}
	
	return val
}
