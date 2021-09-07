// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/spf13/viper"
	"os"
	"path"
)

const (
	ConfigEntryType = "ConfigEntry" // ConfigEntryType is entry type of ConfigEntry
	// ConfigEntryDescription is default description of ConfigEntry
	ConfigEntryDescription = "Internal RK entry which read user config file into viper instance."
)

// BootConfigConfig is bootstrap config of ConfigEntry information.
// 1: Config.Name: Name of viper entry.
// 2: Config.Description: Description of viper entry.
// 3: Config.Locale: <realm>::<region>::<az>::<domain>
// 4: Config.Path: File path of config file, could be either relative or absolute path.
//                 If relative path was provided, then current working directory would be joined as prefix.
type BootConfigConfig struct {
	Config []struct {
		Name        string `yaml:"name" json:"name"`
		Description string `yaml:"description" json:"description"`
		Locale      string `yaml:"locale" json:"locale"`
		Path        string `yaml:"path" json:"name"`
	} `yaml:"config" json:"config"`
}

// ConfigEntry contains bellow fields.
// 1: EntryName: Name of entry.
// 2: EntryType: Type of entry which is ConfigEntryType.
// 3: EntryDescription: Description of ConfigEntry.
// 4: Locale: <realm>::<region>::<az>::<domain>
// 4: Path: File path of config file, could be either relative or absolute path.
//          If relative path was provided, then current working directory would be joined as prefix.
// 5: vp: Viper instance, see viper.Viper for details.
type ConfigEntry struct {
	EntryName        string       `yaml:"entryName" json:"entryName"`
	EntryType        string       `yaml:"entryType" json:"entryType"`
	EntryDescription string       `yaml:"entryDescription" json:"entryDescription"`
	Locale           string       `yaml:"locale" json:"locale"`
	Path             string       `yaml:"path" json:"path"`
	vp               *viper.Viper `yaml:"-" json:"-"`
}

// ConfigEntryOption which used while registering entry from codes.
type ConfigEntryOption func(*ConfigEntry)

// WithNameConfig provide name of entry.
func WithNameConfig(name string) ConfigEntryOption {
	return func(entry *ConfigEntry) {
		if len(name) > 0 {
			entry.EntryName = name
		}
	}
}

// WithDescriptionConfig provide description of entry.
func WithDescriptionConfig(description string) ConfigEntryOption {
	return func(entry *ConfigEntry) {
		if len(description) > 0 {
			entry.EntryDescription = description
		}
	}
}

// WithLocaleConfig provide description of entry.
func WithLocaleConfig(locale string) ConfigEntryOption {
	return func(entry *ConfigEntry) {
		if len(locale) > 0 {
			entry.Locale = locale
		}
	}
}

// WithPathConfig provide path of entry.
func WithPathConfig(path string) ConfigEntryOption {
	return func(entry *ConfigEntry) {
		entry.Path = path
	}
}

// WithViperInstanceConfig provide viper instance of entry.
func WithViperInstanceConfig(vp *viper.Viper) ConfigEntryOption {
	return func(entry *ConfigEntry) {
		if vp != nil {
			entry.vp = vp
		}
	}
}

// RegisterConfigEntriesWithConfig create config entries with config file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterConfigEntriesWithConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: unmarshal user provided config into boot config struct
	config := &BootConfigConfig{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)
	for i := range config.Config {
		element := config.Config[i]

		if len(element.Name) < 1 || !rkcommon.MatchLocaleWithEnv(element.Locale) {
			continue
		}

		entry := RegisterConfigEntry(
			WithNameConfig(element.Name),
			WithLocaleConfig(element.Locale),
			WithDescriptionConfig(element.Description),
			WithPathConfig(element.Path))

		res[element.Name] = entry
	}

	return res
}

// RegisterConfigEntry create ConfigEntry with options.
func RegisterConfigEntry(opts ...ConfigEntryOption) *ConfigEntry {
	entry := &ConfigEntry{
		EntryType:        ConfigEntryType,
		EntryDescription: ConfigEntryDescription,
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.vp == nil {
		// join user provided path with working directory if it is relative path
		if !path.IsAbs(entry.Path) {
			if wd, err := os.Getwd(); err != nil {
				rkcommon.ShutdownWithError(err)
			} else {
				entry.Path = path.Join(wd, entry.Path)
			}
		}

		entry.vp = viper.New()
		// skip this element if path is not valid
		if rkcommon.FileExists(entry.Path) {
			entry.vp.SetConfigFile(entry.Path)
			if err := entry.vp.ReadInConfig(); err != nil {
				rkcommon.ShutdownWithError(errors.New(fmt.Sprintf("failed to read file, path:%s", entry.Path)))
			}
		}
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "config-" + rkcommon.RandString(4)
	}

	GlobalAppCtx.AddConfigEntry(entry)

	return entry
}

// Bootstrap entry.
func (entry *ConfigEntry) Bootstrap(context.Context) {
	// no op
}

// Interrupt entry.
func (entry *ConfigEntry) Interrupt(context.Context) {
	// no op
}

// GetName returns name of entry.
func (entry *ConfigEntry) GetName() string {
	return entry.EntryName
}

// GetType returns type of entry.
func (entry *ConfigEntry) GetType() string {
	return entry.EntryType
}

// String convert entry into JSON style string.
func (entry *ConfigEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON marshal entry.
func (entry *ConfigEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"locale":           entry.Locale,
		"path":             entry.Path,
		"viper":            rkcommon.GeneralizeMapKeyToString(entry.GetViper().AllSettings()),
	}

	return json.Marshal(&m)
}

// UnmarshalJSON is not supported.
func (entry *ConfigEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetDescription return description of entry.
func (entry *ConfigEntry) GetDescription() string {
	return entry.EntryDescription
}

// GetViper returns viper instance.
func (entry *ConfigEntry) GetViper() *viper.Viper {
	return entry.vp
}

// GetViperAsMap convert values in viper instance into map.
func (entry *ConfigEntry) GetViperAsMap() map[string]interface{} {
	return rkcommon.GeneralizeMapKeyToString(entry.GetViper().AllSettings()).(map[string]interface{})
}

// GetLocale returns locale.
func (entry *ConfigEntry) GetLocale() string {
	return entry.Locale
}
