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
	logger := slog.New(otelslog.NewHandler("backend3", otelslog.WithLoggerProvider(lp)))
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
		rollCounter.Add(r.Context(), 1)
		numbersCounter.Add(r.Context(), 1, otelmetric.WithAttributes(attribute.String("number", resStr)))
		if _, err := w.Write([]byte(resStr)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
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
