// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"runtime"
	"time"
)

// Memory stats of current running process
type MemStatsInfo struct {
	MemAllocByte    uint64  `json:"mem_alloc_byte"`
	SysAllocByte    uint64  `json:"sys_alloc_byte"`
	MemPercentage   float64 `json:"mem_usage_percentage"`
	LastGCTimestamp string  `json:"last_gc_timestamp"`
	GCCount         uint32  `json:"gc_count_total"`
	ForceGCCount    uint32  `json:"force_gc_count"`
}

func NewMemStatsInfo() *MemStatsInfo {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return &MemStatsInfo{
		MemAllocByte:    stats.Alloc,
		SysAllocByte:    stats.Sys,
		MemPercentage:   float64(stats.Alloc) / float64(stats.Sys),
		LastGCTimestamp: time.Unix(0, int64(stats.LastGC)).Format(time.RFC3339),
		GCCount:         stats.NumGC,
		ForceGCCount:    stats.NumForcedGC,
	}
}
