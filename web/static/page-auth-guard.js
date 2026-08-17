(function () {
  const target = document.body && document.body.dataset ? document.body.dataset.requireAuthRedirect : "";
  if (!target) {
    return;
  }

  const tokenKey = "twacc_chat_session_token";
  const expiresAtKey = "twacc_chat_session_expires_at";

  function clearSession() {
    localStorage.removeItem(tokenKey);
    localStorage.removeItem(expiresAtKey);
    localStorage.removeItem("twacc_chat_role");
    localStorage.removeItem("twacc_chat_source_system");
    localStorage.removeItem("twacc_chat_external_user_id");
    localStorage.removeItem("twacc_chat_display_name");
    localStorage.removeItem("twacc_chat_email");
  }

  function hasValidSession() {
    const token = localStorage.getItem(tokenKey) || "";
    const expiresAt = localStorage.getItem(expiresAtKey) || "";
    if (!token) {
      return false;
    }
    if (!expiresAt) {
      return true;
    }

    const expiresAtTime = new Date(expiresAt).getTime();
    if (Number.isNaN(expiresAtTime) || expiresAtTime > Date.now()) {
      return true;
    }

    clearSession();
    return false;
  }

  if (!hasValidSession()) {
    window.location.replace(target);
  }
})();
