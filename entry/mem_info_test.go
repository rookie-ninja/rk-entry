// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemInfo_HappyCase(t *testing.T) {
	info := NewMemInfo()
	assert.NotNil(t, info)
	assert.True(t, info.MemAllocByte > 0)
	assert.True(t, info.SysAllocByte > 0)
	assert.True(t, info.MemUsedPercentage > 0)
	assert.NotEmpty(t, info.LastGcTimestamp)
}
