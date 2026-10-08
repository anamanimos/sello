// Sello Product Edit JavaScript

let productId = null;
let uploadedImages = []; // { id, url, size, isPrimary, isNew }
let draggedItem = null;

let quill;

document.addEventListener("DOMContentLoaded", () => {
  setupQuill();
  loadProduct();
  setupImageUpload();
  setupDragAndDrop();
  document
    .getElementById("productForm")
    .addEventListener("submit", handleSubmit);
  checkExternalLinkPermission();
});

function setupQuill() {
  quill = new Quill("#quill-editor", {
    theme: "snow",
    modules: {
      toolbar: {
        container: [
          [{ header: [1, 2, 3, false] }],
          ["bold", "italic", "underline", "strike"],
          [{ list: "ordered" }, { list: "bullet" }],
          ["emoji"],
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

async function loadProduct() {
  const pathParts = window.location.pathname.split("/");
  productId = pathParts[3];

  try {
    const product = await apiRequest(`/api/products/${productId}`);
    if (!product || product.error) {
      showToast("Produk tidak ditemukan", "error");
      setTimeout(() => (window.location.href = "/dashboard/products"), 2000);
      return;
    }

    document.getElementById("name").value = product.name || "";
    document.getElementById("price").value = product.price || "";
    document.getElementById("discountPrice").value =
      product.discount_price || "";
    // document.getElementById("description").value = product.description || "";
    quill.root.innerHTML = product.description || ""; // Load to Quill
    document.getElementById("externalLink").value = product.external_link || "";

    // Load categories and set selected
    await loadCategories(product.category_id);

    // Load images
    const images = await apiRequest(`/api/products/${productId}/images`);

    if (images && Array.isArray(images) && images.length > 0) {
      uploadedImages = images.map((img) => ({
        id: img.id,
        url: img.image_url,
        size: img.image_size,
        isPrimary: img.is_primary,
        isNew: false,
      }));
    } else if (product.image_url && product.image_url !== "") {
      // Fallback to single image from product table
      uploadedImages = [
        {
          id: "legacy",
          url: product.image_url,
          size: product.image_size || 0,
          isPrimary: true,
          isNew: false,
        },
      ];
    }

    renderImages();
    document.getElementById("loading").style.display = "none";
    document.getElementById("productForm").style.display = "block";
  } catch (e) {
    showToast("Gagal memuat produk", "error");
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
      showToast(`${file.name} terlalu besar`, "error");
      continue;
    }
    await uploadNewImage(file);
  }
  e.target.value = "";
}

async function uploadNewImage(file) {
  const storeId = getCurrentStoreId();
  if (!storeId) {
    showToast("Pilih toko terlebih dahulu", "error");
    return;
  }

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
      uploadedImages.push({
        id: "new-" + Date.now(),
        url: data.url,
        size: file.size,
        isPrimary: uploadedImages.length === 0,
        isNew: true,
      });
      renderImages();
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
  addBtn.style.display = uploadedImages.length >= 5 ? "none" : "flex";
}

function setPrimary(index) {
  uploadedImages.forEach((img, i) => (img.isPrimary = i === index));
  renderImages();
}

async function removeImage(index) {
  const img = uploadedImages[index];
  const wasPrimary = img.isPrimary;

  // Delete from server if not new
  if (!img.isNew && img.id !== "legacy") {
    await apiRequest(`/api/products/${productId}/images/${img.id}`, {
      method: "DELETE",
    });
  }

  uploadedImages.splice(index, 1);
  if (wasPrimary && uploadedImages.length > 0)
    uploadedImages[0].isPrimary = true;
  renderImages();
}

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
      const [moved] = uploadedImages.splice(fromIndex, 1);
      uploadedImages.splice(toIndex, 0, moved);
      renderImages();
    }
  });
}

async function loadCategories(selectedId) {
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
    if (categories) {
      const select = document.getElementById("categorySelect");
      if (select) {
        select.innerHTML =
          '<option value="">Pilih Kategori</option>' +
          categories
            .map(
              (c) =>
                `<option value="${c.id}" ${c.id === selectedId ? "selected" : ""}>${c.name}</option>`,
            )
            .join("");
      }
    }
  } catch (e) {
    console.error("Failed to load categories", e);
  }
}

async function handleSubmit(e) {
  e.preventDefault();

  // Sync Quill
  const descInput = document.getElementById("description");
  descInput.value = quill.root.innerHTML;

  const btn = document.getElementById("submitBtn");
  btn.disabled = true;
  btn.textContent = "Menyimpan...";

  try {
    const primaryImage = uploadedImages.find((img) => img.isPrimary);

    const price = parseInt(document.getElementById("price").value) || 0;
    let discountPrice = document.getElementById("discountPrice").value;
    discountPrice = discountPrice ? parseInt(discountPrice) : null;

    if (discountPrice !== null && discountPrice >= price) {
      showToast("Harga diskon harus lebih rendah dari harga normal", "error");
      btn.disabled = false;
      btn.textContent = "Simpan Perubahan";
      return;
    }

    const productData = {
      name: document.getElementById("name").value,
      price: price,
      discount_price: discountPrice,
      external_link: document.getElementById("externalLink").value,
      category_id: document.getElementById("categorySelect").value,
      description: document.getElementById("description").value,
    };

    // Only send image_url if there's a valid primary image
    if (primaryImage && primaryImage.url) {
      productData.image_url = primaryImage.url;
    }

    await apiRequest(`/api/products/${productId}`, {
      method: "PUT",
      body: JSON.stringify(productData),
    });

    // Add new images
    for (const img of uploadedImages.filter((i) => i.isNew)) {
      const res = await apiRequest(`/api/products/${productId}/images`, {
        method: "POST",
        body: JSON.stringify({
          image_url: img.url,
          image_size: img.size,
          is_primary: img.isPrimary,
        }),
      });
      if (res && res.error) throw new Error(res.error);
    }

    // Update primary on existing images
    for (const img of uploadedImages.filter(
      (i) => !i.isNew && i.id !== "legacy",
    )) {
      if (img.isPrimary) {
        await apiRequest(
          `/api/products/${productId}/images/${img.id}/primary`,
          { method: "PUT" },
        );
      }
    }

    // Reorder
    const imageIds = uploadedImages
      .filter((i) => !i.isNew && i.id !== "legacy")
      .map((i) => i.id);
    if (imageIds.length > 0) {
      await apiRequest(`/api/products/${productId}/images/reorder`, {
        method: "PUT",
        body: JSON.stringify({ image_ids: imageIds }),
      });
    }

    Swal.fire({
      icon: "success",
      title: "Berhasil!",
      text: "Produk berhasil diupdate",
      showCancelButton: true,
      confirmButtonText: "Ke Daftar Produk",
      cancelButtonText: "Tetap di Halaman Ini",
      confirmButtonColor: "#10b981",
    }).then((result) => {
      if (result.isConfirmed) {
        window.location.href = "/dashboard/products";
      }
    });
  } catch (err) {
    console.error(err);
    showToast(err.message || "Terjadi kesalahan", "error");
  } finally {
    btn.disabled = false;
    btn.textContent = "Simpan Perubahan";
  }
}
