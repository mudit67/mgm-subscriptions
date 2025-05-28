// Global state
let currentUser = null;
let currentSubscription = null;
let availablePlans = [];

// Initialization
document.addEventListener("DOMContentLoaded", function () {
  if (!authenticateUser()) return;
  loadDashboard();
});

function authenticateUser() {
  const token = localStorage.getItem("token");
  const user = localStorage.getItem("user");

  if (!token || !user) {
    window.location.href = "/";
    return false;
  }

  currentUser = JSON.parse(user);
  document.getElementById(
    "userName"
  ).textContent = `Welcome, ${currentUser.name}`;
  return true;
}

async function loadDashboard() {
  try {
    await loadUserSubscription();
    await loadPlans();
    await renderSubscriptionCard();
    checkAdminAccess();
  } catch (error) {
    console.error("Error loading dashboard:", error);
  }
}

function checkAdminAccess() {
  if (currentUser?.username === "admin") {
    const headerContent = document.querySelector(".header-content");
    const userMenu = headerContent.querySelector(".user-menu");

    const adminLink = document.createElement("a");
    adminLink.href = "/admin";
    adminLink.className = "btn btn-warning";
    adminLink.textContent = "Admin Panel";
    adminLink.style.marginRight = "10px";

    userMenu.insertBefore(adminLink, userMenu.firstChild);
  }
}

// Data loading functions
async function loadPlans() {
  const response = await API.getPlans();
  if (response.success) {
    availablePlans = response.data;
    renderPlans();
  }
}

async function loadUserSubscription() {
  try {
    const response = await API.getSubscription(currentUser.id);
    currentSubscription = response.success ? response.data : null;
  } catch (error) {
    currentSubscription = null;
  }
  renderSubscriptionCard();
}

// Rendering functions
function renderSubscriptionCard() {
  const subscriptionContent = document.getElementById("subscriptionContent");
  const statusBadge = document.getElementById("subscriptionStatusBadge");
  if (!currentSubscription) {
    statusBadge.innerHTML = "";
    subscriptionContent.innerHTML = `
    <div class="no-subscription">
    <h3>No Active Subscription</h3>
    <p>Choose from our amazing plans to get started with SubService</p>
    <button class="btn btn-glass" onclick="showPlans()">Browse Plans</button>
    </div>`;
    return;
  }

  console.log(currentSubscription);
  const plan = availablePlans.find((p) => p.id === currentSubscription.plan_id);
  const formatDate = (date) => new Date(date).toLocaleDateString("en-IN");

  statusBadge.innerHTML = `
        <span class="status-badge status-${currentSubscription.status.toLowerCase()}">
            ${currentSubscription.status}
        </span>`;

  subscriptionContent.innerHTML = `
        <div class="subscription-details">
            <div class="detail-item">
                <div class="detail-label">Plan Name</div>
                <div class="detail-value">${plan?.name || "Unknown Plan"}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Price</div>
                <div class="detail-value price-value">₹${
                  plan?.price || "N/A"
                }</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Duration</div>
                <div class="detail-value">${plan?.duration || "N/A"}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Started On</div>
                <div class="detail-value">${formatDate(
                  currentSubscription.start_date
                )}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Expires On</div>
                <div class="detail-value">${formatDate(
                  currentSubscription.expires_at
                )}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Features</div>
                <div class="detail-value">${
                  plan?.features.length || 0
                } features included</div>
            </div>
        </div>
        ${renderSubscriptionActions()}`;
}

function renderSubscriptionActions() {
  if (!currentSubscription) return "";

  const plan = availablePlans.find((p) => p.id === currentSubscription.plan_id);
  const planName = plan?.name || "Current Plan";

  if (currentSubscription.status === "ACTIVE") {
    return `
            <div class="subscription-actions">
                <button class="btn btn-danger-glass" onclick="cancelSubscription()">Cancel Subscription</button>
            </div>`;
  }

  return `
        <div class="subscription-actions">
            <button class="btn btn-success-glass" onclick="renewCurrentPlan()">Subscribe Again (${planName})</button>
        </div>`;
}

function renderPlans() {
  const plansGrid = document.getElementById("plansGrid");
  plansGrid.innerHTML = "";

  availablePlans.forEach((plan) => {
    const isCurrentActivePlan =
      currentSubscription?.plan_id === plan.id &&
      currentSubscription?.status === "ACTIVE";

    let buttonText = "Subscribe";
    let buttonClass = "btn-primary";
    let isDisabled = false;

    if (currentSubscription) {
      if (currentSubscription.status === "ACTIVE") {
        if (isCurrentActivePlan) {
          buttonText = "Current Plan";
          buttonClass = "btn-secondary";
          isDisabled = true;
        } else {
          buttonText = "Change Plan";
          buttonClass = "btn-warning";
        }
      }
    }

    const planCard = document.createElement("div");
    planCard.className = "plan-card";
    planCard.innerHTML = `
            <div class="plan-name">${plan.name}</div>
            <div class="plan-price">
                <span class="currency">₹</span>${plan.price}
                <span class="period">/${plan.duration}</span>
            </div>
            <ul class="plan-features">
                ${plan.features
                  .map((feature) => `<li>${feature}</li>`)
                  .join("")}
            </ul>
            <button class="btn ${buttonClass}" 
                    onclick="handleSubscriptionAction('${plan.id}')"
                    ${isDisabled ? "disabled" : ""}>
                ${buttonText}
            </button>`;

    plansGrid.appendChild(planCard);
  });
}

// Action handlers
async function handleSubscriptionAction(planId) {
  await executeAction(
    () => API.upsertSubscription(currentUser.id, planId),
    "Subscription updated successfully!"
  );
}

async function renewCurrentPlan() {
  if (!currentSubscription) {
    alert("No subscription found to renew");
    return;
  }

  const plan = availablePlans.find((p) => p.id === currentSubscription.plan_id);
  const planName = plan?.name || "your current plan";

  if (!confirm(`Are you sure you want to renew ${planName}?`)) return;

  await executeAction(
    () => API.upsertSubscription(currentUser.id, currentSubscription.plan_id),
    `${planName} renewed successfully!`
  );
}

async function cancelSubscription() {
  if (!confirm("Are you sure you want to cancel your subscription?")) return;

  await executeAction(
    () => API.cancelSubscription(currentUser.id),
    "Subscription cancelled successfully!"
  );
}

async function upgradeToPlan(planId) {
  await executeAction(
    () => API.upsertSubscription(currentUser.id, planId),
    "Subscription updated successfully!"
  );
}

// Utility functions
async function executeAction(apiCall, successMessage) {
  try {
    const response = await apiCall();
    if (response.success) {
      alert(successMessage);
      await loadDashboard();
    }
  } catch (error) {
    alert("Error: " + error.message);
  }
}

function showPlans() {
  document
    .getElementById("plansSection")
    .scrollIntoView({ behavior: "smooth" });
}

function logout() {
  localStorage.clear();
  window.location.href = "/";
}
