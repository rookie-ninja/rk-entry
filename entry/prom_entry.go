package rkentry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	// Why 1608? It is the year of first telescope was invented
	defaultPort = uint64(1608)
	defaultPath = "/metrics"
)

const (
	// PromEntryType default entry type
	PromEntryType = "PromEntry"
	// PromEntryNameDefault default entry name
	PromEntryNameDefault = "PromDefault"
	// PromEntryDescription default entry description
	PromEntryDescription = "Internal RK entry which implements prometheus client."
)

// BootConfigProm Boot config which is for prom entry.
type BootConfigProm struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path" json:"path"`
	Pusher  struct {
		Enabled       bool   `yaml:"enabled" json:"enabled"`
		IntervalMs    int64  `yaml:"IntervalMs" json:"IntervalMs"`
		JobName       string `yaml:"jobName" json:"jobName"`
		RemoteAddress string `yaml:"remoteAddress" json:"remoteAddress"`
		BasicAuth     string `yaml:"basicAuth" json:"basicAuth"`
		Cert          struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
	} `yaml:"pusher" json:"pusher"`
}

// PromEntry Prometheus entry which implements rkentry.Entry.
type PromEntry struct {
	Pusher           *PushGatewayPusher    `json:"-" yaml:"-"`
	EntryName        string                `json:"entryName" yaml:"entryName"`
	EntryType        string                `json:"entryType" yaml:"entryType"`
	EntryDescription string                `json:"-" yaml:"-"`
	ZapLoggerEntry   *ZapLoggerEntry       `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry     `json:"-" yaml:"-"`
	Port             uint64                `json:"port" yaml:"port"`
	Path             string                `json:"path" yaml:"path"`
	Registry         *prometheus.Registry  `json:"-" yaml:"-"`
	Registerer       prometheus.Registerer `json:"-" yaml:"-"`
	Gatherer         prometheus.Gatherer   `json:"-" yaml:"-"`
}

// PromEntryOption Prom entry option used while initializing prom entry via code
type PromEntryOption func(*PromEntry)

// WithNameProm Name of prom entry
func WithNameProm(name string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.EntryName = name
	}
}

// WithPortProm Port of prom entry
func WithPortProm(port uint64) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Port = port
	}
}

// WithPathProm Path of prom entry
func WithPathProm(path string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Path = path
	}
}

// WithZapLoggerEntryProm rkentry.ZapLoggerEntry of prom entry
func WithZapLoggerEntryProm(zapLoggerEntry *ZapLoggerEntry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// WithEventLoggerEntryProm rkentry.EventLoggerEntry of prom entry
func WithEventLoggerEntryProm(eventLoggerEntry *EventLoggerEntry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// WithPusherProm PushGateway of prom entry
func WithPusherProm(pusher *PushGatewayPusher) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Pusher = pusher
	}
}

// WithPromRegistryProm Provide a new prometheus registry
func WithPromRegistryProm(registry *prometheus.Registry) PromEntryOption {
	return func(entry *PromEntry) {
		if registry != nil {
			entry.Registry = registry
		}
	}
}

// RegisterPromEntryWithConfig create PromEntry with config
func RegisterPromEntryWithConfig(config *BootConfigProm,
	name string,
	port uint64,
	zap *ZapLoggerEntry,
	event *EventLoggerEntry,
	registry *prometheus.Registry) *PromEntry {
	var promEntry *PromEntry
	if config.Enabled {
		var pusher *PushGatewayPusher
		if config.Pusher.Enabled {
			certEntry := GlobalAppCtx.GetCertEntry(config.Pusher.Cert.Ref)
			var certStore *CertStore

			if certEntry != nil {
				certStore = certEntry.Store
			}

			pusher, _ = NewPushGatewayPusher(
				WithIntervalMSPusher(time.Duration(config.Pusher.IntervalMs)*time.Millisecond),
				WithRemoteAddressPusher(config.Pusher.RemoteAddress),
				WithJobNamePusher(config.Pusher.JobName),
				WithBasicAuthPusher(config.Pusher.BasicAuth),
				WithZapLoggerEntryPusher(zap),
				WithEventLoggerEntryPusher(event),
				WithCertStorePusher(certStore))
		}

		if registry == nil {
			registry = prometheus.NewRegistry()
		}

		registry.Register(prometheus.NewGoCollector())
		promEntry = RegisterPromEntry(
			WithNameProm(name),
			WithPortProm(port),
			WithPathProm(config.Path),
			WithZapLoggerEntryProm(zap),
			WithPromRegistryProm(registry),
			WithEventLoggerEntryProm(event),
			WithPusherProm(pusher))

		if promEntry.Pusher != nil {
			promEntry.Pusher.SetGatherer(promEntry.Gatherer)
		}
	}

	return promEntry
}

// RegisterPromEntry Create a prom entry with options and add prom entry to rkentry.GlobalAppCtx
func RegisterPromEntry(opts ...PromEntryOption) *PromEntry {
	entry := &PromEntry{
		Port:             defaultPort,
		Path:             defaultPath,
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		EntryName:        PromEntryNameDefault,
		EntryType:        PromEntryType,
		EntryDescription: PromEntryDescription,
		Registerer:       prometheus.DefaultRegisterer,
		Gatherer:         prometheus.DefaultGatherer,
	}

	for i := range opts {
		opts[i](entry)
	}

	// Trim space by default
	entry.Path = strings.TrimSpace(entry.Path)

	if len(entry.Path) < 1 {
		// Invalid path, use default one
		entry.Path = defaultPath
	}

	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if entry.Registry != nil {
		entry.Registerer = entry.Registry
		entry.Gatherer = entry.Registry
	}

	return entry
}

// Bootstrap Start prometheus client
func (entry *PromEntry) Bootstrap(ctx context.Context) {
	// start pusher
	if entry.Pusher != nil {
		entry.Pusher.Start()
	}
}

// Interrupt Shutdown prometheus client
func (entry *PromEntry) Interrupt(ctx context.Context) {
	if entry.Pusher != nil {
		entry.Pusher.Stop()
	}
}

// GetName Return name of prom entry
func (entry *PromEntry) GetName() string {
	return entry.EntryName
}

// GetType Return type of prom entry
func (entry *PromEntry) GetType() string {
	return entry.EntryType
}

// GetDescription Get description of entry
func (entry *PromEntry) GetDescription() string {
	return entry.EntryDescription
}

// String Stringfy prom entry
func (entry *PromEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry
func (entry *PromEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":         entry.EntryName,
		"entryType":         entry.EntryType,
		"entryDescription":  entry.EntryDescription,
		"pushGateWayPusher": entry.Pusher,
		"eventLoggerEntry":  entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":    entry.ZapLoggerEntry.GetName(),
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
	ZapLoggerEntry   *ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	EventLoggerEntry *EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	CertStore        *CertStore        `json:"certStore" yaml:"certStore"`
	Pusher           *push.Pusher      `json:"-" yaml:"-"`
	IntervalMs       time.Duration     `json:"intervalMs" yaml:"intervalMs"`
	RemoteAddress    string            `json:"remoteAddress" yaml:"remoteAddress"`
	JobName          string            `json:"jobName" yaml:"jobName"`
	Running          *atomic.Bool      `json:"running" yaml:"running"`
	lock             *sync.Mutex       `json:"-" yaml:"-"`
	Credential       string            `json:"-" yaml:"-"`
}

// PushGatewayPusherOption is used while initializing push gateway pusher via code
type PushGatewayPusherOption func(*PushGatewayPusher)

// WithIntervalMSPusher provides interval in milliseconds
func WithIntervalMSPusher(intervalMs time.Duration) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.IntervalMs = intervalMs
	}
}

// WithRemoteAddressPusher provides remote address of pushgateway
func WithRemoteAddressPusher(remoteAddress string) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.RemoteAddress = remoteAddress
	}
}

// WithJobNamePusher provides job name
func WithJobNamePusher(jobName string) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.JobName = jobName
	}
}

// WithBasicAuthPusher provides basic auth of pushgateway
func WithBasicAuthPusher(cred string) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.Credential = cred
	}
}

// WithZapLoggerEntryPusher provides ZapLoggerEntry
func WithZapLoggerEntryPusher(zapLoggerEntry *ZapLoggerEntry) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.ZapLoggerEntry = zapLoggerEntry
	}
}

// WithEventLoggerEntryPusher provides EventLoggerEntry
func WithEventLoggerEntryPusher(eventLoggerEntry *EventLoggerEntry) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.EventLoggerEntry = eventLoggerEntry
	}
}

// WithCertStorePusher provides EventLoggerEntry
func WithCertStorePusher(certStore *CertStore) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.CertStore = certStore
	}
}

// NewPushGatewayPusher creates a new pushGateway periodic job instances with intervalMS, remote URL and job name
// 1: intervalMS: should be a positive integer
// 2: url:        should be a non empty and valid url
// 3: jabName:    should be a non empty string
// 4: cred:       credential of basic auth format as user:pass
// 5: logger:     a logger with stdout output would be assigned if nil
func NewPushGatewayPusher(opts ...PushGatewayPusherOption) (*PushGatewayPusher, error) {
	pg := &PushGatewayPusher{
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		IntervalMs:       1 * time.Second,
		lock:             &sync.Mutex{},
		Running:          atomic.NewBool(false),
	}

	for i := range opts {
		opts[i](pg)
	}

	if pg.IntervalMs < 1 {
		return nil, errors.New("invalid intervalMs")
	}

	if len(pg.RemoteAddress) < 1 {
		return nil, errors.New("empty remoteAddress")
	}

	// certificate was provided, we need to use https for remote address
	if pg.CertStore != nil {
		if !strings.HasPrefix(pg.RemoteAddress, "https://") {
			pg.RemoteAddress = "https://" + pg.RemoteAddress
		}
	}

	if len(pg.JobName) < 1 {
		return nil, errors.New("empty job name")
	}

	if pg.ZapLoggerEntry == nil {
		pg.ZapLoggerEntry = GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if pg.EventLoggerEntry == nil {
		pg.EventLoggerEntry = GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	pg.Pusher = push.New(pg.RemoteAddress, pg.JobName)

	// assign credential of basic auth
	if len(pg.Credential) > 0 && strings.Contains(pg.Credential, ":") {
		pg.Credential = strings.TrimSpace(pg.Credential)
		tokens := strings.Split(pg.Credential, ":")
		if len(tokens) == 2 {
			pg.Pusher = pg.Pusher.BasicAuth(tokens[0], tokens[1])
		}
	}

	httpClient := &http.Client{
		Timeout: DefaultTimeout,
	}

	// deal with tls
	if pg.CertStore != nil {
		certPool := x509.NewCertPool()

		certPool.AppendCertsFromPEM(pg.CertStore.ServerCert)

		conf := &tls.Config{RootCAs: certPool}

		cert, err := tls.X509KeyPair(pg.CertStore.ClientCert, pg.CertStore.ClientKey)

		if err == nil {
			conf.Certificates = []tls.Certificate{cert}
		}

		httpClient.Transport = &http.Transport{TLSClientConfig: conf}
	}

	pg.Pusher.Client(httpClient)

	return pg, nil
}

// Start starts a periodic job
func (pub *PushGatewayPusher) Start() {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	// periodic job already started
	// caution, do not call pub.isRunning() function directory, since it will cause dead lock
	if pub.Running.Load() {
		pub.ZapLoggerEntry.GetLogger().Info("pushGateway publisher already started",
			zap.String("remoteAddress", pub.RemoteAddress),
			zap.String("jobName", pub.JobName))
		return
	}

	pub.Running.CAS(false, true)

	pub.ZapLoggerEntry.GetLogger().Info("starting pushGateway publisher",
		zap.String("remoteAddress", pub.RemoteAddress),
		zap.String("jobName", pub.JobName))

	go pub.push()
}

// Internal use only
func (pub *PushGatewayPusher) push() {
	for pub.Running.Load() {
		event := pub.EventLoggerEntry.GetEventHelper().Start("publish")
		event.AddPayloads(
			zap.String("jobName", pub.JobName),
			zap.String("remoteAddr", pub.RemoteAddress),
			zap.Duration("intervalMs", pub.IntervalMs))

		err := pub.Pusher.Push()

		if err != nil {
			pub.ZapLoggerEntry.GetLogger().Warn("failed to push metrics to PushGateway",
				zap.String("remoteAddress", pub.RemoteAddress),
				zap.String("jobName", pub.JobName),
				zap.Error(err))
			pub.EventLoggerEntry.GetEventHelper().FinishWithError(event, err)
		} else {
			pub.EventLoggerEntry.GetEventHelper().Finish(event)
		}

		time.Sleep(pub.IntervalMs)
	}
}

// IsRunning validate whether periodic job is running or not
func (pub *PushGatewayPusher) IsRunning() bool {
	return pub.Running.Load()
}

// Stop stops periodic job
func (pub *PushGatewayPusher) Stop() {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.Running.CAS(true, false)
}

// GetPusher simply call pusher.Gatherer()
// We add prefix "Add" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) GetPusher() *push.Pusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	return pub.Pusher
}

// String returns string value of PushGatewayPusher
func (pub *PushGatewayPusher) String() string {
	bytes, err := json.Marshal(pub)
	if err != nil {
		// failed to marshal, just return empty string
		return "{}"
	}

	return string(bytes)
}

// SetGatherer sets gatherer of prometheus
func (pub *PushGatewayPusher) SetGatherer(gatherer prometheus.Gatherer) {
	if pub.Pusher != nil {
		pub.Pusher.Gatherer(gatherer)
	}
}
