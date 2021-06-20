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
  - [Built in internal entries](#built-in-internal-entries)
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
    - [ConfigEntry](#configentry)
      - [YAML Hierarchy](#yaml-hierarchy-3)
      - [Access ConfigEntry](#access-configentry)
      - [Stringfy ConfigEntry](#stringfy-configentry)
      - [Dynamically load config based on Environment variable](#dynamically-load-config-based-on-environment-variable)
    - [CertEntry](#certentry)
      - [YAML Hierarchy](#yaml-hierarchy-4)
      - [Access CertEntry](#access-certentry)
      - [Stringfy CertEntry](#stringfy-certentry)
      - [Select cert entry dynamically based on environment](#select-cert-entry-dynamically-based-on-environment)
  - [Info Utility](#info-utility)
    - [ProcessInfo](#processinfo)
      - [Fields](#fields)
      - [Access ProcessInfo](#access-processinfo)
    - [CpuInfo](#cpuinfo)
      - [Fields](#fields-1)
      - [Access CpuInfo](#access-cpuinfo)
    - [MemInfo](#meminfo)
      - [Fields](#fields-2)
      - [Access MemInfo](#access-meminfo)
    - [NetInfo](#netinfo)
      - [Fields](#fields-3)
      - [Access NetInfo](#access-netinfo)
    - [OsInfo](#osinfo)
      - [Fields](#fields-4)
      - [Access OsInfo](#access-osinfo)
    - [GoEnvInfo](#goenvinfo)
      - [Fields](#fields-5)
      - [Access OsInfo](#access-osinfo-1)
    - [PromMetricsInfo](#prommetricsinfo)
      - [Fields](#fields-6)
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

func (entry *MyEntry) GetDescription() string {
	return ""
}
```

- **Step 3:**
Implements **rkentry.EntryRegFunc** and define a struct which could be marshaled from YAML config file.
```go
// A struct which is for unmarshalled YAML.
type BootConfig struct {
	MyEntry struct {
		Enabled   bool   `yaml:"enabled" json:"enabled"`
		Name      string `yaml:"name" json:"name"`
		Key       string `yaml:"key" json:"key"`
	} `yaml:"myEntry" json:"myEntry"`
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
		name: "my-default",
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
GlobalAppCtx
```
**Fields in rkentry.GlobalAppCtx**

| Element | Description | JSON | Default values |
| ------ | ------ | ------ | ------ |
| startTime | Application start time. | startTime | 0001-01-01 00:00:00 +0000 UTC |
| appInfoEntry | See ApplicationInfoEntry for detail. | appInfoEntry | Includes application info specified by user. |
| zapLoggerEntries | See ZapLoggerEntry for detail. | zapLoggerEntries | Includes zap logger entity initiated by user or system by default. |
| eventLoggerEntries | See EventLoggerEntry for detail. | eventLoggerEntries | Includes query logger entity initiated by user or system by default. |
| configEntries | See ConfigEntry for detail. | configEntries | Includes viper config entity initiated by user or system by default. |
| certEntries | See CertEntry for detail. | configEntries | Includes certificates retrieved while bootstrapping with configuration initiated by user. |
| externalEntries | User implemented Entry. | configEntries | Includes user implemented Entry configuration initiated by user. |
| userValues | User K/V registered from code. | userValues | empty map |
| shutdownSig | Shutdown signals which includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT. | shutdown_sig | channel includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT |
| shutdownHooks | Shutdown hooks registered from user code. | shutdown_hooks | empty list |


#### Usage of GlobalAppCtx
- Access start time and up time of application.
```go
// Get start time of application
startTime := rkentry.GlobalAppCtx.GetStartTime()

// Get up time of application/process.
upTime := rkentry.GlobalAppCtx.GetUpTime()
```

- Access AppInfoEntry
```go
// Get AppInfoEntry from GlobalAppCtx
appInfoEntry := rkentry.GlobalAppCtx.GetAppInfoEntry()

// Get AppInfoEntry as Entry type
appInfoEntryRaw := rkentry.GetAppInfoEntryRaw()

// Set AppInfoEntry
entryName := SetAppInfoEntry(<AppInfoEntry>)
```

- Access ZapLoggerEntry
```go
// Get ZapLoggerEntry from GlobalAppCtx
zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry("entry name")

// List ZapLoggerEntry as map[<entry name>]*ZapLoggerEntry
zapLoggerEntries := rkentry.GlobalAppCtx.ListZapLoggerEntries()

// List ZapLoggerEntry as map[<entry name>]*Entry
zapLoggerEntriesRaw := rkentry.GlobalAppCtx.GetZapLoggerEntriesRaw()

// Add ZapLoggerEntry
entryName := rkentry.GlobalAppCtx.AddZapLoggerEntry(<ZapLoggerEntry>)

// Remove ZapLoggerEntry
success := rkentry.GlobalAppCtx.RemoveZapLoggerEntry("entry name")

// Get zap.Logger from ZapLoggerEntry
zapLogger := rkentry.GlobalAppCtx.GetZapLogger("entry name")

// Get zap.Config from ZapLoggerEntry
zapLoggerConfig := rkentry.GlobalAppCtx.GetZapLoggerConfig("entry name")

// Get default ZapLoggerEntry whose output is stdout
zapLoggerEntryDefault := rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()

// Get zap.Logger from default ZapLoggerEntry
zapLoggerDefault := rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()

// Get zap.Config from default ZapLoggerEntry
zapConfigDefault := rkentry.GlobalAppCtx.GetZapLoggerConfigDefault()
```

- Access EventLoggerEntry
```go
// Get EventLoggerEntry from GlobalAppCtx
eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry("entry name")

// List EventLoggerEntry as map[<entry name>]*EventLoggerEntry
eventLoggerEntry := rkentry.GlobalAppCtx.ListEventLoggerEntries()

// List EventLoggerEntry as map[<entry name>]*Entry
eventLoggerEntryRaw := rkentry.GlobalAppCtx.ListEventLoggerEntriesRaw()

// Add EventLoggerEntry
entryName := rkentry.GlobalAppCtx.AddEventLoggerEntry(<EventLoggerEntry>)

// Remove EventLoggerEntry
success := rkentry.GlobalAppCtx.RemoveEventLoggerEntry("entry name")

// Get rkquery.EventFactory from EventLoggerEntry
eventFactory := rkentry.GlobalAppCtx.GetEventFactory("entry name")

// Get rkquery.EventHelper from EventLoggerEntry
eventHelper := rkentry.GlobalAppCtx.GetEventHelper("entry name")

// Get default EventLoggerEntry whose output is stdout
eventLoggerEntryDefault := rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
```

- Access CertEntry
```go
// Get CertEntry from GlobalAppCtx
certEntry := rkentry.GlobalAppCtx.GetCertEntry("entry name")

// List CertEntry as map[<entry name>]*CertEntry
certEntries := rkentry.GlobalAppCtx.ListCertEntries()

// List CertEntry as map[<entry name>]*Entry
certEntriesRaw := rkentry.GlobalAppCtx.ListCertEntriesRaw()

// Remove CertEntry
success := rkentry.GlobalAppCtx.RemoveCertEntry("entry name")

// Add CertEntry
entryName := rkentry.GlobalAppCtx.AddCertEntry(<CertEntry>)
```

- Access ConfigEntry
```go
// Get ConfigEntry from GlobalAppCtx
configEntry := rkentry.GlobalAppCtx.GetConfigEntry("entry name")

// List ConfigEntry as map[<entry name>]*ConfigEntry
configEntries := rkentry.GlobalAppCtx.ListConfigEntryEntries()

// List ConfigEntry as map[<entry name>]*Entry
certEntriesRaw := rkentry.GlobalAppCtx.ListConfigEntryEntriesRaw()

// Remove ConfigEntry
success := rkentry.GlobalAppCtx.RemoveConfigEntry("entry name")

// Add ConfigEntry
entryName := rkentry.GlobalAppCtx.AddConfigEntry(<ConfigEntry>)
```

- Access entries
Entries contains user defined entry.
```go
// Access entries from GlobalAppCtx
entry := rkentry.GlobalAppCtx.GetEntry("entry name")

// Access entries via utility function
entries := rkentry.GlobalAppCtx.ListEntries()

// Add entry
rkentry.GlobalAppCtx.AddEntry()

// Remove ConfigEntry
success := rkentry.GlobalAppCtx.RemoveEntry("entry name")

// Get entry with name
entry := rkentry.GlobalAppCtx.GetEntry("name of your entry")

// Merge map of entries into rkentry.GlobalAppCtx.entries
rkentry.GlobalAppCtx.MergeEntries(mapOfYourEntries map[string]rkentry.Entry)
```

- Access Values

User can add/get/list/remove any values into map of rkentry.GlobalAppCtx.UserValues as needed.

GlobalAppCtx don't provide any locking mechanism.
```go
// Add k/v value into GlobalAppCtx, key should be string and value could be any kind
rkentry.GlobalAppCtx.AddValue(<"key">, <value interface{}>)

// Get value with key
value := rkentry.GlobalAppCtx.GetValue(<"key">)

// List values
entries := rkentry.GlobalAppCtx.ListValues()

// Remove value with key
rkentry.GlobalAppCtx.RemoveValue(<"key">)

// Clear values
rkentry.GlobalAppCtx.ClearValues()
```

- Access shutdown sig
```go
// Access shutdown signal directly
rkentry.GlobalAppCtx.GetShutdownSig()

// Wait for shutdown signal via utility function, otherwise, user must call by himself
rkentry.GlobalAppCtx.WaitForShutdownSig()
```

- Access shutdown hooks

Users can add their own shutdown hook function into GlobalAppCtx.

rkboot will iterate all shutdown hooks in GlobalAppCtx and call every shutdown hook function.
```go
// Get shutdown
rkentry.GlobalAppCtx.GetShutdownHook("name of shutdown hook")

// Add shutdown hook function with name
rkentry.GlobalAppCtx.AddShutdownHook(<"name">, <"function">)

// List shutdown hooks
rkentry.GlobalAppCtx.ListShutdownHooks()

// Remove shutdown hook
success := rkentry.GlobalAppCtx.RemoveShutdownHook("name of shutdown hook")

// Internal 
rkentry.GlobalAppCtx.GetShutdownHook("name of shutdown hook function")
```

### Built in internal entries
#### AppInfoEntry
AppInfoEntry contains bellow fields which could be overridden via YAML file or code.

| Name | Description | YAML | Default value |
| ------ | ------ | ------ | ------ |
| EntryName | Name of entry. | N/A | AppInfoDefault |
| EntryType | Type of entry which is EventEntry. | N/A | AppInfoEntry |
| EntryDescription | Description of entry. | N/A | Internal RK entry which describes application with fields of appName, version and etc. |
| AppName | Application name which refers to go process. | appName | rkApp |
| Version | Application version. | version | v0.0.0 |
| Lang | Programming language <NOT configurable!> .| N/A | golang |
| Description | Description of application itself. | description | "" |
| Keywords | A set of words describe application. | keywords | [] |
| HomeUrl | Home page URL. | homeURL | "" |
| IconUrl | Application Icon URL. | iconURL | "" |
| DocsUrl | A set of URLs of documentations of application. | docsURL | [] |
| Maintainers | Maintainers of application. | maintainers | [] |
| Dependencies | Application dependencies which is parsed from result of go list -json all | depDocPath | [] |

##### YAML Hierarchy
```yaml
rk:
  appName: rk-example-entry           # Optional, default: "rkApp"
  version: v0.0.1                     # Optional, default: "v0.0.0"
  description: "this is description"  # Optional, default: ""
  keywords: ["rk", "golang"]          # Optional, default: []
  homeUrl: "http://example.com"       # Optional, default: ""
  iconUrl: "http://example.com"       # Optional, default: ""
  docsUrl: ["http://example.com"]     # Optional, default: []
  maintainers: ["rk-dev"]             # Optional, default: []
  depDocPath: ""                      # Optional, default: ""
```

##### Access AppInfoEntry
```go
// Access entry
rkentry.GlobalAppCtx.GetAppInfoEntry()

// Access default fields
rkentry.GlobalAppCtx.GetAppInfoEntry().EntryName
rkentry.GlobalAppCtx.GetAppInfoEntry().EntryType
rkentry.GlobalAppCtx.GetAppInfoEntry().EntryDescription

// Access fields in entry
rkentry.GlobalAppCtx.GetAppInfoEntry().AppName
rkentry.GlobalAppCtx.GetAppInfoEntry().Version
rkentry.GlobalAppCtx.GetAppInfoEntry().Lang
rkentry.GlobalAppCtx.GetAppInfoEntry().Description
rkentry.GlobalAppCtx.GetAppInfoEntry().Keywords
rkentry.GlobalAppCtx.GetAppInfoEntry().HomeUrl
rkentry.GlobalAppCtx.GetAppInfoEntry().IconUrl
rkentry.GlobalAppCtx.GetAppInfoEntry().DocsUrl
rkentry.rkentry.GlobalAppCtx.GetAppInfoEntry().Maintainers
rkentry.rkentry.GlobalAppCtx.GetAppInfoEntry().Dependencies
```

##### Stringfy AppInfoEntry
Assuming we have application info YAML as bellow:

```yaml
---
rk: 
  appName: rk-example-entry           # Optional, default: "rkApp"
  version: v0.0.1                     # Optional, default: "v0.0.0"
  description: "this is description"  # Optional, default: ""
  keywords: ["rk", "golang"]          # Optional, default: []
  homeUrl: "http://example.com"       # Optional, default: ""
  iconUrl: "http://example.com"       # Optional, default: ""
  docsUrl: ["http://example.com"]     # Optional, default: []
  maintainers: ["rk-dev"]             # Optional, default: []
  depDocPath: ""                      # Optional, default: ""
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetAppInfoEntry().String())
```

Process information could be printed either.
```json
{
    "entryName":"AppInfoDefault",
    "entryType":"AppInfoEntry",
    "entryDescription":"Internal RK entry which describes application with fields of appName, version and etc.",
    "description":"this is description",
    "appName":"rk-example-entry",
    "version":"v0.0.1",
    "lang":"golang",
    "keywords":[
        "rk",
        "golang"
    ],
    "homeUrl":"http://example.com",
    "iconUrl":"http://example.com",
    "docsUrl":[
        "http://example.com"
    ],
    "maintainers":[
        "rk-dev"
    ]
}
```

#### EventLoggerEntry
EventLoggerEntry is used for [rkquery](https://github.com/rookie-ninja/rk-query) whose responsibility is logging event like RPC or periodical jobs.

| Name | Description |
| ------ | ------ |
| EntryName | Name of entry. |
| EntryType | Type of entry which is EventEntry. |
| EntryDescription | Description of entry. |
| EventFactory | rkquery.EventFactory was initialized at the beginning. |
| EventHelper | rkquery.EventHelper was initialized at the beginning. |
| LoggerConfig | zap.Config which was initialized at the beginning which is not accessible after initialization. |
| LumberjackConfig | lumberjack.Logger which was initialized at the beginning. |

##### YAML Hierarchy
EventLoggerEntry needs application name while creating event log. 
As a result, it is recommended to add AppInfoEntry while initializing event logger entry. 
Otherwise, default application name would be assigned.

| Name | Description | Default |
| ------ | ------ | ------ |
| rk.appName | Application name which refers to go process. | rkapp | 
| eventLogger.name | Required. Name of event logger entry. | N/A |
| eventLogger.description | Description of event logger entry. | N/A |
| eventLogger.encoding | Encoding of event logger, console & json is supported. Please refer rkquery.CONSOLE & rkquery.JSON. | console | 
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
    description: "This is description" # Optional
    encoding: console                  # Optional, default: console, options: console and json
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
  - name: event-logger                    # Required
    description: "Description of entry"   # Optional
    encoding: console                     # Optional, default: console, options: console and json
    outputPaths: ["stdout"]               # Optional
    lumberjack:                           # Optional
      filename: "rkapp-event.log"         # Optional, default: It uses <processname>-lumberjack.log in os.TempDir() if empty.
      maxsize: 1024                       # Optional, default: 1024 (MB)
      maxage: 7                           # Optional, default: 7 (days)
      maxbackups: 3                       # Optional, default: 3 (days)
      localtime: true                     # Optional, default: true
      compress: true                      # Optional, default: true
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetEventLoggerEntry("event-logger").String())
```

Process information could be printed either.
```json
{
    "entryName":"event-logger",
    "entryType":"EventLoggerEntry",
    "entryDescription":"Description of entry",
    "zapConfig":{
        "level":"info",
        "development":false,
        "disableCaller":false,
        "disableStacktrace":false,
        "sampling":null,
        "encoding":"console",
        "encoderConfig":{
            "messageKey":"msg",
            "levelKey":"",
            "timeKey":"",
            "nameKey":"",
            "callerKey":"",
            "functionKey":"",
            "stacktraceKey":"",
            "lineEnding":"",
            "levelEncoder":"capital",
            "timeEncoder":"ISO8601",
            "durationEncoder":"secs",
            "callerEncoder":"full",
            "nameEncoder":"full",
            "consoleSeparator":""
        },
        "outputPaths":[
            "stdout"
        ],
        "errorOutputPaths":[
            "stderr"
        ],
        "initialFields":{

        }
    },
    "lumberjackConfig":{
        "filename":"rkapp-event.log",
        "maxsize":1024,
        "maxage":7,
        "maxbackups":3,
        "localtime":true,
        "compress":true
    }
}
```

#### ZapLoggerEntry
ZapLoggerEntry is used for initializing zap logger.

| Name | Description |
| ------ | ------ |
| EntryName | Name of entry. |
| EntryType | Type of entry which is EventEntry. |
| EntryDescription | Description of entry. |
| Logger | zap.Logger which was initialized at the beginning |
| LoggerConfig | zap.Config which was initialized at the beginning |
| LumberjackConfig | lumberjack.Logger which was initialized at the beginning |

##### YAML Hierarchy
ZapLoggerEntry follows zap and lumberjack YAML hierarchy, please refer to [zap](https://pkg.go.dev/go.uber.org/zap#section-documentation) and [lumberjack](https://github.com/natefinch/lumberjack) site for details.

| Name | Description | Default |
| ------ | ------ | ------ |
| zapLogger.name | Required. Name of zap logger entry | N/A |
| zapLogger.description | Description of zap logger entry. | N/A |
| zapLogger.zap.level | Level is the minimum enabled logging level. | info | 
| zapLogger.zap.development | Development puts the logger in development mode, which changes the behavior of DPanicLevel and takes stacktraces more liberally. | true |
| zapLogger.zap.disableCaller | DisableCaller stops annotating logs with the calling function's file name and line number. | false |
| zapLogger.zap.disableStacktrace | DisableStacktrace completely disables automatic stacktrace capturing. | true |
| zapLogger.zap.sampling | Sampling sets a sampling policy. | nil |
| zapLogger.zap.encoding | Encoding sets the logger's encoding. Valid values are "json" and "console", as well as any third-party encodings registered via RegisterEncoder. | console |
| zapLogger.zap.encoderConfig.messageKey | As name described. | msg |
| zapLogger.zap.encoderConfig.levelKey | As name described. | level |
| zapLogger.zap.encoderConfig.timeKey | As name described. | ts |
| zapLogger.zap.encoderConfig.nameKey | As name described. | logger |
| zapLogger.zap.encoderConfig.callerKey | As name described. | caller |
| zapLogger.zap.encoderConfig.functionKey | As name described. | "" |
| zapLogger.zap.encoderConfig.stacktraceKey | As name described. | stacktraceKey |
| zapLogger.zap.encoderConfig.lineEnding | As name described. | \n |
| zapLogger.zap.encoderConfig.timeEncoder | As name described. | iso8601 |
| zapLogger.zap.encoderConfig.durationEncoder | As name described. | string |
| zapLogger.zap.encoderConfig.callerEncoder | As name described. | "" |
| zapLogger.zap.encoderConfig.nameEncoder | As name described. | "" |
| zapLogger.zap.encoderConfig.consoleSeparator | As name described. | "" |
| zapLogger.zap.outputPaths | Output paths. | ["stdout"] |
| zapLogger.zap.errorOutputPaths | Output paths. | ["stderr"] |
| zapLogger.zap.initialFields | Output paths. | empty map |
| zapLogger.lumberjack.filename | Filename is the file to write logs to | It uses <processname>-lumberjack.log in os.TempDir() if empty. |
| zapLogger.lumberjack.maxsize | MaxSize is the maximum size in megabytes of the log file before it gets rotated. | 1024 |
| zapLogger.lumberjack.maxage | MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename. | 7 |
| zapLogger.lumberjack.maxbackups | axBackups is the maximum number of old log files to retain. | 3 |
| zapLogger.lumberjack.localtime | LocalTime determines if the time used for formatting the timestamps in backup files is the computer's local time. | true |
| zapLogger.lumberjack.compress | Compress determines if the rotated log files should be compressed using gzip. | true |

```yaml
---
zapLogger:
  - name: zap-logger                      # Required
    description: "Description of entry"   # Optional
    zap:
      level: info                         # Optional, default: info, options: [debug, DEBUG, info, INFO, warn, WARN, dpanic, DPANIC, panic, PANIC, fatal, FATAL]
      development: true                   # Optional, default: true
      disableCaller: false                # Optional, default: false
      disableStacktrace: true             # Optional, default: true
      sampling:                           # Optional, default: empty map
        initial: 0
        thereafter: 0
      encoding: console                   # Optional, default: "console", options: [console, json]
      encoderConfig:
        messageKey: "msg"                 # Optional, default: "msg"
        levelKey: "level"                 # Optional, default: "level"
        timeKey: "ts"                     # Optional, default: "ts"
        nameKey: "logger"                 # Optional, default: "logger"
        callerKey: "caller"               # Optional, default: "caller"
        functionKey: ""                   # Optional, default: ""
        stacktraceKey: "stacktrace"       # Optional, default: "stacktrace"
        lineEnding: "\n"                  # Optional, default: "\n"
        levelEncoder: "capitalColor"      # Optional, default: "capitalColor", options: [capital, capitalColor, color, lowercase]
        timeEncoder: "iso8601"            # Optional, default: "iso8601", options: [rfc3339nano, RFC3339Nano, rfc3339, RFC3339, iso8601, ISO8601, millis, nanos]
        durationEncoder: "string"         # Optional, default: "string", options: [string, nanos, ms]
        callerEncoder: ""                 # Optional, default: ""
        nameEncoder: ""                   # Optional, default: ""
        consoleSeparator: ""              # Optional, default: ""
      outputPaths: [ "stdout" ]           # Optional, default: ["stdout"], stdout would be replaced if specified
      errorOutputPaths: [ "stderr" ]      # Optional, default: ["stderr"], stderr would be replaced if specified
      initialFields:                      # Optional, default: empty map
        key: "value"
    lumberjack:                           # Optional
      filename: "rkapp-event.log"         # Optional, default: It uses <processname>-lumberjack.log in os.TempDir() if empty.
      maxsize: 1024                       # Optional, default: 1024 (MB)
      maxage: 7                           # Optional, default: 7 (days)
      maxbackups: 3                       # Optional, default: 3 (days)
      localtime: true                     # Optional, default: true
      compress: true                      # Optional, default: true
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
    "entryName":"zap-logger",
    "entryType":"ZapLoggerEntry",
    "entryDescription":"Description of entry",
    "zapConfig":{
        "level":"info",
        "development":false,
        "disableCaller":false,
        "disableStacktrace":false,
        "sampling":null,
        "encoding":"console",
        "encoderConfig":{
            "messageKey":"msg",
            "levelKey":"level",
            "timeKey":"ts",
            "nameKey":"logger",
            "callerKey":"caller",
            "functionKey":"",
            "stacktraceKey":"stacktrace",
            "lineEnding":"\n",
            "levelEncoder":"capitalColor",
            "timeEncoder":"ISO8601",
            "durationEncoder":"string",
            "callerEncoder":"short",
            "nameEncoder":"full",
            "consoleSeparator":""
        },
        "outputPaths":[
            "stdout"
        ],
        "errorOutputPaths":[
            "stderr"
        ],
        "initialFields":null
    },
    "lumberjackConfig":{
        "filename":"rkapp-event.log",
        "maxsize":1024,
        "maxage":7,
        "maxbackups":3,
        "localtime":true,
        "compress":true
    }
}
```

#### ConfigEntry
ConfigEntry provides convenient way to initialize viper instance. [viper](https://github.com/spf13/viper) is a complete configuration solution for Go applications.
Each viper instance combined with one configuration file. 

| Name | Description |
| ------ | ------ |
| EntryName | Name of entry. |
| EntryType | Type of entry which is EventEntry. |
| EntryDescription | Description of entry. |
| Locale | <realm>::<region>::<az>::<domain> |
| Path | File path of config file, could be either relative or absolute path. |
| vp | Viper instance. |

##### YAML Hierarchy

| Name | Description | Required | Default |
| ------ | ------ | ------ | ------ |
| config.name | Name of config entry. | required | config-<random string> |
| config.path | File path of config file, could be either relative or absolute path | required | "" | 
| config.locale | <realm>::<region>::<az>::<domain> | required | "" |
| config.description | Description of config entry. | "" |

```yaml
config:
  - name: my-config                       # Required
    locale: "*::*::*::*"                  # Required
    path: example/my-config.yaml          # Required
    description: "Description of entry"   # Optional
```

##### Access ConfigEntry
```go
// Access entry
rkentry.GlobalAppCtx.GetConfigEntry("my-config"))

// Access viper instance
rkentry.GlobalAppCtx.GetConfigEntry("my-config").GetViper()
```

##### Stringfy ConfigEntry
Assuming we have config YAML as bellow:

```yaml
---
config:
  - name: my-config
    path: example/my-config.yaml
    locale: "*::*::*::*"
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetViperEntry("my-config").String())
```

Process information could be printed either.
```json
{
    "entryName":"my-config",
    "entryType":"ConfigEntry",
    "entryDescription":"Description of entry",
    "locale":"*::*::*::*",
    "path":"/usr/rk/example/my-config.yaml",
    "viper":{
        "key":"value"
    }
}
```

##### Dynamically load config based on Environment variable
We are using <locale> in yaml config and OS environment variable to distinguish different cert entries.

```
RK use <realm>::<region>::<az>::<domain> to distinguish different environment.
Variable of <locale> could be composed as form of <realm>::<region>::<az>::<domain>
- realm: It could be a company, department and so on, like RK-Corp.
         Environment variable: REALM
         Eg: RK-Corp
         Wildcard: supported

- region: Please see AWS web site: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
          Environment variable: REGION
          Eg: us-east
          Wildcard: supported

- az: Availability zone, please see AWS web site for details: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
      Environment variable: AZ
      Eg: us-east-1
      Wildcard: supported

- domain: Stands for different environment, like dev, test, prod and so on, users can define it by themselves.
          Environment variable: DOMAIN
          Eg: prod
          Wildcard: supported

How it works?
First, we will split locale with "::" and extract realm, region, az and domain.
Second, get environment variable named as REALM, REGION, AZ and DOMAIN.
Finally, compare every element in locale variable and environment variable.
If variables in locale represented as wildcard(*), we will ignore comparison step.

Example:
# let's assuming we are going to define DB address which is different based on environment.
# Then, user can distinguish DB address based on locale.
# We recommend to include locale with wildcard.
---
DB:
  - name: redis-default
    locale: "*::*::*::*"
    addr: "192.0.0.1:6379"
  - name: redis-in-test
    locale: "*::*::*::test"
    addr: "192.0.0.1:6379"
  - name: redis-in-prod
    locale: "*::*::*::prod"
    addr: "176.0.0.1:6379"
```

#### CertEntry
CertEntry provides a convenient way to retrieve certifications from local or remote services.
Supported services listed bellow:
- localFs
- remoteFs
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
| cert.consul.datacenter | consul datacenter. | "" |
| cert.consul.token | Token for access consul. | "" |
| cert.consul.basicAuth | Basic auth for consul server, like <user:pass>. | "" |
| cert.consul.serverCertPath | Path of server cert in Consul server. | "" |
| cert.consul.serverKeyPath | Path of server key in Consul server. | "" |
| cert.consul.clientCertPath | Path of client cert in Consul server. | "" |
| cert.consul.clientCertPath | Path of client key in Consul server. | "" |
| cert.etcd.name | Name of etcd retriever | "" |
| cert.etcd.locale | Represent environment of current process follows schema of \<realm\>::\<region\>::\<az\>::\<domain\> | \*::\*::\*::\* | 
| cert.etcd.endpoint | Endpoint of etcd server, http://x.x.x.x or x.x.x.x both acceptable. | N/A |
| cert.etcd.basicAuth | Basic auth for etcd server, like <user:pass>. | "" |
| cert.etcd.serverCertPath | Path of server cert in etcd server. | "" |
| cert.etcd.serverKeyPath | Path of server key in etcd server. | "" |
| cert.etcd.clientCertPath | Path of client cert in etcd server. | "" |
| cert.etcd.clientCertPath | Path of client key in etcd server. | "" |
| cert.localFs.name | Name of localFs retriever | "" |
| cert.localFs.locale | Represent environment of current process follows schema of \<realm\>::\<region\>::\<az\>::\<domain\> | \*::\*::\*::\* | 
| cert.localFs.serverCertPath | Path of server cert in local file system. | "" |
| cert.localFs.serverKeyPath | Path of server key in local file system. | "" |
| cert.localFs.clientCertPath | Path of client cert in local file system. | "" |
| cert.localFs.clientCertPath | Path of client key in local file system. | "" |
| cert.remoteFs.name | Name of remoteFileStore retriever | "" |
| cert.remoteFs.locale | Represent environment of current process follows schema of \<realm\>::\<region\>::\<az\>::\<domain\> | \*::\*::\*::\* | 
| cert.remoteFs.endpoint | Endpoint of remoteFileStore server, http://x.x.x.x or x.x.x.x both acceptable. | N/A |
| cert.remoteFs.basicAuth | Basic auth for remoteFileStore server, like <user:pass>. | "" |
| cert.remoteFs.serverCertPath | Path of server cert in remoteFs server. | "" |
| cert.remoteFs.serverKeyPath | Path of server key in remoteFs server. | "" |
| cert.remoteFs.clientCertPath | Path of client cert in remoteFs server. | "" |
| cert.remoteFs.clientCertPath | Path of client key in remoteFs server. | "" |

##### Access CertEntry
```go
// Access entry
certEntry := rkentry.GlobalAppCtx.GetCertEntry("name of cert entry")

// Access cert stores which contains certificates as byte array
serverCert := certEntry.Store.ServerCert
serverKey := certEntry.Store.ServerKey
clientCert := certEntry.Store.ClientCert
clientKey := certEntry.Store.ClientKey
```

##### Stringfy CertEntry
Assuming we have cert YAML as bellow:

```yaml
---
cert:
  - name: "local-cert"                       # Required
    description: "Description of entry"      # Optional
    provider: "localFs"                      # Required, etcd, consul, localFS, remoteFs are supported options
    locale: "*::*::*::*"                     # Optional, default: *::*::*::*
    serverCertPath: "example/server.pem"     # Optional, default: "", path of certificate on local FS
    serverKeyPath: "example/server-key.pem"  # Optional, default: "", path of certificate on local FS
    clientCertPath: "example/client.pem"     # Optional, default: "", path of certificate on local FS
    clientKeyPath: "example/client.pem"      # Optional, default: "", path of certificate on local FS
```

```go
fmt.Println(rkentry.GlobalAppCtx.GetCertEntry("local-cert").String())
```

Process information could be printed either.
```json
{
    "entryName":"local-cert",
    "entryType":"CertEntry",
    "entryDescription":"",
    "eventLoggerEntry":"eventLoggerDefault",
    "zapLoggerEntry":"zapLoggerDefault",
    "retriever":{
        "provider":"localFs",
        "locale":"*::*::*::*",
        "serverCertPath":"/usr/rk/example/server.pem",
        "serverKeyPath":"/usr/rk/example/server-key.pem",
        "clientCertPath":"/usr/rk/example/client.pem",
        "clientKeyPath":"/usr/rk/example/client-key.pem"
    },
    "store":{
        "clientCert":"",
        "serverCert":"Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 509952525898157401469127978785735696864294231142 (0x59530fe773bcdf9fa78ecaf6dd46485f62845c66)
    Signature Algorithm: SHA256-RSA
        Issuer: C=CN,ST=Beijing,UnknownOID=2.5.4.7,O=RK,OU=RK Demo,CN=RK Demo CA
        Validity
            Not Before: Apr 7 13:03:00 2021 UTC
            Not After : Apr 6 13:03:00 2026 UTC
        Subject: C=CN,ST=Beijing,UnknownOID=2.5.4.7,CN=example.net
        Subject Public Key Info:
            Public Key Algorithm: ECDSA
                Public-Key: (256 bit)
                X:
                    5f:f2:9f:e7:c6:f6:35:1c:75:24:25:76:64:7c:54:
                    20:0e:d4:36:08:af:43:38:07:ba:cb:79:44:db:51:
                    d3:27
                Y:
                    f5:62:e2:f1:cd:a9:69:28:f6:4a:32:62:02:aa:81:
                    45:f2:8a:ae:65:18:0b:30:a3:af:0a:be:60:fb:73:
                    62:f8
                Curve: P-256
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature, Key Encipherment
            X509v3 Extended Key Usage:
                TLS Web Server Authentication
            X509v3 Basic Constraints: critical
                CA:FALSE
            X509v3 Subject Key Identifier:
                EF:E9:D5:25:10:8E:8D:71:00:73:8A:19:F3:29:9A:F1:2A:96:E7:4D
            X509v3 Authority Key Identifier:
                keyid:60:C2:96:0A:86:07:DE:3B:7A:76:5E:E5:F4:85:ED:F9:71:A7:94:81
            X509v3 Subject Alternative Name:
                DNS:localhost
                IP Address:127.0.0.1, IP Address:0.0.0.0

    Signature Algorithm: SHA256-RSA
         1c:aa:2d:cd:d0:91:a1:8d:af:e4:2a:8c:5c:3b:ce:79:3d:8f:
         45:f8:51:c9:bf:d4:de:a2:18:8a:7c:7d:48:bd:b7:e9:92:d4:
         b6:ba:08:1d:5e:8f:e1:68:5a:44:98:b2:4c:ea:06:97:48:92:
         ac:ea:65:a9:f5:72:1c:85:9e:38:6d:32:c5:a0:65:85:ba:b2:
         b5:cd:46:00:72:03:d5:d2:da:cd:bc:fb:18:3d:a3:07:29:71:
         e5:da:df:01:b8:c1:b9:d3:8f:57:a6:72:e9:f3:4a:b2:95:16:
         34:b5:20:95:23:97:91:d6:08:a8:9b:9d:58:e0:b7:88:34:b3:
         db:5c:f0:a6:29:5e:41:98:97:6d:d4:e0:55:2d:6b:fd:3b:60:
         59:ad:5b:94:83:17:7f:0d:5a:b3:4e:7b:9a:95:10:02:cc:cf:
         54:72:d0:f3:20:63:44:83:b4:a8:5d:74:33:67:35:2d:f6:7f:
         3e:34:20:fd:30:07:b6:b7:d5:3d:b8:79:97:ad:64:af:c8:35:
         a2:22:fb:94:8c:d2:22:f1:91:fa:8d:ef:5c:4a:c4:02:08:62:
         6b:88:f5:8c:16:08:56:99:db:ea:9b:2c:88:83:4d:b4:7a:83:
         2b:8b:21:0d:80:7f:00:4b:67:4e:a0:f7:7f:59:7d:2a:99:b8:
         ff:0f:a3:15
"
    }
}
```

##### Select cert entry dynamically based on environment
We are using <locale> in yaml config and OS environment variable to distinguish different cert entries.

```
RK use <realm>::<region>::<az>::<domain> to distinguish different environment.
Variable of <locale> could be composed as form of <realm>::<region>::<az>::<domain>
- realm: It could be a company, department and so on, like RK-Corp.
         Environment variable: REALM
         Eg: RK-Corp
         Wildcard: supported

- region: Please see AWS web site: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
          Environment variable: REGION
          Eg: us-east
          Wildcard: supported

- az: Availability zone, please see AWS web site for details: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
      Environment variable: AZ
      Eg: us-east-1
      Wildcard: supported

- domain: Stands for different environment, like dev, test, prod and so on, users can define it by themselves.
          Environment variable: DOMAIN
          Eg: prod
          Wildcard: supported

How it works?
First, we will split locale with "::" and extract realm, region, az and domain.
Second, get environment variable named as REALM, REGION, AZ and DOMAIN.
Finally, compare every element in locale variable and environment variable.
If variables in locale represented as wildcard(*), we will ignore comparison step.

Example:
# let's assuming we are going to define DB address which is different based on environment.
# Then, user can distinguish DB address based on locale.
# We recommend to include locale with wildcard.
---
DB:
  - name: redis-default
    locale: "*::*::*::*"
    addr: "192.0.0.1:6379"
  - name: redis-in-test
    locale: "*::*::*::test"
    addr: "192.0.0.1:6379"
  - name: redis-in-prod
    locale: "*::*::*::prod"
    addr: "176.0.0.1:6379"
```

### Info Utility
#### ProcessInfo
Process information for a running application.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| AppName | Application name which refers to go process. | based on user config | appName |
| Version | Application version. | based on user config | version |
| Lang | Programming language <NOT configurable!>. | based on user config | lang |
| Description | Description of application itself. | based on user config | description |
| Keywords | A set of words describe application. | based on user config | keywords |
| HomeUrl | Home page URL. | based on user config | homeUrl |
| IconUrl | Application Icon URL. | based on user config | iconUrl |
| DocsUrl | A set of URLs of documentations of application. | based on user config | docsUrl |
| Maintainers | Maintainers of application. | based on user config | maintainers |
| UID | user id which runs process | based on system | uid |
| GID | group id which runs process | based on system | gid |
| Username | Username which runs process. | based on system | username |
| StartTime | application start time | time.now() | startTime |
| UpTimeSec | application up time in seconds | zero at the beginning | upTimeSec |
| UpTimeStr | application up time in string | zero as string at the beginning | upTimeStr |
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
  "appName": "rk-example-entry",
  "version": "v0.0.1",
  "description": "this is description",
  "keywords": [
    "rk",
    "golang"
  ],
  "homeUrl": "http://example.com",
  "iconUrl": "http://example.com",
  "docsUrl": [
    "http://example.com"
  ],
  "maintainers": [
    "rk-dev"
  ],
  "uid": "501",
  "gid": "20",
  "username": "rk-dev",
  "startTime": "2021-05-14T02:59:51+08:00",
  "upTimeSec": 0,
  "upTimeStr": "12 milliseconds",
  "region": "unknown",
  "az": "unknown",
  "realm": "unknown",
  "domain": "dev"
}
```

#### CpuInfo
CPU information for a running application.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| CpuUsedPercentage | CPU usage. | based on system | cpuUsedPercentage |
| LogicalCoreCount | Logical core count. | based on system | logicalCoreCount |
| PhysicalCoreCount | Physical core count. | based on system | physicalCoreCount |
| VendorId | CPU vendor Id. | based on system | vendorId |
| ModelName | CPU model name. | based on system | modelName |
| Mhz | CPU power. | based on system | mhz |
| CacheSize | CPU cache size in bytes. | based on system | cacheSize |

##### Access CpuInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewCpuInfo()))
```

```json
{
  "cpuUsedPercentage": 11.88,
  "logicalCoreCount": 8,
  "physicalCoreCount": 4,
  "vendorId": "GenuineIntel",
  "modelName": "Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz",
  "mhz": 2000,
  "cacheSize": 256
}
```

#### MemInfo
Memory information for a running application.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| MemUsedPercentage | Memory usage in percentage. | based on system | memUsedPercentage |
| MemUsedMb | Memory usage in megabytes. | based on system | memUsedMb |
| MemAllocByte | Bytes of allocated heap objects. | based on system | memAllocByte |
| SysAllocByte | Total bytes of memory obtained from the OS. | based on system | sysAllocByte |
| LastGcTimestamp | Last GC timestamp. | based on system | lastGcTimestamp |
| GcCount | GC count. | based on system | gcCountTotal |
| ForceGcCount | GC count triggered by user. | based on system | forceGcCount |

##### Access MemInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewMemInfo()))
```

```json
{
  "memUsedPercentage": 0.04,
  "memUsedMb": 2,
  "memAllocByte": 2728536,
  "sysAllocByte": 72436744,
  "lastGcTimestamp": "1970-01-01T08:00:00+08:00",
  "gcCountTotal": 0,
  "forceGcCount": 0
}
```

#### NetInfo
Network information for a running application.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| NetInterface.Name | Name of Network interface. | based on system | name |
| NetInterface.Mtu | Maximum transmission unit. | based on system | mtu |
| NetInterface.HardwareAddr | IEEE MAC-48, EUI-48 and EUI-64 form. | based on system | hardwareAddr |
| NetInterface.Flags | Flags of Network interface. | based on system | flags |
| NetInterface.Addrs | A list of unicast interface addresses for a specific interface. | based on system | addrs |
| NetInterface.MulticastAddrs | List of multicast, joined group addresses for a specific interface. | based on system | multicastAddrs |

##### Access NetInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewNetInfo()))
```

```json
{
  "netInterface": [
    {
      "name": "lo0",
      "mtu": 16384,
      "hardwareAddr": "",
      "flags": [
        "up",
        "loopback",
        "multicast"
      ],
      "addrs": [
        "127.0.0.1/8",
        "::1/128",
        "fe80::1/64"
      ],
      "multicastAddrs": [
        "ff02::fb",
        "224.0.0.251",
        "ff02::2:ff33:9cc0",
        "ff01::1",
        "ff02::1",
        "ff02::1:ff00:1",
        "224.0.0.1"
      ]
    },
    {
      "name": "gif0",
      "mtu": 1280,
      "hardwareAddr": "",
      "flags": [
        "pointtopoint",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "stf0",
      "mtu": 1280,
      "hardwareAddr": "",
      "flags": [
        "0"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "en5",
      "mtu": 1500,
      "hardwareAddr": "ac:de:48:00:11:22",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [
        "fe80::aede:48ff:fe00:1122/64"
      ],
      "multicastAddrs": [
        "ff01::1",
        "ff02::2:ff02:66ae",
        "ff02::1",
        "ff02::1:ff00:1122",
        "ff02::fb"
      ]
    },
    {
      "name": "ap1",
      "mtu": 1500,
      "hardwareAddr": "36:7d:da:84:00:3e",
      "flags": [
        "broadcast",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "en0",
      "mtu": 1500,
      "hardwareAddr": "14:7d:da:84:00:3e",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [
        "192.168.101.5/24"
      ],
      "multicastAddrs": [
        "224.0.0.1",
        "224.0.0.251",
        "ff02::fb"
      ]
    },
    {
      "name": "p2p0",
      "mtu": 2304,
      "hardwareAddr": "06:7d:da:84:00:3e",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "awdl0",
      "mtu": 1484,
      "hardwareAddr": "f2:26:9f:40:0f:0d",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [
        "fe80::f026:9fff:fe40:f0d/64"
      ],
      "multicastAddrs": [
        "ff01::1",
        "ff02::1",
        "ff02::2:ff02:66ae",
        "ff02::1:ff40:f0d",
        "ff02::fb"
      ]
    },
    {
      "name": "llw0",
      "mtu": 1500,
      "hardwareAddr": "f2:26:9f:40:0f:0d",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [
        "fe80::f026:9fff:fe40:f0d/64"
      ],
      "multicastAddrs": [
        "ff01::1",
        "ff02::1",
        "ff02::2:ff02:66ae",
        "ff02::1:ff40:f0d",
        "ff02::fb"
      ]
    },
    {
      "name": "en1",
      "mtu": 1500,
      "hardwareAddr": "ee:49:b4:08:04:04",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "en2",
      "mtu": 1500,
      "hardwareAddr": "ee:49:b4:08:04:05",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "en3",
      "mtu": 1500,
      "hardwareAddr": "ee:49:b4:08:04:01",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "en4",
      "mtu": 1500,
      "hardwareAddr": "ee:49:b4:08:04:00",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "bridge0",
      "mtu": 1500,
      "hardwareAddr": "ee:49:b4:08:04:04",
      "flags": [
        "up",
        "broadcast",
        "multicast"
      ],
      "addrs": [],
      "multicastAddrs": []
    },
    {
      "name": "utun0",
      "mtu": 1380,
      "hardwareAddr": "",
      "flags": [
        "up",
        "pointtopoint",
        "multicast"
      ],
      "addrs": [
        "fe80::5c4e:5009:f1c1:ed7d/64"
      ],
      "multicastAddrs": [
        "ff02::fb",
        "ff01::1",
        "ff02::2:ff02:66ae",
        "ff02::1",
        "ff02::1:ffc1:ed7d"
      ]
    },
    {
      "name": "utun1",
      "mtu": 2000,
      "hardwareAddr": "",
      "flags": [
        "up",
        "pointtopoint",
        "multicast"
      ],
      "addrs": [
        "fe80::b58c:330e:b1be:2785/64"
      ],
      "multicastAddrs": [
        "ff02::fb",
        "ff01::1",
        "ff02::2:ff02:66ae",
        "ff02::1",
        "ff02::1:ffbe:2785"
      ]
    },
    {
      "name": "utun2",
      "mtu": 1500,
      "hardwareAddr": "",
      "flags": [
        "up",
        "pointtopoint",
        "multicast"
      ],
      "addrs": [
        "10.8.0.2/24"
      ],
      "multicastAddrs": [
        "224.0.0.1"
      ]
    }
  ]
}
```

#### OsInfo
Operating system information for a running application.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| Os | OS name. | based on system | os |
| Arch | Architecture of OS. | based on system | arch |
| Hostname | Hostname of OS. | based on system | hostname |

##### Access OsInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewOsInfo()))
```

```json
{
  "os": "darwin",
  "arch": "amd64",
  "hostname": "rk-dev.local"
}
```

#### GoEnvInfo
Go environment information for a running application.
##### Fields
| Element | Description | Default | JSON Key |
| ------ | ------ | ------ | ------ |
| GOOS | OS name. | based on system | goos |
| GOArch | Architecture of OS. | based on system | goArch |
| StartTime | Represent start time of go process. | based on system | startTime |
| UpTimeSec | Represent up time of go process in seconds. | based on system | upTimeSec |
| UpTimeStr | Represent up time of go process as human readable string. | based on system | upTimeStr |
| RoutinesCount | Number of go routines. | based on system | routinesCount |
| Version | Version of GO. | based on system | version |

##### Access OsInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewGoEnvInfo()))
```

```json
{
  "goos": "darwin",
  "goArch": "amd64",
  "startTime": "2021-05-14T03:18:04+08:00",
  "upTimeSec": 0,
  "upTimeStr": "3 milliseconds",
  "routinesCount": 2,
  "version": "go1.15.5"
}
```

#### PromMetricsInfo
Request metrics to struct from prometheus summary collector.
##### Fields
| Element | Description | Default | JSON key |
| ------ | ------ | ------ | ------ |
| RestPath | Restful API path | Based on user  | restPath |
| RestMethod | Restful API method | Based on user  | restMethod |
| GrpcService | Grpc service | Based on user  | grpcService |
| GrpcMethod | Grpc method | Based on user  | grpcMethod |
| ElapsedNanoP50 | Quantile of p50 with time elapsed | based on prometheus collector | elapsedNanoP50 |
| ElapsedNanoP90 | Quantile of p90 with time elapsed | based on prometheus collector | elapsedNanoP90 |
| ElapsedNanoP99 | Quantile of p99 with time elapsed | based on prometheus collector | elapsedNanoP99 |
| ElapsedNanoP999 | Quantile of p999 with time elapsed | based on prometheus collector | elapsedNanoP999 |
| Count | Total number of requests | based on prometheus collector | count |
| ResCode | Response code labels | based on prometheus collector | resCode |

##### Access PromMetricsInfo
```go
fmt.Println(rkcommon.ConvertStructToJSONPretty(rkentry.NewPromMetricsInfo(<your prometheus summary vector>)))
```

- Response from rk-gin
```json
{
  "metrics": [
    {
      "restPath": "/sw/",
      "restMethod": "GET",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 7789178,
      "elapsedNanoP90": 7789178,
      "elapsedNanoP99": 7789178,
      "elapsedNanoP999": 7789178,
      "count": 1,
      "resCode": [
        {
          "resCode": "200",
          "count": 1
        }
      ]
    },
    {
      "restPath": "/rk/v1/apis",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/configs",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/certs",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/gc",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/info",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/sys",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/req",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/entries",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/rk/v1/logs",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/metrics",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    },
    {
      "restPath": "/  /healthy",
      "restMethod": "",
      "grpcService": "",
      "grpcMethod": "",
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "count": 0,
      "resCode": []
    }
  ]
}
```

- Response from rk-grpc
```json
{
  "metrics": [
    {
      "count": 1,
      "elapsedNanoP50": 331538,
      "elapsedNanoP90": 331538,
      "elapsedNanoP99": 331538,
      "elapsedNanoP999": 331538,
      "grpcMethod": "Req",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [
        {
          "count": 1,
          "resCode": "OK"
        }
      ],
      "restMethod": "GET",
      "restPath": "/rk/v1/req"
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Healthy",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Apis",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Logs",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Gc",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Info",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Configs",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Sys",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Entries",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    },
    {
      "count": 0,
      "elapsedNanoP50": 0,
      "elapsedNanoP90": 0,
      "elapsedNanoP99": 0,
      "elapsedNanoP999": 0,
      "grpcMethod": "Certs",
      "grpcService": "rk.api.v1.RkCommonService",
      "resCode": [],
      "restMethod": "",
      "restPath": ""
    }
  ]
}
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