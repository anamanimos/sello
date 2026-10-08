# Panduan Pengguna Sello

Selamat datang di **Sello**, platform pembuatan toko online instan.

## A. Untuk Pemilik Toko (Merchant)

### 1. Memulai (Registrasi)

1.  Buka halaman `http://localhost:8080/register`.
2.  Isi Nama, Email, Password, dan Nama Toko yang diinginkan.
3.  Klik "Daftar".
4.  Atau gunakan tombol **Masuk dengan Google** untuk pendaftaran instan.

### 2. Mengelola Toko (Dashboard)

Setelah login, Anda akan diarahkan ke Dashboard (`/dashboard`).

- **Produk**: Menu untuk menambah produk, upload gambar, mengatur harga dan stok.
- **Toko (Storefront)**: Link toko Anda adalah `http://localhost:8080/s/nama-toko-anda`.
- **Tema**: Buka menu **Tema** untuk mengganti tampilan toko. Klik "Kustomisasi" untuk mengubah warna dan teks.

### 3. Katalog

Halaman toko berfungsi sebagai katalog produk saja. Pengaturan pesanan dan transaksi dikelola oleh Admin Pusat.

---

## B. Untuk Super Admin

### 1. Akses Console Admin

Admin Panel terpisah dari dashboard user.

- URL: `http://localhost:8080/console/login`
- Login menggunakan kredensial Admin.

### 2. Kelola Paket Langganan (Plans)

Fitur untuk mengatur tiering monetisasi platform.

1.  Buka menu **Plans**.
2.  Klik **+ Buat Paket**.
3.  Atur spesifikasi:
    - **Nama Paket**: (misal `Basic`, `Pro`).
    - **Harga**: Bulanan dan Tahunan.
    - **Limit**: Jumlah maksimal toko, produk, dan storage.
    - **Fitur**: Centang "Custom Domain" atau "Tema Premium" jika termasuk dalam paket.
4.  Paket yang dibuat akan menjadi acuan batasan akun user.

### 3. Manajemen User & Toko

- **Users**: Memantau daftar pengguna. Klik "Suspend" untuk memblokir akses pengguna yang melanggar.
- **Stores**: Memantau daftar toko yang dibuat.
- **Analytics**: Melihat ringkasan pertumbuhan platform.

### 4. Manajemen Pesanan (Pusat)

Semua pesanan dari berbagai toko masuk ke menu **Orders** di Super Admin.

1.  Buka menu **Orders**.
2.  Lihat daftar pesanan yang masuk dengan status `Pending` atau `Waiting Confirmation`.
3.  Klik tombol **Update** untuk memproses pesanan (misal konfirmasi pembayaran atau pengiriman).
4.  Status yang diubah admin akan terlihat oleh pembeli.

---

## C. Setup & Konfigurasi (Dev)

### Prasyarat

- Go 1.21+
- PostgreSQL
- Akun Cloudinary (Image Hosting)
- Google Cloud Project (Untuk Google Login)

### Konfigurasi .env

Pastikan file `.env` di root project memiliki variabel berikut:

```env
PORT=8080
DATABASE_URL="host=localhost user=postgres password=root dbname=sello port=5432 sslmode=disable"
JWT_SECRET="ganti_dengan_secret_aman"
SUPABASE_JWT_SECRET="sama_dengan_jwt_secret_jika_tidak_pakai_supabase"

# Cloudinary
CLOUDINARY_CLOUD_NAME="doq..."
CLOUDINARY_API_KEY="736..."
CLOUDINARY_API_SECRET="ABC..."

# Google OAuth
GOOGLE_CLIENT_ID="xxx.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="xxx..."
GOOGLE_REDIRECT_URL="http://localhost:8080/api/auth/google/callback"
```

### Menjalankan Aplikasi

```bash
go run cmd/main.go
```
