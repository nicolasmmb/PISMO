package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

var startTime = time.Now()

func NewRouter(
	log port.Logger,
	accountHandler *AccountHandler,
	transactionHandler *TransactionHandler,
) http.Handler {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/", healthHandler)
	apiMux.HandleFunc("GET /healthz", healthzHandler)
	apiMux.HandleFunc("POST /accounts", accountHandler.CreateAccount)
	apiMux.HandleFunc("GET /accounts/{accountID}", accountHandler.GetAccount)
	apiMux.HandleFunc("POST /transactions", transactionHandler.CreateTransaction)

	apiHandler := Chain(
		apiMux,
		WithTimeout(30*time.Second),
		WithTracing("pismo-api"),
		WithLogging(log),
		WithMetrics(),
		WithRecovery(log),
	)

	rootMux := http.NewServeMux()
	rootMux.Handle("/metrics", MetricsHandler())
	rootMux.Handle("/healthz", WithMetrics()(http.HandlerFunc(healthzHandler)))
	rootMux.Handle("/", apiHandler)

	return rootMux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	response := map[string]any{
		"status":         "healthy",
		"uptime":         uptime.String(),
		"uptime_seconds": uptime.Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
