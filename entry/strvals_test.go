// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBootConfigOverrides(t *testing.T) {
	// For maps
	res, err := parseBootOverrides("key1=value1,key2=value2")
	assert.Nil(t, err)
	assert.Equal(t, "value1", res["key1"])
	assert.Equal(t, "value2", res["key2"])

	// For slice
	res, err = parseBootOverrides("slice[0]=value0,slice[1]=value1")
	assert.Nil(t, err)
	assert.Equal(t, "value0", res["slice"].([]interface{})[0])
	assert.Equal(t, "value1", res["slice"].([]interface{})[1])

	// Mixed
	res, err = parseBootOverrides("key1=value1,slice[0]=value0")
	assert.Nil(t, err)
	assert.Equal(t, "value1", res["key1"])
	assert.Equal(t, "value0", res["slice"].([]interface{})[0])
}
