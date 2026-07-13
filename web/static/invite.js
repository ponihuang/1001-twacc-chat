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
  const emailVerified = app.querySelector("[data-invite-email-verified]");
  const editableInputs = [accountInput, nicknameInput, passwordInput, passwordConfirmationInput];
  const inputByField = {
    account: accountInput,
    nickname: nicknameInput,
    password: passwordInput,
    password_confirmation: passwordConfirmationInput,
    email: emailInput
  };
  const errorByField = {};
  app.querySelectorAll("[data-invite-error-for]").forEach(function (node) {
    errorByField[node.dataset.inviteErrorFor] = node;
  });

  const token = new URLSearchParams(window.location.search).get("token") || "";
  const storageKeys = {
    token: "twacc_chat_session_token",
    expiresAt: "twacc_chat_session_expires_at",
    role: "twacc_chat_role",
    sourceSystem: "twacc_chat_source_system",
    externalUserID: "twacc_chat_external_user_id",
    displayName: "twacc_chat_display_name",
    deviceID: "twacc_chat_device_id"
  };

  function setStatus(message, type) {
    statusNode.textContent = message;
    statusNode.classList.toggle("is-success", type === "success");
    statusNode.classList.toggle("is-error", type === "error");
  }

  function generateDeviceID() {
    if (window.crypto && typeof window.crypto.randomUUID === "function") {
      return window.crypto.randomUUID();
    }
    return "device-" + Date.now();
  }

  function currentDeviceID() {
    const existing = localStorage.getItem(storageKeys.deviceID) || "";
    if (existing.trim()) {
      return existing.trim();
    }
    const next = generateDeviceID();
    localStorage.setItem(storageKeys.deviceID, next);
    return next;
  }

  function clearErrors() {
    fieldError.textContent = "";
    fieldError.hidden = true;
    editableInputs.forEach(function (input) {
      input.classList.remove("is-error");
      input.setAttribute("aria-invalid", "false");
    });
    Object.keys(errorByField).forEach(function (field) {
      errorByField[field].textContent = "";
      errorByField[field].hidden = true;
    });
  }

  function setGeneralError(message) {
    fieldError.textContent = message || "";
    fieldError.hidden = !message;
  }

  function setFieldError(field, message) {
    const input = inputByField[field];
    const errorNode = errorByField[field];
    if (!input || !errorNode) {
      setGeneralError(message);
      return;
    }

    input.classList.add("is-error");
    input.setAttribute("aria-invalid", "true");
    errorNode.textContent = message;
    errorNode.hidden = false;
  }

  function setFormEnabled(enabled) {
    editableInputs.forEach(function (input) {
      input.disabled = !enabled;
    });
    submitButton.disabled = !enabled;
  }

  function setEmailVerified(verified) {
    if (emailVerified) {
      emailVerified.hidden = !verified;
    }
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

  function fieldForErrorCode(code) {
    switch (code) {
    case "USER_ALREADY_EXISTS":
    case "INVALID_EXTERNAL_USER_ID":
      return "account";
    case "EMAIL_ALREADY_EXISTS":
      return "email";
    case "INVALID_PASSWORD":
      return "password";
    case "PASSWORD_CONFIRMATION_MISMATCH":
      return "password_confirmation";
    case "INVALID_REQUEST":
      return "nickname";
    default:
      return "";
    }
  }

  function messageForErrorCode(result, fallback) {
    if (result && result.code === "USER_ALREADY_EXISTS") {
      return "該帳號已存在";
    }
    return responseMessage(result, fallback);
  }

  function showSubmitError(result, fallback) {
    const message = messageForErrorCode(result, fallback);
    const field = fieldForErrorCode(result && result.code ? result.code : "");
    if (field) {
      setFieldError(field, message);
      if (inputByField[field] && !inputByField[field].disabled) {
        inputByField[field].focus();
      }
      setStatus("", "");
      return;
    }

    setGeneralError(message);
    setStatus(message, "error");
  }

  function saveSession(result, loginPayload) {
    const data = result.data || {};
    localStorage.setItem(storageKeys.token, result.token);
    localStorage.setItem(storageKeys.expiresAt, result.expires_at || "");
    localStorage.setItem(storageKeys.role, data.role || "");
    localStorage.setItem(storageKeys.sourceSystem, loginPayload.source_system);
    localStorage.setItem(storageKeys.externalUserID, loginPayload.external_user_id);
    localStorage.setItem(storageKeys.displayName, data.display_name || loginPayload.external_user_id);
    localStorage.setItem(storageKeys.deviceID, loginPayload.device_id);
  }

  async function loadInvitation() {
    if (!token.trim()) {
      emailInput.value = "";
      setEmailVerified(false);
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
      setEmailVerified(true);
      setFormEnabled(true);
      setStatus("邀請有效，請完成帳號設定", "success");
      accountInput.focus();
    } catch (error) {
      emailInput.value = "";
      setEmailVerified(false);
      setFormEnabled(false);
      setStatus(error.message || "邀請無效或已過期", "error");
    }
  }

  async function submitInvitation(event) {
    event.preventDefault();
    clearErrors();

    const password = passwordInput.value.trim();
    const passwordConfirmation = passwordConfirmationInput.value.trim();
    if (password !== passwordConfirmation) {
      setFieldError("password_confirmation", "密碼確認不一致");
      setStatus("", "");
      passwordConfirmationInput.focus();
      return;
    }

    const payload = {
      token: token.trim(),
      account: accountInput.value.trim(),
      nickname: nicknameInput.value.trim(),
      password: password,
      password_confirmation: passwordConfirmation,
      device_id: currentDeviceID()
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
        setFormEnabled(true);
        showSubmitError(result, "註冊失敗");
        return;
      }

      setStatus("註冊完成，正在進入聊天室...", "success");
      if (!result.token) {
        throw new Error("註冊完成，但自動登入失敗");
      }
      saveSession(result, {
        source_system: "office",
        external_user_id: payload.account,
        device_id: payload.device_id
      });
      window.location.assign("/chat");
    } catch (error) {
      setFormEnabled(true);
      setStatus(error.message || "註冊失敗", "error");
    }
  }

  editableInputs.forEach(function (input) {
    input.addEventListener("input", function () {
      clearErrors();
    });
  });
  form.addEventListener("submit", submitInvitation);
  setFormEnabled(false);
  loadInvitation();
})();
