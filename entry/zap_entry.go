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

const ZapKind = "zap"

type ZapConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		Api struct {
			Enabled   bool   `yaml:"enabled"`
			UrlPrefix string `yaml:"urlPrefix"`
		} `yaml:"api"`
		Zap     *rklogger.ZapConfigWrap `yaml:"zap"`
		Rotator *lumberjack.Logger      `yaml:"rotator"`
		Loki    struct {
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

func (c *ZapConfig) YAML() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (c *ZapConfig) JSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *ZapConfig) Header() *EntryConfigHeader {
	return &c.EntryConfigHeader
}

func (c *ZapConfig) Register() (Entry, error) {
	if !c.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(c.Metadata.Domain) {
		return nil, nil
	}

	entry := &ZapEntry{
		config: c,
		once:   sync.Once{},
	}

	// Assign default zap config and lumberjack config
	zapLoggerConfig := rklogger.NewZapStdoutConfig()
	zapLoggerLumberjackConfig := rklogger.NewLumberjackConfigDefault()

	// Override with user provided zap config and lumberjack config
	rku.OverrideZapConfig(zapLoggerConfig, rklogger.TransformToZapConfig(c.Entry.Zap))
	rku.OverrideLumberjackConfig(zapLoggerLumberjackConfig, c.Entry.Rotator)

	// Loki Syncer
	syncers := make([]zapcore.WriteSyncer, 0)
	var lokiSyncer *rklogger.LokiSyncer
	if c.Entry.Loki.Enabled {
		wait := time.Duration(0)
		if len(c.Entry.Loki.MaxBatchWait) > 0 {
			v, err := time.ParseDuration(c.Entry.Loki.MaxBatchWait)
			if err != nil {
				return nil, err
			}
			wait = v
		}

		opts := []rklogger.LokiSyncerOption{
			rklogger.WithLokiAddr(c.Entry.Loki.Addr),
			rklogger.WithLokiPath(c.Entry.Loki.Path),
			rklogger.WithLokiUsername(c.Entry.Loki.Username),
			rklogger.WithLokiPassword(c.Entry.Loki.Password),
			rklogger.WithLokiMaxBatchSize(c.Entry.Loki.MaxBatchSize),
			rklogger.WithLokiMaxBatchWaitMs(wait),
		}

		// labels
		for k, v := range c.Entry.Loki.Labels {
			opts = append(opts, rklogger.WithLokiLabel(k, v))
		}

		// default labels
		opts = append(opts,
			rklogger.WithLokiLabel(rkmid.Domain.Key, rkmid.Domain.String),
			rklogger.WithLokiLabel("service_name", Registry.ServiceName()),
			rklogger.WithLokiLabel("service_version", Registry.ServiceVersion()),
			rklogger.WithLokiLabel("logger_type", "zap"),
		)

		if c.Entry.Loki.InsecureSkipVerify {
			opts = append(opts, rklogger.WithLokiClientTls(&tls.Config{
				InsecureSkipVerify: true,
			}))
		}

		lokiSyncer = rklogger.NewLokiSyncer(opts...)
		syncers = append(syncers, lokiSyncer)
	}

	// Create app logger with config
	zapLogger, err := rklogger.NewZapLoggerWithConfAndSyncer(
		zapLoggerConfig,
		zapLoggerLumberjackConfig,
		syncers,
		zap.AddCaller())

	if err != nil {
		return nil, err
	}

	entry.Logger = zapLogger
	entry.lokiSyncer = lokiSyncer
	entry.LoggerConfig = zapLoggerConfig
	entry.Rotator = zapLoggerLumberjackConfig
	if entry.config.Entry.Api.Enabled {
		if strings.Trim(entry.config.Entry.Api.UrlPrefix, "/") == "" {
			entry.config.Entry.Api.UrlPrefix = "/rk/v1/zap"
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
			case "/rk/v1/zap/level":
				urlLevel := path.Join(c.Entry.Api.UrlPrefix, "level")
				if p != urlLevel {
					inner[urlLevel] = v
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

type ZapEntry struct {
	*zap.Logger
	LoggerConfig *zap.Config
	Rotator      *lumberjack.Logger
	config       *ZapConfig
	lokiSyncer   *rklogger.LokiSyncer
	once         sync.Once
}

func (z *ZapEntry) Category() string {
	return CategoryIndependent
}

func (z *ZapEntry) Kind() string {
	return z.config.Kind
}

func (z *ZapEntry) Name() string {
	return z.config.Metadata.Name
}

func (z *ZapEntry) Config() EntryConfig {
	return z.config
}

func (z *ZapEntry) Bootstrap(ctx context.Context) {
	z.once.Do(func() {
		if z.lokiSyncer != nil {
			z.lokiSyncer.Bootstrap(ctx)
		}
	})
}

func (z *ZapEntry) Interrupt(ctx context.Context) {
	if z.lokiSyncer != nil {
		z.lokiSyncer.Interrupt(ctx)
	}
}

func (z *ZapEntry) FS() *embed.FS {
	return Registry.EntryFS(z.Kind(), z.Name())
}

func (z *ZapEntry) Monitor() *Monitor {
	return nil
}

func (z *ZapEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)
	if !z.config.Entry.Api.Enabled {
		return res
	}

	res = append(res,
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join(z.config.Entry.Api.UrlPrefix, "/level"),
			Handler: z.getLevelHandler(),
		},
		&BuiltinApi{
			Method:  http.MethodPut,
			Path:    path.Join(z.config.Entry.Api.UrlPrefix, "/level"),
			Handler: z.getLevelHandler(),
		},
	)

	return res
}

// Get zap log level
// @Summary Get zap logger level
// @Id 30000
// @Version 1.0
// @Tags     zap
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @Produce json
// @Success 200 {object} zapLevelResp
// @Failure 500 {object} zapErrorResp
// @Router /rk/v1/zap/level [get]
func (z *ZapEntry) getLevelHandler() http.HandlerFunc {
	return z.LoggerConfig.Level.ServeHTTP
}

// Set zap log level
// @Summary Set zap logger level
// @Id 30001
// @Version 1.0
// @Tags     zap
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @Accept x-www-form-urlencoded
// @Param   enumstring  query     string     false  "level"       Enums(debug, info, warn, error, dPanic, panic, fatal)
// @Produce json
// @Success 200 {object} zapLevelResp
// @Failure 500 {object} zapErrorResp
// @Router /rk/v1/zap/level [put]
func (z *ZapEntry) setLevelHandler() http.HandlerFunc {
	return z.LoggerConfig.Level.ServeHTTP
}

type zapLevelResp struct {
	Level string `json:"level"`
}

type zapErrorResp struct {
	Error string `json:"error"`
}

// AddEntryLabelToLokiSyncer add entry name entry type into loki syncer
func (z *ZapEntry) AddEntryLabelToLokiSyncer(e Entry) {
	if z.lokiSyncer != nil && e != nil {
		z.lokiSyncer.AddLabel("entry_name", e.Name())
		z.lokiSyncer.AddLabel("entry_kind", e.Kind())
	}
}

// AddLabelToLokiSyncer add key value pair as label into loki syncer
func (z *ZapEntry) AddLabelToLokiSyncer(k, v string) {
	if z.lokiSyncer != nil {
		z.lokiSyncer.AddLabel(k, v)
	}
}

// Sync underlying logger
func (z *ZapEntry) Sync() {
	if z.Logger != nil {
		z.Logger.Sync()
	}
}

var (
	ZapEntryNoop   = NewZapEntryNoop()
	ZapEntryStdout = NewZapEntryStdout()
)

// NewZapEntryNoop create zap logger entry with noop.
func NewZapEntryNoop() *ZapEntry {
	config := &ZapConfig{}
	config.Kind = "zap"
	config.ApiVersion = "v1"
	config.Metadata.Name = "noop"
	config.Metadata.Version = "v1"
	config.Metadata.Domain = "*"
	config.Metadata.Enabled = true

	return &ZapEntry{
		config: config,
		Logger: rklogger.NoopLogger,
	}
}

// NewZapEntryStdout create zap logger entry with STDOUT.
func NewZapEntryStdout() *ZapEntry {
	config := &ZapConfig{}
	config.Kind = "zap"
	config.ApiVersion = "v1"
	config.Metadata.Name = "stdout"
	config.Metadata.Version = "v1"
	config.Metadata.Domain = "*"
	config.Metadata.Enabled = true

	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	stdoutLoggerConfig := &zap.Config{
		Level:             level,
		Development:       true,
		Encoding:          "console",
		DisableStacktrace: true,
		EncoderConfig:     *rklogger.NewZapStdoutEncoderConfig(),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
	// StdoutLogger is default zap logger whose output path is stdout.
	stdoutLogger, _ := stdoutLoggerConfig.Build()

	return &ZapEntry{
		config:       config,
		Logger:       stdoutLogger,
		LoggerConfig: stdoutLoggerConfig,
	}
}
