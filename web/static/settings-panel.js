(function () {
  const panel = document.querySelector("[data-settings-panel]");
  if (!panel) {
    return;
  }

  const storageKey = "twacc_chat_folder_categories";

  const avatarPreview = panel.querySelector("[data-avatar-preview]");
  const avatarInput = panel.querySelector("[data-avatar-input]");
  const avatarRemoveButton = panel.querySelector("[data-avatar-remove]");
  const avatarEditButton = panel.querySelector("[data-avatar-edit]");
  const folderList = panel.querySelector("[data-folder-list]");
  const folderTabs = document.querySelector("[data-folder-tabs]");
  const addFolderButton = panel.querySelector("[data-folder-add]");
  const subviews = Array.from(panel.querySelectorAll("[data-settings-view]"));
  const openButtons = Array.from(panel.querySelectorAll("[data-settings-open]"));
  const backButton = panel.querySelector("[data-settings-back]");

  function setSubview(name) {
    const target = name || "main";
    panel.dataset.settingsView = target;
    subviews.forEach(function (view) {
      view.classList.toggle("is-active", view.dataset.settingsView === target);
    });
  }

  function refreshFolderOrders() {
    // Name ordering is handled by normalizeFolders; no per-card rank UI is shown.
  }

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function createFolderCard(index, folder) {
    const name = folder && folder.name ? folder.name : ("分類 " + index);
    const isDefaultFolder = name.toUpperCase() === "ALL";
    const article = document.createElement("article");
    article.className = "settings-folder-card";
    article.dataset.folderIndex = String(index - 1);
    article.innerHTML = [
      '<div class="settings-folder-row">',
      '<label class="field-label">資料夾名稱</label>',
      isDefaultFolder
        ? '<span class="settings-folder-default">預設</span>'
        : [
          '<div class="settings-folder-actions">',
          '<button type="button" class="settings-folder-action" data-folder-edit>編輯</button>',
          '<button type="button" class="settings-folder-action is-danger" data-folder-delete>刪除</button>',
          "</div>"
        ].join(""),
      "</div>",
      '<input class="text-input" value="' + escapeHTML(name) + '" data-folder-name readonly>',
    ].join("");
    return article;
  }

  function normalizeFolders(items) {
    const folders = Array.isArray(items) ? items : [];
    const seen = new Set();
    const cleaned = folders
      .map(function (item, index) {
        return {
          name: String((item && item.name) || "").trim() || ("分類 " + (index + 1))
        };
      })
      .filter(function (item) {
        return item.name !== "";
      })
      .filter(function (item) {
        const key = item.name.toUpperCase();
        if (seen.has(key)) {
          return false;
        }
        seen.add(key);
        return true;
      })
      .sort(function (a, b) {
        if (a.name.toUpperCase() === "ALL") {
          return -1;
        }
        if (b.name.toUpperCase() === "ALL") {
          return 1;
        }
        return a.name.localeCompare(b.name, "en", { sensitivity: "base" });
      });

    if (!cleaned.length || cleaned[0].name.toUpperCase() !== "ALL") {
      cleaned.unshift({ name: "ALL" });
    } else {
      cleaned[0].name = "ALL";
    }

    return cleaned.map(function (item, index) {
      return {
        name: index === 0 ? "ALL" : item.name
      };
    });
  }

  function readFolders() {
    try {
      const raw = localStorage.getItem(storageKey);
      if (!raw) {
        return normalizeFolders([{ name: "ALL" }]);
      }
      return normalizeFolders(JSON.parse(raw));
    } catch (_) {
      return normalizeFolders([{ name: "ALL" }]);
    }
  }

  function collectFoldersFromDOM() {
    const cards = Array.from(folderList.querySelectorAll(".settings-folder-card"));
    const folders = cards.map(function (card, index) {
      const nameInput = card.querySelector("[data-folder-name]");
      return {
        name: nameInput ? nameInput.value : ("分類 " + (index + 1))
      };
    });
    return normalizeFolders(folders);
  }

  function saveFolders(folders) {
    const normalized = normalizeFolders(folders);
    localStorage.setItem(storageKey, JSON.stringify(normalized));
    renderFolderTabs(normalized);
    return normalized;
  }

  function renderFolderTabs(folders) {
    if (!folderTabs) {
      return;
    }
    folderTabs.innerHTML = folders.map(function (folder, index) {
      const active = index === 0 ? " is-active" : "";
      return '<button type="button" class="conversation-category-chip' + active + '">' + escapeHTML(folder.name) + "</button>";
    }).join("");
  }

  function renderFolderCards(folders) {
    folderList.innerHTML = "";
    folders.forEach(function (folder, index) {
      folderList.appendChild(createFolderCard(index + 1, folder));
    });
    refreshFolderOrders();
  }

  function handleFolderChange() {
    saveFolders(collectFoldersFromDOM());
    renderFolderCards(readFolders());
  }

  function nextFolderName(folders) {
    const names = new Set(folders.map(function (folder) {
      return folder.name.toUpperCase();
    }));
    let index = folders.length;
    let name = "分類 " + index;
    while (names.has(name.toUpperCase())) {
      index += 1;
      name = "分類 " + index;
    }
    return name;
  }

  function setFolderEditing(card, isEditing) {
    const input = card.querySelector("[data-folder-name]");
    const editButton = card.querySelector("[data-folder-edit]");
    if (!input || !editButton) {
      return;
    }
    card.classList.toggle("is-editing", isEditing);
    input.readOnly = !isEditing;
    editButton.textContent = isEditing ? "儲存" : "編輯";
    if (isEditing) {
      input.focus();
      input.select();
    }
  }

  function deleteFolder(card) {
    const index = Number(card.dataset.folderIndex || 0);
    if (index === 0) {
      return;
    }
    const folders = readFolders().filter(function (_, itemIndex) {
      return itemIndex !== index;
    });
    saveFolders(folders);
    renderFolderCards(readFolders());
  }

  if (avatarInput && avatarPreview) {
    avatarInput.addEventListener("change", function () {
      const file = avatarInput.files && avatarInput.files[0];
      if (!file) {
        return;
      }

      const reader = new FileReader();
      reader.onload = function () {
        avatarPreview.textContent = "";
        avatarPreview.style.backgroundImage = "url('" + reader.result + "')";
        avatarPreview.classList.add("has-image");
      };
      reader.readAsDataURL(file);
    });
  }

  if (avatarRemoveButton && avatarPreview && avatarInput) {
    avatarRemoveButton.addEventListener("click", function () {
      avatarInput.value = "";
      avatarPreview.style.backgroundImage = "";
      avatarPreview.textContent = "E";
      avatarPreview.classList.remove("has-image");
    });
  }

  if (avatarEditButton && avatarInput) {
    avatarEditButton.addEventListener("click", function () {
      avatarInput.click();
    });
  }

  if (addFolderButton && folderList) {
    addFolderButton.addEventListener("click", function () {
      const folders = readFolders();
      folders.push({
        name: nextFolderName(folders)
      });
      saveFolders(folders);
      renderFolderCards(readFolders());
    });
  }

  if (folderList) {
    folderList.addEventListener("click", function (event) {
      const editButton = event.target.closest("[data-folder-edit]");
      const deleteButton = event.target.closest("[data-folder-delete]");

      if (editButton) {
        const card = editButton.closest(".settings-folder-card");
        if (!card) {
          return;
        }
        if (!card.classList.contains("is-editing")) {
          setFolderEditing(card, true);
          return;
        }
        handleFolderChange();
        return;
      }

      if (deleteButton) {
        const card = deleteButton.closest(".settings-folder-card");
        if (card) {
          deleteFolder(card);
        }
      }
    });

    folderList.addEventListener("keydown", function (event) {
      if (!event.target.matches("[data-folder-name]")) {
        return;
      }
      if (event.key === "Enter") {
        event.preventDefault();
        handleFolderChange();
      }
      if (event.key === "Escape") {
        event.preventDefault();
        renderFolderCards(readFolders());
      }
    });
  }

  openButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      setSubview(button.dataset.settingsOpen);
    });
  });

  if (backButton) {
    backButton.addEventListener("click", function () {
      setSubview("main");
    });
  }

  const folders = readFolders();
  renderFolderCards(folders);
  renderFolderTabs(folders);
  setSubview("main");
})();
