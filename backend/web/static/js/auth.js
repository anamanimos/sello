// Sello Auth JavaScript

// Supabase client configuration
const SUPABASE_URL = ""; // Will be set from config
const SUPABASE_KEY = ""; // Will be set from config

// Toggle password visibility
function togglePassword() {
  const passwordInput = document.getElementById("password");
  const type = passwordInput.type === "password" ? "text" : "password";
  passwordInput.type = type;
}

// Password strength checker
function checkPasswordStrength(password) {
  const strengthEl = document.getElementById("passwordStrength");
  if (!strengthEl) return;

  let strength = 0;
  if (password.length >= 8) strength++;
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++;
  if (/\d/.test(password)) strength++;
  if (/[^a-zA-Z0-9]/.test(password)) strength++;

  strengthEl.className = "password-strength";
  if (password.length === 0) return;
  if (strength <= 1) strengthEl.classList.add("weak");
  else if (strength <= 2) strengthEl.classList.add("medium");
  else strengthEl.classList.add("strong");
}

// Generate slug from store name
function generateSlug(name) {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .trim();
}

// Update store slug preview
function updateSlugPreview() {
  const storeNameInput = document.getElementById("storeName");
  const slugPreview = document.getElementById("storeSlug");
  if (storeNameInput && slugPreview) {
    const slug = generateSlug(storeNameInput.value) || "tokoanda";
    slugPreview.textContent = slug;
  }
}

// Show toast notification
function showToast(message, type = "default") {
  // Remove existing toast
  const existingToast = document.querySelector(".toast");
  if (existingToast) existingToast.remove();

  const toast = document.createElement("div");
  toast.className = `toast ${type}`;
  toast.textContent = message;
  document.body.appendChild(toast);

  // Show toast
  setTimeout(() => toast.classList.add("show"), 10);

  // Hide after 3 seconds
  setTimeout(() => {
    toast.classList.remove("show");
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

// Set loading state on button
function setLoading(button, loading) {
  const text = button.querySelector(".btn-text");
  const loader = button.querySelector(".btn-loader");

  if (loading) {
    button.disabled = true;
    if (text) text.style.display = "none";
    if (loader) loader.style.display = "inline-flex";
  } else {
    button.disabled = false;
    if (text) text.style.display = "inline";
    if (loader) loader.style.display = "none";
  }
}

// Handle login form
async function handleLogin(e) {
  e.preventDefault();
  const form = e.target;
  const button = form.querySelector('button[type="submit"]');
  const email = form.email.value;
  const password = form.password.value;

  setLoading(button, true);

  try {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();

    if (response.ok) {
      // Store token and user
      localStorage.setItem("sello_token", data.access_token);
      localStorage.setItem("sello_user", JSON.stringify(data.user));
      showToast("Login berhasil!", "success");

      // Redirect to dashboard
      setTimeout(() => {
        window.location.href = "/dashboard";
      }, 1000);
    } else {
      showToast(data.error || "Login gagal", "error");
    }
  } catch (error) {
    showToast("Terjadi kesalahan. Coba lagi.", "error");
  } finally {
    setLoading(button, false);
  }
}

// Handle register form
async function handleRegister(e) {
  e.preventDefault();
  const form = e.target;
  const button = form.querySelector('button[type="submit"]');

  const name = form.name.value;
  const email = form.email.value;
  let whatsapp = form.whatsapp.value.replace(/[^0-9]/g, "");
  // Strip leading 0 if user typed it out of habit
  if (whatsapp.startsWith("0")) whatsapp = whatsapp.substring(1);
  // Prepend country code 62
  whatsapp = "62" + whatsapp;
  const password = form.password.value;
  const storeName = form.store_name?.value || "";

  if (password.length < 8) {
    showToast("Password minimal 8 karakter", "error");
    return;
  }

  setLoading(button, true);

  try {
    const response = await fetch("/api/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name,
        whatsapp,
        email,
        password,
        store_name: storeName,
      }),
    });

    const data = await response.json();

    if (response.ok) {
      if (data.verification_needed) {
        showToast("Registrasi berhasil! Kode OTP telah dikirim.", "success");
        // Store UID for activation context
        localStorage.setItem("pending_uid", data.user_id);
        setTimeout(() => {
          window.location.href = `/activate?uid=${data.user_id}`;
        }, 1500);
      } else {
        showToast(
          "Registrasi berhasil! Silakan cek email untuk verifikasi.",
          "success",
        );
        // Redirect to login
        setTimeout(() => {
          window.location.href = "/login";
        }, 2000);
      }
    } else {
      showToast(data.error || "Registrasi gagal", "error");
    }
  } catch (error) {
    showToast("Terjadi kesalahan. Coba lagi.", "error");
  } finally {
    setLoading(button, false);
  }
}

// Google Sign In (placeholder - needs Supabase setup)
// Google Sign In
async function signInWithGoogle() {
  window.location.href = "/api/auth/google/login";
}

// Initialize
document.addEventListener("DOMContentLoaded", function () {
  // CHECK AUTH STATE
  const token = localStorage.getItem("sello_token");
  if (token) {
    // If user is on login/register page and already has token, redirect to dashboard
    if (
      document.getElementById("loginForm") ||
      document.getElementById("registerForm")
    ) {
      window.location.href = "/dashboard";
      return; // Stop further execution
    }
  }

  // Login form
  const loginForm = document.getElementById("loginForm");
  if (loginForm) {
    loginForm.addEventListener("submit", handleLogin);
  }

  // Register form
  const registerForm = document.getElementById("registerForm");
  if (registerForm) {
    registerForm.addEventListener("submit", handleRegister);
  }

  // Password strength
  const passwordInput = document.getElementById("password");
  if (passwordInput) {
    passwordInput.addEventListener("input", (e) => {
      checkPasswordStrength(e.target.value);
    });
  }

  // Store name to slug
  const storeNameInput = document.getElementById("storeName");
  if (storeNameInput) {
    storeNameInput.addEventListener("input", updateSlugPreview);
  }

  // ENFORCE ONBOARDING ON DASHBOARD
  if (window.location.pathname.startsWith("/dashboard")) {
    checkOnboardingStatus();
  }
});

// Check if user has completed onboarding
async function checkOnboardingStatus() {
  const token = localStorage.getItem("sello_token");
  if (!token) {
    window.location.href = "/login";
    return;
  }

  // Try to use cached user first to avoid delay
  const userStr = localStorage.getItem("sello_user");
  if (userStr) {
    const user = JSON.parse(userStr);
    // If cached user has no whatsapp, redirect immediately
    if (!user.whatsapp || user.whatsapp.length < 5) {
      window.location.href = "/onboarding";
      return;
    }
    // Note: Store check requires fetching fresh data usually,
    // but let's do a fresh fetch to be safe and authoritative.
  }

  try {
    const response = await fetch("/api/auth/me", {
      headers: { Authorization: "Bearer " + token },
    });

    if (response.ok) {
      const data = await response.json();
      localStorage.setItem("sello_user", JSON.stringify(data.user));

      // Strict Logic: Must have Whatsapp AND at least 1 Store
      const hasWhatsapp = data.user.whatsapp && data.user.whatsapp.length > 5;
      const hasStore = data.stores && data.stores.length > 0;

      if (!hasWhatsapp || !hasStore) {
        window.location.href = "/onboarding";
      }
    } else {
      // Token invalid?
      localStorage.removeItem("sello_token");
      window.location.href = "/login";
    }
  } catch (e) {
    console.error("Auth check failed", e);
  }
}
