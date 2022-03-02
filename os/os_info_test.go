// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewOsInfo_HappyCase(t *testing.T) {
	osInfo := NewOsInfo()
	assert.NotNil(t, osInfo)
	assert.NotEmpty(t, osInfo.Os)
	assert.NotEmpty(t, osInfo.Arch)
	assert.NotEmpty(t, osInfo.Hostname)
}
