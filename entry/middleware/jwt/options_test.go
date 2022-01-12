// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidjwt

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	rkmid "github.com/rookie-ninja/rk-entry/entry/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestToOptions(t *testing.T) {
	// with disabled
	config := &BootConfig{
		Enabled: false,
	}
	assert.Empty(t, ToOptions(config, "", ""))

	// with enabled
	config.Enabled = true
	config.SigningKeys = []string{"key"}
	config.SigningKey = "key"
	assert.NotEmpty(t, ToOptions(config, "", ""))
}

func TestNewOptionSet(t *testing.T) {
	// without option
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.NotNil(t, set.signingKeys)
	assert.Equal(t, AlgorithmHS256, set.signingAlgorithm)
	assert.NotNil(t, set.claims)
	assert.Equal(t, "header:"+rkmid.HeaderAuthorization, set.tokenLookup)
	assert.Equal(t, "Bearer", set.authScheme)
	assert.Empty(t, set.ignorePrefix)
	assert.Len(t, set.extractors, 1)

	// with option
	claim := new(jwt.MapClaims)
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return nil, nil
	}
	parseTokenFunc := func(auth string) (*jwt.Token, error) {
		return nil, nil
	}
	tokenLookupStr := "query:ut,cookie:ut,form:ut,header:ut"

	set = NewOptionSet(
		WithEntryNameAndType("entry", "type"),
		WithIgnorePrefix("/ut-ignore"),
		WithSigningKey("key"),
		WithSigningKeys("key", "value"),
		WithSigningAlgorithm("ut-algo"),
		WithClaims(claim),
		WithTokenLookup(tokenLookupStr),
		WithAuthScheme("ut-scheme"),
		WithKeyFunc(keyFunc),
		WithParseTokenFunc(parseTokenFunc)).(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.NotEmpty(t, set.GetEntryType())
	assert.NotEmpty(t, set.ignorePrefix)
	assert.Equal(t, "key", set.signingKey)
	assert.Equal(t, "value", set.signingKeys["key"])
	assert.Equal(t, "ut-algo", set.signingAlgorithm)
	assert.Equal(t, claim, set.claims)
	assert.Equal(t, "ut-scheme", set.authScheme)
	assert.Equal(t, reflect.ValueOf(keyFunc).Pointer(), reflect.ValueOf(set.keyFunc).Pointer())
	assert.Equal(t, reflect.ValueOf(parseTokenFunc).Pointer(), reflect.ValueOf(set.parseTokenFunc).Pointer())
	assert.Len(t, set.extractors, 4)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil req
	set := NewOptionSet()
	ctx := set.BeforeCtx(nil, nil)
	assert.Nil(t, ctx.Input.Request)
	assert.Empty(t, ctx.Input.UrlPath)

	// happy case
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req, nil)
	assert.NotNil(t, ctx.Input.Request)
	assert.Equal(t, "/ut", ctx.Input.UrlPath)
}

func TestOptionSet_ignore(t *testing.T) {
	// with nil req
	set := NewOptionSet(WithIgnorePrefix("/ut")).(*optionSet)
	assert.True(t, set.ignore("/ut"))
	assert.False(t, set.ignore("/at"))
}

func TestOptionSet_Before(t *testing.T) {
	// with custom extractor and parser
	// expect extract error
	ex := func(ctx context.Context) (string, error) {
		return "", errors.New("ut-error")
	}
	set := NewOptionSet(WithExtractor(ex))
	ctx := set.BeforeCtx(nil, nil)
	set.Before(ctx)
	assert.Nil(t, ctx.Output.JwtToken)
	assert.NotNil(t, ctx.Output.ErrResp)

	// expect parse error
	ex = func(ctx context.Context) (string, error) {
		return "", nil
	}
	par := func(auth string) (*jwt.Token, error) {
		return nil, errors.New("ut-error")
	}
	set = NewOptionSet(
		WithExtractor(ex),
		WithParseTokenFunc(par))
	ctx = set.BeforeCtx(nil, nil)
	set.Before(ctx)
	assert.Nil(t, ctx.Output.JwtToken)
	assert.NotNil(t, ctx.Output.ErrResp)

	// happy case
	ex = func(ctx context.Context) (string, error) {
		return "", nil
	}
	par = func(auth string) (*jwt.Token, error) {
		return &jwt.Token{}, nil
	}
	set = NewOptionSet(
		WithExtractor(ex),
		WithParseTokenFunc(par))
	ctx = set.BeforeCtx(nil, nil)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.JwtToken)
	assert.Nil(t, ctx.Output.ErrResp)

	// Use default one but with invalid jwt
	set = NewOptionSet()
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	ctx = set.BeforeCtx(req, nil)
	set.Before(ctx)
	assert.Nil(t, ctx.Output.JwtToken)
	assert.NotNil(t, ctx.Output.ErrResp)

	// happy case
	set = NewOptionSet(WithSigningKey([]byte("my-secret")))
	req = httptest.NewRequest(http.MethodGet, "/ut", nil)
	req.Header.Set(rkmid.HeaderAuthorization,
		"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.EpM5XBzTJZ4J8AfoJEcJrjth8pfH28LWdjLo90sYb9g")
	ctx = set.BeforeCtx(req, nil)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.JwtToken)
	assert.Nil(t, ctx.Output.ErrResp)
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.GetEntryName())
	assert.NotEmpty(t, mock.GetEntryType())
	assert.NotNil(t, mock.BeforeCtx(nil, nil))
	mock.Before(nil)
}
