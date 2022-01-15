package rkentry

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/markbates/pkger"
	"github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"path"
	"strings"
	"time"
)

var (
	// Templates is a map to store go template
	Templates = map[string][]byte{}
)

const (
	// TvEntryType default entry type
	TvEntryType = "TvEntry"
	// TvEntryNameDefault default entry name
	TvEntryNameDefault = "TvDefault"
	// TvEntryDescription default entry description
	TvEntryDescription = "Internal RK entry which implements RK TV web UI."
)

// Read go TV related template files into memory.
func init() {
	Templates["header"] = readFileFromPkger(ModPath, "/assets/tv/header.tmpl")
	Templates["footer"] = readFileFromPkger(ModPath, "/assets/tv/footer.tmpl")
	Templates["aside"] = readFileFromPkger(ModPath, "/assets/tv/aside.tmpl")
	Templates["head"] = readFileFromPkger(ModPath, "/assets/tv/head.tmpl")
	Templates["svg-sprite"] = readFileFromPkger(ModPath, "/assets/tv/svg-sprite.tmpl")
	Templates["overview"] = readFileFromPkger(ModPath, "/assets/tv/overview.tmpl")
	Templates["apis"] = readFileFromPkger(ModPath, "/assets/tv/apis.tmpl")
	Templates["entries"] = readFileFromPkger(ModPath, "/assets/tv/entries.tmpl")
	Templates["configs"] = readFileFromPkger(ModPath, "/assets/tv/configs.tmpl")
	Templates["certs"] = readFileFromPkger(ModPath, "/assets/tv/certs.tmpl")
	Templates["not-found"] = readFileFromPkger(ModPath, "/assets/tv/not-found.tmpl")
	Templates["internal-error"] = readFileFromPkger(ModPath, "/assets/tv/internal-error.tmpl")
	Templates["os"] = readFileFromPkger(ModPath, "/assets/tv/os.tmpl")
	Templates["env"] = readFileFromPkger(ModPath, "/assets/tv/env.tmpl")
	Templates["prometheus"] = readFileFromPkger(ModPath, "/assets/tv/prometheus.tmpl")
	Templates["deps"] = readFileFromPkger(ModPath, "/assets/tv/deps.tmpl")
	Templates["license"] = readFileFromPkger(ModPath, "/assets/tv/license.tmpl")
	Templates["info"] = readFileFromPkger(ModPath, "/assets/tv/info.tmpl")
	Templates["logs"] = readFileFromPkger(ModPath, "/assets/tv/logs.tmpl")
	Templates["gw-error-mapping"] = readFileFromPkger(ModPath, "/assets/tv/error-mapping.tmpl")
	Templates["git"] = readFileFromPkger(ModPath, "/assets/tv/git.tmpl")
}

