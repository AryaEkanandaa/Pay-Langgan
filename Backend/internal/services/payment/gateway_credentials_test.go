package payment

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestCredentialVaultEncryptsAndDecrypts(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	vault, err := newCredentialVault(key)
	if err != nil {
		t.Fatal(err)
	}
	secret := []byte(`{"server_key":"keep-me-secret"}`)
	ciphertext, err := vault.seal(secret)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ciphertext, []byte("keep-me-secret")) {
		t.Fatal("secret appears in ciphertext")
	}
	plaintext, err := vault.open(ciphertext)
	if err != nil || !bytes.Equal(plaintext, secret) {
		t.Fatalf("decrypted credentials = %q, err = %v", plaintext, err)
	}
	wrongVault, err := newCredentialVault(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{8}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wrongVault.open(ciphertext); err == nil {
		t.Fatal("ciphertext decrypted with an unrelated key")
	}
}

func TestCredentialVaultRequiresDedicatedKey(t *testing.T) {
	if _, err := newCredentialVault(""); err == nil {
		t.Fatal("missing encryption key was accepted")
	}
	if _, err := newCredentialVault(base64.StdEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Fatal("short encryption key was accepted")
	}
}
