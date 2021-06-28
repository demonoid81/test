package otelgqlgen

import "go.opentelemetry.io/otel/trace"

// Config is used to configure the extension.
type Config struct {
	Tracer trace.Tracer
}

// Option specifies instrumentation configuration options.
type Option func(Config) Config

func getConfig(options ...Option) (c Config) {
	for _, f := range options {
		c = f(c)
	}
	return c
}

// WithTracer specifies a tracer to use for creating spans. If none is
// specified, a tracer named
// "github.com/rot1024/otelgqlgen"
// from the global provider is used.
func WithTracer(tracer trace.Tracer) Option {
	return func(o Config) Config {
		o.Tracer = tracer
		return o
	}
}
