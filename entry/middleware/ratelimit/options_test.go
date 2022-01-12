// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidlimit

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.Equal(t, DefaultLimit, set.reqPerSec)
	assert.Empty(t, set.reqPerSecByPath)
	assert.Equal(t, TokenBucket, set.algorithm)
	assert.NotEmpty(t, set.limiter)

	// with option
	l := func() error { return nil }
	set = NewOptionSet(
		WithEntryNameAndType("name", "type"),
		WithReqPerSec(1),
		WithReqPerSecByPath("/ut", 1),
		WithAlgorithm(LeakyBucket),
		WithGlobalLimiter(l),
		WithLimiterByPath("/ut-sub", l),
	).(*optionSet)

	assert.NotEmpty(t, set.GetEntryName())
	assert.Equal(t, 1, set.reqPerSec)
	assert.NotEmpty(t, set.reqPerSecByPath)
	assert.Equal(t, LeakyBucket, set.algorithm)
	assert.NotEmpty(t, t, set.limiter)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil req
	set := NewOptionSet().(*optionSet)
	ctx := set.BeforeCtx(nil)
	assert.Empty(t, ctx.Input.UrlPath)

	// with req
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req)
	assert.NotEmpty(t, ctx.Input.UrlPath)
}

func TestOptionSet_Before(t *testing.T) {
	// with nil ctx
	set := NewOptionSet().(*optionSet)
	set.Before(nil)

	// with error
	l := func() error {
		return errors.New("ut-error")
	}
	set = NewOptionSet(WithGlobalLimiter(l)).(*optionSet)
	beforeCtx := NewBeforeCtx()
	beforeCtx.Input.UrlPath = "/ut"
	set.Before(beforeCtx)

	assert.NotNil(t, beforeCtx.Output.ErrResp)

	// happy case
	l = func() error {
		return nil
	}
	set = NewOptionSet(WithGlobalLimiter(l)).(*optionSet)
	beforeCtx = NewBeforeCtx()
	beforeCtx.Input.UrlPath = "/ut"
	set.Before(beforeCtx)

	assert.Nil(t, beforeCtx.Output.ErrResp)
}

func TestToOptions(t *testing.T) {
	// with disabled
	config := &BootConfig{
		Enabled: false,
	}

	assert.Empty(t, ToOptions(config, "", ""))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", ""))
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
	assert.NotNil(t, mock.BeforeCtx(nil))
	mock.Before(nil)
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
