// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/rookie-ninja/rk-entry/v2/os"
	"os"
	"os/user"
	"time"
)

// aliveResp response of /alive
type aliveResp struct {
	Alive bool `json:"alive" yaml:"alive" example:"true"`
}

// readyResp response of /ready
type readyResp struct {
	Ready bool `json:"ready" yaml:"ready" example:"true"`
}

// gcResp response of /gc
// Returns memory stats of GC before and after.
type gcResp struct {
	MemStatBeforeGc *rkos.MemInfo `json:"memStatBeforeGc" yaml:"memStatBeforeGc"`
	MemStatAfterGc  *rkos.MemInfo `json:"memStatAfterGc" yaml:"memStatAfterGc"`
}

// ProcessInfo process information for a running application.
type ProcessInfo struct {
	AppName     string          `json:"appName" yaml:"appName" example:"rk-app"`
	Version     string          `json:"version" yaml:"version" example:"dev"`
	Description string          `json:"description" yaml:"description" example:"RK application"`
	Keywords    []string        `json:"keywords" yaml:"keywords" example:""`
	HomeUrl     string          `json:"homeUrl" yaml:"homeUrl" example:"https://example.com"`
	DocsUrl     []string        `json:"docsUrl" yaml:"docsUrl" example:""`
	Maintainers []string        `json:"maintainers" yaml:"maintainers" example:"rk-dev"`
	UID         string          `json:"uid" yaml:"uid" example:"501"`
	GID         string          `json:"gid" yaml:"gid" example:"20"`
	Username    string          `json:"username" yaml:"username" example:"lark"`
	StartTime   string          `json:"startTime" yaml:"startTime" example:"2022-03-15T20:43:05+08:00"`
	UpTimeSec   int64           `json:"upTimeSec" yaml:"upTimeSec" example:"13"`
	Region      string          `json:"region" yaml:"region" example:"us-east-1"`
	AZ          string          `json:"az" yaml:"az" example:"us-east-1c"`
	Realm       string          `json:"realm" yaml:"realm" example:"rookie-ninja"`
	Domain      string          `json:"domain" yaml:"domain" example:"dev"`
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
