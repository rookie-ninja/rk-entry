// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

const LoggerEntryType = "LoggerEntry"

// NewLoggerEntryNoop create zap logger entry with noop.
func NewLoggerEntryNoop() *LoggerEntry {
	return &LoggerEntry{
		entryName:        "ZapLoggerEntryNoop",
		entryType:        "ZapLoggerEntry",
		entryDescription: "Internal RK entry which is used for noop logging with zap.Logger.",
		Logger:           rklogger.NoopLogger,
		LoggerConfig:     nil,
		LumberjackConfig: nil,
	}
}

// NewLoggerEntryStdout create zap logger entry with STDOUT.
func NewLoggerEntryStdout() *LoggerEntry {
	return &LoggerEntry{
		entryName:        "ZapLoggerEntryStdout",
		entryType:        "ZapLoggerEntry",
		entryDescription: "Internal RK entry which is used for noop logging with zap.Logger.",
		Logger:           rklogger.StdoutLogger,
		LoggerConfig:     rklogger.StdoutLoggerConfig,
		LumberjackConfig: nil,
	}
}

// RegisterLoggerEntry create event logger entry with options.
func RegisterLoggerEntry(boot *BootLogger) []*LoggerEntry {
	res := make([]*LoggerEntry, 0)

	for _, logger := range boot.Logger {
		if len(logger.Locale) < 1 {
			logger.Locale = "*::*::*::*"
		}

		if !IsLocaleValid(logger.Locale) {
			continue
		}

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
				rklogger.WithLokiLabel(rkmid.Realm.Key, rkmid.Realm.String),
				rklogger.WithLokiLabel(rkmid.Region.Key, rkmid.Region.String),
				rklogger.WithLokiLabel(rkmid.AZ.Key, rkmid.AZ.String),
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

		res = append(res, entry)
	}

	return res
}

// registerLoggerEntry register function
func registerLoggerEntry(raw []byte) map[string]Entry {
	boot := &BootLogger{}
	UnmarshalBoot(raw, boot)

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

type BootLoggerE struct {
	Name        string                  `yaml:"name" json:"name"`
	Description string                  `yaml:"description" json:"description"`
	Locale      string                  `yaml:"locale" json:"locale"`
	Zap         *rklogger.ZapConfigWrap `yaml:"zap" json:"zap"`
	Lumberjack  *lumberjack.Logger      `yaml:"lumberjack" json:"lumberjack"`
	Loki        BootLoki                `yaml:"loki" json:"loki"`
}

// LoggerEntry contains bellow fields.
type LoggerEntry struct {
	entryName        string               `yaml:"name" json:"name"`
	entryType        string               `yaml:"type" json:"type"`
	entryDescription string               `yaml:"description" json:"description"`
	Logger           *zap.Logger          `yaml:"-" json:"-"`
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

func (entry *LoggerEntry) WithOptions(opts ...zap.Option) *zap.Logger {
	return entry.Logger.WithOptions(opts...)
}

func (entry *LoggerEntry) With(fields ...zap.Field) *zap.Logger {
	return entry.Logger.With(fields...)
}

func (entry *LoggerEntry) Debug(msg string, fields ...zap.Field) {
	entry.Logger.Debug(msg, fields...)
}

func (entry *LoggerEntry) Info(msg string, fields ...zap.Field) {
	entry.Logger.Info(msg, fields...)
}

func (entry *LoggerEntry) Warn(msg string, fields ...zap.Field) {
	entry.Logger.Warn(msg, fields...)
}

func (entry *LoggerEntry) Error(msg string, fields ...zap.Field) {
	entry.Logger.Error(msg, fields...)
}

func (entry *LoggerEntry) DPanic(msg string, fields ...zap.Field) {
	entry.Logger.DPanic(msg, fields...)
}

func (entry *LoggerEntry) Panic(msg string, fields ...zap.Field) {
	entry.Logger.Panic(msg, fields...)
}

func (entry *LoggerEntry) Fatal(msg string, fields ...zap.Field) {
	entry.Logger.Fatal(msg, fields...)
}
