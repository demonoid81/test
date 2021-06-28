package otelgqlgen

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
)

const (
	tracerName = "otelgqlgen"
)

var _ interface {
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
	graphql.HandlerExtension
} = Tracer{}

// Tracer is an extension for gqlgen to trace GrapGQL operations
type Tracer struct {
	cfg Config
}

// NewTracer returns a new Tracer
func NewTracer(options ...Option) Tracer {
	return Tracer{
		cfg: getConfig(options...),
	}
}

func (a Tracer) tracer() trace.Tracer {
	if a.cfg.Tracer == nil {
		return otel.Tracer(tracerName)
	}
	return a.cfg.Tracer
}

// ExtensionName implements graphql.HandlerExtension
func (a Tracer) ExtensionName() string {
	return "OpenTelemetryTracer"
}

// Validate implements graphql.HandlerExtension
func (a Tracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptResponse implements graphql.ResponseInterceptor
func (a Tracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) (r *graphql.Response) {
	octx := graphql.GetOperationContext(ctx)

	name := octx.OperationName
	if name == "" {
		if octx.Operation != nil {
			name = string(octx.Operation.Operation)
		} else {
			name = octx.RawQuery
		}
	}

	ctx, span := a.tracer().Start(ctx, name, trace.WithSpanKind(trace.SpanKindUnspecified))
	defer span.End()

	if !span.IsRecording() {
		return next(ctx)
	}

	span.SetAttributes(label.String("graphql.query", octx.RawQuery))
	for name, v := range octx.Variables {
		span.SetAttributes(label.String("graphql.vars."+name, fmt.Sprintf("%+v", v)))
	}
	if stats := extension.GetComplexityStats(ctx); stats != nil {
		span.SetAttributes(
			label.Int("graphql.complexity.value", stats.Complexity),
			label.Int("graphql.complexity.limit", stats.ComplexityLimit),
		)
	}

	defer func() {
		if r.Errors != nil {
			span.SetStatus(1, r.Errors.Error())
			return
		}
		span.SetStatus(2, "")
	}()

	return next(ctx)
}

// InterceptField implements graphql.FieldInterceptor
func (a Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	fc := graphql.GetFieldContext(ctx)

	name := fc.Field.Name
	if object := fc.Field.ObjectDefinition; object != nil {
		name = object.Name + "." + name
	}

	ctx, span := a.tracer().Start(ctx, name, trace.WithSpanKind(trace.SpanKindUnspecified))
	defer span.End()

	if !span.IsRecording() {
		return next(ctx)
	}

	span.SetAttributes(
		label.String("graphql.field.path", fc.Path().String()),
		label.String("graphql.field.name", fc.Field.Name),
		label.String("graphql.field.alias", fc.Field.Alias),
	)
	if object := fc.Field.ObjectDefinition; object != nil {
		span.SetAttributes(
			label.String("graphql.field.object", object.Name),
		)
	}
	for _, arg := range fc.Field.Arguments {
		span.SetAttributes(
			label.String("graphql.field.args."+arg.Name, arg.Value.String()),
		)
	}

	defer func() {
		if errs := graphql.GetFieldErrors(ctx, fc); errs != nil {
			span.SetStatus(1, errs.Error())
			return
		}
		span.SetStatus(2, "")
	}()

	return next(ctx)
}
