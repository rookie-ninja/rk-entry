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

func TestRegisterStaticFileHandlerEntry(t *testing.T) {
	assert.Empty(t, RegisterStaticFileHandlerEntry(&BootStaticFileHandler{}))

	// without options
	entry := RegisterStaticFileHandlerEntry(&BootStaticFileHandler{
		Enabled: true,
		Path:    "/static/",
	})
	assert.NotEmpty(t, entry)
	assert.Equal(t, "/static/", entry.Path)
	assert.NotNil(t, entry.httpFS)
	assert.NotNil(t, entry.Template)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.NotNil(t, entry.GetFileHandler())
}

func TestStaticFileHandlerEntry_Bootstrap_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	// without eventId in context
	entry := RegisterStaticFileHandlerEntry(&BootStaticFileHandler{
		Enabled: true,
		Path:    "/static/",
	})
	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestStaticFileHandlerEntry_EntryFunctions(t *testing.T) {
	entry := RegisterStaticFileHandlerEntry(&BootStaticFileHandler{
		Enabled: true,
		Path:    "/static/",
	})
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestStaticFileHandlerEntry_GetFileHandler(t *testing.T) {
	currDir := t.TempDir()
	os.MkdirAll(path.Join(currDir, "ut-dir"), os.ModePerm)
	os.WriteFile(path.Join(currDir, "ut-file"), []byte("ut content"), os.ModePerm)

	entry := RegisterStaticFileHandlerEntry(&BootStaticFileHandler{
		Enabled: true,
	})
	entry.SetHttpFS(http.Dir(currDir))
	entry.Bootstrap(context.TODO())
	handler := entry.GetFileHandler()

	// expect to get list of files
	writer := httptest.NewRecorder()
	req := &http.Request{
		URL: &url.URL{
			Path: "/static/",
		},
	}
	handler(writer, req)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Contains(t, writer.Body.String(), "Index of")

	// expect to get files to download
	writer = httptest.NewRecorder()
	req = &http.Request{
		URL: &url.URL{
			Path: "/static/ut-file",
		},
	}
	handler(writer, req)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.NotEmpty(t, writer.Header().Get("Content-Disposition"))
	assert.NotEmpty(t, writer.Header().Get("Content-Type"))
	assert.Contains(t, writer.Body.String(), "ut content")
}
