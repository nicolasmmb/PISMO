package e2e_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	adapterhttp "github.com/nicolasmmb/pismo-challenge/internal/adapter/http"
	"github.com/nicolasmmb/pismo-challenge/internal/adapter/logger"
	"github.com/nicolasmmb/pismo-challenge/internal/adapter/repository"
	"github.com/nicolasmmb/pismo-challenge/internal/usecase"
)

var db *sql.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	_ = resource.Expire(120)

	port := resource.GetPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://user_name:secret@localhost:%s/dbname?sslmode=disable", port)

	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("Could not run migrations: %s", err)
	}

	// Seed Operation Types
	if _, err := db.Exec(`INSERT INTO operation_types (id, description, sign) VALUES (4, 'PAGAMENTO', 1)`); err != nil {
		log.Fatalf("failed to seed op types: %v", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func runMigrations(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id BIGSERIAL PRIMARY KEY,
			document_number TEXT UNIQUE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS operation_types (
			id INT PRIMARY KEY,
			description TEXT NOT NULL,
			sign SMALLINT NOT NULL CHECK (sign IN (-1, 1))
		);`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL REFERENCES accounts(id),
			operation_type_id INT NOT NULL REFERENCES operation_types(id),
			amount_cents BIGINT NOT NULL,
			event_date TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func SetupRouter() http.Handler {
	log := logger.New()
	accountRepo := repository.NewAccountRepository(db)
	opTypeRepo := repository.NewOperationTypeRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	tm := repository.NewTransactionManager(db)

	createAccountUC := &usecase.CreateAccount{Accounts: accountRepo}
	getAccountUC := &usecase.GetAccount{Accounts: accountRepo}
	createTxUC := &usecase.CreateTransaction{
		Accounts:           accountRepo,
		OperationTypes:     opTypeRepo,
		Transactions:       txRepo,
		TransactionManager: tm,
	}

	accountHandler := adapterhttp.NewAccountHandler(createAccountUC, getAccountUC)
	txHandler := adapterhttp.NewTransactionHandler(createTxUC)

	return adapterhttp.NewRouter(log, accountHandler, txHandler)
}

func TestE2E_FullFlow(t *testing.T) {
	router := SetupRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := server.Client()

	// 1. Create Account
	t.Run("Create Account", func(t *testing.T) {
		reqBody := `{"document_number": "E2E_DOC_123"}`
		resp, err := client.Post(server.URL+"/accounts", "application/json", bytes.NewBufferString(reqBody))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["account_id"])
	})

	// 2. Create Transaction
	t.Run("Create Transaction", func(t *testing.T) {
		var accountID int64
		err := db.QueryRow("SELECT id FROM accounts WHERE document_number = 'E2E_DOC_123'").Scan(&accountID)
		assert.NoError(t, err)

		reqBody := fmt.Sprintf(`{"account_id": %d, "operation_type_id": 4, "amount": 50.00}`, accountID)
		resp, err := client.Post(server.URL+"/transactions", "application/json", bytes.NewBufferString(reqBody))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}
