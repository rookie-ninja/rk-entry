// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestRegisterConfigEntriesWithConfig_WithoutElement(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
config:
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterConfigEntriesWithConfig_WithoutName(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
config:
  - path: ut-path
    locale: "*::*::*::*"
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterConfigEntriesWithConfig_WithoutPath(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
config:
  - name: unit-test-config
    locale: "*::*::*::*"
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
}

func TestRegisterConfigEntriesWithConfig_WithoutEnvPrefix(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
config:
  - name: unit-test-config
    locale: "*::*::*::*"
    envPrefix: rk
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	assert.Equal(t, "rk", entries["unit-test-config"].(*ConfigEntry).EnvPrefix)
}

func TestRegisterConfigEntriesWithConfig_WithNonExistPath(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
config:
  - name: unit-test-viper
    path: non-exist-path
    locale: "*::*::*::*"
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
}

func TestRegisterConfigEntriesWithConfig_WithDomainAndFileNotExist(t *testing.T) {
	defer assertNotPanic(t)
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := path.Join(t.TempDir(), "ut-viper.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	configFile := `
---
config:
  - name: unit-test-viper
    path: %s
    locale: "*::*::*::*"
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetConfigEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterConfigEntriesWithConfig_WithDomainAndFileExist(t *testing.T) {
	defer assertNotPanic(t)
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := path.Join(t.TempDir(), "ut-viper-prod.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	configFile := `
---
config:
  - name: unit-test-viper
    path: %s
    locale: "*::*::*::prod"
`
	// override path
	configFile = fmt.Sprintf(configFile, path.Join(path.Dir(tempDir), "ut-viper-prod.yaml"))

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetConfigEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterConfigEntriesWithConfig_WithDomainAndBothFileExist(t *testing.T) {
	defer assertNotPanic(t)

	// create default viper config file named as ut-viper.yaml
	viperConfigBeta := `
---
key: beta
`
	// create viper config file in ut temp dir
	tempDirBeta := path.Join(t.TempDir(), "ut-viper-beta.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDirBeta, []byte(viperConfigBeta), os.ModePerm))

	// create prod viper config file named as ut-viper-prod.yaml
	viperConfigProd := `
---
key: prod
`
	// create viper config file in ut temp dir
	tempDirProd := path.Join(path.Dir(tempDirBeta), "ut-viper-prod.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDirProd, []byte(viperConfigProd), os.ModePerm))

	configFile := `
---
config:
  - name: unit-test-beta
    path: %s
    locale: "*::*::*::beta"
  - name: unit-test-prod
    path: %s
    locale: "*::*::*::prod"
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDirBeta, tempDirProd)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetConfigEntry("unit-test-prod")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "prod", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterConfigEntriesWithConfig_WithoutDomainAndBothFileExist(t *testing.T) {
	defer assertNotPanic(t)

	// create default viper config file named as ut-viper.yaml
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := path.Join(t.TempDir(), "ut-viper.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	// create prod viper config file named as ut-viper-prod.yaml
	viperConfigProd := `
---
key: prod
`
	// create viper config file in ut temp dir
	tempDirProd := path.Join(path.Dir(tempDir), "ut-viper-prod.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDirProd, []byte(viperConfigProd), os.ModePerm))

	configFile := `
---
config:
  - name: unit-test-viper
    path: %s
    locale: "*::*::*::*"
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "test"))

	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetConfigEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterConfigEntriesWithConfig_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := path.Join(t.TempDir(), "ut-viper.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	configFile := `
---
config:
  - name: unit-test-viper
    path: %s
    locale: "*::*::*::*"
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetConfigEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
}

func TestRegisterConfigEntry_WithoutOptions(t *testing.T) {
	entry := RegisterConfigEntry()

	assert.NotNil(t, entry)

	// validate default fields
	assert.Contains(t, entry.EntryName, "config-")
	assert.Equal(t, ConfigEntryType, entry.EntryType)

	// validate viper instance
	assert.NotNil(t, entry.vp)

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
}

func TestRegisterConfigEntry_WithInvalidFile(t *testing.T) {
	defer assertPanic(t)
	filePath := createFileAtTestTempDir(t, "invalid")
	RegisterConfigEntry(WithPathConfig(filePath))
}

func TestRegisterConfigEntry_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	description := "unit-test-description"
	vp := viper.New()

	entry := RegisterConfigEntry(
		WithNameConfig(name),
		WithDescriptionConfig(description),
		WithViperInstanceConfig(vp))

	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, name, entry.EntryName)
	assert.Equal(t, description, entry.GetDescription())
	assert.Equal(t, ConfigEntryType, entry.EntryType)

	// validate viper instance
	assert.Equal(t, vp, entry.vp)
	assert.Empty(t, entry.Path)

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
}

func TestConfigEntry_GetDescription_HappyCase(t *testing.T) {
	entry := RegisterConfigEntry(
		WithDescriptionConfig("ut-description"))
	assert.Equal(t, "ut-description", entry.GetDescription())
}

func TestConfigEntry_GetViperAsMap_HappyCase(t *testing.T) {
	vp := viper.New()
	vp.Set("key", "value")
	entry := RegisterConfigEntry(WithViperInstanceConfig(vp))

	res := entry.GetViperAsMap()
	assert.Equal(t, "value", res["key"])
}

func TestConfigEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterConfigEntry()
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestConfigEntry_GetLocale(t *testing.T) {
	entry := RegisterConfigEntry(WithLocaleConfig("ut-locale"))
	assert.Equal(t, "ut-locale", entry.GetLocale())
}

func TestConfigEntry_GetViper_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	vp := viper.New()

	entry := RegisterConfigEntry(
		WithNameConfig(name),
		WithViperInstanceConfig(vp))

	assert.NotNil(t, entry)

	// validate viper instance
	assert.Equal(t, vp, entry.GetViper())

	// clear viper entry
	GlobalAppCtx.clearConfigEntries()
}

func TestConfigEntry_Bootstrap_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterConfigEntry().Bootstrap(context.Background())

	GlobalAppCtx.clearConfigEntries()
}

func TestConfigEntry_Interrupt_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterConfigEntry().Interrupt(context.Background())

	GlobalAppCtx.clearConfigEntries()
}

func TestConfigEntry_GetName_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	vp := viper.New()

	entry := RegisterConfigEntry(
		WithNameConfig(name),
		WithViperInstanceConfig(vp))

	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, name, entry.GetName())

	GlobalAppCtx.clearConfigEntries()
}

func TestConfigEntry_GetType_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	vp := viper.New()

	entry := RegisterConfigEntry(
		WithNameConfig(name),
		WithViperInstanceConfig(vp))

	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, ConfigEntryType, entry.GetType())

	GlobalAppCtx.clearConfigEntries()
}

func TestConfigEntry_String_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := path.Join(t.TempDir(), "ut-viper.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	configFile := `
---
config:
  - name: unit-test-viper
    path: %s
    locale: "*::*::*::*"
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterConfigEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetConfigEntry("unit-test-viper")
	assert.NotNil(t, entry)

	m := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal([]byte(entry.String()), &m))

	assert.Contains(t, m, "entryName")
	assert.Contains(t, m, "entryType")
	assert.Contains(t, m, "path")

	GlobalAppCtx.clearConfigEntries()
}
