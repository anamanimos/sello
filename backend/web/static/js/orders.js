let currentOrder = null;

document.addEventListener("DOMContentLoaded", async () => {
  // Wait for adminLayout to load stores
  setTimeout(loadOrders, 500);
});

async function loadOrders() {
  const storeId = window.getCurrentStoreId();
  if (!storeId) return;

  const tbody = document.getElementById("ordersTableBody");
  tbody.innerHTML =
    '<tr><td colspan="6" class="text-center" style="padding:20px;">Memuat...</td></tr>';

  try {
    const orders = await apiRequest(`/api/stores/${storeId}/orders`);
    renderOrders(orders);
  } catch (error) {
    console.error(error);
    tbody.innerHTML =
      '<tr><td colspan="6" class="text-center text-red-500">Gagal memuat data</td></tr>';
  }
}

function renderOrders(orders) {
  const tbody = document.getElementById("ordersTableBody");
  if (!orders || orders.length === 0) {
    tbody.innerHTML =
      '<tr><td colspan="6" class="text-center" style="padding:40px; color:#6b7280;">Belum ada pesanan</td></tr>';
    return;
  }

  tbody.innerHTML = orders
    .map(
      (order) => `
        <tr>
            <td>
                <div style="font-weight:500;">#${order.id.substring(0, 8)}</div>
                <div style="font-size:12px; color:#6b7280;">${new Date(order.created_at).toLocaleDateString()}</div>
            </td>
            <td>
                <div>${order.customer_name}</div>
                <div style="font-size:12px; color:#6b7280;">${order.customer_phone}</div>
            </td>
            <td style="font-weight:600;">
                Rp ${order.final_amount.toLocaleString("id-ID")}
            </td>
            <td>
                <span class="badge badge-${order.status}">
                    ${formatStatus(order.status)}
                </span>
            </td>
            <td>
                ${
                  order.payment_proof
                    ? `<button onclick="viewOrder('${order.id}')" style="color:var(--accent); text-decoration:underline;">Lihat Bukti</button>`
                    : '<span style="color:#9ca3af;">-</span>'
                }
            </td>
            <td>
                <button onclick="viewOrder('${order.id}')" class="btn btn-sm" style="background:var(--bg-secondary); border:1px solid var(--border-color);">Detail</button>
            </td>
        </tr>
    `,
    )
    .join("");

  // Cache for modal
  window.ordersCache = orders;
}

function formatStatus(status) {
  const map = {
    pending: "Menunggu Bayar",
    waiting_confirmation: "Perlu Konfirmasi",
    paid: "Lunas",
    shipped: "Dikirim",
    completed: "Selesai",
    cancelled: "Dibatalkan",
    expired: "Kedaluwarsa",
  };
  return map[status] || status;
}

function viewOrder(orderId) {
  const order = window.ordersCache.find((o) => o.id === orderId);
  if (!order) return;

  currentOrder = order;

  const modalBody = document.getElementById("modalBody");

  let itemsHtml = (order.items || [])
    .map(
      (item) => `
        <div style="display:flex; justify-content:space-between; padding:8px 0; border-bottom:1px solid #f3f4f6;">
            <div>
                <div style="font-weight:500;">${item.product_name} x${item.quantity}</div>
                ${item.variant_name ? `<div style="font-size:12px; color:#6b7280;">Variant: ${item.variant_name}</div>` : ""}
            </div>
            <div>Rp ${item.subtotal.toLocaleString("id-ID")}</div>
        </div>
    `,
    )
    .join("");

  modalBody.innerHTML = `
        <div style="display:grid; grid-template-columns:1fr 1fr; gap:16px; margin-bottom:16px;">
            <div>
                <label style="font-size:12px; color:#6b7280; display:block;">Pelanggan</label>
                <div style="font-weight:500;">${order.customer_name}</div>
                <div>${order.customer_phone}</div>
                <div>${order.customer_email || "-"}</div>
            </div>
            <div>
                <label style="font-size:12px; color:#6b7280; display:block;">Status</label>
                <div style="margin-bottom:8px;"><span class="badge badge-${order.status}">${formatStatus(order.status)}</span></div>
                <label style="font-size:12px; color:#6b7280;">Metode Bayar</label>
                <div>${(order.payment_method || "").toUpperCase().replace("_", " ")}</div>
            </div>
        </div>
        
        <div style="margin-bottom:16px;">
            <label style="font-size:12px; color:#6b7280; display:block; margin-bottom:4px;">Item Pesanan</label>
            <div style="background:#f9fafb; padding:12px; border-radius:8px;">
                ${itemsHtml}
                <div style="display:flex; justify-content:space-between; margin-top:8px; padding-top:8px; border-top:1px solid #e5e7eb; font-weight:600;">
                    <div>Total</div>
                    <div>Rp ${order.final_amount.toLocaleString("id-ID")}</div>
                </div>
            </div>
        </div>
        
        ${
          order.payment_proof
            ? `
            <div>
                <label style="font-size:12px; color:#6b7280; display:block; margin-bottom:4px;">Bukti Pembayaran</label>
                <a href="${order.payment_proof}" target="_blank">
                    <img src="${order.payment_proof}" class="proof-img" alt="Bukti Transfer">
                </a>
            </div>
        `
            : ""
        }
    `;

  // Show correct buttons
  document.getElementById("btnConfirm").style.display = "none";
  document.getElementById("btnShip").style.display = "none";

  if (order.status === "waiting_confirmation" || order.status === "pending") {
    document.getElementById("btnConfirm").style.display = "block";
  } else if (order.status === "paid") {
    document.getElementById("btnShip").style.display = "block";
  }

  document.getElementById("orderModal").classList.add("show");
}

function closeModal() {
  document.getElementById("orderModal").classList.remove("show");
  currentOrder = null;
}

async function updateStatus(newStatus) {
  if (!currentOrder) return;

  const confirm = await Swal.fire({
    title: "Konfirmasi",
    text: `Ubah status menjadi ${formatStatus(newStatus)}?`,
    icon: "question",
    showCancelButton: true,
    confirmButtonText: "Ya, Ubah",
    confirmButtonColor: "#10b981",
  });

  if (!confirm.isConfirmed) return;

  const storeId = window.getCurrentStoreId();
  try {
    await apiRequest(
      `/api/stores/${storeId}/orders/${currentOrder.id}/status`,
      {
        method: "PUT",
        body: JSON.stringify({ status: newStatus }),
      },
    );

    closeModal();
    loadOrders();
    showToast("Status berhasil diperbarui", "success");
  } catch (e) {
    showToast("Gagal mengubah status", "error");
  }
}
