// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"net"
	"strings"
)

var netInfos = initNetInfos()

func initNetInfos() *NetInfo {
	res := &NetInfo{
		NetInterface: make([]*NetInterface, 0),
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return res
	}

	for i := range interfaces {
		element := interfaces[i]
		info := &NetInterface{
			Name:           element.Name,
			Mtu:            element.MTU,
			HardwareAddr:   element.HardwareAddr.String(),
			Flags:          strings.Split(element.Flags.String(), "|"),
			Addrs:          make([]string, 0),
			MulticastAddrs: make([]string, 0),
		}

		if addrs, err := element.Addrs(); err == nil {
			for j := range addrs {
				info.Addrs = append(info.Addrs, addrs[j].String())
			}
		}

		if multicastAddrs, err := element.MulticastAddrs(); err == nil {
			for j := range multicastAddrs {
				info.MulticastAddrs = append(info.MulticastAddrs, multicastAddrs[j].String())
			}
		}

		res.NetInterface = append(res.NetInterface, info)
	}

	return res
}

type NetInfo struct {
	NetInterface []*NetInterface `json:"netInterface" yaml:"netInterface"`
}

type NetInterface struct {
	// e.g., "en0", "lo0", "eth0.100"
	Name string `json:"name" yaml:"name"`
	// maximum transmission unit
	Mtu int `json:"mtu" yaml:"mtu"`
	// IEEE MAC-48, EUI-48 and EUI-64 form
	HardwareAddr string `json:"hardwareAddr" yaml:"hardwareAddr"`
	// e.g., FlagUp, FlagLoopback, FlagMulticast
	Flags []string `json:"flags" yaml:"flags"`
	// A list of unicast interface addresses for a specific interface.
	Addrs []string `json:"addrs" yaml:"addrs"`
	// A list of multicast, joined group addresses for a specific interface
	MulticastAddrs []string `json:"multicastAddrs" yaml:"multicastAddrs"`
}

func NewNetInfo() *NetInfo {
	return netInfos
}
