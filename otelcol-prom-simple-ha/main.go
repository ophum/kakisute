package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	conn, err := grpc.NewClient(os.Args[1],
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := initMeter(ctx, conn)
	if err != nil {
		return err
	}
	defer func() {
		ctx := context.Background()
		_ = shutdown(ctx)
	}()

	meter := otel.Meter("example")

	testCounter, err := meter.Int64Counter("test")
	if err != nil {
		return err
	}

	for range 1000 {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		testCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("hello", "world"),
			attribute.String("label", os.Args[2]),
		))
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func initMeter(ctx context.Context, conn *grpc.ClientConn) (func(context.Context) error, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("test"),
		),
	)
	if err != nil {
		_ = metricExporter.Shutdown(ctx)
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return metricExporter.Shutdown, nil
}
