// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewViperConfigInfo_WithEmptyViperConfigs(t *testing.T) {
	assert.Empty(t, NewViperConfigInfo())
}

func TestNewViperConfigInfo_HappyCase(t *testing.T) {
	RegisterViperEntry(
		WithNameViper("unit-test-config"),
		WithViperInstanceViper(viper.New()))

	assert.Len(t, NewViperConfigInfo(), 1)

	// clear viper config in app context
	GlobalAppCtx.clearViperEntries()
}
