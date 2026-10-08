// Sello Product Form JavaScript - Multi Image Support

let uploadedImages = []; // { id, url, size, isPrimary }
let draggedItem = null;

// Initialize
document.addEventListener("DOMContentLoaded", () => {
  setupQuill(); // Init Editor
  setupImageUpload();
  setupDragAndDrop();
  document
    .getElementById("productForm")
    .addEventListener("submit", handleSubmit);
  loadStorageInfo();
  loadCategories();
  checkExternalLinkPermission();
});

let quill;

function setupQuill() {
  quill = new Quill("#quill-editor", {
    theme: "snow",
    modules: {
      toolbar: {
        container: [
          [{ header: [1, 2, 3, false] }],
          ["bold", "italic", "underline", "strike"],
          [{ list: "ordered" }, { list: "bullet" }],
          ["emoji"], // Emoji button
          ["clean"],
        ],
      },
      "emoji-toolbar": true,
      "emoji-textarea": false,
      "emoji-shortname": true,
    },
  });
}

// Check if user's plan allows external links
async function checkExternalLinkPermission() {
  try {
    const token = localStorage.getItem("sello_token");
    const response = await fetch("/api/auth/me", {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await response.json();

    if (data.user && data.user.plan && data.user.plan.allow_external_link) {
      document.getElementById("externalLinkGroup").style.display = "block";
    }
  } catch (e) {
    console.error("Failed to check external link permission", e);
  }
}

function setupImageUpload() {
  const addBtn = document.getElementById("addImageBtn");
  const input = document.getElementById("imageInput");

  addBtn.addEventListener("click", () => input.click());
  input.addEventListener("change", handleFilesSelected);
}

async function handleFilesSelected(e) {
  const files = Array.from(e.target.files);
  if (uploadedImages.length + files.length > 5) {
    showToast("Maksimal 5 gambar", "error");
    return;
  }

  for (const file of files) {
    if (file.size > 5 * 1024 * 1024) {
      showToast(`${file.name} terlalu besar (max 5MB)`, "error");
      continue;
    }
    await uploadImage(file);
  }
  e.target.value = "";
}

async function uploadImage(file) {
  const storeId = getCurrentStoreId();
  if (!storeId) {
    showToast("Pilih toko terlebih dahulu", "error");
    return;
  }

  // Show loading state
  const loadingId = "loading-" + Date.now();
  addLoadingPlaceholder(loadingId);

  try {
    const formData = new FormData();
    formData.append("image", file);

    const token = localStorage.getItem("sello_token");
    const response = await fetch(`/api/stores/${storeId}/upload`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    });

    const data = await response.json();
    removeLoadingPlaceholder(loadingId);

    if (response.ok) {
      const imageData = {
        id: "temp-" + Date.now(),
        url: data.url,
        size: file.size,
        isPrimary: uploadedImages.length === 0,
      };
      uploadedImages.push(imageData);
      renderImages();
      loadStorageInfo();
    } else {
      showToast(data.error || "Upload gagal", "error");
    }
  } catch (err) {
    removeLoadingPlaceholder(loadingId);
    showToast("Upload gagal", "error");
  }
}

function addLoadingPlaceholder(id) {
  const grid = document.getElementById("imagesGrid");
  const addBtn = document.getElementById("addImageBtn");
  const placeholder = document.createElement("div");
  placeholder.id = id;
  placeholder.className = "image-item";
  placeholder.innerHTML =
    '<div style="display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-muted);">⏳</div>';
  grid.insertBefore(placeholder, addBtn);
}

function removeLoadingPlaceholder(id) {
  const el = document.getElementById(id);
  if (el) el.remove();
}

function renderImages() {
  const grid = document.getElementById("imagesGrid");
  const addBtn = document.getElementById("addImageBtn");

  // Remove all image items except add button
  Array.from(grid.children).forEach((child) => {
    if (child.id !== "addImageBtn") child.remove();
  });

  uploadedImages.forEach((img, index) => {
    const item = document.createElement("div");
    item.className = `image-item ${img.isPrimary ? "primary" : ""}`;
    item.draggable = true;
    item.dataset.index = index;
    item.innerHTML = `
      <img src="${img.url}" alt="">
      ${img.isPrimary ? '<span class="primary-badge">Utama</span>' : ""}
      <div class="image-actions">
        ${!img.isPrimary ? `<button type="button" class="image-action-btn" onclick="setPrimary(${index})" title="Jadikan Utama">★</button>` : ""}
        <button type="button" class="image-action-btn delete" onclick="removeImage(${index})" title="Hapus">×</button>
      </div>
    `;
    grid.insertBefore(item, addBtn);
  });

  // Hide add button if max reached
  addBtn.style.display = uploadedImages.length >= 5 ? "none" : "flex";
}

function setPrimary(index) {
  uploadedImages.forEach((img, i) => (img.isPrimary = i === index));
  renderImages();
}

function removeImage(index) {
  const wasPrimary = uploadedImages[index].isPrimary;
  uploadedImages.splice(index, 1);
  if (wasPrimary && uploadedImages.length > 0) {
    uploadedImages[0].isPrimary = true;
  }
  renderImages();
}

// Drag and Drop
function setupDragAndDrop() {
  const grid = document.getElementById("imagesGrid");

  grid.addEventListener("dragstart", (e) => {
    if (e.target.classList.contains("image-item")) {
      draggedItem = e.target;
      e.target.classList.add("dragging");
    }
  });

  grid.addEventListener("dragend", (e) => {
    if (e.target.classList.contains("image-item")) {
      e.target.classList.remove("dragging");
      draggedItem = null;
    }
  });

  grid.addEventListener("dragover", (e) => {
    e.preventDefault();
    const target = e.target.closest(".image-item");
    if (
      target &&
      target !== draggedItem &&
      !target.classList.contains("add-image-btn")
    ) {
      target.classList.add("drag-over");
    }
  });

  grid.addEventListener("dragleave", (e) => {
    const target = e.target.closest(".image-item");
    if (target) target.classList.remove("drag-over");
  });

  grid.addEventListener("drop", (e) => {
    e.preventDefault();
    const target = e.target.closest(".image-item");
    if (target && draggedItem && target !== draggedItem) {
      target.classList.remove("drag-over");
      const fromIndex = parseInt(draggedItem.dataset.index);
      const toIndex = parseInt(target.dataset.index);

      // Reorder array
      const [moved] = uploadedImages.splice(fromIndex, 1);
      uploadedImages.splice(toIndex, 0, moved);
      renderImages();
    }
  });
}

async function loadStorageInfo() {
  const storeId = getCurrentStoreId();
  if (!storeId) return;

  try {
    const data = await apiRequest(`/api/stores/${storeId}/storage`);
    if (data) {
      // Use correct keys from API: storage_used, storage_max
      const usedMB = (data.storage_used / (1024 * 1024)).toFixed(1);
      const percentage = Math.min(
        (data.storage_used / data.storage_max) * 100,
        100,
      );
      document.getElementById("storageUsed").textContent = usedMB;
      document.getElementById("storageFill").style.width = percentage + "%";
    }
  } catch (e) {}
}

async function loadCategories() {
  let storeId = getCurrentStoreId();

  // Retry if storeId is missing (race condition with admin layout)
  if (!storeId) {
    for (let i = 0; i < 10; i++) {
      await new Promise((r) => setTimeout(r, 200));
      storeId = getCurrentStoreId();
      if (storeId) break;
    }
  }

  if (!storeId) return;

  try {
    const categories = await apiRequest(`/api/stores/${storeId}/categories`);
    if (categories && categories.length > 0) {
      const select = document.getElementById("categorySelect");
      select.innerHTML =
        '<option value="">Pilih Kategori</option>' +
        categories
          .map((c) => `<option value="${c.id}">${c.name}</option>`)
          .join("");
    }
  } catch (e) {
    console.error("Failed to load categories", e);
  }
}

async function handleSubmit(e) {
  e.preventDefault();

  // Sync Quill to input
  const descInput = document.getElementById("description");
  descInput.value = quill.root.innerHTML;

  if (!document.getElementById("name").value.trim()) {
    showToast("Nama produk wajib diisi", "error");
    return;
  }

  const btn = document.getElementById("submitBtn");
  btn.disabled = true;
  btn.textContent = "Menyimpan...";

  try {
    const storeId = getCurrentStoreId();
    if (!storeId) {
      showToast("Pilih toko terlebih dahulu", "error");
      btn.disabled = false;
      btn.textContent = "Simpan Produk";
      return;
    }

    // Get primary image for backward compat
    const primaryImage = uploadedImages.find((img) => img.isPrimary);

    const price = parseInt(document.getElementById("price").value) || 0;
    let discountPrice = document.getElementById("discountPrice").value;
    discountPrice = discountPrice ? parseInt(discountPrice) : null;

    if (discountPrice !== null && discountPrice >= price) {
      showToast("Harga diskon harus lebih rendah dari harga normal", "error");
      btn.disabled = false;
      btn.textContent = "Simpan Produk";
      return;
    }

    const productData = {
      name: document.getElementById("name").value,
      price: price,
      discount_price: discountPrice,
      external_link: document.getElementById("externalLink").value,
      category_id: document.getElementById("categorySelect").value,
      description: document.getElementById("description").value,
      image_url: primaryImage?.url || "",
      image_size: primaryImage?.size || 0,
      images: uploadedImages.map((img, index) => ({
        image_url: img.url,
        image_size: img.size,
        is_primary: img.isPrimary,
        sort_order: index,
      })),
    };
    const result = await apiRequest(`/api/stores/${storeId}/products`, {
      method: "POST",
      body: JSON.stringify(productData),
    });

    if (result && result.id) {
      // Add all images to product
      for (let i = 0; i < uploadedImages.length; i++) {
        const img = uploadedImages[i];
        const res = await apiRequest(`/api/products/${result.id}/images`, {
          method: "POST",
          body: JSON.stringify({
            image_url: img.url,
            image_size: img.size,
            is_primary: img.isPrimary,
            sort_order: i,
          }),
        });
        if (res && res.error) throw new Error(res.error);
      }

      showToast("Produk berhasil ditambahkan!", "success");
      setTimeout(() => (window.location.href = "/dashboard/products"), 1500);
    } else {
      showToast(result?.error || "Gagal menyimpan", "error");
    }
  } catch (err) {
    console.error(err);
    showToast(err.message || "Terjadi kesalahan", "error");
  } finally {
    btn.disabled = false;
    btn.textContent = "Simpan Produk";
  }
}
