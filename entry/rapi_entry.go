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

const RapiKind = "rapi"

// Inner struct used while initializing swagger entry.
type rapiConfig struct {
	Specs []*rapiSpec `json:"specs" yaml:"specs"`
	Style struct {
		Theme       string `yaml:"theme" json:"theme"`
		RenderStyle string `yaml:"renderStyle" json:"renderStyle"`
		AllowTry    bool   `yaml:"allowTry" json:"allowTry"`
		BgColor     string `yaml:"bgColor" json:"bgColor"`
	} `json:"style" yaml:"style"`
}

// Inner struct used while initializing open API entry.
type rapiSpec struct {
	Name string `json:"name" yaml:"name"`
	Url  string `json:"url" yaml:"url"`
}

type RapiDocConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		Url  string `yaml:"url"`
		Spec struct {
			LocalPath string `yaml:"localPath"`
			EmbedPath string `yaml:"embedPath"`
		} `yaml:"spec"`
		Headers []string `yaml:"headers"`
		Style   struct {
			Theme string `yaml:"theme" json:"theme"`
		} `yaml:"style" json:"style"`
		Debug bool `yaml:"debug" json:"debug"`
	} `yaml:"entry"`
}

func (r *RapiDocConfig) JSON() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *RapiDocConfig) YAML() string {
	b, _ := yaml.Marshal(r)
	return string(b)
}

func (r *RapiDocConfig) Header() *EntryConfigHeader {
	return &r.EntryConfigHeader
}

func (r *RapiDocConfig) Register() (Entry, error) {
	if !r.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(r.Metadata.Domain) {
		return nil, nil
	}

	entry := &RapiDocEntry{
		config:  r,
		headers: map[string]string{},
	}
	entry.embedFS = Registry.EntryFS(entry.Kind(), entry.Name())

	for i := range r.Entry.Headers {
		header := r.Entry.Headers[i]
		tokens := strings.Split(header, ":")
		if len(tokens) == 2 {
			entry.headers[tokens[0]] = tokens[1]
		}
	}

	if len(r.Entry.Url) < 1 {
		r.Entry.Url = "/rapi"
	}

	r.Entry.Url = path.Join("/", r.Entry.Url, "/")

	if r.Entry.Style.Theme != "light" && r.Entry.Style.Theme != "dark" {
		r.Entry.Style.Theme = "light"
	}

	Registry.AddEntry(entry)

	return entry, nil
}

type RapiDocEntry struct {
	config     *RapiDocConfig
	embedFS    *embed.FS
	headers    map[string]string
	rapiSpecs  map[string]string
	rapiConfig string
}

func (r *RapiDocEntry) Category() string {
	return CategoryIndependent
}

func (r *RapiDocEntry) Kind() string {
	return r.config.Kind
}

func (r *RapiDocEntry) Name() string {
	return r.config.Metadata.Name
}

func (r *RapiDocEntry) Config() EntryConfig {
	return r.config
}

func (r *RapiDocEntry) Bootstrap(ctx context.Context) {
	r.initRapiConfig()
}

func (r *RapiDocEntry) Interrupt(ctx context.Context) {}

func (r *RapiDocEntry) Monitor() *Monitor {
	return nil
}

func (r *RapiDocEntry) FS() *embed.FS {
	return r.embedFS
}

func (r *RapiDocEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)

	res = append(res,
		&BuiltinApi{
			Method: http.MethodGet,
			Path:   path.Join("/", r.config.Entry.Url),
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				p := strings.TrimPrefix(strings.TrimSuffix(request.URL.Path, "/"), strings.TrimSuffix(r.config.Entry.Url, "/"))
				p = strings.TrimSuffix(p, "/")
				p = strings.TrimPrefix(p, "/")

				writer.Header().Set("cache-control", "no-cache")

				for k, v := range r.headers {
					writer.Header().Set(k, v)
				}

				switch p {
				case "":
					if file, err := rkembed.AssetsFS.ReadFile("assets/docs/index.html"); err != nil {
						http.Error(writer, "Internal server error", http.StatusInternalServerError)
					} else {
						http.ServeContent(writer, request, "index.html", time.Now(), bytes.NewReader(file))
					}
				case "specs":
					http.ServeContent(writer, request, "specs", time.Now(), strings.NewReader(r.rapiConfig))
				default:
					value, ok := r.rapiSpecs[p]
					if ok {
						http.ServeContent(writer, request, p, time.Now(), strings.NewReader(value))
						return
					}

					http.NotFound(writer, request)
				}
			},
		})

	return res
}

func (r *RapiDocEntry) initRapiConfig() {
	config := &rapiConfig{
		Specs: []*rapiSpec{},
	}

	if len(r.config.Entry.Spec.EmbedPath) > 0 {
		// 1: Add user API swagger JSON
		r.listSpecFromEmbed(config, r.config.Entry.Spec.EmbedPath, false)
	} else {
		// try to read from default directories
		// - docs
		// - api/gen/v1
		// - api/gen
		r.listSpecFromLocal(config, "docs", true)
		r.listSpecFromLocal(config, "api/gen/v1", true)
		r.listSpecFromLocal(config, "api/gen", true)
	}

	// 2: Add rk common APIs
	if len(swaggerSpec) > 0 {
		key := r.Name() + "-rk-common.swagger.json"
		// add common service json file
		r.rapiSpecs[key] = string(swaggerSpec)
		config.Specs = append(config.Specs, &rapiSpec{
			Name: key,
			Url:  path.Join(r.config.Entry.Url, key),
		})
	}

	// 3: Assign style
	config.Style.Theme = r.config.Entry.Style.Theme
	config.Style.RenderStyle = "focused"
	config.Style.AllowTry = false
	if config.Style.Theme == "light" {
		config.Style.BgColor = "#FAFAFA"
	}

	if r.config.Entry.Debug {
		config.Style.RenderStyle = "focused"
		config.Style.AllowTry = true
	}

	// 3: Marshal to swagger-config.json
	bytes, err := json.Marshal(config)
	if err != nil {
		rku.ShutdownWithError(err)
	}

	r.rapiConfig = string(bytes)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (r *RapiDocEntry) listSpecFromEmbed(rapi *rapiConfig, jsonPath string, ignoreError bool) {
	suffix := ".json"

	// 1: read dir
	files, err := r.embedFS.ReadDir(jsonPath)
	if err != nil && !ignoreError {
		return
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := r.embedFS.ReadFile(path.Join(jsonPath, file.Name()))
			key := r.config.Metadata.Name + "-" + file.Name()

			if err != nil && !ignoreError {
				rku.ShutdownWithError(err)
			}

			r.rapiSpecs[key] = string(bytes)

			rapi.Specs = append(rapi.Specs, &rapiSpec{
				Name: key,
				Url:  path.Join(r.config.Entry.Url, key),
			})
		}
	}
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (r *RapiDocEntry) listSpecFromLocal(rapi *rapiConfig, jsonPath string, ignoreError bool) {
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
			key := r.config.Metadata.Name + "-" + file.Name()

			if err != nil && !ignoreError {
				rku.ShutdownWithError(err)
			}

			r.rapiSpecs[key] = string(bytes)

			rapi.Specs = append(rapi.Specs, &rapiSpec{
				Name: key,
				Url:  path.Join(r.config.Entry.Url, key),
			})
		}
	}
}
