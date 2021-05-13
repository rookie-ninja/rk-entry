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
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

const (
	EventLoggerEntryType   = "EventLoggerEntry"
	EventLoggerNameNoop    = "EventLoggerNoop"
	EventLoggerDescription = "Internal RK entry which is used to log event such as RPC request or periodic jobs."
)

// Create event logger entry with noop event factory.
// Event factory and event helper will be created with noop zap logger.
// Since we don't need any log rotation in case of noop, lumberjack config will be nil.
func NoopEventLoggerEntry() *EventLoggerEntry {
	entry := &EventLoggerEntry{
		EntryName:    EventLoggerNameNoop,
		EntryType:    EventLoggerEntryType,
		EventFactory: rkquery.NewEventFactory(rkquery.WithLogger(rklogger.NoopLogger)),
	}

	entry.EventHelper = rkquery.NewEventHelper(entry.EventFactory)

	return entry
}

// Bootstrap config of Event Logger information.
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
	RK struct {
		AppName string `yaml:"appName" json:"appName"`
		Version string `yaml:"version" json:"version"`
	} `yaml:"rk" json:"rk"`
	EventLogger []struct {
		Name        string             `yaml:"name" json:"name"`
		Description string             `yaml:"description" json:"description"`
		Format      string             `yaml:"format" json:"format"`
		OutputPaths []string           `yaml:"outputPaths" json:"outputPaths"`
		Lumberjack  *lumberjack.Logger `yaml:"lumberjack" json:"lumberjack"`
	} `yaml:"eventLogger json:"eventLogger"`
}

// EventLoggerEntry contains bellow fields.
// 1: EntryName: Name of entry.
// 2: EntryType: Type of entry which is EventLoggerEntryType.
// 3: EntryDescription: Description of EventLoggerEntry.
// 4: EventFactory: rkquery.EventFactory was initialized at the beginning.
// 5: EventHelper: rkquery.EventHelper was initialized at the beginning.
// 6: LoggerConfig: zap.Config which was initialized at the beginning which is not accessible after initialization.
// 7: LumberjackConfig: lumberjack.Logger which was initialized at the beginning.
type EventLoggerEntry struct {
	EntryName        string                `yaml:"entryName" json:"entryName"`
	EntryType        string                `yaml:"entryType" json:"entryType"`
	EntryDescription string                `yaml:"entryDescription" json:"entryDescription"`
	EventFactory     *rkquery.EventFactory `yaml:"-" json:"-"`
	EventHelper      *rkquery.EventHelper  `yaml:"-" json:"-"`
	LoggerConfig     *zap.Config           `yaml:"zapConfig" json:"zapConfig"`
	LumberjackConfig *lumberjack.Logger    `yaml:"lumberjackConfig" json:"lumberjackConfig"`
}

// EventLoggerEntry Option which used while registering entry from codes.
type EventLoggerEntryOption func(*EventLoggerEntry)

// Provide name of entry.
func WithNameEvent(name string) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.EntryName = name
	}
}

// Provide description of entry.
func WithDescriptionEvent(description string) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.EntryDescription = description
	}
}

// Provide event factory of entry which refers to rkquery.EventFactory.
func WithEventFactoryEvent(fac *rkquery.EventFactory) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.EventFactory = fac
	}
}

// Create event logger entries with config file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterEventLoggerEntriesWithConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: unmarshal user provided config into boot config struct
	config := &BootConfigEventLogger{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// deal with application name specifically in case of empty string
	if len(config.RK.AppName) < 1 {
		config.RK.AppName = AppNameDefault
	}

	if len(config.RK.Version) < 1 {
		config.RK.Version = "unknown"
	}

	// 2: init event logger entries with boot config
	for i := range config.EventLogger {
		element := config.EventLogger[i]

		var eventFactory *rkquery.EventFactory
		// assign default zap config and lumberjack config
		eventLoggerConfig := rklogger.NewZapEventConfig()
		eventLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()
		// override with user provided zap config and lumberjack config
		rkcommon.OverrideLumberjackConfig(eventLoggerLumberjackConfig, element.Lumberjack)
		// if output paths were provided by user, we will override it which means <stdout> would be omitted
		if len(element.OutputPaths) > 0 {
			eventLoggerConfig.OutputPaths = element.OutputPaths
		}

		if eventLogger, err := rklogger.NewZapLoggerWithConf(eventLoggerConfig, eventLoggerLumberjackConfig); err != nil {
			rkcommon.ShutdownWithError(err)
		} else {
			elements := []string{
				rkcommon.GetDefaultIfEmptyString(os.Getenv("REALM"), "unknown"),
				rkcommon.GetDefaultIfEmptyString(os.Getenv("REGION"), "unknown"),
				rkcommon.GetDefaultIfEmptyString(os.Getenv("AZ"), "unknown"),
				rkcommon.GetDefaultIfEmptyString(os.Getenv("DOMAIN"), "unknown"),
			}

			locale := strings.Join(elements, "::")

			eventFactory = rkquery.NewEventFactory(
				rkquery.WithLogger(eventLogger),
				rkquery.WithAppName(config.RK.AppName),
				rkquery.WithAppVersion(config.RK.Version),
				rkquery.WithLocale(locale),
				rkquery.WithFormat(rkquery.ToFormat(element.Format)))
		}

		entry := RegisterEventLoggerEntry(
			WithNameEvent(element.Name),
			WithDescriptionEvent(element.Description),
			WithEventFactoryEvent(eventFactory))

		// special case for logger config
		entry.LoggerConfig = eventLoggerConfig
		entry.LumberjackConfig = eventLoggerLumberjackConfig

		res[element.Name] = entry
	}

	return res
}

// Create event logger entry with options.
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
func (entry *EventLoggerEntry) Bootstrap(context.Context) {
	// no op
}

// Interrupt entry.
func (entry *EventLoggerEntry) Interrupt(context.Context) {
	// no op
}

// Get name of entry.
func (entry *EventLoggerEntry) GetName() string {
	return entry.EntryName
}

// Get type of entry.
func (entry *EventLoggerEntry) GetType() string {
	return entry.EntryType
}

// Convert entry into JSON style string.
func (entry *EventLoggerEntry) String() string {
	if bytes, err := json.Marshal(entry); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

// Marshal entry.
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

// Not supported
func (entry *EventLoggerEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Return description of entry
func (entry *EventLoggerEntry) GetDescription() string {
	return entry.EntryDescription
}

// Get event factory, refer to rkquery.EventFactory.
func (entry *EventLoggerEntry) GetEventFactory() *rkquery.EventFactory {
	return entry.EventFactory
}

// Get event helperm refer to rkquery.EventHelper.
func (entry *EventLoggerEntry) GetEventHelper() *rkquery.EventHelper {
	return entry.EventHelper
}

// Get lumberjack config, refer to lumberjack.Logger.
func (entry *EventLoggerEntry) GetLumberjackConfig() *lumberjack.Logger {
	return entry.LumberjackConfig
}
