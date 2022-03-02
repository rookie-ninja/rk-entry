// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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
	Enabled  bool     `yaml:"enabled" yaml:"enabled"`
	Path     string   `yaml:"path" yaml:"path"`
	JsonPath string   `yaml:"jsonPath" yaml:"jsonPath"`
	Headers  []string `yaml:"headers" yaml:"headers"`
}

// SWEntry implements rke.Entry interface.
type SWEntry struct {
	entryName        string            `json:"-" yaml:"-"`
	entryType        string            `json:"-" yaml:"-"`
	entryDescription string            `json:"-" yaml:"-"`
	JsonPath         string            `json:"-" yaml:"-"`
	Path             string            `json:"-" yaml:"-"`
	Headers          map[string]string `json:"-" yaml:"-"`
}

func RegisterSWEntry(boot *BootSW) *SWEntry {
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
			entryType:        "SwEntry",
			entryDescription: "Internal RK entry for swagger UI.",
			Path:             boot.Path,
			JsonPath:         boot.JsonPath,
			Headers:          headers,
		}

		// Deal with Path
		// add "/" at start and end side if missing
		if !strings.HasPrefix(swEntry.Path, "/") {
			swEntry.Path = "/" + swEntry.Path
		}

		if !strings.HasSuffix(swEntry.Path, "/") {
			swEntry.Path = swEntry.Path + "/"
		}

		// init swagger configs
		swEntry.initSwaggerConfig()
	}

	return swEntry
}

func (entry *SWEntry) Bootstrap(ctx context.Context) {}

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
		case path.Join(entry.Path, "swagger-config.json"):
			http.ServeContent(writer, request, "swagger-config.json", time.Now(), strings.NewReader(swConfigFileContents))
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
		swaggerUrlConfig.Urls = append(swaggerUrlConfig.Urls, &swUrl{
			Name: key,
			Url:  path.Join(entry.Path, key),
		})
	}

	// 3: Marshal to swagger-config.json and write to pkger
	bytes, err := json.Marshal(swaggerUrlConfig)
	if err != nil {
		ShutdownWithError(err)
	}

	swConfigFileContents = string(bytes)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (entry *SWEntry) listFilesWithSuffix(urlConfig *swUrlConfig, jsonPath string, ignoreError bool) {
	suffix := ".json"
	// re-path it with working directory if not absolute path
	if !path.IsAbs(entry.JsonPath) {
		wd, _ := os.Getwd()
		jsonPath = path.Join(wd, jsonPath)
	}

	files, err := ioutil.ReadDir(jsonPath)
	if err != nil && !ignoreError {
		return
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := ioutil.ReadFile(path.Join(jsonPath, file.Name()))
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

func readFileFromEmbed(filePath string) []byte {
	var file fs.File
	var err error

	if file, err = rkembed.AssetsFS.Open(filePath); err != nil {
		return []byte{}
	}

	var bytes []byte
	if bytes, err = ioutil.ReadAll(file); err != nil {
		return []byte{}
	}

	return bytes
}
