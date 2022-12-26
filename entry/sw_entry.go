// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry/v2"
	rkmid "github.com/rookie-ninja/rk-entry/v2/middleware"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	swaggerJsonFiles     = make(map[string]string, 0)
	swConfigFileContents = ``
)

// Inner struct used while initializing swagger entry.
type swUrlConfig struct {
	Urls []*swUrl `json:"urls" yaml:"urls"`
}

// Inner struct used while initializing swagger entry.
type swUrl struct {
	Name string `json:"name" yaml:"name"`
	Url  string `json:"url" yaml:"url"`
}

// BootSW Bootstrap config of swagger.
// 1: Enabled: Enable swagger.
// 2: Path: Swagger path accessible from restful API.
// 3: JsonPath: The path of where swagger JSON file was located.
// 4: Headers: The headers that would added into each API response.
type BootSW struct {
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	Path     string   `yaml:"path" json:"path"`
	JsonPath string   `yaml:"jsonPath" json:"jsonPath"`
	Headers  []string `yaml:"headers" json:"headers"`
}

// SWEntry implements rke.Entry interface.
type SWEntry struct {
	entryName        string            `json:"-" yaml:"-"`
	entryType        string            `json:"-" yaml:"-"`
	entryDescription string            `json:"-" yaml:"-"`
	JsonPath         string            `json:"-" yaml:"-"`
	Path             string            `json:"-" yaml:"-"`
	Headers          map[string]string `json:"-" yaml:"-"`
	embedFS          *embed.FS         `json:"-" yaml:"-"`
}

type SWEntryOption func(entry *SWEntry)

func WithNameSWEntry(name string) SWEntryOption {
	return func(entry *SWEntry) {
		entry.entryName = name
	}
}

func RegisterSWEntry(boot *BootSW, opts ...SWEntryOption) *SWEntry {
	var swEntry *SWEntry
	if boot.Enabled {
		// Init swagger custom headers from config
		headers := make(map[string]string, 0)
		for i := range boot.Headers {
			header := boot.Headers[i]
			tokens := strings.Split(header, ":")
			if len(tokens) == 2 {
				headers[tokens[0]] = tokens[1]
			}
		}

		swEntry = &SWEntry{
			entryName:        "SwEntry",
			entryType:        SWEntryType,
			entryDescription: "Internal RK entry for swagger UI.",
			Path:             boot.Path,
			JsonPath:         boot.JsonPath,
			Headers:          headers,
		}

		for i := range opts {
			opts[i](swEntry)
		}

		swEntry.embedFS = GlobalAppCtx.GetEmbedFS(swEntry.GetType(), swEntry.GetName())

		if len(swEntry.Path) < 1 {
			swEntry.Path = "/sw"
		}

		// Deal with Path
		// add "/" at start and end side if missing
		if !strings.HasPrefix(swEntry.Path, "/") {
			swEntry.Path = "/" + swEntry.Path
		}

		if !strings.HasSuffix(swEntry.Path, "/") {
			swEntry.Path = swEntry.Path + "/"
		}
	}

	return swEntry
}

func (entry *SWEntry) Bootstrap(ctx context.Context) {
	// init swagger configs
	entry.initSwaggerConfig()
}

func (entry *SWEntry) Interrupt(ctx context.Context) {}

func (entry *SWEntry) GetName() string {
	return entry.entryName
}

func (entry *SWEntry) GetType() string {
	return entry.entryType
}

func (entry *SWEntry) GetDescription() string {
	return entry.entryDescription
}

func (entry *SWEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry
func (entry *SWEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.GetName(),
		"type":        entry.GetType(),
		"description": entry.GetDescription(),
		"jsonPath":    entry.JsonPath,
		"path":        entry.Path,
		"Headers":     entry.Headers,
	}

	return json.Marshal(m)
}

// UnmarshalJSON Unmarshal entry
func (entry *SWEntry) UnmarshalJSON([]byte) error {
	return nil
}

