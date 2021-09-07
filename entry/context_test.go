// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"fmt"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestGlobalAppCtx_init(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx)

	// validate start time recorded.
	assert.NotNil(t, GlobalAppCtx.GetStartTime())

	// validate application logger entry.
	assert.NotNil(t, GlobalAppCtx.GetZapLoggerEntryDefault())

	// validate event logger entry.
	assert.NotNil(t, GlobalAppCtx.GetEventLoggerEntryDefault())

	// validate basic entry reg functions.
	assert.Len(t, internalEntryRegFuncList, 7)

	// validate user entries.
	assert.Empty(t, entryRegFuncList)

	// validate app info entry
	assert.NotNil(t, GlobalAppCtx.GetAppInfoEntry())
	assert.NotNil(t, GlobalAppCtx.GetAppInfoEntryRaw())

	// validate config entries.
	configEntries := GlobalAppCtx.ListConfigEntries()
	assert.Equal(t, 0, len(configEntries))

	// validate zap logger entries.
	zapEntries := GlobalAppCtx.ListZapLoggerEntries()
	assert.Equal(t, 1, len(zapEntries))

	// validate event logger entries.
	eventEntries := GlobalAppCtx.ListEventLoggerEntries()
	assert.Equal(t, 1, len(eventEntries))

	// validate cert entries.
	certEntries := GlobalAppCtx.ListCertEntries()
	fmt.Println(GlobalAppCtx.ListCertEntries())
	assert.Equal(t, 0, len(certEntries))

	// validate shutdown hooks.
	assert.Empty(t, GlobalAppCtx.ListShutdownHooks())

	// validate user values.
	values := GlobalAppCtx.ListValues()
	assert.Equal(t, 0, len(values))
}

func TestRegisterEntryRegFunc_WithNilInput(t *testing.T) {
	length := len(entryRegFuncList)
	RegisterEntryRegFunc(nil)
	assert.Len(t, entryRegFuncList, length)
}

func TestRegisterEntryRegFunc_HappyCase(t *testing.T) {
	regFunc := func(string) map[string]Entry {
		return make(map[string]Entry)
	}

	length := len(entryRegFuncList)

	RegisterEntryRegFunc(regFunc)
	assert.Len(t, entryRegFuncList, length+1)
	// clear reg functions
	entryRegFuncList = entryRegFuncList[:0]
}

func TestListEntryRegFunc_HappyCase(t *testing.T) {
	regFunc := func(string) map[string]Entry {
		return make(map[string]Entry)
	}

	RegisterEntryRegFunc(regFunc)
	assert.Len(t, ListEntryRegFunc(), 1)
	// clear reg functions
	entryRegFuncList = entryRegFuncList[:0]
}

func TestRegisterInternalEntriesFromConfig(t *testing.T) {
	assertNotPanic(t)
	filePath := createFileAtTestTempDir(t, `---`)
	RegisterInternalEntriesFromConfig(filePath)
}

func TestAppContext_SetRkMetaEntry(t *testing.T) {
	assert.Empty(t, GlobalAppCtx.SetRkMetaEntry(nil))
	assert.NotEmpty(t, GlobalAppCtx.SetRkMetaEntry(&EntryMock{Name: "mock"}))
	GlobalAppCtx.rkMetaEntry = nil
}

func TestAppContext_GetRkMetaEntry(t *testing.T) {
	assert.Nil(t, GlobalAppCtx.GetRkMetaEntry())
	metaEntry := &RkMetaEntry{}
	GlobalAppCtx.SetRkMetaEntry(metaEntry)
	assert.NotNil(t, GlobalAppCtx.GetRkMetaEntry())
	GlobalAppCtx.rkMetaEntry = nil
}

func TestAppContext_GetRkMetaEntryRaw(t *testing.T) {
	assert.Nil(t, GlobalAppCtx.GetRkMetaEntryRaw())
	metaEntry := &RkMetaEntry{}
	GlobalAppCtx.SetRkMetaEntry(metaEntry)
	assert.NotNil(t, GlobalAppCtx.GetRkMetaEntryRaw())
	GlobalAppCtx.rkMetaEntry = nil
}

