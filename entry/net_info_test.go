// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

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
