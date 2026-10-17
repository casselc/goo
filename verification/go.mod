module github.com/casselc/goo/verification

go 1.27.0

require (
	github.com/casselc/goo v0.0.0
	github.com/casselc/goo/instrumentation v0.0.0
	github.com/casselc/observability/quintgo v0.0.0
	go.opentelemetry.io/otel v1.46.0
	go.opentelemetry.io/otel/sdk/metric v1.46.0
	hegel.dev/go/hegel v0.9.8
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/ebitengine/purego v0.11.0-alpha.6.0.20260707033313-5f49e7c49322 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/casselc/goo => ..

replace github.com/casselc/observability/quintgo => ../../observability/quintgo

replace github.com/casselc/goo/instrumentation => ../instrumentation
