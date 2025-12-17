package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	adapterhttp "github.com/nicolasmmb/pismo-challenge/internal/adapter/http"
	loggeradapter "github.com/nicolasmmb/pismo-challenge/internal/adapter/logger"
	"github.com/nicolasmmb/pismo-challenge/internal/adapter/repository"
	"github.com/nicolasmmb/pismo-challenge/internal/config"
	"github.com/nicolasmmb/pismo-challenge/internal/usecase"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.Load()
	log := loggeradapter.New()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.OTLPEndpoint != "" {
		shutdownOtel, err := initTracer(ctx, cfg.OTLPEndpoint, "pismo-app")
		if err != nil {
			log.Error("failed to init otel", map[string]any{"error": err})
		} else {
			defer func() {
				if err := shutdownOtel(context.Background()); err != nil {
					log.Error("failed to shutdown otel", map[string]any{"error": err})
				}
			}()
		}
	}

	log.Info("starting service", map[string]any{"port": cfg.Port, "db_driver": cfg.DBDriver})

	if err := run(ctx, cfg, log); err != nil {
		log.Error("service failed", map[string]any{"error": err})
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, log loggeradapter.SlogLogger) error {
	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}

	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}

	prometheus.MustRegister(collectors.NewDBStatsCollector(db, "pismo"))

	accountRepo := repository.NewAccountRepository(db)
	opTypeRepo := repository.NewOperationTypeRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	createAccountUC := &usecase.CreateAccount{
		Accounts: accountRepo,
	}
	getAccountUC := &usecase.GetAccount{
		Accounts: accountRepo,
	}

	tm := repository.NewTransactionManager(db)
	createTxUC := &usecase.CreateTransaction{
		Accounts:           accountRepo,
		OperationTypes:     opTypeRepo,
		Transactions:       txRepo,
		TransactionManager: tm,
	}

	accountHandler := adapterhttp.NewAccountHandler(createAccountUC, getAccountUC)
	txHandler := adapterhttp.NewTransactionHandler(createTxUC)

	handler := adapterhttp.NewRouter(log, accountHandler, txHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server failed", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down http server", nil)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func initTracer(ctx context.Context, otlpEndpoint string, serviceName string) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	conn, err := grpc.NewClient(otlpEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}
