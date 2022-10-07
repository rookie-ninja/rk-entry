// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package jwt

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-entry/v3/entry"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestToOptions_One(t *testing.T) {
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
	signer := rk.NewSymmetricSignerJwt("ut-entry", jwt.SigningMethodHS256.Name, []byte("my-secret"))
	config = &BootConfig{
		Enabled:     true,
		SignerEntry: signer.Name(),
	}
	assert.NotEmpty(t, ToOptions(config, "", ""))
}

func TestToOptions_Two(t *testing.T) {
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
}

func TestToOptions_Three(t *testing.T) {
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
}

func TestToOptions_Four(t *testing.T) {
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
}

func TestToOptions_Five(t *testing.T) {
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
}

func TestNewOptionSet(t *testing.T) {
	// without option
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.EntryName())
	assert.NotNil(t, set.signer)
	assert.Equal(t, "header:"+rkm.HeaderAuthorization, set.tokenLookup)
	assert.Equal(t, "Bearer", set.authScheme)
	assert.Empty(t, set.pathToIgnore)
	assert.Len(t, set.extractors, 1)

	// with option
	tokenLookupStr := "query:ut,cookie:ut,form:ut,header:ut"
	signer := rk.NewSymmetricSignerJwt("ut-entry", jwt.SigningMethodHS256.Name, []byte("ut-key"))

	set = NewOptionSet(
		WithEntryNameAndKind("entry", "kind"),
		WithPathToIgnore("/ut-ignore"),
		WithSigner(signer),
		WithTokenLookup(tokenLookupStr),
		WithAuthScheme("ut-scheme")).(*optionSet)
	assert.NotEmpty(t, set.EntryName())
	assert.NotEmpty(t, set.EntryKind())
	assert.NotEmpty(t, set.pathToIgnore)
	assert.Equal(t, signer, set.signer)
	assert.Equal(t, "ut-scheme", set.authScheme)
	assert.Len(t, set.extractors, 3)
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
		WithSigner(rk.NewSymmetricSignerJwt("ut-entry", jwt.SigningMethodHS256.Name, []byte("my-secret"))))
	req := httptest.NewRequest(http.MethodGet, "/ut", nil)
	req.Header.Set(rkm.HeaderAuthorization,
		"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.EpM5XBzTJZ4J8AfoJEcJrjth8pfH28LWdjLo90sYb9g")
	ctx = set.BeforeCtx(req, nil)
	set.Before(ctx)
	assert.NotNil(t, ctx.Output.JwtToken)
	assert.Nil(t, ctx.Output.ErrResp)
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.EntryName())
	assert.NotEmpty(t, mock.EntryKind())
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
