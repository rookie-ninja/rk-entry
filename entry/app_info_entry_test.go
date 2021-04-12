// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestAppInfoEntryDefault_HappyCase(t *testing.T) {
	entry := AppInfoEntryDefault()

	// assert default values
	assert.NotNil(t, entry)
	assert.Equal(t, AppNameDefault, entry.AppName)
	assert.Equal(t, VersionDefault, entry.Version)
	assert.Equal(t, LangDefault, entry.Lang)
	assert.Equal(t, DescriptionDefault, entry.Description)
	assert.Empty(t, entry.Keywords)
	assert.Empty(t, entry.HomeURL)
	assert.Empty(t, entry.IconURL)
	assert.Empty(t, entry.DocsURL)
	assert.Empty(t, entry.Maintainers)
}

func TestWithAppNameAppInfo_WithEmptyString(t *testing.T) {
	entry := RegisterAppInfoEntry(WithAppNameAppInfo(""))
	assert.NotNil(t, entry)
	assert.Equal(t, AppNameDefault, entry.AppName)
}

func TestWithAppNameAppInfo_HappyCase(t *testing.T) {
	appName := "unit-test-app"
	entry := RegisterAppInfoEntry(WithAppNameAppInfo(appName))
	assert.NotNil(t, entry)
	assert.Equal(t, appName, entry.AppName)
}

func TestWithVersionAppInfo_WithEmptyString(t *testing.T) {
	entry := RegisterAppInfoEntry(WithVersionAppInfo(""))
	assert.NotNil(t, entry)
	assert.Equal(t, VersionDefault, entry.Version)
}

func TestWithVersionAppInfo_HappyCase(t *testing.T) {
	version := "v-unit-test"
	entry := RegisterAppInfoEntry(WithVersionAppInfo(version))
	assert.NotNil(t, entry)
	assert.Equal(t, version, entry.Version)
}

func TestWithDescriptionAppInfo_WithEmptyString(t *testing.T) {
	entry := RegisterAppInfoEntry(WithDescriptionAppInfo(""))
	assert.NotNil(t, entry)
	assert.Equal(t, DescriptionDefault, entry.Description)
}

func TestWithDescriptionAppInfo_HappyCase(t *testing.T) {
	description := "unit-test-description"
	entry := RegisterAppInfoEntry(WithDescriptionAppInfo(description))
	assert.NotNil(t, entry)
	assert.Equal(t, description, entry.Description)
}

func TestWithKeywordsAppInfo_WithEmptySlice(t *testing.T) {
	entry := RegisterAppInfoEntry(WithKeywordsAppInfo())
	assert.NotNil(t, entry)
	assert.Empty(t, entry.Keywords)
}

func TestWithKeywordsAppInfo_HappyCase(t *testing.T) {
	one := "unit-test-one"
	two := "unit-test-two"

	entry := RegisterAppInfoEntry(WithKeywordsAppInfo(one, two))
	assert.NotNil(t, entry)
	assert.Len(t, entry.Keywords, 2)
}

func TestWithHomeURLAppInfo_WithEmptyString(t *testing.T) {
	entry := RegisterAppInfoEntry(WithHomeURLAppInfo(""))
	assert.NotNil(t, entry)
	assert.Empty(t, entry.HomeURL)
}

func TestWithHomeURLAppInfo_HappyCase(t *testing.T) {
	homeURL := "unit-test-home-URL"
	entry := RegisterAppInfoEntry(WithHomeURLAppInfo(homeURL))
	assert.NotNil(t, entry)
	assert.Equal(t, homeURL, entry.HomeURL)
}

func TestWithIconURLAppInfo_WithEmptyString(t *testing.T) {
	entry := RegisterAppInfoEntry(WithIconURLAppInfo(""))
	assert.NotNil(t, entry)
	assert.Empty(t, entry.IconURL)
}

func TestWithIconURLAppInfo_HappyCase(t *testing.T) {
	iconURL := "unit-test-icon-URL"
	entry := RegisterAppInfoEntry(WithIconURLAppInfo(iconURL))
	assert.NotNil(t, entry)
	assert.Equal(t, iconURL, entry.IconURL)
}

func TestWithDocsURLAppInfo_WithEmptySlice(t *testing.T) {
	entry := RegisterAppInfoEntry(WithDocsURLAppInfo())
	assert.NotNil(t, entry)
	assert.Empty(t, entry.DocsURL)
}

func TestWithDocsURLAppInfo_HappyCase(t *testing.T) {
	one := "unit-test-one"
	two := "unit-test-two"

	entry := RegisterAppInfoEntry(WithDocsURLAppInfo(one, two))
	assert.NotNil(t, entry)
	assert.Len(t, entry.DocsURL, 2)
}

