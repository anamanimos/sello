// Sello Dashboard JavaScript

// Check authentication
function checkAuth() {
  const token = localStorage.getItem("sello_token");
  if (!token) {
    window.location.href = "/login";
    return null;
  }
  return token;
}

// API request helper
async function apiRequest(endpoint, options = {}) {
  const token = checkAuth();
  if (!token) return null;

  const response = await fetch(endpoint, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...options.headers,
    },
  });

  if (response.status === 401) {
    localStorage.removeItem("sello_token");
    window.location.href = "/login";
    return null;
  }

  return response.json();
}

// Load dashboard data
async function loadDashboard() {
  try {
    // Load stores
    const stores = await apiRequest("/api/stores");
    if (stores && stores.length > 0) {
      const store = stores[0]; // First store
      updateStoreInfo(store);
      loadStats(store.id); // Load Analytics
      loadProducts(store.id);
    }
  } catch (error) {
    console.error("Failed to load dashboard:", error);
  }
}

// Update store info
function updateStoreInfo(store) {
  // Update header
  const viewStoreBtn = document.querySelector(".header-right a.btn-primary"); // Check selector if needed, usually unrelated.
  // Actually update the "Lihat Toko" link in Quick Actions if exists
  const viewStoreLink = document.getElementById("viewStoreLink");
  if (viewStoreLink) viewStoreLink.href = `/s/${store.slug}`;

  // Check for suspension
  if (store.is_suspended) {
    showSuspensionBanner(store.suspension_reason);
  }
}

function showSuspensionBanner(reason) {
  const banner = document.createElement("div");
  banner.style.cssText = `
        background: #fef2f2;
        border: 1px solid #fca5a5;
        color: #991b1b;
        padding: 16px 24px;
        border-radius: 12px;
        margin-bottom: 24px;
        display: flex;
        align-items: flex-start;
        gap: 12px;
    `;
  banner.innerHTML = `
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="flex-shrink:0; margin-top:2px;">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <div>
            <h3 style="font-size: 16px; font-weight: 600; margin: 0 0 4px 0;">Toko Ditangguhkan (Suspended)</h3>
            <p style="margin: 0; font-size: 14px; opacity: 0.9;">
                Toko Anda saat ini tidak dapat diakses oleh publik.
                <br><strong>Alasan:</strong> ${reason || "Pelanggaran kebijakan layanan."}
            </p>
        </div>
    `;

  // Insert after header
  const header = document.querySelector(".dashboard-header");
  header.parentNode.insertBefore(banner, header.nextSibling);
}

// Load Stats (Analytics & Storage)
async function loadStats(storeId) {
  try {
    const stats = await apiRequest(`/api/stores/${storeId}/stats`);
    if (!stats) return;

    // Update UI
    document.getElementById("productCount").textContent = stats.products;
    document.getElementById("viewCount").textContent = stats.visits;
    document.getElementById("waClickCount").textContent = stats.clicks;

    // Storage
    const usedMB = (stats.storage_used / (1024 * 1024)).toFixed(2);
    const maxMB = (stats.storage_max / (1024 * 1024)).toFixed(0);
    document.getElementById("storageUsed").textContent =
      `${usedMB} / ${maxMB} MB`;

    // Add refresh button to storage card if not exists
    const storageCard = document.getElementById("storageUsed").parentElement;
    if (!storageCard.querySelector(".refresh-btn")) {
      const btn = document.createElement("button");
      btn.className = "refresh-btn";
      btn.innerHTML = "🔄";
      btn.title = "Refresh Storage";
      btn.style.cssText =
        "position:absolute; top:10px; right:10px; background:none; border:none; cursor:pointer; opacity:0.5";
      btn.onclick = (e) => {
        e.preventDefault();
        loadStats(storeId);
        showToast("Storage refreshed", "success");
      };
      // storageCard.style.position = "relative"; // Ensure relative
      storageCard.appendChild(btn);
    }
  } catch (error) {
    console.error("Failed to load stats:", error);
  }
}

// Load products
async function loadProducts(storeId) {
  try {
    const response = await apiRequest(
      `/api/stores/${storeId}/products?limit=5`,
    );
    if (!response) return;

    // Response structure might be {data: [], meta: {}} if using the new handler
    // Or just [] if using the old one. We registered GetProductsByStoreID which returns {data...}

    let products = [];
    if (Array.isArray(response)) {
      products = response;
    } else if (response.data) {
      products = response.data;
    }

    // Render products table
    renderProducts(products);
  } catch (error) {
    console.error("Failed to load products:", error);
  }
}

