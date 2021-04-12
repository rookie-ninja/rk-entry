// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"fmt"
)

// Config information stored in GlobalAppCtx
// 1: Name - Name of config instance
// 2: Raw - Raw data in config instance
type ViperConfigInfo struct {
	Name string `json:"name"`
	Raw  string `json:"raw"`
}

// As struct
func NewViperConfigInfo() []*ViperConfigInfo {
	res := make([]*ViperConfigInfo, 0)

	for k, v := range GlobalAppCtx.ListViperEntries() {
		res = append(res, &ViperConfigInfo{
			Name: k,
			Raw:  fmt.Sprintf("%v", v.vp.AllSettings()),
		})
	}
	return res
}
