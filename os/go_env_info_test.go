// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGoEnvInfo_HappyCase(t *testing.T) {
	info := NewGoEnvInfo()
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.GOOS)
	assert.NotEmpty(t, info.GOArch)
	assert.True(t, info.RoutinesCount >= 0)
	assert.NotEmpty(t, info.Version)
}
