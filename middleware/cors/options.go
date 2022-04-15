// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidcors provide cors related options
package rkmidcors

import (
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	entryName    string
	entryType    string
	pathToIgnore []string
	mock         OptionSetInterface
	// AllowOrigins defines a list of origins that may access the resource.
	// Optional. Default value []string{"*"}.
	allowOrigins []string
	// allowPatterns derived from AllowOrigins by parsing regex fields
	// auto generated when creating new optionSet was created
	allowPatterns []string
	// AllowMethods defines a list methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	// Optional. Default value DefaultCORSConfig.AllowMethods.
	allowMethods []string
	// AllowHeaders defines a list of request headers that can be used when
	// making the actual request. This is in response to a preflight request.
	// Optional. Default value []string{}.
	allowHeaders []string
	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true. When used as part of
	// a response to a preflight request, this indicates whether or not the
	// actual request can be made using credentials.
	// Optional. Default value false.
	allowCredentials bool
	// ExposeHeaders defines a whitelist headers that clients are allowed to
	// access.
	// Optional. Default value []string{}.
	exposeHeaders []string
	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	// Optional. Default value 0.
	maxAge int
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:        "fake-entry",
		entryType:        "",
		pathToIgnore:     []string{},
		allowOrigins:     []string{},
		allowMethods:     []string{},
		allowHeaders:     []string{},
		allowCredentials: false,
		exposeHeaders:    []string{},
		maxAge:           0,
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	if len(set.allowOrigins) < 1 {
		set.allowOrigins = append(set.allowOrigins, "*")
	}

	if len(set.allowMethods) < 1 {
		set.allowMethods = append(set.allowMethods,
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete)
	}

	// parse regex pattern in origins
	set.toPatterns()

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
		ctx.Input.OriginHeader = req.Header.Get(rkmid.HeaderOrigin)
		ctx.Input.AccessControlRequestHeaders = req.Header.Get(rkmid.HeaderAccessControlRequestHeaders)
		ctx.Input.IsPreflight = req.Method == http.MethodOptions
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	if ctx == nil || set.ShouldIgnore(ctx.Input.UrlPath) {
		return
	}

	// case 1: if no origin header was provided, we will return 204 if request is not a OPTION method
	if ctx.Input.OriginHeader == "" {
		// 1.1: if not a preflight request, then pass through
		if !ctx.Input.IsPreflight {
			return
		}

		// 1.2: if it is a preflight request, then return with 204
		ctx.Output.Abort = true
		return
	}

	// case 2: origin not allowed, we will return 204 if request is not a OPTION method
	if !set.isOriginAllowed(ctx.Input.OriginHeader) {
		ctx.Output.Abort = true
		return
	}

	// case 3: not a OPTION method
	if !ctx.Input.IsPreflight {
		ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlAllowOrigin] = ctx.Input.OriginHeader

		// 3.1: add Access-Control-Allow-Credentials
		if set.allowCredentials {
			ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlAllowCredentials] = "true"
		}
		// 3.2: add Access-Control-Expose-Headers
		if len(set.exposeHeaders) > 0 {
			ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlExposeHeaders] = strings.Join(set.exposeHeaders, ",")
		}
		return
	}

	// 4: preflight request, return 204
	// add related headers including:
	//
	// - Vary
	// - Access-Control-Allow-Origin
	// - Access-Control-Allow-Methods
	// - Access-Control-Allow-Credentials
	// - Access-Control-Allow-Headers
	// - Access-Control-Max-Age
	ctx.Output.HeaderVary = append(ctx.Output.HeaderVary,
		rkmid.HeaderAccessControlRequestMethod,
		rkmid.HeaderAccessControlRequestHeaders)
	ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlAllowOrigin] = ctx.Input.OriginHeader
	ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlAllowMethods] = strings.Join(set.allowMethods, ",")

	// 4.1: Access-Control-Allow-Credentials
	if set.allowCredentials {
		ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlAllowCredentials] = "true"
	}

	// 4.2: Access-Control-Allow-Headers
	if len(set.allowHeaders) > 0 {
		ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlAllowHeaders] = strings.Join(set.allowHeaders, ",")
	} else {
		if ctx.Input.AccessControlRequestHeaders != "" {
			ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlAllowHeaders] = ctx.Input.AccessControlRequestHeaders
		}
	}

	if set.maxAge > 0 {
		// 4.3: Access-Control-Max-Age
		ctx.Output.HeadersToReturn[rkmid.HeaderAccessControlMaxAge] = strconv.Itoa(set.maxAge)
	}

	ctx.Output.Abort = true
}

