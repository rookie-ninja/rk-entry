package rkentry

import (
	"context"
	"encoding/json"
	"github.com/markbates/pkger"
	"github.com/markbates/pkger/pkging"
	"github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	swaggerJsonFiles     = make(map[string]string, 0)
	swConfigFileContents = ``
)

const (
	// SwEntryType default entry type
	SwEntryType = "SwEntry"
	// SwEntryNameDefault default entry name
	SwEntryNameDefault = "SwDefault"
	// SwEntryDescription default entry description
	SwEntryDescription = "Internal RK entry for swagger UI."
	// ModPath used while reading files from pkger
	ModPath = "github.com/rookie-ninja/rk-entry"
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

// BootConfigSw Bootstrap config of swagger.
// 1: Enabled: Enable swagger.
// 2: Path: Swagger path accessible from restful API.
// 3: JsonPath: The path of where swagger JSON file was located.
// 4: Headers: The headers that would added into each API response.
type BootConfigSw struct {
	Enabled  bool     `yaml:"enabled" yaml:"enabled"`
	Path     string   `yaml:"path" yaml:"path"`
	JsonPath string   `yaml:"jsonPath" yaml:"jsonPath"`
	Headers  []string `yaml:"headers" yaml:"headers"`
}

// SwEntry implements rkentry.Entry interface.
// 1: Path: Swagger path accessible from restful API.
// 2: JsonPath: The path of where swagger JSON file was located.
// 3: Headers: The headers that would added into each API response.
// 4: Port: The port where swagger would listen to.
// 5: EnableCommonService: Enable common service in swagger.
type SwEntry struct {
	EntryName           string            `json:"entryName" yaml:"entryName"`
	EntryType           string            `json:"entryType" yaml:"entryType"`
	EntryDescription    string            `json:"-" yaml:"-"`
	EventLoggerEntry    *EventLoggerEntry `json:"-" yaml:"-"`
	ZapLoggerEntry      *ZapLoggerEntry   `json:"-" yaml:"-"`
	JsonPath            string            `json:"jsonPath" yaml:"jsonPath"`
	Path                string            `json:"path" yaml:"path"`
	Headers             map[string]string `json:"-" yaml:"-"`
	Port                uint64            `json:"port" yaml:"port"`
	EnableCommonService bool              `json:"-" yaml:"-"`
	AssetsFilePath      string            `json:"-" yaml:"-"`
}

// SwOption Swagger entry option.
type SwOption func(*SwEntry)

// WithPortSw Provide port.
func WithPortSw(port uint64) SwOption {
	return func(entry *SwEntry) {
		entry.Port = port
	}
}

// WithNameSw Provide name.
func WithNameSw(name string) SwOption {
	return func(entry *SwEntry) {
		entry.EntryName = name
	}
}

// WithPathSw Provide path.
func WithPathSw(path string) SwOption {
	return func(entry *SwEntry) {
		if len(path) < 1 {
			path = "sw"
		}
		entry.Path = path
	}
}

// WithJsonPathSw Provide JsonPath.
func WithJsonPathSw(path string) SwOption {
	return func(entry *SwEntry) {
		entry.JsonPath = path
	}
}

// WithHeadersSw Provide headers.
func WithHeadersSw(headers map[string]string) SwOption {
	return func(entry *SwEntry) {
		entry.Headers = headers
	}
}

// WithZapLoggerEntrySw Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntrySw(zapLoggerEntry *ZapLoggerEntry) SwOption {
	return func(entry *SwEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// WithEventLoggerEntrySw Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntrySw(eventLoggerEntry *EventLoggerEntry) SwOption {
	return func(entry *SwEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// WithEnableCommonServiceSw Provide enable common service option.
func WithEnableCommonServiceSw(enable bool) SwOption {
	return func(entry *SwEntry) {
		entry.EnableCommonService = enable
	}
}

func RegisterSwEntryWithConfig(config *BootConfigSw,
	name string, port uint64,
	zap *ZapLoggerEntry,
	event *EventLoggerEntry,
	commonServiceEnabled bool) *SwEntry {
	var swEntry *SwEntry
	if config.Enabled {
		// Init swagger custom headers from config
		headers := make(map[string]string, 0)
		for i := range config.Headers {
			header := config.Headers[i]
			tokens := strings.Split(header, ":")
			if len(tokens) == 2 {
				headers[tokens[0]] = tokens[1]
			}
		}

		swEntry = RegisterSwEntry(
			WithNameSw(name),
			WithZapLoggerEntrySw(zap),
			WithEventLoggerEntrySw(event),
			WithEnableCommonServiceSw(commonServiceEnabled),
			WithPortSw(port),
			WithPathSw(config.Path),
			WithJsonPathSw(config.JsonPath),
			WithHeadersSw(headers))
	}

	return swEntry
}

func RegisterSwEntry(opts ...SwOption) *SwEntry {
	entry := &SwEntry{
		EntryName:        SwEntryNameDefault,
		EntryType:        SwEntryType,
		EntryDescription: SwEntryDescription,
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		Path:             "sw",
		AssetsFilePath:   "/rk/v1/assets/sw/",
	}

	for i := range opts {
		opts[i](entry)
	}

	// Deal with Path
	// add "/" at start and end side if missing
	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if !strings.HasSuffix(entry.Path, "/") {
		entry.Path = entry.Path + "/"
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "SwEntry-" + strconv.FormatUint(entry.Port, 10)
	}

	// init swagger configs
	entry.initSwaggerConfig()

	return entry
}

func (entry *SwEntry) Bootstrap(ctx context.Context) {
	// Noop
}

func (entry *SwEntry) Interrupt(ctx context.Context) {
	// Noop
}

func (entry *SwEntry) GetName() string {
	return entry.EntryName
}

func (entry *SwEntry) GetType() string {
	return entry.EntryType
}

func (entry *SwEntry) GetDescription() string {
	return entry.EntryDescription
}

func (entry *SwEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry
func (entry *SwEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":           entry.EntryName,
		"entryType":           entry.EntryType,
		"entryDescription":    entry.EntryDescription,
		"eventLoggerEntry":    entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":      entry.ZapLoggerEntry.GetName(),
		"jsonPath":            entry.JsonPath,
		"port":                entry.Port,
		"path":                entry.Path,
		"headers":             entry.Headers,
		"enableCommonService": entry.EnableCommonService,
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Unmarshal entry
func (entry *SwEntry) UnmarshalJSON([]byte) error {
	return nil
}

// AssetsFileHandler Handler for swagger assets files.
func (entry *SwEntry) AssetsFileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		p := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/rk/v1"), "/")

		if file, err := openFromPkger(ModPath, p); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
		} else {
			http.ServeContent(writer, request, path.Base(p), time.Now(), file)
		}
	}
}

// ConfigFileHandler handler for swagger config files.
func (entry *SwEntry) ConfigFileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		p := strings.TrimSuffix(request.URL.Path, "/")

		writer.Header().Set("cache-control", "no-cache")

		for k, v := range entry.Headers {
			writer.Header().Set(k, v)
		}

		switch p {
		case strings.TrimSuffix(entry.Path, "/"):
			if file, err := openFromPkger(ModPath, "/assets/sw/index.html"); err != nil {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			} else {
				http.ServeContent(writer, request, "index.html", time.Now(), file)
			}
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
func (entry *SwEntry) initSwaggerConfig() {
	swaggerUrlConfig := &swUrlConfig{
		Urls: make([]*swUrl, 0),
	}

	// 1: Add user API swagger JSON
	entry.listFilesWithSuffix(swaggerUrlConfig)

	// 2: Add rk common APIs
	if entry.EnableCommonService {
		key := entry.EntryName + "-rk-common.swagger.json"
		// add common service json file
		swaggerJsonFiles[key] = string(readFileFromPkger(ModPath, "/assets/sw/config/swagger.json"))
		swaggerUrlConfig.Urls = append(swaggerUrlConfig.Urls, &swUrl{
			Name: key,
			Url:  path.Join(entry.Path, key),
		})
	}

	// 3: Marshal to swagger-config.json and write to pkger
	bytes, err := json.Marshal(swaggerUrlConfig)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Error("Failed to unmarshal swagger-config.json",
			zap.Error(err))
		rkcommon.ShutdownWithError(err)
	}

	swConfigFileContents = string(bytes)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (entry *SwEntry) listFilesWithSuffix(urlConfig *swUrlConfig) {
	jsonPath := entry.JsonPath
	suffix := ".json"
	// re-path it with working directory if not absolute path
	if !path.IsAbs(entry.JsonPath) {
		wd, err := os.Getwd()
		if err != nil {
			entry.ZapLoggerEntry.GetLogger().Info("Failed to get working directory",
				zap.String("error", err.Error()))
			rkcommon.ShutdownWithError(err)
		}
		jsonPath = path.Join(wd, jsonPath)
	}

	files, err := ioutil.ReadDir(jsonPath)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Error("Failed to list files with suffix",
			zap.String("path", jsonPath),
			zap.String("suffix", suffix),
			zap.String("error", err.Error()))
		return
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := ioutil.ReadFile(path.Join(jsonPath, file.Name()))
			key := entry.EntryName + "-" + file.Name()

			if err != nil {
				entry.ZapLoggerEntry.GetLogger().Info("Failed to read file with suffix",
					zap.String("path", path.Join(jsonPath, key)),
					zap.String("suffix", suffix),
					zap.String("error", err.Error()))
				rkcommon.ShutdownWithError(err)
			}

			swaggerJsonFiles[key] = string(bytes)

			urlConfig.Urls = append(urlConfig.Urls, &swUrl{
				Name: key,
				Url:  path.Join(entry.Path, key),
			})
		}
	}
}

// Read go template files with Pkger.
func readFileFromPkger(modPath, filePath string) []byte {
	var file pkging.File
	var err error

	if file, err = pkger.Open(path.Join(modPath+":", filePath)); err != nil {
		return []byte{}
	}

	var bytes []byte
	if bytes, err = ioutil.ReadAll(file); err != nil {
		return []byte{}
	}

	return bytes
}

func openFromPkger(modPath, filePath string) (pkging.File, error) {
	return pkger.Open(path.Join(modPath + ":" + filePath))
}
