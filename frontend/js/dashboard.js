let currentUser = null;
let currentSubscription = null;
let availablePlans = [];

document.addEventListener("DOMContentLoaded", function () {
  const token = localStorage.getItem("token");
  const user = localStorage.getItem("user");

  if (!token || !user) {
    window.location.href = "/";
    return;
  }

  currentUser = JSON.parse(user);
  document.getElementById(
    "userName"
  ).textContent = `Welcome, ${currentUser.name}`;

  loadDashboard();
});

async function loadDashboard() {
  try {
    await loadPlans();
    await loadUserSubscription();
  } catch (error) {
    console.error("Error loading dashboard:", error);
  }
}

async function loadPlans() {
  try {
    const response = await API.getPlans();
    if (response.success) {
      availablePlans = response.data;
      renderPlans();
    }
  } catch (error) {
    console.error("Error loading plans:", error);
  }
}

async function loadUserSubscription() {
  try {
    const response = await API.getSubscription(currentUser.id);
    if (response.success) {
      currentSubscription = response.data;
      renderSubscriptionStatus();
    }
  } catch (error) {
    currentSubscription = null;
    renderSubscriptionStatus();
  }
}

function renderPlans() {
  const plansGrid = document.getElementById("plansGrid");
  plansGrid.innerHTML = "";

  availablePlans.forEach((plan) => {
    const planCard = document.createElement("div");
    planCard.className = "plan-card";

    const isCurrentPlan =
      currentSubscription &&
      currentSubscription.plan_id === plan.id &&
      currentSubscription.status === "ACTIVE";

    const canSubscribe =
      !currentSubscription || currentSubscription.status !== "ACTIVE";

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
            <button class="btn ${
              isCurrentPlan ? "btn-secondary" : "btn-primary"
            }" 
                    onclick="subscribeToPlan('${plan.id}')"
                    ${isCurrentPlan ? "disabled" : ""}>
                ${
                  isCurrentPlan
                    ? "Current Plan"
                    : canSubscribe
                    ? "Subscribe"
                    : "Upgrade/Change Plan"
                }
            </button>
        `;

    plansGrid.appendChild(planCard);
  });
}

function renderSubscriptionStatus() {
  const subscriptionContent = document.getElementById("subscriptionContent");
  const managementSection = document.getElementById("managementSection");

  if (currentSubscription) {
    const plan = availablePlans.find(
      (p) => p.id === currentSubscription.plan_id
    );
    const expiryDate = new Date(
      currentSubscription.expires_at
    ).toLocaleDateString();

    let statusColor = "#28a745"; // Green for active
    if (currentSubscription.status === "EXPIRED") statusColor = "#dc3545"; // Red
    if (currentSubscription.status === "CANCELLED") statusColor = "#ffc107"; // Yellow
    if (currentSubscription.status === "INACTIVE") statusColor = "#6c757d"; // Gray

    subscriptionContent.innerHTML = `
            <div class="subscription-info">
                <h3>${plan ? plan.name : "Unknown Plan"}</h3>
                <p>Status: <span style="color: ${statusColor}; font-weight: 600;">${
      currentSubscription.status
    }</span></p>
                <p>Expires: ${expiryDate}</p>
                <p>Price: ₹${plan ? plan.price : "N/A"}/${
      plan ? plan.duration : "N/A"
    }</p>
            </div>
        `;

    // Show management section for all subscription statuses except EXPIRED
    if (currentSubscription.status !== "EXPIRED") {
      managementSection.classList.remove("hidden");
      updateManagementButtons();
    } else {
      managementSection.classList.add("hidden");
    }
  } else {
    subscriptionContent.innerHTML = `
            <p>No active subscription</p>
            <button class="btn btn-primary" onclick="showPlans()">Subscribe Now</button>
        `;
    managementSection.classList.add("hidden");
  }
}

function updateManagementButtons() {
  const managementActions = document.querySelector(".management-actions");

  if (currentSubscription.status === "ACTIVE") {
    managementActions.innerHTML = `
            <button class="btn btn-warning" onclick="showUpgradePlans()">Change Plan</button>
            <button class="btn btn-danger" onclick="cancelSubscription()">Cancel Subscription</button>
        `;
  } else if (currentSubscription.status === "CANCELLED") {
    managementActions.innerHTML = `
            <button class="btn btn-success" onclick="renewSubscription()">Renew Subscription</button>
            <button class="btn btn-warning" onclick="showUpgradePlans()">Change Plan</button>
        `;
  } else {
    managementActions.innerHTML = `
            <button class="btn btn-primary" onclick="showUpgradePlans()">Subscribe to New Plan</button>
        `;
  }
}

// New function for renewing cancelled subscriptions
async function renewSubscription() {
  if (!confirm("Are you sure you want to renew your subscription?")) {
    return;
  }

  try {
    const response = await API.renewSubscription(currentUser.id);
    if (response.success) {
      alert("Subscription renewed successfully!");
      await loadUserSubscription();
    }
  } catch (error) {
    alert("Error renewing subscription: " + error.message);
  }
}

async function subscribeToPlan(planId) {
  // Allow subscription if no current subscription or if current subscription is not active
  if (currentSubscription && currentSubscription.status === "ACTIVE") {
    alert(
      "You already have an active subscription. Please cancel it first or use the Change Plan option."
    );
    return;
  }

  try {
    const response = await API.createSubscription(currentUser.id, planId);
    if (response.success) {
      alert("Subscription created successfully!");
      await loadUserSubscription();
    }
  } catch (error) {
    alert("Error creating subscription: " + error.message);
  }
}

async function cancelSubscription() {
  if (!confirm("Are you sure you want to cancel your subscription?")) {
    return;
  }

  try {
    const response = await API.cancelSubscription(currentUser.id);
    if (response.success) {
      alert("Subscription cancelled successfully!");
      await loadUserSubscription();
    }
  } catch (error) {
    alert("Error cancelling subscription: " + error.message);
  }
}

function showPlans() {
  document
    .getElementById("plansSection")
    .scrollIntoView({ behavior: "smooth" });
}

function showUpgradePlans() {
  const modal = document.getElementById("planModal");
  const modalPlans = document.getElementById("modalPlans");

  modalPlans.innerHTML = "";
  availablePlans.forEach((plan) => {
    if (currentSubscription && plan.id === currentSubscription.plan_id) {
      return; // Skip current plan
    }

    const planDiv = document.createElement("div");
    planDiv.className = "plan-card";
    planDiv.innerHTML = `
            <div class="plan-name">${plan.name}</div>
            <div class="plan-price">₹${plan.price}/${plan.duration}</div>
            <ul class="plan-features">
                ${plan.features
                  .map((feature) => `<li>${feature}</li>`)
                  .join("")}
            </ul>
            <button class="btn btn-primary" onclick="upgradeToPlan('${
              plan.id
            }')">
                ${
                  currentSubscription && currentSubscription.status === "ACTIVE"
                    ? "Change to This Plan"
                    : "Subscribe"
                }
            </button>
        `;
    modalPlans.appendChild(planDiv);
  });

  modal.classList.remove("hidden");
}

function showReactivationPlans() {
  showUpgradePlans(); // Same as upgrade, but for cancelled subscriptions
}

async function upgradeToPlan(planId) {
  try {
    let response;

    if (currentSubscription && currentSubscription.status === "ACTIVE") {
      // Update existing subscription
      response = await API.updateSubscription(currentUser.id, planId);
    } else {
      // Create new subscription (for cancelled/expired subscriptions)
      response = await API.createSubscription(currentUser.id, planId);
    }

    if (response.success) {
      alert("Subscription updated successfully!");
      closeModal();
      await loadUserSubscription();
    }
  } catch (error) {
    alert("Error updating subscription: " + error.message);
  }
}

function closeModal() {
  document.getElementById("planModal").classList.add("hidden");
}

function logout() {
  localStorage.removeItem("token");
  localStorage.removeItem("user");
  window.location.href = "/";
}
