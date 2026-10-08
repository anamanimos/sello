const app = {
    products: [],
    
    init: async function() {
        this.loadSettings();
        await this.loadProducts();
        
        if (document.getElementById('catalog-container')) {
            this.renderCatalog();
            this.setupFilters();
        }
        
        this.loadDynamicContent();
        this.setupMobileMenu();
        this.setupHeaderScroll();
        this.applyBranding();
    },

    settings: {},

    loadSettings: function() {
        const stored = localStorage.getItem('amerygo_settings');
        if (stored) {
            this.settings = JSON.parse(stored);
        }
    },

    applyBranding: function() {
        // Apply Logo or Title
        const logos = document.querySelectorAll('.logo');
        logos.forEach(logo => {
            if (this.settings.logo) {
                // Show Logo Image
                logo.innerHTML = `<img src="${this.settings.logo}" alt="${this.settings.siteTitle || 'Amerygo'}" style="height: 40px; object-fit: contain;">`;
                logo.classList.remove('logo'); 
                logo.style.display = 'flex';
                logo.style.alignItems = 'center';
            } else {
                // Show Site Title (Text)
                const title = this.settings.siteTitle || 'Amerygo';
                logo.innerHTML = title;
                logo.style.fontFamily = "'Unbounded', sans-serif";
                logo.style.fontSize = "1.5rem";
                logo.style.fontWeight = "700";
                logo.style.color = "var(--primary-color)";
                logo.style.textDecoration = "none";
            }
        });

        // Apply WhatsApp
        if (this.settings.whatsapp) {
            // Update all WA links that match the standard format
            const waLinks = document.querySelectorAll('a[href^="https://wa.me/"]');
            waLinks.forEach(link => {
                const oldHref = link.href;
                // Replace the number part (assumes format https://wa.me/NUMBER?text=...)
                const url = new URL(oldHref);
                // Construct new path: /newNumber
                url.pathname = '/' + this.settings.whatsapp; 
                link.href = url.toString();
            });
        }
    },

    loadProducts: async function() {
        // Try LocalStorage first (Admin updates)
        const stored = localStorage.getItem('amerygo_products');
        if (stored) {
            this.products = JSON.parse(stored);
        } else {
            // Fallback to JSON file
            try {
                const response = await fetch('assets/data/products.json');
                this.products = await response.json();
                // Sync with LS so admin has data to edit
                localStorage.setItem('amerygo_products', JSON.stringify(this.products));
            } catch (e) {
                console.error('Failed to load products', e);
                this.products = [];
            }
        }
    },

    renderCatalog: function(category = 'all') {
        const container = document.getElementById('catalog-container');
        if (!container) return;

        container.innerHTML = '';
        
        const filtered = category === 'all' 
            ? this.products 
            : this.products.filter(p => p.category === category);

        if (filtered.length === 0) {
            container.innerHTML = '<p class="text-center" style="grid-column: 1/-1;">Tidak ada produk di kategori ini.</p>';
            return;
        }

        filtered.forEach(p => {
            const card = document.createElement('div');
            card.className = 'product-card';
            card.setAttribute('data-category', p.category);
            card.onclick = () => window.location.href = `detail.html?id=${p.id}`;
            card.style.cursor = 'pointer';

            const imgHtml = p.image 
                ? `<img src="${p.image}" alt="${p.name}" style="width:100%; height:100%; object-fit:cover;">`
                : `<div class="placeholder-img">${p.name}</div>`;

            // Price HTML with discount support
            let priceHtml;
            if (p.discountPrice && p.discountPrice < p.price) {
                priceHtml = `
                    <p class="price" style="color:var(--primary-color); font-weight:700;">
                        <span style="text-decoration: line-through; color: #999; font-weight: normal; font-size: 0.9rem;">Rp ${p.price.toLocaleString('id-ID')}</span>
                        Rp ${p.discountPrice.toLocaleString('id-ID')}
                    </p>`;
            } else {
                priceHtml = `<p class="price" style="color:var(--primary-color); font-weight:700;">Rp ${p.price.toLocaleString('id-ID')}</p>`;
            }

            card.innerHTML = `
                <div class="product-image">
                    ${imgHtml}
                </div>
                <div class="product-info">
                    <h3>${p.name}</h3>
                    <p class="category">${this.capitalize(p.category)}</p>
                    ${priceHtml}
                </div>
            `;
            container.appendChild(card);
        });
    },

    setupFilters: function() {
        const btns = document.querySelectorAll('.filter-btn');
        btns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                btns.forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
                this.renderCatalog(e.target.dataset.filter);
            });
        });
    },

    setupMobileMenu: function() {
        const mobileBtn = document.querySelector('.mobile-menu-btn');
        if(!mobileBtn) return;
        
        const navLinks = document.querySelector('.nav-links');
        const menuIcon = mobileBtn.querySelector('i');

        mobileBtn.addEventListener('click', function() {
            navLinks.classList.toggle('active');
            if (navLinks.classList.contains('active')) {
                menuIcon.classList.remove('fa-bars');
                menuIcon.classList.add('fa-times');
            } else {
                menuIcon.classList.remove('fa-times');
                menuIcon.classList.add('fa-bars');
            }
        });
    },

    setupHeaderScroll: function() {
        const header = document.querySelector('.navbar');
        if(!header) return;
        
        window.addEventListener('scroll', function() {
            if (window.scrollY > 50) {
                header.style.backgroundColor = 'rgba(26, 26, 26, 0.98)';
                header.style.padding = '15px 0';
            } else {
                header.style.backgroundColor = 'rgba(26, 26, 26, 0.95)';
                header.style.padding = '20px 0';
            }
        });
    },

    // Dynamic Content Loading
    loadDynamicContent: function() {
        this.renderFeatures();
        this.renderPackages();
        this.renderTestimonials();
    },

    renderFeatures: function() {
        const container = document.getElementById('features-container');
        if (!container) return;
        
        const features = JSON.parse(localStorage.getItem('amerygo_features') || '[]');
        if (features.length === 0) return; // Keep static content if no data
        
        container.innerHTML = features.map(f => `
            <div class="feature-card">
                <div class="icon-box"><i class="${f.icon}"></i></div>
                <h3>${f.title}</h3>
                <p>${f.description}</p>
            </div>
        `).join('');
    },

    renderPackages: function() {
        const container = document.getElementById('packages-container');
        if (!container) return;
        
        const packages = JSON.parse(localStorage.getItem('amerygo_packages') || '[]');
        if (packages.length === 0) return;
        
        container.innerHTML = packages.map(p => `
            <div class="package-card ${p.popular ? 'popular' : ''}">
                ${p.popular ? '<div class="badge">Terlaris</div>' : ''}
                <div class="package-header">
                    <h3>${p.name}</h3>
                    <div class="price">Rp ${p.price.toLocaleString('id-ID')}<span>/pcs</span></div>
                </div>
                <ul class="package-features">
                    ${p.features.map(f => `<li><i class="fas fa-check"></i> ${f}</li>`).join('')}
                </ul>
                <a href="https://wa.me/${this.settings.whatsapp || '6281234567890'}?text=${encodeURIComponent('Halo, saya tertarik dengan ' + p.name)}" class="btn ${p.popular ? 'btn-primary' : 'btn-outline-dark'}">Pesan Sekarang</a>
            </div>
        `).join('');
    },

    renderTestimonials: function() {
        const container = document.getElementById('testimonials-container');
        if (!container) return;
        
        const testimonials = JSON.parse(localStorage.getItem('amerygo_testimonials') || '[]');
        if (testimonials.length === 0) return;
        
        container.innerHTML = testimonials.map(t => `
            <div class="testimonial-card">
                <div class="testimonial-content">
                    <i class="fas fa-quote-left quote-icon"></i>
                    <p>${t.content}</p>
                </div>
                <div class="testimonial-author">
                    <div class="author-avatar"><i class="fas fa-user"></i></div>
                    <div class="author-info">
                        <strong>${t.name}</strong>
                        <span>${t.role}</span>
                    </div>
                </div>
            </div>
        `).join('');
    },

    capitalize: function(s) {
        return s.charAt(0).toUpperCase() + s.slice(1);
    }
};

