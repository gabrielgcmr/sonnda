// internal/adapters/inbound/http/web/assets/static/js/register.js
(function () {
  function onReady(fn) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", fn);
      return;
    }
    fn();
  }

  onReady(() => {
    const authForm = document.getElementById("authForm");
    const registerForm = document.getElementById("registerForm");
    const authButton = document.getElementById("authButton");
    const registerButton = document.getElementById("registerButton");
    const backButton = document.getElementById("backButton");
    const alertContainer = document.getElementById("alert-container");
    const accountTypeSelect = document.getElementById("accountType");
    const professionalFields = document.getElementById("professionalFields");

    function showAlert(message, type = "info") {
      const alertClass =
        type === "error"
          ? "bg-red-50 text-red-800 border border-red-200"
          : type === "success"
            ? "bg-green-50 text-green-800 border border-green-200"
            : "bg-blue-50 text-blue-800 border border-blue-200";
      alertContainer.innerHTML = `<div class="px-4 py-3 rounded-lg text-sm font-medium mb-4 ${alertClass}">${message}</div>`;
      alertContainer.scrollIntoView({ behavior: "smooth", block: "nearest" });
    }

    function clearErrors() {
      document
        .querySelectorAll('[id$="-error"]')
        .forEach((el) => (el.textContent = ""));
    }

    function showStep(stepNumber) {
      if (stepNumber === 1) {
        authForm.classList.remove("hidden");
        registerForm.classList.add("hidden");
        document.getElementById("current-step").textContent = "1";
        return;
      }
      if (stepNumber === 2) {
        authForm.classList.add("hidden");
        registerForm.classList.remove("hidden");
        document.getElementById("current-step").textContent = "2";
      }
    }

    accountTypeSelect.addEventListener("change", (e) => {
      if (e.target.value === "professional") {
        professionalFields.classList.remove("hidden");
        document.getElementById("professionalKind").required = true;
        document.getElementById("registrationNumber").required = true;
        document.getElementById("registrationIssuer").required = true;
        return;
      }
      professionalFields.classList.add("hidden");
      document.getElementById("professionalKind").required = false;
      document.getElementById("registrationNumber").required = false;
      document.getElementById("registrationIssuer").required = false;
    });

    authForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      clearErrors();

      const email = document.getElementById("email").value.trim();
      const password = document.getElementById("password").value;
      const confirmPassword = document.getElementById("confirmPassword").value;

      if (!email) {
        document.getElementById("email-error").textContent = "E-mail é obrigatório";
        return;
      }

      if (password.length < 6) {
        document.getElementById("password-error").textContent =
          "Senha deve ter pelo menos 6 caracteres";
        return;
      }

      if (password !== confirmPassword) {
        document.getElementById("confirmPassword-error").textContent =
          "Senhas não conferem";
        return;
      }

      authButton.disabled = true;
      authButton.innerHTML = '<span class="inline-block animate-spin mr-2">⏳</span>Processando...';

      try {
        await firebaseAuth.signUp(email, password);
        showAlert("Autenticação bem-sucedida! Complete seu perfil.", "success");
        showStep(2);
      } catch (error) {
        console.error("Firebase sign up error:", error);

        let errorMessage = "Erro ao criar conta";
        if (error.code === "auth/email-already-in-use") {
          errorMessage = "Este e-mail já está cadastrado";
        } else if (error.code === "auth/invalid-email") {
          errorMessage = "E-mail inválido";
        } else if (error.code === "auth/weak-password") {
          errorMessage = "Senha fraca";
        }

        showAlert(errorMessage, "error");
      } finally {
        authButton.disabled = false;
        authButton.textContent = "Continuar";
      }
    });

    registerForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      clearErrors();

      const fullName = document.getElementById("fullName").value.trim();
      const birthDate = document.getElementById("birthDate").value;
      const cpf = document.getElementById("cpf").value.trim();
      const phone = document.getElementById("phone").value.trim();
      const accountType = document.getElementById("accountType").value;

      if (!fullName) {
        document.getElementById("fullName-error").textContent = "Nome é obrigatório";
        return;
      }

      if (!birthDate) {
        document.getElementById("birthDate-error").textContent =
          "Data de nascimento é obrigatória";
        return;
      }

      if (!cpf) {
        document.getElementById("cpf-error").textContent = "CPF é obrigatório";
        return;
      }

      if (!phone) {
        document.getElementById("phone-error").textContent = "Telefone é obrigatório";
        return;
      }

      if (!accountType) {
        document.getElementById("accountType-error").textContent =
          "Tipo de conta é obrigatório";
        return;
      }

      let professionalData = {};
      if (accountType === "professional") {
        const professionalKind = document.getElementById("professionalKind").value;
        const registrationNumber = document.getElementById("registrationNumber").value.trim();
        const registrationIssuer = document
          .getElementById("registrationIssuer")
          .value.trim();
        const registrationState = document.getElementById("registrationState").value.trim();

        if (!professionalKind) {
          document.getElementById("professionalKind-error").textContent =
            "Tipo de profissional é obrigatório";
          return;
        }

        if (!registrationNumber) {
          document.getElementById("registrationNumber-error").textContent =
            "Número de registro é obrigatório";
          return;
        }

        if (!registrationIssuer) {
          document.getElementById("registrationIssuer-error").textContent =
            "Órgão emissor é obrigatório";
          return;
        }

        professionalData = {
          kind: professionalKind,
          registration_number: registrationNumber,
          registration_issuer: registrationIssuer,
          registration_state: registrationState || null,
        };
      }

      registerButton.disabled = true;
      registerButton.innerHTML =
        '<span class="inline-block animate-spin mr-2">⏳</span>Criando conta...';

      try {
        const idToken = await firebaseAuth.getIdToken();

        const payload = {
          email: document.getElementById("email").value,
          full_name: fullName,
          birth_date: birthDate,
          cpf: cpf,
          phone: phone,
          account_type: accountType,
          professional: accountType === "professional" ? professionalData : null,
        };

        const response = await fetch("/api/registration/register", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${idToken}`,
          },
          body: JSON.stringify(payload),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || "Erro ao registrar usuário");
        }

        showAlert("Conta criada com sucesso! Redirecionando...", "success");
        setTimeout(() => {
          window.location.href = "/dashboard";
        }, 2000);
      } catch (error) {
        console.error("Registration error:", error);
        showAlert(error.message || "Erro ao registrar", "error");
      } finally {
        registerButton.disabled = false;
        registerButton.innerHTML = "Criar Conta";
      }
    });

    backButton.addEventListener("click", () => {
      clearErrors();
      showStep(1);
    });

    if (typeof window.firebaseAuth === "undefined") {
      firebaseAuth.initialize(window.FIREBASE_CONFIG || {});
    }
  });
})();
