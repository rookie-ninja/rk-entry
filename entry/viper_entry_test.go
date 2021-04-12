// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
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

func TestRegisterViperEntriesWithConfig_WithoutElement(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
viper:
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterViperEntriesWithConfig_WithoutName(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
viper:
  - path: ut-path
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterViperEntriesWithConfig_WithoutPath(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
viper:
  - name: unit-test-viper
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterViperEntriesWithConfig_WithNonExistPath(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
viper:
  - name: unit-test-viper
    path: non-exist-path
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterViperEntriesWithConfig_WithDomainAndFileNotExist(t *testing.T) {
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
viper:
  - name: unit-test-viper
    path: %s
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetViperEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterViperEntriesWithConfig_WithDomainAndFileExist(t *testing.T) {
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
viper:
  - name: unit-test-viper
    path: %s
`
	// override path
	configFile = fmt.Sprintf(configFile, path.Join(path.Dir(tempDir), "ut-viper.yaml"))

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetViperEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterViperEntriesWithConfig_WithDomainAndBothFileExist(t *testing.T) {
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
viper:
  - name: unit-test-viper
    path: %s
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetViperEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "prod", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterViperEntriesWithConfig_WithoutDomainAndBothFileExist(t *testing.T) {
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
viper:
  - name: unit-test-viper
    path: %s
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "test"))

	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetViperEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterViperEntriesWithConfig_HappyCase(t *testing.T) {
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
viper:
  - name: unit-test-viper
    path: %s
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetViperEntry("unit-test-viper")
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.GetViper())
	assert.Equal(t, "value", entry.GetViper().GetString("key"))

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
}

func TestRegisterViperEntry_WithoutOptions(t *testing.T) {
	entry := RegisterViperEntry()

	assert.NotNil(t, entry)

	// validate default fields
	assert.Contains(t, entry.entryName, "viper-")
	assert.Equal(t, ViperEntryType, entry.entryType)

	// validate viper instance
	assert.NotNil(t, entry.vp)
	assert.Empty(t, entry.path)

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
}

func TestRegisterViperEntry_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	vp := viper.New()

	entry := RegisterViperEntry(
		WithNameViper(name),
		WithViperInstanceViper(vp))

	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, name, entry.entryName)
	assert.Equal(t, ViperEntryType, entry.entryType)

	// validate viper instance
	assert.Equal(t, vp, entry.vp)
	assert.Empty(t, entry.path)

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
}

func TestViperEntry_GetViper_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	vp := viper.New()

	entry := RegisterViperEntry(
		WithNameViper(name),
		WithViperInstanceViper(vp))

	assert.NotNil(t, entry)

	// validate viper instance
	assert.Equal(t, vp, entry.GetViper())

	// clear viper entry
	GlobalAppCtx.clearViperEntries()
}

func TestViperEntry_Bootstrap_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterViperEntry().Bootstrap(context.Background())
}

func TestViperEntry_Interrupt_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterViperEntry().Interrupt(context.Background())
}

func TestViperEntry_GetName_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	vp := viper.New()

	entry := RegisterViperEntry(
		WithNameViper(name),
		WithViperInstanceViper(vp))

	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, name, entry.GetName())
}

func TestViperEntry_GetType_HappyCase(t *testing.T) {
	name := "unit-test-viper"
	vp := viper.New()

	entry := RegisterViperEntry(
		WithNameViper(name),
		WithViperInstanceViper(vp))

	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, ViperEntryType, entry.GetType())
}

func TestViperEntry_String_HappyCase(t *testing.T) {
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
viper:
  - name: unit-test-viper
    path: %s
`
	// override path
	configFile = fmt.Sprintf(configFile, tempDir)

	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterViperEntriesWithConfig(configFilePath)

	assert.NotEmpty(t, entries)
	entry := GlobalAppCtx.GetViperEntry("unit-test-viper")
	assert.NotNil(t, entry)

	m := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal([]byte(entry.String()), &m))

	assert.Contains(t, m, "entry_name")
	assert.Contains(t, m, "entry_type")
	assert.Contains(t, m, "path")
}
