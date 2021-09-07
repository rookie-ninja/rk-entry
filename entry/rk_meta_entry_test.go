// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithMetaRkMeta(t *testing.T) {
	meta := &rkcommon.RkMeta{}
	entry := RegisterRkMetaEntry(WithMetaRkMeta(meta))
	assert.Equal(t, meta, entry.RkMeta)

	GlobalAppCtx.rkMetaEntry = nil
}

func TestRegisterRkMetaEntriesFromConfig(t *testing.T) {
	config := `
---
git:
  branch: fake-branch
  commit:
    committer:
      email: fake@gmail.com
      name: fake
    date: Fri Sep 03 04:16:10 2021 +0800
    id: fake-id
    idAbbr: fake-id
    sub: fake-sub
  tag: ""
  url: fake-url
name: fake-name
version: fake-version
`

	filePath := createFileAtTestTempDir(t, config)
	entries := RegisterRkMetaEntriesFromConfig(filePath)
	assert.NotEmpty(t, entries)

	metaEntry := entries[RkMetaEntryName]
	assert.NotNil(t, metaEntry)

	GlobalAppCtx.rkMetaEntry = nil
}

func TestRegisterRkMetaEntry(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterRkMetaEntry()
	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
	assert.Equal(t, RkMetaEntryName, entry.GetName())
	assert.Equal(t, RkMetaEntryType, entry.GetType())
	assert.Equal(t, RkMetaEntryDescription, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	GlobalAppCtx.rkMetaEntry = nil
}
