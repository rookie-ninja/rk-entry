// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidjwt is a middleware for JWT
package rkmidjwt

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/error"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	errJwtMissing = rkmid.GetErrorBuilder().New(http.StatusBadRequest, "Missing or malformed jwt")
	errJwtInvalid = rkmid.GetErrorBuilder().New(http.StatusUnauthorized, "Invalid or expired jwt")
)

// ***************** OptionSet Interface *****************

// OptionSetInterface mainly for testing purpose
type OptionSetInterface interface {
	GetEntryName() string

	GetEntryType() string

	Before(*BeforeCtx)

	BeforeCtx(*http.Request, context.Context) *BeforeCtx

	ShouldIgnore(string) bool
}

// ***************** OptionSet Implementation *****************

// optionSet which is used for middleware implementation
type optionSet struct {
	// name of entry
	entryName string

	// type of entry
	entryType string

	// path to ignore
	pathToIgnore []string

	// extractors of JWT token from http.Request
	extractors []jwtHttpExtractor

	// user provided extractor
	extractor JwtExtractor

	// implementation of rkentry.SignerJwt
	signer rkentry.SignerJwt

	// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// Multiply sources example:
	// - "header: Authorization,cookie: myowncookie"
	tokenLookup string

	// AuthScheme to be used in the Authorization header.
	// Optional. Default value "Bearer".
	authScheme string

	// skip the step of validate token,just parse.
	// Optional. Default value "false".
	skipVerify bool

	mock OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:    "fake-entry",
		entryType:    "",
		tokenLookup:  "header:" + rkmid.HeaderAuthorization,
		authScheme:   "Bearer",
		pathToIgnore: []string{},
	}

	for i := range opts {
		opts[i](set)
	}

	if set.signer == nil && !set.skipVerify {
		set.signer = rkentry.RegisterSymmetricJwtSigner(set.entryName, jwt.SigningMethodHS256.Name, []byte("rk jwt key"))
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
		case "header":
			set.extractors = append(set.extractors, jwtFromHeader(parts[1], set.authScheme))
		case "cookie":
			set.extractors = append(set.extractors, jwtFromCookie(parts[1]))
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
	if ctx == nil || set.ShouldIgnore(ctx.Input.UrlPath) {
		return
	}
	var authRaw string
	var err error
	var token *jwt.Token

	if set.extractor != nil { // case 1: if user extractor exists, use it!
		authRaw, err = set.extractor(ctx.Input.UserCtx)
		if err != nil {
			ctx.Output.ErrResp = errJwtInvalid
			return
		}

	} else { // case 2: use default
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
	}
	if set.skipVerify || set.signer == nil {
		// case 1: when skip validate or disable sign, just parse token
		claims := jwt.MapClaims{}
		parser := &jwt.Parser{}
		token, _, err = parser.ParseUnverified(authRaw, claims)
	} else {
		// case 2: parse and validate token
		token, err = set.signer.VerifyJwt(authRaw)
	}

	if err != nil {
		ctx.Output.ErrResp = errJwtInvalid
		return
	}

	ctx.Output.JwtToken = token
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
func (mock *optionSetMock) BeforeCtx(request *http.Request, userCtx context.Context) *BeforeCtx {
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
		ErrResp  rkerror.ErrorInterface
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	Ignore      []string          `yaml:"ignore" json:"ignore"`
	SignerEntry string            `yaml:"signerEntry" json:"signerEntry"`
	Symmetric   *SymmetricConfig  `yaml:"symmetric" json:"symmetric"`
	Asymmetric  *AsymmetricConfig `yaml:"asymmetric" json:"asymmetric"`
	TokenLookup string            `yaml:"tokenLookup" json:"tokenLookup"`
	AuthScheme  string            `yaml:"authScheme" json:"authScheme"`
	SkipVerify  bool              `yaml:"skipVerify" json:"skipVerify"`
}

type SymmetricConfig struct {
	Algorithm string `yaml:"algorithm" json:"algorithm"`
	Token     string `yaml:"token" json:"token"`
	TokenPath string `yaml:"tokenPath" json:"tokenPath"`
}

type AsymmetricConfig struct {
	Algorithm      string `yaml:"algorithm" json:"algorithm"`
	PrivateKey     string `yaml:"privateKey" json:"privateKey"`
	PrivateKeyPath string `yaml:"privateKeyPath" json:"privateKeyPath"`
	PublicKey      string `yaml:"publicKey" json:"publicKey"`
	PublicKeyPath  string `yaml:"publicKeyPath" json:"publicKeyPath"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, entryName, entryType string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		var signerJwt rkentry.SignerJwt

		// check signer entry first
		if v := rkentry.GlobalAppCtx.GetEntry(rkentry.SignerJwtEntryType, config.SignerEntry); v != nil {
			signer, ok := v.(rkentry.SignerJwt)
			if !ok {
				rkentry.ShutdownWithError(errors.New("invalid signer jwt entry"))
			}

			signerJwt = signer
			if signerJwt == nil {
				rkentry.ShutdownWithError(errors.New("cannot find signer entry"))
			}
		} else if config.Asymmetric != nil {
			var pubKey, privKey []byte

			if len(config.Asymmetric.PublicKey) > 0 {
				pubKey = []byte(config.Asymmetric.PublicKey)
			} else {
				pubKey = mustRead(config.Asymmetric.PublicKeyPath)
			}

			if len(config.Asymmetric.PrivateKey) > 0 {
				privKey = []byte(config.Asymmetric.PrivateKey)
			} else {
				privKey = mustRead(config.Asymmetric.PrivateKeyPath)
			}

			signerJwt = rkentry.RegisterAsymmetricJwtSigner(entryName, config.Asymmetric.Algorithm, privKey, pubKey)
			if signerJwt == nil {
				rkentry.ShutdownWithError(errors.New("invalid asymmetric configuration"))
			}
		} else if config.Symmetric != nil {
			var token []byte
			if len(config.Symmetric.Token) > 0 {
				token = []byte(config.Symmetric.Token)
			} else {
				token = mustRead(config.Symmetric.TokenPath)
			}

			signerJwt = rkentry.RegisterSymmetricJwtSigner(entryName, config.Symmetric.Algorithm, token)
			if signerJwt == nil {
				rkentry.ShutdownWithError(errors.New("invalid symmetric configuration"))
			}
		}

		opts = []Option{
			WithEntryNameAndType(entryName, entryType),
			WithTokenLookup(config.TokenLookup),
			WithSigner(signerJwt),
			WithAuthScheme(config.AuthScheme),
			WithPathToIgnore(config.Ignore...),
			WithSkipVerify(config.SkipVerify),
		}

	}

	return opts
}

func mustRead(p string) []byte {
	if !filepath.IsAbs(p) {
		wd, _ := os.Getwd()
		p = filepath.Join(wd, p)
	}

	res, err := os.ReadFile(p)
	if err != nil {
		rkentry.ShutdownWithError(err)
	}

	return res
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

// WithSigner provide rkentry.SignerJwt.
func WithSigner(signer rkentry.SignerJwt) Option {
	return func(opt *optionSet) {
		opt.signer = signer
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
// - "cookie:<name>"
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

// WithSkipVerify provide skipVerify
// Default is false
func WithSkipVerify(skipVerify bool) Option {
	return func(opt *optionSet) {
		opt.skipVerify = skipVerify
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
			return "", errJwtMissing
		}

		auth := req.Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && strings.EqualFold(auth[:l], authScheme) {
			return auth[l+1:], nil
		}
		return "", errJwtMissing
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(name string) jwtHttpExtractor {
	return func(req *http.Request) (string, error) {
		if req == nil || req.URL == nil {
			return "", errJwtMissing
		}

		token := req.URL.Query().Get(name)
		if token == "" {
			return "", errJwtMissing
		}
		return token, nil
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the cookie.
func jwtFromCookie(name string) jwtHttpExtractor {
	return func(req *http.Request) (string, error) {
		if req == nil || req.URL == nil {
			return "", errJwtMissing
		}

		cookie, err := req.Cookie(name)
		if err != nil {
			return "", err
		}

		return cookie.Value, nil
	}
}
