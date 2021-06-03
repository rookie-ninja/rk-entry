// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

// Response of /healthy
type HealthyResponse struct {
	Healthy bool `json:"healthy" yaml:"healthy"`
}

// Response of /gc
// Returns memory stats of GC before and after.
type GcResponse struct {
	MemStatBeforeGc *MemInfo `json:"memStatBeforeGc" yaml:"memStatBeforeGc"`
	MemStatAfterGc  *MemInfo `json:"memStatAfterGc" yaml:"memStatAfterGc"`
}

// Entry for ConfigsResponse
type ConfigsResponse_ConfigEntry struct {
	EntryName        string                 `json:"entryName" yaml:"entryName"`
	EntryType        string                 `json:"entryType" yaml:"entryType"`
	EntryDescription string                 `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        map[string]interface{} `json:"entryMeta" yaml:"entryMeta"`
	Path             string                 `json:"path" yaml:"path"`
}

// Response of /configs
type ConfigsResponse struct {
	Entries []*ConfigsResponse_ConfigEntry `json:"entries" yaml:"entries"`
}

// Response for path of /apis
type ApisResponse struct {
	Entries []*ApisResponse_Entry `json:"entries" yaml:"entries"`
}

type ApisResponse_Entry struct {
	EntryName string             `json:"entryName" yaml:"entryName"`
	Grpc      *ApisResponse_Grpc `json:"grpc" yaml:"grpc"`
	Rest      *ApisResponse_Rest `json:"rest" yaml:"rest"`
}

type ApisResponse_Grpc struct {
	Service string             `json:"service" yaml:"service"`
	Method  string             `json:"method" yaml:"method"`
	Type    string             `json:"type" yaml:"type"`
	Port    uint64             `json:"port" yaml:"port"`
	Gw      *ApisResponse_Rest `json:"gw" yaml:"gw"`
}

type ApisResponse_Rest struct {
	Port    uint64 `json:"port" yaml:"port"`
	Pattern string `json:"pattern" yaml:"pattern"`
	Method  string `json:"method" yaml:"method"`
	SwUrl   string `json:"swUrl" yaml:"swUrl"`
}

// Response of /sys
type SysResponse struct {
	CpuInfo   *CpuInfo   `json:"cpuInfo" yaml:"cpuInfo"`
	MemInfo   *MemInfo   `json:"memInfo" yaml:"memInfo"`
	NetInfo   *NetInfo   `json:"netInfo" yaml:"netInfo"`
	OsInfo    *OsInfo    `json:"osInfo" yaml:"osInfo"`
	GoEnvInfo *GoEnvInfo `json:"goEnvInfo" yaml:"goEnvInfo"`
}

// Response of /entries
type EntriesResponse struct {
	Entries map[string][]*EntriesResponse_Entry `json:"entries" yaml:"entries"`
}

// Entry element which specifies name, type and description.
type EntriesResponse_Entry struct {
	EntryName        string `json:"entryName" yaml:"entryName"`
	EntryType        string `json:"entryType" yaml:"entryType"`
	EntryDescription string `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        Entry  `json:"entryMeta" yaml:"entryMeta"`
}

// Response of /certs
type CertsResponse struct {
	Entries []*CertsResponse_Entry `json:"entries" yaml:"entries"`
}

type CertsResponse_Entry struct {
	EntryName        string `json:"entryName" yaml:"entryName"`
	EntryType        string `json:"entryType" yaml:"entryType"`
	EntryDescription string `json:"entryDescription" yaml:"entryDescription"`
	ServerCertPath   string `json:"serverCertPath" yaml:"serverCertPath"`
	ServerKeyPath    string `json:"serverKeyPath" yaml:"serverKeyPath"`
	ClientCertPath   string `json:"clientCertPath" yaml:"clientCertPath"`
	ClientKeyPath    string `json:"clientKeyPath" yaml:"clientKeyPath"`
	Endpoint         string `json:"endpoint" yaml:"endpoint"`
	Locale           string `json:"locale" yaml:"locale"`
	Provider         string `json:"provider" yaml:"provider"`
	ServerCert       string `json:"serverCert" yaml:"serverCert"`
	ClientCert       string `json:"clientCert" yaml:"clientCert"`
}

// Response of /logs.
type LogsResponse struct {
	Entries map[string][]*LogsResponse_Entry `json:"entries" yaml:"entries"`
}

// Entry element which specifies name, type. description, output path and error output path.
type LogsResponse_Entry struct {
	EntryName        string   `json:"entryName" yaml:"entryName"`
	EntryType        string   `json:"entryType" yaml:"entryType"`
	EntryDescription string   `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        Entry    `json:"entryMeta" yaml:"entryMeta"`
	OutputPaths      []string `json:"outputPaths" yaml:"outputPaths"`
	ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`
}

// Response of /req
type ReqResponse struct {
	Metrics []*ReqMetricsRK `json:"metrics" yaml:"metrics"`
}

// Response of /dep
type DepResponse struct {
	GoMod string `json:"goMod" yaml:"goMod"`
}

// Response of /license
type LicenseResponse struct {
	License string `json:"license" yaml:"license"`
}

// Response of /readme
type ReadmeResponse struct {
	Readme string `json:"readme" yaml:"readme"`
}
