// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkembed is a static files which aims to include assets files into embed.FS
package rkembed

import (
	"embed"
)

//go:embed assets/*
var AssetsFS embed.FS