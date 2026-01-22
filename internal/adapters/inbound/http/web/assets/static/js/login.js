// internal/adapters/inbound/http/web/assets/static/js/login.js
    // Login form handler
    document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const button = document.getElementById('loginButton');
    
    // Clear previous errors
    clearErrors();
    
    // Disable button and show loading
    button.disabled = true;
    button.innerHTML = '<span class="spinner"></span>Entrando...';
    
    try {
      // Sign in with Firebase
      const result = await window.firebaseAuth.signIn(email, password);
      if (result.success) {
        showAlert('Login realizado com sucesso! Redirecionando...', 'success');
        
        // Redirect to home or dashboard after short delay
        setTimeout(() => {
          window.location.href = '/';
        }, 1000);
      } else {
        showAlert(result.error, 'error');
        button.disabled = false;
        button.innerHTML = 'Entrar';
      }
    } catch (error) {
      showAlert('Erro inesperado: ' + error.message, 'error');
      button.disabled = false;
      button.innerHTML = 'Entrar';
    }
  });

  // Helper functions
  function showAlert(message, type) {
    const alertContainer = document.getElementById('alert-container');
    let alertClass = 'bg-blue-50 text-blue-800 border border-blue-200';
    
    if (type === 'error') {
      alertClass = 'bg-red-50 text-red-800 border border-red-200';
    } else if (type === 'success') {
      alertClass = 'bg-green-50 text-green-800 border border-green-200';
    }
    
    alertContainer.innerHTML = `
      <div class="px-4 py-3 rounded-lg text-sm font-medium ${alertClass}">
        ${message}
      </div>
    `;
  }

  function clearErrors() {
    document.getElementById('email-error').textContent = '';
    document.getElementById('password-error').textContent = '';
    document.getElementById('alert-container').innerHTML = '';
  }

  // Check if already authenticated on page load
  window.addEventListener('load', async () => {
    const session = await window.firebaseAuth.checkSession();
    if (session.authenticated) {
      window.location.href = '/';
    }
  });
