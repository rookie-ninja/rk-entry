// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	_ "github.com/rookie-ninja/rk-query"
	"os"
)

//go:embed my-boot.yaml
var boot []byte

func main() {
	os.Setenv("DOMAIN", "prod")

	// 1: register my entry into global rk context
	RegisterMyEntryFromConfig(boot)

	// 2: retrieve entry from global context and convert it into MyEntry
	raw := rkentry.GlobalAppCtx.GetEntry("MyEntry", "MyEntry")

	entry, _ := raw.(*MyEntry)

	// 3: bootstrap entry
	entry.Bootstrap(context.Background())
}

// Register entry, must be in init() function since we need to register entry at beginning
func init() {
	rkentry.RegisterEntryRegFunc(RegisterMyEntryFromConfig)
}

// BootConfig A struct which is for unmarshalled YAML
type BootConfig struct {
	MyEntry struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Description string `yaml:"description" json:"description"`
		Key         string `yaml:"key" json:"key"`
	} `yaml:"myEntry" json:"myEntry"`
}

// RegisterMyEntryFromConfig an implementation of:
// type EntryRegFunc func([]byte) map[string]rke.Entry
func RegisterMyEntryFromConfig(raw []byte) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: decode config map into boot config struct
	config := &BootConfig{}
	rkentry.UnmarshalBootYAML(raw, config)

	// 3: construct entry
	if config.MyEntry.Enabled {
		entry := RegisterMyEntry(
			WithName(config.MyEntry.Name),
			WithDescription(config.MyEntry.Description),
			WithKey(config.MyEntry.Key))
		res[entry.GetName()] = entry
	}

	return res
}

// RegisterMyEntry register entry based on code
func RegisterMyEntry(opts ...MyEntryOption) *MyEntry {
	entry := &MyEntry{
		EntryName:        "MyEntry",
		EntryType:        "MyEntry",
		EntryDescription: "Please contact maintainers to add description of this entry.",
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "my-default"
	}

	if len(entry.EntryDescription) < 1 {
		entry.EntryDescription = "Please contact maintainers to add description of this entry."
	}

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// MyEntryOption options of MyEntry
type MyEntryOption func(*MyEntry)

// WithName provide name of entry
func WithName(name string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.EntryName = name
	}
}

// WithDescription provide description of entry
func WithDescription(description string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.EntryDescription = description
	}
}

// WithKey provide key field in entry
func WithKey(key string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.Key = key
	}
}

// MyEntry is a implementation of Entry
type MyEntry struct {
	EntryName        string `json:"-" yaml:"-"`
	EntryType        string `json:"-" yaml:"-"`
	EntryDescription string `json:"-" yaml:"-"`
	Key              string `json:"-" yaml:"-"`
}

// Bootstrap init required fields in MyEntry
func (entry *MyEntry) Bootstrap(context.Context, ...rkentry.PreloadFunc) {}

// Interrupt noop
func (entry *MyEntry) Interrupt(context.Context) {}

// GetName returns name of entry
func (entry *MyEntry) GetName() string {
	return entry.EntryName
}

// GetType returns type of entry
func (entry *MyEntry) GetType() string {
	return entry.EntryType
}

// String returns string value of entry
func (entry *MyEntry) String() string {
	bytes, _ := json.Marshal(entry)

	return string(bytes)
}

// MarshalJSON marshal entry
func (entry *MyEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.EntryName,
		"type":        entry.EntryType,
		"description": entry.EntryDescription,
		"key":         entry.Key,
	}

	return json.Marshal(&m)
}

// UnmarshalJSON unmarshal entry
func (entry *MyEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetDescription returns description of entry
func (entry *MyEntry) GetDescription() string {
	return entry.EntryDescription
}
