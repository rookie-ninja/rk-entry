// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package secure provide security related options
package secure

import (
	"fmt"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"net/http"
	"strings"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	EntryName() string

	EntryKind() string

	Before(*BeforeCtx)

	BeforeCtx(*http.Request) *BeforeCtx

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	// EntryName name of entry
	entryName string

	// EntryKind type of entry
	entryKind string

	// pathToIgnore ignoring paths prefix
	pathToIgnore []string

	// XSSProtection provides protection against cross-site scripting attack (XSS)
	// by setting the `X-XSS-Protection` header.
	// Optional. Default value "1; mode=block".
	xssProtection string

	// ContentTypeNosniff provides protection against overriding Content-Type
	// header by setting the `X-Content-Type-Options` header.
	// Optional. Default value "nosniff".
	contentTypeNosniff string

	// XFrameOptions can be used to indicate whether or not a browser should
	// be allowed to render a page in a <frame>, <iframe> or <object> .
	// Sites can use this to avoid clickjacking attacks, by ensuring that their
	// content is not embedded into other sites.provides protection against
	// clickjacking.
	// Optional. Default value "SAMEORIGIN".
	// Possible values:
	// - "SAMEORIGIN" - The page can only be displayed in a frame on the same origin as the page itself.
	// - "DENY" - The page cannot be displayed in a frame, regardless of the site attempting to do so.
	// - "ALLOW-FROM uri" - The page can only be displayed in a frame on the specified origin.
	xFrameOptions string

	// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
	// long (in seconds) browsers should remember that this site is only to
	// be accessed using HTTPS. This reduces your exposure to some SSL-stripping
	// man-in-the-middle (MITM) attacks.
	// Optional. Default value 0.
	hstsMaxAge int

	// HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`
	// header, excluding all subdomains from security policy. It has no effect
	// unless HSTSMaxAge is set to a non-zero value.
	// Optional. Default value false.
	hstsExcludeSubdomains bool

	// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
	// security against cross-site scripting (XSS), clickjacking and other code
	// injection attacks resulting from execution of malicious content in the
	// trusted web page context.
	// Optional. Default value "".
	contentSecurityPolicy string

	// HSTSPreloadEnabled will add the preload tag in the `Strict Transport Security`
	// header, which enables the domain to be included in the HSTS preload list
	// maintained by Chrome (and used by Firefox and Safari): https://hstspreload.org/
	// Optional.  Default value false.
	hstsPreloadEnabled bool

	// CSPReportOnly would use the `Content-Security-Policy-Report-Only` header instead
	// of the `Content-Security-Policy` header. This allows iterative updates of the
	// content security policy by only reporting the violations that would
	// have occurred instead of blocking the resource.
	// Optional. Default value false.
	cspReportOnly bool

	// ReferrerPolicy sets the `Referrer-Policy` header providing security against
	// leaking potentially sensitive request paths to third parties.
	// Optional. Default value "".
	referrerPolicy string

	mock OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:          "fake-entry",
		entryKind:          "",
		xssProtection:      "1; mode=block",
		contentTypeNosniff: "nosniff",
		xFrameOptions:      "SAMEORIGIN",
		hstsPreloadEnabled: false,
		pathToIgnore:       make([]string, 0),
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	return set
}

// EntryName returns entry name
func (set *optionSet) EntryName() string {
	return set.entryName
}

// EntryKind returns entry kind
func (set *optionSet) EntryKind() string {
	return set.entryKind
}

