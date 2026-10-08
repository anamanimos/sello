// Sello Categories JS

let currentStoreId = null;
let allCategories = [];
let searchQuery = "";

document.addEventListener("DOMContentLoaded", () => {
  // Wait for admin-layout to set store ID
  waitForStoreId();

  document
    .getElementById("categoryForm")
    .addEventListener("submit", handleFormSubmit);

  // Close modal on overlay click
  document.getElementById("categoryModal").addEventListener("click", (e) => {
    if (e.target.classList.contains("modal-overlay")) closeModal();
  });
});

function waitForStoreId() {
  const storeId = getCurrentStoreId();
  if (storeId) {
    currentStoreId = storeId;
    loadCategories();
  } else {
    // Retry after header load
    setTimeout(waitForStoreId, 100);
  }
}

async function loadCategories() {
  if (!currentStoreId) return;

  const list = document.getElementById("categoriesList");
  // list.innerHTML = `<div class="empty-state"><p>Memuat data...</p></div>`; // Optional, keeps previous state or specific loader

  try {
    const categories = await apiRequest(
      `/api/stores/${currentStoreId}/categories`,
    );
    allCategories = categories || [];
    renderCategories();
  } catch (e) {
    console.error(e);
    allCategories = [];
    list.innerHTML = `<div class="empty-state"><p style="color:var(--error)">Gagal memuat kategori</p></div>`;
    showToast("Gagal memuat kategori", "error");
  }
}

function handleSearch(query) {
  searchQuery = query.toLowerCase();
  renderCategories();
}

function renderCategories() {
  const list = document.getElementById("categoriesList");

  if (allCategories.length === 0) {
    list.innerHTML = `
        <div class="empty-state">
           <div class="empty-state-icon">
             <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
           </div>
           <h3>Belum ada kategori</h3>
           <p>Buat kategori untuk mengelompokkan produk Anda</p>
           <button class="btn btn-primary btn-sm" onclick="openCreateModal()">Tambah Kategori</button>
        </div>`;
    return;
  }

  const filtered = allCategories.filter(
    (cat) =>
      cat.name.toLowerCase().includes(searchQuery) ||
      cat.slug.toLowerCase().includes(searchQuery),
  );

  if (filtered.length === 0) {
    if (searchQuery) {
      list.innerHTML = `
            <div class="empty-state">
                  <div class="empty-state-icon">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                  </div>
                  <h3>Tidak ditemukan</h3>
                  <p>Tidak ada kategori yang cocok dengan "${searchQuery}"</p>
            </div>`;
    } else {
      list.innerHTML = `...`; // Should not reach here
    }
    return;
  }

  list.innerHTML = filtered
    .map(
      (cat) => `
        <div class="product-row">
            <div class="product-row-img">
               ${cat.name.charAt(0).toUpperCase()}
            </div>
            <div class="product-row-info">
              <div class="product-row-name">${cat.name}</div>
              <div class="product-row-price">${cat.slug}</div>
            </div>
            <div class="product-row-actions">
                 <button class="btn-icon" onclick="openEditModal('${cat.id}', '${cat.name}')" title="Edit">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
                <button class="btn-icon danger" onclick="deleteCategory('${cat.id}', '${cat.name}')" title="Hapus">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3,6 5,6 21,6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                </button>
            </div>
        </div>
    `,
    )
    .join("");
}

function openCreateModal() {
  document.getElementById("modalTitle").textContent = "Tambah Kategori";
  document.getElementById("categoryId").value = "";
  document.getElementById("categoryName").value = "";
  document.getElementById("categoryModal").classList.add("show");
  setTimeout(() => document.getElementById("categoryName").focus(), 100);
}

function openEditModal(id, name) {
  document.getElementById("modalTitle").textContent = "Edit Kategori";
  document.getElementById("categoryId").value = id;
  document.getElementById("categoryName").value = name;
  document.getElementById("categoryModal").classList.add("show");
  setTimeout(() => document.getElementById("categoryName").focus(), 100);
}

function closeModal() {
  document.getElementById("categoryModal").classList.remove("show");
}

async function handleFormSubmit(e) {
  e.preventDefault();
  const btn = e.target.querySelector("button[type=submit]");
  const originalText = btn.textContent;
  btn.disabled = true;
  btn.textContent = "Menyimpan...";

  const id = document.getElementById("categoryId").value;
  const name = document.getElementById("categoryName").value;

  if (!name) {
    btn.disabled = false;
    btn.textContent = originalText;
    return;
  }

  try {
    let res;
    if (id) {
      // Update
      res = await apiRequest(`/api/categories/${id}`, {
        method: "PUT",
        body: JSON.stringify({ name }),
      });
    } else {
      // Create
      res = await apiRequest(`/api/stores/${currentStoreId}/categories`, {
        method: "POST",
        body: JSON.stringify({ name }),
      });
    }

    if (res && res.error) {
      showToast(res.error, "error");
    } else {
      showToast(
        id ? "Kategori diperbarui" : "Kategori berhasil dibuat",
        "success",
      );
      closeModal();
      loadCategories(); // Reload list
    }
  } catch (e) {
    showToast("Gagal menyimpan kategori", "error");
  } finally {
    btn.disabled = false;
    btn.textContent = originalText;
  }
}

async function deleteCategory(id, name) {
  const result = await Swal.fire({
    title: "Hapus Kategori?",
    text: `"${name}" akan dihapus. Produk dalam kategori ini akan menjadi 'Tanpa Kategori'.`,
    icon: "warning",
    showCancelButton: true,
    confirmButtonColor: "#ef4444",
    confirmButtonText: "Ya, Hapus!",
    cancelButtonText: "Batal",
  });

  if (!result.isConfirmed) return;

  try {
    await apiRequest(`/api/categories/${id}`, { method: "DELETE" });
    showToast("Kategori dihapus", "success");
    loadCategories();
  } catch (e) {
    showToast("Gagal menghapus kategori", "error");
  }
}
