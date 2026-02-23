package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"publisher-core/storage"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// StorageHandlers 存储处理器
type StorageHandlers struct {
	storage   storage.Storage
	uploadDir string
}

// NewStorageHandlers 创建存储处理器
func NewStorageHandlers(storage storage.Storage, uploadDir string) *StorageHandlers {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	return &StorageHandlers{
		storage:   storage,
		uploadDir: uploadDir,
	}
}

// RegisterRoutes 注册路由
func (h *StorageHandlers) RegisterRoutes(router *mux.Router) {
	storageRouter := router.PathPrefix("/api/v1/storage").Subrouter()

	// 文件操作
	storageRouter.HandleFunc("/upload", h.uploadFile).Methods("POST")
	storageRouter.HandleFunc("/download/{path:.*}", h.downloadFile).Methods("GET")
	storageRouter.HandleFunc("/list", h.listFiles).Methods("GET")
	storageRouter.HandleFunc("/list/{prefix:.*}", h.listFilesByPrefix).Methods("GET")
	storageRouter.HandleFunc("/delete/{path:.*}", h.deleteFile).Methods("DELETE")
	storageRouter.HandleFunc("/exists/{path:.*}", h.fileExists).Methods("GET")
	storageRouter.HandleFunc("/info/{path:.*}", h.getFileInfo).Methods("GET")
	storageRouter.HandleFunc("/url/{path:.*}", h.getFileURL).Methods("GET")

	// 批量操作
	storageRouter.HandleFunc("/batch/delete", h.batchDelete).Methods("POST")
	storageRouter.HandleFunc("/batch/copy", h.batchCopy).Methods("POST")
	storageRouter.HandleFunc("/batch/move", h.batchMove).Methods("POST")
}

