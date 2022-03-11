// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidjwt

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestToOptions_One(t *testing.T) {
	defer rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)

	// with disabled
	config := &BootConfig{
		Enabled: false,
	}
	assert.Empty(t, ToOptions(config, "", ""))

	// with enabled
	config.Enabled = true
	config.Symmetric = &SymmetricConfig{
		Algorithm: jwt.SigningMethodHS256.Name,
		Token:     "ut-key",
	}
	assert.NotEmpty(t, ToOptions(config, "", ""))

	// with signer entry
	signer := rkentry.RegisterSymmetricJwtSigner("ut-entry", jwt.SigningMethodHS256.Name, []byte("my-secret"))
	rkentry.GlobalAppCtx.AddEntry(signer)
	config = &BootConfig{
		Enabled:     true,
		SignerEntry: signer.GetName(),
	}
	assert.NotEmpty(t, ToOptions(config, "", ""))
	rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
}

func TestToOptions_Two(t *testing.T) {
	defer rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
	// with symmetric
	// 1: with raw key
	config := &BootConfig{
		Enabled: true,
		Symmetric: &SymmetricConfig{
			Algorithm: jwt.SigningMethodHS256.Name,
			Token:     "raw key",
		},
	}
	assert.NotEmpty(t, ToOptions(config, "", ""))
	rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
}

func TestToOptions_Three(t *testing.T) {
	defer rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
	// with symmetric
	defer assertPanic(t)
	// 2: with path
	config := &BootConfig{
		Enabled: true,
		Symmetric: &SymmetricConfig{
			Algorithm: jwt.SigningMethodHS256.Name,
			TokenPath: "raw/path",
		},
	}
	assert.NotEmpty(t, ToOptions(config, "", ""))
	rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
}

func TestToOptions_Four(t *testing.T) {
	defer rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
	// with symmetric
	defer assertPanic(t)
	// 2: with path
	config := &BootConfig{
		Enabled: true,
		Asymmetric: &AsymmetricConfig{
			Algorithm:      jwt.SigningMethodRS256.Name,
			PublicKeyPath:  "raw/path",
			PrivateKeyPath: "raw/path",
		},
	}
	assert.NotEmpty(t, ToOptions(config, "", ""))
	rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
}

func TestToOptions_Five(t *testing.T) {
	defer rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
	// with symmetric
	defer assertPanic(t)
	// 2: with path
	config := &BootConfig{
		Enabled: true,
		Asymmetric: &AsymmetricConfig{
			Algorithm:  jwt.SigningMethodRS256.Name,
			PublicKey:  "raw/path",
			PrivateKey: "raw/path",
		},
	}
	assert.NotEmpty(t, ToOptions(config, "", ""))
	rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
}

func TestNewOptionSet(t *testing.T) {
	// without option
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.NotNil(t, set.signer)
	assert.Equal(t, "header:"+rkmid.HeaderAuthorization, set.tokenLookup)
	assert.Equal(t, "Bearer", set.authScheme)
	assert.Empty(t, set.pathToIgnore)
	assert.Len(t, set.extractors, 1)
	rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)

	// with option
	tokenLookupStr := "query:ut,cookie:ut,form:ut,header:ut"
	signer := rkentry.RegisterSymmetricJwtSigner("ut-entry", jwt.SigningMethodHS256.Name, []byte("ut-key"))

	set = NewOptionSet(
		WithEntryNameAndType("entry", "type"),
		WithPathToIgnore("/ut-ignore"),
		WithSigner(signer),
		WithTokenLookup(tokenLookupStr),
		WithAuthScheme("ut-scheme")).(*optionSet)
	assert.NotEmpty(t, set.GetEntryName())
	assert.NotEmpty(t, set.GetEntryType())
	assert.NotEmpty(t, set.pathToIgnore)
	assert.Equal(t, signer, set.signer)
	assert.Equal(t, "ut-scheme", set.authScheme)
	assert.Len(t, set.extractors, 2)
	rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)
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

func TestOptionSet_ShouldIgnore(t *testing.T) {
	// with nil req
	set := NewOptionSet(WithPathToIgnore("/ut")).(*optionSet)
	assert.True(t, set.ShouldIgnore("/ut"))
	assert.False(t, set.ShouldIgnore("/at"))
}

func TestOptionSet_Before(t *testing.T) {
	defer rkentry.GlobalAppCtx.RemoveEntryByType(rkentry.SignerJwtEntryType)

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

	// happy case
	set = NewOptionSet(
		WithSigner(rkentry.RegisterSymmetricJwtSigner("ut-entry", jwt.SigningMethodHS256.Name, []byte("my-secret"))))
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
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

func TestJwtHttpExtractor(t *testing.T) {
	// query
	f := jwtFromQuery("ut-query")
	res, err := f(nil)
	assert.NotNil(t, err)
	assert.Empty(t, res)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL = &url.URL{
		RawQuery: "ut-query=my-jwt",
	}
	res, err = f(req)

	assert.Nil(t, err)
	assert.Equal(t, "my-jwt", res)
}

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// this should never be called in case of a bug
		assert.True(t, false)
	}
}
