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
	EventLoggerEntryType = "event-logger"
)

// Create event logger entry with noop event factory.
// Event factory and event helper will be created with noop zap logger.
// Since we don't need any log rotation in case of noop, lumberjack config will be nil.
func NoopEventLoggerEntry() *EventLoggerEntry {
	entry := &EventLoggerEntry{
		entryName:    "rk-event-logger-noop",
		entryType:    EventLoggerEntryType,
		eventFactory: rkquery.NewEventFactory(rkquery.WithLogger(rklogger.NoopLogger)),
	}

	entry.eventHelper = rkquery.NewEventHelper(entry.eventFactory)

	return entry
}

// Bootstrap config of Event Logger information.
// 1: RK.AppName: Application name which refers to go process.
//                Default application name of AppNameDefault would be assigned if missing in boot config file.
// 2: EventLogger.Name: Name of event logger entry.
// 3: EventLogger.Format: Format of event logger, RK & JSON is supported. Please refer rkquery.RK & rkquery.JSON.
// 4: EventLogger.OutputPaths: Output paths of event logger, stdout would be the default one if not provided.
//                             If one of output path was provided, then stdout would be omitted.
//                             Output path could be relative or absolute paths either.
// 5: EventLogger.Lumberjack: Lumberjack config which follows lumberjack.Logger style.
type BootConfigEventLogger struct {
	RK struct {
		AppName string `yaml:"appName"`
		Version string `yaml:"version"`
	} `yaml:"rk"`
	EventLogger []struct {
		Name        string             `yaml:"name"`
		Format      string             `yaml:"format"`
		OutputPaths []string           `yaml:"outputPaths"`
		Lumberjack  *lumberjack.Logger `yaml:"lumberjack"`
	} `yaml:"eventLogger`
}

// EventLoggerEntry contains bellow fields.
// 1: entryName: Name of entry.
// 2: entryType: Type of entry which is EventLoggerEntryType.
// 3: eventFactory: rkquery.EventFactory was initialized at the beginning.
// 4: eventHelper: rkquery.EventHelper was initialized at the beginning.
// 5: loggerConfig: zap.Config which was initialized at the beginning which is not accessible after initialization.
// 6: lumberjackConfig: lumberjack.Logger which was initialized at the beginning.
type EventLoggerEntry struct {
	entryName        string
	entryType        string
	eventFactory     *rkquery.EventFactory
	eventHelper      *rkquery.EventHelper
	loggerConfig     *zap.Config
	lumberjackConfig *lumberjack.Logger
}

// EventLoggerEntry Option which used while registering entry from codes.
type EventLoggerEntryOption func(*EventLoggerEntry)

// Provide name of entry.
func WithNameEvent(name string) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.entryName = name
	}
}

// Provide event factory of entry which refers to rkquery.EventFactory.
func WithEventFactoryEvent(fac *rkquery.EventFactory) EventLoggerEntryOption {
	return func(entry *EventLoggerEntry) {
		entry.eventFactory = fac
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
		queryLoggerConfig := rklogger.NewZapEventConfig()
		queryLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()
		// override with user provided zap config and lumberjack config
		rkcommon.OverrideLumberjackConfig(queryLoggerLumberjackConfig, element.Lumberjack)
		// if output paths were provided by user, we will override it which means <stdout> would be omitted
		if len(element.OutputPaths) > 0 {
			queryLoggerConfig.OutputPaths = element.OutputPaths
		}

		if queryLogger, err := rklogger.NewZapLoggerWithConf(queryLoggerConfig, queryLoggerLumberjackConfig); err != nil {
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
				rkquery.WithLogger(queryLogger),
				rkquery.WithAppName(config.RK.AppName),
				rkquery.WithAppVersion(config.RK.Version),
				rkquery.WithLocale(locale),
				rkquery.WithFormat(rkquery.ToFormat(element.Format)))
		}

		entry := RegisterEventLoggerEntry(
			WithNameEvent(element.Name),
			WithEventFactoryEvent(eventFactory))

		// special case for logger config
		entry.loggerConfig = queryLoggerConfig
		entry.lumberjackConfig = queryLoggerLumberjackConfig

		res[element.Name] = entry
	}

	return res
}

// Create event logger entry with options.
func RegisterEventLoggerEntry(opts ...EventLoggerEntryOption) *EventLoggerEntry {
	entry := &EventLoggerEntry{
		entryType: EventLoggerEntryType,
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "event-logger-" + rkcommon.RandString(4)
	}

	if entry.eventFactory == nil {
		entry.eventFactory = rkquery.NewEventFactory()
	}

	entry.eventHelper = rkquery.NewEventHelper(entry.eventFactory)

	GlobalAppCtx.addEventLoggerEntry(entry)

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
	return entry.entryName
}

// Get type of entry.
func (entry *EventLoggerEntry) GetType() string {
	return entry.entryType
}

// Convert entry into JSON style string.
func (entry *EventLoggerEntry) String() string {
	m := map[string]interface{}{
		"entry_name": entry.entryName,
		"entry_type": entry.entryType,
	}

	if entry.loggerConfig != nil {
		m["output_path"] = entry.loggerConfig.OutputPaths
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

// Get event factory, refer to rkquery.EventFactory.
func (entry *EventLoggerEntry) GetEventFactory() *rkquery.EventFactory {
	return entry.eventFactory
}

// Get event helperm refer to rkquery.EventHelper.
func (entry *EventLoggerEntry) GetEventHelper() *rkquery.EventHelper {
	return entry.eventHelper
}

// Get lumberjack config, refer to lumberjack.Logger.
func (entry *EventLoggerEntry) GetLumberjackConfig() *lumberjack.Logger {
	return entry.lumberjackConfig
}
