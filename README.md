# rk-entry
[![build](https://github.com/rookie-ninja/rk-entry/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-entry/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/rookie-ninja/rk-entry/branch/master/graph/badge.svg?token=KGKHKIWOEQ)](https://codecov.io/gh/rookie-ninja/rk-entry)
[![Go Report Card](https://goreportcard.com/badge/github.com/rookie-ninja/rk-entry)](https://goreportcard.com/report/github.com/rookie-ninja/rk-entry)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Base library for RK family, includes **Entry interface**, **Builtin Entries** and **Web Framework middleware base**.

| Name                   | Description                                                                              |
|------------------------|------------------------------------------------------------------------------------------|
| Entry                  | Interface                                                                                |
| AppInfoEntry           | Builtin entry, collect application, process info                                         |
| CertEntry              | Builtin entry, parse certificates from path                                              |
| ConfigEntry            | Builtin entry, parse config file from path                                               |
| EventEntry             | Builtin entry, create event logger for recording RPC calls                               |
| LoggerEntry            | Builtin entry, create logger instance                                                    |
| CommonServiceEntry     | Builtin entry, provide http handler of commonly used API, used for web framework entry   |
| DocsEntry              | Builtin entry, provide http handler of rapiDoc UI, used for web framework entry          |
| PromEntry              | Builtin entry, provide http handler of prometheus client, used for web framework entry   |
| StaticFileHandlerEntry | Builtin entry, provide http handler of static file handler, used for web framework entry |
| SWEntry                | Builtin entry, provide http handler of swagger UI, used for web framework entry          |
| auth                   | Middleware base for auth                                                                 |
| cors                   | Middleware base for cors                                                                 |
| csrf                   | Middleware base for csrf                                                                 |
| jwt                    | Middleware base for jwt                                                                  |
| log                    | Middleware base for log                                                                  |
| meta                   | Middleware base for meta                                                                 |
| panic                  | Middleware base for panic                                                                |
| prom                   | Middleware base for prom                                                                 |
| ratelimit              | Middleware base for ratelimit                                                            |
| secure                 | Middleware base for secure                                                               |
| timeout                | Middleware base for timeout                                                              |
| tracing                | Middleware base for tracing                                                              |

## Notice of V2!
> rookie-ninja/rk-entry is under big refactoring stage.
> 
> New version will be start with v2.x.x and may not compatible with v1.x.x.
> 
> Please do not upgrade to the newest master branch.

## Installation
```bash
go get github.com/rookie-ninja/rk-entry/v2
```

## Quick Start
### Entry
**rkentry.Entry** is an interface which can be started from YAML config file by calling UnmarshalBoot() function

Users can implement **rkentry.Entry** interface and bootstrap any service/process with **rkboot.Bootstrapper**

[Example](example)

### Interact with rk-boot.Bootstrapper?

1: Entry will be created and registered into rkentry.GlobalAppCtx.

2: rkboot.Bootstrap() function will iterator all entries in rkentry.GlobalAppCtx.Entries and call Bootstrap().

3: Application will wait for shutdown signal via rkentry.GlobalAppCtx.ShutdownSig.

4: rkboot.Interrupt() function will iterate all entries in rkentry.GlobalAppCtx.Entries and call Interrupt().

### GlobalAppCtx
A struct called AppContext witch contains RK style application metadata.

| Element       | Description                                                                                       | JSON            | Default values                                                                    |
|---------------|---------------------------------------------------------------------------------------------------|-----------------|-----------------------------------------------------------------------------------|
| startTime     | Application start time.                                                                           | startTime       | 0001-01-01 00:00:00 +0000 UTC                                                     |
| appInfoEntry  | See ApplicationInfoEntry for detail.                                                              | appInfoEntry    | Includes application info specified by user.                                      |
| entries       | User implemented Entry.                                                                           | externalEntries | Includes user implemented Entry configuration initiated by user.                  |
| userValues    | User K/V registered from code.                                                                    | userValues      | empty map                                                                         |
| shutdownSig   | Shutdown signals which includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT. | shutdown_sig    | channel includes syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT |
| shutdownHooks | Shutdown hooks registered from user code.                                                         | shutdown_hooks  | empty list                                                                        |

## How to use?
rk-entry should be used as base package for applications which hope to start with YAML.

Please refer [rk-gin](https://github.com/rookie-ninja/rk-gin) as example.

## Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
lark@rkdev.info.

Released under the [Apache 2.0 License](LICENSE).
