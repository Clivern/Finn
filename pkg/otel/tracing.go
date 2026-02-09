// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package otel

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otelapi "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

func init() {
	viper.SetDefault("app.tracing.enabled", false)
	viper.SetDefault("app.tracing.exporter", "stdout")
	viper.SetDefault("app.tracing.endpoint", "http://localhost:4318")
	viper.SetDefault("app.tracing.insecure", true)
	viper.SetDefault("app.tracing.sample_ratio", 1.0)
	viper.SetDefault("app.tracing.service_version", "dev")
}

// InitTracing sets up the global OpenTelemetry tracer provider when tracing is enabled.
func InitTracing(ctx context.Context) (func(context.Context) error, error) {
	if !viper.GetBool("app.tracing.enabled") {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := newTraceExporter(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(viper.GetString("app.name")),
			semconv.ServiceVersion(viper.GetString("app.tracing.service_version")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("tracing resource: %w", err)
	}

	ratio := viper.GetFloat64("app.tracing.sample_ratio")
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otelapi.SetTracerProvider(tp)
	otelapi.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown := func(shutdownCtx context.Context) error {
		return tp.Shutdown(shutdownCtx)
	}

	log.Info().
		Str("exporter", viper.GetString("app.tracing.exporter")).
		Float64("sample_ratio", ratio).
		Msg("OpenTelemetry tracing enabled")

	return shutdown, nil
}

func newTraceExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	switch viper.GetString("app.tracing.exporter") {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlp":
		opts := buildOTLPOptions()
		return otlptracehttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf(
			"unknown app.tracing.exporter %q (use otlp or stdout)",
			viper.GetString("app.tracing.exporter"),
		)
	}
}

func buildOTLPOptions() []otlptracehttp.Option {
	var opts []otlptracehttp.Option

	endpoint := viper.GetString("app.tracing.endpoint")

	if endpoint == "" {
		opts = append(opts, otlptracehttp.WithEndpointURL("http://localhost:4318"))
	} else if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		opts = append(opts, otlptracehttp.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	}

	if viper.GetBool("app.tracing.insecure") {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return opts
}

// WrapHTTPHandler adds HTTP request spans when tracing is enabled.
func WrapHTTPHandler(handler http.Handler) http.Handler {
	if !viper.GetBool("app.tracing.enabled") {
		return handler
	}

	return otelhttp.NewHandler(handler, viper.GetString("app.name"),
		otelhttp.WithFilter(func(r *http.Request) bool {
			p := r.URL.Path
			if p == "/api/v1/public/_health" || p == "/api/v1/public/_ready" {
				return false
			}
			return true
		}),
	)
}
