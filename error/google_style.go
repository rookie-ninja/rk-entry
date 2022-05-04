package rkerror

import (
	"encoding/json"
	"net/http"
)

func NewErrorBuilderGoogle() ErrorBuilder {
	return &ErrorBuilderGoogle{}
}

type ErrorBuilderGoogle struct{}

func (e *ErrorBuilderGoogle) New(code int, msg string, details ...interface{}) ErrorInterface {
	resp := &ErrorGoogle{}

	resp.Err.Code = code
	resp.Err.Status = http.StatusText(code)
	resp.Err.Message = msg
	resp.Err.Details = make([]interface{}, 0)

	if code < 1 {
		resp.Err.Code = http.StatusInternalServerError
		resp.Err.Status = http.StatusText(http.StatusInternalServerError)
	}

	resp.Err.Details = append(resp.Err.Details, details...)

	return resp
}

func (e *ErrorBuilderGoogle) NewCustom() ErrorInterface {
	return e.New(http.StatusInternalServerError, "")
}

// ErrorGoogle is standard google style error
// Referred google style: https://cloud.google.com/apis/design/errors
type ErrorGoogle struct {
	Err struct {
		Code    int           `json:"code" yaml:"code" example:"500"`
		Status  string        `json:"status" yaml:"status" example:"Internal Server Error"`
		Message string        `json:"message" yaml:"message" example:"Internal error occurs"`
		Details []interface{} `json:"details" yaml:"details"`
	} `json:"error" yaml:"error"`
}

func (err *ErrorGoogle) Code() int {
	return err.Err.Code
}

func (err *ErrorGoogle) Message() string {
	return err.Err.Message
}

func (err *ErrorGoogle) Details() []interface{} {
	return err.Err.Details
}

// Error returns string of error
func (err *ErrorGoogle) Error() string {
	res := "{}"

	if bytes, err := json.Marshal(err); err == nil {
		res = string(bytes)
	}

	return res
}
