/**
 * SELLO - Landing Page JavaScript
 * Converted from React to Vanilla JavaScript
 */

// Environment detection and URL configuration
const isProduction =
  window.location.hostname === "sello.id" ||
  window.location.hostname === "www.sello.id";
const APP_URL = isProduction ? "https://app.sello.id" : window.location.origin;
const DEMO_URL = isProduction
  ? "https://demo.sello.id"
  : `${window.location.origin}/s/demo`;
const AUTH_URLS = {
  login: `${APP_URL}/login`,
  register: `${APP_URL}/register`,
  demo: DEMO_URL,
};

document.addEventListener("DOMContentLoaded", function () {
  // ============================================
  // MOBILE MENU TOGGLE
  // ============================================
  const mobileMenuBtn = document.getElementById("mobile-menu-btn");
  const mobileNav = document.getElementById("mobile-nav");
  const menuIcon = document.getElementById("menu-icon");
  const closeIcon = document.getElementById("close-icon");

  if (mobileMenuBtn && mobileNav) {
    mobileMenuBtn.addEventListener("click", function () {
      const isOpen = mobileNav.classList.contains("active");

      mobileNav.classList.toggle("active");
      menuIcon.style.display = isOpen ? "block" : "none";
      closeIcon.style.display = isOpen ? "none" : "block";
    });

    // Close mobile menu when clicking on a link
    const mobileNavLinks = mobileNav.querySelectorAll("a");
    mobileNavLinks.forEach(function (link) {
      link.addEventListener("click", function () {
        mobileNav.classList.remove("active");
        menuIcon.style.display = "block";
        closeIcon.style.display = "none";
      });
    });
  }

  // ============================================
  // SMOOTH SCROLL FOR ANCHOR LINKS
  // ============================================
  const anchorLinks = document.querySelectorAll('a[href^="#"]');

  anchorLinks.forEach(function (link) {
    link.addEventListener("click", function (e) {
      const href = this.getAttribute("href");

      if (href !== "#") {
        e.preventDefault();
        const target = document.querySelector(href);

        if (target) {
          const headerHeight = document.querySelector(".header").offsetHeight;
          const targetPosition = target.offsetTop - headerHeight;

          window.scrollTo({
            top: targetPosition,
            behavior: "smooth",
          });
        }
      }
    });
  });

  // ============================================
  // HEADER SCROLL EFFECT
  // ============================================
  const header = document.querySelector(".header");
  let lastScrollY = window.scrollY;

  window.addEventListener("scroll", function () {
    const currentScrollY = window.scrollY;

    // Add shadow when scrolled
    if (currentScrollY > 10) {
      header.style.boxShadow = "0 4px 6px -1px rgba(0, 0, 0, 0.1)";
    } else {
      header.style.boxShadow = "none";
    }

    lastScrollY = currentScrollY;
  });

  // ============================================
  // INTERSECTION OBSERVER FOR ANIMATIONS
  // ============================================
  const observerOptions = {
    threshold: 0.1,
    rootMargin: "0px 0px -50px 0px",
  };

  const fadeInObserver = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add("fade-in-visible");
        fadeInObserver.unobserve(entry.target);
      }
    });
  }, observerOptions);

  // Observe elements for fade-in animation
  const fadeElements = document.querySelectorAll(
    ".feature-card, .step-card, .pricing-card",
  );
  fadeElements.forEach(function (el) {
    el.classList.add("fade-in");
    fadeInObserver.observe(el);
  });

  // ============================================
  // BUTTON HOVER EFFECTS
  // ============================================
  const buttons = document.querySelectorAll(".btn");

  buttons.forEach(function (btn) {
    btn.addEventListener("mouseenter", function () {
      this.style.transform = "translateY(-2px)";
    });

    btn.addEventListener("mouseleave", function () {
      this.style.transform = "translateY(0)";
    });
  });

  // ============================================
  // AUTH BUTTONS - Redirect to login/register
  // ============================================
  const authButtons = document.querySelectorAll(".open-modal");

  authButtons.forEach(function (btn) {
    btn.addEventListener("click", function (e) {
      e.preventDefault();
      const buttonText = this.textContent.trim().toLowerCase();

      // Check if it's a login or register button
      if (buttonText.includes("masuk")) {
        window.location.href = AUTH_URLS.login;
      } else {
        // Default to register for all other buttons (Daftar, Mulai Gratis, etc.)
        window.location.href = AUTH_URLS.register;
      }
    });
  });

  // ============================================
  // DEMO BUTTON - Redirect to demo store
  // ============================================
  const demoButtons = document.querySelectorAll(".demo-button");

  demoButtons.forEach(function (btn) {
    btn.addEventListener("click", function (e) {
      e.preventDefault();
      window.location.href = AUTH_URLS.demo;
    });
  });

  // Close modal on close button click
  if (modalCloseBtn) {
    modalCloseBtn.addEventListener("click", closeModal);
  }

  // Close modal on overlay click
  if (modal) {
    modal.addEventListener("click", function (e) {
      if (e.target === modal) {
        closeModal();
      }
    });
  }

  // Close modal on Escape key
  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape" && modal && modal.classList.contains("active")) {
      closeModal();
    }
  });

  // Handle form submission
  if (earlyAccessForm) {
    earlyAccessForm.addEventListener("submit", function (e) {
      e.preventDefault();

      // Get form data
      const formData = {
        id: "reg_" + Date.now() + "_" + Math.random().toString(36).substr(2, 9),
        fullName: document.getElementById("full-name").value.trim(),
        email: document.getElementById("email").value.trim(),
        whatsapp: document.getElementById("whatsapp").value.trim(),
        timestamp: new Date().toISOString(),
      };

      // Validate email
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.email)) {
        alert("Format email tidak valid");
        return;
      }

      // Get existing registrations from localStorage
      const registrations = JSON.parse(
        localStorage.getItem("sello_registrations") || "[]",
      );

      // Check if email already registered
      const emailExists = registrations.some(
        (reg) => reg.email.toLowerCase() === formData.email.toLowerCase(),
      );
      if (emailExists) {
        alert("Email sudah terdaftar sebelumnya");
        return;
      }

      // Add new registration
      registrations.push(formData);

      // Save to localStorage
      localStorage.setItem(
        "sello_registrations",
        JSON.stringify(registrations),
      );

      // Log for debugging
      console.log("Early Access Registration:", formData);
      console.log("All Registrations:", registrations);

      // Show success message
      const modalContent = modal.querySelector(".modal");
      const formEl = modalContent.querySelector(".modal-form");
      const headerEl = modalContent.querySelector(".modal-header");
      const footerEl = modalContent.querySelector(".modal-footer-text");

      // Hide form elements
      if (formEl) formEl.style.display = "none";
      if (headerEl) headerEl.style.display = "none";
      if (footerEl) footerEl.style.display = "none";

      // Create success message
      const successDiv = document.createElement("div");
      successDiv.className = "modal-success";
      successDiv.innerHTML = `
        <div class="modal-success-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 6 9 17l-5-5"></path>
          </svg>
        </div>
        <h3 class="modal-success-title">Terima Kasih! 🎉</h3>
        <p class="modal-success-description">
          Pendaftaran Anda berhasil! Kami akan menghubungi <strong>${formData.fullName}</strong> melalui WhatsApp di <strong>${formData.whatsapp}</strong> ketika Sello siap diluncurkan.
        </p>
        <p style="font-size: 0.875rem; color: var(--muted-foreground); margin-top: 0.5rem;">
          Anda adalah pendaftar ke-${registrations.length}
        </p>
        <button class="btn btn-primary btn-lg btn-full" style="margin-top: 1.5rem;" onclick="document.getElementById('modal-close').click()">
          Tutup
        </button>
      `;

      modalContent.appendChild(successDiv);
    });
  }
});

