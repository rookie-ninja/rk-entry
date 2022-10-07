package rk

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	EntryConfigHeader `yaml:",inline"`
}

func (s *ServiceConfig) JSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *ServiceConfig) YAML() string {
	b, _ := yaml.Marshal(s)
	return string(b)
}

func (s *ServiceConfig) Header() *EntryConfigHeader {
	return &s.EntryConfigHeader
}

func (s *ServiceConfig) Register() (Entry, error) {
	if !s.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(s.Metadata.Domain) {
		return nil, nil
	}

	entry := &ServiceEntry{
		config: s,
	}

	Registry.serviceName = s.Metadata.Name
	Registry.serviceVersion = s.Metadata.Version

	return entry, nil
}

type ServiceEntry struct {
	config *ServiceConfig
}

func (s *ServiceEntry) Category() string {
	return CategoryIndependent
}

func (s *ServiceEntry) Kind() string {
	return s.config.Kind
}

func (s *ServiceEntry) Name() string {
	return s.config.Metadata.Name
}

func (s *ServiceEntry) Config() EntryConfig {
	return s.config
}

func (s *ServiceEntry) Bootstrap(ctx context.Context) {}

func (s *ServiceEntry) Interrupt(ctx context.Context) {}

func (s *ServiceEntry) Monitor() *Monitor {
	return nil
}

func (s *ServiceEntry) FS() *embed.FS {
	return Registry.EntryFS(s.Kind(), s.Name())
}

func (s *ServiceEntry) Apis() []*BuiltinApi {
	return make([]*BuiltinApi, 0)
}
