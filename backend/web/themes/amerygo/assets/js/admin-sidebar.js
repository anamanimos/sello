/**
 * Admin Sidebar Component
 * Include this script in all admin pages for consistent sidebar
 */

const AdminSidebar = {
    render: function(activePage) {
        const sidebar = document.createElement('div');
        sidebar.className = 'sidebar';
        sidebar.innerHTML = `
            <div class="sidebar-brand">AMERYGO ADMIN</div>
            <div class="sidebar-menu">
                <a href="dashboard.html" class="${activePage === 'dashboard' ? 'active' : ''}"><i class="fas fa-box"></i> Produk</a>
                <a href="settings.html" class="${activePage === 'settings' ? 'active' : ''}"><i class="fas fa-cog"></i> Pengaturan</a>
                <a href="#" onclick="AdminApp.downloadJSON(); return false;"><i class="fas fa-download"></i> Download JSON</a>
                <a href="#" onclick="AuthService.logout(); return false;"><i class="fas fa-sign-out-alt"></i> Logout</a>
            </div>
        `;
        return sidebar;
    },

    inject: function(activePage) {
        const existing = document.querySelector('.sidebar');
        if (existing) existing.remove();
        
        document.body.insertBefore(this.render(activePage), document.body.firstChild);
    },

    applyTheme: function() {
        const stored = localStorage.getItem('amerygo_settings');
        if (stored) {
            const settings = JSON.parse(stored);
            if (settings.themeColor) {
                document.documentElement.style.setProperty('--primary-color', settings.themeColor);
            }
        }
    }
};

// Auto-apply theme on load
document.addEventListener('DOMContentLoaded', () => {
    AdminSidebar.applyTheme();
});
