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
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

// NewEventEntryNoop create event logger entry with noop event factory.
// Event factory and event helper will be created with noop zap logger.
// Since we don't need any log rotation in case of noop, lumberjack config will be nil.
func NewEventEntryNoop() *EventEntry {
	entry := &EventEntry{
		entryName:        "EventLoggerNoop",
		entryType:        "EventLoggerEntry",
		entryDescription: "Internal RK entry which is used to log event such as RPC request or periodic jobs.",
		EventFactory:     rkquery.NewEventFactory(rkquery.WithZapLogger(rklogger.NoopLogger)),
	}

	entry.EventHelper = rkquery.NewEventHelper(entry.EventFactory)

	return entry
}

func NewEventEntryStdout() *EventEntry {
	entry := &EventEntry{
		entryName:        "EventLoggerNoop",
		entryType:        "EventLoggerEntry",
		entryDescription: "Internal RK entry which is used to log event such as RPC request or periodic jobs.",
		EventFactory:     rkquery.NewEventFactory(rkquery.WithZapLogger(rklogger.EventLogger)),
		LoggerConfig:     rklogger.EventLoggerConfig,
		baseLogger:       rklogger.EventLogger,
	}

	entry.EventHelper = rkquery.NewEventHelper(entry.EventFactory)

	return entry
}

// RegisterEventEntry create event logger entry with options.
func RegisterEventEntry(boot *BootEvent) []*EventEntry {
	res := make([]*EventEntry, 0)

	for _, event := range boot.Event {
		if len(event.Locale) < 1 {
			event.Locale = "*::*::*::*"
		}

		if len(event.Name) < 1 || !IsLocaleValid(event.Locale) {
			continue
		}

		entry := &EventEntry{
			entryName:        event.Name,
			entryType:        EventEntryType,
			entryDescription: event.Description,
		}

		var eventFactory *rkquery.EventFactory
		var lokiSyncer *rklogger.LokiSyncer

		// Assign default zap config and lumberjack config
		eventLoggerConfig := rklogger.NewZapEventConfig()
		eventLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()

		// Override with user provided zap config and lumberjack config
		overrideLumberjackConfig(eventLoggerLumberjackConfig, event.Lumberjack)

		// If output paths were provided by user, we will override it which means <stdout> would be omitted
		if len(event.OutputPaths) > 0 {
			eventLoggerConfig.OutputPaths = event.OutputPaths
		}

		// Loki Syncer
		syncers := make([]zapcore.WriteSyncer, 0)
		if event.Loki.Enabled {
			opts := []rklogger.LokiSyncerOption{
				rklogger.WithLokiAddr(event.Loki.Addr),
				rklogger.WithLokiPath(event.Loki.Path),
				rklogger.WithLokiUsername(event.Loki.Username),
				rklogger.WithLokiPassword(event.Loki.Password),
				rklogger.WithLokiMaxBatchSize(event.Loki.MaxBatchSize),
				rklogger.WithLokiMaxBatchWaitMs(time.Duration(event.Loki.MaxBatchWaitMs) * time.Millisecond),
			}

			// default labels
			opts = append(opts,
				rklogger.WithLokiLabel(rkmid.Realm.Key, rkmid.Realm.String),
				rklogger.WithLokiLabel(rkmid.Region.Key, rkmid.Region.String),
				rklogger.WithLokiLabel(rkmid.AZ.Key, rkmid.AZ.String),
				rklogger.WithLokiLabel(rkmid.Domain.Key, rkmid.Domain.String),
				rklogger.WithLokiLabel("app_name", GlobalAppCtx.GetAppInfoEntry().AppName),
				rklogger.WithLokiLabel("app_version", GlobalAppCtx.GetAppInfoEntry().Version),
				rklogger.WithLokiLabel("logger_type", "event"),
			)

			// labels
			for k, v := range event.Loki.Labels {
				opts = append(opts, rklogger.WithLokiLabel(k, v))
			}

			if event.Loki.InsecureSkipVerify {
				opts = append(opts, rklogger.WithLokiClientTls(&tls.Config{
					InsecureSkipVerify: true,
				}))
			}

			lokiSyncer = rklogger.NewLokiSyncer(opts...)
			syncers = append(syncers, lokiSyncer)
		}

		var eventLogger *zap.Logger
		var err error
		if eventLogger, err = rklogger.NewZapLoggerWithConfAndSyncer(eventLoggerConfig, eventLoggerLumberjackConfig, syncers); err != nil {
			ShutdownWithError(err)
		} else {
			eventFactory = rkquery.NewEventFactory(
				rkquery.WithZapLogger(eventLogger),
				rkquery.WithAppName(GlobalAppCtx.GetAppInfoEntry().AppName),
				rkquery.WithAppVersion(GlobalAppCtx.GetAppInfoEntry().Version),
				rkquery.WithEncoding(rkquery.ToEncoding(event.Encoding)))
		}

		entry.EventFactory = eventFactory
		entry.EventHelper = rkquery.NewEventHelper(eventFactory)
		entry.lokiSyncer = lokiSyncer
		entry.baseLogger = eventLogger
		entry.LoggerConfig = eventLoggerConfig
		entry.LumberjackConfig = eventLoggerLumberjackConfig

		GlobalAppCtx.AddEntry(entry)
		res = append(res, entry)
	}

	return res
}

