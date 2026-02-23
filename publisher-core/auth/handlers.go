package auth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"publisher-core/util"
)

// AuthHandler 认证 API 处理器
type AuthHandler struct {
	userService *UserService
	authService *AuthService
}

// NewAuthHandler 创建认证 API 处理器
func NewAuthHandler(userService *UserService, authService *AuthService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
	}
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// 公开接口（不需要认证）
	api.HandleFunc("/register", h.Register).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")

	// 需要认证的接口
	auth := api.PathPrefix("/auth").Subrouter()
	auth.Use(h.authService.AuthMiddleware)
	auth.HandleFunc("/me", h.GetCurrentUser).Methods("GET")
	auth.HandleFunc("/me", h.UpdateCurrentUser).Methods("PUT")
	auth.HandleFunc("/me/api-key", h.RegenerateAPIKey).Methods("POST")
	auth.HandleFunc("/change-password", h.ChangePassword).Methods("POST")

	// 管理员接口
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(h.authService.AuthMiddleware)
	admin.Use(h.authService.RequireRole("admin"))
	admin.HandleFunc("/users", h.ListUsers).Methods("GET")
	admin.HandleFunc("/users/{id}", h.GetUser).Methods("GET")
	admin.HandleFunc("/users/{id}/role", h.ChangeUserRole).Methods("PUT")
	admin.HandleFunc("/users/{id}/disable", h.DisableUser).Methods("POST")
	admin.HandleFunc("/users/{id}/enable", h.EnableUser).Methods("POST")
}

// =====================================================
// 公开接口
// =====================================================

// Register 用户注册
// POST /api/v1/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Username == "" || req.Password == "" {
		jsonError(w, "INVALID_REQUEST", "username and password are required", http.StatusBadRequest)
		return
	}

	// 验证用户名
	if err := util.NewStringValidator(req.Username).
		Required().
		MinLength(3).
		MaxLength(50).
		AlphaNumeric().
		NoSQLInjection().
		NoXSS().
		Validate(); err != nil {
		jsonError(w, "INVALID_USERNAME", err.Error(), http.StatusBadRequest)
		return
	}

	// 验证密码
	if err := util.NewPasswordValidator(req.Password).
		MinLength(8).
		Validate(); err != nil {
		jsonError(w, "INVALID_PASSWORD", err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.RegisterUser(r.Context(), &req)
	if err != nil {
		jsonError(w, "REGISTER_FAILED", err.Error(), http.StatusBadRequest)
		return
	}

	logrus.Infof("User registered successfully: %s", user.UserID)
	jsonSuccess(w, map[string]interface{}{
		"user_id":  user.UserID,
		"username": user.Username,
		"api_key":  user.APIKey,
		"message":  "User registered successfully",
	})
}

// Login 用户登录
// POST /api/v1/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "INVALID_REQUEST", "username and password are required", http.StatusBadRequest)
		return
	}

	response, err := h.userService.LoginUser(r.Context(), &req)
	if err != nil {
		jsonError(w, "LOGIN_FAILED", "Invalid credentials", http.StatusUnauthorized)
		return
	}

	jsonSuccess(w, response)
}

// =====================================================
// 认证接口
// =====================================================

// GetCurrentUser 获取当前用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r.Context())
	if err != nil {
		jsonError(w, "UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
		return
	}

	jsonSuccess(w, user)
}

// UpdateCurrentUser 更新当前用户信息
// PUT /api/v1/auth/me
func (h *AuthHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r.Context())
	if err != nil {
		jsonError(w, "UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdateUser(r.Context(), user.UserID, &req); err != nil {
		jsonError(w, "UPDATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "User updated successfully",
	})
}

// RegenerateAPIKey 重新生成 API Key
// POST /api/v1/auth/me/api-key
func (h *AuthHandler) RegenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r.Context())
	if err != nil {
		jsonError(w, "UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
		return
	}

	apiKey, err := h.userService.RegenerateAPIKey(r.Context(), user.UserID)
	if err != nil {
		jsonError(w, "REGENERATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"api_key": apiKey,
		"message": "API Key regenerated successfully",
	})
}

// ChangePassword 修改密码
// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromContext(r.Context())
	if err != nil {
		jsonError(w, "UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	// 验证旧密码
	loginReq := &LoginRequest{
		Username: user.Username,
		Password: req.OldPassword,
	}
	if _, err := h.userService.LoginUser(r.Context(), loginReq); err != nil {
		jsonError(w, "INVALID_PASSWORD", "Old password is incorrect", http.StatusBadRequest)
		return
	}

	// 更新密码
	updateReq := &UpdateUserRequest{
		Password: req.NewPassword,
	}
	if err := h.userService.UpdateUser(r.Context(), user.UserID, updateReq); err != nil {
		jsonError(w, "UPDATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "Password changed successfully",
	})
}

// =====================================================
// 管理员接口
// =====================================================

// ListUsers 列出用户
// GET /api/v1/admin/users
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r.URL.Query().Get("limit"), 50)
	offset := parseIntParam(r.URL.Query().Get("offset"), 0)

	users, total, err := h.userService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		jsonError(w, "LIST_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"users": users,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// GetUser 获取用户信息
// GET /api/v1/admin/users/{id}
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		jsonError(w, "USER_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	jsonSuccess(w, user)
}

// ChangeUserRole 修改用户角色
// PUT /api/v1/admin/users/{id}/role
func (h *AuthHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangeUserRole(r.Context(), userID, req.Role); err != nil {
		jsonError(w, "CHANGE_ROLE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "User role changed successfully",
	})
}

// DisableUser 禁用用户
// POST /api/v1/admin/users/{id}/disable
func (h *AuthHandler) DisableUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if err := h.userService.DisableUser(r.Context(), userID); err != nil {
		jsonError(w, "DISABLE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "User disabled successfully",
	})
}

// EnableUser 启用用户
// POST /api/v1/admin/users/{id}/enable
func (h *AuthHandler) EnableUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if err := h.userService.EnableUser(r.Context(), userID); err != nil {
		jsonError(w, "ENABLE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"message": "User enabled successfully",
	})
}

// =====================================================
// 辅助函数
// =====================================================

func jsonSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func jsonError(w http.ResponseWriter, code, message string, statusCode int) {
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

func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	var result int
	if val, err := json.Number(param).Int64(); err == nil {
		result = int(val)
		return result
	}
	return defaultValue
}
