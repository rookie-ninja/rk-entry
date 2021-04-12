// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"os"
)

func main() {
	os.Setenv("DOMAIN", "prod")

	configFilePath := "example/my-boot.yaml"
	// 1: register basic entry into global rk context
	rkentry.RegisterBasicEntriesFromConfig(configFilePath)

	// 2: register my entry into global rk context
	RegisterMyEntriesFromConfig(configFilePath)

	// 3: retrieve entry from global context and convert it into MyEntry
	raw := rkentry.GlobalAppCtx.GetEntry("my-entry")

	entry, _ := raw.(*MyEntry)

	// 4: bootstrap entry
	entry.Bootstrap(context.Background())
}

// Register entry, must be in init() function since we need to register entry at beginning
func init() {
	rkentry.RegisterEntryRegFunc(RegisterMyEntriesFromConfig)
}

// A struct which is for unmarshalled YAML
type BootConfig struct {
	MyEntry struct {
		Enabled   bool   `yaml:"enabled"`
		Name      string `yaml:"name"`
		Key       string `yaml:"key"`
		AppLogger struct {
			Ref string `yaml:"ref"`
		} `yaml:"appLogger"`
		EventLogger struct {
			Ref string `yaml:"ref"`
		}
	} `yaml:"myEntry"`
}

// An implementation of:
// type EntryRegFunc func(string) map[string]rkentry.Entry
func RegisterMyEntriesFromConfig(configFilePath string) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: decode config map into boot config struct
	config := &BootConfig{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 3: construct entry
	if config.MyEntry.Enabled {
		appLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(config.MyEntry.AppLogger.Ref)
		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(config.MyEntry.EventLogger.Ref)

		entry := RegisterMyEntry(
			WithName(config.MyEntry.Name),
			WithKey(config.MyEntry.Key),
			WithAppLoggerEntry(appLoggerEntry),
			WithEventLoggerEntry(eventLoggerEntry))
		res[entry.GetName()] = entry
	}

	return res
}

func RegisterMyEntry(opts ...MyEntryOption) *MyEntry {
	entry := &MyEntry{
		name:             "my-default",
		appLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		eventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.name) < 1 {
		entry.name = "my-default"
	}

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

type MyEntryOption func(*MyEntry)

func WithName(name string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.name = name
	}
}

func WithKey(key string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.key = key
	}
}

func WithAppLoggerEntry(appLoggerEntry *rkentry.ZapLoggerEntry) MyEntryOption {
	return func(entry *MyEntry) {
		if appLoggerEntry != nil {
			entry.appLoggerEntry = appLoggerEntry
		}
	}
}

func WithEventLoggerEntry(eventLoggerEntry *rkentry.EventLoggerEntry) MyEntryOption {
	return func(entry *MyEntry) {
		if eventLoggerEntry != nil {
			entry.eventLoggerEntry = eventLoggerEntry
		}
	}
}

type MyEntry struct {
	name             string
	key              string
	appLoggerEntry   *rkentry.ZapLoggerEntry
	eventLoggerEntry *rkentry.EventLoggerEntry
}

func (entry *MyEntry) Bootstrap(context.Context) {
	event := entry.eventLoggerEntry.GetEventHelper().Start("bootstrap")
	event.AddPair("key", entry.key)
	entry.eventLoggerEntry.GetEventHelper().Finish(event)
}

func (entry *MyEntry) Interrupt(context.Context) {}

func (entry *MyEntry) GetName() string {
	return entry.name
}

func (entry *MyEntry) GetType() string {
	return "example-entry"
}

func (entry *MyEntry) String() string {
	m := map[string]string{
		"name": entry.GetName(),
		"type": entry.GetType(),
		"key":  entry.key,
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}
