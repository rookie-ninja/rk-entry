// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/rookie-ninja/rk-entry/os"
	"os"
	"os/user"
	"time"
)

// healthyResp response of /healthy
type healthyResp struct {
	Healthy bool `json:"healthy" yaml:"healthy"`
}

// gcResp response of /gc
// Returns memory stats of GC before and after.
type gcResp struct {
	MemStatBeforeGc *rkos.MemInfo `json:"memStatBeforeGc" yaml:"memStatBeforeGc"`
	MemStatAfterGc  *rkos.MemInfo `json:"memStatAfterGc" yaml:"memStatAfterGc"`
}

// ProcessInfo process information for a running application.
type ProcessInfo struct {
	AppName     string          `json:"appName" yaml:"appName"`
	Version     string          `json:"version" yaml:"version"`
	Description string          `json:"description" yaml:"description"`
	Keywords    []string        `json:"keywords" yaml:"keywords"`
	HomeUrl     string          `json:"homeUrl" yaml:"homeUrl"`
	DocsUrl     []string        `json:"docsUrl" json:"docsUrl"`
	Maintainers []string        `json:"maintainers" json:"maintainers"`
	UID         string          `json:"uid" json:"uid"`
	GID         string          `json:"gid" json:"gid"`
	Username    string          `json:"username" json:"username"`
	StartTime   string          `json:"startTime" json:"startTime"`
	UpTimeSec   int64           `json:"upTimeSec" json:"upTimeSec"`
	Region      string          `json:"region" json:"region"`
	AZ          string          `json:"az" json:"az"`
	Realm       string          `json:"realm" json:"realm"`
	Domain      string          `json:"domain" json:"domain"`
	CpuInfo     *rkos.CpuInfo   `json:"cpuInfo" yaml:"cpuInfo"`
	MemInfo     *rkos.MemInfo   `json:"memInfo" yaml:"memInfo"`
	NetInfo     *rkos.NetInfo   `json:"netInfo" yaml:"netInfo"`
	OsInfo      *rkos.OsInfo    `json:"osInfo" yaml:"osInfo"`
	GoEnvInfo   *rkos.GoEnvInfo `json:"goEnvInfo" yaml:"goEnvInfo"`
}

// NewProcessInfo creates a new ProcessInfo instance
func NewProcessInfo() *ProcessInfo {
	u, err := user.Current()
	// Assign unknown value to user in order to prevent panic
	if err != nil {
		u = &user.User{
			Name: "",
			Uid:  "",
			Gid:  "",
		}
	}

	return &ProcessInfo{
		AppName:     GlobalAppCtx.GetAppInfoEntry().AppName,
		Version:     GlobalAppCtx.GetAppInfoEntry().Version,
		Description: GlobalAppCtx.GetAppInfoEntry().GetDescription(),
		Keywords:    GlobalAppCtx.GetAppInfoEntry().Keywords,
		HomeUrl:     GlobalAppCtx.GetAppInfoEntry().HomeUrl,
		DocsUrl:     GlobalAppCtx.GetAppInfoEntry().DocsUrl,
		Maintainers: GlobalAppCtx.GetAppInfoEntry().Maintainers,
		Username:    u.Name,
		UID:         u.Uid,
		GID:         u.Gid,
		StartTime:   GlobalAppCtx.GetStartTime().Format(time.RFC3339),
		UpTimeSec:   int64(GlobalAppCtx.GetUpTime().Seconds()),
		Realm:       getDefaultIfEmptyString(os.Getenv("REALM"), ""),
		Region:      getDefaultIfEmptyString(os.Getenv("REGION"), ""),
		AZ:          getDefaultIfEmptyString(os.Getenv("AZ"), ""),
		Domain:      getDefaultIfEmptyString(os.Getenv("DOMAIN"), ""),
		CpuInfo:     rkos.NewCpuInfo(),
		MemInfo:     rkos.NewMemInfo(),
		NetInfo:     rkos.NewNetInfo(),
		OsInfo:      rkos.NewOsInfo(),
		GoEnvInfo:   rkos.NewGoEnvInfo(),
	}
}
