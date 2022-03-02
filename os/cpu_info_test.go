// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCpuInfo_HappyCase(t *testing.T) {
	cpuInfo := NewCpuInfo()
	assert.NotNil(t, cpuInfo)
	assert.True(t, cpuInfo.Count >= 0)
}