// ============================================
// SELLO ADMIN FUNCTIONS (for console use)
// ============================================
window.SelloAdmin = {
  // Get all registrations
  getRegistrations: function () {
    const registrations = JSON.parse(
      localStorage.getItem("sello_registrations") || "[]",
    );
    console.table(registrations);
    return registrations;
  },

  // Get total count
  getCount: function () {
    const registrations = JSON.parse(
      localStorage.getItem("sello_registrations") || "[]",
    );
    console.log("Total Registrations:", registrations.length);
    return registrations.length;
  },

  // Export to JSON file (download)
  exportJSON: function () {
    const registrations = JSON.parse(
      localStorage.getItem("sello_registrations") || "[]",
    );

    if (registrations.length === 0) {
      console.log("No registrations to export");
      return;
    }

    const dataStr = JSON.stringify(registrations, null, 2);
    const dataBlob = new Blob([dataStr], { type: "application/json" });
    const url = URL.createObjectURL(dataBlob);

    const link = document.createElement("a");
    link.href = url;
    link.download = `sello_registrations_${new Date().toISOString().split("T")[0]}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    console.log(`Exported ${registrations.length} registrations to JSON file`);
  },

  // Export to CSV file (download)
  exportCSV: function () {
    const registrations = JSON.parse(
      localStorage.getItem("sello_registrations") || "[]",
    );

    if (registrations.length === 0) {
      console.log("No registrations to export");
      return;
    }

    // CSV header
    let csv = "ID,Nama Lengkap,Email,WhatsApp,Tanggal Daftar\n";

    // CSV rows
    registrations.forEach((reg) => {
      csv += `"${reg.id}","${reg.fullName}","${reg.email}","${reg.whatsapp}","${reg.timestamp}"\n`;
    });

    const dataBlob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(dataBlob);

    const link = document.createElement("a");
    link.href = url;
    link.download = `sello_registrations_${new Date().toISOString().split("T")[0]}.csv`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    console.log(`Exported ${registrations.length} registrations to CSV file`);
  },

  // Clear all registrations
  clearAll: function () {
    if (confirm("Apakah Anda yakin ingin menghapus semua data registrasi?")) {
      localStorage.removeItem("sello_registrations");
      console.log("All registrations cleared");
    }
  },

  // Help - show available commands
  help: function () {
    console.log(`
╔════════════════════════════════════════════════════════════╗
║                   SELLO ADMIN COMMANDS                     ║
╠════════════════════════════════════════════════════════════╣
║  SelloAdmin.getRegistrations()  - Lihat semua pendaftar    ║
║  SelloAdmin.getCount()          - Hitung total pendaftar   ║
║  SelloAdmin.exportJSON()        - Download sebagai JSON    ║
║  SelloAdmin.exportCSV()         - Download sebagai CSV     ║
║  SelloAdmin.clearAll()          - Hapus semua data         ║
║  SelloAdmin.help()              - Tampilkan bantuan ini    ║
╚════════════════════════════════════════════════════════════╝
    `);
  },
};

// ============================================
// ADD FADE-IN ANIMATION STYLES
// ============================================
const style = document.createElement("style");
style.textContent = `
  .fade-in {
    opacity: 0;
    transform: translateY(20px);
    transition: opacity 0.6s ease, transform 0.6s ease;
  }
  
  .fade-in-visible {
    opacity: 1;
    transform: translateY(0);
  }
`;
document.head.appendChild(style);
