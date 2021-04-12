# rk-entry
The entry library mainly used by rk-boot.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Entry](#entry)
    - [Create a new entry](#create-a-new-entry)
    - [Interact with rk-boot.Bootstrapper?](#interact-with-rk-bootbootstrapper)
  - [GlobalAppCtx](#globalappctx)
    - [Access GlobalAppCtx](#access-globalappctx)
    - [Usage of GlobalAppCtx](#usage-of-globalappctx)
  - [Built in entries](#built-in-entries)
    - [AppInfoEntry](#appinfoentry)
      - [YAML Hierarchy](#yaml-hierarchy)
      - [Access AppInfoEntry](#access-appinfoentry)
      - [Stringfy AppInfoEntry](#stringfy-appinfoentry)
    - [EventLoggerEntry](#eventloggerentry)
      - [YAML Hierarchy](#yaml-hierarchy-1)
      - [Access EventLoggerEntry](#access-eventloggerentry)
      - [Stringfy EventLoggerEntry](#stringfy-eventloggerentry)
    - [ZapLoggerEntry](#zaploggerentry)
      - [YAML Hierarchy](#yaml-hierarchy-2)
      - [Access ZapLoggerEntry](#access-zaploggerentry)
      - [Stringfy EventLoggerEntry](#stringfy-eventloggerentry-1)
    - [ViperEntry](#viperentry)
      - [YAML Hierarchy](#yaml-hierarchy-3)
      - [Access ViperEntry](#access-viperentry)
      - [Stringfy ViperEntry](#stringfy-viperentry)
    - [CertEntry](#certentry)
      - [YAML Hierarchy](#yaml-hierarchy-4)
      - [Access CertEntry](#access-certentry)
      - [Stringfy ViperEntry](#stringfy-viperentry-1)
      - [Select config file dynamically](#select-config-file-dynamically)
  - [Info Utility](#info-utility)
    - [ProcessInfo](#processinfo)
      - [Fields](#fields)
      - [Access ProcessInfo](#access-processinfo)
    - [ViperConfigInfo](#viperconfiginfo)
      - [Fields](#fields-1)
      - [Access ViperConfigInfo](#access-viperconfiginfo)
    - [MemStatsInfo](#memstatsinfo)
      - [Fields](#fields-2)
      - [Access MemStatsInfo](#access-memstatsinfo)
    - [PromMetricsInfo](#prommetricsinfo)
      - [Fields](#fields-3)
      - [Access PromMetricsInfo](#access-prommetricsinfo)
- [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation
```bash
go get -u github.com/rookie-ninja/rk-entry
```

## Quick Start
### Entry
**rkentry.Entry** is an interface for rkboot.Bootstrapper to bootstrap entry.

Users can implement **rkentry.Entry** interface and bootstrap any service/process with **rkboot.Bootstrapper**

#### Create a new entry
Please see examples in example/ directory for details.

- **Step 1:**
Construct your own YAML as needed. The elements in YAML file could be used while bootstrapping your entry.
```yaml
---
myEntry:
  enabled: true
  name: my-entry
  key: value
```

- **Step 2:**
Create a struct which implements **rkentry.Entry** interface.
```go
type MyEntry struct {
	name             string
	key              string
}

func (entry *MyEntry) Bootstrap(context.Context) {}

func (entry *MyEntry) Interrupt(context.Context) {}

func (entry *MyEntry) GetName() string {
	return entry.name
}

func (entry *MyEntry) GetType() string {
	return ""
}

func (entry *MyEntry) String() string {
	return ""
}
```

- **Step 3:**
Implements **rkentry.EntryRegFunc** and define a struct which could be marshaled from YAML config file.
```go
// A struct which is for unmarshalled YAML.
type BootConfig struct {
	MyEntry struct {
		Enabled   bool   `yaml:"enabled"`
		Name      string `yaml:"name"`
		Key       string `yaml:"key"`
	} `yaml:"myEntry"`
}

// An implementation of <type EntryRegFunc func(string) map[string]rkentry.Entry>
func RegisterMyEntriesFromConfig(configFilePath string) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: decode config map into boot config struct
	config := &BootConfig{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: construct entry
	if config.MyEntry.Enabled {
		entry := RegisterMyEntry(
			WithName(config.MyEntry.Name),
			WithKey(config.MyEntry.Key))
		res[entry.GetName()] = entry
	}

	return res
}

func RegisterMyEntry(opts ...MyEntryOption) *MyEntry {
	entry := &MyEntry{
		name:             "my-default",
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.name) < 1 {
		entry.name = "my-default"
	}
    
    // add your entry into GlobalAppCtx, user can access this entry via name.
    // eg: rkentry.GlobalAppCtx.GetEntry("myEntry")
	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

type MyEntryOption func(*MyEntry)

func WithName(name string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.name = name
	}
}

func WithKey(key string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.key = key
	}
}
```

- **Step 4:**
Register your reg function in init() in order to register your entry while application/process starts.
```go
// Register entry, must be in init() function since we need to register entry
// at beginning
func init() {
    rkentry.RegisterEntryRegFunc(RegisterMyEntriesFromConfig)
}
```

#### Interact with rk-boot.Bootstrapper?

1: Entry will be created and registered into rkentry.GlobalAppCtx.

2: rkboot.Bootstrap() function will iterator all entries in rkentry.GlobalAppCtx.Entries and call Bootstrap().

3: Application will wait for shutdown signal via rkentry.GlobalAppCtx.ShutdownSig.

4: rkboot.Interrupt() function will iterate all entries in rkentry.GlobalAppCtx.Entries and call Interrupt().

### GlobalAppCtx
A struct called AppContext witch contains RK style application metadata.

#### Access GlobalAppCtx

Access it via GlobalAppCtx variable 
```go
rkentry.GlobalAppCtx
```
**Fields in rkentry.GlobalAppCtx**

| Element | Description | JSON | Default values |
| ------ | ------ | ------ | ------ |
| StartTime | Application start time. | start_time | 0001-01-01 00:00:00 +0000 UTC |
| BasicEntries | Basic entries contains default zap logger entry, event logger entry and rk entry by default. | basic_entries | three default entries including AppInfoEntry, ZapLoggerEntry and EventLoggerEntry |
| Entries | User entries registered from user code. | entries | empty map |
| ViperConfigs | Viper configs with a name as a key. | viper_configs | empty map |
| ShutdownSig | Shutdown signals which includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT. | shutdown_sig | channel includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT |
| ShutdownHooks | Shutdown hooks registered from user code. | shutdown_hooks | empty list |
| UserValues | User K/V registered from code. | user_values | empty map |

#### Usage of GlobalAppCtx
- Access start_time and application up time.
```go
// It is not recommended to override this value since StartTime would be assigned to current time
// at beginning of go process in init() function.
startTime := rkentry.GlobalAppCtx.StartTime

// Get up time of application/process.
upTime := rkentry.GlobalAppCtx.GetUpTime()
```

- Access basic_entries
Basic entries includes default AppInfoEntry if values not overridden by user via config file or codes.
```go
// Access entries from GlobalAppCtx directly
basicEntries := rkentry.GlobalAppCtx.BasicEntries
```

- Access entries

Entries contains user defined or non basic entries.
```go
// Access entries from GlobalAppCtx directly
entries := rkentry.GlobalAppCtx.Entries

// Access entries via utility function
entries := rkentry.GlobalAppCtx.ListEntries()

// Add entry
rkentry.GlobalAppCtx.AddEntry()

// Get entry with name
entry := rkentry.GlobalAppCtx.GetEntry("name of your entry")

// Merge map of entries into rkentry.GlobalAppCtx.entries
rkentry.GlobalAppCtx.MergeEntries(mapOfYourEntries map[string]rkentry.Entry)
```

- Access viper_configs
```go
// Return map of viper instances
viperConfigs := rkentry.GlobalAppCtx.ViperConfigs
viperConfigs = rkentry.GlobalAppCtx.ListViperEntries()

// Retrieve certain viper config instance with name
viperConfig := rkentry.GlobalAppCtx.GetViperEntry("you viper config entry name")

// Add viper entry, just call RegisterViperEntry function
name := "your viper entry name"
vp := viper.New()

entry := RegisterViperEntry(
	WithNameViper(name),
	WithViperInstanceViper(vp))

// Or add viper entry via registration function with config file
rkentry.RegisterEventLoggerEntriesWithConfig("your config file path")
```

- Access user_values

User can add/get/list/remove any values into map of rkentry.GlobalAppCtx.UserValues as needed.

GlobalAppCtx don't provide any locking mechanism.
```go
// Add k/v value into GlobalAppCtx, key should be string and value could be any kind
rkentry.GlobalAppCtx.AddValue()

// Get value with key
value := rkentry.GlobalAppCtx.GetValue()

// Access entries from GlobalAppCtx directly
entries := rkentry.GlobalAppCtx.UserValues

// Access entries via utility function
entries := rkentry.GlobalAppCtx.ListValues()

// Remove value with key
rkentry.GlobalAppCtx.RemoveValue()
```

- Access shutdown_sig
```go
// Access shutdown signal directly
rkentry.GlobalAppCtx.ShutdownSig

// Wait for shutdown signal via utility function, otherwise, user must call by himself
rkentry.GlobalAppCtx.WaitForShutdownSig()
```

- Access shutdown_hooks

Users can add their own shutdown hook function into GlobalAppCtx.

rkboot will iterate all shutdown hooks in GlobalAppCtx and call every shutdown hook function.
```go
// Access shutdown hooks directly
rkentry.GlobalAppCtx.ShutdownHooks

// Access shutdown hooks via utility function
rkentry.GlobalAppCtx.ListSHutdownHooks()

// Add shutdown hook function with name
rkentry.GlobalAppCtx.AddShutdownHook()

// Get shutdown with name
rkentry.GlobalAppCtx.GetShutdownHook("name of shutdown hook function")
```

- Access EventLoggerEntry

Please refer to EventLoggerEntry section in README for details of EventLoggerEntry.
```go
// Add event logger via registration function with config file
rkentry.RegisterEventLoggerEntriesWithConfig("your config file path")
// Or add event logger from code
rkentry.RegisterEventLoggerEntry()

// Get event logger entry with name
rkentry.GlobalAppCtx.GetEventLoggerEntry("name of event logger entry")

// Remove event logger entry with name
rkentry.RemoveEventLoggerEntry("your entry name")

// List event logger entries
rkentry.GlobalAppCtx.ListEventLoggerEntries()

// Get event factory, utility function would get event logger entry first and retrieve event factory in it.
rkentry.GlobalAppCtx.GetEventFactory("name of event logger entry")

// Get event helper, utility function would get event logger entry first and retrieve event helper in it.
rkentry.GlobalAppCtx.GetEventHelper("name of event logger entry")

// Get default event logger entry, default entry would log every thing to stdout
rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
```

- Access ZapLoggerEntry

Please refer to ZapLoggerEntry section in README for details of ZapLoggerEntry. 
```go
// Add zap logger via registration function with config file
rkentry.RegisterZapLoggerEntriesWithConfig("your config file path")
// Or add zap logger from code
rkentry.RegisterZapLoggerEntry()

// Get zap logger entry with name
rkentry.GlobalAppCtx.GetZapLoggerEntry("name of zap logger entry")

// Remove zap logger entry with name
rkentry.RemoveZapLoggerEntry("your entry name")

// List zap logger entries
rkentry.GlobalAppCtx.ListZapLoggerEntries()

// Get zap logger, utility function would get zap logger entry first and retrieve zap logger instance in it.
rkentry.GlobalAppCtx.GetZapLogger("name of zap logger entry")

// Get zap logger config, utility function would get zap logger entry first and retrieve zap logger config in it.
rkentry.GlobalAppCtx.GetZapLoggerConfig("name of zap logger entry")

// Get default zap logger entry, default entry would log every thing to stdout
rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()

// Get default zap logger, default entry would log every thing to stdout
rkentry.GlobalAppCtx.GetZapLoggerDefault()

// Get default zap logger config, default entry would log every thing to stdout
rkentry.GlobalAppCtx.GetZapLoggerConfigDefault()
```

- Access AppInfoEntry
```go
// User don't need to add app info into GlobalAppCtx and it is strongly not recommended.

// Get app info entry
rkentry.GlobalAppCtx.GetAppInfoEntry()
```

- Access up time of application/process
```go
// Get up time of application/process which is calculated based on start_time
rkentry.GlobalAppCtx.GetUpTime()
```

### Built in entries
#### AppInfoEntry
AppInfoEntry contains bellow fields which could be overridden via YAML file or code.

| Name | Description | YAML | Default value |
| ------ | ------ | ------ | ------ |
| AppName | Application name which refers to go process | appName | rkapp |
| Version | Application version | version | v0.0.0 |
| Lang | Programming language <NOT configurable!> | N/A | golang |
| Description | Description of application itself | description | "" |
| Keywords | A set of words describe application | keywords | [] |
| HomeURL | Home page URL | homeURL | "" |
| IconURL | Application Icon URL | iconURL | "" |
| DocsURL | A set of URLs of documentations of application | docsURL | [] |
| Maintainers | Maintainers of application | maintainers | [] |

##### YAML Hierarchy
```yaml
rk:
  appName: rk-example-entry           # Optional, default: "rkapp"
  version: v0.0.1                     # Optional, default: "v0.0.0"
  description: "this is description"  # Optional, default: ""
  keywords: ["rk", "golang"]          # Optional, default: []
  homeURL: "http://example.com"       # Optional, default: ""
  iconURL: "http://example.com"       # Optional, default: ""
  docsURL: ["http://example.com"]     # Optional, default: []
  maintainers: ["rk-dev"]             # Optional, default: []
```

##### Access AppInfoEntry
```go
// Access entry
rkentry.GlobalAppCtx.GetAppInfoEntry()

// Access fields in entry
rkentry.GlobalAppCtx.GetAppInfoEntry().AppName
rkentry.GlobalAppCtx.GetAppInfoEntry().Version
rkentry.GlobalAppCtx.GetAppInfoEntry().Lang
rkentry.GlobalAppCtx.GetAppInfoEntry().Description
rkentry.GlobalAppCtx.GetAppInfoEntry().Keywords
rkentry.GlobalAppCtx.GetAppInfoEntry().HomeURL
rkentry.GlobalAppCtx.GetAppInfoEntry().IconURL
rkentry.GlobalAppCtx.GetAppInfoEntry().DocsURL
rkentry.GlobalAppCtx.GetAppInfoEntry().Maintainers
```

##### Stringfy AppInfoEntry
Assuming we have application info YAML as bellow:

```yaml
---
rk: 
  appName: rk-example-entry           # Optional, default: "rkapp"
  version: v0.0.1                     # Optional, default: "v0.0.0"
  description: "this is description"  # Optional, default: ""
  keywords: ["rk", "golang"]          # Optional, default: []
  homeURL: "http://example.com"       # Optional, default: ""
  iconURL: "http://example.com"       # Optional, default: ""
  docsURL: ["http://example.com"]     # Optional, default: []
  maintainers: ["rk-dev"]             # Optional, default: []
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetAppInfoEntry().String())
```

Process information could be printed either.
```json
{
	"app_name": "rk-example-entry",
	"application_name": "rk-example-entry",
	"az": "unknown",
	"description": "this is description",
	"docs_url": ["http://example.com"],
	"domain": "unknown",
	"entry_name": "rk-app-info-entry",
	"entry_type": "rk-app-info-entry",
	"gid": "20",
	"home_url": "http://example.com",
	"icon_url": "http://example.com",
	"keywords": ["rk", "golang"],
	"lang": "golang",
	"maintainers": ["rk-dev"],
	"realm": "unknown",
	"region": "unknown",
	"start_time": "2021-04-04T07:33:09+08:00",
	"uid": "501",
	"up_time_sec": 0,
	"up_time_str": "4 milliseconds",
	"username": "rk-dev",
	"version": "v0.0.1"
}
```

#### EventLoggerEntry
EventLoggerEntry is used for [rkquery](https://github.com/rookie-ninja/rk-query) whose responsibility is logging event like RPC or periodical jobs.

| Name | Description |
| ------ | ------ |
| eventFactory | rkquery.EventFactory was initialized at the beginning |
| eventHelper | rkquery.EventHelper was initialized at the beginning |
| loggerConfig | zap.Config which was initialized at the beginning which is not accessible after initialization |
| lumberjackConfig | lumberjack.Logger which was initialized at the beginning |

##### YAML Hierarchy
EventLoggerEntry needs application name while creating event log. As a result, it is recommended to add AppInfoEntry while
initializing event logger entry. Otherwise, default application name would be assigned.

| Name | Description | Default |
| ------ | ------ | ------ |
| rk.appName | Application name which refers to go process | rkapp | 
| eventLogger.name | Name of event logger entry | N/A |
| eventLogger.format | Format of event logger, RK & JSON is supported. Please refer rkquery.RK & rkquery.JSON | RK | 
| eventLogger.outputPaths | Output paths of event logger, stdout would be the default one if not provided. If one of output path was provided, then stdout would be omitted. Output path could be relative or absolute paths either. | stdout |
| eventLogger.lumberjack.filename | Filename is the file to write logs to | It uses <processname>-lumberjack.log in os.TempDir() if empty. |
| eventLogger.lumberjack.maxsize | MaxSize is the maximum size in megabytes of the log file before it gets rotated. | 1024 |
| eventLogger.lumberjack.maxage | MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename. | 7 |
| eventLogger.lumberjack.maxbackups | axBackups is the maximum number of old log files to retain. | 3 |
| eventLogger.lumberjack.localtime | LocalTime determines if the time used for formatting the timestamps in backup files is the computer's local time. | true |
| eventLogger.lumberjack.compress | Compress determines if the rotated log files should be compressed using gzip. | true |

```yaml
---
rk:
  appName: rk-example-entry            # Optional, default: "rkapp"
eventLogger:
  - name: event-logger                 # Required
    format: RK                         # Optional, default: RK, options: RK and JSON
    outputPaths: ["stdout"]            # Optional
    lumberjack:                        # Optional
      filename: "rkapp-event.log"      # Optional, default: It uses <processname>-lumberjack.log in os.TempDir() if empty.
      maxsize: 1024                    # Optional, default: 1024 (MB)
      maxage: 7                        # Optional, default: 7 (days)
      maxbackups: 3                    # Optional, default: 3 (days)
      localtime: true                  # Optional, default: true
      compress: true                   # Optional, default: true
```

##### Access EventLoggerEntry
```go
// Access entry
rkentry.GlobalAppCtx.GetEventLoggerEntry("event-logger")

// Access event factory
rkentry.GlobalAppCtx.GetEventLoggerEntry("event-logger").GetEventFactory()

// Access event helper
rkentry.GlobalAppCtx.GetEventLoggerEntry("event-logger").GetEventHelper()

// Access lumberjack config
rkentry.GlobalAppCtx.GetEventLoggerEntry("event-logger").GetLumberjackConfig()
```

##### Stringfy EventLoggerEntry
Assuming we have event logger YAML as bellow:

```yaml
---
rk:
  appName: rk-example-entry            # Optional, default: "rkapp"
eventLogger:
  - name: event-logger                 # Required
    format: RK                         # Optional, default: RK, options: RK and JSON
    outputPaths: ["stdout"]            # Optional
    lumberjack:                        # Optional
      filename: "rkapp-event.log"      # Optional, default: It uses <processname>-lumberjack.log in os.TempDir() if empty.
      maxsize: 1024                    # Optional, default: 1024 (MB)
      maxage: 7                        # Optional, default: 7 (days)
      maxbackups: 3                    # Optional, default: 3 (days)
      localtime: true                  # Optional, default: true
      compress: true                   # Optional, default: true
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetEventLoggerEntry("event-logger").String())
```

Process information could be printed either.
```json
{
	"entry_name": "event-logger",
	"entry_type": "event-logger",
	"lumberjack_compress": true,
	"lumberjack_filename": "rkapp-event.log",
	"lumberjack_localtime": true,
	"lumberjack_maxage": 7,
	"lumberjack_maxbackups": 3,
	"lumberjack_maxsize": 1024,
	"output_path": ["stdout"]
}
```

#### ZapLoggerEntry
ZapLoggerEntry is used for initializing zap logger.

| Name | Description |
| ------ | ------ |
| logger | zap.Logger which was initialized at the beginning |
| loggerConfig | zap.Config which was initialized at the beginning |
| lumberjackConfig | lumberjack.Logger which was initialized at the beginning |

##### YAML Hierarchy
ZapLoggerEntry follows zap and lumberjack YAML hierarchy, please refer to [zap](https://pkg.go.dev/go.uber.org/zap#section-documentation) and [lumberjack](https://github.com/natefinch/lumberjack) site for details.

| Name | Description | Default |
| ------ | ------ | ------ |
| appLogger.name | Name of zap logger entry | N/A |
| eventLogger.format | Format of event logger, RK & JSON is supported. Please refer rkquery.RK & rkquery.JSON | RK | 
| eventLogger.outputPaths | Output paths of event logger, stdout would be the default one if not provided. If one of output path was provided, then stdout would be omitted. Output path could be relative or absolute paths either. | stdout |
| appLogger.lumberjack.filename | Filename is the file to write logs to | It uses <processname>-lumberjack.log in os.TempDir() if empty. |
| appLogger.lumberjack.maxsize | MaxSize is the maximum size in megabytes of the log file before it gets rotated. | 1024 |
| appLogger.lumberjack.maxage | MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename. | 7 |
| appLogger.lumberjack.maxbackups | axBackups is the maximum number of old log files to retain. | 3 |
| appLogger.lumberjack.localtime | LocalTime determines if the time used for formatting the timestamps in backup files is the computer's local time. | true |
| appLogger.lumberjack.compress | Compress determines if the rotated log files should be compressed using gzip. | true |

```yaml
---
zapLogger:
  - name: zap-logger                   # Required
    zap:                              
      level: info                      # Optional, default: info, options: [debug, DEBUG, info, INFO, warn, WARN, dpanic, DPANIC, panic, PANIC, fatal, FATAL]
      development: true                # Optional, default: true
      disableCaller: false             # Optional, default: false
      disableStacktrace: true          # Optional, default: true
      sampling:                        # Optional, default: empty map
        initial: 0
        thereafter: 0
      encoding: console                # Optional, default: "console", options: [console, json]
      encoderConfig:
        messageKey: "msg"              # Optional, default: "msg"
        levelKey: "level"              # Optional, default: "level"
        timeKey: "ts"                  # Optional, default: "ts"
        nameKey: "logger"              # Optional, default: "logger"
        callerKey: "caller"            # Optional, default: "caller"
        functionKey: ""                # Optional, default: ""
        stacktraceKey: "msg"           # Optional, default: "msg"
        lineEnding: "\n"               # Optional, default: "\n"
        levelEncoder: "capitalColor"   # Optional, default: "capitalColor", options: [capital, capitalColor, color, lowercase]
        timeEncoder: "iso8601"         # Optional, default: "iso8601", options: [rfc3339nano, RFC3339Nano, rfc3339, RFC3339, iso8601, ISO8601, millis, nanos]
        durationEncoder: "string"      # Optional, default: "string", options: [string, nanos, ms]
        callerEncoder: ""              # Optional, default: ""
        nameEncoder: ""                # Optional, default: ""
        consoleSeparator: ""           # Optional, default: ""
      outputPaths: [ "stdout" ]        # Optional, default: ["stdout"], stdout would be replaced if specified
      errorOutputPaths: [ "stderr" ]   # Optional, default: ["stderr"], stderr would be replaced if specified
      initialFields:                   # Optional, default: empty map
        key: "value"             
    lumberjack:                        # Optional
      filename: "rkapp-event.log"      # Optional, default: It uses <processname>-lumberjack.log in os.TempDir() if empty.
      maxsize: 1024                    # Optional, default: 1024 (MB)
      maxage: 7                        # Optional, default: 7 (days)
      maxbackups: 3                    # Optional, default: 3 (days)
      localtime: true                  # Optional, default: true
      compress: true                   # Optional, default: true
```

##### Access ZapLoggerEntry
```go
// Access entry
rkentry.GlobalAppCtx.GetZapLoggerEntry("zap-logger")

// Access zap logger
rkentry.GlobalAppCtx.GetZapLoggerEntry("zap-logger").GetLogger()

// Access zap logger config
rkentry.GlobalAppCtx.GetZapLoggerEntry("zap-logger").GetLoggerConfig()

// Access lumberjack config
rkentry.GlobalAppCtx.GetZapLoggerEntry("zap-logger").GetLumberjackConfig()
```

##### Stringfy EventLoggerEntry
Assuming we have event logger YAML as bellow:

```yaml
---
zapLogger:
  - name: zap-logger                   # Required
    zap:                              
      level: info                      # Optional, default: info, options: [debug, DEBUG, info, INFO, warn, WARN, dpanic, DPANIC, panic, PANIC, fatal, FATAL]
    lumberjack:                        # Optional
      filename: "rkapp-event.log"      # Optional, default: It uses <processname>-lumberjack.log in os.TempDir() if empty.
      maxsize: 1024                    # Optional, default: 1024 (MB)
      maxage: 7                        # Optional, default: 7 (days)
      maxbackups: 3                    # Optional, default: 3 (days)
      localtime: true                  # Optional, default: true
      compress: true                   # Optional, default: true
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetZapLoggerEntry("zap-logger").String())
```

Process information could be printed either.
```json
{
	"entry_name": "zap-logger",
	"entry_type": "zap-logger",
	"level": "info",
	"lumberjack_compress": true,
	"lumberjack_filename": "rkapp-event.log",
	"lumberjack_localtime": true,
	"lumberjack_maxage": 7,
	"lumberjack_maxbackups": 3,
	"lumberjack_maxsize": 1024,
	"output_path": ["stdout"]
}
```

#### ViperEntry
ViperEntry provides convenient way to initialize viper instance. [viper](https://github.com/spf13/viper) is a complete configuration solution for Go applications.
Each viper instance combined with one configuration file. 

| Name | Description |
| ------ | ------ |
| path | File path of config file, could be either relative or absolute path |
| vp | Viper instance |

##### YAML Hierarchy

| Name | Description | Default |
| ------ | ------ | ------ |
| viper.name | Name of viper entry | N/A |
| viper.path | File path of config file, could be either relative or absolute path | N/A | 

##### Access ViperEntry
```go
// Access entry
rkentry.GlobalAppCtx.GetViperEntry("my-viper"))

// Access viper instance
rkentry.GlobalAppCtx.GetViperEntry("my-viper").GetViper()
```

##### Stringfy ViperEntry
Assuming we have viper YAML as bellow:

```yaml
---
viper:
  - name: my-viper
    path: example/my-config.yaml
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetViperEntry("my-viper").String())
```

Process information could be printed either.
```json
{
	"entry_name": "my-viper",
	"entry_type": "viper-config",
	"path": "/xxxx/example/my-config.yaml"
}
```

#### CertEntry
CertEntry provides a convenient way to retrieve certifications from local or remote services.
Supported services listed bellow:
- local
- remote file store
- etcd
- consul

| Name | Description |
| ------ | ------ |
| Stores | Map of CertStore which contains server & client keys retrieved from service |
| Retrievers | Map of Retriever specified in YAML file |

##### YAML Hierarchy

| Name | Description | Default |
| ------ | ------ | ------ |
| cert.consul.name | Name of consul retriever | "" |
| cert.consul.locale | Represent environment of current process follows schema of \<realm\>::\<region\>::\<az\>::\<domain\> | \*::\*::\*::\* | 
| cert.consul.endpoint | Endpoint of Consul server, http://x.x.x.x or x.x.x.x both acceptable. | N/A |
| cert.consul.datacenter | Consul datacenter. | "" |
| cert.consul.token | Token for access Consul. | "" |
| cert.consul.basicAuth | Basic auth for Consul server, like <user:pass>. | "" |
| cert.consul.serverCertPath | Key of server cert in Consul server. | "" |
| cert.consul.serverKeyPath | Key of server key in Consul server. | "" |
| cert.consul.clientCertPath | Key of client cert in Consul server. | "" |
| cert.consul.clientCertPath | Key of client key in Consul server. | "" |
| cert.etcd.name | Name of etcd retriever | "" |
| cert.etcd.locale | Represent environment of current process follows schema of \<realm\>::\<region\>::\<az\>::\<domain\> | \*::\*::\*::\* | 
| cert.etcd.endpoint | Endpoint of ETCD server, http://x.x.x.x or x.x.x.x both acceptable. | N/A |
| cert.etcd.basicAuth | Basic auth for ETCD server, like <user:pass>. | "" |
| cert.etcd.serverCertPath | Key of server cert in ETCD server. | "" |
| cert.etcd.serverKeyPath | Key of server key in ETCD server. | "" |
| cert.etcd.clientCertPath | Key of client cert in ETCD server. | "" |
| cert.etcd.clientCertPath | Key of client key in ETCD server. | "" |
| cert.local.name | Name of local retriever | "" |
| cert.local.locale | Represent environment of current process follows schema of \<realm\>::\<region\>::\<az\>::\<domain\> | \*::\*::\*::\* | 
| cert.local.serverCertPath | Key of server cert in local file system. | "" |
| cert.local.serverKeyPath | Key of server key in local file system. | "" |
| cert.local.clientCertPath | Key of client cert in local file system. | "" |
| cert.local.clientCertPath | Key of client key in local file system. | "" |
| cert.remoteFileStore.name | Name of remoteFileStore retriever | "" |
| cert.remoteFileStore.locale | Represent environment of current process follows schema of \<realm\>::\<region\>::\<az\>::\<domain\> | \*::\*::\*::\* | 
| cert.remoteFileStore.endpoint | Endpoint of remoteFileStore server, http://x.x.x.x or x.x.x.x both acceptable. | N/A |
| cert.remoteFileStore.basicAuth | Basic auth for remoteFileStore server, like <user:pass>. | "" |
| cert.remoteFileStore.serverCertPath | Key of server cert in remoteFileStore server. | "" |
| cert.remoteFileStore.serverKeyPath | Key of server key in remoteFileStore server. | "" |
| cert.remoteFileStore.clientCertPath | Key of client cert in remoteFileStore server. | "" |
| cert.remoteFileStore.clientCertPath | Key of client key in remoteFileStore server. | "" |

##### Access CertEntry
```go
// Access entry
certEntry := rkentry.GlobalAppCtx.GetCertEntry()

// Access cert stores which contains certificates as byte array
serverCert := certEntry.Stores["your retriever name"].ServerCert
serverKey := certEntry.Stores["your retriever name"].ServerKey
clientCert := certEntry.Stores["your retriever name"].ClientCert
clientKey := certEntry.Stores["your retriever name"].ClientKey
```

##### Stringfy ViperEntry
Assuming we have viper YAML as bellow:

```yaml
---
cert:
  etcd:
    - name: "etcd-test"                     # Required
      locale: "*::*::*::*"                  # Optional, default: *::*::*::*
      endpoint: "localhost:2379"            # Required, http://x.x.x.x or x.x.x.x both acceptable.
      basicAuth: "root:etcd"                # Optional, default: "", basic auth for Consul server, like <user:pass>
      serverCertPath: "serverCert"          # Optional, default: "", key of value in etcd
      serverKeyPath: "serverKey"            # Optional, default: "", key of value in etcd
      clientCertPath: "clientCert"          # Optional, default: "", key of value in etcd
      clientKeyPath: "clientKey"            # Optional, default: "", key of value in etcd
  local:
    - name: "local-test"                       # Required
      locale: "*::*::*::*"                     # Optional, default: *::*::*::*
      serverCertPath: "example/server.pem"     # Optional, default: "", path of certificate on local FS
      serverKeyPath: "example/server-key.pem"  # Optional, default: "", path of certificate on local FS
      clientCertPath: "example/client.pem"     # Optional, default: "", path of certificate on local FS
      clientKeyPath: "example/client.pem"      # Optional, default: "", path of certificate on local FS
  consul:
    - name: "consul-test"                      # Required
      locale: "*::*::*::*"                     # Optional, default: *::*::*::*
      endpoint: "localhost:8500"               # Required, http://x.x.x.x or x.x.x.x both acceptable.
      basicAuth: "user:pass"                   # Optional, default: "", basic auth for consul server, like <user:pass>
      datacenter: "rk"                         # Optional, default: "", consul datacenter
      token: ""                                # Optional, default: "", token to access consul
      serverCertPath: "serverCert"             # Optional, default: "", key of value in consul 
      serverKeyPath: "serverKey"               # Optional, default: "", key of value in consul
      clientCertPath: "clientCert"             # Optional, default: "", key of value in consul
      clientKeyPath: "clientKey"               # Optional, default: "", key of value in consul
  remoteFileStore:
    - name: "remote-file-store-test"           # Required
      locale: "*::*::*::*"                     # Optional, default: *::*::*::*
      endpoint: "localhost:8080"               # Required, http://x.x.x.x or x.x.x.x both acceptable.
      basicAuth: "user:pass"                   # Optional, default: "", basic auth for remote file store, like <user:pass>
      serverCertPath: "serverCert"             # Optional, default: "", path of file in remote file store
      serverKeyPath: "serverKey"               # Optional, default: "", path of file in remote file store
      clientCertPath: "clientCert"             # Optional, default: "", path of file in remote file store
      clientKeyPath: "clientKey"               # Optional, default: "", path of file in remote file store
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetCertEntry().String())
```

Process information could be printed either.
```json
{
	"entry_name": "rk-cert-entry",
	"entry_type": "rk-cert-entry",
	"local-test_client_cert_exist": false,
	"local-test_client_key_exist": false,
	"local-test_server_cert_exist": true,
	"local-test_server_key_exist": true,
	"retrievers": ["local-test"]
}
```

##### Select config file dynamically
In order to select config file dynamically in different environment, we are using environment variable to choose config files.

**How it works?**
ViperEntry will reconstruct path user provided with DOMAIN(environment variable) as bellow:

Path that user provided: example/my-config.yaml
Value of DOMAIN: prod
Reconstructed path: example/my-config-prod.yaml

As a result, there could be two files named as my-config-test.yaml and my-config-prod.yaml in one path, 
and specify viper.path as my-config.yaml in ViperEntry YAML file.

**Example:**
Assuming we have test and prod environment each needs own configuration file with same index in it.
We need to make sure access correct configuration file without any code changes.

- Step 1: 
Two configuration file described as bellow.

my-config-test.yaml
```yaml
key: test
```

my-config-prod.yaml
```yaml
key: prod
```

- Step 2:
Configure path at bootstrap yaml file.

my-boot.yaml
```yaml
viper:
  - name: my-config
    path: my-config.yaml  # ViperEntry will reconstruct this as my-config-<DOOMAIN>.yaml, if value of DOAMIN is empty, then look up my-config.yaml.
```

- Step 3:
Set environment variable as DOMAIN=prod or DOMAIN=test

- Step 4: 
Access values in configuration file with name of viper entry

```go
rkentry.GlobalAppCtx.GetViperEntry("my-config").GetViper().GetString("key")
```

### Info Utility
#### ProcessInfo
Process information for a running application.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| ApplicationName | Name of current application set by user | based on system | application_name |
| UID | user id which runs process | based on system | uid |
| GID | group id which runs process | based on system | gid |
| Username | username which runs process | based on system | username |
| StartTime | application start time | time.now() | start_time |
| UpTimeSec | application up time in seconds | zero at the beginning | up_time_sec |
| UpTimeStr | application up time in string | zero as string at the beginning | up_time_str |
| Region | region where process runs | based on environment variable REGION | region |
| AZ | availability zone where process runs | based on environment variable AZ | az |
| Realm | realm where process runs | based on environment variable REALM | realm |
| Domain | domain where process runs | based on environment variable DOMAIN | domain |

##### Access ProcessInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewProcessInfo()))
```

```json
{
  "application_name": "rk-example-entry",
  "uid": "501",
  "gid": "20",
  "username": "rk-dev",
  "start_time": "2021-04-04T17:31:45+08:00",
  "up_time_sec": 0,
  "up_time_str": "5 milliseconds",
  "region": "unknown",
  "az": "unknown",
  "realm": "unknown",
  "domain": "prod"
}
```

#### ViperConfigInfo
Viper config information stored in GlobalAppCtx.
##### Fields
| Element | Description | Default | JSON key |
| ------ | ------ | ------ | ------ |
| Name | Name of config instance | N/A | name |
| Raw | Name of config instance | N/A | raw |

##### Access ViperConfigInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewViperConfigInfo()))
```

```json
[
  {
    "name": "my-viper",
    "raw": "map[key:prod]"
  }
]
```

#### MemStatsInfo
Memory stats of current running process.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| MemAllocByte | Bytes of allocated heap objects, from runtime.MemStats.Alloc | based on system | mem_alloc_byte |
| SysAllocByte | Total bytes of memory obtained from the OS, from runtime.MemStats.Sys | based on system | sys_alloc_byte |
| MemPercentage | float64(stats.Alloc) / float64(stats.Sys) | based on system | mem_usage_percentage |
| LastGCTimestamp | The time the last garbage collection finished as RFC3339 | based on system | last_gc_timestamp |
| GCCount | The number of completed GC cycles | zero at the beginning | gc_count_total |
| ForceGCCount | The number of GC cycles that were forced by the application calling the GC function | zero as string at the beginning | force_gc_count |

##### Access MemStatsInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewMemStatsInfo()))
```

```json
{
  "mem_alloc_byte": 1123856,
  "sys_alloc_byte": 73614592,
  "mem_usage_percentage": 0.015266755808413636,
  "last_gc_timestamp": "1970-01-01T08:00:00+08:00",
  "gc_count_total": 0,
  "force_gc_count": 0
}
```

#### PromMetricsInfo
Request metrics to struct from prometheus summary collector.
##### Fields
| Element | Description | Default | JSON key |
| ------ | ------ | ------ | ------ |
| Path | API path | Based on system | path |
| ElapsedNanoP50 | Quantile of p50 with time elapsed | based on prometheus collector | elapsed_nano_p50 |
| ElapsedNanoP90 | Quantile of p90 with time elapsed | based on prometheus collector | elapsed_nano_p90 |
| ElapsedNanoP99 | Quantile of p99 with time elapsed | based on prometheus collector | elapsed_nano_p99 |
| ElapsedNanoP999 | Quantile of p999 with time elapsed | based on prometheus collector | elapsed_nano_p99 |
| Count | Total number of requests | based on prometheus collector | count |
| ResCode | Response code labels | based on prometheus collector | res_code |

##### Access PromMetricsInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewPromMetricsInfo(<your prometheus summary vector>)))
```

**How to get request metrics?** 
Access it via MemStatsToStruct() function 
```go
rk_metrics.GetRequestMetrics()
```

## Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
dongxuny@gmail.com. That email list is a private, safe space; even the zap
maintainers don't have access, so don't hesitate to hold us to a high
standard.

Released under the [MIT License](LICENSE).