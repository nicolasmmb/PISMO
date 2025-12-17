package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/nicolasmmb/pismo-challenge/internal/adapter/repository"
	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// 1. Initialize Dockertest
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// 2. Start Postgres Container
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

	resource.Expire(120) // Kill container after 120 seconds if test hangs

	// 3. Connect to DB with Exponential Backoff
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

	// 4. Run Migrations (Simple Schema Setup for Test)
	if err := runMigrations(db); err != nil {
		log.Fatalf("Could not run migrations: %s", err)
	}

	// 5. Run Tests
	code := m.Run()

	// 6. Cleanup
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

func TestAccountRepository(t *testing.T) {
	repo := repository.NewAccountRepository(db)
	ctx := context.Background()

	t.Run("Create Account", func(t *testing.T) {
		acc := domain.Account{DocumentNumber: "12345678900"}
		id, err := repo.Create(ctx, acc)
		assert.NoError(t, err)
		assert.NotZero(t, id)

		// Verify retrieval
		fetched, err := repo.FindByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, acc.DocumentNumber, fetched.DocumentNumber)
	})

	t.Run("Create Duplicate Account", func(t *testing.T) {
		acc := domain.Account{DocumentNumber: "99999999900"}
		_, err := repo.Create(ctx, acc)
		assert.NoError(t, err)

		_, err = repo.Create(ctx, acc)
		assert.Error(t, err) // Unique constraint violation
	})
}

func TestTransactionLocking(t *testing.T) {
	repo := repository.NewAccountRepository(db)
	acc := domain.Account{DocumentNumber: "LOCK_TEST"}
	id, err := repo.Create(context.Background(), acc)
	assert.NoError(t, err)

	tm := repository.NewTransactionManager(db)

	started1 := make(chan struct{})
	finished1 := make(chan struct{})

	go func() {
		err := tm.RunInTransaction(context.Background(), func(ctx context.Context) error {
			close(started1)
			_, err := repo.FindByIDForUpdate(ctx, id)
			assert.NoError(t, err)

			time.Sleep(1 * time.Second)
			return nil
		})
		assert.NoError(t, err)
		close(finished1)
	}()

	<-started1
	time.Sleep(100 * time.Millisecond)

	start2 := time.Now()
	err = tm.RunInTransaction(context.Background(), func(ctx context.Context) error {
		_, err := repo.FindByIDForUpdate(ctx, id)
		return err
	})
	assert.NoError(t, err)
	duration := time.Since(start2)

	<-finished1

	// If lock worked, goroutine 2 should have waited at least ~900ms (1s minus overhead)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(800), "Transaction 2 should have waited for lock")
}
