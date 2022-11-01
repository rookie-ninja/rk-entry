package rkentry

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
)

type BootPProf struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path" json:"path"`
}

type PProfEntry struct {
	entryName        string `json:"-" yaml:"-"`
	entryType        string `json:"-" yaml:"-"`
	entryDescription string `json:"-" yaml:"-"`
	Path             string `json:"-" yaml:"-"`
}

func (entry *PProfEntry) Bootstrap(ctx context.Context) {}

func (entry *PProfEntry) Interrupt(ctx context.Context) {}

func (entry *PProfEntry) GetName() string {
	return entry.entryName
}

func (entry *PProfEntry) GetType() string {
	return entry.entryType
}

func (entry *PProfEntry) GetDescription() string {
	return entry.entryDescription
}

func (entry *PProfEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry
func (entry *PProfEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.GetName(),
		"type":        entry.GetType(),
		"description": entry.GetDescription(),
		"path":        entry.Path,
	}

	return json.Marshal(m)
}

// UnmarshalJSON Unmarshal entry
func (entry *PProfEntry) UnmarshalJSON([]byte) error {
	return nil
}

type PProfEntryOption func(entry *PProfEntry)

func WithNamePProfEntry(name string) PProfEntryOption {
	return func(entry *PProfEntry) {
		entry.entryName = name
	}
}

// RegisterPProfEntry Create new pprof entry with config
func RegisterPProfEntry(boot *BootPProf, opts ...PProfEntryOption) *PProfEntry {
	if !boot.Enabled {
		return nil
	}

	entry := &PProfEntry{
		entryName:        "PProfEntry",
		entryType:        PProfEntryType,
		entryDescription: "Internal RK entry for pprof.",
		Path:             boot.Path,
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.Path) < 1 {
		entry.Path = "/pprof"
	}

	entry.Path = filepath.Join("/", entry.Path)
	if !strings.HasSuffix(entry.Path, "/") {
		entry.Path = entry.Path + "/"
	}

	return entry
}
