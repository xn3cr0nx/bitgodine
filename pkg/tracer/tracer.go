package tracer

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	sdkexporter "go.opentelemetry.io/otel/sdk/export/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracer *trace.Tracer

// Exporter list of exporter options
type Exporter int

const (
	// Stdout exporter to log tracing to stdout
	Stdout Exporter = iota
	// Jaeger export to send tracing to jaeger
	Jaeger
)

// Config tracer configuration
type Config struct {
	Name     string
	Exporter Exporter
}

// NewTracer singleton implementation returns default tracer
func NewTracer(conf *Config) (*trace.Tracer, echo.MiddlewareFunc, error) {
	if tracer == nil {
		t := otel.Tracer(conf.Name)
		tracer = &t

		if err := configure(conf); err != nil {
			return nil, nil, err
		}
	}

	middleware := otelecho.Middleware(conf.Name)
	return tracer, middleware, nil
}

func configure(conf *Config) (err error) {
	var exporter sdkexporter.SpanExporter

	switch conf.Exporter {
	case Jaeger:
		exporter, err = jaeger.NewRawExporter(jaeger.WithCollectorEndpoint("http://jaeger:14268/api/traces"),
			jaeger.WithProcess(jaeger.Process{
				ServiceName: conf.Name,
				Tags: []label.KeyValue{
					label.String("exporter", "jaeger"),
					label.Float64("test", 312.23),
				},
			}),
			jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		)
		if err != nil {
			return
		}
	default:
		exporter, err = stdout.NewExporter(stdout.WithPrettyPrint())
		if err != nil {
			return
		}
	}

	cfg := sdktrace.Config{
		DefaultSampler: sdktrace.AlwaysSample(),
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(cfg),
		sdktrace.WithSyncer(exporter),
	)
	if err != nil {
		return
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return
}
