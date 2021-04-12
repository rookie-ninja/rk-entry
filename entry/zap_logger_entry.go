// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-logger"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	ZapLoggerEntryType = "zap-logger"
)

// Create zap logger entry with noop.
// Since we don't need any log rotation in case of noop, lumberjack config and logger config will be nil.
func NoopZapLoggerEntry() *ZapLoggerEntry {
	return &ZapLoggerEntry{
		entryName:        "rk-zap-logger-noop",
		entryType:        ZapLoggerEntryType,
		logger:           rklogger.NoopLogger,
		loggerConfig:     nil,
		lumberjackConfig: nil,
	}
}

// Bootstrap config of Zap Logger information.
// 1: AppLogger.Name: Name of app logger entry.
// 2: AppLogger.Zap: zap logger config, refer to zap.Config.
// 3: AppLogger.Lumberjack: lumberjack config, refer to lumberjack.Logger.
type BootConfigZapLogger struct {
	ZapLogger []struct {
		Name       string                  `yaml:"name"`
		Zap        *rklogger.ZapConfigWrap `yaml:"zap"`
		Lumberjack *lumberjack.Logger      `yaml:"lumberjack"`
	} `yaml:"zapLogger"`
}

// ZapLoggerEntry contains bellow fields.
// 1: entryName: Name of entry.
// 2: entryType: Type of entry which is ZapLoggerEntryType.
// 3: logger: zap.Logger which was initialized at the beginning.
// 4: loggerConfig: zap.Logger config which was initialized at the beginning which is not accessible after initialization..
// 5: lumberjackConfig: lumberjack.Logger which was initialized at the beginning.
type ZapLoggerEntry struct {
	entryName        string
	entryType        string
	logger           *zap.Logger
	loggerConfig     *zap.Config
	lumberjackConfig *lumberjack.Logger
}

// ZapLoggerEntry Option which used while registering entry from codes.
type ZapLoggerEntryOption func(*ZapLoggerEntry)

// Provide name of entry.
func WithNameZap(name string) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.entryName = name
	}
}

// Provide zap logger related entity of entry.
func WithLoggerZap(logger *zap.Logger, loggerConfig *zap.Config, lumberjackConfig *lumberjack.Logger) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.logger = logger
		entry.loggerConfig = loggerConfig
		entry.lumberjackConfig = lumberjackConfig
	}
}

// Create zap logger entries with config file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterZapLoggerEntriesWithConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: unmarshal user provided config into boot config struct
	config := &BootConfigZapLogger{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: init zap logger entries with boot config
	for i := range config.ZapLogger {
		element := config.ZapLogger[i]

		// assign default zap config and lumberjack config
		appLoggerConfig := rklogger.NewZapStdoutConfig()
		appLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()
		// override with user provided zap config and lumberjack config
		rkcommon.OverrideZapConfig(appLoggerConfig, rklogger.TransformToZapConfig(element.Zap))
		rkcommon.OverrideLumberjackConfig(appLoggerLumberjackConfig, element.Lumberjack)
		// create app logger with config
		appLogger, err := rklogger.NewZapLoggerWithConf(appLoggerConfig, appLoggerLumberjackConfig)
		if err != nil {
			rkcommon.ShutdownWithError(err)
		}

		entry := RegisterZapLoggerEntry(
			WithNameZap(element.Name),
			WithLoggerZap(appLogger, appLoggerConfig, appLoggerLumberjackConfig))

		res[element.Name] = entry
	}

	return res
}

// Crate event logger entry with options.
func RegisterZapLoggerEntry(opts ...ZapLoggerEntryOption) *ZapLoggerEntry {
	entry := &ZapLoggerEntry{
		entryType: ZapLoggerEntryType,
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.logger == nil {
		entry.loggerConfig = rklogger.NewZapStdoutConfig()
		if logger, err := entry.loggerConfig.Build(); err != nil {
			rkcommon.ShutdownWithError(err)
		} else {
			entry.logger = logger
		}
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "zap-logger-" + rkcommon.RandString(4)
	}

	GlobalAppCtx.addZapLoggerEntry(entry)

	return entry
}

// Bootstrap entry.
func (entry *ZapLoggerEntry) Bootstrap(context.Context) {
	// no op
}

// Interrupt entry.
func (entry *ZapLoggerEntry) Interrupt(context.Context) {
	// no op
}

// Get name of entry.
func (entry *ZapLoggerEntry) GetName() string {
	return entry.entryName
}

// Get type of entry.
func (entry *ZapLoggerEntry) GetType() string {
	return entry.entryType
}

// Convert entry into JSON style string.
func (entry *ZapLoggerEntry) String() string {
	m := map[string]interface{}{
		"entry_name": entry.entryName,
		"entry_type": entry.entryType,
	}

	if entry.loggerConfig != nil {
		m["output_path"] = entry.loggerConfig.OutputPaths
		m["level"] = entry.loggerConfig.Level
	}

	if entry.lumberjackConfig != nil {
		m["lumberjack_filename"] = entry.lumberjackConfig.Filename
		m["lumberjack_compress"] = entry.lumberjackConfig.Compress
		m["lumberjack_maxsize"] = entry.lumberjackConfig.MaxSize
		m["lumberjack_maxage"] = entry.lumberjackConfig.MaxAge
		m["lumberjack_maxbackups"] = entry.lumberjackConfig.MaxBackups
		m["lumberjack_localtime"] = entry.lumberjackConfig.LocalTime
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

// Get zap logger, refer to zap.Logger.
func (entry *ZapLoggerEntry) GetLogger() *zap.Logger {
	return entry.logger
}

// Get zap logger config, refer to zap.Config.
func (entry *ZapLoggerEntry) GetLoggerConfig() *zap.Config {
	return entry.loggerConfig
}

// Get lumberjack config, refer to lumberjack.Logger.
func (entry *ZapLoggerEntry) GetLumberjackConfig() *lumberjack.Logger {
	return entry.lumberjackConfig
}
