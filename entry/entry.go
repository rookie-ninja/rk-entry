// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
)

// An entry could be any kinds of services or pieces of codes which
// needs to be bootstrap/initialized while application starts.
//
// A third party entry could be implemented and inject to rk-boot via rk-boot.yaml file
//
// How to create a new custom entry? Please see example/ for details
// Step 1:
// Construct your own entry YAML struct as needed
// example:
// ---
// myEntry:
//   enabled: true
//   key: value
//
// Step 2:
// Create a struct which implements Entry interface
//
// Step 3:
// Implements EntryRegFunc
//
// Step 4:
// Register your reg function in init() in order to register your entry while application starts
//
// How entry interact with rk-boot.Bootstrapper?
// 1: Entry will be created and registered into rkentry.GlobalAppCtx
// 2: Bootstrap will be called from Bootstrapper.Bootstrap() function
// 3: Application will wait for shutdown signal
// 4: Interrupt will be called from Bootstrapper.Interrupt() function

// New entry function which must be implemented
type EntryRegFunc func(configFilePath string) map[string]Entry

// Entry interface which must be implemented for bootstrapper to bootstrap
type Entry interface {
	// Bootstrap entry
	Bootstrap(context.Context)

	// Interrupt entry
	// Wait for shutdown signal and wait for draining incomplete procedure
	Interrupt(context.Context)

	// Get name of entry
	GetName() string

	// Get type of entry
	GetType() string

	// Get description of entry
	GetDescription() string

	// print entry as string
	String() string
}
