// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-query"
)

const (
	AppNameDefault     = "rkapp"
	VersionDefault     = "v0.0.0"
	LangDefault        = "golang"
	DescriptionDefault = "rk application"
	AppInfoEntryName   = "rk-app-info-entry"
	AppInfoEntryType   = "rk-app-info-entry"
)

// Bootstrap config of application's basic information.
// 1: AppName: Application name which refers to go process.
// 2: Version: Application version.
// 3: Description: Description of application itself.
// 4: Keywords: A set of words describe application.
// 5: HomeURL: Home page URL.
// 6: IconURL: Application Icon URL.
// 7: Maintainers: Maintainers of application.
// 8: DocsURL: A set of URLs of documentations of application.
type BootConfigAppInfo struct {
	RK struct {
		AppName     string   `yaml:"appName"`
		Version     string   `yaml:"version"`
		Description string   `yaml:"description"`
		Keywords    []string `yaml:"keywords"`
		HomeURL     string   `yaml:"homeURL"`
		IconURL     string   `yaml:"iconURL"`
		DocsURL     []string `yaml:"docsURL"`
		Maintainers []string `yaml:"maintainers"`
	} `yaml:"rk"`
}

// AppInfo Entry contains bellow fields.
// 1: AppName: Application name which refers to go process
// 2: Version: Application version
// 3: Lang: Programming language <NOT configurable!>
// 4: Description: Description of application itself
// 5: Keywords: A set of words describe application
// 6: HomeURL: Home page URL
// 7: IconURL: Application Icon URL
// 8: DocsURL: A set of URLs of documentations of application
// 9: Maintainers: Maintainers of application
type AppInfoEntry struct {
	AppName     string
	Version     string
	Lang        string
	Description string
	Keywords    []string
	HomeURL     string
	IconURL     string
	DocsURL     []string
	Maintainers []string
}

// Generate a AppInfo entry with default fields.
// 1: AppName: rkapp
// 2: Version: v0.0.0
// 3: Lang: golang
// 4: Description: rk application
// 5: Keywords: []
// 6: HomeURL: ""
// 7: IconURL: ""
// 8: Maintainers: []
// 9: DocsURL: []]
func AppInfoEntryDefault() *AppInfoEntry {
	return &AppInfoEntry{
		AppName:     AppNameDefault,
		Version:     VersionDefault,
		Lang:        LangDefault,
		Description: DescriptionDefault,
		Keywords:    []string{},
		HomeURL:     "",
		IconURL:     "",
		DocsURL:     []string{},
		Maintainers: []string{},
	}
}

// AppInfo Entry Option which used while registering entry from codes.
type AppInfoEntryOption func(*AppInfoEntry)

// Provide application name.
func WithAppNameAppInfo(AppName string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.AppName = AppName
	}
}

// Provide version.
func WithVersionAppInfo(version string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Version = version
	}
}

// Provide description.
func WithDescriptionAppInfo(description string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Description = description
	}
}

// Provide home page URL.
func WithHomeURLAppInfo(homeURL string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.HomeURL = homeURL
	}
}

// Provide icon URL.
func WithIconURLAppInfo(iconURL string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.IconURL = iconURL
	}
}

// Provide keywords.
func WithKeywordsAppInfo(keywords ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Keywords = append(entry.Keywords, keywords...)
	}
}

// Provide documentation URLs.
func WithDocsURLAppInfo(docsURL ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.DocsURL = append(entry.DocsURL, docsURL...)
	}
}

// Provide maintainers.
func WithMaintainersAppInfo(maintainers ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Maintainers = append(entry.Maintainers, maintainers...)
	}
}

