// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestNoopZapLoggerEntry_HappyCase(t *testing.T) {
	entry := NoopZapLoggerEntry()
	assert.NotNil(t, entry)
	assert.Equal(t, "zap-logger-noop", entry.entryName)
	assert.Equal(t, ZapLoggerEntryType, entry.entryType)
	assert.NotNil(t, entry.logger)
	assert.Nil(t, entry.loggerConfig)
	assert.Nil(t, entry.lumberjackConfig)
}

func TestWithNameZap_WithEmptyString(t *testing.T) {
	entry := RegisterZapLoggerEntry(WithNameZap(""))
	assert.NotNil(t, entry)
	// default name should be assigned with random number
	assert.NotEmpty(t, entry.entryName)
}

func TestWithNameZap_HappyCase(t *testing.T) {
	entry := RegisterZapLoggerEntry(WithNameZap("ut-zap-logger"))
	assert.NotNil(t, entry)
	// default name would be assigned with random number
	assert.Equal(t, "ut-zap-logger", entry.entryName)
}

func TestWithLoggerZap_WithNilInput(t *testing.T) {
	entry := RegisterZapLoggerEntry(WithLoggerZap(nil, nil, nil))
	assert.NotNil(t, entry)
	// default logger and logger config would be assigned
	assert.NotNil(t, entry.logger)
	assert.NotNil(t, entry.loggerConfig)
	assert.Nil(t, entry.lumberjackConfig)
}

func TestWithLoggerZap_HappyCase(t *testing.T) {
	logger := rklogger.StdoutLogger
	loggerConfig := rklogger.StdoutLoggerConfig
	lumberjackConfig := rklogger.LumberjackConfig

	entry := RegisterZapLoggerEntry(WithLoggerZap(logger, loggerConfig, lumberjackConfig))
	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, logger, entry.logger)
	assert.Equal(t, loggerConfig, entry.loggerConfig)
	assert.Equal(t, lumberjackConfig, entry.lumberjackConfig)
}

