// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/rookie-ninja/rk-common/common"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

const (
	// ZapLoggerEntryType name of entry
	ZapLoggerEntryType = "ZapLoggerEntry"
	// ZapLoggerNameNoop type of entry
	ZapLoggerNameNoop = "ZapLoggerNoop"
	// ZapLoggerDescription description of entry
	ZapLoggerDescription = "Internal RK entry which is used for logging with zap.Logger."
)

// NoopZapLoggerEntry create zap logger entry with noop.
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

// BootConfigZapLogger bootstrap config of Zap Logger information.
type BootConfigZapLogger struct {
	ZapLogger []struct {
		Name        string                  `yaml:"name" json:"name"`
		Description string                  `yaml:"description" json:"description"`
		Zap         *rklogger.ZapConfigWrap `yaml:"zap" json:"zap"`
		Lumberjack  *lumberjack.Logger      `yaml:"lumberjack" json:"lumberjack"`
		Loki        BootConfigLoki          `yaml:"loki" json:"loki"`
	} `yaml:"zapLogger" json:"zapLogger"`
}

type BootConfigLoki struct {
	Enabled            bool              `yaml:"enabled" json:"enabled"`
	Addr               string            `yaml:"addr" json:"addr"`
	Path               string            `yaml:"path" json:"path"`
	Username           string            `yaml:"username" json:"username"`
	Password           string            `yaml:"password" json:"password"`
	InsecureSkipVerify bool              `yaml:"insecureSkipVerify" json:"insecureSkipVerify"`
	Labels             map[string]string `yaml:"labels" json:"labels"`
	MaxBatchWaitMs     int               `yaml:"maxBatchWaitMs" json:"maxBatchWaitMs"`
	MaxBatchSize       int               `yaml:"maxBatchSize" json:"maxBatchSize"`
}

// ZapLoggerEntry contains bellow fields.
type ZapLoggerEntry struct {
	EntryName        string               `yaml:"entryName" json:"entryName"`
	EntryType        string               `yaml:"entryType" json:"entryType"`
	EntryDescription string               `yaml:"entryDescription" json:"entryDescription"`
	Logger           *zap.Logger          `yaml:"-" json:"-"`
	LoggerConfig     *zap.Config          `yaml:"zapConfig" json:"zapConfig"`
	LumberjackConfig *lumberjack.Logger   `yaml:"lumberjackConfig" json:"lumberjackConfig"`
	lokiSyncer       *rklogger.LokiSyncer `yaml:"lokiSyncer" json:"lokiSyncer"`
}

// ZapLoggerEntryOption Option which used while registering entry from codes.
type ZapLoggerEntryOption func(*ZapLoggerEntry)

// WithNameZap provide name of entry.
func WithNameZap(name string) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.EntryName = name
	}
}

// WithDescriptionZap provide description of entry.
func WithDescriptionZap(description string) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.EntryDescription = description
	}
}

// WithLoggerZap provide zap logger related entity of entry.
func WithLoggerZap(logger *zap.Logger, loggerConfig *zap.Config, lumberjackConfig *lumberjack.Logger) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.Logger = logger
		entry.LoggerConfig = loggerConfig
		entry.LumberjackConfig = lumberjackConfig
	}
}

// WithLokiSyncerZap provide rklogger.LokiSyncer
func WithLokiSyncerZap(loki *rklogger.LokiSyncer) ZapLoggerEntryOption {
	return func(entry *ZapLoggerEntry) {
		entry.lokiSyncer = loki
	}
}

