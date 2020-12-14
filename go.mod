module github.com/filipecosta90/opentelemetry-go-http-sample

go 1.13

require (
	github.com/go-redis/redis/extra/redisotel v0.2.0
	github.com/go-redis/redis/v8 v8.4.2
	github.com/golangci/golangci-lint v1.33.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/namsral/flag v1.7.4-pre
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.15.0
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
)
