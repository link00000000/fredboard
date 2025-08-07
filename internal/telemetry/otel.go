package telemetry

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	Tracer trace.Tracer
	Meter  metric.Meter
	Logger *slog.Logger
)

func SetupOTelSDK(name string, ctx context.Context) (shutdown func(context.Context) error, err error) {
	// OTel callbacks to notify that the OTelSDK is shutting down
	var shutdownCallbacks []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error

		for _, fn := range shutdownCallbacks {
			err = errors.Join(err, fn(ctx))
		}

		shutdownCallbacks = nil
		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	r := newResource(ctx)

	// Set up trace provider.
	traceExporter := newTraceExporter(ctx)
	tracerProvider := newTracerProvider(traceExporter, r)
	shutdownCallbacks = append(shutdownCallbacks, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterExporter := newMeterExporter(ctx)
	meterProvider := newMeterProvider(meterExporter, r)
	shutdownCallbacks = append(shutdownCallbacks, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerExporter := newLoggerExporter(ctx)
	loggerProvider := newLoggerProvider(loggerExporter, r)
	shutdownCallbacks = append(shutdownCallbacks, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	Tracer = otel.Tracer(name)
	Meter = otel.Meter(name)
	Logger = otelslog.NewLogger(name)

	setupMetrics(Meter)

	return
}

func newResource(ctx context.Context) *resource.Resource {
	r, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("fredboard"),
		),
	)

	if err != nil {
		panic(err)
	}

	return r
}

// TODO: What is this?
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceExporter(ctx context.Context) *otlptrace.Exporter {
	exp, err := otlptracehttp.New(ctx)

	if err != nil {
		panic(err)
	}

	return exp
}

func newTracerProvider(exp sdktrace.SpanExporter, r *resource.Resource) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

func newMeterExporter(ctx context.Context) *otlpmetrichttp.Exporter {
	exp, err := otlpmetrichttp.New(ctx)

	if err != nil {
		panic(err)
	}

	return exp
}

func newMeterProvider(exp sdkmetric.Exporter, r *resource.Resource) *sdkmetric.MeterProvider {
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(r),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
	)
}

func newLoggerExporter(ctx context.Context) *otlploghttp.Exporter {
	exp, err := otlploghttp.New(ctx)

	if err != nil {
		panic(err)
	}

	return exp
}

func newLoggerProvider(exp sdklog.Exporter, r *resource.Resource) *sdklog.LoggerProvider {
	return sdklog.NewLoggerProvider(
		sdklog.WithResource(r),
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exp)),
	)
}
