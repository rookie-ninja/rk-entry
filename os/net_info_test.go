// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNetInfoHappyCase(t *testing.T) {
	netInfo := NewNetInfo()
	assert.NotNil(t, netInfo)
	assert.NotEmpty(t, netInfo.NetInterface)
}
