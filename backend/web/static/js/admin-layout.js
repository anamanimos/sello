// Sello Admin Layout Component
// Reusable sidebar with store switcher, navigation, and theme toggle

// Sello Admin Layout Component
// Reusable sidebar with store switcher, navigation, and theme toggle

if (typeof AdminLayout === "undefined") {
  class AdminLayout {
    constructor() {
      this.stores = [];
      this.currentStoreId = localStorage.getItem("sello_current_store");
      this.theme = localStorage.getItem("sello_theme") || "light";
      this.currentPage = this.detectCurrentPage();
      this.dropdownOpen = false;
    }

    detectCurrentPage() {
      const path = window.location.pathname;
      if (path === "/dashboard") return "dashboard";
      if (path.includes("/stores")) return "stores";
      if (path.includes("/products")) return "products";
      if (path.includes("/orders")) return "orders";
      if (path.includes("/coupons")) return "coupons";
      if (path.includes("/categories")) return "categories";
      if (path.includes("/statistics")) return "statistics";
      if (path.includes("/theme")) return "theme";
      if (path.includes("/billing")) return "billing";
      if (path.includes("/profile")) return "profile";
      if (path.includes("/settings")) return "settings";
      return "dashboard";
    }

    async init() {
      if (this.theme === "dark") {
        document.documentElement.setAttribute("data-theme", "dark");
      }
      this.renderSidebar();
      this.checkBetaMode(); // Async check
      await this.loadStores();
      this.setupDropdown();
    }

    async checkBetaMode() {
      try {
        const res = await apiRequest("/api/auth/me");
        if (res && res.is_beta) {
          console.log("Beta Mode Active");
          // Hide Billing Link
          const style = document.createElement("style");
          style.textContent = `
            a[href="/dashboard/billing"] { display: none !important; }
            .nav-item[href="/dashboard/billing"] { display: none !important; }
        `;
          document.head.appendChild(style);

          // Redirect if on billing page
          if (window.location.pathname === "/dashboard/billing") {
            window.location.href = "/dashboard";
          }
        }
      } catch (e) {
        console.error("Failed to check beta status", e);
      }
    }

    async loadStores() {
      try {
        const token = localStorage.getItem("sello_token");
        if (!token) return;

        const response = await fetch("/api/stores", {
          headers: { Authorization: `Bearer ${token}` },
        });

        if (response.ok) {
          this.stores = (await response.json()) || [];
          this.updateStoreSelector();
        }
      } catch (e) {
        console.error("Failed to load stores:", e);
      } finally {
        // Redirect if no stores (or error) and on restricted page
        if (this.stores.length === 0) {
          const path = window.location.pathname;
          const allowed = [
            "/dashboard/stores",
            "/dashboard/profile",
            "/dashboard/billing",
            "/onboarding",
          ];
          // Check if current path starts with any allowed path
          const isAllowed = allowed.some((p) => path.startsWith(p));

          if (!isAllowed) {
            console.warn("No stores found, redirecting to store management");
            document.body.style.opacity = "0"; // Fade out / Hide
            window.location.href = "/dashboard/stores";
          }
        }
      }
    }

    updateStoreSelector() {
      if (this.stores.length === 0) return;

      // Set current store if not set
      if (!this.currentStoreId && this.stores.length > 0) {
        this.currentStoreId = this.stores[0].id;
        localStorage.setItem("sello_current_store", this.currentStoreId);
      }

      const currentStore =
        this.stores.find((s) => s.id === this.currentStoreId) || this.stores[0];
      document.getElementById("selectedStoreName").textContent =
        currentStore.name;
      document.getElementById("storeSelector").style.display = "block";

      this.updateFavicon(currentStore.favicon_url);
      this.renderStoreList();
    }

    updateFavicon(url) {
      let link = document.querySelector("link[rel~='icon']");
      if (!link) {
        link = document.createElement("link");
        link.rel = "icon";
        document.head.appendChild(link);
      }
      link.href = url || "/static/images/sello-icon.ico";
    }

    renderStoreList(filter = "") {
      const listEl = document.getElementById("storeDropdownList");
      if (!listEl) return;

      const filtered = filter
        ? this.stores.filter((s) =>
            s.name.toLowerCase().includes(filter.toLowerCase()),
          )
        : this.stores;

      listEl.innerHTML = filtered
        .map(
          (s) => `
      <div class="store-option ${s.id === this.currentStoreId ? "selected" : ""}" onclick="adminLayout.selectStore('${s.id}')">
        <span class="store-option-name">${s.name}</span>
        ${s.id === this.currentStoreId ? '<span class="store-option-check">✓</span>' : ""}
      </div>
    `,
        )
        .join("");

      if (filtered.length === 0) {
        listEl.innerHTML =
          '<div class="store-option-empty">Tidak ditemukan</div>';
      }
    }

    setupDropdown() {
      // Close dropdown when clicking outside
      document.addEventListener("click", (e) => {
        const dropdown = document.getElementById("storeDropdown");
        const trigger = document.getElementById("storeSelectorTrigger");
        if (
          dropdown &&
          trigger &&
          !dropdown.contains(e.target) &&
          !trigger.contains(e.target)
        ) {
          this.closeDropdown();
        }
      });
    }

    toggleDropdown() {
      this.dropdownOpen = !this.dropdownOpen;
      const dropdown = document.getElementById("storeDropdown");
      const trigger = document.getElementById("storeSelectorTrigger");
      if (dropdown) {
        dropdown.classList.toggle("open", this.dropdownOpen);
        if (this.dropdownOpen) {
          document.getElementById("storeSearchInput").focus();
          this.renderStoreList();
        }
      }
      if (trigger) trigger.classList.toggle("open", this.dropdownOpen);
    }

    closeDropdown() {
      this.dropdownOpen = false;
      const dropdown = document.getElementById("storeDropdown");
      const trigger = document.getElementById("storeSelectorTrigger");
      if (dropdown) dropdown.classList.remove("open");
      if (trigger) trigger.classList.remove("open");
      const input = document.getElementById("storeSearchInput");
      if (input) input.value = "";
    }

    selectStore(storeId) {
      this.currentStoreId = storeId;
      localStorage.setItem("sello_current_store", storeId);
      this.closeDropdown();
      window.location.reload();
    }

    filterStores(query) {
      this.renderStoreList(query);
    }

    toggleTheme() {
      const isDark = this.theme === "dark";
      this.theme = isDark ? "light" : "dark";
      localStorage.setItem("sello_theme", this.theme);

      if (this.theme === "dark") {
        document.documentElement.setAttribute("data-theme", "dark");
      } else {
        document.documentElement.removeAttribute("data-theme");
      }

      const toggle = document.getElementById("themeToggle");
      if (toggle) toggle.classList.toggle("active", this.theme === "dark");
    }

    async logout() {
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
      localStorage.clear();
      window.location.href = "/login";
    }

    renderSidebar() {
      const sidebarEl = document.getElementById("sidebar");
      if (!sidebarEl) return;

      sidebarEl.innerHTML = `
      <style>
        .store-selector-wrapper { padding: 12px 16px; border-bottom: 1px solid var(--border-light); position: relative; }
        .store-selector-trigger { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: var(--radius); cursor: pointer; transition: var(--transition); }
        .store-selector-trigger:hover { border-color: var(--text-muted); }
        .store-selector-trigger.open { border-color: var(--accent); }
        .store-selector-trigger .selected-store { flex: 1; font-size: 13px; font-weight: 500; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .store-selector-trigger .chevron { color: var(--text-muted); transition: transform 0.2s; }
        .store-selector-trigger.open .chevron { transform: rotate(180deg); }
        .store-dropdown { position: absolute; top: 100%; left: 16px; right: 16px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: var(--radius); box-shadow: var(--shadow-lg); z-index: 1000; display: none; margin-top: 4px; }
        .store-dropdown.open { display: block; }
        .store-search { padding: 8px; border-bottom: 1px solid var(--border-light); }
        .store-search input { width: 100%; padding: 8px 10px; font-size: 13px; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-secondary); color: var(--text-primary); }
        .store-search input:focus { outline: none; border-color: var(--accent); }
        .store-search input::placeholder { color: var(--text-muted); }
        .store-dropdown-list { max-height: 200px; overflow-y: auto; }
        .store-option { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; cursor: pointer; transition: var(--transition); }
        .store-option:hover { background: var(--bg-secondary); }
        .store-option.selected { background: var(--accent-light); }
        .store-option-name { font-size: 13px; color: var(--text-primary); }
        .store-option-check { color: var(--accent); font-size: 12px; }
        .store-option-empty { padding: 12px; text-align: center; color: var(--text-muted); font-size: 13px; }
        .store-dropdown-footer { padding: 8px; border-top: 1px solid var(--border-light); }
        .store-dropdown-footer a { display: flex; align-items: center; justify-content: center; gap: 6px; padding: 8px; background: var(--bg-secondary); border-radius: 6px; color: var(--text-secondary); font-size: 13px; font-weight: 500; text-decoration: none; transition: var(--transition); }
        .store-dropdown-footer a:hover { background: var(--bg-tertiary); color: var(--text-primary); }
      </style>
      <aside class="sidebar">
        <div class="sidebar-header">
          <a href="/" class="logo">
            <img src="/static/images/sello-logo.webp" alt="Sello" height="32">
          </a>
        </div>

        <div class="store-selector-wrapper" id="storeSelector" style="display:none;">
          <div class="store-selector-trigger" id="storeSelectorTrigger" onclick="adminLayout.toggleDropdown()">
            <span class="selected-store" id="selectedStoreName">Pilih Toko</span>
            <svg class="chevron" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6,9 12,15 18,9"/></svg>
          </div>
          <div class="store-dropdown" id="storeDropdown">
            <div class="store-search">
              <input type="text" id="storeSearchInput" placeholder="Cari toko..." oninput="adminLayout.filterStores(this.value)">
            </div>
            <div class="store-dropdown-list" id="storeDropdownList"></div>
            <div class="store-dropdown-footer">
              <a href="/dashboard/stores">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2"/></svg>
                Kelola Toko
              </a>
            </div>
          </div>
        </div>
        
        ${
          !["stores", "profile", "billing"].includes(this.currentPage)
            ? `
        <nav class="sidebar-nav">
          <a href="/dashboard" class="nav-item ${this.currentPage === "dashboard" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <rect x="3" y="3" width="7" height="7" rx="1"/>
              <rect x="14" y="3" width="7" height="7" rx="1"/>
              <rect x="3" y="14" width="7" height="7" rx="1"/>
              <rect x="14" y="14" width="7" height="7" rx="1"/>
            </svg>
            <span>Dashboard</span>
          </a>
          <a href="/dashboard/products" class="nav-item ${this.currentPage === "products" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
              <polyline points="3.27,6.96 12,12.01 20.73,6.96"/>
              <line x1="12" y1="22.08" x2="12" y2="12"/>
            </svg>
            <span>Produk</span>
          </a>

          <a href="/dashboard/categories" class="nav-item ${this.currentPage === "categories" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <rect x="3" y="3" width="7" height="7"></rect>
              <rect x="14" y="3" width="7" height="7"></rect>
              <rect x="14" y="14" width="7" height="7"></rect>
              <rect x="3" y="14" width="7" height="7"></rect>
            </svg>
            <span>Kategori</span>
          </a>
          <a href="/dashboard/statistics" class="nav-item ${this.currentPage === "statistics" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M18 20V10"/>
              <path d="M12 20V4"/>
              <path d="M6 20v-6"/>
            </svg>
            <span>Statistik</span>
          </a>
          <a href="/dashboard/theme" class="nav-item ${this.currentPage === "theme" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="13.5" cy="6.5" r=".5" fill="currentColor"/>
              <circle cx="17.5" cy="10.5" r=".5" fill="currentColor"/>
              <circle cx="8.5" cy="7.5" r=".5" fill="currentColor"/>
              <circle cx="6.5" cy="12.5" r=".5" fill="currentColor"/>
              <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.555C21.965 6.012 17.461 2 12 2z"/>
            </svg>
            <span>Tema</span>
          </a>
          <a href="/dashboard/settings" class="nav-item ${this.currentPage === "settings" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            <span>Pengaturan</span>
          </a>
        </nav>`
            : ""
        }

        <div style="margin: 10px 16px; border-top: 1px solid var(--border-light);"></div>
        
        <nav class="sidebar-nav">
          <div style="padding: 0 12px; margin-bottom: 8px; font-size: 11px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px;">Akun</div>
          <a href="/dashboard/profile" class="nav-item ${this.currentPage === "profile" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <span>Profil Saya</span>
          </a>
          <a href="/dashboard/billing" class="nav-item ${this.currentPage === "billing" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <rect x="2" y="5" width="20" height="14" rx="2" />
              <line x1="2" y1="10" x2="22" y2="10" />
            </svg>
            <span>Billing & Paket</span>
          </a>
        </nav>
        
        <div class="sidebar-footer">
          <div class="theme-toggle">
            <div class="theme-toggle-label">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" width="18" height="18">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
              </svg>
              <span>Dark Mode</span>
            </div>
            <div class="toggle-switch ${this.theme === "dark" ? "active" : ""}" id="themeToggle" onclick="adminLayout.toggleTheme()"></div>
          </div>
          <button onclick="adminLayout.logout()" class="logout-btn">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
              <polyline points="16,17 21,12 16,7"/>
              <line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
            <span>Keluar</span>
          </button>
        </div>
      </aside>
    `;
    }

    renderBottomNav() {
      // Create bottom nav element
      const bottomNav = document.createElement("nav");
      bottomNav.className = "bottom-nav";
      bottomNav.innerHTML = `
      <a href="/dashboard" class="bottom-nav-item ${this.currentPage === "dashboard" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <rect x="3" y="3" width="7" height="7" rx="1"/>
          <rect x="14" y="3" width="7" height="7" rx="1"/>
          <rect x="3" y="14" width="7" height="7" rx="1"/>
          <rect x="14" y="14" width="7" height="7" rx="1"/>
        </svg>
        <span>Home</span>
      </a>
      <a href="/dashboard/stats" class="bottom-nav-item ${this.currentPage === "stats" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <line x1="18" y1="20" x2="18" y2="10"/>
          <line x1="12" y1="20" x2="12" y2="4"/>
          <line x1="6" y1="20" x2="6" y2="14"/>
        </svg>
        <span>Statistik</span>
      </a>
      <a href="/dashboard/products" class="bottom-nav-item center-btn ${this.currentPage === "products" ? "active" : ""}">
        <div class="center-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            <polyline points="3.27,6.96 12,12.01 20.73,6.96"/>
            <line x1="12" y1="22.08" x2="12" y2="12"/>
          </svg>
        </div>
        <span>Produk</span>
      </a>
      <a href="/dashboard/stores" class="bottom-nav-item ${this.currentPage === "stores" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
          <polyline points="9,22 9,12 15,12 15,22"/>
        </svg>
        <span>Toko</span>
      </a>
      <a href="/dashboard/settings" class="bottom-nav-item ${this.currentPage === "settings" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <circle cx="12" cy="12" r="3"/>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
        </svg>
        <span>Setting</span>
      </a>
    `;
      document.body.appendChild(bottomNav);
    }

    renderMobileHeader() {
      const header = document.createElement("header");
      header.className = "mobile-header";
      header.innerHTML = `
      <a href="/dashboard" class="mobile-logo">
        <img src="/static/images/sello-logo.webp" alt="Sello" height="32">
      </a>
      <div class="mobile-header-actions">
        <div class="theme-btn" onclick="adminLayout.toggleTheme()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" width="20" height="20">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
        </div>
      </div>
    `;
      document.body.insertBefore(header, document.body.firstChild);
    }
  }
  // Initialize
  window.AdminLayout = AdminLayout;
}

