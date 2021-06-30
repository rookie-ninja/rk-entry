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
	"github.com/rookie-ninja/rk-query"
	"os"
)

func main() {
	os.Setenv("DOMAIN", "prod")

	configFilePath := "example/my-boot.yaml"
	// 1: register basic entry into global rk context
	rkentry.RegisterInternalEntriesFromConfig(configFilePath)

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
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Description string `yaml:"description" json:"description"`
		Key         string `yaml:"key" json:"key"`
		Logger      struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"myEntry" json:"myEntry"`
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
		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(config.MyEntry.Logger.ZapLogger.Ref)
		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(config.MyEntry.Logger.EventLogger.Ref)

		entry := RegisterMyEntry(
			WithName(config.MyEntry.Name),
			WithDescription(config.MyEntry.Description),
			WithKey(config.MyEntry.Key),
			WithZapLoggerEntry(zapLoggerEntry),
			WithEventLoggerEntry(eventLoggerEntry))
		res[entry.GetName()] = entry
	}

	return res
}

func RegisterMyEntry(opts ...MyEntryOption) *MyEntry {
	entry := &MyEntry{
		EntryName:        "default",
		EntryType:        "myEntry",
		EntryDescription: "Please contact maintainers to add description of this entry.",
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
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

type MyEntryOption func(*MyEntry)

func WithName(name string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.EntryName = name
	}
}

func WithDescription(description string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.EntryDescription = description
	}
}

func WithKey(key string) MyEntryOption {
	return func(entry *MyEntry) {
		entry.Key = key
	}
}

func WithZapLoggerEntry(zapLoggerEntry *rkentry.ZapLoggerEntry) MyEntryOption {
	return func(entry *MyEntry) {
		if zapLoggerEntry != nil {
			entry.ZapLoggerEntry = zapLoggerEntry
		}
	}
}

func WithEventLoggerEntry(eventLoggerEntry *rkentry.EventLoggerEntry) MyEntryOption {
	return func(entry *MyEntry) {
		if eventLoggerEntry != nil {
			entry.EventLoggerEntry = eventLoggerEntry
		}
	}
}

type MyEntry struct {
	EntryName        string                    `json:"entryName" yaml:"entryName"`
	EntryType        string                    `json:"entryType" yaml:"entryType"`
	EntryDescription string                    `json:"entryDescription" yaml:"entryDescription"`
	Key              string                    `json:"key" yaml:"key"`
	ZapLoggerEntry   *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	EventLoggerEntry *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
}

func (entry *MyEntry) Bootstrap(context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))
	event.AddPair("key", entry.Key)
	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

func (entry *MyEntry) Interrupt(context.Context) {}

func (entry *MyEntry) GetName() string {
	return entry.EntryName
}

func (entry *MyEntry) GetType() string {
	return entry.EntryType
}

func (entry *MyEntry) String() string {
	bytes, _ := json.Marshal(entry)

	return string(bytes)
}

// Marshal entry
func (entry *MyEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"eventLoggerEntry": entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":   entry.ZapLoggerEntry.GetName(),
		"key":              entry.Key,
	}

	return json.Marshal(&m)
}

// Unmarshal entry
func (entry *MyEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Get description of entry
func (entry *MyEntry) GetDescription() string {
	return entry.EntryDescription
}
