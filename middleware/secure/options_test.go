// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package secure

import (
	"crypto/tls"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.EntryName())
	assert.Equal(t, "1; mode=block", set.xssProtection)
	assert.Equal(t, "nosniff", set.contentTypeNosniff)
	assert.Equal(t, "SAMEORIGIN", set.xFrameOptions)
	assert.False(t, set.hstsPreloadEnabled)
	assert.Empty(t, set.pathToIgnore)

	// with option
	set = NewOptionSet(
		WithEntryNameAndKind("ut-entry", "ut-type"),
		WithXSSProtection("ut-xss"),
		WithContentTypeNosniff("ut-sniff"),
		WithXFrameOptions("ut-frame"),
		WithHSTSMaxAge(10),
		WithHSTSExcludeSubdomains(true),
		WithHSTSPreloadEnabled(true),
		WithContentSecurityPolicy("ut-policy"),
		WithCSPReportOnly(true),
		WithReferrerPolicy("ut-ref"),
		WithPathToIgnore("ut-prefix"),
	).(*optionSet)

	assert.Equal(t, "ut-entry", set.EntryName())
	assert.Equal(t, "ut-type", set.EntryKind())
	assert.Equal(t, "ut-xss", set.xssProtection)
	assert.Equal(t, "ut-sniff", set.contentTypeNosniff)
	assert.Equal(t, "ut-frame", set.xFrameOptions)
	assert.Equal(t, 10, set.hstsMaxAge)
	assert.True(t, set.hstsExcludeSubdomains)
	assert.True(t, set.hstsPreloadEnabled)
	assert.Equal(t, "ut-policy", set.contentSecurityPolicy)
	assert.True(t, set.cspReportOnly)
	assert.Equal(t, "ut-ref", set.referrerPolicy)
	assert.NotEmpty(t, set.pathToIgnore)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil req
	set := NewOptionSet()
	assert.NotNil(t, set.BeforeCtx(nil))

	// with req
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	req.TLS = &tls.ConnectionState{}
	req.Header.Set(rkm.HeaderXForwardedProto, "https")

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
		rkm.HeaderXXSSProtection,
		rkm.HeaderXContentTypeOptions,
		rkm.HeaderXFrameOptions)

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
		WithPathToIgnore("ut-prefix"))

	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	req.TLS = &tls.ConnectionState{}
	ctx = set.BeforeCtx(req)
	set.Before(ctx)

	containsHeader(t, ctx.Output.HeadersToReturn,
		rkm.HeaderXXSSProtection,
		rkm.HeaderXContentTypeOptions,
		rkm.HeaderXFrameOptions,
		rkm.HeaderStrictTransportSecurity,
		rkm.HeaderContentSecurityPolicyReportOnly,
		rkm.HeaderReferrerPolicy)
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
	assert.NotEmpty(t, mock.EntryName())
	assert.NotEmpty(t, mock.EntryKind())
	assert.NotNil(t, mock.BeforeCtx(nil))
	mock.Before(nil)
}

func containsHeader(t *testing.T, in map[string]string, headers ...string) {
	for _, v := range headers {
		assert.Contains(t, in, v)
	}
}
