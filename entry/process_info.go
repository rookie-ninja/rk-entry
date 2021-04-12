// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/hako/durafmt"
	"github.com/rookie-ninja/rk-common/common"
	"os"
	"os/user"
	"time"
)

// Process information for a running application
// 1: AppName - Name of current application set by user
// 2: Version - Version of application set by user
// 3: Description - Description of application set by user
// 4: Keywords - Keywords of application set by user
// 5: HomeURL - Home URL of application set by user
// 6: IconURL - Icon URL of application set by user
// 7: DocsURL - Document URL list of application set by user
// 8: Maintainers - Maintainers of applicaiton set by user
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
	AppName         string   `json:"app_name"`
	Version         string   `json:"version"`
	Description     string   `json:"description"`
	Keywords        []string `json:"keywords"`
	HomeURL         string   `json:"home_url"`
	IconURL         string   `json:"icon_url"`
	DocsURL         []string `json:"docs_url"`
	Maintainers     []string `json:"maintainers"`
	UID             string   `json:"uid"`
	GID             string   `json:"gid"`
	Username        string   `json:"username"`
	StartTime       string   `json:"start_time"`
	UpTimeSec       int64    `json:"up_time_sec"`
	UpTimeStr       string   `json:"up_time_str"`
	Region          string   `json:"region"`
	AZ              string   `json:"az"`
	Realm           string   `json:"realm"`
	Domain          string   `json:"domain"`
}

func NewProcessInfo() *ProcessInfo {
	u, err := user.Current()
	// assign unknown value to user in order to prevent panic
	if err != nil {
		u = &user.User{
			Name: "unknown",
			Uid:  "unknown",
			Gid:  "unknown",
		}
	}

	return &ProcessInfo{
		AppName:         GlobalAppCtx.GetAppInfoEntry().AppName,
		Version:         GlobalAppCtx.GetAppInfoEntry().Version,
		Description:     GlobalAppCtx.GetAppInfoEntry().Description,
		Keywords:        GlobalAppCtx.GetAppInfoEntry().Keywords,
		HomeURL:         GlobalAppCtx.GetAppInfoEntry().HomeURL,
		IconURL:         GlobalAppCtx.GetAppInfoEntry().IconURL,
		DocsURL:         GlobalAppCtx.GetAppInfoEntry().DocsURL,
		Maintainers:     GlobalAppCtx.GetAppInfoEntry().Maintainers,
		Username:        u.Name,
		UID:             u.Uid,
		GID:             u.Gid,
		StartTime:       GlobalAppCtx.StartTime.Format(time.RFC3339),
		UpTimeSec:       int64(GlobalAppCtx.GetUpTime().Seconds()),
		UpTimeStr:       durafmt.ParseShort(GlobalAppCtx.GetUpTime()).String(),
		Realm:           rkcommon.GetDefaultIfEmptyString(os.Getenv("REALM"), "unknown"),
		Region:          rkcommon.GetDefaultIfEmptyString(os.Getenv("REGION"), "unknown"),
		AZ:              rkcommon.GetDefaultIfEmptyString(os.Getenv("AZ"), "unknown"),
		Domain:          rkcommon.GetDefaultIfEmptyString(os.Getenv("DOMAIN"), "unknown"),
	}
}
