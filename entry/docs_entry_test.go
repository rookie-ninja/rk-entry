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

func TestRegisterDocsEntry(t *testing.T) {
	entry := RegisterDocsEntry(&BootDocs{
		Enabled:  true,
		Path:     "ut-path",
		SpecPath: "ut-spec-path",
		Headers:  []string{"a:b"},
	})

	assert.Equal(t, "/ut-path/", entry.Path)
	assert.Equal(t, "ut-spec-path", entry.SpecPath)
	assert.Len(t, entry.Headers, 1)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
}

func TestDocsEntry_Bootstrap_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterDocsEntry(&BootDocs{
		Enabled:  true,
		Path:     "ut-path",
		SpecPath: "ut-spec-path",
		Headers:  []string{"a:b"},
	})

	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestDocsEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterDocsEntry(&BootDocs{
		Enabled:  true,
		Path:     "ut-path",
		SpecPath: "ut-spec-path",
		Headers:  []string{"a:b"},
	})
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestDocsEntry_ConfigFileHandler(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterDocsEntry(&BootDocs{
		Enabled:  true,
		Path:     "ut-path",
		SpecPath: "ut-spec-path",
		Headers:  []string{"a:b"},
	})

	writer := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rk/v1/spec", nil)

	entry.ConfigFileHandler()(writer, req)
}
