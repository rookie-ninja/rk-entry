// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"math"
)

var cpuInfos = initCpuInfos()

func initCpuInfos() *CpuInfo {
	logicalCoreCount, _ := cpu.Counts(true)
	physicalCoreCount, _ := cpu.Counts(false)

	res := &CpuInfo{
		LogicalCoreCount:  logicalCoreCount,
		PhysicalCoreCount: physicalCoreCount,
	}

	cores, _ := cpu.Info()
	if cores != nil && len(cores) > 0 {
		res.VendorId = cores[0].VendorID
		res.ModelName = cores[0].ModelName
		res.Mhz = cores[0].Mhz
		res.CacheSize = cores[0].CacheSize

	}

	return res
}

// CpuInfo defines CPU information read from system
type CpuInfo struct {
	CpuUsedPercentage float64 `json:"cpuUsedPercentage" yaml:"cpuUsedPercentage"`
	LogicalCoreCount  int     `json:"logicalCoreCount" yaml:"logicalCoreCount"`
	PhysicalCoreCount int     `json:"physicalCoreCount" yaml:"physicalCoreCount"`
	VendorId          string  `json:"vendorId" yaml:"vendorId"`
	ModelName         string  `json:"modelName" yaml:"modelName"`
	Mhz               float64 `json:"mhz" yaml:"mhz"`
	CacheSize         int32   `json:"cacheSize" yaml:"cacheSize"`
}

// NewCpuInfo creates a new CpuInfo instance
func NewCpuInfo() *CpuInfo {
	var cpuUsedPercentage float64
	cpuStat, _ := cpu.Percent(0, false)

	for i := range cpuStat {
		cpuUsedPercentage = math.Round(cpuStat[i]*100) / 100
	}

	cpuInfos.CpuUsedPercentage = cpuUsedPercentage

	return cpuInfos
}
