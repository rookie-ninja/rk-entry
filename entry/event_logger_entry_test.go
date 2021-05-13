// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNoopEventLoggerEntry_HappyCase(t *testing.T) {
	entry := NoopEventLoggerEntry()
	assert.NotNil(t, entry)
	assert.Equal(t, EventLoggerNameNoop, entry.EntryName)
	assert.Equal(t, EventLoggerEntryType, entry.EntryType)
	assert.NotNil(t, entry.EventFactory)
	assert.NotNil(t, entry.EventHelper)
	assert.Nil(t, entry.LumberjackConfig)
}

func TestWithNameEvent_WithEmptyString(t *testing.T) {
	entry := RegisterEventLoggerEntry(WithNameEvent(""))
	assert.NotNil(t, entry)
	// default name should be assigned with random number
	assert.NotEmpty(t, entry.EntryName)
}

func TestWithNameEvent_HappyCase(t *testing.T) {
	entry := RegisterEventLoggerEntry(WithNameEvent("ut-event-logger"))
	assert.NotNil(t, entry)
	// default name would be assigned with random number
	assert.Equal(t, "ut-event-logger", entry.EntryName)
}

func TestWithEventFactoryEvent_WithNilInput(t *testing.T) {
	entry := RegisterEventLoggerEntry(WithEventFactoryEvent(nil))
	assert.NotNil(t, entry)
	// default event factory would be assigned
	assert.NotNil(t, entry.EventFactory)
	assert.NotNil(t, entry.EventHelper)
}

func TestWithEventFactoryEvent_HappyCase(t *testing.T) {
	fac := rkquery.NewEventFactory()
	entry := RegisterEventLoggerEntry(WithEventFactoryEvent(fac))
	assert.NotNil(t, entry)

	assert.Equal(t, fac, entry.EventFactory)
	assert.NotNil(t, entry.EventHelper)
}

func TestRegisterEventLoggerEntriesWithConfig_WithoutRKAppName(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
eventLogger:
  - name: ut-event-logger
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterEventLoggerEntriesWithConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToEventLoggerEntry(t, entries["ut-event-logger"])
	// validate event factory
	assert.NotNil(t, entry.EventFactory)
	assert.Equal(t, AppNameDefault, entry.EventFactory.CreateEvent().GetAppName())
}

func TestRegisterEventLoggerEntriesWithConfig_WithoutElement(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
eventLogger:
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterEventLoggerEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterEventLoggerEntriesWithConfig_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
rk:
  appName: ut-app
eventLogger:
  - name: ut-event-logger
    format: RK
    outputPaths: ["ut.log"]
    lumberjack:
      filename: "ut-lumberjack-filename"
      maxsize: 1
      maxage: 1
      maxbackups: 1
      localtime: true
      compress: true
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterEventLoggerEntriesWithConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToEventLoggerEntry(t, entries["ut-event-logger"])

	// validate default fields
	assert.Equal(t, "ut-event-logger", entry.EntryName)
	assert.Equal(t, EventLoggerEntryType, entry.EntryType)

	// validate event factory
	assert.NotNil(t, entry.EventFactory)
	assert.Equal(t, "ut-app", entry.EventFactory.CreateEvent().GetAppName())

	// validate zap logger config in event factory
	assert.NotNil(t, entry.LoggerConfig)
	assert.Contains(t, entry.LoggerConfig.OutputPaths, "ut.log")

	// validate lumberjack config
	assert.NotNil(t, entry.LumberjackConfig)
	assert.Equal(t, "ut-lumberjack-filename", entry.LumberjackConfig.Filename)
	assert.Equal(t, 1, entry.LumberjackConfig.MaxSize)
	assert.Equal(t, 1, entry.LumberjackConfig.MaxAge)
	assert.Equal(t, 1, entry.LumberjackConfig.MaxBackups)
	assert.True(t, entry.LumberjackConfig.LocalTime)
	assert.True(t, entry.LumberjackConfig.Compress)
}