func TestRegisterZapLoggerEntriesWithConfig_WithoutElement(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
zapLogger:
`
	// create bootstrap config file at ut temp dir
	configFilePath := createFileAtTestTempDir(t, configFile)
	// register entries with config file
	entries := RegisterZapLoggerEntriesWithConfig(configFilePath)

	assert.Empty(t, entries)
}

func TestRegisterZapLoggerEntriesWithConfig_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
zapLogger:
 - name: ut-zap-logger
   zap:
     level: info
     development: false
     disableCaller: false
     disableStacktrace: false
     encoding: console
     outputPaths: ["ut.log"]
     errorOutputPaths: ["ut.log"]
     initialFields: 
       ut-key: "ut-key"
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
	entries := RegisterZapLoggerEntriesWithConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToZapLoggerEntry(t, entries["ut-zap-logger"])

	// validate default fields
	assert.Equal(t, "ut-zap-logger", entry.entryName)
	assert.Equal(t, ZapLoggerEntryType, entry.entryType)

	// validate zap logger
	assert.NotNil(t, entry.logger)

	// validate zap logger config
	assert.NotNil(t, entry.loggerConfig)
	assert.Equal(t, zap.InfoLevel, entry.loggerConfig.Level.Level())
	assert.False(t, entry.loggerConfig.Development)
	assert.False(t, entry.loggerConfig.DisableCaller)
	assert.False(t, entry.loggerConfig.DisableStacktrace)
	assert.Equal(t, "console", entry.loggerConfig.Encoding)
	assert.Contains(t, entry.loggerConfig.OutputPaths, "ut.log")
	assert.Contains(t, entry.loggerConfig.ErrorOutputPaths, "ut.log")
	assert.Contains(t, entry.loggerConfig.InitialFields, "ut-key")

	// validate lumberjack config
	assert.Equal(t, "ut-lumberjack-filename", entry.lumberjackConfig.Filename)
	assert.Equal(t, 1, entry.lumberjackConfig.MaxSize)
	assert.Equal(t, 1, entry.lumberjackConfig.MaxAge)
	assert.Equal(t, 1, entry.lumberjackConfig.MaxBackups)
	assert.True(t, entry.lumberjackConfig.LocalTime)
	assert.True(t, entry.lumberjackConfig.Compress)
}

func TestRegisterZapLoggerEntry_WithoutOptions(t *testing.T) {
	entry := RegisterZapLoggerEntry()

	assert.NotNil(t, entry)

	// validate default fields
	assert.Contains(t, entry.entryName, "zap-logger-")
	assert.Equal(t, ZapLoggerEntryType, entry.entryType)

	// validate zap logger
	assert.NotNil(t, entry.logger)
	assert.NotNil(t, entry.loggerConfig)
	assert.Nil(t, entry.lumberjackConfig)
}

func TestRegisterZapLoggerEntry_HappyCase(t *testing.T) {
	logger := rklogger.StdoutLogger
	loggerConfig := rklogger.StdoutLoggerConfig
	lumberjackConfig := rklogger.LumberjackConfig

	entry := RegisterZapLoggerEntry(WithLoggerZap(logger, loggerConfig, lumberjackConfig))
	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, logger, entry.logger)
	assert.Equal(t, loggerConfig, entry.loggerConfig)
	assert.Equal(t, lumberjackConfig, entry.lumberjackConfig)
}

func TestZapLoggerEntry_Bootstrap_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterZapLoggerEntry().Bootstrap(context.Background())
}

func TestZapLoggerEntry_Interrupt_HappyCase(t *testing.T) {
	assertNotPanic(t)
	RegisterZapLoggerEntry().Interrupt(context.Background())
}

func TestZapLoggerEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterZapLoggerEntry(
		WithNameZap("ut-zap-logger"))
	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, "ut-zap-logger", entry.GetName())
}

func TestZapLoggerEntry_GetType_HappyCase(t *testing.T) {
	entry := RegisterZapLoggerEntry()
	assert.NotNil(t, entry)

	// validate default fields
	assert.Equal(t, ZapLoggerEntryType, entry.GetType())
}

func TestZapLoggerEntry_String_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	configFile := `
---
zapLogger:
 - name: ut-zap-logger
   zap:
     level: info
     development: false
     disableCaller: false
     disableStacktrace: false
     encoding: console
     outputPaths: ["ut.log"]
     errorOutputPaths: ["ut.log"]
     initialFields: 
       ut-key: "ut-key"
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
	entries := RegisterZapLoggerEntriesWithConfig(configFilePath)

	assert.Len(t, entries, 1)

	entry := convertToZapLoggerEntry(t, entries["ut-zap-logger"])

	m := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal([]byte(entry.String()), &m))

	assert.Contains(t, m, "entry_name")
	assert.Contains(t, m, "entry_type")
	assert.Contains(t, m, "output_path")
	assert.Contains(t, m, "level")
	assert.Contains(t, m, "lumberjack_filename")
	assert.Contains(t, m, "lumberjack_compress")
	assert.Contains(t, m, "lumberjack_maxsize")
	assert.Contains(t, m, "lumberjack_maxage")
	assert.Contains(t, m, "lumberjack_maxbackups")
	assert.Contains(t, m, "lumberjack_localtime")
}

func TestZapLoggerEntry_GetLogger_HappyCase(t *testing.T) {
	logger := rklogger.StdoutLogger
	loggerConfig := rklogger.StdoutLoggerConfig
	lumberjackConfig := rklogger.LumberjackConfig

	entry := RegisterZapLoggerEntry(WithLoggerZap(logger, loggerConfig, lumberjackConfig))
	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, logger, entry.GetLogger())
}

func TestZapLoggerEntry_GetLoggerConfig_HappyCase(t *testing.T) {
	logger := rklogger.StdoutLogger
	loggerConfig := rklogger.StdoutLoggerConfig
	lumberjackConfig := rklogger.LumberjackConfig

	entry := RegisterZapLoggerEntry(WithLoggerZap(logger, loggerConfig, lumberjackConfig))
	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, loggerConfig, entry.GetLoggerConfig())
}

func TestZapLoggerEntry_GetLumberjackConfig_HappyCase(t *testing.T) {
	logger := rklogger.StdoutLogger
	loggerConfig := rklogger.StdoutLoggerConfig
	lumberjackConfig := rklogger.LumberjackConfig

	entry := RegisterZapLoggerEntry(WithLoggerZap(logger, loggerConfig, lumberjackConfig))
	assert.NotNil(t, entry)

	// default logger and logger config would be assigned
	assert.Equal(t, lumberjackConfig, entry.GetLumberjackConfig())
}

func convertToZapLoggerEntry(t *testing.T, raw Entry) *ZapLoggerEntry {
	entry, ok := raw.(*ZapLoggerEntry)
	assert.True(t, ok)
	return entry
}
