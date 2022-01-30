// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

const (
	// EventLoggerEntryType type of entry
	EventLoggerEntryType = "EventLoggerEntry"
	// EventLoggerNameNoop name of noop entry
	EventLoggerNameNoop = "EventLoggerNoop"
	// EventLoggerDescription of entry description
	EventLoggerDescription = "Internal RK entry which is used to log event such as RPC request or periodic jobs."
)

// NoopEventLoggerEntry create event logger entry with noop event factory.
// Event factory and event helper will be created with noop zap logger.
// Since we don't need any log rotation in case of noop, lumberjack config will be nil.
func NoopEventLoggerEntry() *EventLoggerEntry {
	entry := &EventLoggerEntry{
		EntryName:        EventLoggerNameNoop,
		EntryType:        EventLoggerEntryType,
		EntryDescription: EventLoggerDescription,
		EventFactory:     rkquery.NewEventFactory(rkquery.WithZapLogger(rklogger.NoopLogger)),
	}

	entry.EventHelper = rkquery.NewEventHelper(entry.EventFactory)

	return entry
}

// BootConfigEventLogger bootstrap config of Event Logger information.
// 1: RK.AppName: Application name which refers to go process.
//                Default application name of AppNameDefault would be assigned if missing in boot config file.
// 2: RK.Version: Version of application.
// 3: EventLogger.Name: Name of event logger entry.
// 4: EventLogger.Description: Description of event logger entry.
// 5: EventLogger.Format: Format of event logger, RK & JSON is supported. Please refer rkquery.RK & rkquery.JSON.
// 6: EventLogger.OutputPaths: Output paths of event logger, stdout would be the default one if not provided.
//                             If one of output path was provided, then stdout would be omitted.
//                             Output path could be relative or absolute paths either.
// 7: EventLogger.Lumberjack: Lumberjack config which follows lumberjack.Logger style.
type BootConfigEventLogger struct {
	EventLogger []struct {
		Name        string             `yaml:"name" json:"name"`
		Description string             `yaml:"description" json:"description"`
		Encoding    string             `yaml:"encoding" json:"encoding"`
		OutputPaths []string           `yaml:"outputPaths" json:"outputPaths"`
		Lumberjack  *lumberjack.Logger `yaml:"lumberjack" json:"lumberjack"`
		Loki        BootConfigLoki     `yaml:"loki" json:"loki"`
	} `yaml:"eventLogger json:"eventLogger"`
}

// EventLoggerEntry contains bellow fields.
type EventLoggerEntry struct {
	EntryName        string                `yaml:"entryName" json:"entryName"`
	EntryType        string                `yaml:"entryType" json:"entryType"`
	EntryDescription string                `yaml:"entryDescription" json:"entryDescription"`
	EventFactory     *rkquery.EventFactory `yaml:"-" json:"-"`
	EventHelper      *rkquery.EventHelper  `yaml:"-" json:"-"`
	LoggerConfig     *zap.Config           `yaml:"zapConfig" json:"zapConfig"`
	LumberjackConfig *lumberjack.Logger    `yaml:"lumberjackConfig" json:"lumberjackConfig"`
	lokiSyncer       *rklogger.LokiSyncer  `yaml:"lokiSyncer" json:"lokiSyncer"`
	baseLogger       *zap.Logger           `yaml:"-" json:"-"`
}

// EventLoggerEntryOption Option which used while registering entry from codes.
type EventLoggerEntryOption func(*EventLoggerEntry)

// WithNameEvent provide name of entry.
func WithNameEvent(name string) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.EntryName = name
	}
}

// WithDescriptionEvent provide description of entry.
func WithDescriptionEvent(description string) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.EntryDescription = description
	}
}

// WithEventFactoryEvent provide event factory of entry which refers to rkquery.EventFactory.
func WithEventFactoryEvent(fac *rkquery.EventFactory) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.EventFactory = fac
	}
}

// WithLokiSyncerEvent provide rklogger.LokiSyncer
func WithLokiSyncerEvent(loki *rklogger.LokiSyncer) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.lokiSyncer = loki
	}
}

// WithBaseLoggerEvent provide zap.Logger
func WithBaseLoggerEvent(base *zap.Logger) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.baseLogger = base
	}
}

