package rk

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rookie-ninja/rk-entry/v3"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path"
)

const ViperKind = "viper"

type ViperConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		Api struct {
			UrlPrefix string `yaml:"urlPrefix"`
		} `yaml:"api"`
		LocalPath string                 `yaml:"localPath"`
		EmbedPath string                 `yaml:"embedPath"`
		EnvPrefix string                 `yaml:"envPrefix"`
		Content   map[string]interface{} `yaml:"content"`
	} `yaml:"entry"`
}

func (v *ViperConfig) JSON() string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (v *ViperConfig) YAML() string {
	b, _ := yaml.Marshal(v)
	return string(b)
}

func (v *ViperConfig) Header() *EntryConfigHeader {
	return &v.EntryConfigHeader
}

func (v *ViperConfig) Register() (Entry, error) {
	if !v.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(v.Metadata.Domain) {
		return nil, nil
	}

	entry := &ViperEntry{
		config: v,
		Viper:  viper.New(),
	}
	entry.embedFS = Registry.EntryFS(entry.Kind(), entry.Name())

	// for embed
	if len(v.Entry.EmbedPath) > 0 && entry.embedFS != nil {
		// read it
		if b, err := entry.embedFS.ReadFile(v.Entry.EmbedPath); err != nil {
			rku.ShutdownWithError(err)
		} else {
			entry.Viper.SetConfigType(path.Ext(v.Entry.EmbedPath))
			if err := entry.Viper.MergeConfig(bytes.NewReader(b)); err != nil {
				rku.ShutdownWithError(err)
			}
		}
	}

	// for local
	if len(v.Entry.LocalPath) > 0 {
		if !path.IsAbs(v.Entry.LocalPath) {
			if wd, err := os.Getwd(); err != nil {
				rku.ShutdownWithError(err)
			} else {
				v.Entry.LocalPath = path.Join(wd, v.Entry.LocalPath)
			}
		}

		entry.Viper.SetConfigType(path.Ext(v.Entry.LocalPath))
		entry.Viper.SetConfigFile(v.Entry.LocalPath)
		if err := entry.Viper.ReadInConfig(); err != nil {
			rku.ShutdownWithError(fmt.Errorf("failed to read file, path:%s", v.Entry.LocalPath))
		}
	}

	// for content
	for k, val := range v.Entry.Content {
		entry.Viper.Set(k, val)
	}

	// enable automatic env
	// issue: https://github.com/rookie-ninja/rk-boot/issues/55
	entry.Viper.AutomaticEnv()
	entry.Viper.SetEnvPrefix(v.Entry.EnvPrefix)

	if len(v.Entry.Api.UrlPrefix) < 1 {
		v.Entry.Api.UrlPrefix = path.Join("/rk/v1/viper", "/")
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

		for p, val := range inner {
			switch p {
			case "/rk/v1/viper/dump":
				urlDump := path.Join(v.Entry.Api.UrlPrefix, "dump")
				if p != urlDump {
					inner[urlDump] = val
					delete(inner, p)
				}
			case "/rk/v1/viper/set":
				urlSet := path.Join(v.Entry.Api.UrlPrefix, "set")
				if p != urlSet {
					inner[urlSet] = val
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

type ViperEntry struct {
	*viper.Viper

	config  *ViperConfig
	embedFS *embed.FS
}

func (v *ViperEntry) Category() string {
	return CategoryIndependent
}

func (v *ViperEntry) Kind() string {
	return v.config.Kind
}

func (v *ViperEntry) Name() string {
	return v.config.Metadata.Name
}

func (v *ViperEntry) Config() EntryConfig {
	return v.config
}

func (v *ViperEntry) Bootstrap(ctx context.Context) {}

func (v *ViperEntry) Interrupt(ctx context.Context) {}

func (v *ViperEntry) Monitor() *Monitor {
	return nil
}

func (v *ViperEntry) FS() *embed.FS {
	return v.embedFS
}

func (v *ViperEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)

	res = append(res,
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", v.config.Entry.Api.UrlPrefix, "dump"),
			Handler: v.dumpConfigHandler(),
		},
		&BuiltinApi{
			Method:  http.MethodPost,
			Path:    path.Join("/", v.config.Entry.Api.UrlPrefix, "set"),
			Handler: v.setConfigHandler(),
		})

	return res
}

// Dump viper config values
// @Summary Dump viper config values
// @Id 32000
// @Version 1.0
// @Tags     viper
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @Produce json
// @Success 200 {object} viperResp
// @Failure 500 {object} viperError
// @Router /rk/v1/viper/dump [get]
func (v *ViperEntry) dumpConfigHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		bytes, _ := json.Marshal(v.AllSettings())
		writer.Write(bytes)
	}
}

// Set viper config values
// @Summary Set viper config values
// @Id 32001
// @Version 1.0
// @Tags     viper
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security JWT
// @Produce json
// @Param     viperReq  body  viperReq  true  "viperReq"
// @Success 200 {object} viperResp
// @Failure 400 {object} viperError
// @Failure 500 {object} viperError
// @Router /rk/v1/viper/set [post]
func (v *ViperEntry) setConfigHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		b, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("empty body"))
			return
		}

		req := &viperReq{}
		if err := json.Unmarshal(b, req); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			b1, _ := json.Marshal(&viperError{
				Error: "bad request",
			})
			writer.Write(b1)
			return
		}

		for i := range req.Values {
			kv := req.Values[i]
			v.Set(kv.Key, kv.Value)
		}

		writer.WriteHeader(http.StatusOK)
		bytes, _ := json.Marshal(v.AllSettings())
		writer.Write(bytes)
	}
}

type viperReq struct {
	Values []struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	} `json:"values"`
}

type viperResp map[string]interface{}

type viperError struct {
	Error string `json:"error"`
}