// RegisterZapLoggerEntriesWithConfig create zap logger entries with config file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterZapLoggerEntriesWithConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: Unmarshal user provided config into boot config struct
	config := &BootConfigZapLogger{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init zap logger entries with boot config
	for i := range config.ZapLogger {
		element := config.ZapLogger[i]

		// Assign default zap config and lumberjack config
		appLoggerConfig := rklogger.NewZapStdoutConfig()
		appLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()

		// Override with user provided zap config and lumberjack config
		rkcommon.OverrideZapConfig(appLoggerConfig, rklogger.TransformToZapConfig(element.Zap))
		rkcommon.OverrideLumberjackConfig(appLoggerLumberjackConfig, element.Lumberjack)

		// Loki Syncer
		syncers := make([]zapcore.WriteSyncer, 0)
		var lokiSyncer *rklogger.LokiSyncer
		if element.Loki.Enabled {
			opts := []rklogger.LokiSyncerOption{
				rklogger.WithLokiAddr(element.Loki.Addr),
				rklogger.WithLokiPath(element.Loki.Path),
				rklogger.WithLokiUsername(element.Loki.Username),
				rklogger.WithLokiPassword(element.Loki.Password),
				rklogger.WithLokiMaxBatchSize(element.Loki.MaxBatchSize),
				rklogger.WithLokiMaxBatchWaitMs(time.Duration(element.Loki.MaxBatchWaitMs) * time.Millisecond),
			}

			// labels
			for k, v := range element.Loki.Labels {
				opts = append(opts, rklogger.WithLokiLabel(k, v))
			}

			// default labels
			opts = append(opts,
				rklogger.WithLokiLabel(rkmid.Realm.Key, rkmid.Realm.String),
				rklogger.WithLokiLabel(rkmid.Region.Key, rkmid.Region.String),
				rklogger.WithLokiLabel(rkmid.AZ.Key, rkmid.AZ.String),
				rklogger.WithLokiLabel(rkmid.Domain.Key, rkmid.Domain.String),
				rklogger.WithLokiLabel("app_name", GlobalAppCtx.GetAppInfoEntry().AppName),
				rklogger.WithLokiLabel("app_version", GlobalAppCtx.GetAppInfoEntry().Version),
				rklogger.WithLokiLabel("logger_type", "zap"),
			)

			if element.Loki.InsecureSkipVerify {
				opts = append(opts, rklogger.WithLokiClientTls(&tls.Config{
					InsecureSkipVerify: true,
				}))
			}

			lokiSyncer = rklogger.NewLokiSyncer(opts...)
			syncers = append(syncers, lokiSyncer)
		}

		// Create app logger with config
		appLogger, err := rklogger.NewZapLoggerWithConfAndSyncer(appLoggerConfig, appLoggerLumberjackConfig, syncers)

		if err != nil {
			rkcommon.ShutdownWithError(err)
		}

		entry := RegisterZapLoggerEntry(
			WithNameZap(element.Name),
			WithDescriptionZap(element.Description),
			WithLoggerZap(appLogger, appLoggerConfig, appLoggerLumberjackConfig),
			WithLokiSyncerZap(lokiSyncer))

		res[element.Name] = entry
	}

	return res
}

// RegisterZapLoggerEntry create event logger entry with options.
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
func (entry *ZapLoggerEntry) Bootstrap(ctx context.Context) {
	if entry.lokiSyncer != nil {
		fmt.Println("boot strapping zap logger")
		entry.lokiSyncer.Bootstrap(ctx)
	}
}

// Interrupt entry.
func (entry *ZapLoggerEntry) Interrupt(ctx context.Context) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.Interrupt(ctx)
	}
}

// GetName returns name of entry.
func (entry *ZapLoggerEntry) GetName() string {
	return entry.EntryName
}

// GetType returns type of entry.
func (entry *ZapLoggerEntry) GetType() string {
	return entry.EntryType
}

// GetDescription returns description of entry.
func (entry *ZapLoggerEntry) GetDescription() string {
	return entry.EntryDescription
}

// String convert entry into JSON style string.
func (entry *ZapLoggerEntry) String() string {
	var bytes []byte
	var err error
	if bytes, err = json.Marshal(entry); err != nil {
		return "{}"
	}

	return string(bytes)
}

// MarshalJSON marshal entry.
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

// UnmarshalJSON not supported.
func (entry *ZapLoggerEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetLogger returns zap logger, refer to zap.Logger.
func (entry *ZapLoggerEntry) GetLogger() *zap.Logger {
	return entry.Logger
}

// GetLoggerConfig returns zap logger config, refer to zap.Config.
func (entry *ZapLoggerEntry) GetLoggerConfig() *zap.Config {
	return entry.LoggerConfig
}

// GetLumberjackConfig returns lumberjack config, refer to lumberjack.Logger.
func (entry *ZapLoggerEntry) GetLumberjackConfig() *lumberjack.Logger {
	return entry.LumberjackConfig
}

// AddEntryLabelToLokiSyncer add entry name entry type into loki syncer
func (entry *ZapLoggerEntry) AddEntryLabelToLokiSyncer(e Entry) {
	if entry.lokiSyncer != nil && e != nil {
		entry.lokiSyncer.AddLabel("entry_name", e.GetName())
		entry.lokiSyncer.AddLabel("entry_type", e.GetType())
	}
}

// AddLabelToLokiSyncer add key value pair as label into loki syncer
func (entry *ZapLoggerEntry) AddLabelToLokiSyncer(k, v string) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.AddLabel(k, v)
	}
}
