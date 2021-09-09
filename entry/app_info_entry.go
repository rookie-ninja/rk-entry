// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkentry defines inner entries for rk-boot.
package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-query"
	"path"
)

const (
	// AppNameDefault will be used if not provided by user
	AppNameDefault = "rk"
	// VersionDefault will be empty if not provided by user
	VersionDefault = ""
	// LangDefault will always be golang
	LangDefault = "golang"
	// AppInfoEntryName is a fixed value
	AppInfoEntryName = "AppInfoDefault"
	// AppInfoEntry is a fixed value
	AppInfoEntryType = "AppInfoEntry"
	// AppInfoEntryDescription is a fixed value
	AppInfoEntryDescription = "Internal RK entry which describes application with fields of appName, version and etc."
)

// BootConfigAppInfo is config of application's basic information.
// 1: Description: Description of application itself.
// 2: Keywords: A set of words describe application.
// 3: HomeUrl: Home page URL.
// 4: IconUrl: Application Icon URL.
// 5: Maintainers: Maintainers of application.
// 6: DocsUrl: A set of URLs of documentations of application.
type BootConfigAppInfo struct {
	App struct {
		Description string   `yaml:"description" json:"description"`
		Keywords    []string `yaml:"keywords" json:"keywords"`
		HomeUrl     string   `yaml:"homeUrl" json:"homeUrl"`
		IconUrl     string   `yaml:"iconUrl" json:"iconUrl"`
		DocsUrl     []string `yaml:"docsUrl" json:"docsUrl"`
		Maintainers []string `yaml:"maintainers" json:"maintainers"`
	} `yaml:"app"`
}

// AppInfoEntry contains bellow fields.
// 1: AppName: Application name which refers to go process
// 2: Version: Application version
// 3: Lang: Programming language <NOT configurable!>
// 4: EntryDescription: Description of application itself
// 5: Keywords: A set of words describe application
// 6: HomeUrl: Home page URL
// 7: IconUrl: Application Icon URL
// 8: DocsUrl: A set of URLs of documentations of application
// 9: Maintainers: Maintainers of application
// 10: License: License of current application.
// 11: Readme: README.md file of current application.
// 11: GoMod: go.mod file of current application.
// 12: UtHtml: cov.html file of current application.
type AppInfoEntry struct {
	EntryName        string   `json:"entryName" yaml:"entryName"`
	EntryType        string   `json:"entryType" yaml:"entryType"`
	EntryDescription string   `json:"description" yaml:"description"`
	AppName          string   `json:"appName" yaml:"appName"`
	Version          string   `json:"version" yaml:"version"`
	Lang             string   `json:"lang" yaml:"lang"`
	Keywords         []string `json:"keywords" yaml:"keywords"`
	HomeUrl          string   `json:"homeUrl" yaml:"homeUrl"`
	IconUrl          string   `json:"iconUrl" yaml:"iconUrl"`
	DocsUrl          []string `json:"docsUrl" yaml:"docsUrl"`
	Maintainers      []string `json:"maintainers" yaml:"maintainers"`
	License          string   `json:"-" yaml:"-"`
	Readme           string   `json:"-" yaml:"-"`
	GoMod            string   `json:"-" yaml:"-"`
	UtHtml           string   `json:"-" yaml:"-"`
}

// AppInfoEntryDefault generate a AppInfo entry with default fields.
func AppInfoEntryDefault() *AppInfoEntry {
	return &AppInfoEntry{
		EntryName:        AppInfoEntryName,
		EntryType:        AppInfoEntryType,
		EntryDescription: AppInfoEntryDescription,
		AppName:          AppNameDefault,
		Version:          VersionDefault,
		Lang:             LangDefault,
		Keywords:         []string{},
		HomeUrl:          "",
		IconUrl:          "",
		DocsUrl:          []string{},
		Maintainers:      []string{},
		License:          "",
		Readme:           "",
		GoMod:            "",
	}
}

// AppInfoEntryOption which used while registering entry from codes.
type AppInfoEntryOption func(*AppInfoEntry)

// WithAppNameAppInfo provide application name.
func WithAppNameAppInfo(name string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.AppName = name
	}
}

// WithVersionAppInfo provide version.
func WithVersionAppInfo(version string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Version = version
	}
}

// WithDescriptionAppInfo provide description.
func WithDescriptionAppInfo(description string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.EntryDescription = description
	}
}

// WithHomeUrlAppInfo provide home page URL.
func WithHomeUrlAppInfo(homeUrl string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.HomeUrl = homeUrl
	}
}

// WithIconUrlAppInfo provide icon URL.
func WithIconUrlAppInfo(iconUrl string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.IconUrl = iconUrl
	}
}

// WithKeywordsAppInfo provide keywords.
func WithKeywordsAppInfo(keywords ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Keywords = append(entry.Keywords, keywords...)
	}
}

// WithDocsUrlAppInfo provide documentation URLs.
func WithDocsUrlAppInfo(docsURL ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.DocsUrl = append(entry.DocsUrl, docsURL...)
	}
}

// WithMaintainersAppInfo provide maintainers.
func WithMaintainersAppInfo(maintainers ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Maintainers = append(entry.Maintainers, maintainers...)
	}
}