document.addEventListener("DOMContentLoaded", () => {
  if (window.adminLayoutInitialized) return;
  window.adminLayoutInitialized = true;

  // Make adminLayout global so onclick handlers work
  window.adminLayout = new window.AdminLayout();
  window.adminLayout.init();
  window.adminLayout.renderMobileHeader();
  window.adminLayout.renderBottomNav();
});

// Keep a reference for backwards compatibility
var adminLayout = null;
document.addEventListener("DOMContentLoaded", () => {
  adminLayout = window.adminLayout;
});

// Toast notification helper
function showToast(message, type = "default") {
  const existing = document.querySelector(".toast");
  if (existing) existing.remove();

  const toast = document.createElement("div");
  toast.className = `toast ${type}`;
  toast.textContent = message;
  document.body.appendChild(toast);
  setTimeout(() => toast.remove(), 3000);
}

// API request helper
async function apiRequest(endpoint, options = {}) {
  const token = localStorage.getItem("sello_token");
  if (!token) {
    window.location.href = "/login";
    return null;
  }

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

  // Handle no_store error - redirect to onboarding
  if (response.status === 403) {
    try {
      const data = await response.json();
      if (data.error === "no_store") {
        window.location.href = "/onboarding";
        return null;
      }
    } catch (e) {}
    return { error: "Access Denied" };
  }

  try {
    const text = await response.text();
    try {
      return JSON.parse(text);
    } catch (e) {
      console.error("Invalid JSON response:", text);
      return { error: "Server Error: Invalid JSON response" };
    }
  } catch (e) {
    return { error: "Network Error or Empty Response" };
  }
}

// Get current store ID
function getCurrentStoreId() {
  return localStorage.getItem("sello_current_store");
}
window.getCurrentStoreId = getCurrentStoreId;
