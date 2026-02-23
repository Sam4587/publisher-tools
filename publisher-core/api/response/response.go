package response

import (
	"encoding/json"
	"net/http"
)

// JSONResponse JSON响应结构
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSONSuccess 返回JSON成功响应
func JSONSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := JSONResponse{
		Success: true,
		Data:    data,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// 如果编码失败,尝试返回简单的错误信息
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// JSONError 返回JSON错误响应
func JSONError(w http.ResponseWriter, code string, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := JSONResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// 如果编码失败,返回状态码
		w.WriteHeader(statusCode)
		return
	}
}

// JSONErrorWithDetails 返回带详细信息的JSON错误响应
func JSONErrorWithDetails(w http.ResponseWriter, code string, message string, statusCode int, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := JSONResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
		Data: details,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(statusCode)
		return
	}
}
