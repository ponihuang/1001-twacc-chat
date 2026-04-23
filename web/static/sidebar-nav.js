(function () {
  const shell = document.querySelector("[data-sidebar-shell]");
  if (!shell) {
    return;
  }

  const views = Array.from(shell.querySelectorAll("[data-sidebar-view]"));
  const openButtons = Array.from(shell.querySelectorAll("[data-sidebar-open]"));
  const toggleButton = shell.querySelector("[data-sidebar-toggle]");
  const dropdown = shell.querySelector("[data-sidebar-dropdown]");
  const titleNode = shell.querySelector("[data-sidebar-title]");
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
    titleOverride = "";
    backAction = "";
    shell.dataset.sidebarView = target;
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

  document.addEventListener("twacc:sidebar-state", function (event) {
    if (currentView() !== "settings") {
      return;
    }
    const detail = event.detail || {};
    titleOverride = detail.title || "";
    backAction = detail.backAction || "";
    applyHeaderState(currentView());
  });

  document.addEventListener("click", function (event) {
    if (!dropdown || !shell.classList.contains("is-menu-open")) {
      return;
    }
    if (shell.contains(event.target)) {
      return;
    }
    shell.classList.remove("is-menu-open");
  });

  setView("home");
})();
