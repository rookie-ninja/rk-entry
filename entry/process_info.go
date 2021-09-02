// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/hako/durafmt"
	"github.com/rookie-ninja/rk-common/common"
	"os"
	"os/user"
	"time"
)

// Process information for a running application.
// 1: AppName - Name of current application set by user
// 2: Version - Version of application set by user
// 3: Description - Description of application set by user
// 4: Keywords - Keywords of application set by user
// 5: HomeUrl - Home URL of application set by user
// 6: IconUrl - Icon URL of application set by user
// 7: DocsUrl; - Document URL list of application set by user
// 8: Maintainers - Maintainers of application set by user
// 9: UID - user id which runs process
// 10: GID - group id which runs process
// 11: Username - username which runs process
// 12: StartTime - application start time
// 13: UpTimeSec - application up time in seconds
// 14: UpTimeStr - application up time in string
// 15: Region - region where process runs
// 16: AZ - availability zone where process runs
// 17: Realm - realm where process runs
// 18: Domain - domain where process runs
type ProcessInfo struct {
	AppName     string   `json:"appName" yaml:"appName"`
	Version     string   `json:"version" yaml:"version"`
	Description string   `json:"description" yaml:"description"`
	Keywords    []string `json:"keywords" yaml:"keywords"`
	HomeUrl     string   `json:"homeUrl" yaml:"homeUrl"`
	IconUrl     string   `json:"iconUrl" yaml:"iconUrl"`
	DocsUrl     []string `json:"docsUrl" json:"docsUrl"`
	Maintainers []string `json:"maintainers" json:"maintainers"`
	UID         string   `json:"uid" json:"uid"`
	GID         string   `json:"gid" json:"gid"`
	Username    string   `json:"username" json:"username"`
	StartTime   string   `json:"startTime" json:"startTime"`
	UpTimeSec   int64    `json:"upTimeSec" json:"upTimeSec"`
	UpTimeStr   string   `json:"upTimeStr" json:"upTimeStr"`
	Region      string   `json:"region" json:"region"`
	AZ          string   `json:"az" json:"az"`
	Realm       string   `json:"realm" json:"realm"`
	Domain      string   `json:"domain" json:"domain"`
}

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
		Description: GlobalAppCtx.GetAppInfoEntry().EntryDescription,
		Keywords:    GlobalAppCtx.GetAppInfoEntry().Keywords,
		HomeUrl:     GlobalAppCtx.GetAppInfoEntry().HomeUrl,
		IconUrl:     GlobalAppCtx.GetAppInfoEntry().IconUrl,
		DocsUrl:     GlobalAppCtx.GetAppInfoEntry().DocsUrl,
		Maintainers: GlobalAppCtx.GetAppInfoEntry().Maintainers,
		Username:    u.Name,
		UID:         u.Uid,
		GID:         u.Gid,
		StartTime:   GlobalAppCtx.GetStartTime().Format(time.RFC3339),
		UpTimeSec:   int64(GlobalAppCtx.GetUpTime().Seconds()),
		UpTimeStr:   durafmt.ParseShort(GlobalAppCtx.GetUpTime()).String(),
		Realm:       rkcommon.GetDefaultIfEmptyString(os.Getenv("REALM"), ""),
		Region:      rkcommon.GetDefaultIfEmptyString(os.Getenv("REGION"), ""),
		AZ:          rkcommon.GetDefaultIfEmptyString(os.Getenv("AZ"), ""),
		Domain:      rkcommon.GetDefaultIfEmptyString(os.Getenv("DOMAIN"), ""),
	}
}
