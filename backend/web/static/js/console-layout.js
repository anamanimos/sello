// Sello Console Layout Component
// Unified sidebar for Super Admin Console

class ConsoleLayout {
  constructor() {
    this.currentPage = this.detectCurrentPage();
    this.theme = localStorage.getItem("sello_theme") || "light";
  }

  detectCurrentPage() {
    const path = window.location.pathname;
    if (path === "/console/dashboard") return "dashboard";
    if (path.includes("/console/orders")) return "orders";
    if (path.includes("/console/users")) return "users";
    if (path.includes("/console/stores")) return "stores";
    if (path.includes("/console/plans")) return "plans";
    if (path.includes("/console/coupons")) return "coupons";
    if (path.includes("/console/analytics")) return "analytics";
    if (path.includes("/console/themes")) return "themes";
    if (path.includes("/console/settings")) return "settings";
    if (path.includes("/console/admins")) return "admins";
    if (path === "/console/coupons") return "coupons"; // Explicit check
    return "dashboard";
  }

  init() {
    if (this.theme === "dark") {
      document.documentElement.setAttribute("data-theme", "dark");
    } else {
      document.documentElement.removeAttribute("data-theme");
    }
    this.renderSidebar();
    this.renderMobileHeader();
    this.renderBottomNav();
    this.loadAdminInfo();
  }

  loadAdminInfo() {
    const admin = JSON.parse(localStorage.getItem("admin_user") || "{}");
    // Update sidebar info if elements exist
    const nameEl = document.getElementById("layoutUserName");
    const roleEl = document.getElementById("layoutUserRole");
    const avatarEl = document.getElementById("layoutUserAvatar");

    if (nameEl) nameEl.textContent = admin.name || "Admin";
    if (roleEl) roleEl.textContent = (admin.role || "admin").replace("_", " ");
    if (avatarEl) avatarEl.textContent = (admin.name || "A")[0].toUpperCase();
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

    // Update toggle button if visible
    const toggle = document.getElementById("themeToggle");
    if (toggle) toggle.classList.toggle("active", this.theme === "dark");
  }

  async logout() {
    const result = await Swal.fire({
      title: "Logout?",
      text: "Are you sure you want to sign out?",
      icon: "question",
      showCancelButton: true,
      confirmButtonColor: "#ef4444",
      confirmButtonText: "Logout",
    });

    if (!result.isConfirmed) return;

    localStorage.removeItem("admin_token");
    localStorage.removeItem("admin_user");
    window.location.href = "/console/login";
  }

  renderSidebar() {
    const sidebarEl = document.getElementById("sidebar");
    if (!sidebarEl) return;
    sidebarEl.classList.add("sidebar");

    // Same CSS styles as dashboard for consistency, but adapted if needed
    // We assume dashboard.css is loaded

    sidebarEl.innerHTML = `
      <div class="sidebar-header">
        <a href="/console/dashboard" class="logo">
          <img src="/static/images/sello-logo.webp" alt="Sello Console" height="32">
        </a>
      </div>

      <nav class="sidebar-nav">
        <a href="/console/dashboard" class="nav-item ${this.currentPage === "dashboard" ? "active" : ""}">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
          <span>Dashboard</span>
        </a>
        
        <div style="font-size:11px; font-weight:600; color:var(--text-muted); padding:12px 12px 4px; text-transform:uppercase; letter-spacing:0.5px;">Billing</div>
        
        <a href="/console/orders" class="nav-item ${this.currentPage === "orders" ? "active" : ""}">
           <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
           <span>Invoices</span>
        </a>
        <a href="/console/plans" class="nav-item ${this.currentPage === "plans" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
            <span>Plans</span>
        </a>
        <a href="/console/coupons" class="nav-item ${this.currentPage === "coupons" ? "active" : ""}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>
            <span>Coupons</span>
        </a>

        <div style="font-size:11px; font-weight:600; color:var(--text-muted); padding:12px 12px 4px; text-transform:uppercase; letter-spacing:0.5px;">Platform</div>

        <a href="/console/stores" class="nav-item ${this.currentPage === "stores" ? "active" : ""}">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>
          <span>Stores</span>
        </a>
        <a href="/console/users" class="nav-item ${this.currentPage === "users" ? "active" : ""}">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
          <span>Users</span>
        </a>
        <a href="/console/themes" class="nav-item ${this.currentPage === "themes" ? "active" : ""}">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"></circle><path d="M14.31 8l5.74 9.94M9.69 8h11.48M7.38 12l5.74-9.94M9.69 16L3.95 6.06M14.31 16H2.83M16.62 12l-5.74 9.94"></path></svg>
          <span>Themes</span>
        </a>
        
        <div style="font-size:11px; font-weight:600; color:var(--text-muted); padding:12px 12px 4px; text-transform:uppercase; letter-spacing:0.5px;">System</div>

        <a href="/console/analytics" class="nav-item ${this.currentPage === "analytics" ? "active" : ""}">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>
          <span>Analytics</span>
        </a>
        <a href="/console/admins" class="nav-item ${this.currentPage === "admins" ? "active" : ""}">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
          <span>Admins</span>
        </a>
        <a href="/console/settings" class="nav-item ${this.currentPage === "settings" ? "active" : ""}">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
          <span>Settings</span>
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
            <div class="toggle-switch ${this.theme === "dark" ? "active" : ""}" id="themeToggle" onclick="consoleLayout.toggleTheme()"></div>
        </div>
        <div class="sidebar-user" style="display: flex; align-items: center; gap: 12px; justify-content: space-between; padding:12px 0;">
          <div style="display: flex; align-items: center; gap: 12px;">
            <div id="layoutUserAvatar" style="width: 32px; height: 32px; background: var(--accent); color: white; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 600;">A</div>
            <div class="sidebar-user-info">
              <h4 id="layoutUserName" style="font-size: 13px; font-weight: 600; color: var(--text-primary); margin:0;">Admin</h4>
              <p id="layoutUserRole" style="font-size: 11px; color: var(--text-secondary); margin:0;">Super Admin</p>
            </div>
          </div>
          <button onclick="consoleLayout.logout()" title="Logout" style="background:none; border:none; cursor:pointer; padding:6px; color:var(--text-muted);">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
          </button>
        </div>
      </div>
    `;
  }

