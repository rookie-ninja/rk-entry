// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCpuInfo_HappyCase(t *testing.T) {
	info := NewCpuInfo()
	assert.NotNil(t, info)
	assert.True(t, info.CpuUsedPercentage >= 0)
	assert.True(t, info.LogicalCoreCount >= 0)
	assert.True(t, info.PhysicalCoreCount >= 0)
}
