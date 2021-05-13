package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewOsInfo_HappyCase(t *testing.T) {
	info := NewOsInfo()
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.Os)
	assert.NotEmpty(t, info.Arch)
	assert.NotEmpty(t, info.Hostname)
}
