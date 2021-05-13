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
	ZapLoggerEntryType   = "ZapLoggerEntry"
	ZapLoggerNameNoop    = "ZapLoggerNoop"
	ZapLoggerDescription = "Internal RK entry which is used for logging with zap.Logger."
)

// Create zap logger entry with noop.
// Since we don't need any log rotation in case of noop, lumberjack config and logger config will be nil.
func NoopZapLoggerEntry() *ZapLoggerEntry {
	return &ZapLoggerEntry{
		EntryName:        ZapLoggerNameNoop,
		EntryType:        ZapLoggerEntryType,
		EntryDescription: ZapLoggerDescription,
		Logger:           rklogger.NoopLogger,
		LoggerConfig:     nil,
		LumberjackConfig: nil,
	}
}

// Bootstrap config of Zap Logger information.
// 1: ZapLogger.Name: Name of zap logger entry.
// 2: ZapLogger.Description: Description of zap logger entry.
// 3: ZapLogger.Zap: zap logger config, refer to zap.Config.
// 4: ZapLogger.Lumberjack: lumberjack config, refer to lumberjack.Logger.
type BootConfigZapLogger struct {
	ZapLogger []struct {
		Name        string                  `yaml:"name" json:"name"`
		Description string                  `yaml:"description" json:"description"`
		Zap         *rklogger.ZapConfigWrap `yaml:"zap" json:"zap"`
		Lumberjack  *lumberjack.Logger      `yaml:"lumberjack" json:"lumberjack"`
	} `yaml:"zapLogger" json:"zapLogger"`
}

// ZapLoggerEntry contains bellow fields.
// 1: EntryName: Name of entry.
// 2: EntryType: Type of entry which is ZapLoggerEntryType.
// 3: EntryDescription: Description of ZapLoggerEntry.
// 4: Logger: zap.Logger which was initialized at the beginning.
// 5: LoggerConfig: zap.Logger config which was initialized at the beginning which is not accessible after initialization..
// 6: LumberjackConfig: lumberjack.Logger which was initialized at the beginning.
type ZapLoggerEntry struct {
	EntryName        string             `yaml:"entryName" json:"entryName"`
	EntryType        string             `yaml:"entryType" json:"entryType"`
	EntryDescription string             `yaml:"entryDescription" json:"entryDescription"`
	Logger           *zap.Logger        `yaml:"-" json:"-"`
	LoggerConfig     *zap.Config        `yaml:"zapConfig" json:"zapConfig"`
	LumberjackConfig *lumberjack.Logger `yaml:"lumberjackConfig" json:"lumberjackConfig"`
}

// ZapLoggerEntry Option which used while registering entry from codes.
type ZapLoggerEntryOption func(*ZapLoggerEntry)

// Provide name of entry.
func WithNameZap(name string) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.EntryName = name
	}
}

// Provide description of entry.
func WithDescriptionZap(description string) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.EntryDescription = description
	}
}

// Provide zap logger related entity of entry.
func WithLoggerZap(logger *zap.Logger, loggerConfig *zap.Config, lumberjackConfig *lumberjack.Logger) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.Logger = logger
		entry.LoggerConfig = loggerConfig
		entry.LumberjackConfig = lumberjackConfig
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
			WithDescriptionZap(element.Description),
			WithLoggerZap(appLogger, appLoggerConfig, appLoggerLumberjackConfig))

		res[element.Name] = entry
	}

	return res
}

// Crate event logger entry with options.
func RegisterZapLoggerEntry(opts ...ZapLoggerEntryOption) *ZapLoggerEntry {
	entry := &ZapLoggerEntry{
		EntryType:        ZapLoggerEntryType,
		EntryDescription: ZapLoggerDescription,
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.Logger == nil {
		entry.LoggerConfig = rklogger.NewZapStdoutConfig()
		if logger, err := entry.LoggerConfig.Build(); err != nil {
			rkcommon.ShutdownWithError(err)
		} else {
			entry.Logger = logger
		}
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "zapLogger-" + rkcommon.RandString(4)
	}

	GlobalAppCtx.AddZapLoggerEntry(entry)

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
	return entry.EntryName
}

// Get type of entry.
func (entry *ZapLoggerEntry) GetType() string {
	return entry.EntryType
}

// Return description of entry
func (entry *ZapLoggerEntry) GetDescription() string {
	return entry.EntryDescription
}

// Convert entry into JSON style string.
func (entry *ZapLoggerEntry) String() string {
	if bytes, err := json.Marshal(entry); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

// Marshal entry.
func (entry *ZapLoggerEntry) MarshalJSON() ([]byte, error) {
	loggerConfigWrap := rklogger.TransformToZapConfigWrap(entry.LoggerConfig)

	type innerZapLoggerEntry struct {
		EntryName        string                  `yaml:"entryName" json:"entryName"`
		EntryType        string                  `yaml:"entryType" json:"entryType"`
		EntryDescription string                  `yaml:"entryDescription" json:"entryDescription"`
		LoggerConfig     *rklogger.ZapConfigWrap `yaml:"zapConfig" json:"zapConfig"`
		LumberjackConfig *lumberjack.Logger      `yaml:"lumberjackConfig" json:"lumberjackConfig"`
	}

	return json.Marshal(&innerZapLoggerEntry{
		EntryName:        entry.EntryName,
		EntryType:        entry.EntryType,
		EntryDescription: entry.EntryDescription,
		LoggerConfig:     loggerConfigWrap,
		LumberjackConfig: entry.LumberjackConfig,
	})
}

// Not supported.
func (entry *ZapLoggerEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Get zap logger, refer to zap.Logger.
func (entry *ZapLoggerEntry) GetLogger() *zap.Logger {
	return entry.Logger
}

// Get zap logger config, refer to zap.Config.
func (entry *ZapLoggerEntry) GetLoggerConfig() *zap.Config {
	return entry.LoggerConfig
}

// Get lumberjack config, refer to lumberjack.Logger.
func (entry *ZapLoggerEntry) GetLumberjackConfig() *lumberjack.Logger {
	return entry.LumberjackConfig
}
