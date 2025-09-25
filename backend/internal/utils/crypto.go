package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// CryptoService 提供 AES-256-GCM 加密/解密
type CryptoService struct {
	key []byte
}

// NewCryptoService 使用 SHA-256 从任意长度密钥派生 32 字节密钥
func NewCryptoService(secretKey string) *CryptoService {
	sum := sha256.Sum256([]byte(secretKey))
	return &CryptoService{key: sum[:]} // 32 bytes
}

// Encrypt 使用 AES-256-GCM 加密，输出 base64(nonce | ciphertext)
func (c *CryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("plaintext cannot be empty")
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 12字节随机 nonce（GCM 推荐长度）
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	out := append(nonce, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt 使用 AES-256-GCM 解密 base64(nonce | ciphertext)
func (c *CryptoService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", errors.New("ciphertext cannot be empty")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}

	nonce := data[:ns]
	enc := data[ns:]

	plain, err := gcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
