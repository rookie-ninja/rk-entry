// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"github.com/rookie-ninja/rk-logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLoggerEntryNoop(t *testing.T) {
	entry := NewLoggerEntryNoop()
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.Logger)
	assert.Nil(t, entry.LoggerConfig)
	assert.Nil(t, entry.LumberjackConfig)
}

func TestNewLoggerEntryStdout(t *testing.T) {
	entry := NewLoggerEntryStdout()
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.Logger)
	assert.NotNil(t, entry.LoggerConfig)
	assert.Nil(t, entry.LumberjackConfig)
}

func TestRegisterLoggerEntry(t *testing.T) {
	entries := RegisterLoggerEntry(&BootLogger{
		Logger: []*BootLoggerE{
			{
				Name: "ut-event",
				Zap: &rklogger.ZapConfigWrap{
					OutputPaths: []string{"stdout"},
					Encoding:    "console",
				},
				Loki: BootLoki{
					Enabled:            true,
					Path:               "mock",
					Addr:               "mock",
					Labels:             map[string]string{"A": "B"},
					InsecureSkipVerify: true,
				},
			},
		},
	})

	assert.Len(t, entries, 1)
	assert.NotNil(t, entries[0].Logger)
	assert.NotNil(t, entries[0].LoggerConfig)
	assert.NotNil(t, entries[0].LumberjackConfig)
	assert.NotNil(t, entries[0].lokiSyncer)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.Empty(t, entries[0].GetDescription())
	assert.NotEmpty(t, entries[0].String())
}

func TestLoggerEntry_UnmarshalJSON(t *testing.T) {
	assert.Nil(t, NewLoggerEntryNoop().UnmarshalJSON(nil))
}

func TestZapEntry_LogRelated(t *testing.T) {
	defer assertPanic(t)

	entry := NewLoggerEntryNoop()

	assert.NotNil(t, entry.WithOptions())
	assert.NotNil(t, entry.With())
	entry.Debug("msg")
	entry.Info("msg")
	entry.Warn("msg")
	entry.Error("msg")
	entry.DPanic("msg")
	entry.Panic("msg")
	entry.Fatal("msg")
}

func TestZapEntry_Bootstrap_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := NewLoggerEntryStdout()
	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestLoggerEntry_Syncer(t *testing.T) {
	defer assertNotPanic(t)

	entries := RegisterLoggerEntry(&BootLogger{
		Logger: []*BootLoggerE{
			{
				Name: "ut-event",
				Zap: &rklogger.ZapConfigWrap{
					OutputPaths: []string{"stdout"},
					Encoding:    "console",
				},
				Loki: BootLoki{
					Enabled: true,
					Path:    "mock",
					Addr:    "mock",
				},
			},
		},
	})

	entries[0].AddEntryLabelToLokiSyncer(GlobalAppCtx.GetAppInfoEntry())
	entries[0].AddLabelToLokiSyncer("key", "value")
	entries[0].Sync()
}