// value related
func TestAppContext_AddValue_WithEmptyKey(t *testing.T) {
	key := ""
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_AddValue_WithEmptyValue(t *testing.T) {
	key := "key"
	value := ""
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_AddValue_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_GetValue_WithEmptyKey(t *testing.T) {
	key := ""
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_GetValue_WithEmptyValue(t *testing.T) {
	key := "key"
	value := ""
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_GetValue_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_ListValues_WithEmptyKey(t *testing.T) {
	key := ""
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)
	assert.Equal(t, value, GlobalAppCtx.ListValues()[key])
	GlobalAppCtx.ClearValues()
}

func TestAppContext_ListValues_WithEmptyValue(t *testing.T) {
	key := "key"
	value := ""
	GlobalAppCtx.AddValue(key, value)
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)
	assert.Equal(t, value, GlobalAppCtx.ListValues()[key])
	GlobalAppCtx.ClearValues()
}

func TestAppContext_ListValues_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)
	assert.Equal(t, value, GlobalAppCtx.ListValues()[key])
	GlobalAppCtx.ClearValues()
}

func TestAppContext_RemoveValue_WithNonExistValue(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	GlobalAppCtx.RemoveValue("non-exist-value")
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)

	GlobalAppCtx.ClearValues()
}

func TestAppContext_RemoveValue_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	GlobalAppCtx.RemoveValue(key)
	assert.Empty(t, GlobalAppCtx.ListValues())

	GlobalAppCtx.ClearValues()
}

func TestAppContext_ClearValues_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)

	GlobalAppCtx.ClearValues()
	assert.Empty(t, GlobalAppCtx.ListValues())
}

// zap logger related
func TestAppContext_AddZapLoggerEntry_WithNilEntry(t *testing.T) {
	assert.Empty(t, GlobalAppCtx.AddZapLoggerEntry(nil))
}

func TestAppContext_RemoveZapLoggerEntry_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"

	entry := RegisterZapLoggerEntry(WithNameZap(loggerName))

	assert.Equal(t, entry, GlobalAppCtx.GetZapLoggerEntry(loggerName))

	// remove zap logger entry
	assert.True(t, GlobalAppCtx.RemoveZapLoggerEntry(loggerName))
}

func TestAppContext_RemoveZapLoggerEntry_WithNonExistEntry(t *testing.T) {
	loggerName := "logger-unit-test"

	RegisterZapLoggerEntry(WithNameZap(loggerName))

	// remove zap logger entry
	assert.False(t, GlobalAppCtx.RemoveZapLoggerEntry("non-exist-entry"))

	// remove zap logger for clearance of global app context
	assert.True(t, GlobalAppCtx.RemoveZapLoggerEntry(loggerName))
}

func TestAppContext_GetZapLogger_WithNonExist(t *testing.T) {
	loggerName := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetZapLogger(loggerName))
}

func TestAppContext_GetZapLogger_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"

	RegisterZapLoggerEntry(
		WithNameZap(loggerName),
		WithLoggerZap(rklogger.StdoutLogger, nil, nil))

	entry := GlobalAppCtx.GetZapLoggerEntry(loggerName)

	assert.Equal(t, entry.GetLogger(), GlobalAppCtx.GetZapLogger(loggerName))

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_GetZapLoggerConfig_WithNonExist(t *testing.T) {
	loggerName := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetZapLoggerConfig(loggerName))
}

func TestAppContext_GetZapLoggerConfig_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"

	RegisterZapLoggerEntry(
		WithNameZap(loggerName),
		WithLoggerZap(nil, rklogger.StdoutLoggerConfig, nil))

	entry := GlobalAppCtx.GetZapLoggerEntry(loggerName)

	assert.Equal(t, entry.GetLoggerConfig(), GlobalAppCtx.GetZapLoggerConfig(loggerName))

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_GetZapLoggerEntry_WithNonExist(t *testing.T) {
	loggerName := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetZapLoggerEntry(loggerName))
}

