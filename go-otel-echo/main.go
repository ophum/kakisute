package main

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
func run() error {
	ctx := context.Background()
	tp, err := initTracer(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	e := echo.New()
	e.Use(otelecho.Middleware("go-otel-echo"))
	e.Use(middleware.Logger())

	e.POST("/", ping)

	return e.Start(":8080")
}

var tracer = otel.Tracer("echo-server")

type PingRequest struct {
	Name string `json:"name"`
}

func ping(c echo.Context) error {
	var req PingRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	res, err := pingImpl(c.Request().Context(), req.Name)
	if err != nil {
		return err
	}
	slow(c.Request().Context())
	return c.JSON(http.StatusOK, map[string]any{
		"pong": res,
	})
}

func pingImpl(ctx context.Context, name string) (string, error) {
	_, span := tracer.Start(ctx, "pingImpl")
	defer span.End()

	span.AddEvent("create response", trace.WithAttributes(
		attribute.String("name", name),
	))
	msg := "hello " + name

	return msg, nil
}

func slow(ctx context.Context) {
	_, span := tracer.Start(ctx, "slow")
	defer span.End()

	for i := range 10 {
		span.AddEvent("loop", trace.WithAttributes(
			attribute.Int("count", i),
		))
		time.Sleep(time.Second)
	}
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	conn, err := grpc.NewClient("clickhouse-single.cloud.t-inagaki.net:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp, nil
}
