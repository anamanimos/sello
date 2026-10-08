/**
 * Initial Data Seeder
 * Run this once to populate localStorage with existing homepage content
 */

const InitialData = {
    features: [
        {
            id: 1,
            icon: "fas fa-pencil-ruler",
            title: "Desain Bebas",
            description: "Request desain sesuka hati atau pilih dari template eksklusif kami secara gratis."
        },
        {
            id: 2,
            icon: "fas fa-medal",
            title: "Bahan Premium",
            description: "Menggunakan bahan DryFit berkualitas yang adem, menyerap keringat, dan nyaman dipakai."
        },
        {
            id: 3,
            icon: "fas fa-truck",
            title: "Pengiriman Cepat",
            description: "Proses produksi 7-14 hari kerja dengan pengiriman ke seluruh Indonesia."
        },
        {
            id: 4,
            icon: "fas fa-hand-holding-usd",
            title: "Harga Bersahabat",
            description: "Kualitas premium dengan harga yang terjangkau untuk semua kalangan."
        }
    ],

    packages: [
        {
            id: 1,
            name: "Paket Starter",
            price: 85000,
            popular: false,
            features: [
                "Jersey Saja",
                "Bahan DryFit Standar",
                "Pilih dari Template",
                "Min. Order 11 pcs"
            ]
        },
        {
            id: 2,
            name: "Paket Juara",
            price: 120000,
            popular: true,
            features: [
                "Jersey + Celana",
                "Bahan DryFit Premium",
                "Custom Desain Bebas",
                "Gratis Nama & Nomor",
                "Min. Order 11 pcs"
            ]
        },
        {
            id: 3,
            name: "Paket Lengkap",
            price: 180000,
            popular: false,
            features: [
                "Jersey + Celana + Kaos Kaki",
                "Bahan DryFit Premium",
                "Custom Desain Bebas",
                "Gratis Nama, Nomor & Logo",
                "Min. Order 11 pcs"
            ]
        }
    ],

    testimonials: [
        {
            id: 1,
            name: "Ahmad Rizki",
            role: "Kapten Tim FC Garuda",
            content: "Jersey-nya keren banget! Desain sesuai request, bahan adem, dan pengiriman cepat. Tim kami makin semangat main!"
        },
        {
            id: 2,
            name: "Siti Fatimah",
            role: "Manager Tim Basket Putri",
            content: "Pelayanan ramah dan responsif. Hasil jersey sangat memuaskan, warna tidak luntur walau dicuci berkali-kali."
        },
        {
            id: 3,
            name: "Budi Santoso",
            role: "Komunitas Sepakbola Jakarta",
            content: "Harga bersahabat, kualitas juara! Jahitan rapi, warna tidak luntur. Pasti order lagi untuk musim depan."
        }
    ],

    seed: function() {
        // Only seed if data doesn't exist
        if (!localStorage.getItem('amerygo_features')) {
            localStorage.setItem('amerygo_features', JSON.stringify(this.features));
            console.log('Features seeded!');
        }
        if (!localStorage.getItem('amerygo_packages')) {
            localStorage.setItem('amerygo_packages', JSON.stringify(this.packages));
            console.log('Packages seeded!');
        }
        if (!localStorage.getItem('amerygo_testimonials')) {
            localStorage.setItem('amerygo_testimonials', JSON.stringify(this.testimonials));
            console.log('Testimonials seeded!');
        }
    }
};

// Auto-seed on load
InitialData.seed();
