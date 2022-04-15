// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"strings"
)

// bootConfigAppInfo is config of application's basic information.
type bootConfigAppInfo struct {
	App struct {
		Name        string   `yaml:"name" json:"name"`
		Version     string   `yaml:"version" json:"version"`
		Description string   `yaml:"description" json:"description"`
		Keywords    []string `yaml:"keywords" json:"keywords"`
		HomeUrl     string   `yaml:"homeUrl" json:"homeUrl"`
		DocsUrl     []string `yaml:"docsUrl" json:"docsUrl"`
		Maintainers []string `yaml:"maintainers" json:"maintainers"`
	} `yaml:"app"`
}

// appInfoEntry contains bellow fields.
type appInfoEntry struct {
	entryName        string   `json:"-" yaml:"-"`
	entryType        string   `json:"-" yaml:"-"`
	entryDescription string   `json:"-" yaml:"-"`
	AppName          string   `json:"-" yaml:"-"`
	Version          string   `json:"-" yaml:"-"`
	Lang             string   `json:"-" yaml:"-"`
	Keywords         []string `json:"-" yaml:"-"`
	HomeUrl          string   `json:"-" yaml:"-"`
	DocsUrl          []string `json:"-" yaml:"-"`
	Maintainers      []string `json:"-" yaml:"-"`
}

// appInfoEntryDefault generate a AppInfo entry with default fields.
func appInfoEntryDefault() *appInfoEntry {
	return &appInfoEntry{
		entryName:        appInfoEntryName,
		entryType:        appInfoEntryType,
		entryDescription: "Internal RK entry which describes application with fields of appName, version and etc.",
		AppName:          "rk",
		Version:          "local",
		Lang:             "golang",
		Keywords:         []string{},
		HomeUrl:          "",
		DocsUrl:          []string{},
		Maintainers:      []string{},
	}
}

// registerAppInfoEntryYAML register appInfoEntry with bytes of YAML
func registerAppInfoEntryYAML(raw []byte) map[string]Entry {
	// Unmarshal user provided config into boot config struct
	config := &bootConfigAppInfo{}
	UnmarshalBootYAML(raw, config)
	res := map[string]Entry{}

	entry := appInfoEntryDefault()
	if len(config.App.Name) > 0 {
		entry.AppName = config.App.Name
	}

	if len(config.App.Version) > 0 {
		entry.Version = config.App.Version
	}

	if len(config.App.Description) > 0 {
		entry.entryDescription = config.App.Description
	}

	entry.Keywords = config.App.Keywords
	entry.HomeUrl = config.App.HomeUrl
	entry.DocsUrl = config.App.DocsUrl
	entry.Maintainers = config.App.Maintainers

	if entry.Keywords == nil {
		entry.Keywords = make([]string, 0)
	}

	if entry.DocsUrl == nil {
		entry.DocsUrl = make([]string, 0)
	}

	if entry.Maintainers == nil {
		entry.Maintainers = make([]string, 0)
	}

	GlobalAppCtx.appInfoEntry = entry

	EventEntryStdout = NewEventEntryStdout()
	LoggerEntryStdout = NewLoggerEntryStdout()

	res[entry.GetName()] = entry
	return res
}

// Bootstrap is noop function.
func (entry *appInfoEntry) Bootstrap(context.Context) {}

// Interrupt is noop function.
func (entry *appInfoEntry) Interrupt(context.Context) {}

// GetName return name of entry.
func (entry *appInfoEntry) GetName() string {
	return entry.entryName
}

// GetType return type of entry.
func (entry *appInfoEntry) GetType() string {
	return entry.entryType
}

// GetDescription return description of entry.
func (entry *appInfoEntry) GetDescription() string {
	return entry.entryDescription
}

// String return string value of entry.
func (entry *appInfoEntry) String() string {
	bytes, _ := json.Marshal(entry)

	return string(bytes)
}

// MarshalJSON Marshal entry.
func (entry *appInfoEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.GetName(),
		"type":        entry.GetType(),
		"description": entry.GetDescription(),
		"appName":     entry.AppName,
		"lang":        entry.Lang,
		"homeUrl":     entry.HomeUrl,
		"docsUrl":     entry.DocsUrl,
		"maintainers": strings.Join(entry.Maintainers, ","),
	}

	return json.Marshal(m)
}

// UnmarshalJSON Not supported.
func (entry *appInfoEntry) UnmarshalJSON([]byte) error {
	return nil
}