// RegisterEventLoggerEntriesWithConfig create event logger entries with config file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterEventLoggerEntriesWithConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: Unmarshal user provided config into boot config struct
	config := &BootConfigEventLogger{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init event logger entries with boot config
	for i := range config.EventLogger {
		element := config.EventLogger[i]

		var eventFactory *rkquery.EventFactory
		// Assign default zap config and lumberjack config
		eventLoggerConfig := rklogger.NewZapEventConfig()
		eventLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()
		// Override with user provided zap config and lumberjack config
		rkcommon.OverrideLumberjackConfig(eventLoggerLumberjackConfig, element.Lumberjack)
		// If output paths were provided by user, we will override it which means <stdout> would be omitted
		if len(element.OutputPaths) > 0 {
			eventLoggerConfig.OutputPaths = element.OutputPaths
		}

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
			for k, v := range element.Loki.Labels {
				opts = append(opts, rklogger.WithLokiLabel(k, v))
			}

			if element.Loki.InsecureSkipVerify {
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
			rkcommon.ShutdownWithError(err)
		} else {
			eventFactory = rkquery.NewEventFactory(
				rkquery.WithZapLogger(eventLogger),
				rkquery.WithAppName(GlobalAppCtx.GetAppInfoEntry().AppName),
				rkquery.WithAppVersion(GlobalAppCtx.GetAppInfoEntry().Version),
				rkquery.WithEncoding(rkquery.ToEncoding(element.Encoding)))
		}

		entry := RegisterEventLoggerEntry(
			WithNameEvent(element.Name),
			WithDescriptionEvent(element.Description),
			WithEventFactoryEvent(eventFactory),
			WithLokiSyncerEvent(lokiSyncer),
			WithBaseLoggerEvent(eventLogger))

		// special case for logger config
		entry.LoggerConfig = eventLoggerConfig
		entry.LumberjackConfig = eventLoggerLumberjackConfig

		res[element.Name] = entry
	}

	return res
}

// RegisterEventLoggerEntry create event logger entry with options.
func RegisterEventLoggerEntry(opts ...EventLoggerEntryOption) *EventLoggerEntry {
	entry := &EventLoggerEntry{
		EntryType:        EventLoggerEntryType,
		EntryDescription: EventLoggerDescription,
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "eventLogger-" + rkcommon.RandString(4)
	}

	if entry.EventFactory == nil {
		entry.EventFactory = rkquery.NewEventFactory()
	}

	entry.EventHelper = rkquery.NewEventHelper(entry.EventFactory)

	GlobalAppCtx.AddEventLoggerEntry(entry)

	return entry
}

// Bootstrap entry.
func (entry *EventLoggerEntry) Bootstrap(ctx context.Context) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.Bootstrap(ctx)
	}
}

// Interrupt entry.
func (entry *EventLoggerEntry) Interrupt(ctx context.Context) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.Interrupt(ctx)
	}
}

// GetName returns name of entry.
func (entry *EventLoggerEntry) GetName() string {
	return entry.EntryName
}

// GetType returns type of entry.
func (entry *EventLoggerEntry) GetType() string {
	return entry.EntryType
}

// String convert entry into JSON style string.
func (entry *EventLoggerEntry) String() string {
	var bytes []byte
	var err error

	if bytes, err = json.Marshal(entry); err != nil {
		return "{}"
	}

	return string(bytes)
}

// MarshalJSON marshal entry.
func (entry *EventLoggerEntry) MarshalJSON() ([]byte, error) {
	loggerConfigWrap := rklogger.TransformToZapConfigWrap(entry.LoggerConfig)

	type innerEventLoggerEntry struct {
		EntryName        string                  `yaml:"entryName" json:"entryName"`
		EntryType        string                  `yaml:"entryType" json:"entryType"`
		EntryDescription string                  `yaml:"entryDescription" json:"entryDescription"`
		LoggerConfig     *rklogger.ZapConfigWrap `yaml:"zapConfig" json:"zapConfig"`
		LumberjackConfig *lumberjack.Logger      `yaml:"lumberjackConfig" json:"lumberjackConfig"`
	}

	return json.Marshal(&innerEventLoggerEntry{
		EntryName:        entry.EntryName,
		EntryType:        entry.EntryType,
		EntryDescription: entry.EntryDescription,
		LoggerConfig:     loggerConfigWrap,
		LumberjackConfig: entry.LumberjackConfig,
	})
}

// UnmarshalJSON not supported.
func (entry *EventLoggerEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetDescription return description of entry.
func (entry *EventLoggerEntry) GetDescription() string {
	return entry.EntryDescription
}

// GetEventFactory return event factory, refer to rkquery.EventFactory.
func (entry *EventLoggerEntry) GetEventFactory() *rkquery.EventFactory {
	return entry.EventFactory
}

// GetEventHelper return event helperm refer to rkquery.EventHelper.
func (entry *EventLoggerEntry) GetEventHelper() *rkquery.EventHelper {
	return entry.EventHelper
}

// GetLumberjackConfig return lumberjack config, refer to lumberjack.Logger.
func (entry *EventLoggerEntry) GetLumberjackConfig() *lumberjack.Logger {
	return entry.LumberjackConfig
}

// AddEntryLabelToLokiSyncer add entry name entry type into loki syncer
func (entry *EventLoggerEntry) AddEntryLabelToLokiSyncer(e Entry) {
	if entry.lokiSyncer != nil && e != nil {
		entry.lokiSyncer.AddLabel("entry_name", e.GetName())
		entry.lokiSyncer.AddLabel("entry_type", e.GetType())
	}
}

// AddLabelToLokiSyncer add key value pair as label into loki syncer
func (entry *EventLoggerEntry) AddLabelToLokiSyncer(k, v string) {
	if entry.lokiSyncer != nil {
		entry.lokiSyncer.AddLabel(k, v)
	}
}

// Sync underlying logger
func (entry *EventLoggerEntry) Sync() {
	if entry.baseLogger != nil {
		entry.baseLogger.Sync()
	}
}
