// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

// HealthyResponse response of /healthy
type HealthyResponse struct {
	Healthy bool `json:"healthy" yaml:"healthy"`
}

// GcResponse response of /gc
// Returns memory stats of GC before and after.
type GcResponse struct {
	MemStatBeforeGc *MemInfo `json:"memStatBeforeGc" yaml:"memStatBeforeGc"`
	MemStatAfterGc  *MemInfo `json:"memStatAfterGc" yaml:"memStatAfterGc"`
}

// ConfigsResponse_ConfigEntry Entry for ConfigsResponse
type ConfigsResponse_ConfigEntry struct {
	EntryName        string                 `json:"entryName" yaml:"entryName"`
	EntryType        string                 `json:"entryType" yaml:"entryType"`
	EntryDescription string                 `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        map[string]interface{} `json:"entryMeta" yaml:"entryMeta"`
	Path             string                 `json:"path" yaml:"path"`
}

// ConfigsResponse response of /configs
type ConfigsResponse struct {
	Entries []*ConfigsResponse_ConfigEntry `json:"entries" yaml:"entries"`
}

// ApisResponse response for path of /apis
type ApisResponse struct {
	Entries []*ApisResponse_Entry `json:"entries" yaml:"entries"`
}

// ApisResponse_Entry Entry for /apis
type ApisResponse_Entry struct {
	EntryName string             `json:"entryName" yaml:"entryName"`
	Grpc      *ApisResponse_Grpc `json:"grpc" yaml:"grpc"`
	Rest      *ApisResponse_Rest `json:"rest" yaml:"rest"`
}

// ApisResponse_Grpc Entry for /apis
type ApisResponse_Grpc struct {
	Service string             `json:"service" yaml:"service"`
	Method  string             `json:"method" yaml:"method"`
	Type    string             `json:"type" yaml:"type"`
	Port    uint64             `json:"port" yaml:"port"`
	Gw      *ApisResponse_Rest `json:"gw" yaml:"gw"`
}

// ApisResponse_Rest Entry for /apis
type ApisResponse_Rest struct {
	Port    uint64 `json:"port" yaml:"port"`
	Pattern string `json:"pattern" yaml:"pattern"`
	Method  string `json:"method" yaml:"method"`
	SwUrl   string `json:"swUrl" yaml:"swUrl"`
}

// SysResponse response of /sys
type SysResponse struct {
	CpuInfo   *CpuInfo   `json:"cpuInfo" yaml:"cpuInfo"`
	MemInfo   *MemInfo   `json:"memInfo" yaml:"memInfo"`
	NetInfo   *NetInfo   `json:"netInfo" yaml:"netInfo"`
	OsInfo    *OsInfo    `json:"osInfo" yaml:"osInfo"`
	GoEnvInfo *GoEnvInfo `json:"goEnvInfo" yaml:"goEnvInfo"`
}

// EntriesResponse response of /entries
type EntriesResponse struct {
	Entries map[string][]*EntriesResponse_Entry `json:"entries" yaml:"entries"`
}

// EntriesResponse_Entry Entry element which specifies name, type and description.
type EntriesResponse_Entry struct {
	EntryName        string `json:"entryName" yaml:"entryName"`
	EntryType        string `json:"entryType" yaml:"entryType"`
	EntryDescription string `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        Entry  `json:"entryMeta" yaml:"entryMeta"`
}

// CertsResponse response of /certs
type CertsResponse struct {
	Entries []*CertsResponse_Entry `json:"entries" yaml:"entries"`
}

// CertsResponse_Entry Entry for /certs
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

// LogsResponse response of /logs.
type LogsResponse struct {
	Entries map[string][]*LogsResponse_Entry `json:"entries" yaml:"entries"`
}

// LogsResponse_Entry Entry element which specifies name, type. description, output path and error output path.
type LogsResponse_Entry struct {
	EntryName        string   `json:"entryName" yaml:"entryName"`
	EntryType        string   `json:"entryType" yaml:"entryType"`
	EntryDescription string   `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        Entry    `json:"entryMeta" yaml:"entryMeta"`
	OutputPaths      []string `json:"outputPaths" yaml:"outputPaths"`
	ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`
}

// ReqResponse response of /req
type ReqResponse struct {
	Metrics []*ReqMetricsRK `json:"metrics" yaml:"metrics"`
}

// DepResponse response of /dep
type DepResponse struct {
	GoMod string `json:"goMod" yaml:"goMod"`
}

// LicenseResponse response of /license
type LicenseResponse struct {
	License string `json:"license" yaml:"license"`
}

// ReadmeResponse response of /readme
type ReadmeResponse struct {
	Readme string `json:"readme" yaml:"readme"`
}

// GwErrorMappingResponse response of /gwErrorMapping
type GwErrorMappingResponse struct {
	Mapping map[int32]*GwErrorMappingResponse_Mapping `json:"mapping" yaml:"mapping"`
}

// GwErrorMappingResponse_Mapping element of mapping of grpc code to restful code with grpc-gateway
type GwErrorMappingResponse_Mapping struct {
	GrpcCode int32  `json:"grpcCode" yaml:"grpcCode"`
	GrpcText string `json:"grpcText" yaml:"grpcText"`
	RestCode int32  `json:"restCode" yaml:"restCode"`
	RestText string `json:"restText" yaml:"restText"`
}

// GitResponse response of /git
type GitResponse struct {
	Package        string `json:"package" yaml:"package"`
	Url            string `json:"url" yaml:"url"`
	Branch         string `yaml:"branch" json:"branch"`
	Tag            string `yaml:"tag" json:"tag"`
	CommitId       string `yaml:"commitId" json:"commitId"`
	CommitIdAbbr   string `yaml:"commitIdAbbr" json:"commitIdAbbr"`
	CommitDate     string `yaml:"commitDate" json:"commitDate"`
	CommitSub      string `yaml:"commitSub" json:"commitSub"`
	CommitterName  string `yaml:"committerName" json:"committerName"`
	CommitterEmail string `yaml:"committerEmail" json:"committerEmail"`
}
