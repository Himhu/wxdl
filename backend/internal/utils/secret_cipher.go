package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// SecretCipher 配置加密接口
type SecretCipher interface {
	Encrypt(plaintext, aad string) (ciphertext string, masked string, checksum string, keyVersion string, err error)
	Decrypt(ciphertext, aad string) (string, error)
}

// CipherEnvelope 加密信封格式
type CipherEnvelope struct {
	Algorithm  string `json:"alg"`
	KeyVersion string `json:"kid"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

type secretCipher struct {
	masterKey  []byte
	keyVersion string
}

type disabledSecretCipher struct {
	reason string
}

// NewDisabledSecretCipher creates a disabled cipher to avoid nil panic.
func NewDisabledSecretCipher(reason string) SecretCipher {
	if reason == "" {
		reason = "secret cipher is disabled"
	}
	return &disabledSecretCipher{reason: reason}
}

func (c *disabledSecretCipher) Encrypt(plaintext, aad string) (string, string, string, string, error) {
	return "", "", "", "", errors.New(c.reason)
}

func (c *disabledSecretCipher) Decrypt(ciphertext, aad string) (string, error) {
	return "", errors.New(c.reason)
}

// NewSecretCipher 创建加密工具
func NewSecretCipher(masterKeyBase64, keyVersion string) (SecretCipher, error) {
	if masterKeyBase64 == "" {
		return nil, fmt.Errorf("master key is required")
	}

	masterKey, err := base64.StdEncoding.DecodeString(masterKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid master key format: %w", err)
	}

	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes")
	}

	if keyVersion == "" {
		keyVersion = "v1"
	}

	return &secretCipher{
		masterKey:  masterKey,
		keyVersion: keyVersion,
	}, nil
}

// Encrypt 加密明文
func (c *secretCipher) Encrypt(plaintext, aad string) (string, string, string, string, error) {
	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return "", "", "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", "", "", err
	}

	ciphertextBytes := gcm.Seal(nil, nonce, []byte(plaintext), []byte(aad))

	envelope := CipherEnvelope{
		Algorithm:  "AES-256-GCM",
		KeyVersion: c.keyVersion,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertextBytes),
	}

	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		return "", "", "", "", err
	}

	masked := maskSecret(plaintext)
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(plaintext)))

	return string(envelopeJSON), masked, checksum, c.keyVersion, nil
}

// Decrypt 解密密文
func (c *secretCipher) Decrypt(ciphertext, aad string) (string, error) {
	var envelope CipherEnvelope
	if err := json.Unmarshal([]byte(ciphertext), &envelope); err != nil {
		return "", fmt.Errorf("invalid cipher envelope: %w", err)
	}

	if envelope.Algorithm != "AES-256-GCM" {
		return "", fmt.Errorf("unsupported algorithm: %s", envelope.Algorithm)
	}

	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return "", fmt.Errorf("invalid nonce: %w", err)
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %w", err)
	}

	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, []byte(aad))
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// maskSecret 脱敏处理
func maskSecret(secret string) string {
	length := len(secret)
	if length <= 4 {
		return "****"
	}
	if length <= 8 {
		return secret[:2] + "****" + secret[length-2:]
	}
	return secret[:4] + "************" + secret[length-4:]
}
