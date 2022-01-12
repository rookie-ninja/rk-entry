// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidauth

import (
	"encoding/base64"
	"fmt"
	rkmid "github.com/rookie-ninja/rk-entry/entry/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOptionSet_BeforeCtx(t *testing.T) {
	set := NewOptionSet()

	// without request
	ctx := set.BeforeCtx(nil)
	assert.NotNil(t, ctx)

	// with http.Request
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, "basic")
	req.Header.Set(rkmid.HeaderApiKey, "apiKey")
	ctx = set.BeforeCtx(req)

	assert.Equal(t, "basic", ctx.Input.BasicAuthHeader)
	assert.Equal(t, "apiKey", ctx.Input.ApiKeyHeader)
	assert.Empty(t, ctx.Output.HeadersToReturn)
	assert.Nil(t, ctx.Output.ErrResp)
}

func TestToOptions(t *testing.T) {
	config := &BootConfig{
		Enabled:      false,
		IgnorePrefix: []string{},
		Basic:        []string{},
		ApiKey:       []string{},
	}

	// with disabled
	assert.Empty(t, ToOptions(config, "", ""))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", ""))
}

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)

	assert.NotEmpty(t, set.GetEntryName())
	assert.NotNil(t, set.basicAccounts)
	assert.NotNil(t, set.apiKey)
	assert.Empty(t, set.ignorePrefix)

	// with options
	set = NewOptionSet(
		WithEntryNameAndType("ut-name", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-key"),
		WithIgnorePrefix("ut-ignore")).(*optionSet)

	assert.NotEmpty(t, set.GetEntryName())
	assert.NotEmpty(t, set.GetEntryType())
	assert.NotEmpty(t, set.basicRealm)
	assert.NotEmpty(t, set.basicAccounts)
	assert.NotEmpty(t, set.apiKey)
	assert.NotEmpty(t, set.ignorePrefix)
}

func TestOptionSet_isBasicAuthorized(t *testing.T) {
	// case 1: auth header is provided
	// case 1.1: invalid basic auth
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, "invalid")

	set := NewOptionSet().(*optionSet)
	resp := set.isBasicAuthorized(set.BeforeCtx(req))
	assert.Equal(t, http.StatusUnauthorized, resp.Err.Code)

	// case 1.2: not authorized
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, "Basic invalid")

	set = NewOptionSet().(*optionSet)
	ctx := set.BeforeCtx(req)
	resp = set.isBasicAuthorized(ctx)
	assert.NotEmpty(t, ctx.Output.HeadersToReturn)
	assert.Equal(t, http.StatusUnauthorized, resp.Err.Code)

	// case 1.3: authorized
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("user:pass"))))

	set = NewOptionSet(WithBasicAuth("", "user:pass")).(*optionSet)
	ctx = set.BeforeCtx(req)
	resp = set.isBasicAuthorized(ctx)
	assert.Nil(t, resp)

	// case 2: auth header missing
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)

	set = NewOptionSet().(*optionSet)
	ctx = set.BeforeCtx(req)
	resp = set.isBasicAuthorized(ctx)
	assert.Equal(t, http.StatusUnauthorized, resp.Err.Code)
}

func TestOptionSet_isApiKeyAuthorized(t *testing.T) {
	// case 1: auth header is provided
	// case 1.1: not authorized
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderApiKey, "invalid")

	set := NewOptionSet().(*optionSet)
	ctx := set.BeforeCtx(req)
	resp := set.isApiKeyAuthorized(ctx)
	assert.Equal(t, http.StatusUnauthorized, resp.Err.Code)

	// case 1.2: authorized
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderApiKey, "key")

	set = NewOptionSet(WithApiKeyAuth("key")).(*optionSet)
	ctx = set.BeforeCtx(req)
	resp = set.isApiKeyAuthorized(ctx)
	assert.Nil(t, resp)

	// case 2: auth header missing
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)

	set = NewOptionSet().(*optionSet)
	ctx = set.BeforeCtx(req)
	resp = set.isApiKeyAuthorized(ctx)
	assert.Equal(t, http.StatusUnauthorized, resp.Err.Code)
}

func TestOptionSet_Before(t *testing.T) {
	// case 0: ignore path
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)

	set := NewOptionSet(WithIgnorePrefix("/ut-path"))
	ctx := set.BeforeCtx(req)
	set.Before(ctx)
	assert.Nil(t, ctx.Output.ErrResp)

	// case 1: basic auth passed
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("user:pass"))))

	set = NewOptionSet(WithBasicAuth("", "user:pass"))
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.Nil(t, ctx.Output.ErrResp)

	// case 2: X-API-Key passed
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderApiKey, "key")

	set = NewOptionSet(WithApiKeyAuth("key"))
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.Nil(t, ctx.Output.ErrResp)

	// case 3: basic auth provided, then return code and response related to basic auth
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, "Basic invalid")

	set = NewOptionSet(WithBasicAuth("", "user:pass"))
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.ErrResp)
	assert.Equal(t, http.StatusUnauthorized, ctx.Output.ErrResp.Err.Code)
	assert.NotEmpty(t, ctx.Output.HeadersToReturn)

	// case 4: X-API-Key provided, then return code and response related to X-API-Key
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderApiKey, "invalid")

	set = NewOptionSet(WithApiKeyAuth("key"))
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.ErrResp)
	assert.Equal(t, http.StatusUnauthorized, ctx.Output.ErrResp.Err.Code)

	// case 5: no auth provided, return bellow code and response
	// case 5.1: basic auth needed
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)

	set = NewOptionSet(WithBasicAuth("", "user:pass"))
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.ErrResp)
	assert.Equal(t, http.StatusUnauthorized, ctx.Output.ErrResp.Err.Code)
	assert.NotEmpty(t, ctx.Output.HeadersToReturn)

	// case 5.2: X-API-Key needed
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)

	set = NewOptionSet(WithApiKeyAuth("key"))
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.ErrResp)
	assert.Equal(t, http.StatusUnauthorized, ctx.Output.ErrResp.Err.Code)
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
	assert.NotNil(t, mock.BeforeCtx(nil))
	mock.Before(nil)
}
