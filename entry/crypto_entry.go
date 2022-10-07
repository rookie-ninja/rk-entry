package rk

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"embed"
	"errors"
	"io"
)

const CryptoKind = "crypto"

func NewCryptoAES(name string, key []byte) (*CryptoAESEntry, error) {
	entry := &CryptoAESEntry{
		name: name,
		key:  key,
	}

	if len(entry.name) < 1 {
		entry.name = "CryptoAES"
	}

	block, err := aes.NewCipher(entry.key)
	if err != nil {
		return nil, err
	}

	entry.block = block

	return entry, nil
}

type CryptoAESEntry struct {
	name  string
	key   []byte
	block cipher.Block
}

func (s *CryptoAESEntry) Category() string {
	return CategoryInline
}

func (s *CryptoAESEntry) Kind() string {
	return "crypto"
}

func (s *CryptoAESEntry) Name() string {
	return s.name
}

func (s *CryptoAESEntry) Config() EntryConfig {
	return nil
}

func (s *CryptoAESEntry) Monitor() *Monitor {
	return nil
}

func (s *CryptoAESEntry) FS() *embed.FS {
	return nil
}

func (s *CryptoAESEntry) Apis() []*BuiltinApi {
	return []*BuiltinApi{}
}

func (s *CryptoAESEntry) Bootstrap(ctx context.Context) {}

func (s *CryptoAESEntry) Interrupt(ctx context.Context) {}

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
