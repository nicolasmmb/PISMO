# Pismo Challenge - API de Transações

API de contas financeiras e transações. Construída com Go + Arquitetura Hexagonal.

## Início Rápido

```bash
./scripts/start.sh
```

## Endpoints

| Método | Path | Descrição |
|--------|------|-----------|
| `POST` | `/accounts` | Criar conta |
| `GET` | `/accounts/{id}` | Buscar conta |
| `POST` | `/transactions` | Criar transação ¹ |
| `GET` | `/healthz` | Health check |
| `GET` | `/metrics` | Métricas Prometheus |

> ¹ Usa transação de banco com lock de linha  

## URLs de Acesso

| Serviço | URL | Credenciais |
|---------|-----|-------------|
| **API** | http://localhost:8080 | - |
| **Grafana** | http://localhost:3000 | Sem login |
| **Prometheus** | http://localhost:9090 | - |
| **pgAdmin** | http://localhost:5050 | `admin@pismo.com` / `admin` |

### pgAdmin → PostgreSQL

| Campo | Valor |
|-------|-------|
| Host | `postgres` |
| Port | `5432` |
| Database | `pismo` |
| User/Pass | `pismo` / `pismo` |

## Exemplos da API

### Criar Conta

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number": "12345678900"}'
```

### Criar Transação

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "operation_type_id": 4, "amount": 123.45}'
```

## Tipos de Operação

| ID | Descrição | Sinal | Efeito |
|----|-----------|-------|--------|
| 1 | COMPRA A VISTA | -1 | Débito |
| 2 | COMPRA PARCELADA | -1 | Débito |
| 3 | SAQUE | -1 | Débito |
| 4 | PAGAMENTO | +1 | Crédito |

## Arquitetura

```
internal/
├── domain/        # Entidades (Account, Transaction)
├── port/          # Interfaces (repositories)
├── usecase/       # Lógica de negócio
└── adapter/
    ├── http/      # Handlers, Router
    └── repository/# Acesso ao banco
```

## Fluxo de Transação

```
POST /transactions
    ↓
Handler → UseCase → TransactionManager.RunInTransaction()
                        ↓
                    BEGIN TX
                    SELECT ... FOR UPDATE (trava conta)
                    INSERT transaction
                    COMMIT
```

## Testes

```bash
make test
```

## Stack Tecnológica

- Go 1.24
- PostgreSQL 15
- Prometheus + Grafana
- OpenTelemetry + Tempo
