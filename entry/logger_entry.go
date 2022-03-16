// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

// NewLoggerEntryNoop create zap logger entry with noop.
func NewLoggerEntryNoop() *LoggerEntry {
	return &LoggerEntry{
		entryName:        "LoggerEntryNoop",
		entryType:        LoggerEntryType,
		entryDescription: "Internal RK entry which is used for noop logging with zap.Logger.",
		Logger:           rklogger.NoopLogger,
		LoggerConfig:     nil,
		LumberjackConfig: nil,
	}
}

// NewLoggerEntryStdout create zap logger entry with STDOUT.
func NewLoggerEntryStdout() *LoggerEntry {
	return &LoggerEntry{
		entryName:        "LoggerEntryStdout",
		entryType:        LoggerEntryType,
		entryDescription: "Internal RK entry which is used for noop logging with zap.Logger.",
		Logger:           rklogger.StdoutLogger,
		LoggerConfig:     rklogger.StdoutLoggerConfig,
		LumberjackConfig: nil,
	}
}

// RegisterLoggerEntry create event logger entry with options.
func RegisterLoggerEntry(boot *BootLogger) []*LoggerEntry {
	res := make([]*LoggerEntry, 0)

	// filter out based domain
	configMap := make(map[string]*BootLoggerE)
	for _, config := range boot.Logger {
		if len(config.Name) < 1 {
			continue
		}

		if !IsValidDomain(config.Domain) {
			continue
		}

		// * or matching domain
		// 1: add it to map if missing
		if _, ok := configMap[config.Name]; !ok {
			configMap[config.Name] = config
			continue
		}

		// 2: already has an entry, then compare domain,
		//    only one case would occur, previous one is already the correct one, continue
		if config.Domain == "" || config.Domain == "*" {
			continue
		}

		configMap[config.Name] = config
	}

	for _, logger := range configMap {
		entry := &LoggerEntry{
			entryName:        logger.Name,
			entryType:        LoggerEntryType,
			entryDescription: logger.Description,
		}

		// Assign default zap config and lumberjack config
		zapLoggerConfig := rklogger.NewZapStdoutConfig()
		zapLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()

		// Override with user provided zap config and lumberjack config
		overrideZapConfig(zapLoggerConfig, rklogger.TransformToZapConfig(logger.Zap))
		overrideLumberjackConfig(zapLoggerLumberjackConfig, logger.Lumberjack)

		// Loki Syncer
		syncers := make([]zapcore.WriteSyncer, 0)
		var lokiSyncer *rklogger.LokiSyncer
		if logger.Loki.Enabled {
			opts := []rklogger.LokiSyncerOption{
				rklogger.WithLokiAddr(logger.Loki.Addr),
				rklogger.WithLokiPath(logger.Loki.Path),
				rklogger.WithLokiUsername(logger.Loki.Username),
				rklogger.WithLokiPassword(logger.Loki.Password),
				rklogger.WithLokiMaxBatchSize(logger.Loki.MaxBatchSize),
				rklogger.WithLokiMaxBatchWaitMs(time.Duration(logger.Loki.MaxBatchWaitMs) * time.Millisecond),
			}

			// labels
			for k, v := range logger.Loki.Labels {
				opts = append(opts, rklogger.WithLokiLabel(k, v))
			}

			// default labels
			opts = append(opts,
				rklogger.WithLokiLabel(rkmid.Domain.Key, rkmid.Domain.String),
				rklogger.WithLokiLabel("app_name", GlobalAppCtx.GetAppInfoEntry().AppName),
				rklogger.WithLokiLabel("app_version", GlobalAppCtx.GetAppInfoEntry().Version),
				rklogger.WithLokiLabel("logger_type", "zap"),
			)

			if logger.Loki.InsecureSkipVerify {
				opts = append(opts, rklogger.WithLokiClientTls(&tls.Config{
					InsecureSkipVerify: true,
				}))
			}

			lokiSyncer = rklogger.NewLokiSyncer(opts...)
			syncers = append(syncers, lokiSyncer)
		}

		// Create app logger with config
		zapLogger, err := rklogger.NewZapLoggerWithConfAndSyncer(zapLoggerConfig, zapLoggerLumberjackConfig, syncers, zap.AddCaller())

		if err != nil {
			ShutdownWithError(err)
		}

		entry.Logger = zapLogger
		entry.LoggerConfig = zapLoggerConfig
		entry.LumberjackConfig = zapLoggerLumberjackConfig
		entry.lokiSyncer = lokiSyncer

		GlobalAppCtx.AddEntry(entry)
		res = append(res, entry)
	}

	return res
}

