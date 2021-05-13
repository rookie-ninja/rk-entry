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
