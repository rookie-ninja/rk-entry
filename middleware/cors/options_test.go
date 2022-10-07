package cors

import (
	"github.com/rookie-ninja/rk-entry/v3/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToOptions(t *testing.T) {
	config := &BootConfig{
		Enabled: false,
	}

	// with disable
	assert.Empty(t, ToOptions(config, "", ""))

	// with enabled
	config.Enabled = true
	assert.NotEmpty(t, ToOptions(config, "", ""))
}

func TestNewOptionSet(t *testing.T) {
	// without options
	set := NewOptionSet().(*optionSet)
	assert.NotEmpty(t, set.EntryName())
	assert.Empty(t, set.pathToIgnore)
	assert.Contains(t, set.allowOrigins, "*")
	assert.NotEmpty(t, set.allowMethods)
	assert.Empty(t, set.allowHeaders)
	assert.False(t, set.allowCredentials)
	assert.Empty(t, set.exposeHeaders)
	assert.Zero(t, set.maxAge)

	// with options
	set = NewOptionSet(
		WithEntryNameAndKind("name", "kind"),
		WithAllowOrigins("localhost:*"),
		WithAllowMethods(http.MethodGet),
		WithAllowHeaders("ut-header"),
		WithAllowCredentials(true),
		WithExposeHeaders("ut-header"),
		WithMaxAge(1),
		WithPathToIgnore("/ut-path")).(*optionSet)

	assert.Equal(t, "name", set.EntryName())
	assert.Equal(t, "type", set.EntryKind())
	assert.Contains(t, set.pathToIgnore, "/ut-path")
	assert.Contains(t, set.allowOrigins, "localhost:*")
	assert.Contains(t, set.allowMethods, http.MethodGet)
	assert.Contains(t, set.allowHeaders, "ut-header")
	assert.True(t, set.allowCredentials)
	assert.Contains(t, set.exposeHeaders, "ut-header")
	assert.Equal(t, 1, set.maxAge)
}

func TestOptionSet_BeforeCtx(t *testing.T) {
	// with nil req
	set := NewOptionSet()
	assert.NotNil(t, set.BeforeCtx(nil))

	// happy case
	req := httptest.NewRequest(http.MethodOptions, "/ut", nil)
	req.Header.Set(rkm.HeaderOrigin, "ut-origin")
	req.Header.Set(rkm.HeaderAccessControlRequestHeaders, "ut-header")
	ctx := set.BeforeCtx(req)
	assert.Equal(t, "/ut", ctx.Input.UrlPath)
	assert.Equal(t, "ut-origin", ctx.Input.OriginHeader)
	assert.Equal(t, "ut-header", ctx.Input.AccessControlRequestHeaders)
	assert.True(t, ctx.Input.IsPreflight)
}

func TestOptionSet_isOriginAllowed(t *testing.T) {
	set := NewOptionSet().(*optionSet)

	// 1: wildcard
	set.allowOrigins = []string{"*"}
	set.toPatterns()
	assert.True(t, set.isOriginAllowed("http://ut.domain"))

	// 2: exact matching
	set.allowOrigins = []string{"http://ut.domain"}
	set.toPatterns()
	assert.True(t, set.isOriginAllowed("http://ut.domain"))
	assert.False(t, set.isOriginAllowed("http://ut.another"))

	// 3: subdomain
	set.allowOrigins = []string{"http://*.ut.domain"}
	set.toPatterns()
	assert.True(t, set.isOriginAllowed("http://sub.ut.domain"))
	assert.True(t, set.isOriginAllowed("http://sub.sub.ut.domain"))
	assert.False(t, set.isOriginAllowed("http://ut.domain"))
	assert.False(t, set.isOriginAllowed("http://ut.another"))

	// 4: wildcard in middle of domain
	set.allowOrigins = []string{"http://ut.*.domain"}
	set.toPatterns()
	assert.True(t, set.isOriginAllowed("http://ut.sub.domain"))
	assert.True(t, set.isOriginAllowed("http://ut.sub.sub.domain"))
	assert.False(t, set.isOriginAllowed("http://ut.domain"))
	assert.False(t, set.isOriginAllowed("http://ut.another"))

	// 5: wildcard in the last
	set.allowOrigins = []string{"http://ut.domain.*"}
	set.toPatterns()
	assert.True(t, set.isOriginAllowed("http://ut.domain.sub"))
	assert.True(t, set.isOriginAllowed("http://ut.domain.sub.sub"))
	assert.False(t, set.isOriginAllowed("http://ut.domain"))
	assert.False(t, set.isOriginAllowed("http://ut.another"))
}

