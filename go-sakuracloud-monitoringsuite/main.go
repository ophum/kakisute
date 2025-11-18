package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	slogmulti "github.com/samber/slog-multi"
	otelconf "go.opentelemetry.io/contrib/otelconf/v0.3.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("failed to run", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	expand := os.ExpandEnv(string(b))
	fmt.Println(expand)
	c, err := otelconf.ParseYAML([]byte(expand))
	if err != nil {
		return err
	}

	s, err := otelconf.NewSDK(
		otelconf.WithContext(ctx),
		otelconf.WithOpenTelemetryConfiguration(*c),
	)
	if err != nil {
		slog.Error("failed to NewSDK")
		return err
	}
	defer func() {
		if err := s.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown", "error", err)
			os.Exit(1)
		}
	}()

	otel.SetTracerProvider(s.TracerProvider())
	otel.SetMeterProvider(s.MeterProvider())
	global.SetLoggerProvider(s.LoggerProvider())

	logger := slog.New(slogmulti.Fanout(
		otelslog.NewHandler("slog-test"),
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	))
	slog.SetDefault(logger)
	return Start(ctx)
}

func Start(ctx context.Context) error {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	i := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			go func(i int) {
				ctx, cancel := context.WithTimeout(ctx, time.Second*2)
				defer cancel()

				if err := fn(ctx, i); err != nil {
					slog.Error("failed", "error", err)
				}
			}(i)

			i++
		}

	}
}

var tracer = otel.Tracer("go-sakuracloud-monitoringsuite")

func fn(ctx context.Context, id int) error {
	ctx, span := tracer.Start(ctx, "fn", trace.WithAttributes(
		attribute.Int("id", id),
	))
	defer span.End()

	slog.Info("fn", "id", id)
	return fn2(ctx)
}

func fn2(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "fn2")
	defer span.End()

	time.Sleep(time.Second)
	return fn3(ctx)
}

func fn3(ctx context.Context) error {
	_, span := tracer.Start(ctx, "fn3")
	defer span.End()

	for i := range 10 {
		span.AddEvent("event", trace.WithTimestamp(time.Now()), trace.WithStackTrace(true), trace.WithAttributes(
			attribute.Int("count", i),
		))
	}
	return nil
}
