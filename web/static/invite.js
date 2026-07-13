(function () {
  const app = document.querySelector("[data-invite-app]");
  if (!app) {
    return;
  }

  const form = app.querySelector("[data-invite-form]");
  const emailInput = app.querySelector("#invite-email");
  const accountInput = app.querySelector("#invite-account");
  const nicknameInput = app.querySelector("#invite-nickname");
  const passwordInput = app.querySelector("#invite-password");
  const passwordConfirmationInput = app.querySelector("#invite-password-confirmation");
  const submitButton = app.querySelector("[data-invite-submit]");
  const statusNode = app.querySelector("[data-invite-status]");
  const fieldError = app.querySelector("[data-invite-field-error]");
  const editableInputs = [accountInput, nicknameInput, passwordInput, passwordConfirmationInput];

  const token = new URLSearchParams(window.location.search).get("token") || "";

  function setStatus(message, type) {
    statusNode.textContent = message;
    statusNode.classList.toggle("is-success", type === "success");
    statusNode.classList.toggle("is-error", type === "error");
  }

  function setFieldError(message) {
    const hasError = Boolean(message);
    fieldError.textContent = message || "";
    fieldError.hidden = !hasError;
    editableInputs.forEach(function (input) {
      input.classList.toggle("is-error", hasError);
      input.setAttribute("aria-invalid", hasError ? "true" : "false");
    });
  }

  function setFormEnabled(enabled) {
    editableInputs.forEach(function (input) {
      input.disabled = !enabled;
    });
    submitButton.disabled = !enabled;
  }

  async function parseJSON(response) {
    const text = await response.text();
    if (!text) {
      return {};
    }
    try {
      return JSON.parse(text);
    } catch (_) {
      return {};
    }
  }

  function responseMessage(result, fallback) {
    return result && result.message ? result.message : fallback;
  }

  async function loadInvitation() {
    if (!token.trim()) {
      setStatus("邀請連結缺少 token", "error");
      return;
    }

    try {
      const response = await fetch("/api/user-invitations/" + encodeURIComponent(token.trim()), {
        method: "GET",
        headers: { Accept: "application/json" }
      });
      const result = await parseJSON(response);
      if (!response.ok || !result.success || !result.data || !result.data.email) {
        throw new Error(responseMessage(result, "邀請無效或已過期"));
      }

      emailInput.value = result.data.email;
      setFormEnabled(true);
      setStatus("邀請有效，請完成帳號設定", "success");
      accountInput.focus();
    } catch (error) {
      setFormEnabled(false);
      setStatus(error.message || "邀請無效或已過期", "error");
    }
  }

  async function submitInvitation(event) {
    event.preventDefault();
    setFieldError("");

    const password = passwordInput.value.trim();
    const passwordConfirmation = passwordConfirmationInput.value.trim();
    if (password !== passwordConfirmation) {
      setFieldError("密碼確認不一致");
      passwordConfirmationInput.focus();
      return;
    }

    const payload = {
      token: token.trim(),
      account: accountInput.value.trim(),
      nickname: nicknameInput.value.trim(),
      password: password,
      password_confirmation: passwordConfirmation
    };

    setFormEnabled(false);
    setStatus("註冊中...", "");

    try {
      const response = await fetch("/api/user-invitations/accept", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json"
        },
        body: JSON.stringify(payload)
      });
      const result = await parseJSON(response);
      if (!response.ok || !result.success) {
        throw new Error(responseMessage(result, "註冊失敗"));
      }

      form.reset();
      emailInput.value = result.data && result.data.email ? result.data.email : "";
      setStatus("註冊完成，請前往登入頁登入", "success");
    } catch (error) {
      setFormEnabled(true);
      setStatus(error.message || "註冊失敗", "error");
    }
  }

  editableInputs.forEach(function (input) {
    input.addEventListener("input", function () {
      setFieldError("");
    });
  });
  form.addEventListener("submit", submitInvitation);
  setFormEnabled(false);
  loadInvitation();
})();
