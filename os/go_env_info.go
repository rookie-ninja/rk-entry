// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"runtime"
)

// GoEnvInfo defines go processor information
type GoEnvInfo struct {
	GOOS          string `json:"goos" yaml:"goos"`
	GOArch        string `json:"goArch" yaml:"goArch"`
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
		RoutinesCount: runtime.NumGoroutine(),
		Version:       runtime.Version(),
	}
}
