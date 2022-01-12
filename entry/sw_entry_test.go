// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithPath_HappyCase(t *testing.T) {
	entry := RegisterSwEntry(WithPathSw("ut-path"))
	assert.Equal(t, "/ut-path/", entry.Path)
}

func TestWithHeaders_HappyCase(t *testing.T) {
	headers := map[string]string{
		"key": "value",
	}
	entry := RegisterSwEntry(WithHeadersSw(headers))
	assert.Len(t, entry.Headers, 1)
}

func TestNewSwEntry(t *testing.T) {
	entry := RegisterSwEntry(
		WithPortSw(1234),
		WithNameSw("ut-name"),
		WithPathSw("ut-path"),
		WithJsonPathSw("ut-json-path"),
		WithHeadersSw(map[string]string{
			"key": "value",
		}),
		WithZapLoggerEntrySw(NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(NoopEventLoggerEntry()),
		WithEnableCommonServiceSw(true))

	assert.Equal(t, uint64(1234), entry.Port)
	assert.Equal(t, "/ut-path/", entry.Path)
	assert.Equal(t, "ut-json-path", entry.JsonPath)
	assert.Len(t, entry.Headers, 1)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.True(t, entry.EnableCommonService)
}

func TestSwEntry_Bootstrap(t *testing.T) {
	//defer assertNotPanic(t)

	entry := RegisterSwEntry(
		WithPortSw(1234),
		WithNameSw("ut-name"),
		WithPathSw("ut-path"),
		WithJsonPathSw("ut-json-path"),
		WithHeadersSw(map[string]string{
			"key": "value",
		}),
		WithZapLoggerEntrySw(NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(NoopEventLoggerEntry()),
		WithEnableCommonServiceSw(true))

	entry.Bootstrap(context.TODO())
}

func TestSwEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterSwEntry(
		WithPortSw(1234),
		WithNameSw("ut-name"),
		WithPathSw("ut-path"),
		WithJsonPathSw("ut-json-path"),
		WithHeadersSw(map[string]string{
			"key": "value",
		}),
		WithZapLoggerEntrySw(NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(NoopEventLoggerEntry()),
		WithEnableCommonServiceSw(true))

	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestSwEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterSwEntry()
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestSwEntry(t *testing.T) {
	entry := RegisterSwEntry()

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
}

func TestSwEntry_AssetsFileHandler(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterSwEntry()

	writer := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rk/v1/assets", nil)

	entry.AssetsFileHandler()(writer, req)
}

func TestSwEntry_ConfigFileHandler(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterSwEntry()

	writer := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rk/v1/assets", nil)

	entry.ConfigFileHandler()(writer, req)
}
func TestRegisterSwEntryWithConfig(t *testing.T) {
	// with disabled
	config := &BootConfigSw{
		Enabled: false,
	}
	assert.Nil(t, RegisterSwEntryWithConfig(config, "", 8080, nil, nil, false))

	// enabled
	config.Enabled = true
	config.Headers = []string{"a:b"}
	assert.NotNil(t, RegisterSwEntryWithConfig(config, "", 8080, nil, nil, false))
}
