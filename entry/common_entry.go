package rk

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"github.com/rookie-ninja/rk-entry/v3"
	"github.com/rookie-ninja/rk-entry/v3/os"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
	"time"
)

const CommonServiceKind = "commonService"

var swaggerSpec []byte

type CommonServiceApiFunc func(resp http.ResponseWriter, req *http.Request)

type CommonServiceConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	UrlPrefix string `yaml:"urlPrefix" json:"urlPrefix"`
}

func (c *CommonServiceConfig) JSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *CommonServiceConfig) YAML() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (c *CommonServiceConfig) Header() *EntryConfigHeader {
	return nil
}

func (c *CommonServiceConfig) Register() (Entry, error) {
	if !c.Enabled {
		return nil, nil
	}

	entry := &CommonServiceEntry{
		config: c,
		gcFunc: func(resp http.ResponseWriter, req *http.Request) {
			before := rkos.NewMemInfo()
			runtime.GC()
			after := rkos.NewMemInfo()

			resp.WriteHeader(http.StatusOK)
			bytes, _ := json.Marshal(&gcResp{
				MemStatBeforeGc: before,
				MemStatAfterGc:  after,
			})
			resp.Write(bytes)
		},
		infoFunc: func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			bytes, _ := json.Marshal(NewProcessInfo())
			resp.Write(bytes)
		},
		readyFunc: func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			bytes, _ := json.Marshal(&readyResp{
				Ready: true,
			})
			resp.Write(bytes)
		},
		aliveFunc: func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			bytes, _ := json.Marshal(&aliveResp{
				Alive: true,
			})
			resp.Write(bytes)
		},
	}

	if len(c.UrlPrefix) < 1 {
		c.UrlPrefix = "/rk/v1"
	} else {
		c.UrlPrefix = path.Join("/", c.UrlPrefix)
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
			case "/rk/v1/ready":
				urlReady := path.Join(c.UrlPrefix, "ready")
				if p != urlReady {
					inner[urlReady] = v
					delete(inner, p)
				}
			case "/rk/v1/alive":
				urlAlive := path.Join(c.UrlPrefix, "alive")
				if p != urlAlive {
					inner[urlAlive] = v
					delete(inner, p)
				}
			case "/rk/v1/gc":
				urlGc := path.Join(c.UrlPrefix, "gc")
				if p != urlGc {
					inner[urlGc] = v
					delete(inner, p)
				}
			case "/rk/v1/info":
				urlInfo := path.Join(c.UrlPrefix, "info")
				if p != urlInfo {
					inner[urlInfo] = v
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

	return entry, nil
}

type CommonServiceEntry struct {
	config    *CommonServiceConfig
	gcFunc    CommonServiceApiFunc
	infoFunc  CommonServiceApiFunc
	readyFunc CommonServiceApiFunc
	aliveFunc CommonServiceApiFunc
}

func (c *CommonServiceEntry) Category() string {
	return CategoryInline
}

func (c *CommonServiceEntry) Kind() string {
	return "commonService"
}

func (c *CommonServiceEntry) Name() string {
	return "default"
}

func (c *CommonServiceEntry) Config() EntryConfig {
	return c.config
}

func (c *CommonServiceEntry) Bootstrap(ctx context.Context) {}

func (c *CommonServiceEntry) Interrupt(ctx context.Context) {}

func (c *CommonServiceEntry) Monitor() *Monitor {
	return nil
}

func (c *CommonServiceEntry) FS() *embed.FS {
	return Registry.EntryFS(c.Kind(), c.Name())
}

func (c *CommonServiceEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)
	if !c.config.Enabled {
		return res
	}

	res = append(res,
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join(c.config.UrlPrefix, "/alive"),
			Handler: http.HandlerFunc(c.aliveFunc),
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join(c.config.UrlPrefix, "/ready"),
			Handler: http.HandlerFunc(c.readyFunc),
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join(c.config.UrlPrefix, "/info"),
			Handler: http.HandlerFunc(c.infoFunc),
		},
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join(c.config.UrlPrefix, "/gc"),
			Handler: http.HandlerFunc(c.gcFunc),
		},
	)

	return res
}

func (c *CommonServiceEntry) OverrideApi(api string, f CommonServiceApiFunc) {
	switch strings.ToLower(api) {
	case "gc":
		c.gcFunc = f
	case "info":
		c.infoFunc = f
	case "alive":
		c.aliveFunc = f
	case "ready":
		c.readyFunc = f
	}
}