func TestWithMaintainersAppInfo_WithEmptySlice(t *testing.T) {
	entry := RegisterAppInfoEntry(WithMaintainersAppInfo())
	assert.NotNil(t, entry)
	assert.Empty(t, entry.Maintainers)
}

func TestWithMaintainersAppInfo_HappyCase(t *testing.T) {
	one := "unit-test-one"
	two := "unit-test-two"

	entry := RegisterAppInfoEntry(WithMaintainersAppInfo(one, two))
	assert.NotNil(t, entry)
	assert.Len(t, entry.Maintainers, 2)
}

func TestRegisterAppInfoEntriesFromConfig_WithNonExistConfigFile(t *testing.T) {
	defer assertPanic(t)
	RegisterAppInfoEntriesFromConfig("invalid-path")
}

func TestRegisterAppInfoEntriesFromConfig_WithInvalidConfigFileExtension(t *testing.T) {
	defer assertPanic(t)
	RegisterAppInfoEntriesFromConfig("invalid-path.invalid")
}

func TestRegisterAppInfoEntriesFromConfig_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
rk:
  appName: ut-app
  version: ut-version
  description: ut-description
  homeURL: ut-homeURL
  iconURL: ut-iconURL
  keywords: ["ut-keyword"]
  maintainers: ["ut-maintainer"]
  docsURL: ["ut-docURL"]
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterAppInfoEntriesFromConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToAppInfoEntry(t, entries[AppInfoEntryName])

	assert.Equal(t, "ut-app", entry.AppName)
	assert.Equal(t, "ut-version", entry.Version)
	assert.Equal(t, "ut-description", entry.Description)
	assert.Equal(t, "ut-homeURL", entry.HomeURL)
	assert.Equal(t, "ut-iconURL", entry.IconURL)
	assert.Contains(t, entry.Keywords, "ut-keyword")
	assert.Contains(t, entry.Maintainers, "ut-maintainer")
	assert.Contains(t, entry.DocsURL, "ut-docURL")
}

