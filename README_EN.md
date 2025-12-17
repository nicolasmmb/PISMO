# Pismo Challenge - Transaction API

Financial accounts and transactions API. Built with Go  

## Quick Start

```bash
./scripts/start.sh
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/accounts` | Create account |
| `GET` | `/accounts/{id}` | Get account |
| `POST` | `/transactions` | Create transaction ¹ |
| `GET` | `/healthz` | Health check |
| `GET` | `/metrics` | Prometheus metrics |

> ¹ Uses database transaction with row locking  

## Access URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| **API** | http://localhost:8080 | - |
| **Grafana** | http://localhost:3000 | No login required |
| **Prometheus** | http://localhost:9090 | - |
| **pgAdmin** | http://localhost:5050 | `admin@pismo.com` / `admin` |

### pgAdmin → PostgreSQL

| Field | Value |
|-------|-------|
| Host | `postgres` |
| Port | `5432` |
| Database | `pismo` |
| User/Pass | `pismo` / `pismo` |

## API Examples

### Create Account

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number": "12345678900"}'
```

### Create Transaction

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "operation_type_id": 4, "amount": 123.45}'
```

## Operation Types

| ID | Description | Sign | Effect |
|----|-------------|------|--------|
| 1 | COMPRA A VISTA | -1 | Debit |
| 2 | COMPRA PARCELADA | -1 | Debit |
| 3 | SAQUE | -1 | Debit |
| 4 | PAGAMENTO | +1 | Credit |

## Architecture

```
internal/
├── domain/        # Entities (Account, Transaction)
├── port/          # Interfaces (repositories)
├── usecase/       # Business logic
└── adapter/
    ├── http/      # Handlers, Router
    └── repository/# Database access
```

## Transaction Flow

```
POST /transactions
    ↓
Handler → UseCase → TransactionManager.RunInTransaction()
                        ↓
                    BEGIN TX
                    SELECT ... FOR UPDATE (lock account)
                    INSERT transaction
                    COMMIT
```

## Tests

```bash
make test
```

## Tech Stack

- Go 1.24
- PostgreSQL 15
- Prometheus + Grafana
- OpenTelemetry + Tempo
