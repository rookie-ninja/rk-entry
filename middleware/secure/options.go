// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// package rkmidsec provide auth related options
package rkmidsec

import (
	"fmt"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rs/xid"
	"net/http"
	"strings"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	Before(*BeforeCtx)

	BeforeCtx(*http.Request) *BeforeCtx
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	// EntryName name of entry
	entryName string

	// EntryType type of entry
	entryType string

	// IgnorePrefix ignoring paths prefix
	ignorePrefix []string

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
		entryName:          xid.New().String(),
		entryType:          "",
		xssProtection:      "1; mode=block",
		contentTypeNosniff: "nosniff",
		xFrameOptions:      "SAMEORIGIN",
		hstsPreloadEnabled: false,
		ignorePrefix:       make([]string, 0),
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
		ctx.Input.isTLS = req.TLS != nil
		ctx.Input.xForwardedProto = req.Header.Get(rkmid.HeaderXForwardedProto)
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	// normalize
	if ctx == nil || set.ignore(ctx.Input.UrlPath) {
		return
	}

	// Add X-XSS-Protection header
	if set.xssProtection != "" {
		ctx.Output.HeadersToReturn[rkmid.HeaderXXSSProtection] = set.xssProtection
	}

	// Add X-Content-Type-Options header
	if set.contentTypeNosniff != "" {
		ctx.Output.HeadersToReturn[rkmid.HeaderXContentTypeOptions] = set.contentTypeNosniff
	}

	// Add X-Frame-Options header
	if set.xFrameOptions != "" {
		ctx.Output.HeadersToReturn[rkmid.HeaderXFrameOptions] = set.xFrameOptions
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
		ctx.Output.HeadersToReturn[rkmid.HeaderStrictTransportSecurity] = fmt.Sprintf("max-age=%d%s", set.hstsMaxAge, subdomains)
	}

	// Add Content-Security-Policy-Report-Only or Content-Security-Policy header
	if set.contentSecurityPolicy != "" {
		if set.cspReportOnly {
			ctx.Output.HeadersToReturn[rkmid.HeaderContentSecurityPolicyReportOnly] = set.contentSecurityPolicy
		} else {
			ctx.Output.HeadersToReturn[rkmid.HeaderContentSecurityPolicy] = set.contentSecurityPolicy
		}
	}

	// Add Referrer-Policy header
	if set.referrerPolicy != "" {
		ctx.Output.HeadersToReturn[rkmid.HeaderReferrerPolicy] = set.referrerPolicy
	}

}

// Ignore determine whether auth should be ignored based on path
func (set *optionSet) ignore(path string) bool {
	for i := range set.ignorePrefix {
		if strings.HasPrefix(path, set.ignorePrefix[i]) {
			return true
		}
	}

	return rkmid.IgnorePrefixGlobal(path)
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
	IgnorePrefix          []string `yaml:"ignorePrefix" json:"ignorePrefix"`
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
			WithEntryNameAndType(entryName, entryType),
			WithXSSProtection(config.XssProtection),
			WithContentTypeNosniff(config.ContentTypeNosniff),
			WithXFrameOptions(config.XFrameOptions),
			WithHSTSMaxAge(config.HstsMaxAge),
			WithHSTSExcludeSubdomains(config.HstsExcludeSubdomains),
			WithHSTSPreloadEnabled(config.HstsPreloadEnabled),
			WithContentSecurityPolicy(config.ContentSecurityPolicy),
			WithCSPReportOnly(config.CspReportOnly),
			WithReferrerPolicy(config.ReferrerPolicy),
			WithIgnorePrefix(config.IgnorePrefix...))
	}

	return opts
}

// ***************** Option *****************

// Option
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(set *optionSet) {
		set.entryName = entryName
		set.entryType = entryType
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

// WithIgnorePrefix provide paths prefix that will ignore.
// Mainly used for swagger main page and RK TV entry.
func WithIgnorePrefix(paths ...string) Option {
	return func(set *optionSet) {
		set.ignorePrefix = append(set.ignorePrefix, paths...)
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}
