// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	rkembed "github.com/rookie-ninja/rk-entry"
	"github.com/rookie-ninja/rk-entry/error"
	"html/template"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var exToIcon = map[string]string{
	// folder
	"folder": "folder.png",
	// compressed file
	"ar":  "pkg.png",
	"zip": "pkg.png",
	"rar": "pkg.png",
	"gz":  "pkg.png",
	"xz":  "pkg.png",
	"gz2": "pkg.png",
	"tar": "pkg.png",
	"dep": "pkg.png",
	"rpm": "pkg.png",
	// image file
	"jpg":  "image.png",
	"jpeg": "image.png",
	"png":  "image.png",
	"gif":  "image.png",
	"svg":  "image.png",
	// audio
	"mp3":  "audio.png",
	"wav":  "image.png",
	"ogg":  "image.png",
	"flac": "image.png",
	// pdf
	"pdf": "pdf.png",
	// docs
	"xls":  "doc.png",
	"odt":  "doc.png",
	"ods":  "doc.png",
	"doc":  "doc.png",
	"docx": "doc.png",
	"xlsx": "doc.png",
	"ppt":  "doc.png",
	"txt":  "doc.png",
	// unknown
	"unknown": "file.png",
}

// BootStaticFileHandler bootstrap config of StaticHandler.
type BootStaticFileHandler struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	Path       string `yaml:"path" json:"path"`
	SourceType string `yaml:"sourceType" json:"sourceType"`
	SourcePath string `yaml:"sourcePath" json:"sourcePath"`
}

// StaticFileHandlerEntry Static file handler entry supports web UI for downloading static files.
type StaticFileHandlerEntry struct {
	entryName        string             `yaml:"-" json:"-"`
	entryType        string             `yaml:"-" json:"-"`
	entryDescription string             `yaml:"-" json:"-"`
	Path             string             `yaml:"-" json:"-"`
	Template         *template.Template `json:"-" yaml:"-"`
	httpFS           http.FileSystem    `yaml:"-" json:"-"`
}

// StaticFileHandlerEntryOption options for StaticFileHandlerEntry
type StaticFileHandlerEntryOption func(entry *StaticFileHandlerEntry)

// WithNameStaticFileHandlerEntry provide entry name
func WithNameStaticFileHandlerEntry(name string) StaticFileHandlerEntryOption {
	return func(entry *StaticFileHandlerEntry) {
		entry.entryName = name
	}
}

// RegisterStaticFileHandlerEntry Create new static file handler entry with config
func RegisterStaticFileHandlerEntry(boot *BootStaticFileHandler, opts ...StaticFileHandlerEntryOption) *StaticFileHandlerEntry {
	if !boot.Enabled {
		return nil
	}

	entry := &StaticFileHandlerEntry{
		entryName:        "StaticFileHandler",
		entryType:        StaticFileHandlerEntryType,
		entryDescription: "Internal RK entry which implements static file handler.",
		Template:         template.New("rk-static"),
		Path:             boot.Path,
		httpFS:           http.Dir(""),
	}

	for i := range opts {
		opts[i](entry)
	}

	if fs := GlobalAppCtx.GetEmbedFS(entry.GetType(), entry.GetName()); fs != nil {
		entry.httpFS = http.FS(fs)
	}

	switch boot.SourceType {
	case "local":
		if !filepath.IsAbs(boot.SourcePath) {
			wd, _ := os.Getwd()
			boot.SourcePath = path.Join(wd, boot.SourcePath)
		}
		entry.httpFS = http.Dir(boot.SourcePath)
	}

	if len(entry.Path) < 1 {
		entry.Path = "/static"
	}

	// Deal with Path
	// add "/" at start and end side if missing
	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if !strings.HasSuffix(entry.Path, "/") {
		entry.Path = entry.Path + "/"
	}

	return entry
}

// Bootstrap entry.
func (entry *StaticFileHandlerEntry) Bootstrap(context.Context) {
	// parse template
	if _, err := entry.Template.Parse(string(readFile("assets/static/index.tmpl", &rkembed.AssetsFS, true))); err != nil {
		ShutdownWithError(err)
	}
}

// Interrupt entry.
func (entry *StaticFileHandlerEntry) Interrupt(context.Context) {
	// Noop
}

