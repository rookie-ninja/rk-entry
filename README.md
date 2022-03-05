# rk-entry
[![build](https://github.com/rookie-ninja/rk-entry/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-entry/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/rookie-ninja/rk-entry/branch/master/graph/badge.svg?token=KGKHKIWOEQ)](https://codecov.io/gh/rookie-ninja/rk-entry)
[![Go Report Card](https://goreportcard.com/badge/github.com/rookie-ninja/rk-entry)](https://goreportcard.com/report/github.com/rookie-ninja/rk-entry)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The entry library mainly used by rk-boot.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Notice of V2!](#notice-of-v2)
- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Entry](#entry)
  - [Interact with rk-boot.Bootstrapper?](#interact-with-rk-bootbootstrapper)
  - [GlobalAppCtx](#globalappctx)
  - [Builtin Entry](#builtin-entry)
- [How to use?](#how-to-use)
- [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Notice of V2!
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

### Builtin Entry
| Name                   | Description                                                                                      |
|------------------------|--------------------------------------------------------------------------------------------------|
| AppInfoEntry           | Contains application, process, OS and go env information                                         |
| CertEntry              | Store certificates                                                                               |
| CommonServiceEntry     | Provide commonly used API                                                                        |
| ConfigEntry            | Init [viper](https://github.com/spf13/viper) instance                                            |
| EventEntry             | Init [rk-query](https://github.com/rookie-ninja/rk-query) instance which is used for logging RPC |
| LoggerEntry            | Init [zap](https://github.com/uber-go/zap) instance for logging                                  |
| PromEntry              | Init [prometheus](github.com/prometheus/client_model) client instance                            |
| StaticFileHandlerEntry | Init static web site UI handler                                                                  |
| SWEntry                | Init swagger UI                                                                                  |
| DocsEntry              | Init [RapiDoc](https://github.com/mrin9/RapiDoc) which can be replace Swagger and RK TV          |

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
