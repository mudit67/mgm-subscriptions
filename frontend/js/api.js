// const API_BASE_URL = "http://localhost:7000/api";

class API {
  static async request(endpoint, options = {}) {
    const url = `/api${endpoint}`;
    const token = localStorage.getItem("token");

    const config = {
      headers: {
        "Content-Type": "application/json",
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || "Request failed");
      }

      return data;
    } catch (error) {
      console.error("API Error:", error);
      throw error;
    }
  }

  // Plan management endpoints
  static async createPlan(planData) {
    return this.request("/plans", {
      method: "POST",
      body: JSON.stringify(planData),
    });
  }

  static async updatePlan(planId, planData) {
    return this.request(`/plans/${planId}`, {
      method: "PUT",
      body: JSON.stringify(planData),
    });
  }

  static async deletePlan(planId) {
    return this.request(`/plans/${planId}`, {
      method: "DELETE",
    });
  }

  // Auth endpoints
  static async login(username, password) {
    return this.request("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
  }

  static async register(username, name, password) {
    return this.request("/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, name, password }),
    });
  }

  // Plans endpoints
  static async getPlans() {
    return this.request("/plans");
  }

  // Subscription endpoints
  static async upsertSubscription(userId, planId) {
    return this.request("/subscriptions", {
      method: "POST",
      body: JSON.stringify({ user_id: userId, plan_id: planId }),
    });
  }
  static async getSubscription(userId) {
    return this.request(`/subscriptions/${userId}`);
  }

  static async cancelSubscription(userId) {
    return this.request(`/subscriptions/${userId}`, {
      method: "DELETE",
    });
  }
}
