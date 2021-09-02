// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBasicInfoToStruct_HappyCase(t *testing.T) {
	assert.Nil(t, os.Setenv("REALM", "unit-test-realm"))
	assert.Nil(t, os.Setenv("REGION", "unit-test-region"))
	assert.Nil(t, os.Setenv("AZ", "unit-test-az"))
	assert.Nil(t, os.Setenv("DOMAIN", "unit-test-domain"))

	info := NewProcessInfo()
	assert.NotNil(t, info)

	assert.NotEmpty(t, info.Username)
	assert.NotEmpty(t, info.UID)
	assert.NotEmpty(t, info.GID)
	assert.NotEmpty(t, info.StartTime)
	assert.True(t, info.UpTimeSec >= 0)
	assert.NotEmpty(t, info.UpTimeStr)
	assert.Equal(t, "unit-test-realm", info.Realm)
	assert.Equal(t, "unit-test-region", info.Region)
	assert.Equal(t, "unit-test-az", info.AZ)
	assert.Equal(t, "unit-test-domain", info.Domain)
}
