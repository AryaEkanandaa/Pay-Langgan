package payment

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// credentialVault encrypts provider secrets before they are persisted.
type credentialVault struct {
	aead cipher.AEAD
}

func newCredentialVault(encodedKey string) (*credentialVault, error) {
	if encodedKey == "" {
		return nil, errors.New("payment credential encryption key is not configured")
	}
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != 32 {
		return nil, errors.New("payment credential encryption key must be base64-encoded 32-byte key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("initialize payment credential encryption: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize payment credential encryption: %w", err)
	}
	return &credentialVault{aead: aead}, nil
}

func (v *credentialVault) seal(plaintext []byte) ([]byte, error) {
	if v == nil || v.aead == nil {
		return nil, errors.New("payment credential encryption is unavailable")
	}
	nonce := make([]byte, v.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("create encryption nonce: %w", err)
	}
	return v.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (v *credentialVault) open(ciphertext []byte) ([]byte, error) {
	if v == nil || v.aead == nil {
		return nil, errors.New("payment credential encryption is unavailable")
	}
	nonceSize := v.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("encrypted payment credentials are invalid")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := v.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("could not decrypt payment credentials")
	}
	return plaintext, nil
}
