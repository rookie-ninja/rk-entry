package rk

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap/zapgrpc"
	"gopkg.in/yaml.v3"
	"net/http"
	"path"
	"time"
)

type PrometheusConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		Url                 string `yaml:"url"`
		ErrorHandling       string `yaml:"errorHandling"`
		DisableCompression  bool   `yaml:"disableCompression"`
		MaxRequestsInFlight int    `yaml:"maxRequestsInFlight"`
		Timeout             string `yaml:"timeout"`
		EnableOpenMetrics   bool   `yaml:"enableOpenMetrics"`
		Logger              struct {
			Kind string `yaml:"kind"`
			Name string `yaml:"name"`
		} `yaml:"logger"`
	} `yaml:"entry"`
}

func (p *PrometheusConfig) JSON() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p *PrometheusConfig) YAML() string {
	b, _ := yaml.Marshal(p)
	return string(b)
}

func (p *PrometheusConfig) Header() *EntryConfigHeader {
	return &p.EntryConfigHeader
}

func (p *PrometheusConfig) Register() (Entry, error) {
	if !p.Metadata.Enabled {
		return nil, nil
	}

	reg := prometheus.NewRegistry()

	errorHandling := promhttp.HTTPErrorOnError
	switch p.Entry.ErrorHandling {
	case "ContinueOnError":
		errorHandling = promhttp.ContinueOnError
	case "PanicOnError":
		errorHandling = promhttp.PanicOnError
	}

	timeout := time.Duration(0)
	if len(p.Entry.Timeout) > 0 {
		v, err := time.ParseDuration(p.Entry.Timeout)
		if err != nil {
			return nil, err
		}
		timeout = v
	}

	loggerEntry := Registry.GetEntryOrDefault(p.Entry.Logger.Kind, p.Entry.Logger.Name)

	var logger promhttp.Logger
	if loggerEntry == nil {
		if v, ok := loggerEntry.(*ZapEntry); ok {
			logger = zapgrpc.NewLogger(v.Logger)
		}
	}

	entry := &PrometheusEntry{
		config:   p,
		Registry: reg,
		WebHandler: func(writer http.ResponseWriter, request *http.Request) {
			promhttp.HandlerFor(reg, promhttp.HandlerOpts{
				ErrorLog:            logger,
				ErrorHandling:       errorHandling,
				Registry:            reg,
				DisableCompression:  p.Entry.DisableCompression,
				MaxRequestsInFlight: p.Entry.MaxRequestsInFlight,
				Timeout:             timeout,
				EnableOpenMetrics:   p.Entry.EnableOpenMetrics,
			}).ServeHTTP(writer, request)
		},
	}

	if len(p.Entry.Url) < 1 {
		p.Entry.Url = "/metrics"
	}
	p.Entry.Url = path.Join("/", p.Entry.Url)

	Registry.AddEntry(entry)

	return entry, nil
}

type PrometheusEntry struct {
	config     *PrometheusConfig
	Registry   *prometheus.Registry
	WebHandler http.HandlerFunc
}

func (p *PrometheusEntry) Bootstrap(ctx context.Context) {
	return
}

func (p *PrometheusEntry) Interrupt(ctx context.Context) {
	return
}

func (p *PrometheusEntry) FS() *embed.FS {
	return Registry.EntryFS(p.Kind(), p.Name())
}

func (p *PrometheusEntry) Category() string {
	return CategoryPlugin
}

func (p *PrometheusEntry) Kind() string {
	return p.config.Kind
}

func (p *PrometheusEntry) Name() string {
	return p.config.Metadata.Name
}

func (p *PrometheusEntry) Config() EntryConfig {
	return p.config
}

func (p *PrometheusEntry) Monitor() *Monitor {
	return nil
}

func (p *PrometheusEntry) Apis() []*BuiltinApi {
	res := []*BuiltinApi{
		{
			Method:  http.MethodGet,
			Path:    path.Join("/", p.config.Entry.Url),
			Handler: p.WebHandler,
		},
	}

	return res
}
