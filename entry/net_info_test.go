package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNetInfoHappyCase(t *testing.T) {
	info := NewNetInfo()
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.NetInterface)
}
