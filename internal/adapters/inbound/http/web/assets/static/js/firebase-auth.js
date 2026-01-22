// internal/adapters/inbound/http/web/assets/static/js/firebase-auth.js

/**
 * Firebase Authentication Manager (httpOnly Cookie Version)
 * Handles Firebase initialization, authentication flows, and session management via httpOnly cookies
 * Configuration is injected via window.firebaseConfig in the HTML template
 */

class FirebaseAuthManager {
  constructor() {
    this.auth = null;
    this.currentUser = null;
    this.initialized = false;
  }

  /**
   * Initialize Firebase with configuration
   * Configuration should be available as window.firebaseConfig
   * Call this before using any auth methods
   */
  async initialize() {
    try {
      if (this.initialized) {
        console.warn('Firebase already initialized');
        return;
      }

      if (!window.firebaseConfig) {
        throw new Error('Firebase configuration not found in window.firebaseConfig');
      }

      // Initialize Firebase
      firebase.initializeApp(window.firebaseConfig);
      this.auth = firebase.auth();
      this.initialized = true;

      // Setup auth state listener
      this.auth.onAuthStateChanged(async (user) => {
        this.currentUser = user;
        
        if (user) {
          // User is signed in - get token and send to backend to create session
          const idToken = await user.getIdToken();
          await this.createSession(idToken);
          console.log('User authenticated:', user.email);
        } else {
          // User is signed out
          console.log('User signed out');
        }
      });

      console.log('Firebase initialized successfully');
    } catch (error) {
      console.error('Firebase initialization error:', error);
      throw error;
    }
  }

  /**
   * Sign up with email and password
   */
  async signUp(email, password) {
    this.ensureInitialized();
    
    try {
      const userCredential = await this.auth.createUserWithEmailAndPassword(email, password);
      const idToken = await userCredential.user.getIdToken();
      
      // Create session on backend
      await this.createSession(idToken);
      
      return {
        success: true,
        user: userCredential.user,
        token: idToken
      };
    } catch (error) {
      console.error('Sign up error:', error);
      return {
        success: false,
        error: this.parseFirebaseError(error)
      };
    }
  }

  /**
   * Sign in with email and password
   */
  async signIn(email, password) {
    this.ensureInitialized();
    
    try {
      const userCredential = await this.auth.signInWithEmailAndPassword(email, password);
      const idToken = await userCredential.user.getIdToken();
      
      // Create session on backend
      await this.createSession(idToken);
      
      return {
        success: true,
        user: userCredential.user,
        token: idToken
      };
    } catch (error) {
      console.error('Sign in error:', error);
      return {
        success: false,
        error: this.parseFirebaseError(error)
      };
    }
  }

  /**
   * Sign out current user
   */
  async signOut() {
    this.ensureInitialized();
    
    try {
      // Delete session on backend first
      await this.deleteSession();
      
      // Then sign out from Firebase
      await this.auth.signOut();
      
      return { success: true };
    } catch (error) {
      console.error('Sign out error:', error);
      return {
        success: false,
        error: this.parseFirebaseError(error)
      };
    }
  }

  /**
   * Create session on backend (stores token in httpOnly cookie)
   */
  async createSession(idToken) {
    try {
      const response = await fetch('/auth/session', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ id_token: idToken }),
        credentials: 'same-origin' // Include cookies
      });

      if (!response.ok) {
        throw new Error('Failed to create session');
      }

      if (response.status === 204) {
        return { success: true };
      }

      return await response.json();
    } catch (error) {
      console.error('Create session error:', error);
      throw error;
    }
  }

  /**
   * Delete session on backend (removes httpOnly cookie)
   */
  async deleteSession() {
    try {
      const response = await fetch('/auth/logout', {
        method: 'POST',
        credentials: 'same-origin' // Include cookies
      });

      if (!response.ok) {
        console.warn('Failed to delete session on backend');
      }
    } catch (error) {
      console.error('Delete session error:', error);
    }
  }

  /**
   * Check session status
   */
  async checkSession() {
    try {
      const response = await fetch('/auth/session', {
        method: 'GET',
        credentials: 'same-origin' // Include cookies
      });

      if (response.ok) {
        return await response.json();
      }
      
      return { authenticated: false };
    } catch (error) {
      console.error('Check session error:', error);
      return { authenticated: false };
    }
  }

  /**
   * Get current ID token (from Firebase, not storage)
   */
  async getIdToken(forceRefresh = false) {
    this.ensureInitialized();
    
    if (!this.currentUser) {
      return null;
    }

    try {
      return await this.currentUser.getIdToken(forceRefresh);
    } catch (error) {
      console.error('Get token error:', error);
      return null;
    }
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated() {
    return !!this.currentUser;
  }

  /**
   * Parse Firebase error into user-friendly message
   */
  parseFirebaseError(error) {
    const errorMessages = {
      'auth/email-already-in-use': 'Este e-mail já está em uso',
      'auth/invalid-email': 'E-mail inválido',
      'auth/operation-not-allowed': 'Operação não permitida',
      'auth/weak-password': 'Senha muito fraca. Use pelo menos 6 caracteres',
      'auth/user-disabled': 'Esta conta foi desabilitada',
      'auth/user-not-found': 'Usuário não encontrado',
      'auth/wrong-password': 'Senha incorreta',
      'auth/too-many-requests': 'Muitas tentativas. Tente novamente mais tarde',
      'auth/network-request-failed': 'Erro de conexão. Verifique sua internet'
    };

    return errorMessages[error.code] || error.message || 'Erro desconhecido';
  }

  /**
   * Ensure Firebase is initialized before operations
   */
  ensureInitialized() {
    if (!this.initialized) {
      throw new Error('Firebase not initialized. Call initialize() first.');
    }
  }
}

// Create global instance
window.firebaseAuth = new FirebaseAuthManager();

// Auto-initialize Firebase when the DOM is ready if config is available
document.addEventListener('DOMContentLoaded', () => {
  if (window.firebaseConfig) {
    window.firebaseAuth.initialize().catch(err => {
      console.error('Failed to initialize Firebase:', err);
    });
  }
});

/**
 * Helper function to make authenticated API calls
 * Note: With httpOnly cookies, the browser automatically sends the cookie
 */
async function authenticatedFetch(url, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers
  };

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'same-origin' // Important: include cookies
  });

  if (response.status === 401) {
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  return response;
}