func TestAppContext_GetZapLoggerEntry_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"

	RegisterZapLoggerEntry(WithNameZap(loggerName))

	entry := GlobalAppCtx.GetZapLoggerEntry(loggerName)

	assert.Equal(t, entry, GlobalAppCtx.GetZapLoggerEntry(loggerName))

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_ListZapLoggerEntries_HappyCase(t *testing.T) {
	length := len(GlobalAppCtx.ListZapLoggerEntries())

	loggerName := "logger-unit-test"

	RegisterZapLoggerEntry(WithNameZap(loggerName))

	assert.Len(t, GlobalAppCtx.ListZapLoggerEntries(), length+1)

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_ListZapLoggerEntriesRaw_HappyCase(t *testing.T) {
	length := len(GlobalAppCtx.ListZapLoggerEntriesRaw())

	loggerName := "logger-unit-test"

	RegisterZapLoggerEntry(WithNameZap(loggerName))

	assert.Len(t, GlobalAppCtx.ListZapLoggerEntriesRaw(), length+1)

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_GetZapLoggerDefault_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetZapLoggerDefault())
}

func TestAppContext_GetZapLoggerConfigDefault_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetZapLoggerConfigDefault())
}

func TestAppContext_GetZapLoggerEntryDefault_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetZapLoggerEntryDefault())
}

func TestAppContext_GetDefaultZapLogger_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetZapLoggerDefault())
}

// event logger related
func TestAppContext_AddEventLoggerEntry_WithNilEntry(t *testing.T) {
	assert.Empty(t, GlobalAppCtx.AddEventLoggerEntry(nil))
}

func TestAppContext_GetEventLoggerEntry_WithNonExist(t *testing.T) {
	loggerName := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetEventLoggerEntry(loggerName))
}

func TestAppContext_GetEventLoggerEntry_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"

	RegisterEventLoggerEntry(WithNameEvent(loggerName))

	entry := GlobalAppCtx.GetZapLoggerEntry(loggerName)

	assert.Equal(t, entry, GlobalAppCtx.GetZapLoggerEntry(loggerName))

	// remove event logger entry
	GlobalAppCtx.RemoveEventLoggerEntry(loggerName)
}

func TestAppContext_ListEventLoggerEntries_HappyCase(t *testing.T) {
	length := len(GlobalAppCtx.ListEventLoggerEntries())

	loggerName := "event-logger-unit-test"

	RegisterEventLoggerEntry(WithNameEvent(loggerName))

	assert.Len(t, GlobalAppCtx.ListEventLoggerEntries(), length+1)

	// remove event logger entry
	GlobalAppCtx.RemoveEventLoggerEntry(loggerName)
}

func TestAppContext_RemoveEventLoggerEntry_WithNonExist(t *testing.T) {
	assert.False(t, GlobalAppCtx.RemoveEventLoggerEntry("non-exist"))
}

func TestAppContext_GetEventHelper_WithNonExist(t *testing.T) {
	assert.Nil(t, GlobalAppCtx.GetEventHelper("non-exist"))
}

func TestAppContext_ListEventLoggerEntriesRaw_HappyCase(t *testing.T) {
	length := len(GlobalAppCtx.ListEventLoggerEntriesRaw())

	loggerName := "logger-unit-test"

	RegisterEventLoggerEntry(WithNameEvent(loggerName))

	assert.Equal(t, length+1, len(GlobalAppCtx.ListEventLoggerEntriesRaw()))

	// remove event logger entry
	GlobalAppCtx.RemoveEventLoggerEntry(loggerName)
}

func TestAppContext_GetEventFactory_WithNonExist(t *testing.T) {
	loggerName := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetEventFactory(loggerName))
}

