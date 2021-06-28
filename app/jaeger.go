package app

import (
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func (app *App) InitTraceProvider() (trace.TracerProvider, func()) {
	// Create and install Jaeger export pipeline
	tp, flush, err := jaeger.NewExportPipeline(
		// "http://localhost:14268/api/traces"
		jaeger.WithCollectorEndpoint(app.Cfg.Jaeger.Endpoint,
			jaeger.WithUsername(app.Cfg.Jaeger.Username),
			jaeger.WithPassword(app.Cfg.Jaeger.Password),
		),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: app.Cfg.NodeName,
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		app.Logger.Fatal().Err(err).Msg("")
	}
	return tp, func() {
		flush()
	}
}
