-- Demo Store Seed Script
-- Run this in PostgreSQL to create demo store with sample data

-- 1. Create demo user
INSERT INTO users (id, email, name, whatsapp, is_phone_verified, is_email_verified, created_at, updated_at)
VALUES (
    'demo0000-0000-0000-0000-000000000001',
    'demo@sello.id',
    'Demo User',
    '628123456789',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 2. Create demo store
INSERT INTO stores (id, user_id, slug, name, description, whatsapp, is_verified, created_at, updated_at)
VALUES (
    'demo0000-0000-0000-0000-000000000002',
    'demo0000-0000-0000-0000-000000000001',
    'demo',
    'Toko Demo Sello',
    'Ini adalah toko demo untuk menunjukkan fitur-fitur Sello. Anda dapat membuat toko seperti ini secara gratis!',
    '628123456789',
    true,
    NOW(),
    NOW()
) ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    whatsapp = EXCLUDED.whatsapp,
    is_verified = EXCLUDED.is_verified,
    updated_at = NOW();

-- 3. Get the store ID for the demo store (in case of conflict)
-- Using demo0000-0000-0000-0000-000000000002

-- 4. Create store theme with starter template
INSERT INTO store_themes (id, store_id, theme_name, is_active, custom_config, created_at, updated_at)
VALUES (
    'demo0000-0000-0000-0000-000000000003',
    'demo0000-0000-0000-0000-000000000002',
    'starter',
    true,
    '{"primary_color": "#22c55e", "products_title": "Produk Unggulan Kami", "cta_button": "Pesan Sekarang"}',
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- 5. Delete existing demo products (for clean re-seed)
DELETE FROM products WHERE store_id = 'demo0000-0000-0000-0000-000000000002';

-- 6. Create 10 demo products with Unsplash images
INSERT INTO products (id, store_id, name, price, discount_price, description, image_url, created_at, updated_at) VALUES
-- Product 1: Electronics
('demop000-0000-0000-0000-000000000001', 'demo0000-0000-0000-0000-000000000002', 
 'Headphone Bluetooth Premium', 450000, 399000, 
 'Headphone wireless dengan kualitas suara jernih dan bass yang dalam. Dilengkapi noise cancelling aktif untuk pengalaman mendengarkan yang imersif.',
 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 2: Fashion
('demop000-0000-0000-0000-000000000002', 'demo0000-0000-0000-0000-000000000002',
 'Jaket Denim Classic', 275000, NULL,
 'Jaket denim berkualitas tinggi dengan potongan classic. Cocok untuk gaya casual sehari-hari.',
 'https://images.unsplash.com/photo-1551028719-00167b16eac5?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 3: Accessories
('demop000-0000-0000-0000-000000000003', 'demo0000-0000-0000-0000-000000000002',
 'Tas Ransel Minimalist', 185000, 165000,
 'Tas ransel dengan desain minimalist dan bahan anti air. Kapasitas laptop hingga 15 inch.',
 'https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 4: Electronics
('demop000-0000-0000-0000-000000000004', 'demo0000-0000-0000-0000-000000000002',
 'Smartwatch Sport Edition', 850000, 799000,
 'Smartwatch dengan fitur lengkap: heart rate monitor, GPS, waterproof, dan battery tahan 7 hari.',
 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 5: Home
('demop000-0000-0000-0000-000000000005', 'demo0000-0000-0000-0000-000000000002',
 'Lampu Meja LED Modern', 125000, NULL,
 'Lampu meja dengan desain modern dan 3 tingkat kecerahan. Hemat energi dan ramah mata.',
 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 6: Fashion
('demop000-0000-0000-0000-000000000006', 'demo0000-0000-0000-0000-000000000002',
 'Sepatu Sneakers Urban', 320000, 289000,
 'Sneakers nyaman untuk aktivitas sehari-hari. Sol karet anti slip dan bahan breathable.',
 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 7: Electronics
('demop000-0000-0000-0000-000000000007', 'demo0000-0000-0000-0000-000000000002',
 'Power Bank 20000mAh', 199000, NULL,
 'Power bank kapasitas besar dengan fast charging 22.5W. Cukup untuk mengisi penuh smartphone 4-5 kali.',
 'https://images.unsplash.com/photo-1609091839311-d5365f9ff1c5?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 8: Accessories
('demop000-0000-0000-0000-000000000008', 'demo0000-0000-0000-0000-000000000002',
 'Kacamata Hitam Polarized', 145000, 120000,
 'Kacamata hitam dengan lensa polarized untuk perlindungan UV maksimal. Frame ringan dan stylish.',
 'https://images.unsplash.com/photo-1572635196237-14b3f281503f?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 9: Electronics
('demop000-0000-0000-0000-000000000009', 'demo0000-0000-0000-0000-000000000002',
 'Speaker Bluetooth Portable', 225000, NULL,
 'Speaker portable dengan suara bass yang kuat. Waterproof IPX7 dan battery tahan 12 jam.',
 'https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?w=400&h=400&fit=crop',
 NOW(), NOW()),

-- Product 10: Fashion
('demop000-0000-0000-0000-000000000010', 'demo0000-0000-0000-0000-000000000002',
 'Topi Baseball Cap', 85000, 75000,
 'Topi baseball dengan bahan katun premium. Adjustable strap untuk kenyamanan maksimal.',
 'https://images.unsplash.com/photo-1588850561407-ed78c282e89b?w=400&h=400&fit=crop',
 NOW(), NOW());

-- Done! Demo store is ready at /s/demo
SELECT 'Demo store created successfully!' as status;
