// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkos

import (
	"net"
	"strings"
)

func NewNetInfo() *NetInfo {
	res := &NetInfo{
		NetInterface: make([]*netInterface, 0),
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return res
	}

	for i := range interfaces {
		element := interfaces[i]
		info := &netInterface{
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

// NetInfo defines network interface information about local machine
type NetInfo struct {
	NetInterface []*netInterface `json:"netInterface" yaml:"netInterface"`
}

// netInterface describes network interface
type netInterface struct {
	Name           string   `json:"name" yaml:"name" example:"lo0"`
	Mtu            int      `json:"mtu" yaml:"mtu" example:"16384"`
	HardwareAddr   string   `json:"hardwareAddr" yaml:"hardwareAddr" example:""`
	Flags          []string `json:"flags" yaml:"flags" example:"up"`
	Addrs          []string `json:"addrs" yaml:"addrs" example:"127.0.0.1/8"`
	MulticastAddrs []string `json:"multicastAddrs" yaml:"multicastAddrs" example:"ff02::fb"`
}
