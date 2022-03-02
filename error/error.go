// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkerror defines RK style API errors.
package rkerror

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorResp is standard rk style error
type ErrorResp struct {
	Err *Error `json:"error" yaml:"error"` // Err is RK style error type
}

// New error response with options
func New(code int, details ...interface{}) *ErrorResp {
	resp := &ErrorResp{
		Err: &Error{
			Code:    code,
			Status:  http.StatusText(code),
			Details: make([]interface{}, 0),
		},
	}

	if code < 1 {
		resp.Err.Code = http.StatusInternalServerError
		resp.Err.Status = http.StatusText(http.StatusInternalServerError)
	}

	resp.Err.Message = http.StatusText(resp.Err.Code)

	resp.Err.Details = append(resp.Err.Details, details...)

	return resp
}

func NewUnauthorized(details ...interface{}) *ErrorResp {
	return New(http.StatusUnauthorized, details...)
}

func NewInternalError(details ...interface{}) *ErrorResp {
	return New(http.StatusInternalServerError, details...)
}

func NewBadRequest(details ...interface{}) *ErrorResp {
	return New(http.StatusBadRequest, details...)
}

func NewForbidden(details ...interface{}) *ErrorResp {
	return New(http.StatusForbidden, details...)
}

func NewTooManyRequests(details ...interface{}) *ErrorResp {
	return New(http.StatusTooManyRequests, details...)
}

func NewTimeout(details ...interface{}) *ErrorResp {
	return New(http.StatusRequestTimeout, details...)
}

// FromError converts error to ErrorResp
func FromError(err error) *ErrorResp {
	if err == nil {
		err = errors.New("unknown error")
	}

	return &ErrorResp{
		Err: &Error{
			Code:    http.StatusInternalServerError,
			Status:  http.StatusText(http.StatusInternalServerError),
			Details: make([]interface{}, 0),
			Message: err.Error(),
		},
	}
}

// Error defines standard error types of rk style
type Error struct {
	Code    int           `json:"code" yaml:"code"`       // Code represent codes in response
	Status  string        `json:"status" yaml:"status"`   // Status represent string value of code
	Message string        `json:"message" yaml:"message"` // Message represent detail message
	Details []interface{} `json:"details" yaml:"details"` // Details is a list of details in any types in string
}

// Error returns string of error
func (err *Error) Error() string {
	return fmt.Sprintf("[%s] %s", err.Status, err.Message)
}
