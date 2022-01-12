// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestWithNameCommonService_WithEmptyString(t *testing.T) {
	entry := RegisterCommonServiceEntry(
		WithNameCommonService(""))

	assert.NotEmpty(t, entry.GetName())
}

func TestWithNameCommonService_HappyCase(t *testing.T) {
	entry := RegisterCommonServiceEntry(
		WithNameCommonService("unit-test"))

	assert.Equal(t, "unit-test", entry.GetName())
}

func TestWithEventLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := RegisterCommonServiceEntry(
		WithEventLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.EventLoggerEntry)
}

func TestWithEventLoggerEntryCommonService_HappyCase(t *testing.T) {
	eventLoggerEntry := NoopEventLoggerEntry()
	entry := RegisterCommonServiceEntry(
		WithEventLoggerEntryCommonService(eventLoggerEntry))

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := RegisterCommonServiceEntry(
		WithZapLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.ZapLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_HappyCase(t *testing.T) {
	zapLoggerEntry := NoopZapLoggerEntry()
	entry := RegisterCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
}

func TestNewCommonServiceEntry_WithoutOptions(t *testing.T) {
	entry := RegisterCommonServiceEntry()

	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestNewCommonServiceEntry_HappyCase(t *testing.T) {
	zapLoggerEntry := NoopZapLoggerEntry()
	eventLoggerEntry := NoopEventLoggerEntry()

	entry := RegisterCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry),
		WithEventLoggerEntryCommonService(eventLoggerEntry),
		WithNameCommonService("ut"))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, "ut", entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestCommonServiceEntry_Bootstrap_WithoutRouter(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(
		WithZapLoggerEntryCommonService(NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())
}

func TestCommonServiceEntry_Bootstrap_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(
		WithZapLoggerEntryCommonService(NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())
}

func TestCommonServiceEntry_Interrupt_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(
		WithZapLoggerEntryCommonService(NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(NoopEventLoggerEntry()))
	entry.Interrupt(context.Background())
}

func TestCommonServiceEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterCommonServiceEntry(
		WithNameCommonService("ut"))

	assert.Equal(t, "ut", entry.GetName())
}

func TestCommonServiceEntry_GetType_HappyCase(t *testing.T) {
	entry := RegisterCommonServiceEntry()

	assert.Equal(t, "CommonServiceEntry", entry.GetType())
}

func TestCommonServiceEntry_String_HappyCase(t *testing.T) {
	entry := RegisterCommonServiceEntry()

	assert.NotEmpty(t, entry.String())
}

func TestCommonServiceEntry_Healthy_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Healthy(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.Equal(t, `{"healthy":true}`, writer.Body.String())
}

func TestCommonServiceEntry_GC_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Gc(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Info_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Info(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Config_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	vp := viper.New()
	vp.Set("unit-test-key", "unit-test-value")

	viperEntry := RegisterConfigEntry(
		WithNameConfig("unit-test"),
		WithViperInstanceConfig(vp))

	GlobalAppCtx.AddConfigEntry(viperEntry)

	entry.Configs(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
	assert.Contains(t, writer.Body.String(), "unit-test-key")
	assert.Contains(t, writer.Body.String(), "unit-test-value")
}

func TestCommonServiceEntry_Sys_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Sys(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Entries_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Entries(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Certs_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	RegisterCertEntry(WithNameCert("ut-cert"))
	certEntry := GlobalAppCtx.GetCertEntry("ut-cert")
	certEntry.Retriever = &CredRetrieverLocalFs{}
	certEntry.Store = &CertStore{}

	entry.Certs(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	GlobalAppCtx.RemoveCertEntry("ut-cert")
}

func TestCommonServiceEntry_Logs_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Logs(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Deps_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Deps(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_License_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.License(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Readme_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	entry.Readme(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Git_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry()

	writer := httptest.NewRecorder()

	GlobalAppCtx.SetRkMetaEntry(&RkMetaEntry{
		RkMeta: &rkcommon.RkMeta{
			Git: &rkcommon.Git{
				Commit: &rkcommon.Commit{
					Committer: &rkcommon.Committer{},
				},
			},
		},
	})

	entry.Git(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}
