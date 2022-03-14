package rkentry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

// WithRegistryPromEntry provide prometheus.Registry
func WithRegistryPromEntry(registry *prometheus.Registry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Registry = registry
	}
}

// RegisterPromEntry Create a prom entry with options and add prom entry to rkentry.GlobalAppCtx
func RegisterPromEntry(boot *BootProm, opts ...PromEntryOption) *PromEntry {
	if !boot.Enabled {
		return nil
	}

	entry := &PromEntry{
		entryName:        "PromEntry",
		entryType:        PromEntryType,
		entryDescription: "Internal RK entry which implements prometheus client.",
		Path:             boot.Path,
		Registerer:       prometheus.DefaultRegisterer,
		Gatherer:         prometheus.DefaultGatherer,
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.Registry == nil {
		entry.Registry = prometheus.NewRegistry()
	}
	entry.Registry.Register(collectors.NewGoCollector())

	if entry.Registry != nil {
		entry.Registerer = entry.Registry
		entry.Gatherer = entry.Registry
	}

	// Trim space by default
	entry.Path = strings.TrimSpace(entry.Path)

	if len(entry.Path) < 1 {
		// Invalid path, use default one
		entry.Path = "/metrics"
	}

	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	entry.Pusher = newPushGatewayPusher(boot, entry.Gatherer)

	return entry
}

// BootProm Boot config which is for prom entry.
type BootProm struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path" json:"path"`
	Pusher  struct {
		Enabled       bool   `yaml:"enabled" json:"enabled"`
		IntervalMs    int64  `yaml:"IntervalMs" json:"IntervalMs"`
		JobName       string `yaml:"jobName" json:"jobName"`
		RemoteAddress string `yaml:"remoteAddress" json:"remoteAddress"`
		BasicAuth     string `yaml:"basicAuth" json:"basicAuth"`
		CertEntry     string `yaml:"certEntry" json:"certEntry"`
		LoggerEntry   string `yaml:"loggerEntry" json:"loggerEntry"`
	} `yaml:"pusher" json:"pusher"`
}

// PromEntry Prometheus entry which implements rkentry.Entry.
type PromEntry struct {
	*prometheus.Registry
	prometheus.Registerer
	prometheus.Gatherer

	entryName        string             `json:"-" yaml:"-"`
	entryType        string             `json:"-" yaml:"-"`
	entryDescription string             `json:"-" yaml:"-"`
	Path             string             `json:"-" yaml:"-"`
	Pusher           *PushGatewayPusher `json:"-" yaml:"-"`
}

type PromEntryOption func(entry *PromEntry)

// Bootstrap Start prometheus client
func (entry *PromEntry) Bootstrap(ctx context.Context) {
	// start pusher
	if entry.Pusher != nil {
		entry.Pusher.Bootstrap(ctx)
	}
}

// Interrupt Shutdown prometheus client
func (entry *PromEntry) Interrupt(ctx context.Context) {
	if entry.Pusher != nil {
		entry.Pusher.Interrupt(ctx)
	}
}

// GetName Return name of prom entry
func (entry *PromEntry) GetName() string {
	return entry.entryName
}

// GetType Return type of prom entry
func (entry *PromEntry) GetType() string {
	return entry.entryType
}

// GetDescription Get description of entry
func (entry *PromEntry) GetDescription() string {
	return entry.entryDescription
}

