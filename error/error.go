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
func New(code int, msg string, details ...interface{}) *ErrorResp {
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

	resp.Err.Message = msg
	resp.Err.Details = append(resp.Err.Details, details...)

	return resp
}

func NewUnauthorized(msg string, details ...interface{}) *ErrorResp {
	return New(http.StatusUnauthorized, msg, details...)
}

func NewInternalError(msg string, details ...interface{}) *ErrorResp {
	return New(http.StatusInternalServerError, msg, details...)
}

func NewBadRequest(msg string, details ...interface{}) *ErrorResp {
	return New(http.StatusBadRequest, msg, details...)
}

func NewForbidden(msg string, details ...interface{}) *ErrorResp {
	return New(http.StatusForbidden, msg, details...)
}

func NewTooManyRequests(msg string, details ...interface{}) *ErrorResp {
	return New(http.StatusTooManyRequests, msg, details...)
}

func NewTimeout(msg string, details ...interface{}) *ErrorResp {
	return New(http.StatusRequestTimeout, msg, details...)
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
