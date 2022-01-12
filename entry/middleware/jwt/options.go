// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidjwt is a middleware for JWT
package rkmidjwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-common/error"
	rkmid "github.com/rookie-ninja/rk-entry/entry/middleware"
	"github.com/rs/xid"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/jwt.go

var (
	errJwtMissing = rkerror.New(
		rkerror.WithHttpCode(http.StatusBadRequest),
		rkerror.WithMessage("missing or malformed jwt"))
	errJwtInvalid = rkerror.New(
		rkerror.WithHttpCode(http.StatusUnauthorized),
		rkerror.WithMessage("invalid or expired jwt"))
)

const (
	// AlgorithmHS256 is default algorithm for jwt
	AlgorithmHS256 = "HS256"
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	Before(*BeforeCtx)

	BeforeCtx(*http.Request, context.Context) *BeforeCtx
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	// EntryName name of entry
	entryName string

	// EntryType type of entry
	entryType string

	ignorePrefix []string

	// extractors of JWT token from http.Request
	extractors []jwtHttpExtractor

	// user provided extractor
	extractor JwtExtractor

	// SigningKey Signing key to validate token.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither user-defined KeyFunc nor SigningKeys is provided.
	signingKey interface{}

	// SigningKeys Map of signing keys to validate token with kid field usage.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither user-defined KeyFunc nor SigningKey is provided.
	signingKeys map[string]interface{}

	// SigningAlgorithm Signing algorithm used to check the token's signing algorithm.
	// Optional. Default value HS256.
	signingAlgorithm string

	// Claims are extendable claims data defining token content. Used by default ParseTokenFunc implementation.
	// Not used if custom ParseTokenFunc is set.
	// Optional. Default value jwt.MapClaims
	claims jwt.Claims

	// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "param:<name>"
	// - "cookie:<name>"
	// - "form:<name>"
	// Multiply sources example:
	// - "header: Authorization,cookie: myowncookie"
	tokenLookup string

	// AuthScheme to be used in the Authorization header.
	// Optional. Default value "Bearer".
	authScheme string

	// KeyFunc defines a user-defined function that supplies the public key for a token validation.
	// The function shall take care of verifying the signing algorithm and selecting the proper key.
	// A user-defined KeyFunc can be useful if tokens are issued by an external party.
	// Used by default ParseTokenFunc implementation.
	//
	// When a user-defined KeyFunc is provided, SigningKey, SigningKeys, and SigningMethod are ignored.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither SigningKeys nor SigningKey is provided.
	// Not used if custom ParseTokenFunc is set.
	// Default to an internal implementation verifying the signing algorithm and selecting the proper key.
	keyFunc jwt.Keyfunc

	// ParseTokenFunc defines a user-defined function that parses token from given auth. Returns an error when token
	// parsing fails or parsed token is invalid.
	// Defaults to implementation using `github.com/golang-jwt/jwt` as JWT implementation library
	parseTokenFunc func(auth string) (*jwt.Token, error)

	mock OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:        xid.New().String(),
		entryType:        "",
		signingKeys:      make(map[string]interface{}),
		signingAlgorithm: AlgorithmHS256,
		claims:           jwt.MapClaims{},
		tokenLookup:      "header:" + rkmid.HeaderAuthorization,
		authScheme:       "Bearer",
		ignorePrefix:     []string{},
	}

	set.keyFunc = set.defaultKeyFunc
	set.parseTokenFunc = set.defaultParseToken

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	sources := strings.Split(set.tokenLookup, ",")
	for _, source := range sources {
		parts := strings.Split(source, ":")

		switch parts[0] {
		case "query":
			set.extractors = append(set.extractors, jwtFromQuery(parts[1]))
		case "cookie":
			set.extractors = append(set.extractors, jwtFromCookie(parts[1]))
		case "form":
			set.extractors = append(set.extractors, jwtFromForm(parts[1]))
		case "header":
			set.extractors = append(set.extractors, jwtFromHeader(parts[1], set.authScheme))
		}
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
func (set *optionSet) BeforeCtx(req *http.Request, userCtx context.Context) *BeforeCtx {
	ctx := NewBeforeCtx()

	ctx.Input.Request = req
	ctx.Input.UserCtx = userCtx

	if req != nil && req.URL != nil {
		ctx.Input.UrlPath = req.URL.Path
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	if ctx == nil || set.ignore(ctx.Input.UrlPath) {
		return
	}

	// case 1: if user extractor exists, use it!
	if set.extractor != nil {
		// case 1.1: extract
		authRaw, err := set.extractor(ctx.Input.UserCtx)
		if err != nil {
			ctx.Output.ErrResp = errJwtInvalid
			return
		}

		// case 1.2: parse
		token, err := set.parseTokenFunc(authRaw)
		if err != nil {
			ctx.Output.ErrResp = errJwtInvalid
			return
		}

		ctx.Output.JwtToken = token
		return
	}

	// case 2: use default
	var authRaw string
	var err error
	var token *jwt.Token

	for _, extractor := range set.extractors {
		// Extract token from extractor, if it's not fail break the loop and
		// set auth
		authRaw, err = extractor(ctx.Input.Request)
		if err == nil {
			break
		}
	}

	if err != nil {
		ctx.Output.ErrResp = errJwtInvalid
		return
	}

	token, err = set.parseTokenFunc(authRaw)
	if err != nil {
		ctx.Output.ErrResp = errJwtInvalid
		return
	}

	ctx.Output.JwtToken = token
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

// Default key parsing func
func (set *optionSet) defaultKeyFunc(t *jwt.Token) (interface{}, error) {
	// check the signing method
	if t.Method.Alg() != set.signingAlgorithm {
		return nil, fmt.Errorf("unexpected jwt signing algorithm=%v", t.Header["alg"])
	}

	// check kid in token first
	// https://www.rfc-editor.org/rfc/rfc7515#section-4.1.4
	if len(set.signingKeys) > 0 {
		if kid, ok := t.Header["kid"].(string); ok {
			if key, ok := set.signingKeys[kid]; ok {
				return key, nil
			}
		}
		return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
	}

	// return signing key
	return set.signingKey, nil
}

// Default token parsing func
func (set *optionSet) defaultParseToken(auth string) (*jwt.Token, error) {
	token := new(jwt.Token)
	var err error

	// implementation of jwt.MapClaims
	if _, ok := set.claims.(jwt.MapClaims); ok {
		token, err = jwt.Parse(auth, set.keyFunc)
	} else {
		// custom implementation of jwt.Claims
		t := reflect.ValueOf(set.claims).Type().Elem()
		claims := reflect.New(t).Interface().(jwt.Claims)
		token, err = jwt.ParseWithClaims(auth, claims, set.keyFunc)
	}

	// return error
	if err != nil {
		return nil, err
	}

	// invalid token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
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
func (mock *optionSetMock) BeforeCtx(request *http.Request, userCtx context.Context) *BeforeCtx {
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
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		UrlPath string
		Request *http.Request
		UserCtx context.Context
	}
	Output struct {
		JwtToken *jwt.Token
		ErrResp  *rkerror.ErrorResp
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled      bool     `yaml:"enabled" json:"enabled"`
	IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
	SigningKey   string   `yaml:"signingKey" json:"signingKey"`
	SigningKeys  []string `yaml:"signingKeys" json:"signingKeys"`
	SigningAlgo  string   `yaml:"signingAlgo" json:"signingAlgo"`
	TokenLookup  string   `yaml:"tokenLookup" json:"tokenLookup"`
	AuthScheme   string   `yaml:"authScheme" json:"authScheme"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		var signingKey []byte
		if len(config.SigningKey) > 0 {
			signingKey = []byte(config.SigningKey)
		}

		opts = []Option{
			WithEntryNameAndType(entryName, entryType),
			WithSigningKey(signingKey),
			WithSigningAlgorithm(config.SigningAlgo),
			WithTokenLookup(config.TokenLookup),
			WithAuthScheme(config.AuthScheme),
			WithIgnorePrefix(config.IgnorePrefix...),
		}

		for _, v := range config.SigningKeys {
			tokens := strings.SplitN(v, ":", 2)
			if len(tokens) == 2 {
				opts = append(opts, WithSigningKeys(tokens[0], tokens[1]))
			}
		}
	}

	return opts
}

// ***************** Option *****************

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.entryName = entryName
		opt.entryType = entryType
	}
}

// WithSigningKey provide SigningKey.
func WithSigningKey(key interface{}) Option {
	return func(opt *optionSet) {
		if key != nil {
			opt.signingKey = key
		}
	}
}

// WithSigningKeys provide SigningKey with key and value.
func WithSigningKeys(key string, value interface{}) Option {
	return func(opt *optionSet) {
		if len(key) > 0 {
			opt.signingKeys[key] = value
		}
	}
}

// WithSigningAlgorithm provide signing algorithm.
// Default is HS256.
func WithSigningAlgorithm(algo string) Option {
	return func(opt *optionSet) {
		if len(algo) > 0 {
			opt.signingAlgorithm = algo
		}
	}
}

// WithClaims provide jwt.Claims.
func WithClaims(claims jwt.Claims) Option {
	return func(opt *optionSet) {
		opt.claims = claims
	}
}

// WithExtractor provide user extractor
func WithExtractor(ex JwtExtractor) Option {
	return func(opt *optionSet) {
		opt.extractor = ex
	}
}

// WithTokenLookup provide lookup configs.
// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
// to extract token from the request.
// Optional. Default value "header:Authorization".
// Possible values:
// - "header:<name>"
// - "query:<name>"
// - "param:<name>"
// - "cookie:<name>"
// - "form:<name>"
// Multiply sources example:
// - "header: Authorization,cookie: myowncookie"
func WithTokenLookup(lookup string) Option {
	return func(opt *optionSet) {
		if len(lookup) > 0 {
			opt.tokenLookup = lookup
		}
	}
}

// WithAuthScheme provide auth scheme.
// Default is Bearer
func WithAuthScheme(scheme string) Option {
	return func(opt *optionSet) {
		if len(scheme) > 0 {
			opt.authScheme = scheme
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

// WithKeyFunc provide user defined key func.
func WithKeyFunc(f jwt.Keyfunc) Option {
	return func(opt *optionSet) {
		opt.keyFunc = f
	}
}

// WithParseTokenFunc provide user defined token parse func.
func WithParseTokenFunc(f ParseTokenFunc) Option {
	return func(opt *optionSet) {
		opt.parseTokenFunc = f
	}
}

// WithMockOptionSet provide mock OptionSetInterface
func WithMockOptionSet(mock OptionSetInterface) Option {
	return func(set *optionSet) {
		set.mock = mock
	}
}

// ***************** Types *****************

// ParseTokenFunc parse token func
type ParseTokenFunc func(auth string) (*jwt.Token, error)

type JwtExtractor func(ctx context.Context) (string, error)

// jwt http extractor
type jwtHttpExtractor func(r *http.Request) (string, error)

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtHttpExtractor {
	return func(req *http.Request) (string, error) {
		if req == nil {
			return "", errJwtMissing.Err
		}

		auth := req.Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && strings.EqualFold(auth[:l], authScheme) {
			return auth[l+1:], nil
		}
		return "", errJwtMissing.Err
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(name string) jwtHttpExtractor {
	return func(req *http.Request) (string, error) {
		if req == nil || req.URL == nil {
			return "", errJwtMissing.Err
		}

		token := req.URL.Query().Get(name)
		if token == "" {
			return "", errJwtMissing.Err
		}
		return token, nil
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the named cookie.
func jwtFromCookie(name string) jwtHttpExtractor {
	return func(req *http.Request) (string, error) {
		if req == nil {
			return "", errJwtMissing.Err
		}

		cookie, err := req.Cookie(name)
		if err != nil {
			return "", errJwtMissing.Err
		}

		return url.QueryUnescape(cookie.Value)
	}
}

// jwtFromForm returns a `jwtExtractor` that extracts token from the form field.
func jwtFromForm(name string) jwtHttpExtractor {
	return func(req *http.Request) (string, error) {
		if req == nil {
			return "", errJwtMissing.Err
		}

		field := req.Form.Get(name)
		if field == "" {
			return "", errJwtMissing.Err
		}
		return field, nil
	}
}
