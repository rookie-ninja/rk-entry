// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rookie-ninja/rk-entry/v2"
	"github.com/rookie-ninja/rk-entry/v2/os"
	"net/http"
	"net/url"
	"runtime"
)

var swAssetsFile []byte

// @title RK Common Service
// @version 2.0
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

// BootCommonService Bootstrap config of common service.
type BootCommonService struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	PathPrefix string `yaml:"pathPrefix" json:"pathPrefix"`
}

// CommonServiceEntry RK common service which contains commonly used APIs
type CommonServiceEntry struct {
	entryName        string `json:"-" yaml:"-"`
	entryType        string `json:"-" yaml:"-"`
	entryDescription string `json:"-" yaml:"-"`
	pathPrefix       string `json:"-" yaml:"-"`
	ReadyPath        string `json:"-" yaml:"-"`
	AlivePath        string `json:"-" yaml:"-"`
	GcPath           string `json:"-" yaml:"-"`
	InfoPath         string `json:"-" yaml:"-"`
}

// CommonServiceEntryOption option for CommonServiceEntry
type CommonServiceEntryOption func(entry *CommonServiceEntry)

// WithNameCommonServiceEntry provide entry name
func WithNameCommonServiceEntry(name string) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.entryName = name
	}
}

// RegisterCommonServiceEntry Create new common service entry with options.
func RegisterCommonServiceEntry(boot *BootCommonService, opts ...CommonServiceEntryOption) *CommonServiceEntry {
	if boot.Enabled {
		entry := &CommonServiceEntry{
			entryName:        "CommonServiceEntry",
			entryType:        CommonServiceEntryType,
			entryDescription: "Internal RK entry which implements commonly used API.",
			ReadyPath:        "ready",
			AlivePath:        "alive",
			GcPath:           "gc",
			InfoPath:         "info",
			pathPrefix:       boot.PathPrefix,
		}

		for i := range opts {
			opts[i](entry)
		}

		if len(boot.PathPrefix) < 1 {
			entry.pathPrefix = "/rk/v1"
		}

		// append prefix
		entry.ReadyPath, _ = url.JoinPath("/", entry.pathPrefix, entry.ReadyPath)
		entry.AlivePath, _ = url.JoinPath("/", entry.pathPrefix, entry.AlivePath)
		entry.GcPath, _ = url.JoinPath("/", entry.pathPrefix, entry.GcPath)
		entry.InfoPath, _ = url.JoinPath("/", entry.pathPrefix, entry.InfoPath)

		// change swagger config file
		oldSwAssets := readFile("assets/sw/config/swagger.json", &rkembed.AssetsFS, true)
		m := map[string]interface{}{}

		if err := json.Unmarshal(oldSwAssets, &m); err != nil {
			ShutdownWithError(err)
		}

		if ps, ok := m["paths"]; ok {
			var inner map[string]interface{}
			if inner, ok = ps.(map[string]interface{}); !ok {
				ShutdownWithError(errors.New("Invalid format of swagger.json"))
			}

			for p, v := range inner {
				switch p {
				case "/rk/v1/ready":
					if p != entry.ReadyPath {
						inner[entry.ReadyPath] = v
						delete(inner, p)
					}
				case "/rk/v1/alive":
					if p != entry.AlivePath {
						inner[entry.AlivePath] = v
						delete(inner, p)
					}
				case "/rk/v1/gc":
					if p != entry.GcPath {
						inner[entry.GcPath] = v
						delete(inner, p)
					}
				case "/rk/v1/info":
					if p != entry.InfoPath {
						inner[entry.InfoPath] = v
						delete(inner, p)
					}
				}
			}
		}

		if newSwAssets, err := json.Marshal(&m); err != nil {
			ShutdownWithError(err)
		} else {
			swAssetsFile = newSwAssets
		}

		return entry
	}

	return nil
}

// Bootstrap common service entry.
func (entry *CommonServiceEntry) Bootstrap(context.Context) {}

// Interrupt common service entry.
func (entry *CommonServiceEntry) Interrupt(context.Context) {}

// GetName Get name of entry.
func (entry *CommonServiceEntry) GetName() string {
	return entry.entryName
}

// GetType Get entry type.
func (entry *CommonServiceEntry) GetType() string {
	return entry.entryType
}

// GetDescription Get description of entry.
func (entry *CommonServiceEntry) GetDescription() string {
	return entry.entryDescription
}

// String Stringfy entry.
func (entry *CommonServiceEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry.
func (entry *CommonServiceEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.GetName(),
		"type":        entry.GetType(),
		"description": entry.GetDescription(),
		"readyPath":   entry.ReadyPath,
		"alivePath":   entry.AlivePath,
		"gcPath":      entry.GcPath,
		"infoPath":    entry.InfoPath,
	}

	return json.Marshal(m)
}

// UnmarshalJSON Not supported.
func (entry *CommonServiceEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Ready handler
// @Summary Get application readiness status
// @Id 8001
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} readyResp
// @Failure 500 {object} rkerror.ErrorInterface
// @Router /rk/v1/ready [get]
func (entry *CommonServiceEntry) Ready(writer http.ResponseWriter, request *http.Request) {
	if GlobalAppCtx.readinessCheck != nil && !GlobalAppCtx.readinessCheck(request, writer) {
		return
	}

	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.MarshalIndent(&readyResp{
		Ready: true,
	}, "", "  ")
	writer.Write(bytes)
}

// Alive handler
// @Summary Get application liveness status
// @Id 8002
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} aliveResp
// @Router /rk/v1/alive [get]
func (entry *CommonServiceEntry) Alive(writer http.ResponseWriter, request *http.Request) {
	if GlobalAppCtx.livenessCheck != nil && !GlobalAppCtx.livenessCheck(request, writer) {
		return
	}

	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.MarshalIndent(&aliveResp{
		Alive: true,
	}, "", "  ")
	writer.Write(bytes)
}

// Gc handler
// @Summary Trigger Gc
// @Id 8003
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} gcResp
// @Router /rk/v1/gc [get]
func (entry *CommonServiceEntry) Gc(writer http.ResponseWriter, request *http.Request) {
	before := rkos.NewMemInfo()
	runtime.GC()
	after := rkos.NewMemInfo()

	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.MarshalIndent(&gcResp{
		MemStatBeforeGc: before,
		MemStatAfterGc:  after,
	}, "", "  ")
	writer.Write(bytes)
}

// Info handler
// @Summary Get application and process info
// @Id 8004
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @produce application/json
// @Success 200 {object} ProcessInfo
// @Router /rk/v1/info [get]
func (entry *CommonServiceEntry) Info(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	bytes, _ := json.MarshalIndent(NewProcessInfo(), "", "  ")
	writer.Write(bytes)
}