// String Stringfy prom entry
func (entry *PromEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry
func (entry *PromEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":              entry.entryName,
		"type":              entry.entryType,
		"description":       entry.entryDescription,
		"pushGateWayPusher": entry.Pusher,
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Unmarshal entry
func (entry *PromEntry) UnmarshalJSON([]byte) error {
	return nil
}

// RegisterCollectors Register collectors in default registry
func (entry *PromEntry) RegisterCollectors(collectors ...prometheus.Collector) {
	for i := range collectors {
		entry.Registerer.Register(collectors[i])
	}
}

// PushGatewayPusher is a pusher which contains bellow instances
type PushGatewayPusher struct {
	loggerEntry   *LoggerEntry  `json:"-" yaml:"-"`
	Pusher        *push.Pusher  `json:"-" yaml:"-"`
	IntervalMs    time.Duration `json:"-" yaml:"-"`
	RemoteAddress string        `json:"-" yaml:"-"`
	JobName       string        `json:"-" yaml:"-"`
	running       *atomic.Bool  `json:"-" yaml:"-"`
	certEntry     *CertEntry    `json:"-" yaml:"-"`
}

// newPushGatewayPusher creates a new pushGateway periodic job instances with intervalMS, remote URL and job name
func newPushGatewayPusher(boot *BootProm, gatherer prometheus.Gatherer) *PushGatewayPusher {
	if !boot.Pusher.Enabled {
		return nil
	}

	certEntry := GlobalAppCtx.GetCertEntry(boot.Pusher.CertEntry)

	pg := &PushGatewayPusher{
		IntervalMs:    time.Duration(boot.Pusher.IntervalMs) * time.Millisecond,
		JobName:       boot.Pusher.JobName,
		RemoteAddress: boot.Pusher.RemoteAddress,
		running:       atomic.NewBool(false),
		certEntry:     certEntry,
	}

	if pg.IntervalMs < 1 {
		pg.IntervalMs = 5 * time.Second
	}

	if len(pg.RemoteAddress) < 1 {
		pg.RemoteAddress = "http://localhost:9091"
	}

	if len(pg.JobName) < 1 {
		pg.JobName = "rk"
	}

	pg.loggerEntry = GlobalAppCtx.GetLoggerEntry(boot.Pusher.LoggerEntry)
	if pg.loggerEntry == nil {
		pg.loggerEntry = LoggerEntryStdout
	}

	pg.Pusher = push.New(pg.RemoteAddress, pg.JobName)

	// assign credential of basic auth
	if len(boot.Pusher.BasicAuth) > 0 && strings.Contains(boot.Pusher.BasicAuth, ":") {
		boot.Pusher.BasicAuth = strings.TrimSpace(boot.Pusher.BasicAuth)
		tokens := strings.Split(boot.Pusher.BasicAuth, ":")
		if len(tokens) == 2 {
			pg.Pusher = pg.Pusher.BasicAuth(tokens[0], tokens[1])
		}
	}

	pg.Pusher.Gatherer(gatherer)

	return pg
}

// Bootstrap starts a periodic job
func (pub *PushGatewayPusher) Bootstrap(ctx context.Context) {
	httpClient := http.DefaultClient

	var conf *tls.Config
	if pub.certEntry != nil {
		if pub.certEntry.Certificate != nil {
			conf = &tls.Config{}
			conf.Certificates = []tls.Certificate{
				*pub.certEntry.Certificate,
			}
		}

		if pub.certEntry.RootCA != nil {
			caCert := x509.NewCertPool()
			caCert.AddCert(pub.certEntry.RootCA)

			if conf != nil {
				conf.RootCAs = caCert
			} else {
				conf = &tls.Config{}
				conf.RootCAs = caCert
			}
		}
	}

	if conf != nil {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: conf,
		}
	}

	pub.Pusher.Client(httpClient)

	pub.running.CAS(false, true)

	pub.loggerEntry.Info("Starting pushGateway publisher",
		zap.String("remoteAddress", pub.RemoteAddress),
		zap.String("jobName", pub.JobName))

	go pub.push()
}

// Interrupt stops periodic job
func (pub *PushGatewayPusher) Interrupt(ctx context.Context) {
	pub.running.CAS(true, false)
}

// Internal use only
func (pub *PushGatewayPusher) push() {
	for pub.running.Load() {
		err := pub.Pusher.Push()

		if err != nil {
			pub.loggerEntry.Warn("Failed to push metrics to PushGateway",
				zap.String("remoteAddress", pub.RemoteAddress),
				zap.String("jobName", pub.JobName),
				zap.Error(err))
		}

		time.Sleep(pub.IntervalMs)
	}
}