// RegisterEventEntryYAML register function
func RegisterEventEntryYAML(raw []byte) map[string]Entry {
	boot := &BootEvent{}
	UnmarshalBootYAML(raw, boot)

	res := map[string]Entry{}

	entries := RegisterEventEntry(boot)
	for i := range entries {
		entry := entries[i]
		res[entry.GetName()] = entry
	}

	return res
}

// BootEvent bootstrap config of Event Logger information.
type BootEvent struct {
	Event []*BootEventE `yaml:"event" json:"event"`
}

type BootLoki struct {
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

type BootEventE struct {
	Name        string             `yaml:"name" json:"name"`
	Description string             `yaml:"description" json:"description"`
	Locale      string             `yaml:"locale" json:"locale"`
	Encoding    string             `yaml:"encoding" json:"encoding"`
	OutputPaths []string           `yaml:"outputPaths" json:"outputPaths"`
	Lumberjack  *lumberjack.Logger `yaml:"lumberjack" json:"lumberjack"`
	Loki        BootLoki           `yaml:"loki" json:"loki"`
}

// EventEntry contains bellow fields.
type EventEntry struct {
	*rkquery.EventFactory
	*rkquery.EventHelper
	entryName        string               `yaml:"-" json:"-"`
	entryType        string               `yaml:"-" json:"-"`
	entryDescription string               `yaml:"-" json:"-"`
	LoggerConfig     *zap.Config          `yaml:"-" json:"-"`
	LumberjackConfig *lumberjack.Logger   `yaml:"-" json:"-"`
	lokiSyncer       *rklogger.LokiSyncer `yaml:"-" json:"-"`
	baseLogger       *zap.Logger          `yaml:"-" json:"-"`
}

// Bootstrap entry.
func (entry *EventEntry) Bootstrap(ctx context.Context) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.Bootstrap(ctx)
	}
}

// Interrupt entry.
func (entry *EventEntry) Interrupt(ctx context.Context) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.Interrupt(ctx)
	}
}

// GetName returns name of entry.
func (entry *EventEntry) GetName() string {
	return entry.entryName
}

// GetType returns type of entry.
func (entry *EventEntry) GetType() string {
	return entry.entryType
}

// String convert entry into JSON style string.
func (entry *EventEntry) String() string {
	var bytes []byte
	var err error

	if bytes, err = json.Marshal(entry); err != nil {
		return "{}"
	}

	return string(bytes)
}

// MarshalJSON marshal entry.
func (entry *EventEntry) MarshalJSON() ([]byte, error) {
	loggerConfigWrap := rklogger.TransformToZapConfigWrap(entry.LoggerConfig)

	type innerEventLoggerEntry struct {
		EntryName        string                  `yaml:"name" json:"name"`
		EntryType        string                  `yaml:"type" json:"type"`
		EntryDescription string                  `yaml:"description" json:"description"`
		LoggerConfig     *rklogger.ZapConfigWrap `yaml:"zapConfig" json:"zapConfig"`
		LumberjackConfig *lumberjack.Logger      `yaml:"lumberjackConfig" json:"lumberjackConfig"`
	}

	return json.Marshal(&innerEventLoggerEntry{
		EntryName:        entry.entryName,
		EntryType:        entry.entryType,
		EntryDescription: entry.entryDescription,
		LoggerConfig:     loggerConfigWrap,
		LumberjackConfig: entry.LumberjackConfig,
	})
}

// UnmarshalJSON not supported.
func (entry *EventEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetDescription return description of entry.
func (entry *EventEntry) GetDescription() string {
	return entry.entryDescription
}

// AddEntryLabelToLokiSyncer add entry name entry type into loki syncer
func (entry *EventEntry) AddEntryLabelToLokiSyncer(e Entry) {
	if entry.lokiSyncer != nil && e != nil {
		entry.lokiSyncer.AddLabel("entry_name", e.GetName())
		entry.lokiSyncer.AddLabel("entry_type", e.GetType())
	}
}

// AddLabelToLokiSyncer add key value pair as label into loki syncer
func (entry *EventEntry) AddLabelToLokiSyncer(k, v string) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.AddLabel(k, v)
	}
}

// Sync underlying logger
func (entry *EventEntry) Sync() {
	if entry.baseLogger != nil {
		entry.baseLogger.Sync()
	}
}