func newReq(method string, headers ...header) *http.Request {
	req := httptest.NewRequest(method, "/ut", nil)
	for _, h := range headers {
		req.Header.Set(h.Key, h.Value)
	}

	return req
}

type header struct {
	Key   string
	Value string
}

func TestOptionSet_Before(t *testing.T) {
	originHeaderValue := "http://ut-origin"
	set := NewOptionSet()

	// with empty option, all request will be passed
	req := newReq(http.MethodGet, header{rkm.HeaderOrigin, originHeaderValue})
	ctx := set.BeforeCtx(req)
	set.Before(ctx)
	assert.False(t, ctx.Output.Abort)

	// match 1.1
	req = newReq(http.MethodGet)
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.False(t, ctx.Output.Abort)

	// match 1.2
	req = newReq(http.MethodOptions)
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.True(t, ctx.Output.Abort)

	// match 2
	set = NewOptionSet(WithAllowOrigins("http://do-not-pass-through"))
	req = newReq(http.MethodGet, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.True(t, ctx.Output.Abort)

	// match 3
	set = NewOptionSet()
	req = newReq(http.MethodGet, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.False(t, ctx.Output.Abort)
	assert.Equal(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowOrigin])

	// match 3.1
	set = NewOptionSet(WithAllowCredentials(true))
	req = newReq(http.MethodGet, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.False(t, ctx.Output.Abort)
	assert.Equal(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowOrigin])
	assert.Equal(t, "true", ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowCredentials])

	// match 3.2
	set = NewOptionSet(WithAllowCredentials(true), WithExposeHeaders("expose"))
	req = newReq(http.MethodGet, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.False(t, ctx.Output.Abort)
	assert.Equal(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowOrigin])
	assert.Equal(t, "true", ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowCredentials])
	assert.Equal(t, "expose", ctx.Output.HeadersToReturn[rkm.HeaderAccessControlExposeHeaders])

	// match 4
	set = NewOptionSet()
	req = newReq(http.MethodOptions, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.True(t, ctx.Output.Abort)
	assert.Len(t, ctx.Output.HeaderVary, 2)
	assert.Equal(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowOrigin])

	// match 4.1
	set = NewOptionSet(WithAllowCredentials(true))
	req = newReq(http.MethodOptions, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.True(t, ctx.Output.Abort)
	assert.Len(t, ctx.Output.HeaderVary, 2)
	assert.Equal(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowOrigin])
	assert.NotEmpty(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowMethods])
	assert.Equal(t, "true", ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowCredentials])

	// match 4.2
	set = NewOptionSet(WithAllowHeaders("ut-header"))
	req = newReq(http.MethodOptions, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.True(t, ctx.Output.Abort)
	assert.Len(t, ctx.Output.HeaderVary, 2)
	assert.Equal(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowOrigin])
	assert.NotEmpty(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowMethods])
	assert.Equal(t, "ut-header", ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowHeaders])

	// match 4.3
	set = NewOptionSet(WithMaxAge(1))
	req = newReq(http.MethodOptions, header{rkm.HeaderOrigin, originHeaderValue})
	ctx = set.BeforeCtx(req)
	set.Before(ctx)
	assert.True(t, ctx.Output.Abort)
	assert.Len(t, ctx.Output.HeaderVary, 2)
	assert.Equal(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowOrigin])
	assert.NotEmpty(t, originHeaderValue, ctx.Output.HeadersToReturn[rkm.HeaderAccessControlAllowMethods])
	assert.Equal(t, "1", ctx.Output.HeadersToReturn[rkm.HeaderAccessControlMaxAge])
}

func TestNewOptionSetMock(t *testing.T) {
	mock := NewOptionSetMock(NewBeforeCtx())
	assert.NotEmpty(t, mock.EntryName())
	assert.NotEmpty(t, mock.EntryKind())
	assert.NotNil(t, mock.BeforeCtx(nil))
	mock.Before(nil)
}