func TestRegisterEventLoggerEntry_WithoutOptions(t *testing.T) {
	entry := RegisterEventLoggerEntry()

	assert.NotNil(t, entry)

	// validate default fields
	assert.Contains(t, entry.EntryName, "eventLogger")
	assert.Equal(t, EventLoggerEntryType, entry.EntryType)

	// validate event factory
	assert.NotNil(t, entry.EventFactory)
	assert.Equal(t, "unknown", entry.EventFactory.CreateEvent().GetAppName())
}

func TestRegisterEventLoggerEntry_HappyCase(t *testing.T) {
	fac := rkquery.NewEventFactory()
	entry := RegisterEventLoggerEntry(
		WithNameEvent("ut-event-logger"),
		WithEventFactoryEvent(fac))

	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, "ut-event-logger", entry.EntryName)
	assert.Equal(t, EventLoggerEntryType, entry.EntryType)

	// validate event factory
	assert.Equal(t, fac, entry.EventFactory)
	assert.Equal(t, "unknown", entry.EventFactory.CreateEvent().GetAppName())
}

func TestEventLoggerEntry_Bootstrap_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterEventLoggerEntry().Bootstrap(context.Background())
}

func TestEventLoggerEntry_Interrupt_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterEventLoggerEntry().Interrupt(context.Background())
}

func TestEventLoggerEntry_GetName_HappyCase(t *testing.T) {
	fac := rkquery.NewEventFactory()
	entry := RegisterEventLoggerEntry(
		WithNameEvent("ut-event-logger"),
		WithEventFactoryEvent(fac))

	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, "ut-event-logger", entry.GetName())
}

func TestEventLoggerEntry_GetType_HappyCase(t *testing.T) {
	fac := rkquery.NewEventFactory()
	entry := RegisterEventLoggerEntry(
		WithNameEvent("ut-event-logger"),
		WithEventFactoryEvent(fac))

	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, EventLoggerEntryType, entry.GetType())
}

func TestEventLoggerEntry_String_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
rk:
  appName: ut-app
eventLogger:
  - name: ut-event-logger
    format: RK
    outputPaths: ["ut.log"]
    lumberjack:
      filename: "ut-lumberjack-filename"
      maxsize: 1
      maxage: 1
      maxbackups: 1
      localtime: true
      compress: true
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterEventLoggerEntriesWithConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToEventLoggerEntry(t, entries["ut-event-logger"])

	m := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal([]byte(entry.String()), &m))

	assert.Contains(t, m, "entryName")
	assert.Contains(t, m, "entryType")
	assert.Contains(t, m, "zapConfig")
	assert.Contains(t, m, "lumberjackConfig")
}

func TestEventLoggerEntry_GetEventFactory_HappyCase(t *testing.T) {
	fac := rkquery.NewEventFactory()
	entry := RegisterEventLoggerEntry(WithEventFactoryEvent(fac))

	assert.NotNil(t, entry)
	assert.Equal(t, fac, entry.GetEventFactory())
}

func TestEventLoggerEntry_GetEventHelper_HappyCase(t *testing.T) {
	fac := rkquery.NewEventFactory()
	entry := RegisterEventLoggerEntry(WithEventFactoryEvent(fac))

	assert.NotNil(t, entry)
	assert.NotNil(t, fac, entry.GetEventHelper())
}

func TestEventLoggerEntry_GetLumberjackConfig_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
rk:
  appName: ut-app
eventLogger:
  - name: ut-event-logger
    format: RK
    outputPaths: ["ut.log"]
    lumberjack:
      filename: "ut-lumberjack-filename"
      maxsize: 1
      maxage: 1
      maxbackups: 1
      localtime: true
      compress: true
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterEventLoggerEntriesWithConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToEventLoggerEntry(t, entries["ut-event-logger"])
	assert.NotNil(t, entry.GetEventHelper())
}

func convertToEventLoggerEntry(t *testing.T, raw Entry) *EventLoggerEntry {
	entry, ok := raw.(*EventLoggerEntry)
	assert.True(t, ok)
	return entry
}