// Implements rkentry.EntryRegFunc which generate RKEntry based on boot configuration file.
func RegisterAppInfoEntriesFromConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: unmarshal user provided config into boot config struct
	config := &BootConfigAppInfo{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: init rk entry from config
	entry := RegisterAppInfoEntry(
		WithAppNameAppInfo(config.RK.AppName),
		WithVersionAppInfo(config.RK.Version),
		WithDescriptionAppInfo(config.RK.Description),
		WithKeywordsAppInfo(config.RK.Keywords...),
		WithHomeURLAppInfo(config.RK.HomeURL),
		WithIconURLAppInfo(config.RK.IconURL),
		WithDocsURLAppInfo(config.RK.DocsURL...),
		WithMaintainersAppInfo(config.RK.Maintainers...))

	res[AppInfoEntryName] = entry

	return res
}

// Register RKEntry with options.
// This function is used while creating entry from code instead of config file.
// We will override RKEntry fields if value is nil or empty if necessary.
//
// Generally, we recommend call rkctx.GlobalAppCtx.AddEntry() inside this function,
// however, we recommend to register RKEntry, ZapLoggerEntry, EventLoggerEntry with
// function of rkctx.RegisterBasicEntriesWithConfig which will register these entries to
// global context automatically.
func RegisterAppInfoEntry(opts ...AppInfoEntryOption) *AppInfoEntry {
	entry := &AppInfoEntry{
		AppName:     AppNameDefault,
		Version:     VersionDefault,
		Lang:        LangDefault,
		Description: DescriptionDefault,
		Keywords:    []string{},
		HomeURL:     "",
		IconURL:     "",
		DocsURL:     []string{},
		Maintainers: []string{},
	}

	for i := range opts {
		opts[i](entry)
	}

	// override elements which should not be nil
	if len(entry.Keywords) < 1 {
		entry.Keywords = []string{}
	}

	if len(entry.DocsURL) < 1 {
		entry.DocsURL = []string{}
	}

	if len(entry.Maintainers) < 1 {
		entry.Maintainers = []string{}
	}

	// override elements which should not be empty
	if len(entry.AppName) < 1 {
		entry.AppName = AppNameDefault
	}

	if len(entry.Version) < 1 {
		entry.Version = VersionDefault
	}

	if len(entry.Lang) < 1 {
		entry.Lang = LangDefault
	}

	if len(entry.Description) < 1 {
		entry.Description = DescriptionDefault
	}

	GlobalAppCtx.addAppInfoEntry(entry)

	// override default event logger entry in order to use correct application name.
	// this is special case for default event logger entry.
	eventLoggerConfig := GlobalAppCtx.GetEventLoggerEntryDefault().loggerConfig
	eventLogger, _ := eventLoggerConfig.Build()
	eventLoggerEntry := RegisterEventLoggerEntry(
		WithNameEvent(DefaultEventLoggerEntryName),
		WithEventFactoryEvent(
			rkquery.NewEventFactory(
				rkquery.WithLogger(eventLogger),
				rkquery.WithAppName(entry.AppName))))

	eventLoggerEntry.loggerConfig = eventLoggerConfig

	return entry
}

// No op
func (entry *AppInfoEntry) Bootstrap(context.Context) {
	// no op
}

// No op
func (entry *AppInfoEntry) Interrupt(context.Context) {
	// no op
}

// Return name of entry
func (entry *AppInfoEntry) GetName() string {
	return AppInfoEntryName
}

// Return type of entry
func (entry *AppInfoEntry) GetType() string {
	return AppInfoEntryType
}

// Return string of entry
func (entry *AppInfoEntry) String() string {
	m := map[string]interface{}{
		"entry_name":  AppInfoEntryName,
		"entry_type":  AppInfoEntryType,
		"app_name":    entry.AppName,
		"version":     entry.Version,
		"lang":        entry.Lang,
		"description": entry.Description,
		"home_url":    entry.HomeURL,
		"icon_url":    entry.IconURL,
		"keywords":    entry.Keywords,
		"docs_url":    entry.DocsURL,
		"maintainers": entry.Maintainers,
	}

	// add process info
	for k, v := range rkcommon.ConvertStructToMap(NewProcessInfo()) {
		m[k] = v
	}

	if bytes, err := json.Marshal(&m); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}