// RegisterAppInfoEntriesFromConfig implements rkentry.EntryRegFunc which generate RKEntry based on boot configuration file.
func RegisterAppInfoEntriesFromConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: Unmarshal user provided config into boot config struct
	config := &BootConfigAppInfo{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init rk entry from config
	entry := RegisterAppInfoEntry(
		WithDescriptionAppInfo(config.App.Description),
		WithKeywordsAppInfo(config.App.Keywords...),
		WithHomeUrlAppInfo(config.App.HomeUrl),
		WithIconUrlAppInfo(config.App.IconUrl),
		WithDocsUrlAppInfo(config.App.DocsUrl...),
		WithMaintainersAppInfo(config.App.Maintainers...))

	res[AppInfoEntryName] = entry

	return res
}

// RegisterAppInfoEntry register RKEntry with options.
// This function is used while creating entry from code instead of config file.
// We will override RKEntry fields if value is nil or empty if necessary.
//
// Generally, we recommend call rkctx.GlobalAppCtx.AddEntry() inside this function,
// however, we recommend to register RKEntry, ZapLoggerEntry, EventLoggerEntry with
// function of rkctx.RegisterBasicEntriesWithConfig which will register these entries to
// global context automatically.
func RegisterAppInfoEntry(opts ...AppInfoEntryOption) *AppInfoEntry {
	entry := &AppInfoEntry{
		EntryName:        AppInfoEntryName,
		EntryType:        AppInfoEntryType,
		EntryDescription: AppInfoEntryDescription,
		AppName:          AppNameDefault,
		Version:          VersionDefault,
		Lang:             LangDefault,
		Keywords:         []string{},
		HomeUrl:          "",
		IconUrl:          "",
		DocsUrl:          []string{},
		Maintainers:      []string{},
		License:          "",
		Readme:           "",
		GoMod:            "",
	}

	for i := range opts {
		opts[i](entry)
	}

	// Read git info to retrieve package name
	rkMetaEntry := GlobalAppCtx.GetRkMetaEntry()
	if rkMetaEntry != nil {
		entry.AppName = rkMetaEntry.RkMeta.Name
		entry.Version = rkMetaEntry.RkMeta.Version
	}

	// Override elements which should not be nil
	if len(entry.Keywords) < 1 {
		entry.Keywords = []string{}
	}

	if len(entry.DocsUrl) < 1 {
		entry.DocsUrl = []string{}
	}

	if len(entry.Maintainers) < 1 {
		entry.Maintainers = []string{}
	}

	// Override elements which should not be empty
	if len(entry.AppName) < 1 {
		entry.AppName = AppNameDefault
	}

	if len(entry.Version) < 1 {
		entry.Version = VersionDefault
	}

	if len(entry.EntryDescription) < 1 {
		entry.EntryDescription = AppInfoEntryDescription
	}

	GlobalAppCtx.SetAppInfoEntry(entry)

	// Override default event logger entry in order to use correct application name.
	// this is special case for default event logger entry.
	eventLoggerConfig := GlobalAppCtx.GetEventLoggerEntryDefault().LoggerConfig
	eventLogger, _ := eventLoggerConfig.Build()
	eventLoggerEntry := RegisterEventLoggerEntry(
		WithNameEvent(DefaultEventLoggerEntryName),
		WithEventFactoryEvent(
			rkquery.NewEventFactory(
				rkquery.WithZapLogger(eventLogger),
				rkquery.WithAppName(entry.AppName),
				rkquery.WithAppVersion(entry.Version))))

	eventLoggerEntry.LoggerConfig = eventLoggerConfig

	return entry
}

// Read rk meta file.
func (entry *AppInfoEntry) readRkMetaFile(filePath string) string {
	var bytes []byte
	// read file from gen/rk directory
	if bytes = rkcommon.TryReadFile(filePath); len(bytes) < 1 {
		// read from current working directory as backoff, since user may run program from IDE directory.
		return string(rkcommon.TryReadFile(path.Base(filePath)))
	}

	return string(bytes)
}

// Bootstrap will read meta files.
func (entry *AppInfoEntry) Bootstrap(context.Context) {
	// read license file
	entry.License = entry.readRkMetaFile(rkcommon.RkLicenseFilePath)
	// read readme file
	entry.Readme = entry.readRkMetaFile(rkcommon.RkReadmeFilePath)
	// read go.mod file
	entry.GoMod = entry.readRkMetaFile(rkcommon.RkDepFilePath)
	// read ut coverage report as HTML
	entry.UtHtml = entry.readRkMetaFile(rkcommon.RkUtHtmlFilePath)
}

// Interrupt is noop function.
func (entry *AppInfoEntry) Interrupt(context.Context) {
	// No op
}

// GetName return name of entry.
func (entry *AppInfoEntry) GetName() string {
	return entry.EntryName
}

// GetType return type of entry.
func (entry *AppInfoEntry) GetType() string {
	return entry.EntryType
}

// GetDescription return description of entry.
func (entry *AppInfoEntry) GetDescription() string {
	return entry.EntryDescription
}

// String return string value of entry.
func (entry *AppInfoEntry) String() string {
	var bytes []byte
	var err error
	if bytes, err = json.Marshal(entry); err != nil {
		return "{}"
	}

	return string(bytes)
}
