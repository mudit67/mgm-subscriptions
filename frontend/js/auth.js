document.addEventListener("DOMContentLoaded", function () {
  if (localStorage.getItem("token")) {
    window.location.href = "/dashboard";
    return;
  }

  const loginForm = document.getElementById("loginForm");
  const registerForm = document.getElementById("registerForm");
  const errorMessage = document.getElementById("errorMessage");

  loginForm.addEventListener("submit", handleLogin);
  registerForm.addEventListener("submit", handleRegister);
});

function showLogin() {
  document.getElementById("loginForm").classList.remove("hidden");
  document.getElementById("registerForm").classList.add("hidden");
  document.querySelectorAll(".tab-btn")[0].classList.add("active");
  document.querySelectorAll(".tab-btn")[1].classList.remove("active");
  hideError();
}

function showRegister() {
  document.getElementById("loginForm").classList.add("hidden");
  document.getElementById("registerForm").classList.remove("hidden");
  document.querySelectorAll(".tab-btn")[1].classList.add("active");
  document.querySelectorAll(".tab-btn")[0].classList.remove("active");
  hideError();
}

async function handleLogin(e) {
  e.preventDefault();

  const username = document.getElementById("loginUsername").value;
  const password = document.getElementById("loginPassword").value;

  try {
    const response = await API.login(username, password);

    if (response.success) {
      localStorage.setItem("token", response.data.token);
      localStorage.setItem("user", JSON.stringify(response.data.user));
      window.location.href = "/dashboard";
    }
  } catch (error) {
    showError(error.message);
  }
}

async function handleRegister(e) {
  e.preventDefault();

  const username = document.getElementById("registerUsername").value;
  const name = document.getElementById("registerName").value;
  const password = document.getElementById("registerPassword").value;

  try {
    const response = await API.register(username, name, password);

    if (response.success) {
      showError("Registration successful! Please login.", "success");
      showLogin();
      document.getElementById("registerForm").reset();
    }
  } catch (error) {
    showError(error.message);
  }
}

function showError(message, type = "error") {
  const errorDiv = document.getElementById("errorMessage");
  errorDiv.textContent = message;
  errorDiv.className = type === "success" ? "success-message" : "error-message";
  errorDiv.classList.remove("hidden");
}

function hideError() {
  document.getElementById("errorMessage").classList.add("hidden");
}
