// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// RegisterConfigEntry create ConfigEntry with BootConfigConfig.
func RegisterConfigEntry(boot *BootConfig) []*ConfigEntry {
	res := make([]*ConfigEntry, 0)

	// filter out based domain
	configMap := make(map[string]*BootConfigE)
	for _, config := range boot.Config {
		if len(config.Name) < 1 {
			continue
		}

		if !IsValidDomain(config.Domain) {
			continue
		}

		// * or matching domain
		// 1: add it to map if missing
		if _, ok := configMap[config.Name]; !ok {
			configMap[config.Name] = config
			continue
		}

		// 2: already has an entry, then compare domain,
		//    only one case would occur, previous one is already the correct one, continue
		if config.Domain == "" || config.Domain == "*" {
			continue
		}

		configMap[config.Name] = config
	}

	for _, config := range configMap {
		entry := &ConfigEntry{
			entryName:        config.Name,
			entryType:        ConfigEntryType,
			entryDescription: config.Description,
			content:          config.Content,
			Viper:            viper.New(),
			Path:             config.Path,
			EnvPrefix:        config.EnvPrefix,
		}

		// if file path was provided
		if len(entry.Path) > 0 {
			if !filepath.IsAbs(entry.Path) {
				if wd, err := os.Getwd(); err != nil {
					ShutdownWithError(err)
				} else {
					entry.Path = filepath.ToSlash(filepath.Join(wd, entry.Path))
				}
			}

			// skip this element if path is not valid
			if fileExists(entry.Path) {
				entry.Viper.SetConfigFile(entry.Path)
				if err := entry.Viper.ReadInConfig(); err != nil {
					ShutdownWithError(fmt.Errorf("failed to read file, path:%s", entry.Path))
				}
			}
		}

		// if content exist, then fill viper
		for k, v := range entry.content {
			entry.Viper.Set(k, v)
		}

		// enable automatic env
		// issue: https://github.com/rookie-ninja/rk-boot/issues/55
		entry.Viper.AutomaticEnv()
		entry.Viper.SetEnvPrefix(entry.EnvPrefix)

		GlobalAppCtx.AddEntry(entry)
		res = append(res, entry)
	}

	return res
}

// RegisterConfigEntryYAML register function
func RegisterConfigEntryYAML(raw []byte) map[string]Entry {
	boot := &BootConfig{}
	UnmarshalBootYAML(raw, boot)

	res := map[string]Entry{}

	entries := RegisterConfigEntry(boot)
	for i := range entries {
		entry := entries[i]
		res[entry.GetName()] = entry
	}

	return res
}

// BootConfig is bootstrap config of ConfigEntry information.
type BootConfig struct {
	Config []*BootConfigE `yaml:"config" json:"config"`
}

// BootConfigE element of ConfigEntry
type BootConfigE struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Domain      string                 `yaml:"domain" json:"domain"`
	Path        string                 `yaml:"path" json:"name"`
	EnvPrefix   string                 `yaml:"envPrefix" json:"envPrefix"`
	Content     map[string]interface{} `yaml:"content" json:"content"`
}

// ConfigEntry contains bellow fields.
type ConfigEntry struct {
	*viper.Viper

	entryName        string                 `yaml:"-" json:"-"`
	entryType        string                 `yaml:"-" json:"-"`
	entryDescription string                 `yaml:"-" json:"-"`
	Locale           string                 `yaml:"-" json:"-"`
	Path             string                 `yaml:"-" json:"-"`
	EnvPrefix        string                 `yaml:"-" json:"-"`
	content          map[string]interface{} `yaml:"-" json:"-"`
}

// Bootstrap entry.
func (entry *ConfigEntry) Bootstrap(context.Context) {}

// Interrupt entry.
func (entry *ConfigEntry) Interrupt(context.Context) {}

// GetName returns name of entry.
func (entry *ConfigEntry) GetName() string {
	return entry.entryName
}

// GetType returns type of entry.
func (entry *ConfigEntry) GetType() string {
	return entry.entryType
}

// String convert entry into JSON style string.
func (entry *ConfigEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON marshal entry.
func (entry *ConfigEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.GetName(),
		"type":        entry.GetType(),
		"description": entry.GetDescription(),
		"locale":      entry.Locale,
		"path":        entry.Path,
		"envPrefix":   entry.EnvPrefix,
	}

	return json.Marshal(m)
}

// UnmarshalJSON is not supported.
func (entry *ConfigEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetDescription return description of entry.
func (entry *ConfigEntry) GetDescription() string {
	return entry.entryDescription
}
