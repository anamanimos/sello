#!/bin/bash

# Script Migrasi Database dari Supabase ke Local VPS
# Simpan sebagai migrate_supabase.sh dan beri izin eksekusi (chmod +x)

echo "=== Sello Database Migration Tool ==="
echo "Tools ini akan memindahkan data dari Supabase (Cloud) ke Database Lokal VPS."
echo ""

# 1. Input Supabase Connection String
echo "Masukkan Connection String SUPABASE (Transaction Mode):"
echo "Format: postgres://postgres.[ref]:[password]@[host]:6543/postgres"
read -r SUPABASE_URL

if [ -z "$SUPABASE_URL" ]; then
    echo "Error: URL Supabase tidak boleh kosong."
    exit 1
fi

# 2. Input Local Connection String
echo ""
echo "Masukkan Connection String LOCAL (Default: postgres://sello_user:sello123@localhost:5432/sello):"
read -r LOCAL_URL

if [ -z "$LOCAL_URL" ]; then
    LOCAL_URL="postgres://sello_user:sello123@localhost:5432/sello"
fi

echo ""
echo "Starting Migration..."
echo "1. Dumping data from Supabase (Schema: public)..."

# Dump hanya schema public, data-only atau schema+data?
# Kita asumsikan full schema+data untuk public, tapi hati-hati conflict create table.
# Sebaiknya gunakan --clean --if-exists untuk menimpa db lokal.
pg_dump "$SUPABASE_URL" -n public --clean --if-exists --no-owner --no-acl -f supabase_dump.sql

if [ $? -ne 0 ]; then
    echo "Error: Gagal melakukan pg_dump. Cek koneksi internet atau URL Supabase."
    exit 1
fi

echo "Dump Selesai (supabase_dump.sql)."
echo "2. Restoring to Local Database..."

psql "$LOCAL_URL" < supabase_dump.sql

if [ $? -ne 0 ]; then
    echo "Error: Gagal melakukan restore ke database lokal."
    exit 1
fi

echo ""
echo "=== MIGRASI SUKSES ==="
echo "Data telah dipindahkan. Jangan lupa update .env Anda untuk menggunakan DB_HOST=localhost."
rm supabase_dump.sql
echo "File dump sementara telah dihapus."
