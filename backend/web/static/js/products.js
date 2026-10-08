// Products Page JS

let currentPage = 1;
let currentLimit = 10;
let searchQuery = "";
let currentCategory = "";
let debounceTimer;

document.addEventListener("DOMContentLoaded", () => {
  loadCategories();
  loadProducts();
});

// Load Categories for Filter
async function loadCategories() {
  const storeId = getCurrentStoreId();
  if (!storeId) return;
  try {
    const categories = await apiRequest(`/api/stores/${storeId}/categories`);
    if (categories && categories.length > 0) {
      const select = document.getElementById("categoryFilter");
      // Keep first option
      const first = select.firstElementChild.outerHTML;
      select.innerHTML =
        first +
        categories
          .map((c) => `<option value="${c.id}">${c.name}</option>`)
          .join("");
    }
  } catch (e) {
    console.error("Failed filter categories", e);
  }
}

// Main Load Function
async function loadProducts() {
  const storeId = getCurrentStoreId();
  if (!storeId) return;

  // Show loading
  const listEl = document.getElementById("productsList");
  // Only show loading text if empty or previous was empty state
  // listEl.innerHTML = '<div class="empty-state"><p>Memuat...</p></div>';

  try {
    const queryParams = new URLSearchParams({
      page: currentPage,
      limit: currentLimit,
      search: searchQuery,
      category_id: currentCategory,
    });

    const res = await apiRequest(
      `/api/stores/${storeId}/products?${queryParams.toString()}`,
    );

    // Handle new response format { data: [], meta: {} }
    const products = res.data || [];
    const meta = res.meta || { page: 1, limit: 10, total: 0, total_pages: 0 };

    renderProducts(products);
    renderPagination(meta);
  } catch (e) {
    console.error(e);
    listEl.innerHTML =
      '<div class="empty-state"><p>Gagal memuat produk</p></div>';
  }
}

function renderProducts(products) {
  const listEl = document.getElementById("productsList");

  if (products.length === 0) {
    listEl.innerHTML =
      '<div class="empty-state"><p>Tidak ada produk ditemukan</p></div>';
    return;
  }

  const placeholder =
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100' viewBox='0 0 24 24' fill='%23f1f5f9' stroke='%2394a3b8' stroke-width='1' stroke-linecap='round' stroke-linejoin='round'%3E%3Crect x='3' y='3' width='18' height='18' rx='2' ry='2'/%3E%3Ccircle cx='8.5' cy='8.5' r='1.5'/%3E%3Cpolyline points='21 15 16 10 5 21'/%3E%3C/svg%3E";

  listEl.innerHTML = products
    .map((p) => {
      const priceDisplay = p.discount_price
        ? `<div class="price-wrapper">
                 <span class="product-price strikethrough">Rp ${p.price.toLocaleString()}</span>
                 <span class="product-price discounted">Rp ${p.discount_price.toLocaleString()}</span>
               </div>`
        : `<div class="product-price">Rp ${p.price.toLocaleString()}</div>`;

      const badge = p.discount_price
        ? `<span class="discount-badge">DISKON</span>`
        : "";

      const categoryLabel = p.category
        ? `<div style="font-size: 11px; color: var(--text-muted); margin-bottom: 2px;">${p.category.name}</div>`
        : "";

      return `
        <div class="product-row">
          ${badge}
          <img src="${p.image_url || placeholder}" class="product-img" alt="" onerror="this.src='${placeholder}'">
          <div class="product-info">
            <div class="product-name">${p.name}</div>
            ${categoryLabel}
            ${priceDisplay}
          </div>
          <div class="product-actions">
            <a href="/dashboard/products/${p.id}/edit" class="btn btn-sm btn-outline">Edit</a>
            <button class="btn btn-sm btn-danger" onclick="deleteProduct('${p.id}')">Hapus</button>
          </div>
        </div>
        `;
    })
    .join("");
}

function renderPagination(meta) {
  const footer = document.getElementById("paginationFooter");
  const controls = document.getElementById("paginationControls");
  const showingStart = document.getElementById("showingStart");
  const showingEnd = document.getElementById("showingEnd");
  const totalItems = document.getElementById("totalItems");

  if (meta.total === 0) {
    footer.style.display = "none";
    return;
  }
  footer.style.display = "flex";

  // Info
  const start = (meta.page - 1) * meta.limit + 1;
  const end = Math.min(meta.page * meta.limit, meta.total);
  showingStart.textContent = start;
  showingEnd.textContent = end;
  totalItems.textContent = meta.total;

  // Controls
  let html = "";

  // Prev
  html += `<button class="page-btn" ${meta.page === 1 ? "disabled" : ""} onclick="changePage(${meta.page - 1})">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
    </button>`;

  // Page Numbers (Simple logic: show all or simplified range? Let's show surrounding)
  const totalPages = meta.total_pages;
  for (let i = 1; i <= totalPages; i++) {
    // Show 1, Last, Current, Current-1, Current+1. Ellipsis otherwise.
    if (
      i === 1 ||
      i === totalPages ||
      (i >= meta.page - 1 && i <= meta.page + 1)
    ) {
      html += `<button class="page-btn ${i === meta.page ? "active" : ""}" onclick="changePage(${i})">${i}</button>`;
    } else if (i === meta.page - 2 || i === meta.page + 2) {
      html += `<span style="padding: 0 4px; color: var(--text-muted);">...</span>`;
    }
  }

  // Next
  html += `<button class="page-btn" ${meta.page === totalPages ? "disabled" : ""} onclick="changePage(${meta.page + 1})">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
    </button>`;

  controls.innerHTML = html;
}

// Handlers
function changePage(page) {
  if (page < 1) return;
  currentPage = page;
  loadProducts();
}

function handleSearch(val) {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    searchQuery = val;
    currentPage = 1; // Reset to page 1
    loadProducts();
  }, 500);
}

function handleCategoryFilter(val) {
  currentCategory = val;
  currentPage = 1;
  loadProducts();
}

function handleLimitChange(val) {
  currentLimit = parseInt(val);
  currentPage = 1;
  loadProducts();
}

async function deleteProduct(id) {
  const result = await Swal.fire({
    title: "Hapus Produk?",
    text: "Tindakan ini tidak dapat dibatalkan!",
    icon: "warning",
    showCancelButton: true,
    confirmButtonColor: "#ef4444",
    confirmButtonText: "Ya, Hapus!",
    cancelButtonText: "Batal",
  });

  if (!result.isConfirmed) return;

  try {
    await apiRequest(`/api/products/${id}`, { method: "DELETE" });
    showToast("Produk berhasil dihapus", "success");
    loadProducts(); // Reload current page
  } catch (e) {
    showToast("Gagal menghapus produk", "error");
  }
}
