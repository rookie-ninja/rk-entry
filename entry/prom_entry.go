package rkentry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
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

var (
	// Why 1608? It is the year of first telescope was invented
	defaultPort = uint64(1608)
	defaultPath = "/metrics"
)

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
		Cert          struct {
			RootCertPath   string `json:"rootCertPath" yaml:"rootCertPath"`
			ClientKeyPath  string `json:"clientKeyPath" yaml:"clientKeyPath"`
			ClientCertPath string `json:"clientCertPath" yaml:"clientCertPath"`
		} `yaml:"cert" json:"cert"`
	} `yaml:"pusher" json:"pusher"`
}

// PromEntry Prometheus entry which implements rkentry.Entry.
type PromEntry struct {
	entryName        string                `json:"-" yaml:"-"`
	entryType        string                `json:"-" yaml:"-"`
	entryDescription string                `json:"-" yaml:"-"`
	Pusher           *PushGatewayPusher    `json:"-" yaml:"-"`
	loggerEntry      *LoggerEntry          `json:"-" yaml:"-"`
	Port             uint64                `json:"-" yaml:"-"`
	Path             string                `json:"-" yaml:"-"`
	Registry         *prometheus.Registry  `json:"-" yaml:"-"`
	Registerer       prometheus.Registerer `json:"-" yaml:"-"`
	Gatherer         prometheus.Gatherer   `json:"-" yaml:"-"`
}

// RegisterPromEntry Create a prom entry with options and add prom entry to rkentry.GlobalAppCtx
func RegisterPromEntry(boot *BootProm, port uint64, registry *prometheus.Registry, loggerEntry *LoggerEntry) *PromEntry {
	if !boot.Enabled {
		return nil
	}

	entry := &PromEntry{
		Port:             port,
		Path:             boot.Path,
		loggerEntry:      loggerEntry,
		entryName:        "PromEntry",
		entryType:        "PromEntry",
		entryDescription: "Internal RK entry which implements prometheus client.",
		Registry:         registry,
		Registerer:       prometheus.DefaultRegisterer,
		Gatherer:         prometheus.DefaultGatherer,
	}

	if entry.Registry == nil {
		entry.Registry = prometheus.NewRegistry()
	}

	entry.Registry.Register(collectors.NewGoCollector())

	// Trim space by default
	entry.Path = strings.TrimSpace(entry.Path)

	if len(entry.Path) < 1 {
		// Invalid path, use default one
		entry.Path = defaultPath
	}

	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if entry.loggerEntry == nil {
		entry.loggerEntry = LoggerEntryStdout
	}

	if entry.Registry != nil {
		entry.Registerer = entry.Registry
		entry.Gatherer = entry.Registry
	}

	entry.Pusher = newPushGatewayPusher(boot, entry.loggerEntry, entry.Gatherer)

	return entry
}

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
		"port":              entry.Port,
		"path":              entry.Path,
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Unmarshal entry
func (entry *PromEntry) UnmarshalJSON(b []byte) error {
	return nil
}

// RegisterCollectors Register collectors in default registry
func (entry *PromEntry) RegisterCollectors(collectors ...prometheus.Collector) error {
	var err error
	for i := range collectors {
		if innerErr := entry.Registerer.Register(collectors[i]); innerErr != nil {
			err = innerErr
		}
	}

	return err
}

// PushGatewayPusher is a pusher which contains bellow instances
type PushGatewayPusher struct {
	loggerEntry    *LoggerEntry  `json:"-" yaml:"-"`
	Pusher         *push.Pusher  `json:"-" yaml:"-"`
	IntervalMs     time.Duration `json:"-" yaml:"-"`
	RemoteAddress  string        `json:"-" yaml:"-"`
	JobName        string        `json:"-" yaml:"-"`
	running        *atomic.Bool  `json:"-" yaml:"-"`
	embedFS        *embed.FS     `json:"-" yaml:"-"`
	clientCertPath string        `json:"-" yaml:"-"`
	clientKeyPath  string        `json:"-" yaml:"-"`
	rootCertPath   string        `json:"-" yaml:"-"`
}

// newPushGatewayPusher creates a new pushGateway periodic job instances with intervalMS, remote URL and job name
func newPushGatewayPusher(boot *BootProm, loggerEntry *LoggerEntry, gatherer prometheus.Gatherer) *PushGatewayPusher {
	if !boot.Pusher.Enabled {
		return nil
	}

	pg := &PushGatewayPusher{
		loggerEntry:    loggerEntry,
		IntervalMs:     time.Duration(boot.Pusher.IntervalMs) * time.Millisecond,
		JobName:        boot.Pusher.JobName,
		RemoteAddress:  boot.Pusher.RemoteAddress,
		running:        atomic.NewBool(false),
		clientCertPath: boot.Pusher.Cert.ClientCertPath,
		clientKeyPath:  boot.Pusher.Cert.ClientKeyPath,
		rootCertPath:   boot.Pusher.Cert.RootCertPath,
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

func (pub *PushGatewayPusher) SetEmbedFS(fs *embed.FS) {
	pub.embedFS = fs
}

// Bootstrap starts a periodic job
func (pub *PushGatewayPusher) Bootstrap(ctx context.Context) {
	httpClient := http.DefaultClient

	// deal with tls
	if len(pub.clientCertPath) > 0 && len(pub.clientKeyPath) > 0 {
		clientCert, err := tls.X509KeyPair(
			readFile(pub.clientCertPath, pub.embedFS),
			readFile(pub.clientKeyPath, pub.embedFS))
		if err != nil {
			ShutdownWithError(err)
		}

		rootCert := x509.NewCertPool()
		rootCert.AppendCertsFromPEM(readFile(pub.rootCertPath, pub.embedFS))
		conf := &tls.Config{
			RootCAs: rootCert,
			Certificates: []tls.Certificate{
				clientCert,
			},
		}

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
