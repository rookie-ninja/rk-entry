// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// package rkmidcsrf provide auth related options
package rkmidcsrf

import (
	"context"
	"crypto/subtle"
	"errors"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rs/xid"
	"net/http"
	"net/url"
	"strings"
	"time"
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

	// TokenLength is the length of the generated token.
	tokenLength int

	// TokenLookup is a string in the form of "<source>:<key>" that is used
	// to extract token from the request.
	// Optional. Default value "header:X-CSRF-Token".
	// Possible values:
	// - "header:<name>"
	// - "form:<name>"
	// - "query:<name>"
	tokenLookup string

	// CookieName Name of the CSRF cookie. This cookie will store CSRF token.
	// Optional. Default value "_csrf".
	cookieName string

	// CookieDomain Domain of the CSRF cookie.
	// Optional. Default value none.
	cookieDomain string

	// CookiePath Path of the CSRF cookie.
	// Optional. Default value none.
	cookiePath string

	// CookieMaxAge Max age (in seconds) of the CSRF cookie.
	// Optional. Default value 86400 (24hr).
	cookieMaxAge int

	// CookieHTTPOnly Indicates if CSRF cookie is HTTP only.
	// Optional. Default value false.
	cookieHTTPOnly bool

	// CookieSameSite Indicates SameSite mode of the CSRF cookie.
	// Optional. Default value SameSiteDefaultMode.
	cookieSameSite http.SameSite

	extractor csrfHttpExtractor

	userExtractor CsrfExtractor

	mock OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:      xid.New().String(),
		entryType:      "",
		tokenLength:    32,
		tokenLookup:    "header:" + rkmid.HeaderXCSRFToken,
		cookieName:     "_csrf",
		cookieMaxAge:   86400,
		cookieSameSite: http.SameSiteDefaultMode,
		ignorePrefix:   make([]string, 0),
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	// initialize extractor
	parts := strings.Split(set.tokenLookup, ":")
	set.extractor = csrfTokenFromHeader(parts[1])
	switch parts[0] {
	case "form":
		set.extractor = csrfTokenFromForm(parts[1])
	case "query":
		set.extractor = csrfTokenFromQuery(parts[1])
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
		ctx.Input.Method = req.Method
		if cookie, err := req.Cookie(set.cookieName); err != nil {
			ctx.Input.Token = rkcommon.RandString(set.tokenLength)
		} else {
			ctx.Input.Token, _ = url.QueryUnescape(cookie.Value)
		}
		ctx.Input.Request = req
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	// normalize
	if ctx == nil || set.ignore(ctx.Input.UrlPath) {
		return
	}

	// 3.1: do not check http methods of GET, HEAD, OPTIONS and TRACE
	switch ctx.Input.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
	default:
		var clientToken string
		var err error
		// 3.2: validate token only for requests which are not defined as 'safe' by RFC7231
		if set.userExtractor != nil {
			clientToken, err = set.userExtractor(ctx.Input.UserCtx)
		} else {
			clientToken, err = set.extractor(ctx.Input.Request)
		}

		if err != nil {
			ctx.Output.ErrResp = rkerror.New(
				rkerror.WithHttpCode(http.StatusBadRequest),
				rkerror.WithMessage("failed to extract client token"),
				rkerror.WithDetails(err))
			return
		}

		// 3.3: return 403 to client if token is not matched
		if !set.isValidToken(ctx.Input.Token, clientToken) {
			ctx.Output.ErrResp = rkerror.New(
				rkerror.WithHttpCode(http.StatusForbidden),
				rkerror.WithMessage("invalid csrf token"),
				rkerror.WithDetails(err))
			return
		}
	}

	// set CSRF cookie
	cookie := new(http.Cookie)
	cookie.Name = set.cookieName
	cookie.Value = ctx.Input.Token
	// 4.1
	if set.cookiePath != "" {
		cookie.Path = set.cookiePath
	}
	// 4.2
	if set.cookieDomain != "" {
		cookie.Domain = set.cookieDomain
	}
	// 4.3
	if set.cookieSameSite != http.SameSiteDefaultMode {
		cookie.SameSite = set.cookieSameSite
	}
	cookie.Expires = time.Now().Add(time.Duration(set.cookieMaxAge) * time.Second)
	cookie.Secure = set.cookieSameSite == http.SameSiteNoneMode
	cookie.HttpOnly = set.cookieHTTPOnly
	ctx.Output.Cookie = cookie

	ctx.Output.VaryHeaders = append(ctx.Output.VaryHeaders, rkmid.HeaderCookie)
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

func (set *optionSet) isValidToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
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
	ctx.Output.VaryHeaders = make([]string, 0)
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		UrlPath string
		Method  string
		Token   string
		Request *http.Request
		UserCtx context.Context
	}
	Output struct {
		VaryHeaders []string
		Cookie      *http.Cookie
		ErrResp     *rkerror.ErrorResp
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	IgnorePrefix   []string `yaml:"ignorePrefix" json:"ignorePrefix"`
	TokenLength    int      `yaml:"tokenLength" json:"tokenLength"`
	TokenLookup    string   `yaml:"tokenLookup" json:"tokenLookup"`
	CookieName     string   `yaml:"cookieName" json:"cookieName"`
	CookieDomain   string   `yaml:"cookieDomain" json:"cookieDomain"`
	CookiePath     string   `yaml:"cookiePath" json:"cookiePath"`
	CookieMaxAge   int      `yaml:"cookieMaxAge" json:"cookieMaxAge"`
	CookieHttpOnly bool     `yaml:"cookieHttpOnly" json:"cookieHttpOnly"`
	CookieSameSite string   `yaml:"cookieSameSite" json:"cookieSameSite"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts,
			WithEntryNameAndType(entryName, entryType),
			WithTokenLength(config.TokenLength),
			WithTokenLookup(config.TokenLookup),
			WithCookieName(config.CookieName),
			WithCookieDomain(config.CookieDomain),
			WithCookiePath(config.CookiePath),
			WithCookieMaxAge(config.CookieMaxAge),
			WithCookieHTTPOnly(config.CookieHttpOnly),
			WithIgnorePrefix(config.IgnorePrefix...))

		// convert to string to cookie same sites
		sameSite := http.SameSiteDefaultMode

		switch strings.ToLower(config.CookieSameSite) {
		case "lax":
			sameSite = http.SameSiteLaxMode
		case "strict":
			sameSite = http.SameSiteStrictMode
		case "none":
			sameSite = http.SameSiteNoneMode
		default:
			sameSite = http.SameSiteDefaultMode
		}

		opts = append(opts, WithCookieSameSite(sameSite))
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

// WithTokenLength the length of the generated token.
// Optional. Default value 32.
func WithTokenLength(val int) Option {
	return func(opt *optionSet) {
		if val > 0 {
			opt.tokenLength = val
		}
	}
}

// WithTokenLookup a string in the form of "<source>:<key>" that is used
// to extract token from the request.
// Optional. Default value "header:X-CSRF-Token".
// Possible values:
// - "header:<name>"
// - "form:<name>"
// - "query:<name>"
// Optional. Default value "header:X-CSRF-Token".
func WithTokenLookup(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.tokenLookup = val
		}
	}
}

// WithCookieName provide name of the CSRF cookie. This cookie will store CSRF token.
// Optional. Default value "csrf".
func WithCookieName(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.cookieName = val
		}
	}
}

// WithCookieDomain provide domain of the CSRF cookie.
// Optional. Default value "".
func WithCookieDomain(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.cookieDomain = val
		}
	}
}

// WithCookiePath provide path of the CSRF cookie.
// Optional. Default value "".
func WithCookiePath(val string) Option {
	return func(opt *optionSet) {
		if len(val) > 0 {
			opt.cookiePath = val
		}
	}
}

// WithCookieMaxAge provide max age (in seconds) of the CSRF cookie.
// Optional. Default value 86400 (24hr).
func WithCookieMaxAge(val int) Option {
	return func(opt *optionSet) {
		if val > 0 {
			opt.cookieMaxAge = val
		}
	}
}

// WithCookieHTTPOnly indicates if CSRF cookie is HTTP only.
// Optional. Default value false.
func WithCookieHTTPOnly(val bool) Option {
	return func(opt *optionSet) {
		opt.cookieHTTPOnly = val
	}
}

// WithCookieSameSite indicates SameSite mode of the CSRF cookie.
// Optional. Default value SameSiteDefaultMode.
func WithCookieSameSite(val http.SameSite) Option {
	return func(opt *optionSet) {
		opt.cookieSameSite = val
	}
}

// WithExtractor provide user extractor
func WithExtractor(ex CsrfExtractor) Option {
	return func(opt *optionSet) {
		opt.userExtractor = ex
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

// ***************** Extractor *****************

type CsrfExtractor func(ctx context.Context) (string, error)

// CsrfTokenExtractor defines a function that takes `echo.Context` and returns
// either a token or an error.
type csrfHttpExtractor func(*http.Request) (string, error)

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided request header.
func csrfTokenFromHeader(header string) csrfHttpExtractor {
	return func(req *http.Request) (string, error) {
		token := req.Header.Get(header)
		if token == "" {
			return "", errors.New("missing csrf token in header")
		}
		return token, nil
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided form parameter.
func csrfTokenFromForm(param string) csrfHttpExtractor {
	return func(req *http.Request) (string, error) {
		token := req.Form.Get(param)
		if token == "" {
			return "", errors.New("missing csrf token in the form parameter")
		}
		return token, nil
	}
}

// csrfTokenFromQuery returns a `csrfTokenExtractor` that extracts token from the
// provided query parameter.
func csrfTokenFromQuery(param string) csrfHttpExtractor {
	return func(req *http.Request) (string, error) {
		token := req.URL.Query().Get(param)
		if token == "" {
			return "", errors.New("missing csrf token in the query string")
		}
		return token, nil
	}
}
