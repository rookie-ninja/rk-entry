package rk

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/http/pprof"
	"path"
)

type PprofConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		UrlPrefix string `yaml:"urlPrefix"`
	} `yaml:"entry"`
}

func (p *PprofConfig) JSON() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p *PprofConfig) YAML() string {
	b, _ := yaml.Marshal(p)
	return string(b)
}

func (p *PprofConfig) Header() *EntryConfigHeader {
	return &p.EntryConfigHeader
}

func (p *PprofConfig) Register() (Entry, error) {
	if !p.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(p.Metadata.Domain) {
		return nil, nil
	}

	entry := &PprofEntry{
		config: p,
	}

	if len(p.Entry.UrlPrefix) < 1 {
		p.Entry.UrlPrefix = "/pprof"
	}

	p.Entry.UrlPrefix = path.Join("/", p.Entry.UrlPrefix, "/")

	Registry.AddEntry(entry)

	return entry, nil
}

type PprofEntry struct {
	config *PprofConfig
}

func (p *PprofEntry) Category() string {
	return CategoryIndependent
}

func (p *PprofEntry) Kind() string {
	return p.config.Kind
}

func (p *PprofEntry) Name() string {
	return p.config.Metadata.Name
}

func (p *PprofEntry) Config() EntryConfig {
	return p.config
}

func (p *PprofEntry) Bootstrap(ctx context.Context) {}

func (p *PprofEntry) Interrupt(ctx context.Context) {}

func (p *PprofEntry) Monitor() *Monitor {
	return nil
}

func (p *PprofEntry) FS() *embed.FS {
	return Registry.EntryFS(p.Kind(), p.Name())
}

func (p *PprofEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)

	res = append(res,
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix),
			Handler: pprof.Index,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "cmdline"),
			Handler: pprof.Cmdline,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "profile"),
			Handler: pprof.Profile,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "symbol"),
			Handler: pprof.Symbol,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "trace"),
			Handler: pprof.Trace,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "allocs"),
			Handler: pprof.Handler("allocs").ServeHTTP,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "block"),
			Handler: pprof.Handler("block").ServeHTTP,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "goroutine"),
			Handler: pprof.Handler("goroutine").ServeHTTP,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "heap"),
			Handler: pprof.Handler("heap").ServeHTTP,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "mutex"),
			Handler: pprof.Handler("mutex").ServeHTTP,
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.UrlPrefix, "threadcreate"),
			Handler: pprof.Handler("threadcreate").ServeHTTP,
		},
	)

	return res
}
