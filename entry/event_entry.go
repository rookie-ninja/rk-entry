package rk

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"errors"
	"github.com/rookie-ninja/rk-entry/v3"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

const EventKind = "event"

type EventConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		Api struct {
			Enabled   bool   `yaml:"enabled"`
			UrlPrefix string `yaml:"urlPrefix"`
		} `yaml:"api"`
		Encoding    string             `yaml:"encoding"`
		OutputPaths []string           `yaml:"outputPaths"`
		Rotator     *lumberjack.Logger `yaml:"rotator"`
		Loki        struct {
			Enabled            bool              `yaml:"enabled"`
			Addr               string            `yaml:"addr"`
			Path               string            `yaml:"path"`
			Username           string            `yaml:"username"`
			Password           string            `yaml:"password"`
			InsecureSkipVerify bool              `yaml:"insecureSkipVerify"`
			Labels             map[string]string `yaml:"labels" json:"labels"`
			MaxBatchWait       string            `yaml:"maxBatchWait"`
			MaxBatchSize       int               `yaml:"maxBatchSize"`
		} `yaml:"loki"`
	} `yaml:"entry"`
}

func (e *EventConfig) JSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e *EventConfig) YAML() string {
	b, _ := yaml.Marshal(e)
	return string(b)
}

func (e *EventConfig) Header() *EntryConfigHeader {
	return &e.EntryConfigHeader
}

func (e *EventConfig) Register() (Entry, error) {
	if !e.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(e.Metadata.Domain) {
		return nil, nil
	}

	entry := &EventEntry{
		config: e,
		once:   sync.Once{},
	}

	var eventFactory *rkquery.EventFactory
	var lokiSyncer *rklogger.LokiSyncer

	// Assign default zap config and lumberjack config
	eventLoggerConfig := rklogger.NewZapEventConfig()
	eventLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()

	// Override with user provided zap config and lumberjack config
	rku.OverrideLumberjackConfig(eventLoggerLumberjackConfig, e.Entry.Rotator)

	// If output paths were provided by user, we will override it which means <stdout> would be omitted
	if len(e.Entry.OutputPaths) > 0 {
		eventLoggerConfig.OutputPaths = e.Entry.OutputPaths
	}

	// Loki Syncer
	syncers := make([]zapcore.WriteSyncer, 0)
	if e.Entry.Loki.Enabled {
		wait := time.Duration(0)
		if len(e.Entry.Loki.MaxBatchWait) > 0 {
			v, err := time.ParseDuration(e.Entry.Loki.MaxBatchWait)
			if err != nil {
				return nil, err
			}
			wait = v
		}

		opts := []rklogger.LokiSyncerOption{
			rklogger.WithLokiAddr(e.Entry.Loki.Addr),
			rklogger.WithLokiPath(e.Entry.Loki.Path),
			rklogger.WithLokiUsername(e.Entry.Loki.Username),
			rklogger.WithLokiPassword(e.Entry.Loki.Password),
			rklogger.WithLokiMaxBatchSize(e.Entry.Loki.MaxBatchSize),
			rklogger.WithLokiMaxBatchWaitMs(wait),
		}

		// default labels
		opts = append(opts,
			rklogger.WithLokiLabel(rkm.Domain.Key, rkm.Domain.String),
			rklogger.WithLokiLabel("service_name", Registry.ServiceName()),
			rklogger.WithLokiLabel("service_version", Registry.ServiceVersion()),
			rklogger.WithLokiLabel("logger_type", "event"),
		)

		// labels
		for k, v := range e.Entry.Loki.Labels {
			opts = append(opts, rklogger.WithLokiLabel(k, v))
		}

		if e.Entry.Loki.InsecureSkipVerify {
			opts = append(opts, rklogger.WithLokiClientTls(&tls.Config{
				InsecureSkipVerify: true,
			}))
		}

		lokiSyncer = rklogger.NewLokiSyncer(opts...)
		syncers = append(syncers, lokiSyncer)
	}

	var eventLogger *zap.Logger
	var err error
	if eventLogger, err = rklogger.NewZapLoggerWithConfAndSyncer(eventLoggerConfig, eventLoggerLumberjackConfig, syncers); err != nil {
		return nil, err
	} else {
		eventFactory = rkquery.NewEventFactory(
			rkquery.WithZapLogger(eventLogger),
			rkquery.WithServiceName(Registry.ServiceName()),
			rkquery.WithServiceVersion(Registry.ServiceVersion()),
			rkquery.WithEncoding(rkquery.ToEncoding(e.Entry.Encoding)))
	}

	entry.EventFactory = eventFactory
	entry.EventHelper = rkquery.NewEventHelper(eventFactory)
	entry.lokiSyncer = lokiSyncer
	entry.BaseLogger = eventLogger
	entry.BaseLoggerConfig = eventLoggerConfig

	if entry.config.Entry.Api.Enabled {
		if strings.Trim(entry.config.Entry.Api.UrlPrefix, "/") == "" {
			entry.config.Entry.Api.UrlPrefix = "/rk/v1/event"
		}
	}

	// change swagger config file
	oldSwaggerSpec, err := rkembed.AssetsFS.ReadFile("assets/sw/config/swagger.json")
	if err != nil {
		rku.ShutdownWithError(err)
	}

	m := map[string]interface{}{}

	if err := json.Unmarshal(oldSwaggerSpec, &m); err != nil {
		rku.ShutdownWithError(err)
	}

	if ps, ok := m["paths"]; ok {
		var inner map[string]interface{}
		if inner, ok = ps.(map[string]interface{}); !ok {
			rku.ShutdownWithError(errors.New("invalid format of swagger.json"))
		}

		for p, v := range inner {
			switch p {
			case "/rk/v1/event/activate":
				urlActivate := path.Join(e.Entry.Api.UrlPrefix, "activate")
				if p != urlActivate {
					inner[urlActivate] = v
					delete(inner, p)
				}
			case "/rk/v1/event/deactivate":
				urlDeactivate := path.Join(e.Entry.Api.UrlPrefix, "deactivate")
				if p != urlDeactivate {
					inner[urlDeactivate] = v
					delete(inner, p)
				}
			}
		}
	}

	if newSwaggerSpec, err := json.Marshal(&m); err != nil {
		rku.ShutdownWithError(err)
	} else {
		swaggerSpec = newSwaggerSpec
	}

	Registry.AddEntry(entry)

	return entry, nil
}

