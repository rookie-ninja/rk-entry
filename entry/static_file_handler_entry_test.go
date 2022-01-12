package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"
)

func TestNewStaticFileHandlerEntry(t *testing.T) {
	// without options
	entry := RegisterStaticFileHandlerEntry()
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.Equal(t, "/rk/v1/static/", entry.Path)
	assert.NotNil(t, entry.Fs)
	assert.NotNil(t, entry.Template)

	// with options
	utFs := http.Dir("")
	utPath := "/ut-path/"
	utZapLogger := NoopZapLoggerEntry()
	utEventLogger := NoopEventLoggerEntry()
	utName := "ut-entry"

	entry = RegisterStaticFileHandlerEntry(
		WithPathStatic(utPath),
		WithEventLoggerEntryStatic(utEventLogger),
		WithZapLoggerEntryStatic(utZapLogger),
		WithNameStatic(utName),
		WithFileSystemStatic(utFs))

	assert.NotNil(t, entry)
	assert.Equal(t, utZapLogger, entry.ZapLoggerEntry)
	assert.Equal(t, utEventLogger, entry.EventLoggerEntry)
	assert.Equal(t, utPath, entry.Path)
	assert.Equal(t, utFs, entry.Fs)
	assert.NotNil(t, entry.Template)
	assert.Equal(t, utName, entry.EntryName)
}

func TestStaticFileHandlerEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without eventId in context
	entry := RegisterStaticFileHandlerEntry()
	entry.Bootstrap(context.TODO())

	// with eventId in context
	entry.Bootstrap(context.TODO())
}

func TestStaticFileHandlerEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	// without eventId in context
	entry := RegisterStaticFileHandlerEntry()
	entry.Interrupt(context.TODO())

	// with eventId in context
	entry.Interrupt(context.TODO())
}

func TestStaticFileHandlerEntry_EntryFunctions(t *testing.T) {
	entry := RegisterStaticFileHandlerEntry()
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestStaticFileHandlerEntry_GetFileHandler(t *testing.T) {
	currDir := t.TempDir()
	os.MkdirAll(path.Join(currDir, "ut-dir"), os.ModePerm)
	os.WriteFile(path.Join(currDir, "ut-file"), []byte("ut content"), os.ModePerm)

	entry := RegisterStaticFileHandlerEntry(
		WithFileSystemStatic(http.Dir(currDir)))
	entry.Bootstrap(context.TODO())
	handler := entry.GetFileHandler()

	// expect to get list of files
	writer := httptest.NewRecorder()
	req := &http.Request{
		URL: &url.URL{
			Path: "/rk/v1/static/",
		},
	}
	handler(writer, req)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Contains(t, writer.Body.String(), "Index of")

	// expect to get files to download
	writer = httptest.NewRecorder()
	req = &http.Request{
		URL: &url.URL{
			Path: "/rk/v1/static/ut-file",
		},
	}
	handler(writer, req)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.NotEmpty(t, writer.Header().Get("Content-Disposition"))
	assert.NotEmpty(t, writer.Header().Get("Content-Type"))
	assert.Contains(t, writer.Body.String(), "ut content")
}

func TestRegisterStaticFileHandlerEntryWithConfig(t *testing.T) {
	// with disabled
	config := &BootConfigStaticHandler{
		Enabled: false,
	}
	assert.Nil(t, RegisterStaticFileHandlerEntryWithConfig(config, "", nil, nil))

	// enabled
	config.Enabled = true
	config.SourceType = "local"
	config.SourcePath = t.TempDir()
	assert.NotNil(t, RegisterStaticFileHandlerEntryWithConfig(config, "", nil, nil))
}
