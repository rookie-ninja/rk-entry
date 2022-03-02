// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestRegisterCommonServiceEntry(t *testing.T) {
	entry := RegisterCommonServiceEntry(&BootCommonService{
		Enabled:    true,
		PathPrefix: "ut-prefix",
	})

	assert.Contains(t, entry.HealthyPath, "/ut-prefix")
	assert.Contains(t, entry.GcPath, "/ut-prefix")
	assert.Contains(t, entry.InfoPath, "/ut-prefix")

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
}

func TestCommonServiceEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(&BootCommonService{
		Enabled: true,
	})
	entry.Bootstrap(context.Background())
}

func TestCommonServiceEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(&BootCommonService{
		Enabled: true,
	})
	entry.Bootstrap(context.Background())
	entry.Interrupt(context.Background())
}

func TestCommonServiceEntry_Healthy(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(&BootCommonService{
		Enabled: true,
	})

	writer := httptest.NewRecorder()

	entry.Healthy(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.Contains(t, writer.Body.String(), "true")
}

func TestCommonServiceEntry_GC(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(&BootCommonService{
		Enabled: true,
	})

	writer := httptest.NewRecorder()

	entry.Gc(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_Info(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterCommonServiceEntry(&BootCommonService{
		Enabled: true,
	})

	writer := httptest.NewRecorder()

	entry.Info(writer, nil)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())
}

func TestCommonServiceEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterCommonServiceEntry(&BootCommonService{
		Enabled: true,
	})
	assert.Nil(t, entry.UnmarshalJSON(nil))
}