type EventEntry struct {
	*rkquery.EventFactory
	*rkquery.EventHelper
	config           *EventConfig
	lokiSyncer       *rklogger.LokiSyncer
	BaseLogger       *zap.Logger
	BaseLoggerConfig *zap.Config
	Rotator          *lumberjack.Logger
	once             sync.Once
}

func (e *EventEntry) Category() string {
	return CategoryIndependent
}

func (e *EventEntry) Kind() string {
	return e.config.Kind
}

func (e *EventEntry) Name() string {
	return e.config.Metadata.Name
}

func (e *EventEntry) Config() EntryConfig {
	return e.config
}

func (e *EventEntry) Bootstrap(ctx context.Context) {
	e.once.Do(func() {
		if e.lokiSyncer != nil {
			e.lokiSyncer.Bootstrap(ctx)
		}
	})
}

func (e *EventEntry) Interrupt(ctx context.Context) {
	if e.lokiSyncer != nil {
		e.lokiSyncer.Interrupt(ctx)
	}
}

func (e *EventEntry) Monitor() *Monitor {
	return nil
}

func (e *EventEntry) FS() *embed.FS {
	return Registry.EntryFS(e.Kind(), e.Name())
}

func (e *EventEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)
	if !e.config.Entry.Api.Enabled {
		return res
	}

	res = append(res,
		&BuiltinApi{
			Method:  http.MethodPost,
			Path:    path.Join(e.config.Entry.Api.UrlPrefix, "/deactivate"),
			Handler: e.deactivateEventHandler(),
		},
		&BuiltinApi{
			Method:  http.MethodPost,
			Path:    path.Join(e.config.Entry.Api.UrlPrefix, "/activate"),
			Handler: e.activateEventHandler(),
		})

	return res
}

