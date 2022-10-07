package rk

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/rookie-ninja/rk-entry/v3"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"gopkg.in/yaml.v3"
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

const FileKind = "file"

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

type FileConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		Url        string `yaml:"url"`
		SourceType string `yaml:"sourceType"`
		SourcePath string `yaml:"sourcePath"`
	} `yaml:"entry"`
}

func (f *FileConfig) JSON() string {
	b, _ := json.Marshal(f)
	return string(b)
}

func (f *FileConfig) YAML() string {
	b, _ := yaml.Marshal(f)
	return string(b)
}

func (f *FileConfig) Header() *EntryConfigHeader {
	return &f.EntryConfigHeader
}

func (f *FileConfig) Register() (Entry, error) {
	if !f.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(f.Metadata.Domain) {
		return nil, nil
	}

	entry := &FileEntry{
		config: f,
		tpl:    template.New("rk-static"),
		httpFS: http.Dir(""),
	}
	entry.embedFS = Registry.EntryFS(entry.Kind(), entry.Name())
	if entry.embedFS != nil {
		entry.httpFS = http.FS(entry.embedFS)
	}

	switch f.Entry.SourceType {
	case "local":
		if !filepath.IsAbs(f.Entry.SourcePath) {
			wd, _ := os.Getwd()
			f.Entry.SourcePath = path.Join(wd, f.Entry.SourcePath)
		}
		entry.httpFS = http.Dir(f.Entry.SourcePath)
	}

	if len(f.Entry.Url) < 1 {
		f.Entry.Url = "/file"
	}

	f.Entry.Url = path.Join("/", f.Entry.Url, "/")

	Registry.AddEntry(entry)

	return entry, nil
}

type FileEntry struct {
	config  *FileConfig
	tpl     *template.Template
	embedFS *embed.FS
	httpFS  http.FileSystem
}

func (f *FileEntry) Category() string {
	return CategoryIndependent
}

func (f *FileEntry) Kind() string {
	return f.config.Kind
}

func (f *FileEntry) Name() string {
	return f.config.Metadata.Name
}

func (f *FileEntry) Config() EntryConfig {
	return f.config
}

func (f *FileEntry) Bootstrap(ctx context.Context) {
	if _, err := f.tpl.Parse(string(rku.ReadFileFromEmbed("assets/static/index.tmpl", &rkembed.AssetsFS, true))); err != nil {
		rku.ShutdownWithError(err)
	}
}

func (f *FileEntry) Interrupt(ctx context.Context) {}

func (f *FileEntry) Monitor() *Monitor {
	return nil
}

func (f *FileEntry) FS() *embed.FS {
	return f.embedFS
}

func (f *FileEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)

	res = append(res,
		&BuiltinApi{
			Method:  http.MethodGet,
			Path:    path.Join("/", f.config.Entry.Url),
			Handler: f.GetFileHandler(),
		})

	return res
}

// GetFileHandler handles requests sent from user.
func (f *FileEntry) GetFileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !strings.HasSuffix(request.URL.Path, "/") {
			request.URL.Path = request.URL.Path + "/"
		}

		// Trim prefix with path user defined in order to get file path
		p := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, f.config.Entry.Url), "/")

		if len(p) < 1 {
			p = "/"
		}
		p = path.Join("/", p)

		var file http.File
		var err error
		// open file
		if file, err = f.httpFS.Open(p); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			bytes, _ := json.Marshal(rkm.GetErrorBuilder().New(http.StatusInternalServerError, "Failed to open file", err))
			writer.Write(bytes)
			return
		}

		// get file info
		fileInfo, err := file.Stat()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			bytes, _ := json.Marshal(rkm.GetErrorBuilder().New(http.StatusInternalServerError, "Failed to stat file", err))
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
					Icon:     base64.StdEncoding.EncodeToString(rku.ReadFileFromEmbed(path.Join("assets/static/icons", f.getIconPath(v)), &rkembed.AssetsFS, false)),
					FileUrl:  path.Join(f.config.Entry.Url, p, v.Name()),
					FileName: v.Name(),
					Size:     v.Size(),
					ModTime:  v.ModTime(),
				})
			}

			f.sortFiles(files)

			data := &indexResp{
				PrevPath: path.Join(f.config.Entry.Url, path.Dir(p)),
				PrevIcon: base64.StdEncoding.EncodeToString(rku.ReadFileFromEmbed(path.Join("assets/static/icons/folder.png"), &rkembed.AssetsFS, false)),
				Path:     p,
				Files:    files,
			}

			buf := new(bytes.Buffer)
			if err := f.tpl.ExecuteTemplate(buf, "index", data); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				bytes, _ := json.Marshal(rkm.GetErrorBuilder().New(http.StatusInternalServerError, "Failed to execute go template", err))
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
func (f *FileEntry) sortFiles(res []*fileResp) {
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
func (f *FileEntry) getIconPath(info fs.FileInfo) string {
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
type indexResp struct {
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
