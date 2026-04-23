(function () {
  const shell = document.querySelector("[data-sidebar-shell]");
  if (!shell) {
    return;
  }

  const views = Array.from(shell.querySelectorAll("[data-sidebar-view]"));
  const openButtons = Array.from(shell.querySelectorAll("[data-sidebar-open]"));
  const toggleButton = shell.querySelector("[data-sidebar-toggle]");
  const dropdown = shell.querySelector("[data-sidebar-dropdown]");

  function setView(name) {
    const target = name || "home";
    shell.dataset.sidebarView = target;
    views.forEach(function (view) {
      view.classList.toggle("is-active", view.dataset.sidebarView === target);
    });
    shell.classList.toggle("is-menu-open", false);
    if (toggleButton) {
      toggleButton.setAttribute("aria-label", target === "home" ? "打開選單" : "返回");
    }
  }

  openButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      setView(button.dataset.sidebarOpen);
    });
  });

  if (toggleButton) {
    toggleButton.addEventListener("click", function () {
      if ((shell.dataset.sidebarView || "home") === "home") {
        shell.classList.toggle("is-menu-open");
        return;
      }
      setView("home");
    });
  }

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
