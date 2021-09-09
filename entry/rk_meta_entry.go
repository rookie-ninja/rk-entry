// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	"os"
	"path"
)

const (
	// RkMetaEntryName name of entry
	RkMetaEntryName = "RkMetaDefault"
	// RkMetaEntryType type of entry
	RkMetaEntryType = "RkMetaEntry"
	// RkMetaEntryDescription description of entry
	RkMetaEntryDescription = "Internal RK entry which describes rk metadata."
)

// BootConfigRkMeta bootstrap config of application's meta information.
type BootConfigRkMeta struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	Git     struct {
		Url    string `yaml:"url" json:"url"`
		Branch string `yaml:"branch" json:"branch"`
		Tag    string `yaml:"tag" json:"tag"`
		Commit struct {
			Id        string `yaml:"id" json:"id"`
			IdAbbr    string `yaml:"idAbbr" json:"idAbbr"`
			Date      string `yaml:"date" json:"date"`
			Sub       string `yaml:"sub" json:"sub"`
			Committer struct {
				Name  string `yaml:"name" json:"name"`
				Email string `yaml:"email" json:"email"`
			} `yaml:"committer" json:"committer"`
		} `yaml:"commit" json:"commit"`
	} `yaml:"git" json:"git"`
}

// RkMetaEntry contains bellow fields.
type RkMetaEntry struct {
	EntryName        string           `json:"entryName" yaml:"entryName"`
	EntryType        string           `json:"entryType" yaml:"entryType"`
	EntryDescription string           `json:"entryDescription" yaml:"entryDescription"`
	RkMeta           *rkcommon.RkMeta `json:"meta" yaml:"meta"`
}

// RkMetaEntryOption Option which used while registering entry from codes.
type RkMetaEntryOption func(*RkMetaEntry)

// WithMetaRkMeta provide git information.
func WithMetaRkMeta(meta *rkcommon.RkMeta) RkMetaEntryOption {
	return func(entry *RkMetaEntry) {
		entry.RkMeta = meta
	}
}

// RegisterRkMetaEntriesFromConfig implements rkentry.EntryRegFunc which generate RKEntry based on boot configuration file.
func RegisterRkMetaEntriesFromConfig(string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: Unmarshal user provided config into boot config struct
	config := &BootConfigRkMeta{}

	var meta *rkcommon.RkMeta

	// We will looking for .rk/rk.yaml file in current working directory
	//
	// In case, there are no .rk/rk.yaml file in working directory, we will try to call local command
	// to full fill entry.
	//
	// For example, if we run main.go files in IDE or directory with command line without rk build command.
	// join the path with current working directory if user provided path is relative path
	wd, _ := os.Getwd()
	filePath := path.Join(wd, rkcommon.RkMetaFilePath)
	if rkcommon.FileExists(filePath) {
		rkcommon.UnmarshalBootConfig(filePath, config)
		meta = &rkcommon.RkMeta{
			Name:    config.Name,
			Version: config.Version,
			Git: &rkcommon.Git{
				Url:    config.Git.Url,
				Branch: config.Git.Branch,
				Tag:    config.Git.Tag,
				Commit: &rkcommon.Commit{
					Id:     config.Git.Commit.Id,
					IdAbbr: config.Git.Commit.IdAbbr,
					Date:   config.Git.Commit.Date,
					Sub:    config.Git.Commit.Sub,
					Committer: &rkcommon.Committer{
						Name:  config.Git.Commit.Committer.Name,
						Email: config.Git.Commit.Committer.Email,
					},
				},
			},
		}
	} else {
		meta = rkcommon.GetRkMetaFromCmd()
	}

	// 2: Init rk entry from config
	entry := RegisterRkMetaEntry(
		WithMetaRkMeta(meta))

	res[RkMetaEntryName] = entry

	return res
}

// RegisterRkMetaEntry register Entry with options.
// This function is used while creating entry from code instead of config file.
// We will override RKEntry fields if value is nil or empty if necessary.
//
// Generally, we recommend call rkctx.GlobalAppCtx.AddEntry() inside this function,
// however, we recommend to register RKEntry, ZapLoggerEntry, EventLoggerEntry with
// function of rkctx.RegisterBasicEntriesWithConfig which will register these entries to
// global context automatically.
func RegisterRkMetaEntry(opts ...RkMetaEntryOption) *RkMetaEntry {
	entry := &RkMetaEntry{
		EntryName:        RkMetaEntryName,
		EntryType:        RkMetaEntryType,
		EntryDescription: RkMetaEntryDescription,
	}

	for i := range opts {
		opts[i](entry)
	}

	GlobalAppCtx.SetRkMetaEntry(entry)

	return entry
}

// Bootstrap No op.
func (entry *RkMetaEntry) Bootstrap(context.Context) {
	// No op
}

// Interrupt No op.
func (entry *RkMetaEntry) Interrupt(context.Context) {
	// No op
}

// GetName return name of entry.
func (entry *RkMetaEntry) GetName() string {
	return entry.EntryName
}

// GetType return type of entry.
func (entry *RkMetaEntry) GetType() string {
	return entry.EntryType
}

// GetDescription return description of entry.
func (entry *RkMetaEntry) GetDescription() string {
	return entry.EntryDescription
}

// String return string of entry.
func (entry *RkMetaEntry) String() string {
	var bytes []byte
	var err error
	if bytes, err = json.Marshal(entry); err != nil {
		return "{}"
	}

	return string(bytes)
}
