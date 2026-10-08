document.addEventListener("DOMContentLoaded", async () => {
  // Wait for adminLayout
  setTimeout(loadCoupons, 500);
});

async function loadCoupons() {
  const storeId = window.getCurrentStoreId();
  if (!storeId) return;

  const tbody = document.getElementById("couponsTableBody");
  tbody.innerHTML =
    '<tr><td colspan="6" class="text-center" style="padding:20px;">Memuat...</td></tr>';

  try {
    const coupons = await apiRequest(`/api/stores/${storeId}/coupons`);
    renderCoupons(coupons);
  } catch (error) {
    console.error(error);
    tbody.innerHTML =
      '<tr><td colspan="6" class="text-center text-red-500">Gagal memuat data</td></tr>';
  }
}

function renderCoupons(coupons) {
  const tbody = document.getElementById("couponsTableBody");
  if (!coupons || coupons.length === 0) {
    tbody.innerHTML =
      '<tr><td colspan="6" class="text-center" style="padding:40px; color:#6b7280;">Belum ada kupon</td></tr>';
    return;
  }

  tbody.innerHTML = coupons
    .map(
      (coupon) => `
        <tr>
            <td style="font-weight:600; font-family:monospace;">${coupon.code}</td>
            <td>
                ${
                  coupon.type === "fixed"
                    ? `Rp ${coupon.value.toLocaleString("id-ID")}`
                    : `${coupon.value}% ${coupon.max_discount > 0 ? `(Max Rp ${coupon.max_discount.toLocaleString("id-ID")})` : ""}`
                }
            </td>
            <td>Rp ${coupon.min_purchase.toLocaleString("id-ID")}</td>
            <td>
                ${coupon.usage_limit > 0 ? `${coupon.used_count} / ${coupon.usage_limit}` : `${coupon.used_count} / ∞`}
            </td>
            <td>${coupon.expires_at ? new Date(coupon.expires_at).toLocaleDateString() : "Selamanya"}</td>
            <td>
                <button onclick="deleteCoupon('${coupon.id}')" class="btn btn-sm" style="background:#fee2e2; color:#dc2626; border:none; padding: 4px 8px; border-radius:4px; cursor:pointer;">Hapus</button>
            </td>
        </tr>
    `,
    )
    .join("");
}

function openCreateModal() {
  document.getElementById("couponForm").reset();
  document.getElementById("maxDiscountGroup").style.display = "none";
  document.getElementById("createModal").classList.add("show");
}

function closeCreateModal() {
  document.getElementById("createModal").classList.remove("show");
}

function toggleMaxDiscount(type) {
  const group = document.getElementById("maxDiscountGroup");
  if (type === "percentage") {
    group.style.display = "block";
  } else {
    group.style.display = "none";
  }
}

async function handleCreate(e) {
  e.preventDefault();

  const storeId = window.getCurrentStoreId();
  const formData = new FormData(e.target);
  const data = {
    code: formData.get("code").toUpperCase(),
    type: formData.get("type"),
    value: parseFloat(formData.get("value")),
    min_purchase: parseFloat(formData.get("min_purchase") || 0),
    max_discount: parseFloat(formData.get("max_discount") || 0),
    usage_limit: parseInt(formData.get("usage_limit") || 0),
    expires_at: formData.get("expires_at") || "",
  };

  try {
    const res = await fetch(`/api/stores/${storeId}/coupons`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("sello_token")}`,
      },
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || "Gagal membuat kupon");
    }

    closeCreateModal();
    loadCoupons();
    showToast("Kupon berhasil dibuat", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function deleteCoupon(id) {
  const confirm = await Swal.fire({
    title: "Hapus Kupon?",
    text: "Kupon yang dihapus tidak bisa dikembalikan",
    icon: "warning",
    showCancelButton: true,
    confirmButtonColor: "#dc2626",
    confirmButtonText: "Hapus",
  });

  if (!confirm.isConfirmed) return;

  try {
    await apiRequest(`/api/coupons/${id}`, { method: "DELETE" });
    loadCoupons();
    showToast("Kupon dihapus", "success");
  } catch (e) {
    showToast("Gagal menghapus kupon", "error");
  }
}
