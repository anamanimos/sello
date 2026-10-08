# Sello — SaaS Katalog Produk + WhatsApp

> "Vibe coding antigravity" untuk membuat website katalog produk yang mudah, cepat, dan bisa dikustomisasi.

---

## 1. Deskripsi

Sello adalah SaaS yang memungkinkan pengguna membuat **website katalog produk gratis** dengan subdomain sendiri.  
Fitur utama:

- Maks. 20 produk per user (Free Plan)
- Penyimpanan gambar & template 50MB per user
- Template unik & custom per toko
- Tombol WhatsApp untuk kontak cepat
- Scalable ke ribuan toko

---

## 2. Teknologi & Stack

| Layer           | Teknologi / Tools                    | Keterangan                |
| --------------- | ------------------------------------ | ------------------------- |
| Backend         | Go + Gin                             | API ringan & cepat        |
| ORM / DB        | GORM + Supabase PostgreSQL           | CRUD & multi-tenant       |
| Auth            | Supabase Auth                        | Register / login / JWT    |
| Storage         | Supabase Storage / Cloudinary        | Gambar & template         |
| Frontend        | HTML/CSS/JS ringan                   | Render template dinamis   |
| Deployment      | Cloud (Vercel / Fly.io / Cloudflare) | Wildcard subdomain        |
| Caching         | Optional: Redis                      | Template & config caching |
| Version Control | Git                                  | Source code management    |

---

## 3. Database Schema

### Users

- Disediakan Supabase Auth (`auth.users`)
- Field tambahan:
  - `storage_used BIGINT` → track quota

### Stores

| Field      | Type      | Keterangan            |
| ---------- | --------- | --------------------- |
| id         | uuid PK   | Auto-generated        |
| user_id    | uuid FK   | References auth.users |
| slug       | text      | Subdomain unik        |
| name       | text      | Nama toko             |
| whatsapp   | text      | Nomor WhatsApp        |
| created_at | timestamp | Timestamp             |

### Products

| Field       | Type      | Keterangan        |
| ----------- | --------- | ----------------- |
| id          | uuid PK   | Auto-generated    |
| store_id    | uuid FK   | References stores |
| name        | text      | Nama produk       |
| price       | bigint    | Harga             |
| description | text      | Deskripsi         |
| image_url   | text      | Gambar produk     |
| image_size  | bigint    | Ukuran file       |
| created_at  | timestamp | Timestamp         |

### Templates

| Field      | Type      | Keterangan        |
| ---------- | --------- | ----------------- |
| id         | uuid PK   | Auto-generated    |
| store_id   | uuid FK   | References stores |
| version    | int       | Versi template    |
| html_url   | text      | URL HTML template |
| css_url    | text      | URL CSS           |
| js_url     | text      | URL JS            |
| created_at | timestamp | Timestamp         |

---

## 4. Arsitektur High-Level

```text
[User / Editor Dashboard]
        ↓ (JSON config)
[Supabase Storage] → HTML / CSS / JS
        ↓
[Backend Go + Gin] → fetch template + inject data
        ↓
[Supabase PostgreSQL] → config, products, store info
        ↓
[Cache: optional Redis]
        ↓
[Public Store Page] → subdomain user.sello.app

## 5. Fitur MVP
Register / login user (Supabase Auth)

CRUD Store & Product

Upload template HTML/CSS/JS ke Supabase Storage

Render template dinamis per subdomain

Button WhatsApp untuk tiap toko

Limit Free Plan: 20 produk + 50MB storage

## 6. Roadmap “Antigravity”
Stage	Fitur / Goal
MVP	Backend Go + Supabase, CRUD produk, template render
Stage 1	Editor text + color + image upload per store
Stage 2	Advanced editor (drag & drop ala Elementor)
Stage 3	Premium plan → lebih banyak produk & storage
Stage 4	Analytics, order tracking, multi-language
Stage 5	Marketplace integration (optional)

## 7. Setup Lokal (Windows / PowerShell)
# Clone project
git clone <repo-url>
cd sello-backend

# Init Go module
go mod init sello

# Install dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/supabase/supabase-go
go get github.com/golang-jwt/jwt/v5

# Jalankan server
go run cmd/main.go
Cek endpoint:

http://localhost:8080/
http://localhost:8080/ping

## 8. Notes
Tiap toko punya template unik → HTML/CSS/JS di storage, config di DB

Caching template → untuk performa

Fallback template → untuk error / corrupt template

Security → sanitize HTML/JS & batasi storage

Supabase memudahkan MVP → bisa fokus ke fitur, tidak repot setup lokal PostgreSQL
```