// Convert allowed origins to patterns
func (set *optionSet) toPatterns() {
	set.allowPatterns = []string{}

	for _, raw := range set.allowOrigins {
		var result strings.Builder
		result.WriteString("^")
		for i, literal := range strings.Split(raw, "*") {

			// Replace * with .*
			if i > 0 {
				result.WriteString(".*")
			}

			result.WriteString(literal)
		}
		result.WriteString("$")
		set.allowPatterns = append(set.allowPatterns, result.String())
	}
}

// Check based on origin header
func (set *optionSet) isOriginAllowed(originHeader string) bool {
	res := false

	for _, pattern := range set.allowPatterns {
		res, _ = regexp.MatchString(pattern, originHeader)
		if res {
			break
		}
	}

	return res
}

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(path string) bool {
	for i := range set.pathToIgnore {
		if strings.HasPrefix(path, set.pathToIgnore[i]) {
			return true
		}
	}

	return rkmid.ShouldIgnoreGlobal(path)
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
	ctx.Output.HeaderVary = make([]string, 0)
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		UrlPath                     string
		OriginHeader                string
		IsPreflight                 bool
		AccessControlRequestHeaders string
	}
	Output struct {
		HeadersToReturn map[string]string
		HeaderVary      []string
		Abort           bool
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled"`
	AllowOrigins     []string `yaml:"allowOrigins" json:"allowOrigins"`
	AllowCredentials bool     `yaml:"allowCredentials" json:"allowCredentials"`
	AllowHeaders     []string `yaml:"allowHeaders" json:"allowHeaders"`
	AllowMethods     []string `yaml:"allowMethods" json:"allowMethods"`
	ExposeHeaders    []string `yaml:"exposeHeaders" json:"exposeHeaders"`
	MaxAge           int      `yaml:"maxAge" json:"maxAge"`
	Ignore           []string `yaml:"ignore" json:"ignore"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithAllowOrigins(config.AllowOrigins...),
			WithAllowCredentials(config.AllowCredentials),
			WithExposeHeaders(config.ExposeHeaders...),
			WithMaxAge(config.MaxAge),
			WithAllowHeaders(config.AllowHeaders...),
			WithAllowMethods(config.AllowMethods...),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.entryName = entryName
		opt.entryType = entryType
	}
}

// WithAllowOrigins provide allowed origins.
func WithAllowOrigins(origins ...string) Option {
	return func(opt *optionSet) {
		opt.allowOrigins = append(opt.allowOrigins, origins...)
	}
}

// WithAllowMethods provide allowed http methods
func WithAllowMethods(methods ...string) Option {
	return func(opt *optionSet) {
		opt.allowMethods = append(opt.allowMethods, methods...)
	}
}

// WithAllowHeaders provide allowed headers
func WithAllowHeaders(headers ...string) Option {
	return func(opt *optionSet) {
		opt.allowHeaders = append(opt.allowHeaders, headers...)
	}
}

// WithAllowCredentials allow credentials or not
func WithAllowCredentials(allow bool) Option {
	return func(opt *optionSet) {
		opt.allowCredentials = allow
	}
}

// WithExposeHeaders provide expose headers
func WithExposeHeaders(headers ...string) Option {
	return func(opt *optionSet) {
		opt.exposeHeaders = append(opt.exposeHeaders, headers...)
	}
}

// WithMaxAge provide max age
func WithMaxAge(age int) Option {
	return func(opt *optionSet) {
		opt.maxAge = age
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