// BeforeCtx should be created before Before()
func (set *optionSet) BeforeCtx(req *http.Request) *BeforeCtx {
	ctx := NewBeforeCtx()

	if req != nil && req.URL != nil && req.Header != nil {
		ctx.Input.UrlPath = req.URL.Path
		ctx.Input.isTLS = req.TLS != nil
		ctx.Input.xForwardedProto = req.Header.Get(rkm.HeaderXForwardedProto)
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	// normalize
	if ctx == nil || set.ShouldIgnore(ctx.Input.UrlPath) {
		return
	}

	// Add X-XSS-Protection header
	if set.xssProtection != "" {
		ctx.Output.HeadersToReturn[rkm.HeaderXXSSProtection] = set.xssProtection
	}

	// Add X-Content-Type-Options header
	if set.contentTypeNosniff != "" {
		ctx.Output.HeadersToReturn[rkm.HeaderXContentTypeOptions] = set.contentTypeNosniff
	}

	// Add X-Frame-Options header
	if set.xFrameOptions != "" {
		ctx.Output.HeadersToReturn[rkm.HeaderXFrameOptions] = set.xFrameOptions
	}

	// Add Strict-Transport-Security header
	if (ctx.Input.isTLS || (ctx.Input.xForwardedProto == "https")) && set.hstsMaxAge != 0 {
		subdomains := ""
		if !set.hstsExcludeSubdomains {
			subdomains = "; includeSubdomains"
		}
		if set.hstsPreloadEnabled {
			subdomains = fmt.Sprintf("%s; preload", subdomains)
		}
		ctx.Output.HeadersToReturn[rkm.HeaderStrictTransportSecurity] = fmt.Sprintf("max-age=%d%s", set.hstsMaxAge, subdomains)
	}

	// Add Content-Security-Policy-Report-Only or Content-Security-Policy header
	if set.contentSecurityPolicy != "" {
		if set.cspReportOnly {
			ctx.Output.HeadersToReturn[rkm.HeaderContentSecurityPolicyReportOnly] = set.contentSecurityPolicy
		} else {
			ctx.Output.HeadersToReturn[rkm.HeaderContentSecurityPolicy] = set.contentSecurityPolicy
		}
	}

	// Add Referrer-Policy header
	if set.referrerPolicy != "" {
		ctx.Output.HeadersToReturn[rkm.HeaderReferrerPolicy] = set.referrerPolicy
	}

}

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(path string) bool {
	for i := range set.pathToIgnore {
		if strings.HasPrefix(path, set.pathToIgnore[i]) {
			return true
		}
	}

	return rkm.ShouldIgnoreGlobal(path)
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

// EntryName returns entry name
func (mock *optionSetMock) EntryName() string {
	return "mock"
}

// EntryKind returns entry kind
func (mock *optionSetMock) EntryKind() string {
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
		UrlPath         string
		xForwardedProto string
		isTLS           bool
	}
	Output struct {
		HeadersToReturn map[string]string
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled               bool     `yaml:"enabled" json:"enabled"`
	Ignore                []string `yaml:"ignore" json:"ignore"`
	XssProtection         string   `yaml:"xssProtection" json:"xssProtection"`
	ContentTypeNosniff    string   `yaml:"contentTypeNosniff" json:"contentTypeNosniff"`
	XFrameOptions         string   `yaml:"xFrameOptions" json:"xFrameOptions"`
	HstsMaxAge            int      `yaml:"hstsMaxAge" json:"hstsMaxAge"`
	HstsExcludeSubdomains bool     `yaml:"hstsExcludeSubdomains" json:"hstsExcludeSubdomains"`
	HstsPreloadEnabled    bool     `yaml:"hstsPreloadEnabled" json:"hstsPreloadEnabled"`
	ContentSecurityPolicy string   `yaml:"contentSecurityPolicy" json:"contentSecurityPolicy"`
	CspReportOnly         bool     `yaml:"cspReportOnly" json:"cspReportOnly"`
	ReferrerPolicy        string   `yaml:"referrerPolicy" json:"referrerPolicy"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndKind(entryName, entryType),
			WithXSSProtection(config.XssProtection),
			WithContentTypeNosniff(config.ContentTypeNosniff),
			WithXFrameOptions(config.XFrameOptions),
			WithHSTSMaxAge(config.HstsMaxAge),
			WithHSTSExcludeSubdomains(config.HstsExcludeSubdomains),
			WithHSTSPreloadEnabled(config.HstsPreloadEnabled),
			WithContentSecurityPolicy(config.ContentSecurityPolicy),
			WithCSPReportOnly(config.CspReportOnly),
			WithReferrerPolicy(config.ReferrerPolicy),
			WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option
type Option func(*optionSet)

// WithEntryNameAndKind provide entry name and entry kind.
func WithEntryNameAndKind(name, kind string) Option {
	return func(set *optionSet) {
		set.entryName = name
		set.entryKind = kind
	}
}

// WithXSSProtection provide X-XSS-Protection header value.
// Optional. Default value "1; mode=block".
func WithXSSProtection(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.xssProtection = val
		}
	}
}

// WithContentTypeNosniff provide X-Content-Type-Options header value.
// Optional. Default value "nosniff".
func WithContentTypeNosniff(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.contentTypeNosniff = val
		}
	}
}

// WithXFrameOptions provide X-Frame-Options header value.
// Optional. Default value "SAMEORIGIN".
func WithXFrameOptions(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.xFrameOptions = val
		}
	}
}

// WithHSTSMaxAge provide Strict-Transport-Security header value.
func WithHSTSMaxAge(val int) Option {
	return func(opt *optionSet) {
		opt.hstsMaxAge = val
	}
}

// WithHSTSExcludeSubdomains provide excluding subdomains of HSTS.
func WithHSTSExcludeSubdomains(val bool) Option {
	return func(opt *optionSet) {
		opt.hstsExcludeSubdomains = val
	}
}

// WithHSTSPreloadEnabled provide enabling HSTS preload.
// Optional. Default value false.
func WithHSTSPreloadEnabled(val bool) Option {
	return func(opt *optionSet) {
		opt.hstsPreloadEnabled = val
	}
}

// WithContentSecurityPolicy provide Content-Security-Policy header value.
// Optional. Default value "".
func WithContentSecurityPolicy(val string) Option {
	return func(opt *optionSet) {
		opt.contentSecurityPolicy = val
	}
}

// WithCSPReportOnly provide Content-Security-Policy-Report-Only header value.
// Optional. Default value false.
func WithCSPReportOnly(val bool) Option {
	return func(set *optionSet) {
		set.cspReportOnly = val
	}
}

// WithReferrerPolicy provide Referrer-Policy header value.
// Optional. Default value "".
func WithReferrerPolicy(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.referrerPolicy = val
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
