// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLabelerHttp_Keys(t *testing.T) {
	l := &labelerHttp{}

	assert.Equal(t, labelKeysHttp, l.Keys())
	assert.Equal(t, len(labelKeysHttp), len(l.Values()))
}

func TestLabelerHttp_Values(t *testing.T) {
	l := &labelerHttp{
		entryName: "entryName",
		entryKind: "entryKind",
		domain:    "domain",
		instance:  "instance",
		method:    "GET",
		path:      "/",
		resCode:   "200",
	}

	// key -> value should match
	keys := l.Keys()
	values := l.Values()

	assert.Equal(t, "entryName", keys[0])
	assert.Equal(t, "entryName", values[0])

	assert.Equal(t, "entryType", keys[1])
	assert.Equal(t, "entryType", values[1])

	assert.Equal(t, "domain", keys[2])
	assert.Equal(t, "domain", values[2])

	assert.Equal(t, "instance", keys[3])
	assert.Equal(t, "instance", values[3])

	assert.Equal(t, "restMethod", keys[4])
	assert.Equal(t, "GET", values[4])

	assert.Equal(t, "restPath", keys[5])
	assert.Equal(t, "/", values[5])

	assert.Equal(t, "resCode", keys[6])
	assert.Equal(t, "200", values[6])
}

func TestLabelerGrpc_Keys(t *testing.T) {
	l := &labelerGrpc{}

	assert.Equal(t, labelKeysGrpc, l.Keys())
	assert.Equal(t, len(labelKeysGrpc), len(l.Values()))
}

func TestLabelerGrpc_Values(t *testing.T) {
	l := &labelerGrpc{
		entryName:   "entryName",
		entryKind:   "entryKind",
		domain:      "domain",
		instance:    "instance",
		restMethod:  "GET",
		restPath:    "/",
		grpcMethod:  "Hello",
		grpcType:    "Unary",
		grpcService: "Common",
		resCode:     "200",
	}

	// key -> value should match
	keys := l.Keys()
	values := l.Values()

	assert.Equal(t, "entryName", keys[0])
	assert.Equal(t, "entryName", values[0])

	assert.Equal(t, "entryType", keys[1])
	assert.Equal(t, "entryType", values[1])

	assert.Equal(t, "domain", keys[2])
	assert.Equal(t, "domain", values[2])

	assert.Equal(t, "instance", keys[3])
	assert.Equal(t, "instance", values[3])

	assert.Equal(t, "grpcService", keys[4])
	assert.Equal(t, "Common", values[4])

	assert.Equal(t, "grpcMethod", keys[5])
	assert.Equal(t, "Hello", values[5])

	assert.Equal(t, "grpcType", keys[6])
	assert.Equal(t, "Unary", values[6])

	assert.Equal(t, "restMethod", keys[7])
	assert.Equal(t, "GET", values[7])

	assert.Equal(t, "restPath", keys[8])
	assert.Equal(t, "/", values[8])

	assert.Equal(t, "resCode", keys[9])
	assert.Equal(t, "200", values[9])
}

func TestToOptions(t *testing.T) {
	config := &BootConfig{
		Enabled: false,
		Ignore:  []string{},
	}

	// with disabled
	assert.Empty(t, ToOptions(config, "", "", nil, ""))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", "", nil, ""))
}

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.EntryName())
	assert.Empty(t, set.pathToIgnore)
	assert.Equal(t, LabelerTypeHttp, set.labelerType)
	assert.NotNil(t, set.registerer)
	assert.NotNil(t, set.metricsSet)
	assert.NotNil(t, set.metricsSet.GetSummary(MetricsNameElapsedNano))
	assert.NotNil(t, set.metricsSet.GetCounter(MetricsNameResCode))
	assert.NotNil(t, GetServerMetricsSet(set.EntryName()))

	ClearAllMetrics()

	// with options
	reg := prometheus.NewRegistry()

	set = NewOptionSet(
		WithEntryNameAndKind("name", "kind"),
		WithRegisterer(reg),
		WithPathToIgnore("/ut-ignore"),
		WithLabelerType(LabelerTypeGrpc)).(*optionSet)

	assert.NotEmpty(t, set.EntryName())
	assert.NotEmpty(t, set.EntryKind())
	assert.Contains(t, set.pathToIgnore, "/ut-ignore")
	assert.Equal(t, LabelerTypeGrpc, set.labelerType)
	assert.Equal(t, reg, set.registerer)
	assert.NotNil(t, set.metricsSet)
	assert.NotNil(t, set.metricsSet.GetSummary(MetricsNameElapsedNano))
	assert.NotNil(t, set.metricsSet.GetCounter(MetricsNameResCode))

	ClearAllMetrics()
}

func TestOptionSet_ignore(t *testing.T) {
	set := NewOptionSet(WithPathToIgnore("/ut-ignore")).(*optionSet)
	assert.True(t, set.ShouldIgnore("/ut-ignore"))
	assert.False(t, set.ShouldIgnore("/"))

	ClearAllMetrics()
}

func TestOptionSet_getServerResCodeMetrics(t *testing.T) {
	set := NewOptionSet().(*optionSet)
	assert.NotNil(t, set.getServerResCodeMetrics(&labelerHttp{}))

	ClearAllMetrics()
}

func TestOptionSet_getServerDurationMetrics(t *testing.T) {
	set := NewOptionSet().(*optionSet)
	assert.NotNil(t, set.getServerDurationMetrics(&labelerHttp{}))

	ClearAllMetrics()
}

func TestOptionSet_AfterCtx(t *testing.T) {
	set := NewOptionSet()

	ctx := set.AfterCtx("200")
	assert.NotNil(t, ctx)
	assert.Equal(t, "200", ctx.Input.ResCode)

	ClearAllMetrics()
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	set := NewOptionSet()

	// with nil req
	assert.NotNil(t, set.BeforeCtx(nil))

	// happy case
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx := set.BeforeCtx(req)
	assert.Equal(t, http.MethodGet, ctx.Input.RestMethod)
	assert.Equal(t, "/ut", ctx.Input.RestPath)

	ClearAllMetrics()
}

func TestOptionSet_Before(t *testing.T) {
	defer assertNotPanic(t)

	set := NewOptionSet()
	set.Before(set.BeforeCtx(nil))
	ClearAllMetrics()
}

func TestOptionSet_After(t *testing.T) {
	defer assertNotPanic(t)

	set := NewOptionSet()

	// with nil input
	set.After(nil, nil)

	// happy case with http label type
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	beforeCtx := set.BeforeCtx(req)

	set.Before(beforeCtx)

	afterCtx := set.AfterCtx("200")
	set.After(beforeCtx, afterCtx)

	ClearAllMetrics()

	// happy case with grpc label type
	set = NewOptionSet(WithLabelerType(LabelerTypeGrpc))
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	beforeCtx = set.BeforeCtx(req)

	set.Before(beforeCtx)

	afterCtx = set.AfterCtx("200")
	set.After(beforeCtx, afterCtx)

	ClearAllMetrics()

	// happy case with default
	set = NewOptionSet(WithLabelerType("unknown"))
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	beforeCtx = set.BeforeCtx(req)

	set.Before(beforeCtx)

	afterCtx = set.AfterCtx("200")
	set.After(beforeCtx, afterCtx)

	ClearAllMetrics()
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx(), NewAfterCtx())
	assert.NotEmpty(t, mock.EntryName())
	assert.NotEmpty(t, mock.EntryKind())
	assert.NotNil(t, mock.BeforeCtx(nil))
	assert.NotNil(t, mock.AfterCtx(""))
	mock.Before(nil)
	mock.After(nil, nil)
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}
