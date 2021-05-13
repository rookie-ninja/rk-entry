// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"math"
	"runtime"
	"time"
)

// Memory stats of current running process
type MemInfo struct {
	MemUsedPercentage float64 `json:"memUsedPercentage" yaml:"memUsedPercentage"`
	MemUsedMb         uint64  `json:"memUsedMb" yaml:"memUsedMb"`
	MemAllocByte      uint64  `json:"memAllocByte" yaml:"memAllocByte"`
	SysAllocByte      uint64  `json:"sysAllocByte" yaml:"sysAllocByte"`
	LastGcTimestamp   string  `json:"lastGcTimestamp" yaml:"lastGcTimestamp"`
	GcCount           uint32  `json:"gcCountTotal" yaml:"gcCountTotal"`
	ForceGcCount      uint32  `json:"forceGcCount" yaml:"forceGcCount"`
}

func NewMemInfo() *MemInfo {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return &MemInfo{
		MemUsedPercentage: math.Round(float64(stats.Alloc)/float64(stats.Sys)*100) / 100,
		MemUsedMb:         stats.Alloc / (1024 * 1024),
		MemAllocByte:      stats.Alloc,
		SysAllocByte:      stats.Sys,
		LastGcTimestamp:   time.Unix(0, int64(stats.LastGC)).Format(time.RFC3339),
		GcCount:           stats.NumGC,
		ForceGcCount:      stats.NumForcedGC,
	}
}
