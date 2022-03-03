package rkmid

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"regexp"
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
	Realm = zap.String("realm", getEnvValueOrDefault("REALM", "*"))
	// Region environment variable
	Region = zap.String("region", getEnvValueOrDefault("REGION", "*"))
	// AZ environment variable
	AZ = zap.String("az", getEnvValueOrDefault("AZ", "*"))
	// Domain environment variable
	Domain = zap.String("domain", getEnvValueOrDefault("DOMAIN", "*"))
	// LocalIp read local IP from localhost
	LocalIp = zap.String("localIp", getLocalIP())
	// LocalHostname read hostname from localhost
	LocalHostname = zap.String("localHostname", getLocalHostname())

	ignorePrefix = []string{}

	IgnorePrefixGlobal = ignorePathPrefix
)

func AddIgnorePrefixGlobal(prefix ...string) {
	ignorePrefix = append(ignorePrefix, prefix...)
}

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

// GenerateRequestId generate request id based on google/uuid.
// UUIDs are based on RFC 4122 and DCE 1.1: Authentication and Security Services.
//
// A UUID is a 16 byte (128 bit) array. UUIDs may be used as keys to maps or compared directly.
func GenerateRequestId() string {
	// do not use uuid.New() since it would panic if any error occurs
	requestId, err := uuid.NewRandom()

	// currently, we will return empty string if error occurs
	if err != nil {
		return ""
	}

	return requestId.String()
}

// GenerateRequestIdWithPrefix generate request id based on google/uuid.
// UUIDs are based on RFC 4122 and DCE 1.1: Authentication and Security Services.
//
// A UUID is a 16 byte (128 bit) array. UUIDs may be used as keys to maps or compared directly.
func GenerateRequestIdWithPrefix(prefix string) string {
	// Do not use uuid.New() since it would panic if any error occurs
	requestId, err := uuid.NewRandom()

	// Currently, we will return empty string if error occurs
	if err != nil {
		return ""
	}

	if len(prefix) > 0 {
		return prefix + "-" + requestId.String()
	}

	return requestId.String()
}

// getEnvValueOrDefault returns default value if environment variable is empty or not exist.
func getEnvValueOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)

	if len(value) < 1 {
		return defaultValue
	}

	return value
}

// getLocalHostname returns hostname of localhost, return "" if error occurs or hostname is empty.
func getLocalHostname() string {
	hostname, err := os.Hostname()
	if err != nil || len(hostname) < 1 {
		hostname = ""
	}

	return hostname
}

// getLocalIP
// This is a tricky function.
// We will iterate through all the network interfaces，but will choose the first one since we are assuming that
// eth0 will be the default one to use in most of the case.
//
// Currently, we do not have any interfaces for selecting the network interface yet.
func getLocalIP() string {
	localIP := "localhost"

	// skip the error since we don't want to break RPC calls because of it
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return localIP
	}

	for _, addr := range addresses {
		items := strings.Split(addr.String(), "/")
		if len(items) < 2 || items[0] == "127.0.0.1" {
			continue
		}

		if match, err := regexp.MatchString(`\d+\.\d+\.\d+\.\d+`, items[0]); err == nil && match {
			localIP = items[0]
		}
	}

	return localIP
}