// Render products in table
function renderProducts(products) {
  const container = document.getElementById("productsTable");
  if (!container) return;

  if (products.length === 0) {
    return; // Keep empty state
  }

  const html = products
    .slice(0, 5)
    .map(
      (product) => `
        <div class="product-row">
            ${
              product.image_url
                ? `<img src="${product.image_url}" alt="${product.name}" class="product-row-img">`
                : `<div class="product-row-img" style="display:flex;align-items:center;justify-content:center;">📦</div>`
            }
            <div class="product-row-info">
                <div class="product-row-name">${product.name}</div>
                <div class="product-row-price">${formatRupiah(product.price)}</div>
            </div>
            <div class="product-row-actions">
                <button class="btn-icon" onclick="editProduct('${product.id}')" title="Edit">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                        <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                    </svg>
                </button>
                <button class="btn-icon danger" onclick="deleteProduct('${product.id}')" title="Hapus">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="3,6 5,6 21,6"/>
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                    </svg>
                </button>
            </div>
        </div>
    `,
    )
    .join("");

  container.innerHTML = html;
}

// Format Rupiah
function formatRupiah(number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(number);
}

// Copy store link
function copyStoreLink() {
  const storeLink = document.querySelector(".header-right a")?.href;
  if (storeLink) {
    navigator.clipboard.writeText(storeLink).then(() => {
      showToast("Link toko berhasil disalin!", "success");
    });
  }
}

// Edit product
function editProduct(id) {
  window.location.href = `/dashboard/products/${id}/edit`;
}

// Delete product
async function deleteProduct(id) {
  const result = await Swal.fire({
    title: "Hapus Produk?",
    text: "Produk yang dihapus tidak dapat dikembalikan",
    icon: "warning",
    showCancelButton: true,
    confirmButtonColor: "#ef4444",
    cancelButtonColor: "#6b7280",
    confirmButtonText: "Ya, Hapus!",
    cancelButtonText: "Batal",
    background: "#1f2937",
    color: "#f9fafb",
  });
  if (!result.isConfirmed) return;

  try {
    const apiResult = await apiRequest(`/api/products/${id}`, {
      method: "DELETE",
    });

    if (apiResult) {
      showToast("Produk berhasil dihapus", "success");
      loadDashboard(); // Reload
    }
  } catch (error) {
    showToast("Gagal menghapus produk", "error");
  }
}

// Logout
async function logout() {
  const result = await Swal.fire({
    title: "Keluar?",
    text: "Anda yakin ingin keluar dari akun ini?",
    icon: "question",
    showCancelButton: true,
    confirmButtonColor: "#ef4444",
    cancelButtonColor: "#6b7280",
    confirmButtonText: "Ya, Keluar",
    cancelButtonText: "Batal",
    background: "#1f2937",
    color: "#f9fafb",
  });
  if (!result.isConfirmed) return;
  localStorage.removeItem("sello_token");
  localStorage.removeItem("sello_user");
  window.location.href = "/login";
}

// Toast notification
function showToast(message, type = "default") {
  const existingToast = document.querySelector(".toast");
  if (existingToast) existingToast.remove();

  const toast = document.createElement("div");
  toast.className = `toast ${type}`;
  toast.textContent = message;
  toast.style.cssText = `
        position: fixed;
        bottom: 24px;
        right: 24px;
        padding: 16px 24px;
        background: ${type === "success" ? "#22c55e" : type === "error" ? "#ef4444" : "#1f2937"};
        color: white;
        border-radius: 12px;
        box-shadow: 0 25px 50px -12px rgba(0,0,0,0.25);
        z-index: 9999;
        animation: slideIn 0.3s ease;
    `;
  document.body.appendChild(toast);

  setTimeout(() => {
    toast.style.animation = "slideOut 0.3s ease";
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

// Initialize
document.addEventListener("DOMContentLoaded", function () {
  checkAuth();
  loadDashboard();
});

// Add CSS animations
const style = document.createElement("style");
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100px); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100px); opacity: 0; }
    }
`;
document.head.appendChild(style);
