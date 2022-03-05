// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkerror

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	res := New(http.StatusInternalServerError, "ut-error")

	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), res.Err.Status)
	assert.Empty(t, res.Err.Details)
}

func TestNewUnauthorized(t *testing.T) {
	// With rk error type
	res := NewUnauthorized("rk error")
	assert.Equal(t, http.StatusUnauthorized, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusUnauthorized), res.Err.Status)
	assert.Equal(t, "rk error", res.Err.Message)
}

func TestNewInternalError(t *testing.T) {
	// With rk error type
	res := NewInternalError("rk error")
	assert.Equal(t, http.StatusInternalServerError, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), res.Err.Status)
	assert.Equal(t, "rk error", res.Err.Message)
}

func TestNewBadRequest(t *testing.T) {
	// With rk error type
	res := NewBadRequest("rk error")
	assert.Equal(t, http.StatusBadRequest, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusBadRequest), res.Err.Status)
	assert.Equal(t, "rk error", res.Err.Message)
}

func TestNewForbidden(t *testing.T) {
	// With rk error type
	res := NewForbidden("rk error")
	assert.Equal(t, http.StatusForbidden, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusForbidden), res.Err.Status)
	assert.Equal(t, "rk error", res.Err.Message)
}

func TestNewTooManyRequests(t *testing.T) {
	// With rk error type
	res := NewTooManyRequests("rk error")
	assert.Equal(t, http.StatusTooManyRequests, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusTooManyRequests), res.Err.Status)
	assert.Equal(t, "rk error", res.Err.Message)
}

func TestNewTimeout(t *testing.T) {
	// With rk error type
	res := NewTimeout("rk error")
	assert.Equal(t, http.StatusRequestTimeout, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusRequestTimeout), res.Err.Status)
	assert.Equal(t, "rk error", res.Err.Message)
}

func TestFromError_WithNilError(t *testing.T) {
	res := FromError(nil)
	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), res.Err.Status)
	assert.Empty(t, res.Err.Details)
	assert.Equal(t, "unknown error", res.Err.Message)
}

func TestFromError_HappyCase(t *testing.T) {
	res := FromError(errors.New("ut error"))
	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.Err.Code)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), res.Err.Status)
	assert.Empty(t, res.Err.Details)
	assert.Equal(t, "ut error", res.Err.Message)
}
