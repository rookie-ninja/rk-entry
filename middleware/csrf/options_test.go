// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package csrf

import (
	"context"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.EntryName())
	assert.Equal(t, 32, set.tokenLength)
	assert.Equal(t, "header:"+rkm.HeaderXCSRFToken, set.tokenLookup)
	assert.Equal(t, "_csrf", set.cookieName)
	assert.Equal(t, 86400, set.cookieMaxAge)
	assert.Equal(t, http.SameSiteDefaultMode, set.cookieSameSite)
	assert.Empty(t, set.pathToIgnore)
	assert.NotNil(t, set.extractor)

	// with option
	set = NewOptionSet(
		WithEntryNameAndKind("ut-entry", "ut-kind"),
		WithExtractor(func(context.Context) (string, error) {
			return "", nil
		}),
		WithTokenLength(10),
		WithTokenLookup("header:ut-header"),
		WithCookieName("ut-cookie"),
		WithCookieDomain("ut-domain"),
		WithCookiePath("ut-path"),
		WithCookieMaxAge(10),
		WithCookieHTTPOnly(true),
		WithCookieSameSite(http.SameSiteDefaultMode),
	).(*optionSet)

	assert.Equal(t, "ut-entry", set.EntryName())
	assert.Equal(t, "ut-type", set.EntryKind())
	assert.NotNil(t, set.userExtractor)
	assert.Equal(t, 10, set.tokenLength)
	assert.Equal(t, "header:ut-header", set.tokenLookup)
	assert.Equal(t, "ut-cookie", set.cookieName)
	assert.Equal(t, "ut-domain", set.cookieDomain)
	assert.Equal(t, "ut-path", set.cookiePath)
	assert.True(t, set.cookieHTTPOnly)
	assert.Equal(t, 10, set.cookieMaxAge)
	assert.Equal(t, http.SameSiteDefaultMode, set.cookieSameSite)
	assert.Empty(t, set.pathToIgnore)
	assert.NotNil(t, set.extractor)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil req
	set := NewOptionSet()
	assert.NotNil(t, set.BeforeCtx(nil))

	// with cookie
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	req.AddCookie(&http.Cookie{
		Value: "value",
	})
	ctx := set.BeforeCtx(req)

	assert.NotNil(t, ctx.Input.Request)
	assert.NotEmpty(t, ctx.Input.Token)
	assert.Equal(t, http.MethodGet, ctx.Input.Method)
	assert.Equal(t, "/ut", ctx.Input.UrlPath)

	// without cookie
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req)
	assert.NotNil(t, ctx.Input.Request)
	assert.NotEmpty(t, ctx.Input.Token)
	assert.Equal(t, http.MethodGet, ctx.Input.Method)
	assert.Equal(t, "/ut", ctx.Input.UrlPath)
}

func TestOptionSet_Before(t *testing.T) {
	set := NewOptionSet()

	// match 3.1
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx := set.BeforeCtx(req)
	set.Before(ctx)
	assert.Nil(t, ctx.Output.ErrResp)
	assert.NotNil(t, ctx.Output.Cookie)

	// match 3.2
	req = httptest.NewRequest(http.MethodPost, "/ut", nil)
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.ErrResp)
	assert.Contains(t, ctx.Output.ErrResp.Error(), http.StatusText(http.StatusBadRequest))
	assert.Nil(t, ctx.Output.Cookie)

	// match 3.3
	req = httptest.NewRequest(http.MethodPost, "/ut", nil)
	req.Header.Set(rkm.HeaderXCSRFToken, "ut-csrf-token")
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.Contains(t, ctx.Output.ErrResp.Error(), http.StatusText(http.StatusForbidden))
	assert.Nil(t, ctx.Output.Cookie)

	// match 4.1
	set = NewOptionSet(WithCookiePath("/ut"))
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.Equal(t, "/ut", ctx.Output.Cookie.Path)

	// match 4.2
	set = NewOptionSet(WithCookieDomain("ut-domain"))
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.Equal(t, "ut-domain", ctx.Output.Cookie.Domain)

	// match 4.3
	set = NewOptionSet(WithCookieSameSite(http.SameSiteStrictMode))
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.Equal(t, http.SameSiteStrictMode, ctx.Output.Cookie.SameSite)
}

func TestOptionSet_IsValidToken(t *testing.T) {
	set := NewOptionSet().(*optionSet)

	// expect ture
	token := "my-token"
	clientToken := "my-token"

	assert.True(t, set.isValidToken(token, clientToken))

	// expect false
	assert.False(t, set.isValidToken(token, clientToken+"-invalid"))
}

func TestCsrfTokenFromHeader(t *testing.T) {
	set := NewOptionSet(WithTokenLookup("header:ut-header")).(*optionSet)

	// happy case
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("ut-header", "ut-header-value")
	res, err := set.extractor(req)
	assert.Nil(t, err)
	assert.Equal(t, "ut-header-value", res)

	// expect error
	req = &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("ut-header-invalid", "ut-header-value")
	res, err = set.extractor(req)
	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestCsrfTokenFromForm(t *testing.T) {
	set := NewOptionSet(WithTokenLookup("form:ut-form")).(*optionSet)

	// happy case
	req := &http.Request{
		Form: url.Values{},
	}
	req.Form.Set("ut-form", "ut-form-value")
	res, err := set.extractor(req)
	assert.Nil(t, err)
	assert.Equal(t, "ut-form-value", res)

	// expect error
	req = &http.Request{
		Form: url.Values{},
	}
	req.Form.Set("ut-form-invalid", "ut-form-value")
	res, err = set.extractor(req)
	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestCsrfTokenFromQuery(t *testing.T) {
	set := NewOptionSet(WithTokenLookup("query:ut-query")).(*optionSet)

	// happy case
	req := &http.Request{
		URL: &url.URL{},
	}
	req.URL.RawQuery = "ut-query=ut-query-value"
	res, err := set.extractor(req)
	assert.Nil(t, err)
	assert.Equal(t, "ut-query-value", res)

	// expect error
	req = &http.Request{
		URL: &url.URL{},
	}
	req.URL.RawQuery = "ut-query-invalid=ut-query-value"
	res, err = set.extractor(req)
	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.EntryName())
	assert.NotEmpty(t, mock.EntryKind())
	assert.NotNil(t, mock.BeforeCtx(nil))
	mock.Before(nil)
}