// RegisterLoggerEntryYAML register function
func RegisterLoggerEntryYAML(raw []byte) map[string]Entry {
	boot := &BootLogger{}
	UnmarshalBootYAML(raw, boot)

	res := map[string]Entry{}

	entries := RegisterLoggerEntry(boot)
	for i := range entries {
		entry := entries[i]
		res[entry.GetName()] = entry
	}

	return res
}

// BootLogger bootstrap config of Zap Logger information.
type BootLogger struct {
	Logger []*BootLoggerE `json:"logger" yaml:"logger"`
}

// BootLoggerE bootstrap element of LoggerEntry
type BootLoggerE struct {
	Name        string                  `yaml:"name" json:"name"`
	Description string                  `yaml:"description" json:"description"`
	Domain      string                  `yaml:"domain" json:"domain"`
	Zap         *rklogger.ZapConfigWrap `yaml:"zap" json:"zap"`
	Lumberjack  *lumberjack.Logger      `yaml:"lumberjack" json:"lumberjack"`
	Loki        BootLoki                `yaml:"loki" json:"loki"`
}

// LoggerEntry contains bellow fields.
type LoggerEntry struct {
	*zap.Logger
	entryName        string               `yaml:"name" json:"name"`
	entryType        string               `yaml:"type" json:"type"`
	entryDescription string               `yaml:"description" json:"description"`
	LoggerConfig     *zap.Config          `yaml:"zapConfig" json:"zapConfig"`
	LumberjackConfig *lumberjack.Logger   `yaml:"lumberjackConfig" json:"lumberjackConfig"`
	lokiSyncer       *rklogger.LokiSyncer `yaml:"lokiSyncer" json:"lokiSyncer"`
}

// Bootstrap entry.
func (entry *LoggerEntry) Bootstrap(ctx context.Context) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.Bootstrap(ctx)
	}
}

// Interrupt entry.
func (entry *LoggerEntry) Interrupt(ctx context.Context) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.Interrupt(ctx)
	}
}

// GetName returns name of entry.
func (entry *LoggerEntry) GetName() string {
	return entry.entryName
}

// GetType returns type of entry.
func (entry *LoggerEntry) GetType() string {
	return entry.entryType
}

// GetDescription returns description of entry.
func (entry *LoggerEntry) GetDescription() string {
	return entry.entryDescription
}

// String convert entry into JSON style string.
func (entry *LoggerEntry) String() string {
	var bytes []byte
	var err error
	if bytes, err = json.Marshal(entry); err != nil {
		return "{}"
	}

	return string(bytes)
}

// MarshalJSON marshal entry.
func (entry *LoggerEntry) MarshalJSON() ([]byte, error) {
	loggerConfigWrap := rklogger.TransformToZapConfigWrap(entry.LoggerConfig)

	type innerZapLoggerEntry struct {
		EntryName        string                  `yaml:"name" json:"name"`
		EntryType        string                  `yaml:"type" json:"type"`
		EntryDescription string                  `yaml:"description" json:"description"`
		LoggerConfig     *rklogger.ZapConfigWrap `yaml:"zapConfig" json:"zapConfig"`
		LumberjackConfig *lumberjack.Logger      `yaml:"lumberjackConfig" json:"lumberjackConfig"`
	}

	return json.Marshal(&innerZapLoggerEntry{
		EntryName:        entry.entryName,
		EntryType:        entry.entryType,
		EntryDescription: entry.entryDescription,
		LoggerConfig:     loggerConfigWrap,
		LumberjackConfig: entry.LumberjackConfig,
	})
}

// UnmarshalJSON not supported.
func (entry *LoggerEntry) UnmarshalJSON([]byte) error {
	return nil
}

// AddEntryLabelToLokiSyncer add entry name entry type into loki syncer
func (entry *LoggerEntry) AddEntryLabelToLokiSyncer(e Entry) {
	if entry.lokiSyncer != nil && e != nil {
		entry.lokiSyncer.AddLabel("entry_name", e.GetName())
		entry.lokiSyncer.AddLabel("entry_type", e.GetType())
	}
}

// AddLabelToLokiSyncer add key value pair as label into loki syncer
func (entry *LoggerEntry) AddLabelToLokiSyncer(k, v string) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.AddLabel(k, v)
	}
}

// Sync underlying logger
func (entry *LoggerEntry) Sync() {
	if entry.Logger != nil {
		entry.Logger.Sync()
	}
}
