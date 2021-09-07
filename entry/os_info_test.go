// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

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
