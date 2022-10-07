package rk

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"github.com/rookie-ninja/rk-entry/v3"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const SwaggerKind = "swagger"

// Inner struct used while initializing swagger entry.
type swUrlConfig struct {
	Urls []*swUrl `json:"urls" yaml:"urls"`
}

// Inner struct used while initializing swagger entry.
type swUrl struct {
	Name string `json:"name" yaml:"name"`
	Url  string `json:"url" yaml:"url"`
}

type SwaggerConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		Url  string `yaml:"url"`
		Spec struct {
			LocalPath string `yaml:"localPath"`
			EmbedPath string `yaml:"embedPath"`
		} `yaml:"spec"`
		Headers []string `yaml:"headers"`
	} `yaml:"entry"`
}

func (s *SwaggerConfig) JSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *SwaggerConfig) YAML() string {
	b, _ := yaml.Marshal(s)
	return string(b)
}

func (s *SwaggerConfig) Header() *EntryConfigHeader {
	return &s.EntryConfigHeader
}

func (s *SwaggerConfig) Register() (Entry, error) {
	if !s.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(s.Metadata.Domain) {
		return nil, nil
	}

	entry := &SwaggerEntry{
		config:  s,
		headers: map[string]string{},
	}
	entry.embedFS = Registry.EntryFS(entry.Kind(), entry.Name())

	for i := range s.Entry.Headers {
		header := s.Entry.Headers[i]
		tokens := strings.Split(header, ":")
		if len(tokens) == 2 {
			entry.headers[tokens[0]] = tokens[1]
		}
	}

	if len(s.Entry.Url) < 1 {
		s.Entry.Url = "/sw"
	}

	s.Entry.Url = path.Join("/", s.Entry.Url, "/")

	Registry.AddEntry(entry)

	return entry, nil
}

type SwaggerEntry struct {
	config        *SwaggerConfig
	embedFS       *embed.FS
	headers       map[string]string
	swaggerSpecs  map[string]string
	swaggerConfig string
}

func (s *SwaggerEntry) Category() string {
	return CategoryIndependent
}

func (s *SwaggerEntry) Kind() string {
	return s.config.Kind
}

func (s *SwaggerEntry) Name() string {
	return s.config.Metadata.Name
}

func (s *SwaggerEntry) Config() EntryConfig {
	return s.config
}

func (s *SwaggerEntry) Bootstrap(ctx context.Context) {
	s.initSwaggerSpec()
}

func (s *SwaggerEntry) Interrupt(ctx context.Context) {}

func (s *SwaggerEntry) Monitor() *Monitor {
	return nil
}

func (s *SwaggerEntry) FS() *embed.FS {
	return s.embedFS
}

func (s *SwaggerEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)

	res = append(res,
		&BuiltinApi{
			Method: http.MethodGet,
			Path:   path.Join("/", s.config.Entry.Url),
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				p := strings.TrimSuffix(request.URL.Path, "/")

				writer.Header().Set("cache-control", "no-cache")

				for k, v := range s.headers {
					writer.Header().Set(k, v)
				}

				switch p {
				case strings.TrimSuffix(s.config.Entry.Url, "/"):
					if file, err := rkembed.AssetsFS.ReadFile("assets/sw/index.html"); err != nil {
						http.Error(writer, "Internal server error", http.StatusInternalServerError)
					} else {
						http.ServeContent(writer, request, "index.html", time.Now(), bytes.NewReader(file))
					}
				case path.Join(s.config.Entry.Url, "swagger-config.json"):
					http.ServeContent(writer, request, "swagger-config.json", time.Now(), strings.NewReader(s.swaggerConfig))
				default:
					p = strings.TrimPrefix(p, s.config.Entry.Url)
					value, ok := s.swaggerSpecs[p]

					if ok {
						http.ServeContent(writer, request, p, time.Now(), strings.NewReader(value))
					} else {
						http.NotFound(writer, request)
					}
				}
			},
		})

	return res
}

func (s *SwaggerEntry) initSwaggerSpec() {
	swaggerUrlConfig := &swUrlConfig{
		Urls: make([]*swUrl, 0),
	}

	if len(s.config.Entry.Spec.EmbedPath) > 0 {
		// 1: Add user API swagger JSON
		s.listSpecFromEmbed(swaggerUrlConfig, s.config.Entry.Spec.EmbedPath, false)
	} else {
		// try to read from default directories
		// - docs
		// - api/gen/v1
		// - api/gen
		s.listSpecFromLocal(swaggerUrlConfig, "docs", true)
		s.listSpecFromLocal(swaggerUrlConfig, "api/gen/v1", true)
		s.listSpecFromLocal(swaggerUrlConfig, "api/gen", true)
	}

	// 2: Add rk common APIs
	if len(swaggerSpec) > 0 {
		key := s.Name() + "-rk-common.swagger.json"
		// add common service json file
		s.swaggerSpecs[key] = string(swaggerSpec)
		swaggerUrlConfig.Urls = append(swaggerUrlConfig.Urls, &swUrl{
			Name: key,
			Url:  path.Join(s.config.Entry.Url, key),
		})
	}

	// 3: Marshal to swagger-config.json
	bytes, err := json.Marshal(swaggerUrlConfig)
	if err != nil {
		rku.ShutdownWithError(err)
	}

	s.swaggerConfig = string(bytes)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (s *SwaggerEntry) listSpecFromEmbed(urlConfig *swUrlConfig, jsonPath string, ignoreError bool) {
	suffix := ".json"

	// 1: read dir
	files, err := s.embedFS.ReadDir(jsonPath)
	if err != nil && !ignoreError {
		return
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := s.embedFS.ReadFile(path.Join(jsonPath, file.Name()))
			key := s.config.Metadata.Name + "-" + file.Name()

			if err != nil && !ignoreError {
				rku.ShutdownWithError(err)
			}

			s.swaggerSpecs[key] = string(bytes)

			urlConfig.Urls = append(urlConfig.Urls, &swUrl{
				Name: key,
				Url:  path.Join(s.config.Entry.Url, key),
			})
		}
	}
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (s *SwaggerEntry) listSpecFromLocal(urlConfig *swUrlConfig, jsonPath string, ignoreError bool) {
	suffix := ".json"

	// re-path it with working directory if not absolute path
	if !path.IsAbs(jsonPath) {
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
			key := s.config.Metadata.Name + "-" + file.Name()

			if err != nil && !ignoreError {
				rku.ShutdownWithError(err)
			}

			s.swaggerSpecs[key] = string(bytes)

			urlConfig.Urls = append(urlConfig.Urls, &swUrl{
				Name: key,
				Url:  path.Join(s.config.Entry.Url, key),
			})
		}
	}
}
