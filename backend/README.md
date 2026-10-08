# Sello Backend

SaaS Katalog Produk + WhatsApp

## Tech Stack

- **Backend**: Go + Gin
- **Database**: Supabase PostgreSQL + GORM
- **Auth**: Supabase Auth (JWT)
- **Storage**: Supabase Storage

## Quick Start

```bash
# Clone & setup
git clone <repo-url>
cd sello-backend
cp .env.example .env

# Edit .env with your Supabase credentials

# Install dependencies
go mod tidy

# Run server
go run cmd/main.go
```

## API Endpoints

### Public

| Method | Endpoint                | Description            |
| ------ | ----------------------- | ---------------------- |
| GET    | `/ping`                 | Health check           |
| GET    | `/store/:slug`          | Get store by subdomain |
| GET    | `/store/:slug/products` | Get store products     |

### Protected (require JWT)

| Method | Endpoint                         | Description    |
| ------ | -------------------------------- | -------------- |
| GET    | `/api/stores`                    | List my stores |
| POST   | `/api/stores`                    | Create store   |
| PUT    | `/api/stores/:id`                | Update store   |
| DELETE | `/api/stores/:id`                | Delete store   |
| POST   | `/api/stores/:store_id/products` | Create product |
| PUT    | `/api/products/:id`              | Update product |
| DELETE | `/api/products/:id`              | Delete product |

## Project Structure

```
sello-backend/
├── cmd/main.go              # Entry point
├── config/config.go         # Environment config
├── internal/
│   ├── database/            # DB connection
│   ├── handlers/            # API handlers
│   ├── middleware/          # Auth middleware
│   └── models/              # GORM models
├── .env.example             # Env template
└── go.mod                   # Dependencies
```
