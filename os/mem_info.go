// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"math"
	"runtime"
	"time"
)

// MemInfo memory stats of current running process
type MemInfo struct {
	MemUsedPercentage float64 `json:"memUsedPercentage" yaml:"memUsedPercentage" example:"0.21"`
	MemUsedMb         uint64  `json:"memUsedMb" yaml:"memUsedMb" example:"3"`
	MemAllocByte      uint64  `json:"memAllocByte" yaml:"memAllocByte" example:"4182336"`
	SysAllocByte      uint64  `json:"sysAllocByte" yaml:"sysAllocByte" example:"19876624"`
	LastGcTimestamp   string  `json:"lastGcTimestamp" yaml:"lastGcTimestamp" example:"2022-03-15T20:43:06+08:00"`
	GcCount           uint32  `json:"gcCountTotal" yaml:"gcCountTotal" example:"1"`
	ForceGcCount      uint32  `json:"forceGcCount" yaml:"forceGcCount" example:"0"`
}

// NewMemInfo creates a new MemInfo
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
