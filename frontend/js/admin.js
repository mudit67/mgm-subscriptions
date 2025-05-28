let currentUser = null;
let allPlans = [];

document.addEventListener("DOMContentLoaded", function () {
  // Check authentication and admin access
  const token = localStorage.getItem("token");
  const user = localStorage.getItem("user");

  if (!token || !user) {
    window.location.href = "/";
    return;
  }

  currentUser = JSON.parse(user);

  // Check if user is admin
  if (currentUser.username !== "admin") {
    alert("Access denied. Admin privileges required.");
    window.location.href = "/dashboard";
    return;
  }

  document.getElementById(
    "adminName"
  ).textContent = `Welcome, ${currentUser.name}`;

  // Load admin data
  loadAdminDashboard();

  // Setup form handler
  document
    .getElementById("createPlanForm")
    .addEventListener("submit", handleCreatePlan);
});

async function loadAdminDashboard() {
  try {
    await loadAllPlans();
  } catch (error) {
    console.error("Error loading admin dashboard:", error);
  }
}

async function loadAllPlans() {
  try {
    const response = await API.getPlans();
    if (response.success) {
      allPlans = response.data;
      renderPlansTable();
    }
  } catch (error) {
    console.error("Error loading plans:", error);
  }
}

function renderPlansTable() {
  const tableBody = document.getElementById("plansTableBody");
  tableBody.innerHTML = "";

  allPlans.forEach((plan) => {
    const row = document.createElement("div");
    row.className = "table-row";

    row.innerHTML = `
            <div class="table-cell" data-label="Plan Name">${plan.name}</div>
            <div class="table-cell" data-label="Price">₹${plan.price}</div>
            <div class="table-cell" data-label="Duration">${plan.duration}</div>
            <div class="table-cell" data-label="Features">
                <div class="plan-features-list">
                    ${plan.features.slice(0, 2).join(", ")}
                    ${
                      plan.features.length > 2
                        ? ` (+${plan.features.length - 2} more)`
                        : ""
                    }
                </div>
            </div>
            <div class="table-cell" data-label="Actions">
                <div class="plan-actions">
                    <button class="btn btn-danger btn-small" onclick="deletePlan('${
                      plan.id
                    }', '${plan.name}')">
                        Delete
                    </button>
                </div>
            </div>
        `;

    tableBody.appendChild(row);
  });
}

function showCreatePlanModal() {
  document.getElementById("createPlanModal").classList.remove("hidden");
}

function closeCreatePlanModal() {
  document.getElementById("createPlanModal").classList.add("hidden");
  document.getElementById("createPlanForm").reset();
}

async function handleCreatePlan(e) {
  e.preventDefault();

  const name = document.getElementById("planName").value;
  const price = parseFloat(document.getElementById("planPrice").value);
  const duration = document.getElementById("planDuration").value;
  const featuresText = document.getElementById("planFeatures").value;

  // Parse features from textarea (one per line)
  const features = featuresText
    .split("\n")
    .map((feature) => feature.trim())
    .filter((feature) => feature.length > 0);

  if (features.length === 0) {
    alert("Please add at least one feature");
    return;
  }

  const planData = {
    name,
    price,
    duration,
    features,
  };

  try {
    const response = await API.createPlan(planData);
    if (response.success) {
      alert("Plan created successfully!");
      closeCreatePlanModal();
      await loadAllPlans();
    }
  } catch (error) {
    alert("Error creating plan: " + error.message);
  }
}

async function deletePlan(planId, planName) {
  if (
    !confirm(
      `Are you sure you want to delete the plan "${planName}"? This action cannot be undone.`
    )
  ) {
    return;
  }

  try {
    const response = await API.deletePlan(planId);
    if (response.success) {
      alert("Plan deleted successfully!");
      await loadAllPlans();
    }
  } catch (error) {
    alert("Error deleting plan: " + error.message);
  }
}

function logout() {
  localStorage.removeItem("token");
  localStorage.removeItem("user");
  window.location.href = "/";
}
