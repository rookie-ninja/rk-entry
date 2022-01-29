module github.com/rookie-ninja/rk-entry

go 1.15

require (
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/grantae/certinfo v0.0.0-20170412194111-59d56a35515b
	github.com/hako/durafmt v0.0.0-20200710122514-c0fb7b4da026
	github.com/hashicorp/consul/api v1.8.1
	github.com/juju/ratelimit v1.0.1
	github.com/markbates/pkger v0.17.1
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/client_model v0.2.0
	github.com/rookie-ninja/rk-common v1.2.3
	github.com/rookie-ninja/rk-logger v1.2.7
	github.com/rookie-ninja/rk-prom v1.1.4
	github.com/rookie-ninja/rk-query v1.2.7
	github.com/rs/xid v1.3.0
	github.com/shirou/gopsutil/v3 v3.21.4
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	go.etcd.io/etcd/client/v3 v3.5.0-alpha.0
	go.opentelemetry.io/contrib v1.3.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/exporters/jaeger v1.3.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.3.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
	go.uber.org/atomic v1.7.0
	go.uber.org/ratelimit v0.2.0
	go.uber.org/zap v1.20.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