  renderMobileHeader() {
    const existing = document.querySelector(".mobile-header");
    if (existing) return;

    const header = document.createElement("header");
    header.className = "mobile-header";
    header.innerHTML = `
      <div style="display:flex;align-items:center;gap:12px;">
        <button class="mobile-menu-btn" onclick="consoleLayout.toggleSidebar()" aria-label="Menu">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="22" height="22">
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
        <a href="/console/dashboard" class="mobile-logo">
          <img src="/static/images/sello-logo.webp" alt="Sello Console" height="32">
        </a>
      </div>
      <div class="mobile-header-actions">
        <div class="theme-btn" onclick="consoleLayout.toggleTheme()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" width="20" height="20">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
        </div>
      </div>
    `;
    document.body.insertBefore(header, document.body.firstChild);

    // Create sidebar overlay backdrop
    const overlay = document.createElement("div");
    overlay.className = "sidebar-overlay";
    overlay.onclick = () => this.toggleSidebar(false);
    document.body.appendChild(overlay);
  }

  toggleSidebar(forceState) {
    const sidebar = document.getElementById("sidebar");
    const overlay = document.querySelector(".sidebar-overlay");
    if (!sidebar) return;

    const isOpen = sidebar.classList.contains("mobile-open");
    const shouldOpen = forceState !== undefined ? forceState : !isOpen;

    if (shouldOpen) {
      sidebar.classList.add("mobile-open");
      if (overlay) overlay.classList.add("show");
      document.body.style.overflow = "hidden";
    } else {
      sidebar.classList.remove("mobile-open");
      if (overlay) overlay.classList.remove("show");
      document.body.style.overflow = "";
    }
  }

  renderBottomNav() {
    const existing = document.querySelector(".bottom-nav");
    if (existing) return;

    const nav = document.createElement("nav");
    nav.className = "bottom-nav";
    nav.innerHTML = `
      <a href="/console/dashboard" class="bottom-nav-item ${this.currentPage === "dashboard" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
        <span>Dashboard</span>
      </a>
      <a href="/console/stores" class="bottom-nav-item ${this.currentPage === "stores" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>
        <span>Stores</span>
      </a>
      <a href="/console/users" class="bottom-nav-item ${this.currentPage === "users" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle></svg>
        <span>Users</span>
      </a>
      <a href="/console/themes" class="bottom-nav-item ${this.currentPage === "themes" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="10"></circle><path d="M14.31 8l5.74 9.94M9.69 8h11.48M7.38 12l5.74-9.94M9.69 16L3.95 6.06M14.31 16H2.83M16.62 12l-5.74 9.94"></path></svg>
        <span>Themes</span>
      </a>
      <a href="/console/settings" class="bottom-nav-item ${this.currentPage === "settings" ? "active" : ""}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.32 9c.3.56.82.98 1.41 1.07H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
        <span>Settings</span>
      </a>
    `;
    document.body.appendChild(nav);
  }
}

// Initialize
const consoleLayout = new ConsoleLayout();
document.addEventListener("DOMContentLoaded", () => {
  consoleLayout.init();
});
