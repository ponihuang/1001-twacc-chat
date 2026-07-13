(function () {
  const app = document.querySelector("[data-auth-app]");
  if (!app) {
    return;
  }

  const loginForm = app.querySelector("[data-login-form]");
  const logoutButtons = Array.from(app.querySelectorAll("[data-logout-button]"));
  const statusNode = app.querySelector("[data-auth-status]");
  const loginStateNode = app.querySelector("[data-login-state]");
  const tokenStateNode = app.querySelector("[data-token-state]");
  const sourceSystemInput = app.querySelector("#source-system");
  const externalUserIDInput = app.querySelector("#external-user-id");
  const passwordInput = app.querySelector("#login-password");
  const loginFieldError = app.querySelector("[data-login-field-error]");
  const deviceIDInput = app.querySelector("#device-id");
  const integrationTokenInput = app.querySelector("#integration-token");
  const forgotPasswordToggle = app.querySelector("[data-forgot-password-toggle]");
  const forgotPasswordHint = app.querySelector("[data-forgot-password-hint]");

  const storageKeys = {
    token: "twacc_chat_session_token",
    expiresAt: "twacc_chat_session_expires_at",
    role: "twacc_chat_role",
    sourceSystem: "twacc_chat_source_system",
    externalUserID: "twacc_chat_external_user_id",
    displayName: "twacc_chat_display_name",
    deviceID: "twacc_chat_device_id",
    integrationToken: "twacc_chat_integration_token"
  };

  function generateDeviceID() {
    if (window.crypto && typeof window.crypto.randomUUID === "function") {
      return window.crypto.randomUUID();
    }
    return "device-" + Date.now();
  }

  function maskToken(token) {
    if (!token) {
      return "尚未保存 session token";
    }
    if (token.length <= 12) {
      return token;
    }
    return token.slice(0, 6) + "..." + token.slice(-4);
  }

  function formatDateTime(value) {
    if (!value) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString("zh-TW", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false
    });
  }

  function readSession() {
    const token = localStorage.getItem(storageKeys.token) || "";
    const expiresAt = localStorage.getItem(storageKeys.expiresAt) || "";
    if (token && sessionExpired(expiresAt)) {
      clearSession();
      return {
        token: "",
        expiresAt: "",
        role: "",
        sourceSystem: "",
        externalUserID: "",
        displayName: "",
        deviceID: localStorage.getItem(storageKeys.deviceID) || "",
        integrationToken: localStorage.getItem(storageKeys.integrationToken) || ""
      };
    }

    return {
      token: token,
      expiresAt: expiresAt,
      role: localStorage.getItem(storageKeys.role) || "",
      sourceSystem: localStorage.getItem(storageKeys.sourceSystem) || "",
      externalUserID: localStorage.getItem(storageKeys.externalUserID) || "",
      displayName: localStorage.getItem(storageKeys.displayName) || "",
      deviceID: localStorage.getItem(storageKeys.deviceID) || "",
      integrationToken: localStorage.getItem(storageKeys.integrationToken) || ""
    };
  }

  function sessionExpired(expiresAt) {
    if (!expiresAt) {
      return false;
    }
    const expiresAtTime = new Date(expiresAt).getTime();
    return !Number.isNaN(expiresAtTime) && expiresAtTime <= Date.now();
  }

  function clearSession() {
    localStorage.removeItem(storageKeys.token);
    localStorage.removeItem(storageKeys.expiresAt);
    localStorage.removeItem(storageKeys.role);
    localStorage.removeItem(storageKeys.sourceSystem);
    localStorage.removeItem(storageKeys.externalUserID);
    localStorage.removeItem(storageKeys.displayName);
  }

  function sessionEventDetail(session) {
    return {
      loggedIn: Boolean(session.token),
      token: session.token || "",
      role: session.role || "",
      sourceSystem: session.sourceSystem || "",
      externalUserID: session.externalUserID || "",
      displayName: session.displayName || "",
      actor: session.sourceSystem && session.externalUserID ? [session.sourceSystem, session.externalUserID].join(" / ") : ""
    };
  }

  function setLoginFieldError(message) {
    if (!loginFieldError) {
      return;
    }
    const hasError = Boolean(message);
    loginFieldError.textContent = message || "";
    loginFieldError.hidden = !hasError;
    externalUserIDInput.classList.toggle("is-error", hasError);
    passwordInput.classList.toggle("is-error", hasError);
    externalUserIDInput.setAttribute("aria-invalid", hasError ? "true" : "false");
    passwordInput.setAttribute("aria-invalid", hasError ? "true" : "false");
  }

  function renderSessionState() {
    const session = readSession();
    deviceIDInput.value = session.deviceID || generateDeviceID();
    if (!session.deviceID) {
      localStorage.setItem(storageKeys.deviceID, deviceIDInput.value);
    }
    integrationTokenInput.value = session.integrationToken || integrationTokenInput.value;
    sourceSystemInput.value = session.sourceSystem || sourceSystemInput.value;
    externalUserIDInput.value = session.externalUserID || externalUserIDInput.value;

    if (session.token) {
      const accountLabel = session.sourceSystem && session.externalUserID ? [session.sourceSystem, session.externalUserID].join(" / ") : "未知帳號";
      loginStateNode.textContent = "已登入：" + (session.displayName || session.externalUserID || "使用者");
      tokenStateNode.textContent = accountLabel + "，登入憑證有效";
      statusNode.textContent = session.expiresAt
        ? "已保存 session token，過期時間：" + formatDateTime(session.expiresAt)
        : "已保存 session token";
      statusNode.classList.remove("is-error");
      statusNode.classList.add("is-success");
      app.dataset.sessionToken = session.token;
      app.dataset.userRole = session.role || "";
      app.dataset.sourceSystem = session.sourceSystem || "";
      app.dataset.externalUserId = session.externalUserID || "";
      app.dataset.actor = [session.sourceSystem || "unknown", session.externalUserID || "unknown"].join(" / ");
      document.dispatchEvent(new CustomEvent("twacc:session-changed", { detail: sessionEventDetail(session) }));
      if (app.dataset.redirectWhenAuthenticated) {
        window.location.replace(app.dataset.redirectWhenAuthenticated);
      }
      return;
    }

    loginStateNode.textContent = "未登入";
    tokenStateNode.textContent = "尚未保存 session token";
    statusNode.textContent = "尚未登入";
    statusNode.classList.remove("is-error", "is-success");
    delete app.dataset.sessionToken;
    delete app.dataset.userRole;
    delete app.dataset.sourceSystem;
    delete app.dataset.externalUserId;
    delete app.dataset.actor;
    document.dispatchEvent(new CustomEvent("twacc:session-changed", { detail: sessionEventDetail(session) }));
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

  function requestHeaders(integrationToken) {
    return {
      "Content-Type": "application/json",
      ...(integrationToken ? { Authorization: "Bearer " + integrationToken } : {})
    };
  }

  async function requestLogin(payload, integrationToken) {
    const response = await fetch("/api/erp/login", {
      method: "POST",
      headers: requestHeaders(integrationToken),
      body: JSON.stringify(payload)
    });

    const result = await parseJSON(response);
    return { response, result, payload };
  }

  async function loginUser(payload, integrationToken) {
    const primarySource = payload.source_system || "office";
    const sources = primarySource === "office" || primarySource === "erp"
      ? ["office", "erp"]
      : [primarySource];
    let lastResponseData = null;

    for (const sourceSystem of sources) {
      const nextPayload = { ...payload, source_system: sourceSystem };
      const responseData = await requestLogin(nextPayload, integrationToken);
      if (responseData.response.ok && responseData.result.success && responseData.result.token) {
        return responseData;
      }

      lastResponseData = responseData;
      if (!responseData.result || responseData.result.code !== "USER_NOT_FOUND") {
        break;
      }
    }

    return lastResponseData;
  }

  async function submitLogin(event) {
    event.preventDefault();
    setLoginFieldError("");

    const payload = {
      source_system: sourceSystemInput.value.trim(),
      external_user_id: externalUserIDInput.value.trim(),
      password: passwordInput.value.trim(),
      device_id: deviceIDInput.value.trim() || generateDeviceID()
    };
    const integrationToken = integrationTokenInput.value.trim();
    localStorage.setItem(storageKeys.deviceID, payload.device_id);
    localStorage.setItem(storageKeys.integrationToken, integrationToken);

    statusNode.textContent = "登入中...";
    statusNode.classList.remove("is-error", "is-success");

    try {
      const responseData = await loginUser(payload, integrationToken);

      if (!responseData.response.ok || !responseData.result.success || !responseData.result.token) {
        throw new Error("帳號或密碼錯誤");
      }

      const loggedInPayload = responseData.payload || payload;
      localStorage.setItem(storageKeys.token, responseData.result.token);
      localStorage.setItem(storageKeys.expiresAt, responseData.result.expires_at || "");
      localStorage.setItem(storageKeys.role, responseData.result.data && responseData.result.data.role ? responseData.result.data.role : "");
      localStorage.setItem(storageKeys.displayName, responseData.result.data && responseData.result.data.display_name ? responseData.result.data.display_name : loggedInPayload.external_user_id);
      localStorage.setItem(storageKeys.sourceSystem, loggedInPayload.source_system);
      localStorage.setItem(storageKeys.externalUserID, loggedInPayload.external_user_id);

      renderSessionState();
      if (app.dataset.loginRedirect) {
        window.location.assign(app.dataset.loginRedirect);
      }
    } catch (error) {
      statusNode.textContent = error.message || "登入失敗";
      statusNode.classList.remove("is-success");
      statusNode.classList.add("is-error");
      setLoginFieldError(error.message || "帳號或密碼錯誤");
    }
  }

  function logout() {
    clearSession();
    sourceSystemInput.value = "erp";
    externalUserIDInput.value = "";
    passwordInput.value = "";
    renderSessionState();
    statusNode.textContent = "已登出，可重新登入";
    if (app.dataset.logoutRedirect) {
      window.location.assign(app.dataset.logoutRedirect);
    }
  }

  function setForgotPasswordHintOpen(isOpen) {
    if (!forgotPasswordToggle || !forgotPasswordHint) {
      return;
    }
    forgotPasswordHint.hidden = !isOpen;
    forgotPasswordToggle.setAttribute("aria-expanded", isOpen ? "true" : "false");
  }

  loginForm.addEventListener("submit", submitLogin);
  [externalUserIDInput, passwordInput].forEach(function (input) {
    input.addEventListener("input", function () {
      setLoginFieldError("");
    });
  });
  logoutButtons.forEach(function (button) {
    button.addEventListener("click", logout);
  });
  if (forgotPasswordToggle && forgotPasswordHint) {
    forgotPasswordToggle.addEventListener("click", function (event) {
      event.stopPropagation();
      setForgotPasswordHintOpen(forgotPasswordHint.hidden);
    });
    document.addEventListener("click", function (event) {
      if (forgotPasswordHint.hidden || forgotPasswordToggle.contains(event.target) || forgotPasswordHint.contains(event.target)) {
        return;
      }
      setForgotPasswordHintOpen(false);
    });
    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        setForgotPasswordHintOpen(false);
      }
    });
  }
  renderSessionState();
})();
