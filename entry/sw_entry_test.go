// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterSWEntry(t *testing.T) {
	entry := RegisterSWEntry(&BootSW{
		Enabled:   true,
		Path:      "ut-path",
		JsonPaths: []string{"ut-json-path"},
		Headers:   []string{"a:b"},
	})

	assert.Equal(t, "/ut-path/", entry.Path)
	assert.Equal(t, "ut-json-path", entry.JsonPaths)
	assert.Len(t, entry.Headers, 1)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
}

func TestSwEntry_Bootstrap_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterSWEntry(&BootSW{
		Enabled:   true,
		Path:      "ut-path",
		JsonPaths: []string{"ut-json-path"},
		Headers:   []string{"a:b"},
	})

	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestSWEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterSWEntry(&BootSW{
		Enabled:   true,
		Path:      "ut-path",
		JsonPaths: []string{"ut-json-path"},
		Headers:   []string{"a:b"},
	})
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestSWEntry_ConfigFileHandler(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterSWEntry(&BootSW{
		Enabled:   true,
		Path:      "ut-path",
		JsonPaths: []string{"ut-json-path"},
		Headers:   []string{"a:b"},
	})

	writer := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rk/v1/rk-common-swagger.json", nil)

	entry.ConfigFileHandler()(writer, req)
}
