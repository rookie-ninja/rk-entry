// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/hako/durafmt"
	"runtime"
	"time"
)

// GoEnvInfo defines go processor information
type GoEnvInfo struct {
	GOOS          string `json:"goos" yaml:"goos"`
	GOArch        string `json:"goArch" yaml:"goArch"`
	StartTime     string `json:"startTime" json:"startTime"`
	UpTimeSec     int64  `json:"upTimeSec" json:"upTimeSec"`
	UpTimeStr     string `json:"upTimeStr" json:"upTimeStr"`
	RoutinesCount int    `json:"routinesCount" yaml:"routinesCount"`
	Version       string `json:"version" yaml:"version"`
}

// NewGoEnvInfo creates a new instance of GoEnvInfo
func NewGoEnvInfo() *GoEnvInfo {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return &GoEnvInfo{
		GOOS:          runtime.GOOS,
		GOArch:        runtime.GOARCH,
		StartTime:     GlobalAppCtx.GetStartTime().Format(time.RFC3339),
		UpTimeSec:     int64(GlobalAppCtx.GetUpTime().Seconds()),
		UpTimeStr:     durafmt.ParseShort(GlobalAppCtx.GetUpTime()).String(),
		RoutinesCount: runtime.NumGoroutine(),
		Version:       runtime.Version(),
	}
}
