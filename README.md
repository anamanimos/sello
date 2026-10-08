# Sello — SaaS Katalog Produk & Toko Online WhatsApp

> Platform katalog produk instan dan toko online multi-tenant yang mudah, cepat, dan terintegrasi dengan pemesanan WhatsApp.

---

## 📁 Struktur Monorepo

```
sello/
├── backend/          # Core backend (Go + Gin + GORM PostgreSQL)
│   ├── cmd/          # Entry points (main.go, seeder, migration)
│   ├── config/       # Environment configuration
│   ├── internal/     # Models, database, middleware, handlers, services
│   ├── web/          # Merchant dashboard, Super Admin console, theme engine
│   └── scripts/      # Database & migration scripts
├── frontend/         # Marketing landing page (React 18 + Vite + Tailwind + shadcn/ui)
│   ├── src/          # Components, pages, hooks, styling
│   └── package.json
└── landing-static/   # Static HTML/CSS/JS export versi landing page
```

---

## 🚀 Quick Start

### 1. Backend (Go + Gin)

```bash
cd backend
cp .env.example .env
# Sesuaikan kredensial PostgreSQL dan konfigurasi di .env

go mod tidy
go run cmd/main.go
```

- **Dashboard Merchant**: `http://localhost:8080/dashboard`
- **Super Admin Console**: `http://localhost:8080/console/login` (Default: `admin@sello.id` / `admin123`)
- **Storefront**: `http://localhost:8080/s/:slug`

### 2. Frontend (React + Vite)

```bash
cd frontend
npm install
npm run dev
```

---

## 🛠️ Tech Stack

- **Backend**: Go (Gin Web Framework), GORM ORM
- **Database**: PostgreSQL (Supabase compatible)
- **Frontend**: React 18, Vite, TypeScript, Tailwind CSS, shadcn/ui
- **Auth**: JWT, Google OAuth 2.0, OTP (WhatsApp & Email)
- **Storefront**: Dynamic file-based Theme Engine (`/backend/web/themes`)
- **Storage**: Cloudinary & Local Storage