func TestAppContext_GetEventFactory_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"
	fac := rkquery.NewEventFactory()

	RegisterEventLoggerEntry(
		WithNameEvent(loggerName),
		WithEventFactoryEvent(fac))

	assert.Equal(t, fac, GlobalAppCtx.GetEventFactory(loggerName))

	// remove event logger entry
	GlobalAppCtx.RemoveEventLoggerEntry(loggerName)
}

func TestAppContext_GetEventHelper_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"
	fac := rkquery.NewEventFactory()

	RegisterEventLoggerEntry(
		WithNameEvent(loggerName),
		WithEventFactoryEvent(fac))

	assert.NotNil(t, GlobalAppCtx.GetEventHelper(loggerName))

	// remove event logger entry
	GlobalAppCtx.RemoveEventLoggerEntry(loggerName)
}

func TestAppContext_GetEventLoggerEntryDefault_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetEventLoggerEntryDefault())
}

func TestAppContext_GetAppInfoEntry_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetAppInfoEntry())
}

// Cert entry related
func TestAppContext_AddCertEntry_WithNilEntry(t *testing.T) {
	assert.Empty(t, GlobalAppCtx.AddCertEntry(nil))
	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

func TestAppContext_GetCertEntry_WithNonExist(t *testing.T) {
	name := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetCertEntry(name))
	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

func TestAppContext_GetCertEntry_HappyCase(t *testing.T) {
	RegisterCertEntry()

	assert.True(t, len(GlobalAppCtx.ListCertEntries()) == 1)

	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

func TestAppContext_ListCertEntries_WithEmptyList(t *testing.T) {
	assert.True(t, len(GlobalAppCtx.ListCertEntries()) == 0)
	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

func TestAppContext_ListCertEntries_HappyCase(t *testing.T) {
	RegisterCertEntry()

	assert.True(t, len(GlobalAppCtx.ListCertEntries()) == 1)
	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

func TestAppContext_ListCertEntriesRaw_WithEmptyList(t *testing.T) {
	assert.True(t, len(GlobalAppCtx.ListCertEntriesRaw()) == 0)
	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

func TestAppContext_ListCertEntriesRaw_HappyCase(t *testing.T) {
	RegisterCertEntry()

	assert.True(t, len(GlobalAppCtx.ListCertEntriesRaw()) == 1)
	assert.NotNil(t, GlobalAppCtx.GetCertEntry(CertEntryName))
	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

func TestAppContext_RemoveCertEntry(t *testing.T) {
	RegisterCertEntry()

	assert.True(t, GlobalAppCtx.RemoveCertEntry(CertEntryName))
	assert.False(t, GlobalAppCtx.RemoveCertEntry("non-exist"))
	// clear cert entries
	GlobalAppCtx.clearCertEntries()
}

// Cred entry related
func TestAppContext_AddCredEntry_WithNilEntry(t *testing.T) {
	assert.Empty(t, GlobalAppCtx.AddCredEntry(nil))
}

func TestAppContext_GetCredEntry_WithNonExist(t *testing.T) {
	name := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetCredEntry(name))
	// clear cred entries
	GlobalAppCtx.clearCredEntries()
}

func TestAppContext_GetCredEntry_HappyCase(t *testing.T) {
	RegisterCredEntry()

	assert.True(t, len(GlobalAppCtx.ListCredEntries()) == 1)
	assert.NotNil(t, GlobalAppCtx.GetCredEntry(CredEntryName))
	assert.Nil(t, GlobalAppCtx.GetCredEntry("non-exist"))

	// clear cred entries
	GlobalAppCtx.clearCredEntries()
}

func TestAppContext_ListCredEntries_WithEmptyList(t *testing.T) {
	assert.True(t, len(GlobalAppCtx.ListCredEntries()) == 0)
	// clear cred entries
	GlobalAppCtx.clearCredEntries()
}

func TestAppContext_ListCredEntries_HappyCase(t *testing.T) {
	RegisterCredEntry()

	assert.True(t, len(GlobalAppCtx.ListCredEntries()) == 1)
	// clear cred entries
	GlobalAppCtx.clearCredEntries()
}

func TestAppContext_ListCredEntriesRaw_WithEmptyList(t *testing.T) {
	assert.True(t, len(GlobalAppCtx.ListCredEntriesRaw()) == 0)
	// clear cred entries
	GlobalAppCtx.clearCredEntries()
}

func TestAppContext_RemoveCredEntry(t *testing.T) {
	RegisterCredEntry()

	assert.True(t, GlobalAppCtx.RemoveCredEntry(CredEntryName))
	assert.False(t, GlobalAppCtx.RemoveCredEntry("non-exist"))

	// clear cred entries
	GlobalAppCtx.clearCredEntries()
}

func TestAppContext_ListCredEntriesRaw_HappyCase(t *testing.T) {
	RegisterCertEntry()

	assert.True(t, len(GlobalAppCtx.ListCertEntriesRaw()) == 1)
	// clear viper entries
	GlobalAppCtx.clearCertEntries()
}

// Config entry related
func TestAppContext_AddConfigEntry_WithNilEntry(t *testing.T) {
	assert.Empty(t, GlobalAppCtx.AddConfigEntry(nil))
}

func TestAppContext_GetConfigEntry_WithNonExist(t *testing.T) {
	name := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetConfigEntry(name))
	// clear viper entries
	GlobalAppCtx.clearConfigEntries()
}

func TestAppContext_GetConfigEntry_HappyCase(t *testing.T) {
	name := "viper-config"
	vp := viper.New()
	RegisterConfigEntry(
		WithNameConfig(name),
		WithViperInstanceConfig(vp))

	assert.True(t, len(GlobalAppCtx.ListConfigEntries()) == 1)
	assert.NotNil(t, GlobalAppCtx.GetConfigEntry(name))
	// clear viper entries
	GlobalAppCtx.clearConfigEntries()
}

func TestAppContext_ListConfigEntries_WithEmptyList(t *testing.T) {
	assert.True(t, len(GlobalAppCtx.ListConfigEntries()) == 0)
	// clear viper entries
	GlobalAppCtx.clearConfigEntries()
}

func TestAppContext_ListConfigEntries_HappyCase(t *testing.T) {
	name := "viper-config"
	vp := viper.New()
	RegisterConfigEntry(
		WithNameConfig(name),
		WithViperInstanceConfig(vp))

	assert.True(t, len(GlobalAppCtx.ListConfigEntries()) == 1)
	// clear viper entries
	GlobalAppCtx.clearConfigEntries()
}

func TestAppContext_ListConfigEntriesRaw_WithEmptyList(t *testing.T) {
	assert.True(t, len(GlobalAppCtx.ListConfigEntriesRaw()) == 0)
	// clear viper entries
	GlobalAppCtx.clearConfigEntries()
}

func TestAppContext_ListConfigEntriesRaw_HappyCase(t *testing.T) {
	name := "viper-config"
	vp := viper.New()
	RegisterConfigEntry(
		WithNameConfig(name),
		WithViperInstanceConfig(vp))

	assert.True(t, len(GlobalAppCtx.ListConfigEntriesRaw()) == 1)
	// clear viper entries
	GlobalAppCtx.clearConfigEntries()
}

func TestAppContext_RemoveConfigEntry(t *testing.T) {
	assert.False(t, GlobalAppCtx.RemoveConfigEntry("non-exist"))

	name := "viper-config"
	vp := viper.New()
	RegisterConfigEntry(
		WithNameConfig(name),
		WithViperInstanceConfig(vp))

	assert.True(t, len(GlobalAppCtx.ListConfigEntriesRaw()) == 1)

	assert.True(t, GlobalAppCtx.RemoveConfigEntry(name))
	// clear viper entries
	GlobalAppCtx.clearConfigEntries()
}

// shutdown signal related
func TestAppContext_GetShutdownSig_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetShutdownSig())
}

// shutdown hook related
func TestAppContext_AddShutdownHook_WithEmptyName(t *testing.T) {
	name := ""
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.Equal(t, 1, len(GlobalAppCtx.ListShutdownHooks()))
	assert.NotNil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_AddShutdownHook_WithNilFunc(t *testing.T) {
	name := ""
	GlobalAppCtx.AddShutdownHook(name, nil)
	assert.Equal(t, 0, len(GlobalAppCtx.ListShutdownHooks()))
	assert.Nil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_AddShutdownHook_HappyCase(t *testing.T) {
	name := "unit-test-hook"
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.Equal(t, 1, len(GlobalAppCtx.ListShutdownHooks()))
	assert.NotNil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_GetShutdownHook_WithNonExistHooks(t *testing.T) {
	name := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_GetShutdownHook_HappyCase(t *testing.T) {
	name := "unit-test-hook"
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.NotNil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_ListShutdownHooks_WithEmptyHooks(t *testing.T) {
	assert.Equal(t, 0, len(GlobalAppCtx.ListShutdownHooks()))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_ListShutdownHooks_HappyCase(t *testing.T) {
	name := "unit-test-hook"
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.Equal(t, 1, len(GlobalAppCtx.ListShutdownHooks()))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

// entry related
func TestAppContext_AddEntry_WithEmptyName(t *testing.T) {
	name := "unit-test-entry"
	entry := &EntryMock{
		Name: name,
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, 1, len(GlobalAppCtx.ListEntries()))
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(name))

	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_AddEntry_WithNilEntry(t *testing.T) {
	GlobalAppCtx.AddEntry(nil)
	assert.Equal(t, 0, len(GlobalAppCtx.ListEntries()))
	assert.Nil(t, GlobalAppCtx.GetEntry(""))

	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_AddEntry_HappyCase(t *testing.T) {
	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, 1, len(GlobalAppCtx.ListEntries()))
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(entry.GetName()))

	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_GetEntry_WithNonExistEntry(t *testing.T) {
	name := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetEntry(name))

	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_GetEntry_HappyCase(t *testing.T) {
	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(entry.GetName()))

	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_ListEntries_WithEmptyEntry(t *testing.T) {
	assert.Equal(t, 0, len(GlobalAppCtx.ListEntries()))

	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_ListEntries_HappyCase(t *testing.T) {
	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, 1, len(GlobalAppCtx.ListEntries()))
	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_MergeEntries_HappyCase(t *testing.T) {
	target := map[string]Entry{
		"target-entry": &EntryMock{},
	}

	entry := &EntryMock{
		Name: "unit-test-entry",
	}

	GlobalAppCtx.AddEntry(entry)
	GlobalAppCtx.MergeEntries(target)
	assert.Equal(t, 2, len(GlobalAppCtx.ListEntries()))
	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_RemoveEntry(t *testing.T) {
	assert.False(t, GlobalAppCtx.RemoveEntry("non-exist"))

	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(entry.GetName()))
	assert.True(t, GlobalAppCtx.RemoveEntry(entry.GetName()))
	// clear entries
	GlobalAppCtx.clearEntries()
}

func TestAppContext_RemoveShutdownHook(t *testing.T) {
	assert.False(t, GlobalAppCtx.RemoveShutdownHook("non-exist"))
	GlobalAppCtx.AddShutdownHook("ut-shutdownhook", func() {})
	assert.True(t, GlobalAppCtx.RemoveShutdownHook("ut-shutdownhook"))
}

func TestAppContext_WaitForShutdownSig(t *testing.T) {
	go func() {
		time.Sleep(1 * time.Second)
		GlobalAppCtx.shutdownSig <- syscall.SIGTERM
	}()

	GlobalAppCtx.WaitForShutdownSig()
}

type EntryMock struct {
	Name string
}

func (entry *EntryMock) Bootstrap(context.Context) {}

func (entry *EntryMock) Interrupt(context.Context) {}

func (entry *EntryMock) GetName() string {
	return entry.Name
}

func (entry *EntryMock) GetType() string {
	return ""
}

func (entry *EntryMock) String() string {
	return ""
}

func (entry *EntryMock) GetDescription() string {
	return ""
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
