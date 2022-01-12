package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTvEntry(t *testing.T) {
	entry := RegisterTvEntry(
		WithEventLoggerEntryTv(NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(NoopZapLoggerEntry()))

	assert.Equal(t, TvEntryNameDefault, entry.GetName())
	assert.Equal(t, TvEntryType, entry.GetType())
	assert.Equal(t, TvEntryDescription, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestTvEntry_Bootstrap(t *testing.T) {
	entry := RegisterTvEntry(
		WithEventLoggerEntryTv(NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(NoopZapLoggerEntry()))

	entry.Bootstrap(context.TODO())
}

func TestTvEntry_Interrupt(t *testing.T) {
	entry := RegisterTvEntry(
		WithEventLoggerEntryTv(NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(NoopZapLoggerEntry()))

	entry.Interrupt(context.TODO())
}

func TestRegisterTvEntryWithConfig(t *testing.T) {
	// with disabled
	config := &BootConfigTv{
		Enabled: false,
	}
	assert.Nil(t, RegisterTvEntryWithConfig(config, "", nil, nil))

	// with enabled
	config.Enabled = true
	assert.NotNil(t, RegisterTvEntryWithConfig(config, "", nil, nil))
}

func TestTvEntry_AssetsFileHandler(t *testing.T) {
	entry := RegisterTvEntry()

	handler := entry.AssetsFileHandler()

	// ok
	req := httptest.NewRequest(http.MethodGet, "/rk/v1/assets/tv/css/all.min.css", nil)
	writer := httptest.NewRecorder()
	handler(writer, req)
	assert.Equal(t, http.StatusOK, writer.Code)

	// not exist
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/assets/tv/css/not-exist", nil)
	writer = httptest.NewRecorder()
	handler(writer, req)
	assert.Equal(t, http.StatusInternalServerError, writer.Code)
}

func TestTvEntry_Action(t *testing.T) {
	entry := RegisterTvEntry()
	entry.Bootstrap(context.TODO())
	defer entry.Interrupt(context.TODO())
	logger := GlobalAppCtx.GetZapLoggerDefault()

	assert.NotNil(t, entry.Action("/", logger))
	assert.NotNil(t, entry.Action("/entries", logger))
	assert.NotNil(t, entry.Action("/configs", logger))
	assert.NotNil(t, entry.Action("/certs", logger))
	assert.NotNil(t, entry.Action("/os", logger))
	assert.NotNil(t, entry.Action("/env", logger))
	assert.NotNil(t, entry.Action("/prometheus", logger))
	assert.NotNil(t, entry.Action("/logs", logger))
	assert.NotNil(t, entry.Action("/deps", logger))
	assert.NotNil(t, entry.Action("/license", logger))
	assert.NotNil(t, entry.Action("/info", logger))
	assert.NotNil(t, entry.Action("/git", logger))
	assert.NotNil(t, entry.Action("/I-dont-know", logger))
}
