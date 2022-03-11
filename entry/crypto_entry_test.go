package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	key       = []byte("the-key-has-to-be-32-bytes-long!")
	plaintext = "ut-message"
)

func TestNewCryptoAES(t *testing.T) {
	// with invalid length
	crypto, err := NewCryptoAES("ut", []byte("invalid"))
	assert.Nil(t, crypto)
	assert.NotNil(t, err)

	// with valid length
	crypto, err = NewCryptoAES("ut", key)
	assert.NotNil(t, crypto)
	assert.Nil(t, err)
}

func TestCryptoAESEntry_Encrypt_And_Decrypt(t *testing.T) {
	crypto, err := NewCryptoAES("ut", key)
	assert.Nil(t, err)

	encrypted, err := crypto.Encrypt([]byte(plaintext))
	assert.Nil(t, err)

	decrypted, err := crypto.Decrypt(encrypted)
	assert.Nil(t, err)

	assert.Equal(t, plaintext, string(decrypted))
}

func TestCryptoAESEntry(t *testing.T) {
	crypto, err := NewCryptoAES("ut", key)
	assert.Nil(t, err)

	assert.NotEmpty(t, crypto.GetName())
	assert.Equal(t, CryptoEntryType, crypto.GetType())
	assert.NotEmpty(t, crypto.GetDescription())

	crypto.Bootstrap(context.TODO())
	crypto.Interrupt(context.TODO())
}