// Global function for detail page to use
async function loadProductDetail(id) {
    await app.loadProducts(); // Ensure data is loaded
    
    // Convert id to number for comparison
    const product = app.products.find(p => p.id == id);
    
    const loader = document.getElementById('product-detail-loader');
    const content = document.getElementById('product-content');
    
    if (loader) loader.style.display = 'none';
    
    if (!product) {
        if(content) content.innerHTML = '<div class="text-center"><h3>Produk tidak ditemukan</h3><a href="index.html" class="btn btn-primary">Kembali ke Katalog</a></div>';
        if(content) content.style.display = 'block';
        return;
    }

    if(content) content.style.display = 'grid'; // Restore grid display for detail container

    // Populate data
    document.getElementById('product-title').textContent = product.name;
    document.getElementById('product-price').textContent = product.price.toLocaleString('id-ID');
    document.getElementById('product-desc').textContent = product.description;
    document.getElementById('breadcrumb-current').textContent = product.name;
    
    const img = document.getElementById('main-product-img');
    if(img) img.src = product.image || 'assets/img/placeholder.jpg';

    // Update WhatsApp Link
    const waBtn = document.getElementById('whatsapp-btn');
    if(waBtn) {
        const text = `Halo Amerygo, saya mau order ${product.name}`;
        waBtn.href = `https://wa.me/6281234567890?text=${encodeURIComponent(text)}`;
    }
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    app.init();
});
