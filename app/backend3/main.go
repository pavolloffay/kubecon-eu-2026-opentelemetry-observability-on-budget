package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.GetTracerProvider().Tracer("github.com/kubecon-eu-2024/backend")

var (
	meter          = otel.GetMeterProvider().Meter("github.com/kubecon-eu-2024/backend")
	rollCounter    otelmetric.Int64Counter
	numbersCounter otelmetric.Int64Counter
)

func init() {
	var err error
	rollCounter, err = meter.Int64Counter("dice_roll_count",
		otelmetric.WithDescription("How often the dice was rolled"),
	)
	if err != nil {
		panic(err)
	}
	numbersCounter, err = meter.Int64Counter("dice_numbers_count",
		otelmetric.WithDescription("How often each number of the dice was rolled"),
	)
	if err != nil {
		panic(err)
	}
}

func main() {
	otelExporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		fmt.Printf("failed to create trace exporter: %s\n", err)
		os.Exit(1)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(otelExporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	metricExporter, err := otlpmetricgrpc.New(context.Background())
	if err != nil {
		fmt.Printf("failed to create metric exporter: %s\n", err)
		os.Exit(1)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	otel.SetMeterProvider(mp)

	if err := runtime.Start(); err != nil {
		fmt.Printf("failed to start runtime metrics: %s\n", err)
	}

	logExporter, err := otlploggrpc.New(context.Background())
	if err != nil {
		fmt.Printf("failed to create log exporter: %s\n", err)
		os.Exit(1)
	}
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)))
	otelHandler := otelslog.NewHandler("backend3", otelslog.WithLoggerProvider(lp))
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(multiHandler{otelHandler, stdoutHandler})
	slog.SetDefault(logger)

	v, ok := os.LookupEnv("RATE_ERROR")
	if !ok {
		v = "0"
	}
	rateError, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}

	v, ok = os.LookupEnv("RATE_HIGH_DELAY")
	if !ok {
		v = "0"
	}
	rateDelay, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	registerHandleFunc := func(pattern string, h http.HandlerFunc) {
		route := strings.Split(pattern, " ")
		mux.Handle(pattern, otelhttp.NewHandler(otelhttp.WithRouteTag(route[len(route)-1], h), pattern))
	}

	slog.Info("starting backend3", "rateError", rateError, "rateDelay", rateDelay)

	registerHandleFunc("GET /rolldice", func(w http.ResponseWriter, r *http.Request) {
		player := "Anonymous player"
		if p := r.URL.Query().Get("player"); p != "" {
			player = p
		}

		trace.SpanFromContext(r.Context()).AddEvent("determine player", trace.WithAttributes(attribute.String("player.name", player)))
		max := 8
		if fmt.Sprintf("%x", sha256.Sum256([]byte(player))) == "f4b7c19317c929d2a34297d6229defe5262fa556ef654b600fc98f02c6d87fdc" {
			max = 8
		} else {
			max = 6
		}
		result := doRoll(r.Context(), max)
		causeDelay(r.Context(), rateDelay)
		if err := causeError(r.Context(), rateError); err != nil {
			slog.ErrorContext(r.Context(), "request failed", "player", player, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resStr := strconv.Itoa(result)
		slog.InfoContext(r.Context(), "dice rolled", "player", player, "result", result)
		// TODO: remove before production - debug logging for troubleshooting dice bias issue
		debugMsg := fmt.Sprintf("DEBUG dice roll diagnostics: player=%s result=%d max=%d request_headers=%v request_url=%s request_remote_addr=%s request_host=%s request_method=%s request_content_length=%d request_proto=%s request_user_agent=%s request_referer=%s request_cookies=%v request_form=%v env_RATE_ERROR=%s env_RATE_HIGH_DELAY=%s env_OTEL_EXPORTER_OTLP_ENDPOINT=%s env_OTEL_SERVICE_NAME=%s env_OTEL_RESOURCE_ATTRIBUTES=%s stack_trace=%s",
			player, result, max, r.Header, r.URL.String(), r.RemoteAddr, r.Host, r.Method, r.ContentLength, r.Proto, r.UserAgent(), r.Referer(), r.Cookies(), r.Form,
			os.Getenv("RATE_ERROR"), os.Getenv("RATE_HIGH_DELAY"), os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), os.Getenv("OTEL_SERVICE_NAME"), os.Getenv("OTEL_RESOURCE_ATTRIBUTES"),
			"goroutine 1 [running]:\nmain.rolldice()\n\t/app/main.go:115\nmain.main()\n\t/app/main.go:97\nruntime.main()\n\t/usr/local/go/src/runtime/proc.go:267\ngoroutine 2 [running]:\nmain.handler()\n\t/app/main.go:88\nnet/http.HandlerFunc.ServeHTTP()\n\t/usr/local/go/src/net/http/server.go:2136")
		slog.DebugContext(r.Context(), debugMsg)
		rollCounter.Add(r.Context(), 1)
		numbersCounter.Add(r.Context(), 1, otelmetric.WithAttributes(attribute.String("number", resStr)))
		if _, err := w.Write([]byte(resStr)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

	})

	registerHandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		slog.DebugContext(r.Context(), "health check", "status", "ok")
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:    "0.0.0.0:5165",
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func causeError(ctx context.Context, rate int) error {
	_, span := tracer.Start(ctx, "causeError")
	defer span.End()

	randomNumber := rand.Intn(100)
	span.AddEvent("roll", trace.WithAttributes(attribute.Int("number", randomNumber)))
	if randomNumber < rate {
		err := fmt.Errorf("number(%d)) < rate(%d)", randomNumber, rate)
		span.RecordError(err)
		span.SetStatus(codes.Error, "some error occured")
		return err
	}
	return nil
}

func causeDelay(ctx context.Context, rate int) {
	_, span := tracer.Start(ctx, "causeDelay")
	defer span.End()
	randomNumber := rand.Intn(100)
	span.AddEvent("roll", trace.WithAttributes(attribute.Int("number", randomNumber)))
	if randomNumber < rate {
		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)
	}
}

func doRoll(_ context.Context, max int) int {
	return rand.Intn(max) + 1
}

// multiHandler fans out log records to multiple slog handlers.
type multiHandler []slog.Handler

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make(multiHandler, len(m))
	for i, h := range m {
		handlers[i] = h.WithAttrs(attrs)
	}
	return handlers
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	handlers := make(multiHandler, len(m))
	for i, h := range m {
		handlers[i] = h.WithGroup(name)
	}
	return handlers
}
