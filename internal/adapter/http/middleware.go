package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicolasmmb/pismo-challenge/internal/port"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var requestCounter uint64

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, m ...Middleware) http.Handler {
	wrapped := h
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}
	return wrapped
}

func WithRecovery(log port.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", map[string]any{"error": rec})
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func WithLogging(log port.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(sr, r)
			log.Info("request",
				map[string]any{
					"method": r.Method,
					"path":   r.URL.Path,
					"status": sr.status,
					"dur_ms": time.Since(start).Milliseconds(),
				})
		})
	}
}

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method"})

	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"path", "method", "status"})
)

func WithMetrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			next.ServeHTTP(sr, r)

			duration := time.Since(start).Seconds()

			httpDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
			httpRequests.WithLabelValues(r.URL.Path, r.Method, http.StatusText(sr.status)).Inc()
		})
	}
}

func MetricsHandler() http.Handler {
	return promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

func WithTimeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func WithTracing(serviceName string) Middleware {
	rootTracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			spanName := r.Method + " " + r.URL.Path
			ctx, span := rootTracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(r.Method),
					semconv.HTTPTargetKey.String(r.URL.Path),
					semconv.HTTPURLKey.String(r.URL.String()),
					attribute.String("http.host", r.Host),
				),
			)
			defer span.End()

			span.AddEvent("Request received")

			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			next.ServeHTTP(sr, r.WithContext(ctx))

			span.AddEvent("Response sent", trace.WithAttributes(
				attribute.Int("http.status_code", sr.status),
			))

			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(sr.status),
			)

			if sr.status >= 400 {
				span.SetAttributes(attribute.Bool("error", true))
			}
		})
	}
}
