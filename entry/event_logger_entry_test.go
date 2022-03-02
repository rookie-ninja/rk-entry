// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEventEntryNoop(t *testing.T) {
	entry := NewEventEntryNoop()
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.EventFactory)
	assert.Nil(t, entry.LoggerConfig)
	assert.Nil(t, entry.LumberjackConfig)
	assert.NotNil(t, entry.eventHelper)
}

func TestNewEventEntryStdout(t *testing.T) {
	entry := NewEventEntryStdout()
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.EventFactory)
	assert.NotNil(t, entry.LoggerConfig)
	assert.Nil(t, entry.LumberjackConfig)
	assert.NotNil(t, entry.eventHelper)
}

func TestRegisterEventEntry(t *testing.T) {
	boot := &BootEvent{
		Event: []*BootEventE{
			{
				Name:        "ut-event",
				Encoding:    "console",
				OutputPaths: []string{"stdout"},
				Loki: BootLoki{
					Enabled: true,
					Path:    "mock",
					Addr:    "mock",
				},
			},
		},
	}

	entries := RegisterEventEntry(boot)

	assert.Len(t, entries, 1)
	assert.NotNil(t, entries[0].EventFactory)
	assert.NotNil(t, entries[0].eventHelper)
	assert.NotNil(t, entries[0].lokiSyncer)
	assert.NotNil(t, entries[0].baseLogger)
	assert.NotNil(t, entries[0].EventFactory)
	assert.NotNil(t, entries[0].LoggerConfig)
	assert.NotNil(t, entries[0].LumberjackConfig)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.Empty(t, entries[0].GetDescription())
	assert.NotEmpty(t, entries[0].String())
}

func TestEventEntry_UnmarshalJSON(t *testing.T) {
	assert.Nil(t, NewEventEntryNoop().UnmarshalJSON(nil))
}

func TestEventEntry_EventRelated(t *testing.T) {
	entry := NewEventEntryNoop()

	assert.NotNil(t, entry.CreateEvent("op"))
	entry.FinishEvent(entry.CreateEvent("op"))
	entry.FinishEventWithError(entry.CreateEvent("op"), nil)
	entry.FinishEventWithCond(entry.CreateEvent("op"), true)
}

func TestEventEntry_Bootstrap_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := NewEventEntryStdout()
	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestEventEntry_Syncer(t *testing.T) {
	defer assertNotPanic(t)

	boot := &BootEvent{
		Event: []*BootEventE{
			{
				Name:        "ut-event",
				Encoding:    "console",
				OutputPaths: []string{"stdout"},
				Loki: BootLoki{
					Enabled: true,
					Path:    "mock",
					Addr:    "mock",
				},
			},
		},
	}

	entries := RegisterEventEntry(boot)

	entries[0].AddEntryLabelToLokiSyncer(GlobalAppCtx.GetAppInfoEntry())
	entries[0].AddLabelToLokiSyncer("key", "value")
	entries[0].Sync()
}