// Activate event log level
// @Summary Activate event logger level
// @Id 31000
// @Version 1.0
// @Tags     event
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @Produce json
// @Success 200 {object} eventLevelResp
// @Failure 500 {object} eventErrorResp
// @Router /rk/v1/event/activate [post]
func (e *EventEntry) activateEventHandler() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		e.BaseLoggerConfig.Level.SetLevel(zapcore.InfoLevel)
		b, _ := json.Marshal(&eventLevelResp{
			Level: zapcore.InfoLevel.String(),
		})

		resp.WriteHeader(http.StatusOK)
		resp.Write(b)
	}
}

// Deactivate event log level
// @Summary Deactivate event logger level
// @Id 31001
// @Version 1.0
// @Tags     event
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @Produce json
// @Success 200 {object} eventLevelResp
// @Failure 500 {object} eventErrorResp
// @Router /rk/v1/event/deactivate [post]
func (e *EventEntry) deactivateEventHandler() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		e.BaseLoggerConfig.Level.SetLevel(zapcore.FatalLevel)
		b, _ := json.Marshal(&eventLevelResp{
			Level: zapcore.FatalLevel.String(),
		})

		resp.WriteHeader(http.StatusOK)
		resp.Write(b)
	}
}

type eventLevelResp struct {
	Level string `json:"level"`
}

type eventErrorResp struct {
	Error string `json:"error"`
}

// AddEntryLabelToLokiSyncer add entry name entry type into loki syncer
func (e *EventEntry) AddEntryLabelToLokiSyncer(in Entry) {
	if e.lokiSyncer != nil && e != nil {
		e.lokiSyncer.AddLabel("entry_name", in.Name())
		e.lokiSyncer.AddLabel("entry_kind", in.Kind())
	}
}

// AddLabelToLokiSyncer add key value pair as label into loki syncer
func (e *EventEntry) AddLabelToLokiSyncer(k, v string) {
	if e.lokiSyncer != nil {
		e.lokiSyncer.AddLabel(k, v)
	}
}

// Sync underlying logger
func (e *EventEntry) Sync() {
	if e.BaseLogger != nil {
		e.BaseLogger.Sync()
	}
}

var (
	EventEntryNoop   = NewEventEntryNoop()
	EventEntryStdout = NewEventEntryStdout()
)

// NewEventEntryNoop create event logger entry with noop event factory.
// Event factory and event helper will be created with noop zap logger.
// Since we don't need any log rotation in case of noop, lumberjack config will be nil.
func NewEventEntryNoop() *EventEntry {
	config := &EventConfig{}
	config.Kind = EventKind
	config.ApiVersion = "v1"
	config.Metadata.Name = "noop"
	config.Metadata.Version = "v1"
	config.Metadata.Domain = "*"
	config.Metadata.Enabled = true

	entry := &EventEntry{
		config:       config,
		EventFactory: rkquery.NewEventFactory(rkquery.WithZapLogger(rklogger.NoopLogger)),
		BaseLogger:   rklogger.NoopLogger,
	}

	entry.EventHelper = rkquery.NewEventHelper(entry.EventFactory)

	return entry
}

// NewEventEntryStdout create event logger entry with stdout event factory.
func NewEventEntryStdout() *EventEntry {
	config := &EventConfig{}
	config.Kind = EventKind
	config.ApiVersion = "v1"
	config.Metadata.Name = "stdout"
	config.Metadata.Version = "v1"
	config.Metadata.Domain = "*"
	config.Metadata.Enabled = true

	logger, loggerConfig, _ := rklogger.NewZapLoggerWithBytes(rklogger.EventLoggerConfigBytes, rklogger.JSON)

	entry := &EventEntry{
		config: config,
		EventFactory: rkquery.NewEventFactory(
			rkquery.WithZapLogger(rklogger.EventLogger),
			rkquery.WithServiceName(Registry.serviceName),
			rkquery.WithServiceVersion(Registry.serviceVersion)),
		BaseLogger:       logger,
		BaseLoggerConfig: loggerConfig,
	}

	entry.EventHelper = rkquery.NewEventHelper(entry.EventFactory)

	return entry
}
