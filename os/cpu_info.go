// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"runtime"
)

func NewCpuInfo() *CpuInfo {
	res := &CpuInfo{
		Count: runtime.NumCPU(),
	}

	return res
}

// CpuInfo defines CPU information read from system
type CpuInfo struct {
	Count int `json:"count" yaml:"count"`
}
