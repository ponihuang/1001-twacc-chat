(function () {
  const shell = document.querySelector("[data-sidebar-shell]");
  if (!shell) {
    return;
  }

  const views = Array.from(shell.querySelectorAll("[data-sidebar-view]"));
  const openButtons = Array.from(shell.querySelectorAll("[data-sidebar-open]"));
  const authRequiredButtons = Array.from(shell.querySelectorAll("[data-auth-required]"));
  const toggleButton = shell.querySelector("[data-sidebar-toggle]");
  const dropdown = shell.querySelector("[data-sidebar-dropdown]");
  const settingsActions = shell.querySelector("[data-settings-actions]");
  const settingsMenuToggle = shell.querySelector("[data-settings-menu-toggle]");
  const settingsMenu = shell.querySelector("[data-settings-menu]");
  const titleNode = shell.querySelector("[data-sidebar-title]");
  const accountNameNode = shell.querySelector("[data-sidebar-account-name]");
  let titleOverride = "";
  let backAction = "";

  const viewTitles = {
    home: "聊天系統",
    login: "登入",
    status: "登入狀態",
    direct: "建立一對一對話",
    settings: "設定"
  };

  function currentView() {
    return shell.dataset.sidebarView || "home";
  }

  function isLoggedIn() {
    return Boolean(localStorage.getItem("twacc_chat_session_token"));
  }

  function viewRequiresAuth(name) {
    return name === "direct" || name === "settings";
  }

  function accountName() {
    return localStorage.getItem("twacc_chat_external_user_id") || "使用者";
  }

  function applyAccountMenuLabel() {
    if (!accountNameNode) {
      return;
    }
    accountNameNode.textContent = accountName();
  }

  function applyAuthVisibility() {
    const loggedIn = isLoggedIn();
    shell.classList.toggle("is-authenticated", loggedIn);
    applyAccountMenuLabel();
    authRequiredButtons.forEach(function (button) {
      button.hidden = !loggedIn;
    });

    if (!loggedIn && viewRequiresAuth(currentView())) {
      setView("login");
    }
  }

  function applyHeaderState(target) {
    if (titleNode) {
      titleNode.textContent = titleOverride || viewTitles[target] || viewTitles.home;
    }
    if (toggleButton) {
      const isBack = Boolean(backAction) || target !== "home";
      toggleButton.setAttribute("aria-label", isBack ? "返回" : "打開選單");
    }
  }

  function setView(name) {
    const target = name || "home";
    if (viewRequiresAuth(target) && !isLoggedIn()) {
      setView("login");
      return;
    }
    titleOverride = "";
    backAction = "";
    shell.dataset.sidebarView = target;
    shell.dataset.settingsView = "";
    if (settingsMenu) {
      settingsMenu.hidden = true;
    }
    views.forEach(function (view) {
      view.classList.toggle("is-active", view.dataset.sidebarView === target);
    });
    shell.classList.toggle("is-menu-open", false);
    applyHeaderState(target);
    document.dispatchEvent(new CustomEvent("twacc:sidebar-view-changed", {
      detail: { view: target }
    }));
  }

  openButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      if (button.hidden) {
        return;
      }
      setView(button.dataset.sidebarOpen);
    });
  });

  if (toggleButton) {
    toggleButton.addEventListener("click", function () {
      if (currentView() === "home") {
        shell.classList.toggle("is-menu-open");
        return;
      }
      if (backAction) {
        document.dispatchEvent(new CustomEvent("twacc:sidebar-back", {
          detail: { action: backAction, view: currentView() }
        }));
        return;
      }
      setView("home");
    });
  }

  if (settingsMenuToggle && settingsMenu) {
    settingsMenuToggle.addEventListener("click", function () {
      if (settingsActions && settingsActions.hidden) {
        return;
      }
      settingsMenu.hidden = !settingsMenu.hidden;
    });
  }

  document.addEventListener("twacc:sidebar-state", function (event) {
    if (currentView() !== "settings") {
      return;
    }
    const detail = event.detail || {};
    titleOverride = detail.title || "";
    backAction = detail.backAction || "";
    shell.dataset.settingsView = detail.settingsView || "main";
    applyHeaderState(currentView());
  });

  document.addEventListener("twacc:session-changed", applyAuthVisibility);

  document.addEventListener("click", function (event) {
    if (settingsMenu && settingsMenuToggle && !settingsMenu.hidden && !settingsMenu.contains(event.target) && !settingsMenuToggle.contains(event.target)) {
      settingsMenu.hidden = true;
    }
    if (!dropdown || !shell.classList.contains("is-menu-open")) {
      return;
    }
    if (shell.contains(event.target)) {
      return;
    }
    shell.classList.remove("is-menu-open");
  });

  setView("home");
  applyAuthVisibility();
})();
