package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	logPerSec   int
	duration    time.Duration
	otlpAddress string
}

func (c *Config) ParseFlag() {
	flag.IntVar(&c.logPerSec, "log-per-sec", 100, "log count per second")
	flag.DurationVar(&c.duration, "duration", time.Minute, "duration")
	flag.StringVar(&c.otlpAddress, "otlp-address", "localhost:4317", "otlp address (grpc)")
	flag.Parse()
}

var config Config

func init() {
	config.ParseFlag()
}

func main() {
	conn, err := grpc.NewClient(config.otlpAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceNameKey.String("testloggenerator"),
	))
	if err != nil {
		log.Fatal(err)
	}

	shutdownMeterProvider, err := initSelfMetricsProvider(ctx, res, conn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownMeterProvider(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	meter := otel.Meter("github.com/ophum/kakisute/testloggenerator")
	gaugeSendMetricElapsed, err := meter.Float64Gauge("log_generate_elapsed")
	if err != nil {
		log.Fatal(err)
	}

	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatal(err)
	}
	defer logExporter.Shutdown(ctx)

	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	defer logProvider.Shutdown(ctx)
	global.SetLoggerProvider(logProvider)

	logger := otelslog.NewLogger("testloggenerator")

	ctx, cancel = context.WithTimeout(ctx, config.duration)
	defer cancel()

	limiter := rate.NewLimiter(rate.Limit(config.logPerSec), 1)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := limiter.Wait(ctx); err != nil {
			log.Fatal(err)
		}

		go func() {
			s := time.Now()
			msg, attrs := generateLog()
			if err != nil {
				return
			}
			logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
			elapsed := time.Since(s)

			gaugeSendMetricElapsed.Record(ctx, float64(elapsed)/float64(time.Millisecond))
		}()
	}
}

func generateLog() (string, []slog.Attr) {
	return apiLog()
}

type APILog struct {
	Method   string
	Path     string
	Statuses []int
}

func (a *APILog) Log() (string, []slog.Attr) {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(len(a.Statuses))))
	if err != nil {
		log.Fatal(err)
	}

	status := a.Statuses[i.Int64()]
	return fmt.Sprintf("%s %d %s", a.Method, status, a.Path), []slog.Attr{
		slog.String("method", a.Method),
		slog.String("path", "/api/v1/posts"),
		slog.Int("status", http.StatusOK),
	}
}

var apiLogs = []APILog{
	{Method: "GET", Path: "/api/v1/posts", Statuses: []int{200, 400, 401, 403, 404, 500, 503}},
	{Method: "POST", Path: "/api/v1/posts", Statuses: []int{201, 400, 401, 403, 500, 503}},
}

func apiLog() (string, []slog.Attr) {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(len(apiLogs))))
	if err != nil {
		log.Fatal(err)
	}

	return apiLogs[i.Int64()].Log()
}
