package rkerror

import (
	"encoding/json"
	"net/http"
)

func NewErrorBuilderAMZN() ErrorBuilder {
	return &ErrorBuilderAMZN{}
}

type ErrorBuilderAMZN struct{}

func (e *ErrorBuilderAMZN) New(code int, msg string, details ...interface{}) ErrorInterface {
	resp := &ErrorAMZN{}
	resp.Resp.Errors = make([]*ErrorElementAMZN, 0)

	element := &ErrorElementAMZN{}
	element.Err.Code = code
	element.Err.Status = http.StatusText(code)
	element.Err.Message = msg
	element.Err.Details = make([]interface{}, 0)

	if code < 1 {
		element.Err.Code = http.StatusInternalServerError
		element.Err.Status = http.StatusText(http.StatusInternalServerError)
	}

	for i := range details {
		detail := details[i]
		if v, ok := detail.(error); ok {
			element.Err.Details = append(element.Err.Details, v.Error())
		} else {
			element.Err.Details = append(element.Err.Details, detail)
		}
	}

	resp.Resp.Errors = append(resp.Resp.Errors, element)

	return resp
}

func (e *ErrorBuilderAMZN) NewCustom() ErrorInterface {
	return e.New(http.StatusInternalServerError, "")
}

type ErrorAMZN struct {
	Resp struct {
		Errors []*ErrorElementAMZN `json:"errors" yaml:"errors"`
	} `json:"response" yaml:"response"`
}

type ErrorElementAMZN struct {
	Err struct {
		Code    int           `json:"code" yaml:"code" example:"500"`
		Status  string        `json:"status" yaml:"status" example:"Internal Server Error"`
		Message string        `json:"message" yaml:"message" example:"Internal error occurs"`
		Details []interface{} `json:"details" yaml:"details"`
	} `json:"error" json:"error"`
}

func (err *ErrorAMZN) Code() int {
	res := 0

	if len(err.Resp.Errors) > 0 {
		return err.Resp.Errors[0].Err.Code
	}

	return res
}

func (err *ErrorAMZN) Message() string {
	res := ""

	if len(err.Resp.Errors) > 0 {
		return err.Resp.Errors[0].Err.Message
	}

	return res
}

func (err *ErrorAMZN) Details() []interface{} {
	res := make([]interface{}, 0)

	if len(err.Resp.Errors) > 0 {
		return err.Resp.Errors[0].Err.Details
	}

	return res
}

// Error returns string of error
func (err *ErrorAMZN) Error() string {
	res := "{}"

	if bytes, marshalErr := json.Marshal(err); marshalErr == nil {
		res = string(bytes)
	}

	return res
}