// ConfigFileHandler handler for swagger config files.
func (entry *SWEntry) ConfigFileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		p := strings.TrimSuffix(request.URL.Path, "/")

		writer.Header().Set("cache-control", "no-cache")

		for k, v := range entry.Headers {
			writer.Header().Set(k, v)
		}

		switch p {
		// request index.html file
		case strings.TrimSuffix(entry.Path, "/"):
			if file := readFile("assets/sw/index.html", &rkembed.AssetsFS, false); len(file) < 1 {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			} else {
				http.ServeContent(writer, request, "index.html", time.Now(), bytes.NewReader(file))
			}
		// css files
		case path.Join(entry.Path, "swagger-ui.css"):
			if file := readFile("assets/sw/css/swagger-ui.css", &rkembed.AssetsFS, false); len(file) < 1 {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			} else {
				writer.Header().Set("Content-Type", "text/css; charset=utf-8")
				http.ServeContent(writer, request, "swagger-ui.css", time.Now(), bytes.NewReader(file))
			}
		// favicon files
		case path.Join(entry.Path, "favicon-32x32.png"),
			path.Join(entry.Path, "favicon-16x16.png"):
			base := path.Base(p)
			if file := readFile(filepath.Join("assets/sw/favicon", base), &rkembed.AssetsFS, false); len(file) < 1 {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			} else {
				writer.Header().Set("Content-Type", "image/png")
				http.ServeContent(writer, request, base, time.Now(), bytes.NewReader(file))
			}
		// js files
		case path.Join(entry.Path, "swagger-ui-bundle.js"),
			path.Join(entry.Path, "swagger-ui-standalone-preset.js"):
			base := path.Base(p)
			if file := readFile(filepath.Join("assets/sw/js", base), &rkembed.AssetsFS, false); len(file) < 1 {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			} else {
				writer.Header().Set("Content-Type", "application/javascript")
				http.ServeContent(writer, request, base, time.Now(), bytes.NewReader(file))
			}
		// request config.json
		case path.Join(entry.Path, "swagger-config.json"):
			writer.Header().Set("Content-Type", "application/json")
			http.ServeContent(writer, request, "swagger-config.json", time.Now(), strings.NewReader(swConfigFileContents))
		// swagger spec config
		default:
			p = strings.TrimPrefix(p, entry.Path)
			value, ok := swaggerJsonFiles[p]

			if ok {
				http.ServeContent(writer, request, p, time.Now(), strings.NewReader(value))
			} else {
				http.NotFound(writer, request)
			}
		}
	}
}

// Init swagger config.
// This function do the things bellow:
// 1: List swagger files from entry.JSONPath.
// 2: Read user swagger json files and deduplicate.
// 3: Assign swagger contents into swaggerConfigJson variable
func (entry *SWEntry) initSwaggerConfig() {
	swaggerUrlConfig := &swUrlConfig{
		Urls: make([]*swUrl, 0),
	}

	if len(entry.JsonPath) > 0 {
		// 1: Add user API swagger JSON
		entry.listFilesWithSuffix(swaggerUrlConfig, entry.JsonPath, false)
	} else {
		// try to read from default directories
		// - docs
		// - api/gen/v1
		// - api/gen
		entry.listFilesWithSuffix(swaggerUrlConfig, "docs", true)
		entry.listFilesWithSuffix(swaggerUrlConfig, "api/gen/v1", true)
		entry.listFilesWithSuffix(swaggerUrlConfig, "api/gen", true)
	}

	// 2: Add rk common APIs
	if len(swAssetsFile) > 0 {
		key := entry.entryName + "-rk-common.swagger.json"
		// add common service json file
		swaggerJsonFiles[key] = string(swAssetsFile)
		e := &swUrl{
			Name: key,
			Url:  path.Join(entry.Path, key),
		}
		swaggerUrlConfig.Urls = append(swaggerUrlConfig.Urls, e)
	}

	// 3: Marshal to swagger-config.json and write to pkger
	bytes, err := json.Marshal(swaggerUrlConfig)
	if err != nil {
		ShutdownWithError(err)
	}

	swConfigFileContents = string(bytes)

	// 4: ignore swagger assets for middleware
	rkmid.AddPathToIgnoreGlobal(entry.Path)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (entry *SWEntry) listFilesWithSuffix(urlConfig *swUrlConfig, jsonPath string, ignoreError bool) {
	suffix := ".json"

	if entry.embedFS != nil {
		// 1: read dir
		files, err := entry.embedFS.ReadDir(jsonPath)
		if err != nil && !ignoreError {
			return
		}

		for i := range files {
			file := files[i]
			if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
				bytes, err := entry.embedFS.ReadFile(filepath.Join(jsonPath, file.Name()))
				key := entry.entryName + "-" + file.Name()

				if err != nil && !ignoreError {
					ShutdownWithError(err)
				}

				swaggerJsonFiles[key] = string(bytes)

				urlConfig.Urls = append(urlConfig.Urls, &swUrl{
					Name: key,
					Url:  path.Join(entry.Path, key),
				})
			}
		}

		return
	}

	// re-path it with working directory if not absolute path
	if !filepath.IsAbs(jsonPath) {
		wd, _ := os.Getwd()
		jsonPath = filepath.Join(wd, jsonPath)
	}

	files, err := ioutil.ReadDir(jsonPath)
	if err != nil && !ignoreError {
		return
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := os.ReadFile(filepath.Join(jsonPath, file.Name()))
			key := entry.entryName + "-" + file.Name()

			if err != nil && !ignoreError {
				ShutdownWithError(err)
			}

			swaggerJsonFiles[key] = string(bytes)

			urlConfig.Urls = append(urlConfig.Urls, &swUrl{
				Name: key,
				Url:  path.Join(entry.Path, key),
			})
		}
	}
}
