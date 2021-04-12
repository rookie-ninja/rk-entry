// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGlobalAppCtx_init(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx)

	// validate start time recorded.
	assert.NotNil(t, GlobalAppCtx.StartTime)

	// validate application logger entry.
	assert.NotNil(t, GlobalAppCtx.GetZapLoggerEntryDefault())

	// validate event logger entry.
	assert.NotNil(t, GlobalAppCtx.GetEventLoggerEntryDefault())

	// validate basic entry reg functions.
	assert.Len(t, basicEntryRegFuncList, 5)

	// validate user entries.
	assert.Empty(t, entryRegFuncList)

	// validate viper configs.
	assert.Empty(t, GlobalAppCtx.ViperConfigs)

	// validate shutdown hooks.
	assert.Empty(t, GlobalAppCtx.ShutdownHooks)

	// validate user values.
	assert.Empty(t, GlobalAppCtx.UserValues)
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

// logger related
func TestAppContext_addZapLoggerEntry_WithNilEntry(t *testing.T) {
	assert.Empty(t, GlobalAppCtx.addZapLoggerEntry(nil))
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

func TestAppContext_GetEventLoggerEntry_WithNonExist(t *testing.T) {
	loggerName := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetEventLoggerEntry(loggerName))
}

func TestAppContext_GetEventLoggerEntry_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"

	RegisterEventLoggerEntry(WithNameEvent(loggerName))

	entry := GlobalAppCtx.GetZapLoggerEntry(loggerName)

	assert.Equal(t, entry, GlobalAppCtx.GetZapLoggerEntry(loggerName))

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_ListEventLoggerEntries_HappyCase(t *testing.T) {
	length := len(GlobalAppCtx.ListEventLoggerEntries())

	loggerName := "event-logger-unit-test"

	RegisterEventLoggerEntry(WithNameEvent(loggerName))

	assert.Len(t, GlobalAppCtx.ListEventLoggerEntries(), length+1)

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
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

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_GetEventHelper_HappyCase(t *testing.T) {
	loggerName := "logger-unit-test"
	fac := rkquery.NewEventFactory()

	RegisterEventLoggerEntry(
		WithNameEvent(loggerName),
		WithEventFactoryEvent(fac))

	assert.NotNil(t, GlobalAppCtx.GetEventHelper(loggerName))

	// remove zap logger entry
	GlobalAppCtx.RemoveZapLoggerEntry(loggerName)
}

func TestAppContext_GetEventLoggerEntryDefault_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetEventLoggerEntryDefault())
}

func TestAppContext_GetAppInfoEntry_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetAppInfoEntry())
}

// viper entry related
func TestAppContext_GetViperConfig_WithNonExist(t *testing.T) {
	name := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetViperEntry(name))
	// clear viper entries
	GlobalAppCtx.clearViperEntries()
}

func TestAppContext_GetViperEntry_HappyCase(t *testing.T) {
	name := "viper-config"
	vp := viper.New()
	RegisterViperEntry(
		WithNameViper(name),
		WithViperInstanceViper(vp))

	assert.True(t, len(GlobalAppCtx.ViperConfigs) == 1)
	assert.NotNil(t, GlobalAppCtx.GetViperEntry(name))
	// clear viper entries
	GlobalAppCtx.clearViperEntries()
}

func TestAppContext_ListViperEntries_WithEmptyList(t *testing.T) {
	assert.True(t, len(GlobalAppCtx.ListViperEntries()) == 0)
	// clear viper entries
	GlobalAppCtx.clearViperEntries()
}

func TestAppContext_ListViperEntries_HappyCase(t *testing.T) {
	name := "viper-config"
	vp := viper.New()
	RegisterViperEntry(
		WithNameViper(name),
		WithViperInstanceViper(vp))

	assert.True(t, len(GlobalAppCtx.ListViperEntries()) == 1)
	// clear viper entries
	GlobalAppCtx.clearViperEntries()
}

// shutdown signal related
func TestAppContext_GetShutdownSig_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.ShutdownSig)
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
	name := ""
	entry := &EntryMock{
		Name: "unit-test-entry",
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

type EntryMock struct {
	Name string
}

func (entry *EntryMock) Bootstrap(context.Context) {}

func (entry *EntryMock) Interrupt(context.Context) {}

func (entry *EntryMock) GetName() string {
	return ""
}

func (entry *EntryMock) GetType() string {
	return ""
}

func (entry *EntryMock) String() string {
	return ""
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