// GetName Get name of entry.
func (entry *StaticFileHandlerEntry) GetName() string {
	return entry.entryName
}

// GetType Get entry type.
func (entry *StaticFileHandlerEntry) GetType() string {
	return entry.entryType
}

// GetDescription Get description of entry.
func (entry *StaticFileHandlerEntry) GetDescription() string {
	return entry.entryDescription
}

// String Stringfy entry.
func (entry *StaticFileHandlerEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry.
func (entry *StaticFileHandlerEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.GetName(),
		"type":        entry.GetType(),
		"description": entry.GetDescription(),
		"path":        entry.Path,
	}

	return json.Marshal(m)
}

// UnmarshalJSON Not supported.
func (entry *StaticFileHandlerEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetFileHandler handles requests sent from user.
func (entry *StaticFileHandlerEntry) GetFileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !strings.HasSuffix(request.URL.Path, "/") {
			request.URL.Path = request.URL.Path + "/"
		}

		// Trim prefix with path user defined in order to get file path
		p := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, entry.Path), "/")

		if len(p) < 1 {
			p = "/"
		}
		p = path.Join("/", p)

		var file http.File
		var err error
		// open file
		if file, err = entry.httpFS.Open(p); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			bytes, _ := json.Marshal(rkerror.NewInternalError("Failed to open file", err))
			writer.Write(bytes)
			return
		}

		// get file info
		fileInfo, err := file.Stat()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			bytes, _ := json.Marshal(rkerror.NewInternalError("Failed to stat file", err))
			writer.Write(bytes)
			return
		}

		// list files if file is directory
		if fileInfo.IsDir() {
			infos, _ := file.Readdir(math.MaxInt32)
			files := make([]*fileResp, 0)

			for _, v := range infos {
				files = append(files, &fileResp{
					isDir:    v.IsDir(),
					Icon:     base64.StdEncoding.EncodeToString(readFile(path.Join("assets/static/icons", entry.getIconPath(v)), &rkembed.AssetsFS, false)),
					FileUrl:  path.Join(entry.Path, p, v.Name()),
					FileName: v.Name(),
					Size:     v.Size(),
					ModTime:  v.ModTime(),
				})
			}

			entry.sortFiles(files)
			resp := &resp{
				PrevPath: path.Join(entry.Path, path.Dir(p)),
				PrevIcon: base64.StdEncoding.EncodeToString(readFile(path.Join("assets/static/icons/folder.png"), &rkembed.AssetsFS, false)),
				Path:     p,
				Files:    files,
			}

			buf := new(bytes.Buffer)
			if err := entry.Template.ExecuteTemplate(buf, "index", resp); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				bytes, _ := json.Marshal(rkerror.NewInternalError("Failed to execute go template", err))
				writer.Write(bytes)
				return
			}

			writer.WriteHeader(http.StatusOK)
			writer.Write(buf.Bytes())
		} else {
			// make browser download file
			writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
			writer.Header().Set("Content-Type", "application/octet-stream")
			http.ServeContent(writer, request, path.Base(p), time.Now(), file)
		}
	}
}

// sort file response
func (entry *StaticFileHandlerEntry) sortFiles(res []*fileResp) {
	sort.SliceStable(res, func(i, j int) bool {
		if res[i].isDir && res[j].isDir {
			return strings.Compare(res[i].FileName, res[j].FileName) < 0
		}

		if res[i].isDir {
			return true
		}

		if res[j].isDir {
			return false
		}

		return strings.Compare(res[i].FileName, res[j].FileName) < 0
	})
}

// get icon path based on file information
func (entry *StaticFileHandlerEntry) getIconPath(info fs.FileInfo) string {
	if info.IsDir() {
		return exToIcon["folder"]
	}

	ex := strings.TrimPrefix(filepath.Ext(info.Name()), ".")
	res := exToIcon[ex]

	if len(res) < 1 {
		return exToIcon["unknown"]
	}

	return res
}

// response for inner
type resp struct {
	PrevPath string
	PrevIcon string
	Path     string
	Files    []*fileResp
}

// file response for inner
type fileResp struct {
	isDir    bool
	FileName string
	FileUrl  string
	Icon     string
	Size     int64
	ModTime  time.Time
}
