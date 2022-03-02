// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppInfoEntryDefault_HappyCase(t *testing.T) {
	entry := appInfoEntryDefault()

	// Assert default values
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.AppName)
	assert.NotEmpty(t, entry.Version)
	assert.NotEmpty(t, entry.Lang)
	assert.NotEmpty(t, entry.entryDescription)
	assert.Empty(t, entry.Keywords)
	assert.Empty(t, entry.HomeUrl)
	assert.Empty(t, entry.DocsUrl)
	assert.Empty(t, entry.Maintainers)
}

func TestRegisterAppInfoEntry(t *testing.T) {
	bootStr := `
---
app:
  name: rk
  version: version
  description: desc
  homeUrl: ut-homeUrl
  docsUrl: ["ut-docUrl"]
  keywords: ["ut-keyword"]
  maintainers: ["ut-maintainer"]
`

	entries := registerAppInfoEntry([]byte(bootStr))
	entry := entries[appInfoEntryName].(*appInfoEntry)

	assert.Equal(t, "rk", entry.AppName)
	assert.Equal(t, "version", entry.Version)
	assert.Equal(t, "desc", entry.GetDescription())
	assert.Equal(t, "ut-homeUrl", entry.HomeUrl)
	assert.Contains(t, entry.DocsUrl, "ut-docUrl")
	assert.Contains(t, entry.Maintainers, "ut-maintainer")
	assert.Contains(t, entry.Keywords, "ut-keyword")
	assert.NotEmpty(t, entry.String())
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
}

func TestAppInfoEntry_UnmarshalJSON(t *testing.T) {
	defer assertNotPanic(t)

	GlobalAppCtx.GetAppInfoEntry().UnmarshalJSON(nil)
}

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// this should never be called in case of a bug
		assert.True(t, false)
	}
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}
