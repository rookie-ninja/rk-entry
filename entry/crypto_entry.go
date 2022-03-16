package rkentry

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
)

func NewCryptoAES(entryName string, key []byte) (*CryptoAESEntry, error) {
	entry := &CryptoAESEntry{
		entryName: entryName,
		key:       key,
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "CryptoAES"
	}

	block, err := aes.NewCipher(entry.key)
	if err != nil {
		return nil, err
	}

	entry.block = block

	return entry, nil
}

type CryptoAESEntry struct {
	entryName string
	key       []byte
	block     cipher.Block
}

func (s *CryptoAESEntry) Bootstrap(ctx context.Context) {}

func (s *CryptoAESEntry) Interrupt(ctx context.Context) {}

func (s *CryptoAESEntry) GetName() string {
	return s.entryName
}

func (s *CryptoAESEntry) GetType() string {
	return CryptoEntryType
}

func (s *CryptoAESEntry) GetDescription() string {
	return "Symmetric crypto entry with AES"
}

func (s *CryptoAESEntry) String() string {
	m := map[string]string{
		"name":      s.entryName,
		"algorithm": "AES",
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

func (s *CryptoAESEntry) Encrypt(plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(s.block)
	if err != nil {
		return nil, err
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (s *CryptoAESEntry) Decrypt(ciphertext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(s.block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("Cipher text is too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