// @title RK Common Service
// @version 1.0
// @description.markdown This is builtin common service.
// @contact.name rk-dev
// @contact.url https://github.com/rookie-ninja/rk-entry
// @contact.email lark@pointgoal.io
// @license.name Apache 2.0 License
// @license.url https://github.com/rookie-ninja/rk-entry/blob/master/LICENSE.txt
// @securityDefinitions.basic BasicAuth
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
// @schemes http https

// Ready handler
// @Summary Get application readiness status
// @Id 30001
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} readyResp
// @Failure 500 {object} rkerror.ErrorInterface
// @Router /rk/v1/ready [get]
func (c *CommonServiceEntry) Ready(writer http.ResponseWriter, request *http.Request) {
	c.readyFunc(writer, request)
}

// Alive handler
// @Summary Get application liveness status
// @Id 30002
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} aliveResp
// @Router /rk/v1/alive [get]
func (c *CommonServiceEntry) Alive(writer http.ResponseWriter, request *http.Request) {
	c.aliveFunc(writer, request)
}

// Gc handler
// @Summary Trigger Gc
// @Id 30003
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} gcResp
// @Router /rk/v1/gc [get]
func (c *CommonServiceEntry) Gc(writer http.ResponseWriter, request *http.Request) {
	c.gcFunc(writer, request)
}

// Info handler
// @Summary Get application and process info
// @Id 30004
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} ProcessInfo
// @Router /rk/v1/info [get]
func (c *CommonServiceEntry) Info(writer http.ResponseWriter, request *http.Request) {
	c.infoFunc(writer, request)
}

// aliveResp response of /alive
type aliveResp struct {
	Alive bool `json:"alive" yaml:"alive" example:"true"`
}

// readyResp response of /ready
type readyResp struct {
	Ready bool `json:"ready" yaml:"ready" example:"true"`
}

// gcResp response of /gc
// Returns memory stats of GC before and after.
type gcResp struct {
	MemStatBeforeGc *rkos.MemInfo `json:"memStatBeforeGc" yaml:"memStatBeforeGc"`
	MemStatAfterGc  *rkos.MemInfo `json:"memStatAfterGc" yaml:"memStatAfterGc"`
}

// ProcessInfo process information for a running application.
type ProcessInfo struct {
	ServiceName    string          `json:"serviceName" yaml:"serviceName" example:"rk-app"`
	ServiceVersion string          `json:"serviceVersion" yaml:"serviceVersion" example:"dev"`
	UID            string          `json:"uid" yaml:"uid" example:"501"`
	GID            string          `json:"gid" yaml:"gid" example:"20"`
	Username       string          `json:"username" yaml:"username" example:"lark"`
	StartTime      string          `json:"startTime" yaml:"startTime" example:"2022-03-15T20:43:05+08:00"`
	UpTime         string          `json:"upTime" yaml:"upTime" example:"1h"`
	Domain         string          `json:"domain" yaml:"domain" example:"dev"`
	CpuInfo        *rkos.CpuInfo   `json:"cpuInfo" yaml:"cpuInfo"`
	MemInfo        *rkos.MemInfo   `json:"memInfo" yaml:"memInfo"`
	NetInfo        *rkos.NetInfo   `json:"netInfo" yaml:"netInfo"`
	OsInfo         *rkos.OsInfo    `json:"osInfo" yaml:"osInfo"`
	GoEnvInfo      *rkos.GoEnvInfo `json:"goEnvInfo" yaml:"goEnvInfo"`
}

// NewProcessInfo creates a new ProcessInfo instance
func NewProcessInfo() *ProcessInfo {
	u, err := user.Current()
	// Assign unknown value to user in order to prevent panic
	if err != nil {
		u = &user.User{
			Name: "",
			Uid:  "",
			Gid:  "",
		}
	}

	return &ProcessInfo{
		ServiceName:    Registry.serviceName,
		ServiceVersion: Registry.serviceVersion,
		Username:       u.Name,
		UID:            u.Uid,
		GID:            u.Gid,
		StartTime:      Registry.startTime.Format(time.RFC3339),
		UpTime:         time.Since(Registry.startTime).String(),
		Domain:         rku.GetDefaultIfEmptyString(os.Getenv("DOMAIN"), ""),
		CpuInfo:        rkos.NewCpuInfo(),
		MemInfo:        rkos.NewMemInfo(),
		NetInfo:        rkos.NewNetInfo(),
		OsInfo:         rkos.NewOsInfo(),
		GoEnvInfo:      rkos.NewGoEnvInfo(),
	}
}
