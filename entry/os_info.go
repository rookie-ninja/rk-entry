// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"os"
	"runtime"
)

var osInfo = initOsInfo()

func initOsInfo() *OsInfo {
	hostname, _ := os.Hostname()

	return &OsInfo{
		Os:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: hostname,
	}
}

type OsInfo struct {
	Os       string `json:"os" yaml:"os"`
	Arch     string `json:"arch" yaml:"arch"`
	Hostname string `json:"hostname" yaml:"hostname"`
}

func NewOsInfo() *OsInfo {
	return osInfo
}
