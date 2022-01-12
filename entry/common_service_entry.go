// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"runtime"
)

const (
	// CommonServiceEntryType type of entry
	CommonServiceEntryType = "CommonServiceEntry"
	// CommonServiceEntryNameDefault name of entry
	CommonServiceEntryNameDefault = "CommonServiceDefault"
	// CommonServiceEntryDescription description of entry
	CommonServiceEntryDescription = "Internal RK entry which implements commonly used API."
)

// @title RK Common Service
// @version 1.0
// @description This is builtin RK common service.

// @contact.name rk-dev
// @contact.url https://github.com/rookie-ninja/rk-entry
// @contact.email lark@pointgoal.io

// @license.name Apache 2.0 License
// @license.url https://github.com/rookie-ninja/rk-entry/blob/master/LICENSE.txt

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// @securityDefinitions.apikey JWT
// @in header
// @name Authorization

// @schemes http https

// BootConfigCommonService Bootstrap config of common service.
// 1: Enabled: Enable common service.
type BootConfigCommonService struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// CommonServiceEntry RK common service which contains commonly used APIs
// 1: Healthy Returns true if process is alive
// 2: Gc Trigger gc()
// 3: Info Returns entry basic information
// 4: Configs Returns viper configs in GlobalAppCtx
// 5: Apis Returns list of apis registered in gin router
// 6: Sys Returns CPU and Memory information
// 7: Req Returns request metrics
// 8: Certs Returns certificates
// 9: Entries Returns entries
// 10: Logs Returns log entries
// 12: Deps Returns dependency which is full  go.mod file content
// 13: License Returns license file content
// 14: Readme Returns README file content
type CommonServiceEntry struct {
	EntryName        string            `json:"entryName" yaml:"entryName"`
	EntryType        string            `json:"entryType" yaml:"entryType"`
	EntryDescription string            `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry `json:"-" yaml:"-"`
	ZapLoggerEntry   *ZapLoggerEntry   `json:"-" yaml:"-"`
	HealthyPath      string            `json:"-" yaml:"-"`
	GcPath           string            `json:"-" yaml:"-"`
	InfoPath         string            `json:"-" yaml:"-"`
	ConfigsPath      string            `json:"-" yaml:"-"`
	SysPath          string            `json:"-" yaml:"-"`
	EntriesPath      string            `json:"-" yaml:"-"`
	CertsPath        string            `json:"-" yaml:"-"`
	LogsPath         string            `json:"-" yaml:"-"`
	DepsPath         string            `json:"-" yaml:"-"`
	LicensePath      string            `json:"-" yaml:"-"`
	ReadmePath       string            `json:"-" yaml:"-"`
	GitPath          string            `json:"-" yaml:"-"`
	ApisPath         string            `json:"-" yaml:"-"`
	ReqPath          string            `json:"-" yaml:"-"`
}

// CommonServiceEntryOption Common service entry option.
type CommonServiceEntryOption func(*CommonServiceEntry)

// WithNameCommonService Provide name.
func WithNameCommonService(name string) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.EntryName = name
	}
}

// WithEventLoggerEntryCommonService Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryCommonService(eventLoggerEntry *EventLoggerEntry) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// WithZapLoggerEntryCommonService Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryCommonService(zapLoggerEntry *ZapLoggerEntry) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// RegisterCommonServiceEntryWithConfig Create new common service entry with config
func RegisterCommonServiceEntryWithConfig(config *BootConfigCommonService, name string, zap *ZapLoggerEntry, event *EventLoggerEntry) *CommonServiceEntry {
	var commonServiceEntry *CommonServiceEntry

	if config.Enabled {
		commonServiceEntry = RegisterCommonServiceEntry(
			WithNameCommonService(name),
			WithZapLoggerEntryCommonService(zap),
			WithEventLoggerEntryCommonService(event))
	}

	return commonServiceEntry
}

// RegisterCommonServiceEntry Create new common service entry with options.
func RegisterCommonServiceEntry(opts ...CommonServiceEntryOption) *CommonServiceEntry {
	entry := &CommonServiceEntry{
		EntryName:        CommonServiceEntryNameDefault,
		EntryType:        CommonServiceEntryType,
		EntryDescription: CommonServiceEntryDescription,
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		HealthyPath:      "/rk/v1/healthy",
		GcPath:           "/rk/v1/gc",
		InfoPath:         "/rk/v1/info",
		ConfigsPath:      "/rk/v1/configs",
		SysPath:          "/rk/v1/sys",
		EntriesPath:      "/rk/v1/entries",
		CertsPath:        "/rk/v1/certs",
		LogsPath:         "/rk/v1/logs",
		DepsPath:         "/rk/v1/deps",
		LicensePath:      "/rk/v1/license",
		ReadmePath:       "/rk/v1/readme",
		GitPath:          "/rk/v1/git",
		ApisPath:         "/rk/v1/apis",
		ReqPath:          "/rk/v1/req",
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = CommonServiceEntryNameDefault
	}

	return entry
}

// Bootstrap common service entry.
func (entry *CommonServiceEntry) Bootstrap(context.Context) {
	// Noop
}

// Interrupt common service entry.
func (entry *CommonServiceEntry) Interrupt(context.Context) {
	// Noop
}

// GetName Get name of entry.
func (entry *CommonServiceEntry) GetName() string {
	return entry.EntryName
}

// GetType Get entry type.
func (entry *CommonServiceEntry) GetType() string {
	return entry.EntryType
}

// String Stringfy entry.
func (entry *CommonServiceEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// GetDescription Get description of entry.
func (entry *CommonServiceEntry) GetDescription() string {
	return entry.EntryDescription
}

// MarshalJSON Marshal entry.
func (entry *CommonServiceEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"zapLoggerEntry":   entry.ZapLoggerEntry.GetName(),
		"eventLoggerEntry": entry.EventLoggerEntry.GetName(),
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *CommonServiceEntry) UnmarshalJSON([]byte) error {
	return nil
}

func doHealthy() *HealthyResponse {
	return &HealthyResponse{
		Healthy: true,
	}
}

// Healthy handler
// @Summary Get application healthy status
// @Id 1
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} HealthyResponse
// @Router /rk/v1/healthy [get]
func (entry *CommonServiceEntry) Healthy(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doHealthy())
	writer.Write(bytes)
}

func doGc() *GcResponse {
	before := NewMemInfo()
	runtime.GC()
	after := NewMemInfo()

	return &GcResponse{
		MemStatBeforeGc: before,
		MemStatAfterGc:  after,
	}
}

// Gc handler
// @Summary Trigger Gc
// @Id 2
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} GcResponse
// @Router /rk/v1/gc [get]
func (entry *CommonServiceEntry) Gc(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doGc())
	writer.Write(bytes)
}

func doInfo() *ProcessInfo {
	return NewProcessInfo()
}

// Info handler
// @Summary Get application and process info
// @Id 3
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} ProcessInfo
// @Router /rk/v1/info [get]
func (entry *CommonServiceEntry) Info(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doInfo())
	writer.Write(bytes)
}

func doConfigs() *ConfigsResponse {
	res := &ConfigsResponse{
		Entries: make([]*ConfigsResponse_ConfigEntry, 0),
	}

	for _, v := range GlobalAppCtx.ListConfigEntries() {
		configEntry := &ConfigsResponse_ConfigEntry{
			EntryName:        v.GetName(),
			EntryType:        v.GetType(),
			EntryDescription: v.GetDescription(),
			EntryMeta:        v.GetViperAsMap(),
			Path:             v.Path,
		}

		res.Entries = append(res.Entries, configEntry)
	}

	return res
}

// Configs handler
// @Summary List ConfigEntry
// @Id 4
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} ConfigsResponse
// @Router /rk/v1/configs [get]
func (entry *CommonServiceEntry) Configs(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doConfigs())
	writer.Write(bytes)
}

func doSys() *SysResponse {
	return &SysResponse{
		CpuInfo:   NewCpuInfo(),
		MemInfo:   NewMemInfo(),
		NetInfo:   NewNetInfo(),
		OsInfo:    NewOsInfo(),
		GoEnvInfo: NewGoEnvInfo(),
	}
}

// Sys handler
// @Summary Get OS Stat
// @Id 5
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} SysResponse
// @Router /rk/v1/sys [get]
func (entry *CommonServiceEntry) Sys(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doSys())
	writer.Write(bytes)
}

func doEntries() *EntriesResponse {
	res := &EntriesResponse{
		Entries: make(map[string][]*EntriesResponse_Entry),
	}

	// Iterate all internal and external entries in GlobalAppCtx
	entriesHelper(GlobalAppCtx.ListEntries(), res)
	entriesHelper(GlobalAppCtx.ListEventLoggerEntriesRaw(), res)
	entriesHelper(GlobalAppCtx.ListZapLoggerEntriesRaw(), res)
	entriesHelper(GlobalAppCtx.ListConfigEntriesRaw(), res)
	entriesHelper(GlobalAppCtx.ListCertEntriesRaw(), res)
	entriesHelper(GlobalAppCtx.ListCredEntriesRaw(), res)

	// App info entry
	appInfoEntry := GlobalAppCtx.GetAppInfoEntry()
	res.Entries[appInfoEntry.GetType()] = []*EntriesResponse_Entry{
		{
			EntryName:        appInfoEntry.GetName(),
			EntryType:        appInfoEntry.GetType(),
			EntryDescription: appInfoEntry.GetDescription(),
			EntryMeta:        appInfoEntry,
		},
	}

	return res
}

// Helper function of /entries
func entriesHelper(m map[string]Entry, res *EntriesResponse) {
	// Iterate entries and construct EntryElement
	for i := range m {
		entry := m[i]
		element := &EntriesResponse_Entry{
			EntryName:        entry.GetName(),
			EntryType:        entry.GetType(),
			EntryDescription: entry.GetDescription(),
			EntryMeta:        entry,
		}

		if entries, ok := res.Entries[entry.GetType()]; ok {
			entries = append(entries, element)
		} else {
			res.Entries[entry.GetType()] = []*EntriesResponse_Entry{element}
		}
	}
}

// Entries handler
// @Summary List all Entry
// @Id 6
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} EntriesResponse
// @Router /rk/v1/entries [get]
func (entry *CommonServiceEntry) Entries(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doEntries())
	writer.Write(bytes)
}

func doCerts() *CertsResponse {
	res := &CertsResponse{
		Entries: make([]*CertsResponse_Entry, 0),
	}

	entries := GlobalAppCtx.ListCertEntries()

	// Iterator cert entries and construct CertResponse
	for i := range entries {
		entry := entries[i]

		certEntry := &CertsResponse_Entry{
			EntryName:        entry.GetName(),
			EntryType:        entry.GetType(),
			EntryDescription: entry.GetDescription(),
		}

		if entry.Retriever != nil {
			certEntry.Endpoint = entry.Retriever.GetEndpoint()
			certEntry.Locale = entry.Retriever.GetLocale()
			certEntry.Provider = entry.Retriever.GetProvider()
			certEntry.ServerCertPath = entry.ServerCertPath
			certEntry.ServerKeyPath = entry.ServerKeyPath
			certEntry.ClientCertPath = entry.ClientCertPath
			certEntry.ClientKeyPath = entry.ClientKeyPath
		}

		if entry.Store != nil {
			certEntry.ServerCert = entry.Store.SeverCertString()
			certEntry.ClientCert = entry.Store.ClientCertString()
		}

		res.Entries = append(res.Entries, certEntry)
	}

	return res
}

// Certs handler
// @Summary List CertEntry
// @Id 7
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} CertsResponse
// @Router /rk/v1/certs [get]
func (entry *CommonServiceEntry) Certs(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doCerts())
	writer.Write(bytes)
}

func doLogs() *LogsResponse {
	res := &LogsResponse{
		Entries: make(map[string][]*LogsResponse_Entry),
	}

	logsHelper(GlobalAppCtx.ListEventLoggerEntriesRaw(), res)
	logsHelper(GlobalAppCtx.ListZapLoggerEntriesRaw(), res)

	return res
}

// Helper function of /logs
func logsHelper(m map[string]Entry, res *LogsResponse) {
	entries := make([]*LogsResponse_Entry, 0)

	// Iterate logger related entries and construct LogEntryElement
	for i := range m {
		entry := m[i]
		logEntry := &LogsResponse_Entry{
			EntryName:        entry.GetName(),
			EntryType:        entry.GetType(),
			EntryDescription: entry.GetDescription(),
			EntryMeta:        entry,
		}

		if val, ok := entry.(*ZapLoggerEntry); ok {
			if val.LoggerConfig != nil {
				logEntry.OutputPaths = val.LoggerConfig.OutputPaths
				logEntry.ErrorOutputPaths = val.LoggerConfig.ErrorOutputPaths
			}
		}

		if val, ok := entry.(*EventLoggerEntry); ok {
			if val.LoggerConfig != nil {
				logEntry.OutputPaths = val.LoggerConfig.OutputPaths
				logEntry.ErrorOutputPaths = val.LoggerConfig.ErrorOutputPaths
			}
		}

		entries = append(entries, logEntry)
	}

	var entryType string

	if len(entries) > 0 {
		entryType = entries[0].EntryType
	}

	res.Entries[entryType] = entries
}

// Logs handler
// @Summary List logger related entries
// @Id 8
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} LogsResponse
// @Router /rk/v1/logs [get]
func (entry *CommonServiceEntry) Logs(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doLogs())
	writer.Write(bytes)
}

func doDeps() *DepResponse {
	return &DepResponse{
		GoMod: GlobalAppCtx.GetAppInfoEntry().GoMod,
	}
}

// Deps handler
// @Summary List dependencies related application
// @Id 9
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} DepResponse
// @Router /rk/v1/deps [get]
func (entry *CommonServiceEntry) Deps(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doDeps())
	writer.Write(bytes)
}

func doLicense() *LicenseResponse {
	return &LicenseResponse{
		License: GlobalAppCtx.GetAppInfoEntry().License,
	}
}

// License handler
// @Summary Get license related application
// @Id 10
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} LicenseResponse
// @Router /rk/v1/license [get]
func (entry *CommonServiceEntry) License(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doLicense())
	writer.Write(bytes)
}

func doReadme() *ReadmeResponse {
	return &ReadmeResponse{
		Readme: GlobalAppCtx.GetAppInfoEntry().Readme,
	}
}

// Readme handler
// @Summary Get README file.
// @Id 11
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} ReadmeResponse
// @Router /rk/v1/readme [get]
func (entry *CommonServiceEntry) Readme(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doReadme())
	writer.Write(bytes)
}

func doGit() *GitResponse {
	res := &GitResponse{}

	rkMetaEntry := GlobalAppCtx.GetRkMetaEntry()
	if rkMetaEntry != nil {
		res.Package = path.Base(rkMetaEntry.RkMeta.Git.Url)
		res.Branch = rkMetaEntry.RkMeta.Git.Branch
		res.Tag = rkMetaEntry.RkMeta.Git.Tag
		res.Url = rkMetaEntry.RkMeta.Git.Url
		res.CommitId = rkMetaEntry.RkMeta.Git.Commit.Id
		res.CommitIdAbbr = rkMetaEntry.RkMeta.Git.Commit.IdAbbr
		res.CommitSub = rkMetaEntry.RkMeta.Git.Commit.Sub
		res.CommitterName = rkMetaEntry.RkMeta.Git.Commit.Committer.Name
		res.CommitterEmail = rkMetaEntry.RkMeta.Git.Commit.Committer.Email
		res.CommitDate = rkMetaEntry.RkMeta.Git.Commit.Date
	}

	return res
}

// Git handler
// @Summary Get Git information.
// @Id 12
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} GitResponse
// @Router /rk/v1/git [get]
func (entry *CommonServiceEntry) Git(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(doGit())
	writer.Write(bytes)
}

// Apis handler
// @Summary List API
// @Id 13
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} ApisResponse
// @Router /rk/v1/apis [get]
func (entry *CommonServiceEntry) apisNoop(writer http.ResponseWriter, request *http.Request) {}

// Req handler
// @Summary List prometheus metrics of requests
// @Id 14
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @success 200 {object} ReqResponse
// @Router /rk/v1/req [get]
func (entry *CommonServiceEntry) reqNoop(writer http.ResponseWriter, request *http.Request) {}