// BootConfigTv Bootstrap config of tv.
// 1: Enabled: Enable tv service.
type BootConfigTv struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// TvEntry RK TV entry supports web UI for application & process information.
// 1: EntryName: Name of entry.
// 2: EntryType: Type of entry.
// 2: EntryDescription: Description of entry.
// 3: ZapLoggerEntry: ZapLoggerEntry used for logging.
// 4: EventLoggerEntry: EventLoggerEntry used for logging.
// 5: Template: GO template for rendering web UI.
type TvEntry struct {
	EntryName        string             `json:"entryName" yaml:"entryName"`
	EntryType        string             `json:"entryType" yaml:"entryType"`
	EntryDescription string             `json:"-" yaml:"-"`
	ZapLoggerEntry   *ZapLoggerEntry    `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry  `json:"-" yaml:"-"`
	Template         *template.Template `json:"-" yaml:"-"`
	AssetsFilePath   string             `json:"-" yaml:"-"`
	BasePath         string             `json:"-" yaml:"-"`
}

// TvEntryOption TV entry option.
type TvEntryOption func(entry *TvEntry)

// WithNameTv Provide name.
func WithNameTv(name string) TvEntryOption {
	return func(entry *TvEntry) {
		entry.EntryName = name
	}
}

// WithEventLoggerEntryTv Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryTv(eventLoggerEntry *EventLoggerEntry) TvEntryOption {
	return func(entry *TvEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// WithZapLoggerEntryTv Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryTv(zapLoggerEntry *ZapLoggerEntry) TvEntryOption {
	return func(entry *TvEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

func RegisterTvEntryWithConfig(config *BootConfigTv, name string, zap *ZapLoggerEntry, event *EventLoggerEntry) *TvEntry {
	var tvEntry *TvEntry
	if config.Enabled {
		tvEntry = RegisterTvEntry(
			WithNameTv(name),
			WithZapLoggerEntryTv(zap),
			WithEventLoggerEntryTv(event))
	}

	return tvEntry
}

// RegisterTvEntry Create new TV entry with options.
func RegisterTvEntry(opts ...TvEntryOption) *TvEntry {
	entry := &TvEntry{
		EntryName:        TvEntryNameDefault,
		EntryType:        TvEntryType,
		EntryDescription: TvEntryDescription,
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		AssetsFilePath:   "/rk/v1/assets/tv/",
		BasePath:         "/rk/v1/tv/",
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = TvEntryNameDefault
	}

	return entry
}

// AssetsFileHandler Handler which returns js, css, images and html files for TV web UI.
func (entry *TvEntry) AssetsFileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		p := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/rk/v1"), "/")

		if file, err := pkger.Open(path.Join("/", p)); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
		} else {
			http.ServeContent(writer, request, path.Base(p), time.Now(), file)
		}
	}
}

// Bootstrap TV entry.
func (entry *TvEntry) Bootstrap(context.Context) {
	entry.Template = template.New("rk-tv")

	// Parse templates
	for _, v := range Templates {
		if _, err := entry.Template.Parse(string(v)); err != nil {
			rkcommon.ShutdownWithError(err)
		}
	}
}

// Interrupt TV entry.
func (entry *TvEntry) Interrupt(context.Context) {
	// Noop
}

// GetName Get name of entry.
func (entry *TvEntry) GetName() string {
	return entry.EntryName
}

// GetType Get type of entry.
func (entry *TvEntry) GetType() string {
	return entry.EntryType
}

// GetDescription Get description of entry.
func (entry *TvEntry) GetDescription() string {
	return entry.EntryDescription
}

// String Stringfy entry.
func (entry *TvEntry) String() string {
	bytesStr, _ := json.Marshal(entry)
	return string(bytesStr)
}

// MarshalJSON Marshal entry
func (entry *TvEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"eventLoggerEntry": entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":   entry.ZapLoggerEntry.GetName(),
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *TvEntry) UnmarshalJSON([]byte) error {
	return nil
}

// TV handler
// @Summary Get HTML page of /tv
// @Id 15
// @version 1.0
// @Security ApiKeyAuth
// @Security BasicAuth
// @produce text/html
// @Success 200 string HTML
// @Router /rk/v1/tv [get]
func (entry *TvEntry) noop() {}

func (entry *TvEntry) Action(subPath string, logger *zap.Logger) *bytes.Buffer {
	var buf *bytes.Buffer

	switch subPath {
	case "", "/", "/overview", "/application", "overview", "application":
		buf = entry.ExecuteTemplate("overview", doReadme(), logger)
	case "/entries", "entries":
		buf = entry.ExecuteTemplate("entries", doEntries(), logger)
	case "/configs", "configs":
		buf = entry.ExecuteTemplate("configs", doConfigs(), logger)
	case "/certs", "certs":
		buf = entry.ExecuteTemplate("certs", doCerts(), logger)
	case "/os", "os":
		buf = entry.ExecuteTemplate("os", doSys(), logger)
	case "/env", "env":
		buf = entry.ExecuteTemplate("env", doSys(), logger)
	case "/prometheus", "prometheus":
		buf = entry.ExecuteTemplate("prometheus", nil, logger)
	case "/logs", "logs":
		buf = entry.ExecuteTemplate("logs", doLogs(), logger)
	case "/deps", "deps":
		buf = entry.ExecuteTemplate("deps", doDeps(), logger)
	case "/license", "license":
		buf = entry.ExecuteTemplate("license", doLicense(), logger)
	case "/info", "info":
		buf = entry.ExecuteTemplate("info", doInfo(), logger)
	case "/git", "git":
		buf = entry.ExecuteTemplate("git", doGit(), logger)
	default:
		buf = entry.ExecuteTemplate("not-found", nil, logger)
	}

	return buf
}

// Execute go template into buffer.
func (entry *TvEntry) ExecuteTemplate(templateName string, data interface{}, logger *zap.Logger) *bytes.Buffer {
	buf := new(bytes.Buffer)

	if err := entry.Template.ExecuteTemplate(buf, templateName, data); err != nil {
		logger.Warn("Failed to execute template", zap.Error(err))
		buf.Reset()
		entry.Template.ExecuteTemplate(buf, "internal-error", nil)
	}

	return buf
}
