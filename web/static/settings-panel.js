(function () {
  const panel = document.querySelector("[data-settings-panel]");
  if (!panel) {
    return;
  }

  const storageKey = "twacc_chat_folder_categories";
  const activeFolderStorageKey = "twacc_chat_active_folder_tab_id";
  const conversationCatalogStorageKey = "twacc_chat_conversation_catalog";
  const sessionStorageKeys = {
    token: "twacc_chat_session_token",
    sourceSystem: "twacc_chat_source_system",
    externalUserID: "twacc_chat_external_user_id",
    displayName: "twacc_chat_display_name"
  };

  const avatarPreview = panel.querySelector("[data-avatar-preview]");
  const avatarInput = panel.querySelector("[data-avatar-input]");
  const avatarRemoveButton = panel.querySelector("[data-avatar-remove]");
  const avatarEditButton = panel.querySelector("[data-avatar-edit]");
  const avatarEditTrigger = panel.querySelector("[data-avatar-edit-trigger]");
  const profileName = panel.querySelector("[data-settings-profile-name]");
  const profileSource = panel.querySelector("[data-settings-profile-source]");
  const profileAccount = panel.querySelector("[data-settings-profile-account]");
  const avatarPreviewEdit = panel.querySelector("[data-avatar-preview-edit]");
  const profileEditName = panel.querySelector("[data-profile-edit-name]");
  const profileEditSource = panel.querySelector("[data-profile-edit-source]");
  const profileEditAccount = panel.querySelector("[data-profile-edit-account]");
  const profileEditSaveButton = panel.querySelector("[data-profile-edit-save]");
  const folderList = panel.querySelector("[data-folder-list]");
  const folderTabs = document.querySelector("[data-folder-tabs]");
  const addFolderButton = panel.querySelector("[data-folder-add]");
  const folderListView = panel.querySelector("[data-folder-list-view]");
  const folderEditView = panel.querySelector("[data-folder-edit-view]");
  const folderEditBackButton = panel.querySelector("[data-folder-edit-back]");
  const folderEditNameInput = panel.querySelector("[data-folder-edit-name]");
  const folderChatList = panel.querySelector("[data-folder-chat-list]");
  const folderEditSaveButton = panel.querySelector("[data-folder-edit-save]");
  const folderEditStatus = panel.querySelector("[data-folder-edit-status]");
  const folderEditMenuButton = panel.querySelector("[data-folder-edit-menu]");
  const folderEditMenuPanel = panel.querySelector("[data-folder-edit-menu-panel]");
  const folderEditDeleteButton = panel.querySelector("[data-folder-edit-delete]");
  const subviews = Array.from(panel.querySelectorAll("[data-settings-view]"));
  const openButtons = Array.from(panel.querySelectorAll("[data-settings-open]"));
  const editProfileButton = document.querySelector("[data-settings-edit-open]");
  const backButton = panel.querySelector("[data-settings-back]");
  const conversationList = document.querySelector("[data-conversation-list]");

  let activeFolderID = "";
  let activeFolderTabID = "";
  let profileEditOriginalName = "";

  function readSession() {
    return {
      token: localStorage.getItem(sessionStorageKeys.token) || "",
      sourceSystem: localStorage.getItem(sessionStorageKeys.sourceSystem) || "",
      externalUserID: localStorage.getItem(sessionStorageKeys.externalUserID) || "",
      displayName: localStorage.getItem(sessionStorageKeys.displayName) || ""
    };
  }

  function authHeaders() {
    const session = readSession();
    if (!session.token) {
      return null;
    }
    return {
      "Content-Type": "application/json",
      Authorization: "Bearer " + session.token
    };
  }

  function isLoggedIn() {
    return Boolean(readSession().token);
  }

  function accountStorageKey(base) {
    const session = readSession();
    if (!session.token || !session.sourceSystem || !session.externalUserID) {
      return "";
    }
    return base + "::" + encodeURIComponent(session.sourceSystem + ":" + session.externalUserID);
  }

  function avatarStorageKey() {
    return accountStorageKey("twacc_chat_avatar_data_url");
  }

  function applyAvatarImage(dataURL) {
    if (avatarPreview) {
      if (dataURL) {
        avatarPreview.textContent = "";
        avatarPreview.style.backgroundImage = "url('" + dataURL + "')";
        avatarPreview.classList.add("has-image");
      } else {
        avatarPreview.style.backgroundImage = "";
        avatarPreview.classList.remove("has-image");
      }
    }
    if (avatarPreviewEdit && avatarPreviewEdit.parentElement) {
      if (dataURL) {
        avatarPreviewEdit.textContent = "";
        avatarPreviewEdit.parentElement.style.backgroundImage = "url('" + dataURL + "')";
        avatarPreviewEdit.parentElement.classList.add("has-image");
      } else {
        avatarPreviewEdit.parentElement.style.backgroundImage = "";
        avatarPreviewEdit.parentElement.classList.remove("has-image");
      }
    }
  }

  function renderProfile() {
    const session = readSession();
    if (!session.token) {
      if (profileName) {
        profileName.textContent = "未登入";
      }
      if (profileAccount) {
        profileAccount.textContent = "請先登入";
      }
      if (profileSource) {
        profileSource.textContent = "尚未登入";
      }
      if (avatarPreview) {
        avatarPreview.textContent = "";
        avatarPreview.style.backgroundImage = "";
        avatarPreview.classList.remove("has-image");
      }
      if (avatarPreviewEdit) {
        avatarPreviewEdit.textContent = "";
        if (avatarPreviewEdit.parentElement) {
          avatarPreviewEdit.parentElement.style.backgroundImage = "";
          avatarPreviewEdit.parentElement.classList.remove("has-image");
        }
      }
      if (profileEditName) {
        profileEditName.value = "";
      }
      if (profileEditSource) {
        profileEditSource.value = "";
      }
      if (profileEditAccount) {
        profileEditAccount.value = "";
      }
      updateProfileSaveVisibility();
      return;
    }

    if (profileName) {
      profileName.textContent = session.displayName || session.externalUserID || "使用者";
    }
    if (profileSource) {
      profileSource.textContent = session.sourceSystem || "unknown";
    }
    if (profileAccount) {
      profileAccount.textContent = session.externalUserID || "unknown";
    }
    const avatarKey = avatarStorageKey();
    const avatarDataURL = avatarKey ? localStorage.getItem(avatarKey) || "" : "";
    applyAvatarImage(avatarDataURL);
    const avatarInitial = (session.displayName || session.externalUserID || "U").slice(0, 1).toUpperCase();
    if (avatarPreview && !avatarPreview.classList.contains("has-image")) {
      avatarPreview.textContent = avatarInitial;
    }
    if (avatarPreviewEdit && avatarPreviewEdit.parentElement && !avatarPreviewEdit.parentElement.classList.contains("has-image")) {
      avatarPreviewEdit.textContent = avatarInitial;
    }
    if (profileEditName) {
      profileEditOriginalName = session.displayName || session.externalUserID || "";
      profileEditName.value = profileEditOriginalName;
    }
    if (profileEditSource) {
      profileEditSource.value = session.sourceSystem || "";
    }
    if (profileEditAccount) {
      profileEditAccount.value = session.externalUserID || "";
    }
    updateProfileSaveVisibility();
  }

  function updateProfileSaveVisibility() {
    if (!profileEditSaveButton || !profileEditName) {
      return;
    }
    const displayName = profileEditName.value.trim();
    const changed = Boolean(displayName) && displayName !== profileEditOriginalName;
    profileEditSaveButton.hidden = !changed;
  }

  function setSubview(name) {
    const target = name || "main";
    panel.dataset.settingsView = target;
    subviews.forEach(function (view) {
      view.classList.toggle("is-active", view.dataset.settingsView === target);
    });
    syncSidebarState();
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

  async function saveProfileName() {
    if (!profileEditName) {
      return false;
    }
    const displayName = profileEditName.value.trim();
    if (!displayName) {
      profileEditName.focus();
      return false;
    }
    if (displayName === profileEditOriginalName) {
      return true;
    }
    const headers = authHeaders();
    if (!headers) {
      return false;
    }

    const response = await fetch("/api/users/me/profile", {
      method: "PATCH",
      headers: headers,
      body: JSON.stringify({ display_name: displayName })
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      throw new Error(result.message || "個人資料更新失敗");
    }

    const data = result.data || {};
    const savedName = data.display_name || displayName;
    localStorage.setItem(sessionStorageKeys.displayName, savedName);
    profileEditOriginalName = savedName;
    renderProfile();
    updateProfileSaveVisibility();
    document.dispatchEvent(new CustomEvent("twacc:session-changed", {
      detail: {
        loggedIn: true,
        token: readSession().token,
        sourceSystem: data.source_system || readSession().sourceSystem,
        externalUserID: data.external_user_id || readSession().externalUserID,
        displayName: savedName
      }
    }));
    return true;
  }

  function syncSidebarState() {
    const shell = document.querySelector("[data-sidebar-shell]");
    if (!shell || (shell.dataset.sidebarView || "home") !== "settings") {
      return;
    }
    const target = panel.dataset.settingsView || "main";
    document.dispatchEvent(new CustomEvent("twacc:sidebar-state", {
      detail: {
        title: target === "folders" ? "聊天室分類" : (target === "profile-edit" ? "編輯個人資料" : "設定"),
        backAction: target === "folders" || target === "profile-edit" ? "settings-main" : "",
        settingsView: target
      }
    }));
  }

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function folderID(name) {
    return "folder_" + encodeURIComponent(String(name || "").trim().toLowerCase()).replaceAll("%", "_");
  }

  function createDefaultFolder() {
    return {
      id: "all",
      name: "ALL",
      conversation_ids: []
    };
  }

  function normalizeFolders(items) {
    const folders = Array.isArray(items) ? items : [];
    const seen = new Set();
    const cleaned = folders.map(function (item, index) {
      const name = String((item && item.name) || "").trim() || ("分類 " + (index + 1));
      return {
        id: item && item.id ? String(item.id) : folderID(name),
        name: name,
        conversation_ids: Array.isArray(item && item.conversation_ids)
          ? item.conversation_ids.map(String)
          : []
      };
    }).filter(function (item) {
      const key = item.name.toUpperCase();
      if (!item.name || seen.has(key)) {
        return false;
      }
      seen.add(key);
      return true;
    }).sort(function (a, b) {
      if (a.name.toUpperCase() === "ALL") {
        return -1;
      }
      if (b.name.toUpperCase() === "ALL") {
        return 1;
      }
      return a.name.localeCompare(b.name, "en", { sensitivity: "base" });
    });

    if (!cleaned.length || cleaned[0].name.toUpperCase() !== "ALL") {
      cleaned.unshift(createDefaultFolder());
    } else {
      cleaned[0].id = "all";
      cleaned[0].name = "ALL";
    }

    return cleaned;
  }

  function readFolders() {
    const key = accountStorageKey(storageKey);
    if (!key) {
      return normalizeFolders([createDefaultFolder()]);
    }
    try {
      const raw = localStorage.getItem(key);
      if (!raw) {
        return normalizeFolders([createDefaultFolder()]);
      }
      return normalizeFolders(JSON.parse(raw));
    } catch (_) {
      return normalizeFolders([createDefaultFolder()]);
    }
  }

  function saveFolders(folders) {
    const key = accountStorageKey(storageKey);
    if (!key) {
      return normalizeFolders([createDefaultFolder()]);
    }
    const normalized = normalizeFolders(folders);
    localStorage.setItem(key, JSON.stringify(normalized));
    renderFolderTabs(normalized);
    return normalized;
  }

  function customFolders(folders) {
    return folders.filter(function (folder) {
      return folder.name.toUpperCase() !== "ALL";
    });
  }

  function renderFolderTabs(folders) {
    if (!folderTabs) {
      return;
    }
    if (!isLoggedIn()) {
      folderTabs.hidden = true;
      folderTabs.innerHTML = "";
      return;
    }
    folderTabs.hidden = false;
    const items = customFolders(folders);
    const allActive = !activeFolderTabID ? " is-active" : "";
    folderTabs.innerHTML = [
      '<button type="button" class="conversation-category-chip' + allActive + '" data-folder-tab-id="">ALL</button>'
    ].concat(items.map(function (folder) {
      const active = folder.id === activeFolderTabID ? " is-active" : "";
      return [
        '<button type="button" class="conversation-category-chip' + active + '" data-folder-tab-id="' + escapeHTML(folder.id) + '">',
        escapeHTML(folder.name),
        "</button>"
      ].join("");
    })).join("");
  }

  function activeFolderTab() {
    if (!activeFolderTabID) {
      return null;
    }
    return readFolders().find(function (folder) {
      return folder.id === activeFolderTabID;
    }) || null;
  }

  function applyConversationFolderFilter() {
    const folder = activeFolderTab();
    const allowed = folder ? new Set((folder.conversation_ids || []).map(String)) : null;
    document.dispatchEvent(new CustomEvent("twacc:conversation-filter-changed", {
      detail: {
        folderID: folder ? folder.id : "",
        conversationIDs: folder ? Array.from(allowed) : []
      }
    }));
  }

  function conversationCount(folder) {
    const count = Array.isArray(folder.conversation_ids) ? folder.conversation_ids.length : 0;
    return count ? (count + " 個聊天") : "尚未加入聊天";
  }

  function renderFolderCards(folders) {
    const items = customFolders(folders);
    folderList.innerHTML = items.length ? items.map(function (folder) {
      return [
        '<article class="settings-folder-item" data-folder-id="' + escapeHTML(folder.id) + '">',
        '<button type="button" class="settings-folder-item-main" data-folder-open>',
        "<strong>" + escapeHTML(folder.name) + "</strong>",
        "<span>" + escapeHTML(conversationCount(folder)) + "</span>",
        "</button>",
        '<button type="button" class="settings-folder-more" data-folder-more aria-label="資料夾選單">⋮</button>',
        '<div class="settings-folder-menu" data-folder-menu hidden>',
        '<button type="button" data-folder-delete>刪除</button>',
        "</div>",
        "</article>"
      ].join("");
    }).join("") : [
      '<div class="settings-folder-empty">',
      "<strong>尚未建立資料夾</strong>",
      "<span>按「建立新資料夾」開始分類聊天室。</span>",
      "</div>"
    ].join("");
  }

  function nextFolderName(folders) {
    const names = new Set(folders.map(function (folder) {
      return folder.name.toUpperCase();
    }));
    let index = customFolders(folders).length + 1;
    let name = "資料夾 " + index;
    while (names.has(name.toUpperCase())) {
      index += 1;
      name = "資料夾 " + index;
    }
    return name;
  }

  function readConversationsFromDOM() {
    try {
      const key = accountStorageKey(conversationCatalogStorageKey);
      const raw = key ? localStorage.getItem(key) : "";
      if (raw) {
        const parsed = JSON.parse(raw);
        if (Array.isArray(parsed) && parsed.length) {
          return parsed.map(function (item) {
            return {
              id: String(item.id || ""),
              title: String(item.title || ""),
              preview: String(item.preview || "聊天室")
            };
          }).filter(function (item) {
            return item.id !== "";
          });
        }
      }
    } catch (_) {
    }
    return Array.from(document.querySelectorAll("[data-conversation-id]")).map(function (button) {
      const id = String(button.dataset.conversationId || "");
      const titleNode = button.querySelector(".conversation-card-header strong");
      const previewNode = button.querySelector(".conversation-card-header + span");
      return {
        id: id,
        title: titleNode ? titleNode.textContent.trim() : ("對話 " + id),
        preview: previewNode ? previewNode.textContent.trim() : "聊天室"
      };
    }).filter(function (item) {
      return item.id !== "";
    });
  }

  function findFolder(folderIDValue) {
    return readFolders().find(function (folder) {
      return folder.id === folderIDValue;
    });
  }

  function setFolderView(mode) {
    const editing = mode === "edit";
    folderListView.hidden = editing;
    folderEditView.hidden = !editing;
  }

  function renderFolderChatList(folder) {
    const conversations = readConversationsFromDOM();
    const selected = new Set((folder.conversation_ids || []).map(String));
    folderChatList.innerHTML = conversations.length ? conversations.map(function (conversation) {
      const checked = selected.has(conversation.id) ? " checked" : "";
      return [
        '<label class="settings-folder-chat-item">',
        '<input type="checkbox" value="' + escapeHTML(conversation.id) + '"' + checked + '>',
        '<span class="settings-folder-chat-avatar">' + escapeHTML(conversation.title.slice(0, 1).toUpperCase()) + "</span>",
        "<span>",
        "<strong>" + escapeHTML(conversation.title) + "</strong>",
        "<small>" + escapeHTML(conversation.preview) + "</small>",
        "</span>",
        "</label>"
      ].join("");
    }).join("") : [
      '<div class="settings-folder-empty">',
      "<strong>目前沒有可加入的聊天室</strong>",
      "<span>請先載入或建立對話。</span>",
      "</div>"
    ].join("");
  }

  function openFolderEditor(folderIDValue) {
    const folder = findFolder(folderIDValue);
    if (!folder) {
      return;
    }
    activeFolderID = folder.id;
    folderEditNameInput.value = folder.name;
    if (folderEditMenuPanel) {
      folderEditMenuPanel.hidden = true;
    }
    if (folderEditStatus) {
      folderEditStatus.hidden = true;
      folderEditStatus.textContent = "已儲存";
      folderEditStatus.classList.remove("is-error");
      folderEditStatus.classList.add("is-success");
    }
    renderFolderChatList(folder);
    setFolderView("edit");
  }

  function saveActiveFolder() {
    if (!activeFolderID) {
      return;
    }
    const folders = readFolders();
    const target = folders.find(function (folder) {
      return folder.id === activeFolderID;
    });
    if (!target) {
      return;
    }
    const selectedIDs = Array.from(folderChatList.querySelectorAll('input[type="checkbox"]:checked')).map(function (input) {
      return String(input.value);
    });
    target.name = folderEditNameInput.value.trim() || target.name;
    target.conversation_ids = selectedIDs;
    const saved = saveFolders(folders);
    renderFolderCards(saved);
    const updated = saved.find(function (folder) {
      return folder.id === activeFolderID;
    });
    if (updated) {
      folderEditNameInput.value = updated.name;
      renderFolderChatList(updated);
    }
    if (folderEditStatus) {
      folderEditStatus.hidden = false;
      folderEditStatus.textContent = "已儲存";
      folderEditStatus.classList.remove("is-error");
      folderEditStatus.classList.add("is-success");
    }
    applyConversationFolderFilter();
  }

  function deleteFolder(folderIDValue) {
    const saved = saveFolders(readFolders().filter(function (folder) {
      return folder.id !== folderIDValue && folder.name.toUpperCase() !== "ALL";
    }));
    if (activeFolderTabID === folderIDValue) {
      activeFolderTabID = "";
      renderFolderTabs(saved);
      applyConversationFolderFilter();
    }
    renderFolderCards(saved);
    if (activeFolderID === folderIDValue) {
      activeFolderID = "";
      setFolderView("list");
    }
  }

  function renderLoggedOutSettings() {
    activeFolderID = "";
    activeFolderTabID = "";
    renderProfile();
    if (folderTabs) {
      folderTabs.hidden = true;
      folderTabs.innerHTML = "";
    }
    if (folderList) {
      folderList.innerHTML = "";
    }
    if (folderChatList) {
      folderChatList.innerHTML = "";
    }
    setFolderView("list");
    applyConversationFolderFilter();
  }

  function renderLoggedInSettings() {
    const activeFolderKey = accountStorageKey(activeFolderStorageKey);
    activeFolderTabID = activeFolderKey ? localStorage.getItem(activeFolderKey) || "" : "";
    renderProfile();
    const folders = readFolders();
    renderFolderCards(folders);
    renderFolderTabs(folders);
    applyConversationFolderFilter();
    setFolderView("list");
  }

  function renderSettingsForSession() {
    if (!isLoggedIn()) {
      renderLoggedOutSettings();
      return;
    }
    renderLoggedInSettings();
  }

  if (avatarInput && avatarPreview) {
    avatarInput.addEventListener("change", function () {
      const file = avatarInput.files && avatarInput.files[0];
      if (!file) {
        return;
      }

      const reader = new FileReader();
      reader.onload = function () {
        const dataURL = String(reader.result || "");
        const key = avatarStorageKey();
        if (key) {
          localStorage.setItem(key, dataURL);
        }
        applyAvatarImage(dataURL);
        document.dispatchEvent(new CustomEvent("twacc:avatar-changed"));
      };
      reader.readAsDataURL(file);
    });
  }

  if (avatarEditTrigger && avatarInput) {
    avatarEditTrigger.addEventListener("click", function (event) {
      if (avatarEditTrigger.dataset.avatarDisabled === "true") {
        event.preventDefault();
        return;
      }
      avatarInput.click();
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
      const folder = {
        id: "folder_" + Date.now(),
        name: nextFolderName(folders),
        conversation_ids: []
      };
      folders.push(folder);
      saveFolders(folders);
      renderFolderCards(readFolders());
      openFolderEditor(folder.id);
    });
  }

  if (folderList) {
    folderList.addEventListener("click", function (event) {
      const card = event.target.closest("[data-folder-id]");
      if (!card) {
        return;
      }

      if (event.target.closest("[data-folder-more]")) {
        const menu = card.querySelector("[data-folder-menu]");
        if (menu) {
          menu.hidden = !menu.hidden;
        }
        return;
      }

      if (event.target.closest("[data-folder-delete]")) {
        deleteFolder(card.dataset.folderId);
        return;
      }

      if (event.target.closest("[data-folder-open]")) {
        openFolderEditor(card.dataset.folderId);
      }
    });
  }

  if (folderEditBackButton) {
    folderEditBackButton.addEventListener("click", function () {
      saveActiveFolder();
      setFolderView("list");
    });
  }

  if (folderEditSaveButton) {
    folderEditSaveButton.addEventListener("click", saveActiveFolder);
  }

  if (folderEditNameInput) {
    folderEditNameInput.addEventListener("keydown", function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        saveActiveFolder();
      }
    });
  }

  if (folderEditMenuButton && folderEditMenuPanel) {
    folderEditMenuButton.addEventListener("click", function () {
      folderEditMenuPanel.hidden = !folderEditMenuPanel.hidden;
    });
  }

  if (folderEditDeleteButton) {
    folderEditDeleteButton.addEventListener("click", function () {
      deleteFolder(activeFolderID);
    });
  }

  if (folderTabs) {
    folderTabs.addEventListener("click", function (event) {
      const button = event.target.closest("[data-folder-tab-id]");
      if (!button) {
        return;
      }
      const folderIDValue = button.dataset.folderTabId || "";
      activeFolderTabID = activeFolderTabID === folderIDValue ? "" : folderIDValue;
      const key = accountStorageKey(activeFolderStorageKey);
      if (activeFolderTabID) {
        if (key) {
          localStorage.setItem(key, activeFolderTabID);
        }
      } else if (key) {
        localStorage.removeItem(key);
      }
      renderFolderTabs(readFolders());
      applyConversationFolderFilter();
    });
  }

  openButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      setSubview(button.dataset.settingsOpen);
      if (button.dataset.settingsOpen === "folders") {
        setFolderView("list");
      }
    });
  });

  if (editProfileButton) {
    editProfileButton.addEventListener("click", function () {
      renderProfile();
      setSubview("profile-edit");
      updateProfileSaveVisibility();
    });
  }

  if (profileEditName) {
    profileEditName.addEventListener("input", updateProfileSaveVisibility);
    profileEditName.addEventListener("keydown", function (event) {
      if (event.key !== "Enter") {
        return;
      }
      event.preventDefault();
    });
  }

  if (backButton) {
    backButton.addEventListener("click", function () {
      setSubview("main");
      setFolderView("list");
    });
  }

  document.addEventListener("twacc:sidebar-view-changed", function (event) {
    const view = event.detail && event.detail.view;
    if (view !== "settings") {
      return;
    }
    renderSettingsForSession();
    setSubview("main");
    setFolderView("list");
  });

  document.addEventListener("twacc:sidebar-back", function (event) {
    const detail = event.detail || {};
    if (detail.action !== "settings-main") {
      return;
    }
    setSubview("main");
    setFolderView("list");
  });

  if (profileEditSaveButton) {
    profileEditSaveButton.addEventListener("click", async function () {
      profileEditSaveButton.disabled = true;
      try {
        const saved = await saveProfileName();
        if (saved) {
          setSubview("main");
        }
      } catch (_) {
      } finally {
        profileEditSaveButton.disabled = false;
      }
    });
  }

  document.addEventListener("twacc:session-changed", renderSettingsForSession);

  renderSettingsForSession();
  setSubview("main");
})();
