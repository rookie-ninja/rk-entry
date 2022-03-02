# rk-entry
[![build](https://github.com/rookie-ninja/rk-entry/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-entry/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/rookie-ninja/rk-entry/branch/master/graph/badge.svg?token=KGKHKIWOEQ)](https://codecov.io/gh/rookie-ninja/rk-entry)
[![Go Report Card](https://goreportcard.com/badge/github.com/rookie-ninja/rk-entry)](https://goreportcard.com/report/github.com/rookie-ninja/rk-entry)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The entry library mainly used by rk-boot.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Important!](#important)
  - [Installation](#installation)
  - [Quick Start](#quick-start)
    - [Entry](#entry)
      - [Interact with rk-boot.Bootstrapper?](#interact-with-rk-bootbootstrapper)
    - [AppCtx](#appctx)
      - [Access AppCtx](#access-appctx)
      - [Usage of AppCtx](#usage-of-appctx)
  - [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Important!
> rookie-ninja/rk-entry is under big refactoring stage.
> 
> New version will be start with v2.x.x and may not compatible with v1.x.x.
> 
> Please do not upgrade to the newest master branch.

## Installation
```bash
go get github.com/rookie-ninja/rk-entry
```

## Quick Start
### Entry
**rkentry.Entry** is an interface for rkboot.Bootstrapper to bootstrap entry.

Users can implement **rkentry.Entry** interface and bootstrap any service/process with **rkb.Bootstrapper**

[Example](example)

#### Interact with rk-boot.Bootstrapper?

1: Entry will be created and registered into rke.AppCtx.

2: rkb.Bootstrap() function will iterator all entries in rke.AppCtx.Entries and call Bootstrap().

3: Application will wait for shutdown signal via rke.AppCtx.ShutdownSig.

4: rkb.Interrupt() function will iterate all entries in rke.AppCtx.Entries and call Interrupt().

### AppCtx
A struct called AppContext witch contains RK style application metadata.

#### Access AppCtx

Access it via AppCtx variable 
```go
AppCtx
```
**Fields in rke.AppCtx**

| Element       | Description                                                                                       | JSON            | Default values                                                                    |
|---------------|---------------------------------------------------------------------------------------------------|-----------------|-----------------------------------------------------------------------------------|
| startTime     | Application start time.                                                                           | startTime       | 0001-01-01 00:00:00 +0000 UTC                                                     |
| appInfoEntry  | See ApplicationInfoEntry for detail.                                                              | appInfoEntry    | Includes application info specified by user.                                      |
| entries       | User implemented Entry.                                                                           | externalEntries | Includes user implemented Entry configuration initiated by user.                  |
| userValues    | User K/V registered from code.                                                                    | userValues      | empty map                                                                         |
| shutdownSig   | Shutdown signals which includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT. | shutdown_sig    | channel includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT |
| shutdownHooks | Shutdown hooks registered from user code.                                                         | shutdown_hooks  | empty list                                                                        |


#### Usage of AppCtx
- Access start time and uptime of application.
```go
// Get start time of application
startTime := rkentry.GlobalAppCtx.GetStartTime()

// Get uptime of application/process.
upTime := rkentry.GlobalAppCtx.GetUpTime()
```

- Access AppInfoEntry
```go
// Get AppInfoEntry from rkentry.GlobalAppCtx
appInfoEntry := rkentry.GlobalAppCtx.GetAppInfoEntry()
```

- Access entries
Entries contains user defined entry.
```go
// Access entries from rkentry.GlobalAppCtx
entry := rkentry.GlobalAppCtx.GetEntry("entry type", "entry name")

// Access entries via utility function
entries := rkentry.GlobalAppCtx.ListEntries()

// Add entry
rkentry.GlobalAppCtx.AddEntry()
```

- Access Values

User can add/get/list/remove any values into map of rke.AppCtx.UserValues as needed.

rke.AppCtx don't provide any locking mechanism.
```go
// Add k/v value into AppCtx, key should be string and value could be any kind
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

Users can add their own shutdown hook function into rke.AppCtx.

rkb will iterate all shutdown hooks in rke.AppCtx and call every shutdown hook function.
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

## Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
lark@rkdev.info.

Released under the [Apache 2.0 License](LICENSE).
