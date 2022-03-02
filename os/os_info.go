// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"os"
	"runtime"
)

func NewOsInfo() *OsInfo {
	hostname, _ := os.Hostname()

	return &OsInfo{
		Os:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: hostname,
	}
}

// OsInfo defines OS information
type OsInfo struct {
	Os       string `json:"os" yaml:"os"`
	Arch     string `json:"arch" yaml:"arch"`
	Hostname string `json:"hostname" yaml:"hostname"`
}
