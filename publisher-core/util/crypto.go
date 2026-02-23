package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// PBKDF2迭代次数
	keyIteration = 100000
	// 密钥长度
	keyLength = 32
	// 盐值长度
	saltLength = 16
)

// CryptoManager 加密管理器
type CryptoManager struct {
	encryptionKey []byte
}

// NewCryptoManager 创建加密管理器
func NewCryptoManager(secret string) (*CryptoManager, error) {
	if secret == "" {
		secret = os.Getenv("ENCRYPTION_SECRET")
		if secret == "" {
			return nil, errors.New("encryption secret not provided")
		}
	}

	// 使用固定盐值从密钥派生加密密钥
	salt := []byte("publisher-core-salt")
	key := pbkdf2.Key([]byte(secret), salt, keyIteration, keyLength, sha256.New)

	return &CryptoManager{
		encryptionKey: key,
	}, nil
}

// Encrypt 加密数据
func (cm *CryptoManager) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// 创建GCM模式的加密器
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回Base64编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (cm *CryptoManager) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// 提取nonce和密文
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Hash 哈希数据（用于检测数据变化）
func (cm *CryptoManager) Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}

// GenerateRandomKey 生成随机密钥
func GenerateRandomKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// 全局加密管理器实例
var globalCryptoManager *CryptoManager

// InitCrypto 初始化加密管理器
func InitCrypto(secret string) error {
	cm, err := NewCryptoManager(secret)
	if err != nil {
		return err
	}
	globalCryptoManager = cm
	return nil
}

// GetCryptoManager 获取全局加密管理器
func GetCryptoManager() (*CryptoManager, error) {
	if globalCryptoManager == nil {
		return nil, errors.New("crypto manager not initialized")
	}
	return globalCryptoManager, nil
}
