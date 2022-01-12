package rkmid

import (
	"github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
)

const (
	HeaderAuthorization                   = "authorization"
	HeaderApiKey                          = "X-API-Key"
	HeaderRequestId                       = "X-Request-Id"
	HeaderTraceId                         = "X-Trace-Id"
	HeaderOrigin                          = "Origin"
	HeaderAccessControlAllowOrigin        = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowCredentials   = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders      = "Access-Control-Expose-Headers"
	HeaderVary                            = "Vary"
	HeaderAccessControlRequestMethod      = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders     = "Access-Control-Request-Headers"
	HeaderAccessControlAllowMethods       = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders       = "Access-Control-Allow-Headers"
	HeaderAccessControlMaxAge             = "Access-Control-Max-Age"
	HeaderContentEncoding                 = "Content-Encoding"
	HeaderContentLength                   = "Content-Length"
	HeaderContentType                     = "Content-Type"
	HeaderAcceptEncoding                  = "Accept-Encoding"
	HeaderXXSSProtection                  = "X-Xss-Protection"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderXForwardedProto                 = "X-Forwarded-Proto"
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderReferrerPolicy                  = "Referrer-Policy"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderCookie                          = "Cookie"
)

var (
	EntryNameKey      = &entryNameKey{}
	EntryTypeKey      = &entryTypeKey{}
	EventKey          = &eventKey{}
	LoggerKey         = &loggerKey{}
	TracerKey         = &tracerKey{}
	SpanKey           = &spanKey{}
	TracerProviderKey = &tracerProviderKey{}
	PropagatorKey     = &propagatorKey{}
	JwtTokenKey       = &jwtTokenKey{}
	CsrfTokenKey      = &csrfTokenKey{}
	// Realm environment variable
	Realm = zap.String("realm", rkcommon.GetEnvValueOrDefault("REALM", "*"))
	// Region environment variable
	Region = zap.String("region", rkcommon.GetEnvValueOrDefault("REGION", "*"))
	// AZ environment variable
	AZ = zap.String("az", rkcommon.GetEnvValueOrDefault("AZ", "*"))
	// Domain environment variable
	Domain = zap.String("domain", rkcommon.GetEnvValueOrDefault("DOMAIN", "*"))
	// LocalIp read local IP from localhost
	LocalIp = zap.String("localIp", rkcommon.GetLocalIP())
	// LocalHostname read hostname from localhost
	LocalHostname = zap.String("localHostname", rkcommon.GetLocalHostname())

	ignorePrefix = []string{
		"/rk/v1/assets",
	}

	IgnorePrefixGlobal = ignorePathPrefix
)

type entryNameKey struct{}

func (key *entryNameKey) String() string {
	return "entryNameKeyRk"
}

type entryTypeKey struct{}

func (key *entryTypeKey) String() string {
	return "entryTypeKeyRk"
}

type eventKey struct{}

func (key *eventKey) String() string {
	return "eventKeyRk"
}

type loggerKey struct{}

func (key *loggerKey) String() string {
	return "loggerKeyRk"
}

type tracerKey struct{}

func (key *tracerKey) String() string {
	return "tracerKeyRk"
}

type spanKey struct{}

func (key *spanKey) String() string {
	return "spanKeyRk"
}

type tracerProviderKey struct{}

func (key *tracerProviderKey) String() string {
	return "tracerProviderKeyRk"
}

type propagatorKey struct{}

func (key *propagatorKey) String() string {
	return "propagatorKeyRk"
}

type jwtTokenKey struct{}

func (key *jwtTokenKey) String() string {
	return "jwtTokenKeyRk"
}

type csrfTokenKey struct{}

func (key *csrfTokenKey) String() string {
	return "csrfTokenKeyRk"
}

// GetRemoteAddressSet returns remote endpoint information set including IP, Port.
// We will do as best as we can to determine it.
// If fails, then just return default ones.
func GetRemoteAddressSet(req *http.Request) (remoteIp, remotePort string) {
	remoteIp, remotePort = "0.0.0.0", "0"

	if req == nil {
		return
	}

	var err error
	if remoteIp, remotePort, err = net.SplitHostPort(req.RemoteAddr); err != nil {
		return
	}

	forwardedRemoteIp := req.Header.Get("x-forwarded-for")

	// Deal with forwarded remote ip
	if len(forwardedRemoteIp) > 0 {
		if forwardedRemoteIp == "::1" {
			forwardedRemoteIp = "localhost"
		}

		remoteIp = forwardedRemoteIp
	}

	if remoteIp == "::1" {
		remoteIp = "localhost"
	}

	return remoteIp, remotePort
}

func ignorePathPrefix(urlPath string) bool {
	for i := range ignorePrefix {
		if strings.HasPrefix(urlPath, ignorePrefix[i]) {
			return true
		}
	}

	return false
}
