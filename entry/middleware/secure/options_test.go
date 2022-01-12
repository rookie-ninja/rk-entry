// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidsec

import (
	"crypto/tls"
	rkmid "github.com/rookie-ninja/rk-entry/entry/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.Equal(t, "1; mode=block", set.xssProtection)
	assert.Equal(t, "nosniff", set.contentTypeNosniff)
	assert.Equal(t, "SAMEORIGIN", set.xFrameOptions)
	assert.False(t, set.hstsPreloadEnabled)
	assert.Empty(t, set.ignorePrefix)

	// with option
	set = NewOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithXSSProtection("ut-xss"),
		WithContentTypeNosniff("ut-sniff"),
		WithXFrameOptions("ut-frame"),
		WithHSTSMaxAge(10),
		WithHSTSExcludeSubdomains(true),
		WithHSTSPreloadEnabled(true),
		WithContentSecurityPolicy("ut-policy"),
		WithCSPReportOnly(true),
		WithReferrerPolicy("ut-ref"),
		WithIgnorePrefix("ut-prefix"),
	).(*optionSet)

	assert.Equal(t, "ut-entry", set.GetEntryName())
	assert.Equal(t, "ut-type", set.GetEntryType())
	assert.Equal(t, "ut-xss", set.xssProtection)
	assert.Equal(t, "ut-sniff", set.contentTypeNosniff)
	assert.Equal(t, "ut-frame", set.xFrameOptions)
	assert.Equal(t, 10, set.hstsMaxAge)
	assert.True(t, set.hstsExcludeSubdomains)
	assert.True(t, set.hstsPreloadEnabled)
	assert.Equal(t, "ut-policy", set.contentSecurityPolicy)
	assert.True(t, set.cspReportOnly)
	assert.Equal(t, "ut-ref", set.referrerPolicy)
	assert.NotEmpty(t, set.ignorePrefix)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil req
	set := NewOptionSet()
	assert.NotNil(t, set.BeforeCtx(nil))

	// with req
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	req.TLS = &tls.ConnectionState{}
	req.Header.Set(rkmid.HeaderXForwardedProto, "https")

	ctx := set.BeforeCtx(req)
	assert.Equal(t, "/ut", ctx.Input.UrlPath)
	assert.True(t, ctx.Input.isTLS)
	assert.Equal(t, "https", ctx.Input.xForwardedProto)
}

func TestOptionSet_Before(t *testing.T) {
	// without options
	set := NewOptionSet()
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx := set.BeforeCtx(req)
	set.Before(ctx)
	containsHeader(t, ctx.Output.HeadersToReturn,
		rkmid.HeaderXXSSProtection,
		rkmid.HeaderXContentTypeOptions,
		rkmid.HeaderXFrameOptions)

	// with options
	set = NewOptionSet(
		WithXSSProtection("ut-xss"),
		WithContentTypeNosniff("ut-sniff"),
		WithXFrameOptions("ut-frame"),
		WithHSTSMaxAge(10),
		WithHSTSExcludeSubdomains(true),
		WithHSTSPreloadEnabled(true),
		WithContentSecurityPolicy("ut-policy"),
		WithCSPReportOnly(true),
		WithReferrerPolicy("ut-ref"),
		WithIgnorePrefix("ut-prefix"))

	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	req.TLS = &tls.ConnectionState{}
	ctx = set.BeforeCtx(req)
	set.Before(ctx)

	containsHeader(t, ctx.Output.HeadersToReturn,
		rkmid.HeaderXXSSProtection,
		rkmid.HeaderXContentTypeOptions,
		rkmid.HeaderXFrameOptions,
		rkmid.HeaderStrictTransportSecurity,
		rkmid.HeaderContentSecurityPolicyReportOnly,
		rkmid.HeaderReferrerPolicy)
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

func containsHeader(t *testing.T, in map[string]string, headers ...string) {
	for _, v := range headers {
		assert.Contains(t, in, v)
	}
}
