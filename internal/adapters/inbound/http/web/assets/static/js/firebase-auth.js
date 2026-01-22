// internal/adapters/inbound/http/web/assets/static/js/firebase-auth.js
import { initializeApp } from "https://www.gstatic.com/firebasejs/12.8.0/firebase-app.js";
import {
  createUserWithEmailAndPassword,
  getAuth,
  onAuthStateChanged,
  signInWithEmailAndPassword,
  signOut as firebaseSignOut,
} from "https://www.gstatic.com/firebasejs/12.8.0/firebase-auth.js";

/**
 * Firebase Authentication Manager (httpOnly Cookie Version)
 * Handles Firebase initialization, authentication flows, and session management via httpOnly cookies
 * Configuration is injected via window.firebaseConfig in the HTML template
 */
class FirebaseAuthManager {
  constructor() {
    this.app = null;
    this.auth = null;
    this.currentUser = null;
    this.initialized = false;
  }

  async initialize() {
    try {
      if (this.initialized) {
        console.warn("Firebase already initialized");
        return;
      }

      if (!window.firebaseConfig) {
        throw new Error("Firebase configuration not found in window.firebaseConfig");
      }

      this.app = initializeApp(window.firebaseConfig);
      this.auth = getAuth(this.app);
      this.initialized = true;

      onAuthStateChanged(this.auth, async (user) => {
        this.currentUser = user;

        if (user) {
          const idToken = await user.getIdToken();
          await this.createSession(idToken);
          console.log("User authenticated:", user.email);
          return;
        }

        console.log("User signed out");
      });

      console.log("Firebase initialized successfully");
    } catch (error) {
      console.error("Firebase initialization error:", error);
      throw error;
    }
  }

  async signUp(email, password) {
    this.ensureInitialized();

    try {
      const userCredential = await createUserWithEmailAndPassword(
        this.auth,
        email,
        password
      );
      const idToken = await userCredential.user.getIdToken();

      await this.createSession(idToken);

      return {
        success: true,
        user: userCredential.user,
        token: idToken,
      };
    } catch (error) {
      console.error("Sign up error:", error);
      return {
        success: false,
        error: this.parseFirebaseError(error),
      };
    }
  }

  async signIn(email, password) {
    this.ensureInitialized();

    try {
      const userCredential = await signInWithEmailAndPassword(
        this.auth,
        email,
        password
      );
      const idToken = await userCredential.user.getIdToken();

      await this.createSession(idToken);

      return {
        success: true,
        user: userCredential.user,
        token: idToken,
      };
    } catch (error) {
      console.error("Sign in error:", error);
      return {
        success: false,
        error: this.parseFirebaseError(error),
      };
    }
  }

  async signOut() {
    this.ensureInitialized();

    try {
      await this.deleteSession();
      await firebaseSignOut(this.auth);
      return { success: true };
    } catch (error) {
      console.error("Sign out error:", error);
      return {
        success: false,
        error: this.parseFirebaseError(error),
      };
    }
  }

  async createSession(idToken) {
    try {
      const response = await fetch("/auth/session", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ id_token: idToken }),
        credentials: "same-origin",
      });

      if (!response.ok) {
        throw new Error("Failed to create session");
      }

      if (response.status === 204) {
        return { success: true };
      }

      return await response.json();
    } catch (error) {
      console.error("Create session error:", error);
      throw error;
    }
  }

  async deleteSession() {
    try {
      const response = await fetch("/auth/logout", {
        method: "POST",
        credentials: "same-origin",
      });

      if (!response.ok) {
        console.warn("Failed to delete session on backend");
      }
    } catch (error) {
      console.error("Delete session error:", error);
    }
  }

  async checkSession() {
    try {
      const response = await fetch("/auth/session", {
        method: "GET",
        credentials: "same-origin",
      });

      if (response.ok) {
        return await response.json();
      }

      return { authenticated: false };
    } catch (error) {
      console.error("Check session error:", error);
      return { authenticated: false };
    }
  }

  async getIdToken(forceRefresh = false) {
    this.ensureInitialized();

    if (!this.currentUser) {
      return null;
    }

    try {
      return await this.currentUser.getIdToken(forceRefresh);
    } catch (error) {
      console.error("Get token error:", error);
      return null;
    }
  }

  isAuthenticated() {
    return !!this.currentUser;
  }

  parseFirebaseError(error) {
    const errorMessages = {
      "auth/email-already-in-use": "Este e-mail já está em uso",
      "auth/invalid-email": "E-mail inválido",
      "auth/operation-not-allowed": "Operação não permitida",
      "auth/weak-password": "Senha muito fraca. Use pelo menos 6 caracteres",
      "auth/user-disabled": "Esta conta foi desabilitada",
      "auth/user-not-found": "Usuário não encontrado",
      "auth/wrong-password": "Senha incorreta",
      "auth/too-many-requests": "Muitas tentativas. Tente novamente mais tarde",
      "auth/network-request-failed": "Erro de conexão. Verifique sua internet",
    };

    return errorMessages[error.code] || error.message || "Erro desconhecido";
  }

  ensureInitialized() {
    if (!this.initialized) {
      throw new Error("Firebase not initialized. Call initialize() first.");
    }
  }
}

window.firebaseAuth = new FirebaseAuthManager();

document.addEventListener("DOMContentLoaded", () => {
  if (window.firebaseConfig) {
    window.firebaseAuth.initialize().catch((err) => {
      console.error("Failed to initialize Firebase:", err);
    });
  }
});

async function authenticatedFetch(url, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: "same-origin",
  });

  if (response.status === 401) {
    window.location.href = "/login";
    throw new Error("Unauthorized");
  }

  return response;
}
