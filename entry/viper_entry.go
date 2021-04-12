// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/spf13/viper"
	"os"
	"path"
	"strings"
)

const (
	ViperEntryType = "viper-config"
)

// Bootstrap config of Zap Logger information.
// 1: Viper.Name: Name of viper entry.
// 2: Viper.Path: File path of config file, could be either relative or absolute path.
//                If relative path was provided, then current working directory would be joined as prefix.
type BootConfigViper struct {
	Viper []struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	} `yaml:"viper"`
}

// ViperEntry contains bellow fields.
// 1: entryName: Name of entry.
// 2: entryType: Type of entry which is ViperEntryType.
// 3: path: File path of config file, could be either relative or absolute path.
//          If relative path was provided, then current working directory would be joined as prefix.
// 4: vp: Viper instance, see viper.Viper for details.
type ViperEntry struct {
	entryName string
	entryType string
	path      string
	vp        *viper.Viper
}

// ViperEntry Option which used while registering entry from codes.
type ViperEntryOption func(*ViperEntry)

// Provide name of entry.
func WithNameViper(name string) ViperEntryOption {
	return func(entry *ViperEntry) {
		if len(name) > 0 {
			entry.entryName = name
		}
	}
}

// Provide viper instance of entry.
func WithViperInstanceViper(vp *viper.Viper) ViperEntryOption {
	return func(entry *ViperEntry) {
		if vp != nil {
			entry.vp = vp
		}
	}
}

// Create viper entries with config file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterViperEntriesWithConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: unmarshal user provided config into boot config struct
	config := &BootConfigViper{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)
	for i := range config.Viper {
		element := config.Viper[i]

		// skip this element name is empty
		if len(element.Name) < 1 {
			continue
		}

		// join user provided path with working directory if it is relative path
		if !path.IsAbs(element.Path) {
			if wd, err := os.Getwd(); err != nil {
				rkcommon.ShutdownWithError(err)
			} else {
				element.Path = path.Join(wd, element.Path)
			}
		}

		// check DOMAIN in environment variable
		domain := os.Getenv("DOMAIN")
		if len(domain) > 0 {
			tokens := strings.Split(path.Base(element.Path), ".")
			// the size of token slice should be two
			if len(tokens) == 2 {
				newPathWithDomain := path.Join(
					path.Dir(element.Path),
					fmt.Sprintf("%s-%s.%s", tokens[0], domain, tokens[1]))

				if rkcommon.FileExists(newPathWithDomain) {
					element.Path = newPathWithDomain
				}
			}
		}

		// skip this element if path is not valid
		if !rkcommon.FileExists(element.Path) {
			continue
		}

		vp := viper.New()
		vp.SetConfigFile(element.Path)

		if err := vp.ReadInConfig(); err != nil {
			continue
		}

		entry := RegisterViperEntry(
			WithNameViper(element.Name),
			WithViperInstanceViper(vp))

		entry.path = element.Path

		res[element.Name] = entry
	}

	return res
}

// Crate viper entry with options.
func RegisterViperEntry(opts ...ViperEntryOption) *ViperEntry {
	entry := &ViperEntry{
		entryType: ViperEntryType,
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.vp == nil {
		entry.vp = viper.New()
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "viper-" + rkcommon.RandString(4)
	}

	GlobalAppCtx.AddViperEntry(entry)

	return entry
}

// Bootstrap entry.
func (entry *ViperEntry) Bootstrap(context.Context) {
	// no op
}

// Interrupt entry.
func (entry *ViperEntry) Interrupt(context.Context) {
	// no op
}

// Get name of entry.
func (entry *ViperEntry) GetName() string {
	return entry.entryName
}

// Get type of entry.
func (entry *ViperEntry) GetType() string {
	return entry.entryType
}

// Convert entry into JSON style string.
func (entry *ViperEntry) String() string {
	m := map[string]interface{}{
		"entry_name": entry.entryName,
		"entry_type": entry.entryType,
		"path":       entry.path,
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

// Get viper instance.
func (entry *ViperEntry) GetViper() *viper.Viper {
	return entry.vp
}