func TestRegisterAppInfoEntriesFromConfig_WithInvalidElementType(t *testing.T) {
	defer assertPanic(t)

	configFile := `
---
rk:
  appName: ut-app
  version: ut-version
  description: ut-description
  homeURL: ut-homeURL
  iconURL: ut-iconURL
  keywords: "ut-keyword" # this should be a string slice
  maintainers: ["ut-maintainer"]
  docsURL: ["ut-docURL"]
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	RegisterAppInfoEntriesFromConfig(configFilePath)
}

func TestRegisterAppInfoEntriesFromConfig_WithoutElements(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
rk:
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterAppInfoEntriesFromConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToAppInfoEntry(t, entries[AppInfoEntryName])

	assert.Equal(t, AppNameDefault, entry.AppName)
	assert.Equal(t, VersionDefault, entry.Version)
	assert.Equal(t, DescriptionDefault, entry.Description)
	assert.Empty(t, entry.HomeURL)
	assert.Empty(t, entry.IconURL)
	assert.Empty(t, entry.Keywords)
	assert.Empty(t, entry.Maintainers)
	assert.Empty(t, entry.DocsURL)
}

func TestRegisterAppInfoEntriesFromConfig_WithoutRKSection(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterAppInfoEntriesFromConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToAppInfoEntry(t, entries[AppInfoEntryName])

	assert.Equal(t, AppNameDefault, entry.AppName)
	assert.Equal(t, VersionDefault, entry.Version)
	assert.Equal(t, DescriptionDefault, entry.Description)
	assert.Empty(t, entry.HomeURL)
	assert.Empty(t, entry.IconURL)
	assert.Empty(t, entry.Keywords)
	assert.Empty(t, entry.Maintainers)
	assert.Empty(t, entry.DocsURL)
}

func TestRegisterAppInfoEntriesFromConfig_WithEmptyElements(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
rk:
  appName:
  version:
  description:
  homeURL:
  iconURL:
  keywords:
  maintainers:
  docsURL:
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterAppInfoEntriesFromConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToAppInfoEntry(t, entries[AppInfoEntryName])

	assert.Equal(t, AppNameDefault, entry.AppName)
	assert.Equal(t, VersionDefault, entry.Version)
	assert.Equal(t, DescriptionDefault, entry.Description)
	assert.Empty(t, entry.HomeURL)
	assert.Empty(t, entry.IconURL)
	assert.Empty(t, entry.Keywords)
	assert.Empty(t, entry.Maintainers)
	assert.Empty(t, entry.DocsURL)
}

func TestRegisterAppInfoEntry_WithoutOptions(t *testing.T) {
	entry := RegisterAppInfoEntry()
	assert.NotNil(t, entry)

	assert.Equal(t, AppNameDefault, entry.AppName)
	assert.Equal(t, VersionDefault, entry.Version)
	assert.Equal(t, DescriptionDefault, entry.Description)
	assert.Empty(t, entry.HomeURL)
	assert.Empty(t, entry.IconURL)
	assert.Empty(t, entry.Keywords)
	assert.Empty(t, entry.Maintainers)
	assert.Empty(t, entry.DocsURL)
}

func TestRegisterAppInfoEntry_WithEmptyElements(t *testing.T) {
	entry := RegisterAppInfoEntry(
		WithAppNameAppInfo(""),
		WithVersionAppInfo(""),
		WithDescriptionAppInfo(""))

	assert.NotNil(t, entry)

	assert.Equal(t, AppNameDefault, entry.AppName)
	assert.Equal(t, VersionDefault, entry.Version)
	assert.Equal(t, DescriptionDefault, entry.Description)
	assert.Empty(t, entry.HomeURL)
	assert.Empty(t, entry.IconURL)
	assert.Empty(t, entry.Keywords)
	assert.Empty(t, entry.Maintainers)
	assert.Empty(t, entry.DocsURL)
}

func TestRegisterAppInfoEntry_HappyCase(t *testing.T) {
	entry := RegisterAppInfoEntry(
		WithAppNameAppInfo("ut-app"),
		WithVersionAppInfo("ut-version"),
		WithDescriptionAppInfo("ut-description"),
		WithHomeURLAppInfo("ut-homeURL"),
		WithIconURLAppInfo("ut-iconURL"),
		WithKeywordsAppInfo("ut-keyword"),
		WithDocsURLAppInfo("ut-docURL"),
		WithMaintainersAppInfo("ut-maintainer"))

	assert.NotNil(t, entry)

	assert.Equal(t, "ut-app", entry.AppName)
	assert.Equal(t, "ut-version", entry.Version)
	assert.Equal(t, "ut-description", entry.Description)
	assert.Equal(t, "ut-homeURL", entry.HomeURL)
	assert.Equal(t, "ut-iconURL", entry.IconURL)
	assert.Contains(t, entry.Keywords, "ut-keyword")
	assert.Contains(t, entry.DocsURL, "ut-docURL")
	assert.Contains(t, entry.Maintainers, "ut-maintainer")
}

func TestAppInfoEntry_Bootstrap_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	RegisterAppInfoEntry().Bootstrap(context.Background())
}

func TestAppInfoEntry_Interrupt_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	RegisterAppInfoEntry().Interrupt(context.Background())
}

func TestAppInfoEntry_GetName_HappyCase(t *testing.T) {
	assert.Equal(t, AppInfoEntryName, RegisterAppInfoEntry().GetName())
}

func TestAppInfoEntry_GetType_HappyCase(t *testing.T) {
	assert.Equal(t, AppInfoEntryType, RegisterAppInfoEntry().GetType())
}

func TestAppInfoEntry_String_HappyCase(t *testing.T) {
	entry := RegisterAppInfoEntry()
	str := entry.String()

	m := make(map[string]interface{})

	// assert unmarshalling without error
	assert.Nil(t, json.Unmarshal([]byte(str), &m))

	assert.Equal(t, AppInfoEntryName, m["entry_name"])
	assert.Equal(t, AppInfoEntryType, m["entry_type"])
	assert.Equal(t, entry.AppName, m["app_name"])
	assert.Equal(t, entry.Version, m["version"])
	assert.Equal(t, entry.Lang, m["lang"])
	assert.Equal(t, entry.Description, m["description"])
	assert.Equal(t, entry.HomeURL, m["home_url"])
	assert.Equal(t, entry.IconURL, m["icon_url"])
	assert.Empty(t, entry.Keywords, m["keywords"])
	assert.Empty(t, entry.DocsURL, m["docs_url"])
	assert.Empty(t, entry.Maintainers, m["maintainers"])
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
		// expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// this should never be called in case of a bug
		assert.True(t, true)
	}
}

func createFileAtTestTempDir(t *testing.T, content string) string {
	tempDir := path.Join(t.TempDir(), "ut-boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(content), os.ModePerm))
	return tempDir
}

func convertToAppInfoEntry(t *testing.T, raw Entry) *AppInfoEntry {
	entry, ok := raw.(*AppInfoEntry)
	assert.True(t, ok)
	return entry
}