// uploadFile 上传文件
func (h *StorageHandlers) uploadFile(w http.ResponseWriter, r *http.Request) {
	maxSize := int64(100 * 1024 * 1024) // 100MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, "PARSE_FORM_FAILED", err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "MISSING_FILE", "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 获取自定义路径
	customPath := r.FormValue("path")
	var filePath string
	if customPath != "" {
		filePath = customPath
	} else {
		// 按日期组织目录
		dateDir := time.Now().Format("2006/01/02")
		filePath = fmt.Sprintf("%s/%d_%s", dateDir, time.Now().UnixNano(), header.Filename)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// 使用流式写入
	if err := h.storage.WriteStream(ctx, filePath, file); err != nil {
		jsonError(w, "UPLOAD_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取文件信息
	info, err := h.storage.Stat(ctx, filePath)
	if err != nil {
		logrus.Warnf("Failed to get file info: %v", err)
	}

	// 获取文件URL
	url, err := h.storage.GetURL(ctx, filePath)
	if err != nil {
		logrus.Warnf("Failed to get file URL: %v", err)
	}

	jsonSuccess(w, map[string]interface{}{
		"path":      filePath,
		"file_name": header.Filename,
		"size":      header.Size,
		"mime_type": info.MimeType,
		"hash":      info.Hash,
		"url":       url,
	})
}

// downloadFile 下载文件
func (h *StorageHandlers) downloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// 检查文件是否存在
	exists, err := h.storage.Exists(ctx, filePath)
	if err != nil {
		jsonError(w, "CHECK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		jsonError(w, "FILE_NOT_FOUND", "file not found", http.StatusNotFound)
		return
	}

	// 获取文件流
	reader, err := h.storage.ReadStream(ctx, filePath)
	if err != nil {
		jsonError(w, "READ_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// 获取文件信息
	info, err := h.storage.Stat(ctx, filePath)
	if err != nil {
		logrus.Warnf("Failed to get file info: %v", err)
	}

	// 设置响应头
	if info != nil && info.MimeType != "" {
		w.Header().Set("Content-Type", info.MimeType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// 获取文件名
	fileName := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))

	if info != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
	}

	// 流式传输
	if _, err := io.Copy(w, reader); err != nil {
		logrus.Errorf("Failed to send file: %v", err)
	}
}

// listFiles 列出文件
func (h *StorageHandlers) listFiles(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = ""
	}

	h.doListFiles(w, r, prefix)
}

// listFilesByPrefix 按前缀列出文件
func (h *StorageHandlers) listFilesByPrefix(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	prefix := vars["prefix"]

	h.doListFiles(w, r, prefix)
}

// doListFiles 执行文件列表查询
func (h *StorageHandlers) doListFiles(w http.ResponseWriter, r *http.Request, prefix string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	files, err := h.storage.List(ctx, prefix)
	if err != nil {
		jsonError(w, "LIST_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取每个文件的详细信息
	type fileInfo struct {
		Path     string `json:"path"`
		Size     int64  `json:"size"`
		MimeType string `json:"mime_type"`
		Modified string `json:"modified"`
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := h.storage.Stat(ctx, file)
		if err != nil {
			logrus.Warnf("Failed to get info for %s: %v", file, err)
			fileInfos = append(fileInfos, fileInfo{Path: file})
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			Path:     file,
			Size:     info.Size,
			MimeType: info.MimeType,
			Modified: info.UpdatedAt.Format(time.RFC3339),
		})
	}

	jsonSuccess(w, map[string]interface{}{
		"files": fileInfos,
		"total": len(fileInfos),
		"prefix": prefix,
	})
}

// deleteFile 删除文件
func (h *StorageHandlers) deleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// 检查文件是否存在
	exists, err := h.storage.Exists(ctx, filePath)
	if err != nil {
		jsonError(w, "CHECK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		jsonError(w, "FILE_NOT_FOUND", "file not found", http.StatusNotFound)
		return
	}

	// 删除文件
	if err := h.storage.Delete(ctx, filePath); err != nil {
		jsonError(w, "DELETE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"message": "File deleted successfully",
		"path":    filePath,
	})
}

// fileExists 检查文件是否存在
func (h *StorageHandlers) fileExists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	exists, err := h.storage.Exists(ctx, filePath)
	if err != nil {
		jsonError(w, "CHECK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"path":   filePath,
		"exists": exists,
	})
}

// getFileInfo 获取文件信息
func (h *StorageHandlers) getFileInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := h.storage.Stat(ctx, filePath)
	if err != nil {
		jsonError(w, "GET_INFO_FAILED", err.Error(), http.StatusNotFound)
		return
	}

	// 获取URL
	url, err := h.storage.GetURL(ctx, filePath)
	if err != nil {
		logrus.Warnf("Failed to get URL: %v", err)
	}

	jsonSuccess(w, map[string]interface{}{
		"path":      info.Path,
		"size":      info.Size,
		"mime_type": info.MimeType,
		"hash":      info.Hash,
		"updated":   info.UpdatedAt.Format(time.RFC3339),
		"url":       url,
	})
}

// getFileURL 获取文件URL
func (h *StorageHandlers) getFileURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// 检查文件是否存在
	exists, err := h.storage.Exists(ctx, filePath)
	if err != nil {
		jsonError(w, "CHECK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		jsonError(w, "FILE_NOT_FOUND", "file not found", http.StatusNotFound)
		return
	}

	// 获取URL
	url, err := h.storage.GetURL(ctx, filePath)
	if err != nil {
		jsonError(w, "GET_URL_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取签名URL（可选）
	expiryStr := r.URL.Query().Get("expiry")
	var signedURL string
	if expiryStr != "" {
		expiry, err := time.ParseDuration(expiryStr)
		if err == nil {
			signedURL, err = h.storage.GetSignedURL(ctx, filePath, expiry)
			if err != nil {
				logrus.Warnf("Failed to get signed URL: %v", err)
			}
		}
	}

	response := map[string]interface{}{
		"path": filePath,
		"url":  url,
	}
	if signedURL != "" {
		response["signed_url"] = signedURL
	}

	jsonSuccess(w, response)
}

// batchDelete 批量删除
func (h *StorageHandlers) batchDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Paths []string `json:"paths"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Paths) == 0 {
		jsonError(w, "NO_PATHS", "paths list is empty", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	var deleted []string
	var failed []map[string]string

	for _, path := range req.Paths {
		if err := h.storage.Delete(ctx, path); err != nil {
			failed = append(failed, map[string]string{
				"path":  path,
				"error": err.Error(),
			})
		} else {
			deleted = append(deleted, path)
		}
	}

	jsonSuccess(w, map[string]interface{}{
		"deleted": deleted,
		"failed":  failed,
		"summary": map[string]int{
			"total":   len(req.Paths),
			"success": len(deleted),
			"failed":  len(failed),
		},
	})
}

// batchCopy 批量复制
func (h *StorageHandlers) batchCopy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		jsonError(w, "NO_ITEMS", "items list is empty", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	var copied []string
	var failed []map[string]string

	for _, item := range req.Items {
		if err := storage.Copy(h.storage, ctx, item.Src, item.Dst); err != nil {
			failed = append(failed, map[string]string{
				"src":   item.Src,
				"dst":   item.Dst,
				"error": err.Error(),
			})
		} else {
			copied = append(copied, item.Dst)
		}
	}

	jsonSuccess(w, map[string]interface{}{
		"copied": copied,
		"failed": failed,
		"summary": map[string]int{
			"total":   len(req.Items),
			"success": len(copied),
			"failed":  len(failed),
		},
	})
}

// batchMove 批量移动
func (h *StorageHandlers) batchMove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		jsonError(w, "NO_ITEMS", "items list is empty", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	var moved []string
	var failed []map[string]string

	for _, item := range req.Items {
		if err := storage.Move(h.storage, ctx, item.Src, item.Dst); err != nil {
			failed = append(failed, map[string]string{
				"src":   item.Src,
				"dst":   item.Dst,
				"error": err.Error(),
			})
		} else {
			moved = append(moved, item.Dst)
		}
	}

	jsonSuccess(w, map[string]interface{}{
		"moved": moved,
		"failed": failed,
		"summary": map[string]int{
			"total":   len(req.Items),
			"success": len(moved),
			"failed":  len(failed),
		},
	})
}

// StorageAPIAdapter 存储API适配器，实现server.go中的StorageAPI接口
type StorageAPIAdapter struct {
	storage storage.Storage
}

// NewStorageAPIAdapter 创建存储API适配器
func NewStorageAPIAdapter(s storage.Storage) *StorageAPIAdapter {
	return &StorageAPIAdapter{storage: s}
}

// Upload 上传文件
func (a *StorageAPIAdapter) Upload(file []byte, path string) (string, error) {
	ctx := context.Background()
	if err := a.storage.Write(ctx, path, file); err != nil {
		return "", err
	}
	return a.storage.GetURL(ctx, path)
}

// Download 下载文件
func (a *StorageAPIAdapter) Download(path string) ([]byte, error) {
	ctx := context.Background()
	return a.storage.Read(ctx, path)
}

// List 列出文件
func (a *StorageAPIAdapter) List(prefix string) ([]string, error) {
	ctx := context.Background()
	return a.storage.List(ctx, prefix)
}

// Delete 删除文件
func (a *StorageAPIAdapter) Delete(path string) error {
	ctx := context.Background()
	return a.storage.Delete(ctx, path)
}

// parsePathList 解析逗号分隔的路径列表
func parsePathList(paths string) []string {
	if paths == "" {
		return nil
	}
	result := strings.Split(paths, ",")
	for i, p := range result {
		result[i] = strings.TrimSpace(p)
	}
	return result
}
