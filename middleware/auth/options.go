// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidauth provide auth related options
package rkmidauth

import (
	"encoding/base64"
	"fmt"
	"github.com/rookie-ninja/rk-entry/v2/error"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"net/http"
	"strings"
)

const (
	authTypeBasic = "Basic"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	Before(*BeforeCtx)

	BeforeCtx(*http.Request) *BeforeCtx

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	entryName     string
	entryType     string
	basicRealm    string
	basicAccounts map[string]bool
	apiKey        map[string]bool
	pathToIgnore  []string
	mock          OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:     "fake-entry",
		entryType:     "",
		basicRealm:    "",
		basicAccounts: make(map[string]bool),
		apiKey:        make(map[string]bool),
		pathToIgnore:  []string{},
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	return set
}

// GetEntryName returns entry name
func (set *optionSet) GetEntryName() string {
	return set.entryName
}

// GetEntryType returns entry type
func (set *optionSet) GetEntryType() string {
	return set.entryType
}

// BeforeCtx should be created before Before()
func (set *optionSet) BeforeCtx(req *http.Request) *BeforeCtx {
	ctx := NewBeforeCtx()

	if req != nil && req.URL != nil && req.Header != nil {
		ctx.Input.UrlPath = req.URL.Path
		ctx.Input.BasicAuthHeader = req.Header.Get(rkmid.HeaderAuthorization)
		ctx.Input.ApiKeyHeader = req.Header.Get(rkmid.HeaderApiKey)
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	// normalize
	if ctx == nil {
		ctx.Output.ErrResp = rkerror.NewInternalError("Nil context")
		return
	}

	// case 0: ignore path
	if set.ShouldIgnore(ctx.Input.UrlPath) {
		return
	}

	// case 1: basic auth passed
	errBasic := set.isBasicAuthorized(ctx)
	if errBasic == nil {
		return
	}

	// case 2: X-API-Key passed
	errApiKey := set.isApiKeyAuthorized(ctx)
	if errApiKey == nil {
		return
	}

	// case 3: basic auth provided, then return code and response related to basic auth
	if len(ctx.Input.BasicAuthHeader) > 0 {
		ctx.Output.ErrResp = errBasic
		return
	}

	// case 4: X-API-Key provided, then return code and response related to X-API-Key
	if len(ctx.Input.ApiKeyHeader) > 0 {
		ctx.Output.ErrResp = errApiKey
		return
	}

	// case 5: no auth provided, return bellow code and response
	tmp := make([]string, 0)
	// case 5.1: basic auth needed
	if len(set.basicAccounts) > 0 {
		ctx.Output.HeadersToReturn["WWW-Authenticate"] = fmt.Sprintf(`%s realm="%s"`, authTypeBasic, set.basicRealm)
		tmp = append(tmp, "Basic Auth")
	}
	// case 5.2: X-API-Key needed
	if len(set.apiKey) > 0 {
		tmp = append(tmp, "X-API-Key")
	}

	ctx.Output.ErrResp = rkerror.NewUnauthorized(fmt.Sprintf("missing authorization, provide one of bellow auth header:[%s]", strings.Join(tmp, ",")))
}

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(path string) bool {
	if len(set.basicAccounts) < 1 && len(set.apiKey) < 1 {
		return true
	}

	for i := range set.pathToIgnore {
		if strings.HasPrefix(path, set.pathToIgnore[i]) {
			return true
		}
	}

	return rkmid.ShouldIgnoreGlobal(path)
}

// Validate basic auth
func (set *optionSet) isBasicAuthorized(ctx *BeforeCtx) *rkerror.ErrorResp {
	// case 1: auth header is provided
	if len(ctx.Input.BasicAuthHeader) > 0 {
		tokens := strings.SplitN(ctx.Input.BasicAuthHeader, " ", 2)
		// case 1.1: invalid basic auth
		if len(tokens) != 2 {
			return rkerror.NewUnauthorized("Invalid Basic Auth format")
		}

		// case 1.2: not authorized
		_, ok := set.basicAccounts[tokens[1]]
		if !ok {
			if tokens[0] == authTypeBasic {
				ctx.Output.HeadersToReturn["WWW-Authenticate"] = fmt.Sprintf(`%s realm="%s"`, authTypeBasic, set.basicRealm)
			}

			return rkerror.NewUnauthorized("Invalid credential")
		}

		// case 1.3: authorized
		return nil
	}

	// case 2: auth header missing
	return rkerror.NewUnauthorized("Missing authorization header")
}

// Validate X-API-Key
func (set *optionSet) isApiKeyAuthorized(ctx *BeforeCtx) *rkerror.ErrorResp {
	// case 1: auth header is provided
	if len(ctx.Input.ApiKeyHeader) > 0 {
		// case 1.1: not authorized
		_, ok := set.apiKey[ctx.Input.ApiKeyHeader]
		if !ok {
			return rkerror.NewUnauthorized("Invalid X-API-Key")
		}

		// case 1.2: authorized
		return nil
	}

	// case 2: auth header missing
	return rkerror.NewUnauthorized("Missing authorization header")
}

// ***************** OptionSet Mock *****************

// NewOptionSetMock for testing purpose
func NewOptionSetMock(before *BeforeCtx) OptionSetInterface {
	return &optionSetMock{
		before: before,
	}
}

type optionSetMock struct {
	before *BeforeCtx
}

// GetEntryName returns entry name
func (mock *optionSetMock) GetEntryName() string {
	return "mock"
}

// GetEntryType returns entry type
func (mock *optionSetMock) GetEntryType() string {
	return "mock"
}

// BeforeCtx should be created before Before()
func (mock *optionSetMock) BeforeCtx(request *http.Request) *BeforeCtx {
	return mock.before
}

// Before should run before user handler
func (mock *optionSetMock) Before(ctx *BeforeCtx) {
	return
}

// ShouldIgnore should run before user handler
func (mock *optionSetMock) ShouldIgnore(string) bool {
	return false
}

// ***************** Context *****************

// NewBeforeCtx create new BeforeCtx with fields initialized
func NewBeforeCtx() *BeforeCtx {
	ctx := &BeforeCtx{}
	ctx.Output.HeadersToReturn = make(map[string]string)
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		BasicAuthHeader string
		ApiKeyHeader    string
		UrlPath         string
	}
	Output struct {
		HeadersToReturn map[string]string
		ErrResp         *rkerror.ErrorResp
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Ignore  []string `yaml:"ignore" json:"ignore"`
	Basic   []string `yaml:"basic" json:"basic"`
	ApiKey  []string `yaml:"apiKey" json:"apiKey"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithBasicAuth(entryName, config.Basic...),
			WithApiKeyAuth(config.ApiKey...),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option for optionSet
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(set *optionSet) {
		set.entryName = entryName
		set.entryType = entryType
	}
}

// WithBasicAuth provide basic auth credentials formed as user:pass.
// We will encode credential with base64 since incoming credential from client would be encoded.
func WithBasicAuth(realm string, cred ...string) Option {
	return func(set *optionSet) {
		for i := range cred {
			set.basicAccounts[base64.StdEncoding.EncodeToString([]byte(cred[i]))] = true
		}

		set.basicRealm = realm
	}
}

// WithApiKeyAuth provide API Key auth credentials.
// An API key is a token that a client provides when making API calls.
// With API key auth, you send a key-value pair to the API either in the request headers or query parameters.
// Some APIs use API keys for authorization.
//
// The API key was injected into incoming header with key of X-API-Key
func WithApiKeyAuth(key ...string) Option {
	return func(set *optionSet) {
		for i := range key {
			set.apiKey[key[i]] = true
		}
	}
}

// WithPathToIgnore provide paths prefix that will ignore.
func WithPathToIgnore(paths ...string) Option {
	return func(set *optionSet) {
		for i := range paths {
			if len(paths[i]) > 0 {
				set.pathToIgnore = append(set.pathToIgnore, paths[i])
			}
		}
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}
