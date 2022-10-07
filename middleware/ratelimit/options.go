// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package limit provide options
package limit

import (
	"errors"
	"github.com/rookie-ninja/rk-entry/v3/error"
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	uber "go.uber.org/ratelimit"
	"net/http"
	"strings"
)

const (
	LeakyBucket   = "leakyBucket"
	DefaultLimit  = 1000000
	GlobalLimiter = "rk-limiter"
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
	entryName       string
	entryKind       string
	reqPerSec       int
	reqPerSecByPath map[string]int
	algorithm       string
	pathToIgnore    []string
	limiter         map[string]Limiter
	mock            OptionSetInterface
}

// NewOptionSet Create new optionSet with options.
func NewOptionSet(opts ...Option) OptionSetInterface {
	set := &optionSet{
		entryName:       "fake-entry",
		entryKind:       "",
		reqPerSec:       DefaultLimit,
		reqPerSecByPath: make(map[string]int),
		algorithm:       LeakyBucket,
		limiter:         make(map[string]Limiter),
		pathToIgnore:    []string{},
	}

	for i := range opts {
		opts[i](set)
	}

	if set.mock != nil {
		return set.mock
	}

	switch set.algorithm {
	case LeakyBucket:
		if set.reqPerSec < 1 {
			l := &ZeroRateLimiter{}
			set.setLimiter(GlobalLimiter, l.Limit)
		} else {
			l := &leakyBucketLimiter{
				delegator: uber.New(set.reqPerSec),
			}
			set.setLimiter(GlobalLimiter, l.Limit)
		}

		for k, v := range set.reqPerSecByPath {
			if v < 1 {
				l := &ZeroRateLimiter{}
				set.setLimiter(k, l.Limit)
			} else {
				l := &leakyBucketLimiter{
					delegator: uber.New(v),
				}
				set.setLimiter(k, l.Limit)
			}
		}
	default:
		l := &NoopLimiter{}
		set.setLimiter(GlobalLimiter, l.Limit)
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

	if req != nil && req.URL != nil {
		ctx.Input.UrlPath = req.URL.Path
	}

	return ctx
}

// Before should run before user handler
func (set *optionSet) Before(ctx *BeforeCtx) {
	if ctx == nil {
		return
	}

	// case 0: ignore path
	if set.ShouldIgnore(ctx.Input.UrlPath) {
		return
	}

	limiter := set.getLimiter(ctx.Input.UrlPath)
	if err := limiter(); err != nil {
		ctx.Output.ErrResp = rkm.GetErrorBuilder().New(http.StatusTooManyRequests, err.Error())
		return
	}

	return
}

func (set *optionSet) getLimiter(method string) Limiter {
	if v, ok := set.limiter[method]; ok {
		return v
	}

	return set.limiter[GlobalLimiter]
}

// Set limiter if not exists
func (set *optionSet) setLimiter(method string, l Limiter) {
	if _, ok := set.limiter[method]; ok {
		return
	}

	set.limiter[method] = l
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
	return ctx
}

// BeforeCtx context for Before() function
type BeforeCtx struct {
	Input struct {
		UrlPath string
	}
	Output struct {
		ErrResp rkerror.ErrorInterface
	}
}

// ***************** BootConfig *****************

// BootConfig for YAML
type BootConfig struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	Ignore    []string `yaml:"ignore" json:"ignore"`
	Algorithm string   `yaml:"algorithm" json:"algorithm"`
	ReqPerSec int      `yaml:"reqPerSec" json:"reqPerSec"`
	Paths     []struct {
		Path      string `yaml:"path" json:"path"`
		ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
	} `yaml:"paths" json:"paths"`
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *BootConfig, name, kind string) []Option {
	opts := make([]Option, 0)

	if config.Enabled {
		opts = append(opts, WithEntryNameAndKind(name, kind))

		if len(config.Algorithm) > 0 {
			opts = append(opts, WithAlgorithm(config.Algorithm))
		}

		opts = append(opts, WithReqPerSec(config.ReqPerSec))

		for i := range config.Paths {
			e := config.Paths[i]
			opts = append(opts, WithReqPerSecByPath(e.Path, e.ReqPerSec))
		}

		opts = append(opts, WithPathToIgnore(config.Ignore...))
	}

	return opts
}

// ***************** Option *****************

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndKind provide entry name and entry kind.
func WithEntryNameAndKind(name, kind string) Option {
	return func(opt *optionSet) {
		opt.entryName = name
		opt.entryKind = kind
	}
}

// WithReqPerSec Provide request per second.
func WithReqPerSec(reqPerSec int) Option {
	return func(opt *optionSet) {
		if reqPerSec >= 0 {
			opt.reqPerSec = reqPerSec
		}
	}
}

// WithReqPerSecByPath Provide request per second by method.
func WithReqPerSecByPath(path string, reqPerSec int) Option {
	return func(opt *optionSet) {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if reqPerSec >= 0 {
			opt.reqPerSecByPath[path] = reqPerSec
		}
	}
}

// WithAlgorithm provide algorithm of rate limit.
// - leakyBucket
func WithAlgorithm(algo string) Option {
	return func(opt *optionSet) {
		opt.algorithm = algo
	}
}

// WithGlobalLimiter provide user defined Limiter.
func WithGlobalLimiter(l Limiter) Option {
	return func(opt *optionSet) {
		if l != nil {
			opt.limiter[GlobalLimiter] = l
		}
	}
}

// WithLimiterByPath provide user defined Limiter by method.
func WithLimiterByPath(path string, l Limiter) Option {
	return func(opt *optionSet) {
		if l == nil {
			return
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		opt.limiter[path] = l
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

// ***************** Limiter *****************

// Limiter User could implement it
type Limiter func() error

// NoopLimiter will do nothing
type NoopLimiter struct{}

// Limit will do nothing
func (l *NoopLimiter) Limit() error {
	return nil
}

// ZeroRateLimiter will block requests.
type ZeroRateLimiter struct{}

// Limit will block request and return error
func (l *ZeroRateLimiter) Limit() error {
	return errors.New("slow down your request")
}

// leakyBucketLimiter delegates limit logic to uber.Limiter
type leakyBucketLimiter struct {
	delegator uber.Limiter
}

// Limit delegates limit logic to uber.Limiter
func (l *leakyBucketLimiter) Limit() error {
	l.delegator.Take()
	return nil
}
