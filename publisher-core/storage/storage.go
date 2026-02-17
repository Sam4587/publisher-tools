package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Storage interface {
	Write(ctx context.Context, path string, data []byte) error
	WriteStream(ctx context.Context, path string, reader io.Reader) error
	Read(ctx context.Context, path string) ([]byte, error)
	ReadStream(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	Stat(ctx context.Context, path string) (*FileInfo, error)
	List(ctx context.Context, prefix string) ([]string, error)
	GetURL(ctx context.Context, path string) (string, error)
	GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)
}

type FileInfo struct {
	Path      string
	Size      int64
	MimeType  string
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
	StorageTypeOSS   StorageType = "oss"
	StorageTypeCOS   StorageType = "cos"
)

type Config struct {
	Type      StorageType
	RootDir   string
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
	BaseURL   string
}

type LocalStorage struct {
	rootDir string
	baseURL string
	mu      sync.RWMutex
}

func NewLocalStorage(rootDir string, baseURL string) (*LocalStorage, error) {
	if rootDir == "" {
		rootDir = "./uploads"
	}

	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		rootDir: rootDir,
		baseURL: baseURL,
	}, nil
}

func (s *LocalStorage) normalizePath(path string) string {
	path = strings.TrimPrefix(path, "/")
	return filepath.FromSlash(path)
}

func (s *LocalStorage) resolvePath(path string) (string, error) {
	normalized := s.normalizePath(path)
	absPath := filepath.Join(s.rootDir, normalized)

	relPath, err := filepath.Rel(s.rootDir, absPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	if strings.HasPrefix(relPath, "..") {
		return "", errors.New("path outside storage root")
	}

	return absPath, nil
}

func (s *LocalStorage) Write(ctx context.Context, path string, data []byte) error {
	absPath, err := s.resolvePath(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *LocalStorage) WriteStream(ctx context.Context, path string, reader io.Reader) error {
	absPath, err := s.resolvePath(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(absPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func (s *LocalStorage) Read(ctx context.Context, path string) ([]byte, error) {
	absPath, err := s.resolvePath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func (s *LocalStorage) ReadStream(ctx context.Context, path string) (io.ReadCloser, error) {
	absPath, err := s.resolvePath(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	absPath, err := s.resolvePath(path)
	if err != nil {
		return err
	}

	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	absPath, err := s.resolvePath(path)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *LocalStorage) Stat(ctx context.Context, path string) (*FileInfo, error) {
	absPath, err := s.resolvePath(path)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)

	return &FileInfo{
		Path:      path,
		Size:      stat.Size(),
		MimeType:  detectMimeType(absPath, data),
		Hash:      hex.EncodeToString(hash[:]),
		UpdatedAt: stat.ModTime(),
	}, nil
}

func (s *LocalStorage) List(ctx context.Context, prefix string) ([]string, error) {
	absPath, err := s.resolvePath(prefix)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(s.rootDir, path)
			files = append(files, filepath.ToSlash(relPath))
		}
		return nil
	})

	return files, err
}

func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	normalized := s.normalizePath(path)
	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.baseURL, "/"), normalized), nil
	}
	return fmt.Sprintf("file://%s", filepath.Join(s.rootDir, normalized)), nil
}

func (s *LocalStorage) GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return s.GetURL(ctx, path)
}

func detectMimeType(path string, data []byte) string {
	mimeType := http.DetectContentType(data)
	if mimeType != "application/octet-stream" {
		return mimeType
	}

	ext := filepath.Ext(path)
	if ext != "" {
		mimeType = mime.TypeByExtension(ext)
		if mimeType != "" {
			return mimeType
		}
	}

	return "application/octet-stream"
}

func ImageToBase64(storage Storage, ctx context.Context, path string) (string, error) {
	data, err := storage.Read(ctx, path)
	if err != nil {
		return "", err
	}

	info, err := storage.Stat(ctx, path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("data:%s;base64,%s", info.MimeType, encodeBase64(data)), nil
}

func Base64ToImage(storage Storage, ctx context.Context, path string, base64Data string) error {
	var data []byte
	if strings.HasPrefix(base64Data, "data:") {
		idx := strings.Index(base64Data, ",")
		if idx == -1 {
			return errors.New("invalid Base64 data")
		}
		data = decodeBase64(base64Data[idx+1:])
	} else {
		data = decodeBase64(base64Data)
	}

	return storage.Write(ctx, path, data)
}

func DownloadFile(storage Storage, ctx context.Context, path string, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	return storage.WriteStream(ctx, path, resp.Body)
}

func Copy(storage Storage, ctx context.Context, src, dst string) error {
	data, err := storage.Read(ctx, src)
	if err != nil {
		return err
	}
	return storage.Write(ctx, dst, data)
}

func Move(storage Storage, ctx context.Context, src, dst string) error {
	if err := Copy(storage, ctx, src, dst); err != nil {
		return err
	}
	return storage.Delete(ctx, src)
}

func encodeBase64(data []byte) string {
	return hex.EncodeToString(data)
}

func decodeBase64(s string) []byte {
	data, _ := hex.DecodeString(s)
	return data
}

type BufferStorage struct {
	mu    sync.RWMutex
	files map[string]*bytes.Buffer
}

func NewBufferStorage() *BufferStorage {
	return &BufferStorage{
		files: make(map[string]*bytes.Buffer),
	}
}

func (s *BufferStorage) Write(ctx context.Context, path string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[path] = bytes.NewBuffer(data)
	return nil
}

func (s *BufferStorage) WriteStream(ctx context.Context, path string, reader io.Reader) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, reader)
	if err != nil {
		return err
	}
	s.files[path] = buf
	return nil
}

func (s *BufferStorage) Read(ctx context.Context, path string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	buf, exists := s.files[path]
	if !exists {
		return nil, os.ErrNotExist
	}
	return buf.Bytes(), nil
}

func (s *BufferStorage) ReadStream(ctx context.Context, path string) (io.ReadCloser, error) {
	data, err := s.Read(ctx, path)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *BufferStorage) Delete(ctx context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, path)
	return nil
}

func (s *BufferStorage) Exists(ctx context.Context, path string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.files[path]
	return exists, nil
}

func (s *BufferStorage) Stat(ctx context.Context, path string) (*FileInfo, error) {
	data, err := s.Read(ctx, path)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	return &FileInfo{
		Path: path,
		Size: int64(len(data)),
		Hash: hex.EncodeToString(hash[:]),
	}, nil
}

func (s *BufferStorage) List(ctx context.Context, prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []string
	for path := range s.files {
		if strings.HasPrefix(path, prefix) {
			result = append(result, path)
		}
	}
	return result, nil
}

func (s *BufferStorage) GetURL(ctx context.Context, path string) (string, error) {
	return fmt.Sprintf("memory://%s", path), nil
}

func (s *BufferStorage) GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return s.GetURL(ctx, path)
}
