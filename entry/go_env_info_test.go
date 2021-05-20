// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGoEnvInfo_HappyCase(t *testing.T) {
	info := NewGoEnvInfo()
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.GOOS)
	assert.NotEmpty(t, info.GOArch)
	assert.NotEmpty(t, info.StartTime)
	assert.True(t, info.UpTimeSec >= 0)
	assert.NotEmpty(t, info.UpTimeStr)
	assert.True(t, info.RoutinesCount >= 0)
	assert.NotEmpty(t, info.Version)
}
