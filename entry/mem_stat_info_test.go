// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemStatsInfo_HappyCase(t *testing.T) {
	stats := NewMemStatsInfo()
	assert.NotNil(t, stats)
	assert.True(t, stats.MemAllocByte > 0)
	assert.True(t, stats.SysAllocByte > 0)
	assert.True(t, stats.MemPercentage > 0)
	assert.NotEmpty(t, stats.LastGCTimestamp)
}
