(function () {
  const app = document.querySelector("[data-auth-app]");
  if (!app) {
    return;
  }

  const storageKeys = {
    token: "twacc_chat_session_token",
    sourceSystem: "twacc_chat_source_system",
    activeConversationID: "twacc_chat_active_conversation_id",
    activeFolderID: "twacc_chat_active_folder_tab_id",
    folderCategories: "twacc_chat_folder_categories",
    conversationCatalog: "twacc_chat_conversation_catalog"
  };

  const conversationList = app.querySelector("[data-conversation-list]");
  const conversationSummary = app.querySelector("[data-conversation-summary]");
  const conversationSearchInput = app.querySelector("[data-conversation-search]");
  const sidebarShell = app.querySelector("[data-sidebar-shell]");
  const conversationSearchPanel = app.querySelector("[data-conversation-search-panel]");
  const searchDirectSection = app.querySelector("[data-search-direct-section]");
  const searchDirectResults = app.querySelector("[data-search-direct-results]");
  const searchMessageSection = app.querySelector("[data-search-message-section]");
  const searchMessageResults = app.querySelector("[data-search-message-results]");
  const searchAllMessagesButton = app.querySelector("[data-search-all-messages]");
  const conversationCreateActions = app.querySelector("[data-conversation-create-actions]");
  const conversationCreateToggle = app.querySelector("[data-conversation-create-toggle]");
  const conversationCreateMenu = app.querySelector("[data-conversation-create-menu]");
  const newGroupSearchInput = app.querySelector("[data-new-group-search]");
  const newGroupSelectedNode = app.querySelector("[data-new-group-selected]");
  const newGroupResultsNode = app.querySelector("[data-new-group-results]");
  const newGroupNextButton = app.querySelector("[data-new-group-next]");
  const newGroupNameInput = app.querySelector("[data-new-group-name]");
  const newGroupMemberCountNode = app.querySelector("[data-new-group-member-count]");
  const newGroupMemberSummaryNode = app.querySelector("[data-new-group-member-summary]");
  const newGroupCreateButton = app.querySelector("[data-new-group-create]");
  const contactsList = app.querySelector("[data-contacts-list]");
  const folderTabs = app.querySelector("[data-folder-tabs]");
  const messageBoard = app.querySelector("[data-message-board]");
  const messageBoardShell = messageBoard ? messageBoard.closest(".message-board-shell") : null;
  const messageForm = app.querySelector("[data-message-form]");
  const chatShell = app.querySelector(".chat-shell");
  const conversationTitle = app.querySelector("[data-conversation-title]");
  const messageStatus = app.querySelector("[data-message-status]");
  const contactMenuToggle = app.querySelector("[data-contact-menu-toggle]");
  const contactMenu = app.querySelector("[data-contact-menu]");
  const contactPanel = app.querySelector("[data-contact-panel]");
  const contactPanelClose = app.querySelector("[data-contact-panel-close]");
  const contactForm = app.querySelector("[data-contact-form]");
  const contactAvatar = app.querySelector("[data-contact-avatar]");
  const contactDisplayName = app.querySelector("[data-contact-display-name]");
  const contactOriginalName = app.querySelector("[data-contact-original-name]");
  const contactAliasName = app.querySelector("[data-contact-alias-name]");
  const contactAccount = app.querySelector("[data-contact-account]");
  const realtimeStateNode = app.querySelector("[data-realtime-state]");
  const composerInput = messageForm.querySelector(".composer-input");
  const composerFileInput = messageForm.querySelector(".composer-file-input");
  const fileStateNode = messageForm.querySelector("[data-file-state]");
  const attachmentToggle = messageForm.querySelector("[data-attachment-toggle]");
  const attachmentMenu = messageForm.querySelector("[data-attachment-menu]");
  const attachmentDialog = app.querySelector("[data-attachment-dialog]");
  const attachmentTitle = app.querySelector("[data-attachment-title]");
  const attachmentPreview = app.querySelector("[data-attachment-preview]");
  const attachmentCaption = app.querySelector("[data-attachment-caption]");
  const attachmentClose = app.querySelector("[data-attachment-close]");
  const attachmentSend = app.querySelector("[data-attachment-send]");
  const chatDropOverlay = app.querySelector("[data-chat-drop-overlay]");
  const chatDropZones = app.querySelectorAll("[data-drop-kind]");
  const messageContextMenu = app.querySelector("[data-message-context-menu]");
  const replyPreview = app.querySelector("[data-reply-preview]");
  const replySenderNode = app.querySelector("[data-reply-sender]");
  const replyExcerptNode = app.querySelector("[data-reply-excerpt]");
  const replyCancelButton = app.querySelector("[data-reply-cancel]");

  let conversations = [];
  let activeConversationID = 0;
  let activeFolderID = "";
  let conversationSearchQuery = "";
  let pendingDirectTarget = null;
  let userSearchRequestToken = 0;
  let groupUserSearchRequestToken = 0;
  let messageSearchRequestToken = 0;
  let realtimeSocket = null;
  let reconnectTimer = 0;
  let reconnectAttempts = 0;
  let messageLoadToken = 0;
  let selectedAttachmentKind = "document";
  let selectedAttachmentFiles = [];
  let attachmentPreviewURLs = [];
  let dragDepth = 0;
  let contextMessage = null;
  let replyMessage = null;
  let newGroupSearchQuery = "";
  let newGroupSelectedMembers = [];
  let contactsLoadToken = 0;
  let contactKeys = new Set();
  let contactExternalIDs = new Set();
  let contactsLoaded = false;
  let contactsIndexLoading = false;

  function readSessionToken() {
    return localStorage.getItem(storageKeys.token) || "";
  }

  function readSourceSystem() {
    return localStorage.getItem(storageKeys.sourceSystem) || "erp";
  }

  function readExternalUserID() {
    return localStorage.getItem("twacc_chat_external_user_id") || "";
  }

  function accountStorageKey(base) {
    const token = readSessionToken();
    const sourceSystem = readSourceSystem();
    const externalUserID = readExternalUserID();
    if (!token || !sourceSystem || !externalUserID) {
      return "";
    }
    return base + "::" + encodeURIComponent(sourceSystem + ":" + externalUserID);
  }

  function loadSessionScopedState() {
    const activeConversationKey = accountStorageKey(storageKeys.activeConversationID);
    const activeFolderKey = accountStorageKey(storageKeys.activeFolderID);
    activeConversationID = activeConversationKey ? Number(localStorage.getItem(activeConversationKey) || 0) : 0;
    activeFolderID = activeFolderKey ? localStorage.getItem(activeFolderKey) || "" : "";
  }

  function authHeaders() {
    const token = readSessionToken();
    if (!token) {
      return null;
    }
    return {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token
    };
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

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function linkifyMessageText(value) {
    const escaped = escapeHTML(value);
    return escaped.replace(/\bhttps?:\/\/[^\s<]+/gi, function (url) {
      return '<a href="' + url + '" target="_blank" rel="noreferrer noopener">' + url + "</a>";
    });
  }

  function formatTime(value) {
    if (!value) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString();
  }

  function formatBytes(size) {
    if (!size) {
      return "0 B";
    }
    if (size < 1024) {
      return size + " B";
    }
    if (size < 1024 * 1024) {
      return (size / 1024).toFixed(1) + " KB";
    }
    return (size / (1024 * 1024)).toFixed(1) + " MB";
  }

  function setMessageStatus(text, isError) {
    messageStatus.textContent = text;
    messageStatus.classList.toggle("is-error", Boolean(isError));
  }

  function setConversationTitle(text) {
    if (!conversationTitle) {
      return;
    }
    conversationTitle.textContent = text || "目前對話";
  }

  function setConversationUIActive(isActive) {
    app.classList.toggle("is-conversation-active", Boolean(isActive));
    if (!isActive) {
      setContactMenuOpen(false);
      setContactPanelOpen(false);
    }
    updateContactActions();
  }

  function setContactMenuOpen(isOpen) {
    if (!contactMenu || !contactMenuToggle) {
      return;
    }
    contactMenu.hidden = !isOpen;
    contactMenuToggle.setAttribute("aria-expanded", isOpen ? "true" : "false");
  }

  function setContactPanelOpen(isOpen) {
    if (!contactPanel) {
      return;
    }
    contactPanel.hidden = !isOpen;
    const stage = contactPanel.closest(".desktop-stage");
    if (stage) {
      stage.classList.toggle("is-contact-panel-open", Boolean(isOpen));
    }
  }

  function updateContactActions() {
    if (!contactMenuToggle) {
      return;
    }
    const peer = activeDirectPeer();
    if (peer && !contactsLoaded) {
      contactMenuToggle.hidden = true;
      ensureContactsLoaded();
      return;
    }
    const canAddContact = Boolean(peer) && !isContactPeer(peer);
    contactMenuToggle.hidden = !canAddContact;
    if (!canAddContact) {
      setContactMenuOpen(false);
      setContactPanelOpen(false);
    }
  }

  function fillContactPanel(peer) {
    const title = peer.title || peer.externalUserID || "聯絡人";
    const initial = title.slice(0, 1).toUpperCase();
    if (contactAvatar) {
      contactAvatar.textContent = initial || "?";
    }
    if (contactDisplayName) {
      contactDisplayName.textContent = title;
    }
    if (contactOriginalName) {
      contactOriginalName.textContent = "original name";
    }
    if (contactAliasName) {
      contactAliasName.value = title;
    }
    if (contactAccount) {
      contactAccount.value = peer.externalUserID || "";
    }
  }

  function activeConversationTitle() {
    if (pendingDirectTarget) {
      return pendingDirectTarget.title || pendingDirectTarget.externalUserID || "";
    }
    const active = conversations.find(function (item) {
      return item.conversation_id === activeConversationID;
    });
    return active ? active.title : "";
  }

  function activeDirectPeer() {
    if (pendingDirectTarget) {
      return {
        sourceSystem: pendingDirectTarget.sourceSystem || readSourceSystem(),
        externalUserID: pendingDirectTarget.externalUserID || "",
        title: pendingDirectTarget.title || pendingDirectTarget.externalUserID || ""
      };
    }
    const active = conversations.find(function (item) {
      return item.conversation_id === activeConversationID;
    });
    if (!active || active.type !== "direct" || !active.direct_external_user_id) {
      return null;
    }
    return {
      sourceSystem: active.direct_source_system || readSourceSystem(),
      externalUserID: active.direct_external_user_id || "",
      title: active.title || active.direct_external_user_id || ""
    };
  }

  function contactKey(sourceSystem, externalUserID) {
    return String(sourceSystem || readSourceSystem()).trim() + ":" + String(externalUserID || "").trim().toLowerCase();
  }

  function isContactPeer(peer) {
    if (!peer || !peer.externalUserID) {
      return false;
    }
    const externalID = String(peer.externalUserID || "").trim().toLowerCase();
    return contactKeys.has(contactKey(peer.sourceSystem, peer.externalUserID)) || contactExternalIDs.has(externalID);
  }

  function findExistingDirectConversation(sourceSystem, externalUserID) {
    const source = String(sourceSystem || "").trim();
    const externalID = String(externalUserID || "").trim().toLowerCase();
    if (!source || !externalID) {
      return null;
    }
    return conversations.find(function (item) {
      return item.type === "direct"
        && String(item.direct_source_system || "").trim() === source
        && String(item.direct_external_user_id || "").trim().toLowerCase() === externalID;
    }) || null;
  }

  function openConversation(conversationID) {
    pendingDirectTarget = null;
    activeConversationID = Number(conversationID || 0);
    if (!activeConversationID) {
      return;
    }
    const key = accountStorageKey(storageKeys.activeConversationID);
    if (key) {
      localStorage.setItem(key, String(activeConversationID));
    }
    closeSearchMode();
    renderConversationList(conversations);
    updateContactActions();
    loadMessages(activeConversationID);
  }

  function readFolderCategories() {
    const key = accountStorageKey(storageKeys.folderCategories);
    if (!key) {
      return [];
    }
    try {
      const raw = localStorage.getItem(key);
      return raw ? JSON.parse(raw) : [];
    } catch (_) {
      return [];
    }
  }

  function filteredConversations(items) {
    let visibleItems = items;
    if (activeFolderID) {
      const folder = readFolderCategories().find(function (item) {
        return String(item.id || "") === activeFolderID;
      });
      if (folder && Array.isArray(folder.conversation_ids)) {
        const allowed = new Set(folder.conversation_ids.map(String));
        visibleItems = visibleItems.filter(function (item) {
          return allowed.has(String(item.conversation_id));
        });
      }
    }
    if (!conversationSearchQuery) {
      return visibleItems;
    }
    const query = conversationSearchQuery.toLowerCase();
    return visibleItems.filter(function (item) {
      const text = [
        item.title || "",
        conversationPreview(item),
        item.last_message_type || ""
      ].join(" ").toLowerCase();
      return text.indexOf(query) >= 0;
    });
  }

  function syncConversationCatalog(items) {
    const key = accountStorageKey(storageKeys.conversationCatalog);
    if (!key) {
      return;
    }
    localStorage.setItem(key, JSON.stringify(items.map(function (item) {
      return {
        id: String(item.conversation_id),
        title: item.title || "",
        preview: conversationPreview(item)
      };
    })));
  }

  function renderEmptyConversationState(title, detail) {
    const stateTitle = title || "請選擇對話對象，開始傳訊息";
    const stateDetail = detail || stateTitle;
    pendingDirectTarget = null;
    setConversationUIActive(false);
    setConversationTitle(stateTitle);
    setMessageStatus(stateDetail, false);
    messageBoard.innerHTML = "";
  }

  function renderLoadingConversationState() {
    messageBoard.innerHTML = [
      '<div class="empty-chat-state">',
      "<strong>載入中...</strong>",
      "</div>"
    ].join("");
  }

  function closeActiveConversation() {
    if (!activeConversationID && !pendingDirectTarget) {
      return;
    }
    pendingDirectTarget = null;
    activeConversationID = 0;
    const key = accountStorageKey(storageKeys.activeConversationID);
    if (key) {
      localStorage.removeItem(key);
    }
    renderConversationList(conversations);
    renderEmptyConversationState();
  }

  function setRealtimeState(text, isConnected) {
    if (!realtimeStateNode) {
      return;
    }
    realtimeStateNode.textContent = text;
    realtimeStateNode.classList.toggle("is-live", Boolean(isConnected));
  }

  function renderConversationList(items) {
    conversations = items;
    syncConversationCatalog(items);

    const visibleItems = filteredConversations(items);
    if (!visibleItems.length) {
      const isSearching = Boolean(conversationSearchQuery);
      conversationList.innerHTML = [
        '<article class="conversation-card is-placeholder">',
        "<strong>" + (isSearching ? "找不到對話" : (items.length ? "此分類尚無對話" : "目前沒有對話")) + "</strong>",
        "<span>" + (isSearching ? "請嘗試其他關鍵字" : (items.length ? "請到聊天室分類加入對話" : "先建立一個一對一對話")) + "</span>",
        "<small>" + (isSearching ? "搜尋中" : (items.length ? "等待分類" : "等待建立")) + "</small>",
        "</article>"
      ].join("");
      conversationSummary.textContent = isSearching ? "沒有符合搜尋條件的對話。" : (items.length ? "此聊天室分類目前沒有對應對話。" : "目前尚無對話，請先建立一個一對一對話。");
      if (!pendingDirectTarget) {
        renderEmptyConversationState("請選擇對話對象，開始傳訊息");
      }
      return;
    }

    conversationList.innerHTML = visibleItems.map(function (item) {
      const active = item.conversation_id === activeConversationID ? " is-active" : "";
      const preview = conversationPreviewExcerpt(item);
      const meta = formatTime(item.last_message_at) || (item.member_count + " 位成員");
      const unread = item.unread_count > 0
        ? ('<span class="conversation-unread-badge" aria-label="未讀訊息 ' + item.unread_count + ' 則">' + item.unread_count + "</span>")
        : "";
      return [
        '<button type="button" class="conversation-card conversation-button' + active + '" data-conversation-id="' + item.conversation_id + '">',
        '<span class="conversation-card-header"><strong>' + escapeHTML(item.title) + "</strong>" + unread + "</span>",
        "<span>" + escapeHTML(preview) + "</span>",
        "<small>" + escapeHTML(meta) + "</small>",
        "</button>"
      ].join("");
    }).join("");

    conversationSummary.textContent = "已載入 " + visibleItems.length + " 筆對話。";

    conversationList.querySelectorAll("[data-conversation-id]").forEach(function (button) {
      button.addEventListener("click", function () {
        openConversation(button.dataset.conversationId || 0);
      });
    });
  }

  function normalizedSearchExternalID() {
    return conversationSearchQuery.replace(/^@+/, "").trim();
  }

  function setSearchMode(active) {
    if (sidebarShell) {
      sidebarShell.classList.toggle("is-searching", Boolean(active));
    }
    if (conversationSearchPanel) {
      conversationSearchPanel.hidden = !active;
    }
    if (folderTabs) {
      folderTabs.hidden = Boolean(active);
    }
    if (conversationList) {
      conversationList.hidden = Boolean(active);
    }
  }

  function renderDirectSearchResults(items, loading) {
    if (!searchDirectSection || !searchDirectResults) {
      return;
    }
    if (!conversationSearchQuery) {
      searchDirectSection.hidden = true;
      searchDirectResults.innerHTML = "";
      return;
    }
    searchDirectSection.hidden = false;
    if (loading) {
      searchDirectSection.hidden = true;
      searchDirectResults.innerHTML = "";
      return;
    }
    if (!items.length) {
      searchDirectSection.hidden = true;
      searchDirectResults.innerHTML = "";
      return;
    }
    searchDirectSection.hidden = false;
    searchDirectResults.innerHTML = items.map(function (item) {
      const title = item.display_name || item.external_user_id || "使用者";
      const externalID = item.external_user_id || "";
      const sourceSystem = item.source_system || readSourceSystem();
      const initial = title.slice(0, 1).toUpperCase();
      return [
        '<button type="button" class="conversation-search-result" data-search-direct-id="' + escapeHTML(externalID) + '" data-search-direct-title="' + escapeHTML(title) + '" data-search-direct-source="' + escapeHTML(sourceSystem) + '">',
        '<span class="conversation-search-avatar">' + escapeHTML(initial) + '</span>',
        '<span class="conversation-search-text">',
        '<strong>' + escapeHTML(title) + '</strong>',
        '<small>@' + escapeHTML(externalID) + '</small>',
        '</span>',
        '</button>'
      ].join("");
    }).join("");
    searchDirectResults.querySelectorAll("[data-search-direct-id]").forEach(function (button) {
      button.addEventListener("click", function () {
        openPendingDirectConversation({
          sourceSystem: button.dataset.searchDirectSource || "",
          externalUserID: button.dataset.searchDirectId || "",
          title: button.dataset.searchDirectTitle || ""
        });
      });
    });
  }

  async function searchUsers() {
    const headers = authHeaders();
    const externalID = normalizedSearchExternalID();
    if (!headers || !externalID) {
      renderDirectSearchResults([], false);
      return;
    }
    const requestToken = ++userSearchRequestToken;
    renderDirectSearchResults([], true);
    const response = await fetch("/api/users/search?source_system=" + encodeURIComponent(readSourceSystem()) + "&q=" + encodeURIComponent(externalID), {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (requestToken !== userSearchRequestToken) {
      return;
    }
    if (!response.ok || !result.success) {
      renderDirectSearchResults([], false);
      return;
    }
    const data = result.data || {};
    renderDirectSearchResults(data.users || [], false);
  }

  function renderMessageSearchResults(items, loading) {
    if (!searchMessageSection || !searchMessageResults) {
      return;
    }
    searchMessageSection.hidden = !conversationSearchQuery;
    if (!conversationSearchQuery) {
      searchMessageResults.innerHTML = "";
      return;
    }
    if (loading) {
      searchMessageResults.innerHTML = '<div class="conversation-search-empty">搜尋聊天記錄中...</div>';
      return;
    }
    if (!items.length) {
      searchMessageResults.innerHTML = '<div class="conversation-search-empty">沒有符合的聊天記錄</div>';
      return;
    }
    searchMessageResults.innerHTML = items.map(function (item) {
      const title = item.conversation_title || "聊天記錄";
      const initial = title.slice(0, 1).toUpperCase();
      const sender = item.sender_name ? item.sender_name + "：" : "";
      return [
        '<button type="button" class="conversation-search-result" data-search-message-conversation-id="' + item.conversation_id + '">',
        '<span class="conversation-search-avatar is-muted">' + escapeHTML(initial) + '</span>',
        '<span class="conversation-search-text">',
        '<strong>' + escapeHTML(title) + '</strong>',
        '<small>' + escapeHTML(sender + (item.content || "")) + '</small>',
        '</span>',
        '<span class="conversation-search-date">' + escapeHTML(formatTime(item.created_at)) + '</span>',
        '</button>'
      ].join("");
    }).join("");
    searchMessageResults.querySelectorAll("[data-search-message-conversation-id]").forEach(function (button) {
      button.addEventListener("click", function () {
        const conversationID = Number(button.dataset.searchMessageConversationId || 0);
        if (!conversationID) {
          return;
        }
        openConversation(conversationID);
      });
    });
  }

  async function searchMessages() {
    const headers = authHeaders();
    if (!headers || !conversationSearchQuery) {
      renderMessageSearchResults([], false);
      return;
    }
    const requestToken = ++messageSearchRequestToken;
    renderMessageSearchResults([], true);
    const response = await fetch("/api/messages/search?q=" + encodeURIComponent(conversationSearchQuery), {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (requestToken !== messageSearchRequestToken) {
      return;
    }
    if (!response.ok || !result.success) {
      renderMessageSearchResults([], false);
      return;
    }
    const data = result.data || {};
    renderMessageSearchResults(data.messages || [], false);
  }

  function renderSearchMode() {
    const searching = Boolean(conversationSearchQuery);
    setSearchMode(searching);
    if (!searching) {
      renderDirectSearchResults([], false);
      renderMessageSearchResults([], false);
      return;
    }
    searchUsers();
    searchMessages();
  }

  function closeSearchMode() {
    conversationSearchQuery = "";
    if (conversationSearchInput) {
      conversationSearchInput.value = "";
    }
    setConversationCreateMenuOpen(false);
    setSearchMode(false);
    renderDirectSearchResults([], false);
    renderMessageSearchResults([], false);
  }

  function setConversationCreateMenuOpen(open) {
    if (!conversationCreateActions || !conversationCreateToggle || !conversationCreateMenu) {
      return;
    }
    conversationCreateActions.classList.toggle("is-open", Boolean(open));
    conversationCreateToggle.setAttribute("aria-expanded", open ? "true" : "false");
    conversationCreateMenu.hidden = !open;
  }

  function memberKey(member) {
    return String((member && member.sourceSystem) || readSourceSystem()) + ":" + String((member && member.externalUserID) || "");
  }

  function userSearchItemToGroupMember(item) {
    const title = item.display_name || item.external_user_id || "使用者";
    return {
      sourceSystem: item.source_system || readSourceSystem(),
      externalUserID: item.external_user_id || "",
      title: title
    };
  }

  function isNewGroupMemberSelected(member) {
    const key = memberKey(member);
    return newGroupSelectedMembers.some(function (item) {
      return memberKey(item) === key;
    });
  }

  function updateNewGroupSelectedMember(member, selected) {
    const key = memberKey(member);
    newGroupSelectedMembers = newGroupSelectedMembers.filter(function (item) {
      return memberKey(item) !== key;
    });
    if (selected) {
      newGroupSelectedMembers.push(member);
    }
    renderNewGroupSelectedMembers();
  }

  function renderNewGroupSelectedMembers() {
    if (!newGroupSelectedNode || !newGroupNextButton) {
      return;
    }
    const pickerNode = newGroupSelectedNode.closest(".new-group-picker");
    newGroupSelectedNode.hidden = !newGroupSelectedMembers.length;
    newGroupNextButton.disabled = newGroupSelectedMembers.length < 1;
    if (pickerNode) {
      pickerNode.classList.toggle("has-selected", newGroupSelectedMembers.length > 0);
    }
    newGroupSelectedNode.innerHTML = newGroupSelectedMembers.map(function (member) {
      const title = member.title || member.externalUserID || "使用者";
      const initial = title.slice(0, 1).toUpperCase();
      return [
        '<span class="new-group-selected-chip" data-new-group-selected-key="' + escapeHTML(memberKey(member)) + '">',
        '<span class="new-group-selected-avatar">' + escapeHTML(initial) + "</span>",
        "<span>" + escapeHTML(title) + "</span>",
        '<button type="button" aria-label="移除 ' + escapeHTML(title) + '">×</button>',
        "</span>"
      ].join("");
    }).join("");
  }

  function renderNewGroupResults(items, loading) {
    if (!newGroupResultsNode) {
      return;
    }
    if (!newGroupSearchQuery) {
      newGroupResultsNode.innerHTML = '<div class="new-group-empty">輸入名稱或帳號搜尋成員</div>';
      return;
    }
    if (loading) {
      newGroupResultsNode.innerHTML = '<div class="new-group-empty">搜尋成員中...</div>';
      return;
    }
    if (!items.length) {
      newGroupResultsNode.innerHTML = '<div class="new-group-empty">找不到符合的成員</div>';
      return;
    }
    newGroupResultsNode.innerHTML = items.map(function (item) {
      const member = userSearchItemToGroupMember(item);
      const selected = isNewGroupMemberSelected(member);
      const initial = (member.title || member.externalUserID || "U").slice(0, 1).toUpperCase();
      return [
        '<button type="button" class="new-group-member' + (selected ? " is-selected" : "") + '"',
        ' data-new-group-source="' + escapeHTML(member.sourceSystem) + '"',
        ' data-new-group-id="' + escapeHTML(member.externalUserID) + '"',
        ' data-new-group-title="' + escapeHTML(member.title) + '">',
        '<span class="new-group-member-check" aria-hidden="true"></span>',
        '<span class="new-group-member-avatar">' + escapeHTML(initial) + "</span>",
        '<span class="new-group-member-text">',
        "<strong>" + escapeHTML(member.title) + "</strong>",
        "<small>@" + escapeHTML(member.externalUserID) + "</small>",
        "</span>",
        "</button>"
      ].join("");
    }).join("");
  }

  async function searchNewGroupUsers() {
    const headers = authHeaders();
    const query = String(newGroupSearchQuery || "").replace(/^@+/, "").trim();
    if (!headers || !query) {
      renderNewGroupResults([], false);
      return;
    }
    const requestToken = ++groupUserSearchRequestToken;
    renderNewGroupResults([], true);
    const response = await fetch("/api/users/search?source_system=" + encodeURIComponent(readSourceSystem()) + "&q=" + encodeURIComponent(query), {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (requestToken !== groupUserSearchRequestToken) {
      return;
    }
    if (!response.ok || !result.success) {
      renderNewGroupResults([], false);
      return;
    }
    const data = result.data || {};
    renderNewGroupResults(data.users || [], false);
  }

  function resetNewGroupDraft() {
    newGroupSearchQuery = "";
    newGroupSelectedMembers = [];
    if (newGroupSearchInput) {
      newGroupSearchInput.value = "";
    }
    if (newGroupNameInput) {
      newGroupNameInput.value = "";
    }
    renderNewGroupSelectedMembers();
    renderNewGroupResults([], false);
    renderNewGroupDetails();
  }

  function suggestedNewGroupName() {
    return newGroupSelectedMembers.map(function (member) {
      return member.title || member.externalUserID || "使用者";
    }).join("、");
  }

  function renderNewGroupDetails() {
    if (!newGroupMemberCountNode || !newGroupMemberSummaryNode) {
      return;
    }
    const totalMembers = newGroupSelectedMembers.length + 1;
    newGroupMemberCountNode.textContent = totalMembers + " 名成員";
    newGroupMemberSummaryNode.innerHTML = newGroupSelectedMembers.map(function (member) {
      const initial = (member.title || member.externalUserID || "U").slice(0, 1).toUpperCase();
      return [
        '<div class="new-group-summary-member">',
        '<span class="new-group-member-avatar">' + escapeHTML(initial) + "</span>",
        '<span class="new-group-member-text">',
        "<strong>" + escapeHTML(member.title || member.externalUserID || "使用者") + "</strong>",
        "<small>@" + escapeHTML(member.externalUserID || "") + "</small>",
        "</span>",
        "</div>"
      ].join("");
    }).join("");
  }

  async function createGroupConversation() {
    const headers = authHeaders();
    const name = newGroupNameInput ? newGroupNameInput.value.trim() : "";
    if (!headers) {
      setMessageStatus("請先登入。", true);
      return;
    }
    if (!name) {
      setMessageStatus("請輸入群組名稱。", true);
      if (newGroupNameInput) {
        newGroupNameInput.focus();
      }
      return;
    }
    if (!newGroupSelectedMembers.length) {
      setMessageStatus("請至少選擇一位成員。", true);
      return;
    }
    if (newGroupCreateButton) {
      newGroupCreateButton.disabled = true;
    }
    setMessageStatus("建立群組中...", false);
    const response = await fetch("/api/conversations/group", {
      method: "POST",
      headers: headers,
      body: JSON.stringify({
        name: name,
        members: newGroupSelectedMembers.map(function (member) {
          return {
            source_system: member.sourceSystem,
            external_user_id: member.externalUserID
          };
        })
      })
    });
    const result = await parseJSON(response);
    if (newGroupCreateButton) {
      newGroupCreateButton.disabled = false;
    }
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "群組建立失敗。", true);
      return;
    }
    const data = result.data || {};
    resetNewGroupDraft();
    document.dispatchEvent(new CustomEvent("twacc:sidebar-home"));
    await loadConversations();
    if (data.conversation_id) {
      openConversation(Number(data.conversation_id));
    }
    setMessageStatus("群組已建立。", false);
  }

  async function addActivePeerToContacts() {
    const headers = authHeaders();
    const peer = activeDirectPeer();
    if (!headers || !peer) {
      setMessageStatus("目前沒有可加入的聯絡人。", true);
      return;
    }
    const submitButton = contactForm ? contactForm.querySelector("button[type='submit']") : null;
    if (submitButton) {
      submitButton.disabled = true;
    }
    setMessageStatus("加入聯絡人中...", false);
    let response;
    let result = {};
    try {
      response = await fetch("/api/contacts", {
        method: "POST",
        headers: headers,
        body: JSON.stringify({
          source_system: peer.sourceSystem,
          external_user_id: peer.externalUserID,
          alias_name: contactAliasName ? contactAliasName.value.trim() : ""
        })
      });
      result = await parseJSON(response);
    } catch (_) {
      setMessageStatus("加入聯絡人失敗，請稍後再試。", true);
      if (submitButton) {
        submitButton.disabled = false;
      }
      return;
    }
    if (submitButton) {
      submitButton.disabled = false;
    }
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "加入聯絡人失敗。", true);
      return;
    }
    contactKeys.add(contactKey(peer.sourceSystem, peer.externalUserID));
    contactExternalIDs.add(String(peer.externalUserID || "").trim().toLowerCase());
    contactsLoaded = true;
    setContactPanelOpen(false);
    updateContactActions();
    if (sidebarShell && sidebarShell.dataset.sidebarView === "contacts") {
      loadContacts();
    }
    setMessageStatus("已加入聯絡人。", false);
  }

  function contactListTitle(item) {
    return item.alias_name || item.display_name || item.external_user_id || "聯絡人";
  }

  function renderContacts(items, loading) {
    if (!contactsList) {
      return;
    }
    if (loading) {
      contactsList.innerHTML = '<div class="contacts-empty">載入聯絡人中...</div>';
      return;
    }
    if (!items.length) {
      contactsList.innerHTML = '<div class="contacts-empty">尚未加入聯絡人</div>';
      return;
    }
    contactsList.innerHTML = items.map(function (item) {
      const title = contactListTitle(item);
      const initial = title.slice(0, 1).toUpperCase();
      return [
        '<button type="button" class="contact-list-item"',
        ' data-contact-source="' + escapeHTML(item.source_system || readSourceSystem()) + '"',
        ' data-contact-id="' + escapeHTML(item.external_user_id || "") + '"',
        ' data-contact-title="' + escapeHTML(title) + '">',
        '<span class="contact-list-avatar">' + escapeHTML(initial || "?") + '</span>',
        '<span class="contact-list-text">',
        '<strong>' + escapeHTML(title) + '</strong>',
        '<small>@' + escapeHTML(item.external_user_id || "") + '</small>',
        '</span>',
        '</button>'
      ].join("");
    }).join("");
    contactsList.querySelectorAll("[data-contact-id]").forEach(function (button) {
      button.addEventListener("click", function () {
        openPendingDirectConversation({
          sourceSystem: button.dataset.contactSource || readSourceSystem(),
          externalUserID: button.dataset.contactId || "",
          title: button.dataset.contactTitle || ""
        });
        document.dispatchEvent(new CustomEvent("twacc:sidebar-home"));
      });
    });
  }

  async function loadContacts() {
    const headers = authHeaders();
    if (!contactsList || !headers) {
      if (contactsList) {
        renderContacts([], false);
      }
      return;
    }
    const requestToken = ++contactsLoadToken;
    renderContacts([], true);
    const contacts = await fetchContacts(requestToken);
    if (!contacts) {
      return;
    }
    renderContacts(contacts, false);
  }

  async function fetchContacts(requestToken) {
    const headers = authHeaders();
    if (!headers) {
      contactKeys = new Set();
      contactExternalIDs = new Set();
      contactsLoaded = false;
      contactsIndexLoading = false;
      updateContactActions();
      return [];
    }
    let response;
    let result = {};
    try {
      contactsIndexLoading = true;
      response = await fetch("/api/contacts", {
        headers: { Authorization: headers.Authorization }
      });
      result = await parseJSON(response);
    } catch (_) {
      contactsIndexLoading = false;
      if (contactsList && requestToken === contactsLoadToken) {
        contactsList.innerHTML = '<div class="contacts-empty is-error">聯絡人載入失敗</div>';
      }
      return null;
    }
    if (requestToken && requestToken !== contactsLoadToken) {
      contactsIndexLoading = false;
      return null;
    }
    if (!response.ok || !result.success) {
      contactsIndexLoading = false;
      if (contactsList && requestToken === contactsLoadToken) {
        contactsList.innerHTML = '<div class="contacts-empty is-error">' + escapeHTML(result.message || "聯絡人載入失敗") + '</div>';
      }
      return null;
    }
    const data = result.data || {};
    const contacts = data.contacts || [];
    contactKeys = new Set(contacts.map(function (item) {
      return contactKey(item.source_system, item.external_user_id);
    }));
    contactExternalIDs = new Set(contacts.map(function (item) {
      return String(item.external_user_id || "").trim().toLowerCase();
    }).filter(Boolean));
    contactsLoaded = true;
    contactsIndexLoading = false;
    updateContactActions();
    return contacts;
  }

  async function ensureContactsLoaded() {
    if (contactsLoaded || contactsIndexLoading || !authHeaders()) {
      updateContactActions();
      return;
    }
    await fetchContacts(0);
  }

  function conversationPreview(item) {
    if (item.last_message_type === "image") {
      return "圖片";
    }
    if (item.last_message_type === "file") {
      return "附件";
    }
    const preview = currentMessageBodyText(item.last_message_preview || "");
    return preview || (item.type === "direct" ? "一對一對話" : "群組對話");
  }

  function conversationPreviewExcerpt(item) {
    const preview = conversationPreview(item);
    const limit = 15;
    if (preview.length <= limit) {
      return preview;
    }
    return preview.slice(0, limit) + "...";
  }

  function isOutgoingMessage(message) {
    const senderName = String(message && message.sender_name ? message.sender_name : "").trim();
    if (!senderName) {
      return false;
    }
    const displayName = String(localStorage.getItem("twacc_chat_display_name") || "").trim();
    const externalUserID = String(readExternalUserID() || "").trim();
    return senderName === displayName || senderName === externalUserID;
  }

  function trimMessageExcerpt(text) {
    const normalized = String(text || "").replace(/\s+/g, " ").trim();
    return normalized.length > 40 ? normalized.slice(0, 40) + "..." : normalized;
  }

  function currentMessageBodyText(content) {
    let text = String(content || "").trim();
    let nestedDepth = 0;
    while (nestedDepth < 5) {
      const reply = parseReplyContent(text);
      if (!reply) {
        break;
      }
      text = String(reply.body || "").trim();
      nestedDepth += 1;
    }
    return text;
  }

  function messageExcerpt(message) {
    const text = currentMessageBodyText(displayMessageContent(message));
    if (text) {
      return trimMessageExcerpt(text);
    }
    const attachments = messageAttachments(message);
    if (attachments.length) {
      return allAttachmentsAreImages(attachments) ? "圖片" : "附件";
    }
    return "訊息";
  }

  function parseReplyContent(content) {
    const text = String(content || "").replace(/\r\n/g, "\n");
    const match = text.match(/^回覆 ([^：]+)：(?:「([\s\S]*?)」)?\n([\s\S]*)$/);
    if (!match) {
      return null;
    }
    return {
      senderName: match[1] || "訊息",
      excerpt: match[2] || "",
      body: match[3] || ""
    };
  }

  function renderMessageContent(contentText) {
    if (!contentText) {
      return "";
    }
    const reply = parseReplyContent(contentText);
    if (!reply) {
      return "<p>" + linkifyMessageText(contentText) + "</p>";
    }
    const quote = [
      '<div class="message-reply-quote">',
      "<strong>" + escapeHTML(reply.senderName) + "</strong>",
      "<span>" + linkifyMessageText(reply.excerpt || "訊息") + "</span>",
      "</div>"
    ].join("");
    const body = reply.body ? "<p>" + linkifyMessageText(reply.body) + "</p>" : "";
    return quote + body;
  }

  function renderMessages(data) {
    if (!data.messages || !data.messages.length) {
      messageBoard.innerHTML = "";
      scrollMessageBoardToLatest();
      return;
    }

    const firstUnreadMessageID = Number(data.first_unread_message_id || 0);
    let unreadDividerRendered = false;
    messageBoard.innerHTML = data.messages.map(function (message) {
      const isOutgoing = isOutgoingMessage(message);
      const outgoing = isOutgoing ? " outgoing" : "";
      const attachments = messageAttachments(message);
      const photoMessage = allAttachmentsAreImages(attachments);
      const attachment = renderAttachments(message, attachments);
      const contentText = displayMessageContent(message);
      const content = renderMessageContent(contentText);
      const unreadDivider = firstUnreadMessageID
        && !unreadDividerRendered
        && Number(message.message_id) === firstUnreadMessageID
        ? '<div class="message-unread-divider" data-unread-divider><span>未讀訊息</span></div>'
        : "";
      if (unreadDivider) {
        unreadDividerRendered = true;
      }
      return unreadDivider + [
        '<div class="message-row' + outgoing + '">',
        '<div class="message-bubble' + (photoMessage ? " photo-bubble" : "") + '"',
        ' data-message-id="' + escapeHTML(message.message_id) + '"',
        ' data-sender-name="' + escapeHTML(message.sender_name) + '"',
        ' data-message-excerpt="' + escapeHTML(messageExcerpt(message)) + '"',
        ' data-outgoing="' + (isOutgoing ? "true" : "false") + '">',
        "<strong>" + escapeHTML(message.sender_name) + "</strong>",
        content,
        attachment,
        "<small>" + escapeHTML(formatTime(message.created_at)) + "</small>",
        "</div>",
        "</div>"
      ].join("");
    }).join("");
    if (!scrollMessageBoardToUnreadDivider() && !scrollMessageBoardToMessage(firstUnreadMessageID)) {
      scrollMessageBoardToLatest();
    }
  }

  function closeMessageContextMenu() {
    if (!messageContextMenu) {
      return;
    }
    messageContextMenu.hidden = true;
    contextMessage = null;
  }

  function findContextMessageBubble(target) {
    if (!target || !messageBoard || !target.closest) {
      return null;
    }
    const bubble = target.closest(".message-bubble");
    return bubble && messageBoard.contains(bubble) ? bubble : null;
  }

  function bubbleFromContextEvent(event) {
    let node = event.target;
    while (node && node !== messageBoard) {
      if (node.classList && node.classList.contains("message-bubble")) {
        return node;
      }
      node = node.parentElement;
    }
    return null;
  }

  function positionMessageContextMenu(event) {
    if (!messageContextMenu) {
      return;
    }
    messageContextMenu.hidden = false;
    const menuRect = messageContextMenu.getBoundingClientRect();
    const boardRect = messageBoard ? messageBoard.getBoundingClientRect() : { left: 0, top: 0, width: window.innerWidth, height: window.innerHeight };
    const shellRect = messageBoardShell ? messageBoardShell.getBoundingClientRect() : boardRect;
    const margin = 12;
    const gap = 2;
    const preferredLeft = event.clientX - shellRect.left + gap;
    const preferredTop = event.clientY - shellRect.top + gap;
    const left = Math.max(margin, Math.min(preferredLeft, shellRect.width - menuRect.width - margin));
    const top = Math.max(margin, Math.min(preferredTop, shellRect.height - menuRect.height - margin));
    messageContextMenu.style.left = left + "px";
    messageContextMenu.style.top = top + "px";
  }

  function openMessageContextMenu(event, bubble) {
    if (!messageContextMenu || !bubble) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();

    contextMessage = {
      messageID: Number(bubble.dataset.messageId || 0),
      senderName: bubble.dataset.senderName || "",
      excerpt: bubble.dataset.messageExcerpt || "",
      outgoing: bubble.dataset.outgoing === "true"
    };

    const replyButton = messageContextMenu.querySelector('[data-message-action="reply"]');
    if (replyButton) {
      replyButton.hidden = false;
    }
    const deleteButton = messageContextMenu.querySelector('[data-message-action="delete"]');
    if (deleteButton) {
      deleteButton.hidden = !contextMessage.outgoing;
    }
    positionMessageContextMenu(event);
  }

  function updateReplyPreview() {
    if (!replyPreview) {
      return;
    }
    if (!replyMessage) {
      replyPreview.hidden = true;
      return;
    }
    if (replySenderNode) {
      replySenderNode.textContent = replyMessage.senderName || "訊息";
    }
    if (replyExcerptNode) {
      replyExcerptNode.textContent = replyMessage.excerpt || "";
    }
    replyPreview.hidden = false;
  }

  function setReplyMessage(message) {
    replyMessage = message ? {
      senderName: message.senderName || "",
      excerpt: message.excerpt || ""
    } : null;
    updateReplyPreview();
  }

  async function recallContextMessage() {
    if (!contextMessage || !contextMessage.messageID || !activeConversationID) {
      return;
    }
    const headers = authHeaders();
    if (!headers) {
      setMessageStatus("請先登入。", true);
      return;
    }
    const messageID = contextMessage.messageID;
    closeMessageContextMenu();
    setMessageStatus("收回訊息中...", false);
    const response = await fetch("/api/conversations/" + activeConversationID + "/messages/" + messageID + "/recall", {
      method: "POST",
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "訊息收回失敗。", true);
      return;
    }
    await loadMessages(activeConversationID);
    await loadConversations();
    setMessageStatus("訊息已收回。", false);
  }

  function scrollMessageBoardToLatest() {
    window.requestAnimationFrame(function () {
      messageBoard.scrollTop = messageBoard.scrollHeight;
    });
  }

  function scrollMessageBoardToMessage(messageID) {
    if (!messageBoard || !messageID) {
      return false;
    }
    const target = messageBoard.querySelector('[data-message-id="' + CSS.escape(String(messageID)) + '"]');
    if (!target) {
      return false;
    }
    window.requestAnimationFrame(function () {
      const row = target.closest(".message-row") || target;
      messageBoard.scrollTop = Math.max(0, row.offsetTop - 16);
    });
    return true;
  }

  function scrollMessageBoardToUnreadDivider() {
    if (!messageBoard) {
      return false;
    }
    const target = messageBoard.querySelector("[data-unread-divider]");
    if (!target) {
      return false;
    }
    window.requestAnimationFrame(function () {
      messageBoard.scrollTop = Math.max(0, target.offsetTop - 12);
    });
    return true;
  }

  function displayMessageContent(message) {
    if (!message) {
      return "";
    }
    if (message.attachment && message.content === message.attachment.original_name) {
      return "";
    }
    return message.content || "";
  }

  function messageAttachments(message) {
    if (!message) {
      return [];
    }
    if (Array.isArray(message.attachments) && message.attachments.length) {
      return message.attachments;
    }
    return message.attachment ? [message.attachment] : [];
  }

  function renderAttachments(message, knownAttachments) {
    const attachments = knownAttachments || messageAttachments(message);
    if (!attachments.length) {
      return "";
    }

    if (allAttachmentsAreImages(attachments)) {
      return renderAttachmentPhotoMessage(attachments);
    }

    return '<div class="message-attachments">' + attachments.map(renderAttachment).join("") + "</div>";
  }

  function renderAttachmentPhotoMessage(attachments) {
    return [
      '<div class="message-photo-grid ' + photoGridClass(attachments.length) + '">',
      attachments.map(function (attachment) {
        return [
          '<a href="' + escapeHTML(attachment.url) + '" target="_blank" rel="noreferrer">',
          '<img src="' + escapeHTML(attachment.url) + '" alt="' + escapeHTML(attachment.original_name || "photo") + '">',
          '</a>'
        ].join("");
      }).join(""),
      '</div>'
    ].join("");
  }

  function renderAttachment(attachment) {
    if (!attachment || !attachment.url) {
      return "";
    }

    if ((attachment.mime_type || "").indexOf("image/") === 0) {
      return [
        '<a class="message-attachment image-attachment" href="' + escapeHTML(attachment.url) + '" target="_blank" rel="noreferrer">',
        '<img src="' + escapeHTML(attachment.url) + '" alt="' + escapeHTML(attachment.original_name || "image") + '">',
        "</a>"
      ].join("");
    }

    return [
      '<a class="message-attachment file-attachment" href="' + escapeHTML(attachment.url) + '" target="_blank" rel="noreferrer">',
      "<strong>" + escapeHTML(attachment.original_name || "附件") + "</strong>",
      "<small>" + escapeHTML(formatBytes(attachment.size_bytes)) + "</small>",
      "</a>"
    ].join("");
  }

  function updateSelectedFileState() {
    if (!fileStateNode || !composerFileInput) {
      return;
    }
    const files = composerFileInput.files ? Array.from(composerFileInput.files) : [];
    fileStateNode.textContent = "";
    if (files.length) {
      openAttachmentDialog(files, selectedAttachmentKind);
    } else {
      closeAttachmentDialog(false);
    }
  }

  function setAttachmentMenuOpen(open) {
    if (!attachmentMenu || !attachmentToggle) {
      return;
    }
    attachmentMenu.hidden = !open;
    attachmentToggle.setAttribute("aria-expanded", open ? "true" : "false");
  }

  function chooseAttachment(kind) {
    if (!composerFileInput) {
      return;
    }
    selectedAttachmentKind = kind === "photo" ? "photo" : "document";
    if (kind === "photo") {
      composerFileInput.accept = ".jpg,.jpeg,.png";
    } else {
      composerFileInput.accept = ".doc,.docx,.pdf,.xlsx";
    }
    composerFileInput.multiple = true;
    setAttachmentMenuOpen(false);
    composerFileInput.click();
  }

  function fileExtension(name) {
    const parts = String(name || "").split(".");
    return parts.length > 1 ? parts.pop().toLowerCase() : "file";
  }

  function isImageFile(file) {
    const name = String(file && file.name ? file.name : "").toLowerCase();
    return file && (file.type.indexOf("image/") === 0 || /\.(jpg|jpeg|png)$/.test(name));
  }

  function isImageAttachment(attachment) {
    const name = String(attachment && attachment.original_name ? attachment.original_name : "").toLowerCase();
    return !!attachment && ((attachment.mime_type || "").indexOf("image/") === 0 || /\.(jpg|jpeg|png)$/.test(name));
  }

  function allFilesAreImages(files) {
    return files.length > 0 && files.every(isImageFile);
  }

  function allAttachmentsAreImages(attachments) {
    return attachments.length > 0 && attachments.every(isImageAttachment);
  }

  function allowedAttachmentFiles(files) {
    const allowedAttachmentPattern = /\.(jpg|jpeg|png|doc|docx|pdf|xlsx)$/i;
    return files.filter(function (file) {
      return allowedAttachmentPattern.test(file.name || "");
    });
  }

  function revokeAttachmentPreviewURLs() {
    attachmentPreviewURLs.forEach(function (url) {
      URL.revokeObjectURL(url);
    });
    attachmentPreviewURLs = [];
  }

  function closeAttachmentDialog(clearFile) {
    if (attachmentDialog) {
      attachmentDialog.hidden = true;
    }
    revokeAttachmentPreviewURLs();
    if (attachmentPreview) {
      attachmentPreview.innerHTML = "";
    }
    if (attachmentCaption) {
      attachmentCaption.value = "";
      syncAttachmentCaptionHeight();
    }
    if (clearFile && composerFileInput) {
      composerFileInput.value = "";
    }
    if (clearFile) {
      selectedAttachmentFiles = [];
    }
  }

  function renderAttachmentFileList(files) {
    return [
      '<div class="attachment-file-list">',
      files.map(function (file) {
        let media = '<span class="attachment-file-ext">' + escapeHTML(fileExtension(file.name)) + '</span>';
        if (isImageFile(file)) {
          const url = URL.createObjectURL(file);
          attachmentPreviewURLs.push(url);
          media = '<img class="attachment-file-thumb" src="' + escapeHTML(url) + '" alt="' + escapeHTML(file.name) + '">';
        }
        return [
          '<div class="attachment-file-item">',
          media,
          '<span class="attachment-file-meta">',
          '<strong>' + escapeHTML(file.name || "附件") + '</strong>',
          '<small>' + escapeHTML(formatBytes(file.size)) + '</small>',
          '</span>',
          '</div>'
        ].join("");
      }).join(""),
      '</div>'
    ].join("");
  }

  function photoGridClass(count) {
    return "attachment-photo-grid count-" + Math.min(Math.max(count, 1), 5);
  }

  function renderAttachmentPhotoGrid(files) {
    return [
      '<div class="' + photoGridClass(files.length) + '">',
      files.map(function (file) {
        const url = URL.createObjectURL(file);
        attachmentPreviewURLs.push(url);
        return [
          '<img src="' + escapeHTML(url) + '" alt="' + escapeHTML(file.name || "photo") + '">'
        ].join("");
      }).join(""),
      '</div>'
    ].join("");
  }

  function openAttachmentDialog(files, kind) {
    if (!attachmentDialog || !attachmentPreview || !attachmentTitle) {
      return;
    }
    closeAttachmentDialog(false);
    selectedAttachmentKind = kind === "photo" ? "photo" : "document";
    selectedAttachmentFiles = Array.isArray(files) ? files : [files];
    const isSingleImage = selectedAttachmentFiles.length === 1 && isImageFile(selectedAttachmentFiles[0]);
    const isPhotoSet = selectedAttachmentFiles.length > 1 && allFilesAreImages(selectedAttachmentFiles);
    if (selectedAttachmentFiles.length > 1) {
      attachmentTitle.textContent = isPhotoSet
        ? "傳送 " + selectedAttachmentFiles.length + " 張照片"
        : "傳送 " + selectedAttachmentFiles.length + " 個檔案";
    } else {
      attachmentTitle.textContent = isSingleImage || selectedAttachmentKind === "photo" ? "傳送照片" : "傳送文件";
    }
    if (isSingleImage) {
      const file = selectedAttachmentFiles[0];
      const previewURL = URL.createObjectURL(file);
      attachmentPreviewURLs.push(previewURL);
      attachmentPreview.innerHTML = '<img src="' + escapeHTML(previewURL) + '" alt="' + escapeHTML(file.name) + '">';
    } else if (isPhotoSet) {
      attachmentPreview.innerHTML = renderAttachmentPhotoGrid(selectedAttachmentFiles);
    } else if (selectedAttachmentFiles.length > 1) {
      attachmentPreview.innerHTML = renderAttachmentFileList(selectedAttachmentFiles);
    } else {
      const file = selectedAttachmentFiles[0];
      attachmentPreview.innerHTML = [
        '<div class="attachment-document-preview">',
        '<span>' + escapeHTML(fileExtension(file.name)) + '</span>',
        '<strong>' + escapeHTML(file.name) + '</strong>',
        '<small>' + escapeHTML(formatBytes(file.size)) + '</small>',
        '</div>'
      ].join("");
    }
    attachmentDialog.hidden = false;
    if (attachmentCaption) {
      syncAttachmentCaptionHeight();
      attachmentCaption.focus();
    }
  }

  function syncAttachmentCaptionHeight() {
    if (!attachmentCaption) {
      return;
    }
    attachmentCaption.style.height = "44px";
    const maxHeight = 144;
    const nextHeight = Math.min(attachmentCaption.scrollHeight, maxHeight);
    attachmentCaption.style.height = nextHeight + "px";
    attachmentCaption.style.overflowY = attachmentCaption.scrollHeight > maxHeight ? "auto" : "hidden";
  }

  function sendAttachmentFromDialog() {
    if (attachmentCaption) {
      composerInput.value = attachmentCaption.value.trim();
    }
    messageForm.requestSubmit();
  }

  function openDraggedFiles(files, kind) {
    const acceptedFiles = allowedAttachmentFiles(Array.from(files || []));
    if (!acceptedFiles.length) {
      setMessageStatus("請拖放 jpg、png、doc、docx、pdf 或 xlsx 檔案。", true);
      return;
    }
    if (composerFileInput) {
      composerFileInput.value = "";
    }
    openAttachmentDialog(acceptedFiles, kind);
  }

  function dragEventHasFiles(event) {
    const types = event.dataTransfer && event.dataTransfer.types;
    return types && Array.from(types).indexOf("Files") !== -1;
  }

  function setDropOverlayVisible(visible) {
    if (!chatDropOverlay) {
      return;
    }
    chatDropOverlay.hidden = !visible;
    if (!visible) {
      chatDropZones.forEach(function (zone) {
        zone.classList.remove("is-drag-over");
      });
    }
  }

  function handleChatDragEnter(event) {
    if (!dragEventHasFiles(event) || attachmentDialog && !attachmentDialog.hidden) {
      return;
    }
    event.preventDefault();
    dragDepth += 1;
    setDropOverlayVisible(true);
  }

  function handleChatDragOver(event) {
    if (!dragEventHasFiles(event)) {
      return;
    }
    event.preventDefault();
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "copy";
    }
    setDropOverlayVisible(true);
  }

  function handleChatDragLeave(event) {
    if (!dragEventHasFiles(event)) {
      return;
    }
    dragDepth = Math.max(0, dragDepth - 1);
    if (!dragDepth) {
      setDropOverlayVisible(false);
    }
  }

  function handleDropZoneDragOver(event) {
    if (!dragEventHasFiles(event)) {
      return;
    }
    event.preventDefault();
    chatDropZones.forEach(function (zone) {
      zone.classList.toggle("is-drag-over", zone === event.currentTarget);
    });
  }

  function handleDropZoneDrop(event) {
    if (!dragEventHasFiles(event)) {
      return;
    }
    event.preventDefault();
    dragDepth = 0;
    setDropOverlayVisible(false);
    openDraggedFiles(event.dataTransfer.files, event.currentTarget.dataset.dropKind || "document");
  }

  function handleAttachmentCaptionKeydown(event) {
    if (event.key !== "Enter" || event.shiftKey || event.isComposing) {
      return;
    }

    event.preventDefault();
    sendAttachmentFromDialog();
  }

  async function loadConversations() {
    loadSessionScopedState();
    const items = await fetchConversationItems();
    if (!items) {
      return;
    }

    renderConversationList(items);
    updateContactActions();
    if (activeConversationID) {
      await loadMessages(activeConversationID);
      return;
    }

    if (!pendingDirectTarget) {
      renderEmptyConversationState();
    }
  }

  async function fetchConversationItems() {
    const headers = authHeaders();
    if (!headers) {
      activeConversationID = 0;
      activeFolderID = "";
      if (folderTabs) {
        folderTabs.hidden = true;
      }
      renderConversationList([]);
      conversationSummary.textContent = "請先登入，再載入對話列表。";
      renderEmptyConversationState("請選擇對話對象，開始傳訊息");
      return null;
    }

    if (folderTabs) {
      folderTabs.hidden = false;
    }

    const response = await fetch("/api/conversations", {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      renderConversationList([]);
      conversationSummary.textContent = result.message || "對話列表讀取失敗。";
      return null;
    }

    return result.data || [];
  }

  async function refreshConversationListOnly() {
    const items = await fetchConversationItems();
    if (!items) {
      return;
    }
    renderConversationList(items);
  }

  async function loadMessages(conversationID, refreshListAfterRead) {
    const headers = authHeaders();
    if (!headers || !conversationID) {
      return;
    }

    const requestToken = ++messageLoadToken;
    setConversationUIActive(true);
    setMessageStatus("正在載入訊息...", false);
    setConversationTitle(activeConversationTitle());
    renderLoadingConversationState();
    const response = await fetch("/api/conversations/" + conversationID + "/messages", {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (requestToken !== messageLoadToken || conversationID !== activeConversationID) {
      return;
    }
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "訊息讀取失敗。", true);
      return;
    }

    const data = result.data || {};
    setConversationTitle(data.title || activeConversationTitle() || "目前對話");
    renderMessages(data);
    setMessageStatus("已載入訊息。", false);
    if (refreshListAfterRead !== false) {
      await refreshConversationListOnly();
    }
  }

  function websocketURL() {
    const token = readSessionToken();
    if (!token) {
      return "";
    }
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return protocol + "//" + window.location.host + "/ws?token=" + encodeURIComponent(token);
  }

  function clearReconnectTimer() {
    if (!reconnectTimer) {
      return;
    }
    window.clearTimeout(reconnectTimer);
    reconnectTimer = 0;
  }

  function closeRealtimeSocket() {
    clearReconnectTimer();
    if (!realtimeSocket) {
      setRealtimeState("WS 未連線", false);
      return;
    }

    const socket = realtimeSocket;
    realtimeSocket = null;
    socket.onopen = null;
    socket.onmessage = null;
    socket.onerror = null;
    socket.onclose = null;
    socket.close();
    setRealtimeState("WS 未連線", false);
  }

  function scheduleRealtimeReconnect() {
    if (!readSessionToken() || reconnectTimer) {
      return;
    }

    const delay = Math.min(1000 * Math.max(reconnectAttempts, 1), 5000);
    reconnectTimer = window.setTimeout(function () {
      reconnectTimer = 0;
      connectRealtime();
    }, delay);
    setRealtimeState("WS 重連中", false);
  }

  function handleRealtimeEvent(event) {
    if (!event || !event.event_type) {
      return;
    }

    if (event.event_type === "conversation.ready") {
      loadConversations();
      return;
    }

    if (event.event_type === "message.created") {
      loadConversations();
      if (event.conversation_id && Number(event.conversation_id) === activeConversationID) {
        loadMessages(activeConversationID);
      }
      return;
    }

    if (event.event_type === "message.recalled" || event.event_type === "message.deleted") {
      loadConversations();
      if (event.conversation_id && Number(event.conversation_id) === activeConversationID) {
        loadMessages(activeConversationID);
      }
    }
  }

  function connectRealtime() {
    const url = websocketURL();
    if (!url) {
      closeRealtimeSocket();
      return;
    }
    if (realtimeSocket && (realtimeSocket.readyState === WebSocket.OPEN || realtimeSocket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    closeRealtimeSocket();
    setRealtimeState("WS 連線中", false);

    const socket = new WebSocket(url);
    realtimeSocket = socket;

    socket.onopen = function () {
      reconnectAttempts = 0;
      setRealtimeState("WS 已連線", true);
    };

    socket.onmessage = function (raw) {
      try {
        handleRealtimeEvent(JSON.parse(raw.data));
      } catch (_) {
      }
    };

    socket.onerror = function () {
      setRealtimeState("WS 連線異常", false);
    };

    socket.onclose = function () {
      if (realtimeSocket === socket) {
        realtimeSocket = null;
      }
      reconnectAttempts += 1;
      scheduleRealtimeReconnect();
    };
  }

  function openPendingDirectConversation(target) {
    const externalUserID = String(target.externalUserID || "").trim();
    if (!externalUserID) {
      return;
    }
    const sourceSystem = String(target.sourceSystem || readSourceSystem()).trim();
    const existing = findExistingDirectConversation(sourceSystem, externalUserID);
    if (existing) {
      openConversation(existing.conversation_id);
      return;
    }

    activeConversationID = 0;
    pendingDirectTarget = {
      sourceSystem: sourceSystem,
      externalUserID: externalUserID,
      title: target.title || externalUserID
    };
    const key = accountStorageKey(storageKeys.activeConversationID);
    if (key) {
      localStorage.removeItem(key);
    }
    closeSearchMode();
    renderConversationList(conversations);
    updateContactActions();
    messageLoadToken += 1;
    setConversationUIActive(true);
    setConversationTitle(activeConversationTitle());
    setMessageStatus("尚未建立對話，送出第一則訊息後會建立。", false);
    messageBoard.innerHTML = "";
    scrollMessageBoardToLatest();
    composerInput.focus();
  }

  async function ensureActiveConversationForSend(headers) {
    if (activeConversationID) {
      return activeConversationID;
    }
    if (!pendingDirectTarget) {
      return 0;
    }

    setMessageStatus("建立對話中...", false);
    const response = await fetch("/api/conversations/direct", {
      method: "POST",
      headers: headers,
      body: JSON.stringify({
        source_system: pendingDirectTarget.sourceSystem,
        external_user_id: pendingDirectTarget.externalUserID
      })
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "建立對話失敗", true);
      return 0;
    }

    activeConversationID = result.data.conversation_id;
    pendingDirectTarget = null;
    const key = accountStorageKey(storageKeys.activeConversationID);
    if (key) {
      localStorage.setItem(key, String(activeConversationID));
    }
    return activeConversationID;
  }

  async function sendMessage(event) {
    event.preventDefault();
    const headers = authHeaders();
    let content = composerInput.value.trim();
    const files = selectedAttachmentFiles.length
      ? selectedAttachmentFiles.slice()
      : (composerFileInput && composerFileInput.files ? Array.from(composerFileInput.files) : []);
    if (!headers) {
      setMessageStatus("請先登入。", true);
      return;
    }
    if (!activeConversationID && !pendingDirectTarget) {
      setMessageStatus("請先建立或選擇對話。", true);
      return;
    }
    if (!content && !files.length) {
      setMessageStatus("請輸入訊息內容或選擇檔案。", true);
      return;
    }
    if (replyMessage && content) {
      content = "回覆 " + (replyMessage.senderName || "訊息") + "：「" + (replyMessage.excerpt || "") + "」\n" + content;
    }

    setMessageStatus("訊息送出中...", false);
    const conversationID = await ensureActiveConversationForSend(headers);
    if (!conversationID) {
      return;
    }

    if (files.length) {
      const formData = new FormData();
      formData.set("content", content);
      files.forEach(function (file) {
        formData.append("file", file);
      });
      const response = await fetch("/api/conversations/" + conversationID + "/messages", {
        method: "POST",
        headers: { Authorization: headers.Authorization },
        body: formData
      });
      const result = await parseJSON(response);
      if (!response.ok || !result.success) {
        setMessageStatus(result.message || "檔案送出失敗，請確認格式或大小後重試。", true);
        return;
      }
    } else {
      const response = await fetch("/api/conversations/" + conversationID + "/messages", {
        method: "POST",
        headers: headers,
        body: JSON.stringify({ type: "text", content: content })
      });
      const result = await parseJSON(response);
      if (!response.ok || !result.success) {
        setMessageStatus(result.message || "送出失敗。", true);
        return;
      }
    }

    composerInput.value = "";
    syncComposerInputHeight();
    if (composerFileInput) {
      composerFileInput.value = "";
    }
    selectedAttachmentFiles = [];
    setReplyMessage(null);
    updateSelectedFileState();
    closeAttachmentDialog(false);
    await loadMessages(activeConversationID);
    await loadConversations();
    setMessageStatus("訊息已送出。", false);
  }

  function handleComposerKeydown(event) {
    if (event.key !== "Enter" || event.shiftKey || event.isComposing) {
      return;
    }

    event.preventDefault();
    if (!composerInput.value.trim()) {
      return;
    }
    messageForm.requestSubmit();
  }

  function syncComposerInputHeight() {
    if (!composerInput) {
      return;
    }
    composerInput.style.height = "48px";
    const maxHeight = 132;
    const nextHeight = Math.min(composerInput.scrollHeight, maxHeight);
    composerInput.style.height = nextHeight + "px";
    composerInput.style.overflowY = composerInput.scrollHeight > maxHeight ? "auto" : "hidden";
  }

  document.addEventListener("twacc:session-changed", function () {
    contactKeys = new Set();
    contactExternalIDs = new Set();
    contactsLoaded = false;
    contactsIndexLoading = false;
    connectRealtime();
    ensureContactsLoaded();
    loadConversations();
  });

  document.addEventListener("twacc:conversation-filter-changed", function (event) {
    const detail = event.detail || {};
    activeFolderID = detail.folderID || "";
    const key = accountStorageKey(storageKeys.activeFolderID);
    if (activeFolderID) {
      if (key) {
        localStorage.setItem(key, activeFolderID);
      }
    } else if (key) {
      localStorage.removeItem(key);
    }
    renderConversationList(conversations);
    if (!activeConversationID) {
      return;
    }
    if (!detail.folderID || !Array.isArray(detail.conversationIDs) || detail.conversationIDs.indexOf(String(activeConversationID)) < 0) {
      closeActiveConversation();
    }
  });

  if (conversationSearchInput) {
    conversationSearchInput.addEventListener("focus", function () {
      conversationSearchQuery = conversationSearchInput.value.trim();
      renderSearchMode();
    });
    conversationSearchInput.addEventListener("input", function () {
      conversationSearchQuery = conversationSearchInput.value.trim();
      renderSearchMode();
    });
  }

  if (searchAllMessagesButton && conversationSearchInput) {
    searchAllMessagesButton.addEventListener("click", function () {
      conversationSearchInput.focus();
    });
  }

  if (conversationCreateToggle) {
    conversationCreateToggle.addEventListener("click", function (event) {
      event.stopPropagation();
      setConversationCreateMenuOpen(conversationCreateMenu ? conversationCreateMenu.hidden : true);
    });
  }

  if (conversationCreateMenu) {
    conversationCreateMenu.addEventListener("click", function (event) {
      const actionButton = event.target.closest("[data-conversation-create-action]");
      if (!actionButton) {
        return;
      }
      const action = actionButton.dataset.conversationCreateAction;
      setConversationCreateMenuOpen(false);
      if (action === "group") {
        resetNewGroupDraft();
        document.dispatchEvent(new CustomEvent("twacc:open-new-group"));
      }
    });
  }

  if (newGroupSearchInput) {
    newGroupSearchInput.addEventListener("input", function () {
      newGroupSearchQuery = newGroupSearchInput.value.trim();
      searchNewGroupUsers();
    });
  }

  if (newGroupResultsNode) {
    newGroupResultsNode.addEventListener("click", function (event) {
      const memberButton = event.target.closest("[data-new-group-id]");
      if (!memberButton) {
        return;
      }
      const member = {
        sourceSystem: memberButton.dataset.newGroupSource || readSourceSystem(),
        externalUserID: memberButton.dataset.newGroupId || "",
        title: memberButton.dataset.newGroupTitle || ""
      };
      updateNewGroupSelectedMember(member, !isNewGroupMemberSelected(member));
      searchNewGroupUsers();
    });
  }

  if (newGroupSelectedNode) {
    newGroupSelectedNode.addEventListener("click", function (event) {
      const removeButton = event.target.closest("button");
      const chip = event.target.closest("[data-new-group-selected-key]");
      if (!removeButton || !chip) {
        return;
      }
      newGroupSelectedMembers = newGroupSelectedMembers.filter(function (member) {
        return memberKey(member) !== chip.dataset.newGroupSelectedKey;
      });
      renderNewGroupSelectedMembers();
      searchNewGroupUsers();
    });
  }

  if (newGroupNextButton) {
    newGroupNextButton.addEventListener("click", function () {
      if (!newGroupSelectedMembers.length) {
        return;
      }
      renderNewGroupDetails();
      if (newGroupNameInput && !newGroupNameInput.value.trim()) {
        newGroupNameInput.value = suggestedNewGroupName();
      }
      document.dispatchEvent(new CustomEvent("twacc:open-new-group-details"));
    });
  }

	  if (newGroupCreateButton) {
	    newGroupCreateButton.addEventListener("click", createGroupConversation);
	  }

  if (contactMenuToggle) {
    contactMenuToggle.addEventListener("click", async function (event) {
      event.stopPropagation();
      await ensureContactsLoaded();
      if (contactMenuToggle.hidden) {
        setContactMenuOpen(false);
        return;
      }
      setContactMenuOpen(contactMenu ? contactMenu.hidden : true);
    });
  }

  if (contactMenu) {
    contactMenu.addEventListener("click", function (event) {
      const actionButton = event.target.closest("[data-contact-action]");
      if (!actionButton) {
        return;
      }
      const peer = activeDirectPeer();
      setContactMenuOpen(false);
      if (actionButton.dataset.contactAction === "add" && peer) {
        fillContactPanel(peer);
        setContactPanelOpen(true);
        if (contactAliasName) {
          window.requestAnimationFrame(function () {
            contactAliasName.focus();
            contactAliasName.select();
          });
        }
      }
    });
  }

  if (contactPanelClose) {
    contactPanelClose.addEventListener("click", function () {
      setContactPanelOpen(false);
    });
  }

  if (contactForm) {
    contactForm.addEventListener("submit", function (event) {
      event.preventDefault();
      addActivePeerToContacts();
    });
  }

	  document.addEventListener("twacc:sidebar-view-changed", function (event) {
    const detail = event.detail || {};
    if (detail.view === "contacts") {
      loadContacts();
    }
    if (detail.view === "new-group" && newGroupSearchInput) {
      window.requestAnimationFrame(function () {
        newGroupSearchInput.focus();
      });
    }
    if (detail.view === "new-group-details" && newGroupNameInput) {
      renderNewGroupDetails();
      window.requestAnimationFrame(function () {
        newGroupNameInput.focus();
        newGroupNameInput.select();
      });
    }
  });

  document.addEventListener("twacc:search-close", closeSearchMode);

  window.addEventListener("storage", function (event) {
    if (!event.key || event.key.indexOf("twacc_chat_") !== 0) {
      return;
    }
    const activeFolderKey = accountStorageKey(storageKeys.activeFolderID);
    if (event.key === activeFolderKey) {
      activeFolderID = activeFolderKey ? localStorage.getItem(activeFolderKey) || "" : "";
    }
    loadConversations();
  });

  messageForm.addEventListener("submit", sendMessage);
  composerInput.addEventListener("input", syncComposerInputHeight);
  composerInput.addEventListener("keydown", handleComposerKeydown);
  messageBoard.addEventListener("contextmenu", function (event) {
    const bubble = bubbleFromContextEvent(event);
    if (!bubble) {
      return;
    }
    openMessageContextMenu(event, bubble);
  }, true);
  messageBoard.addEventListener("scroll", closeMessageContextMenu, { passive: true });
  messageBoard.addEventListener("wheel", closeMessageContextMenu, { passive: true });
  if (messageContextMenu) {
    messageContextMenu.addEventListener("click", function (event) {
      const actionButton = event.target.closest("[data-message-action]");
      if (!actionButton || !contextMessage) {
        return;
      }
      const action = actionButton.dataset.messageAction;
      if (action === "reply") {
        setReplyMessage(contextMessage);
        closeMessageContextMenu();
        composerInput.focus();
        return;
      }
      if (action === "delete") {
        recallContextMessage();
      }
    });
  }
  if (replyCancelButton) {
    replyCancelButton.addEventListener("click", function () {
      setReplyMessage(null);
      composerInput.focus();
    });
  }
	  document.addEventListener("click", function (event) {
	    if (conversationCreateActions && !conversationCreateActions.contains(event.target)) {
	      setConversationCreateMenuOpen(false);
	    }
    if (contactMenu && contactMenuToggle && !contactMenu.hidden && !contactMenu.contains(event.target) && !contactMenuToggle.contains(event.target)) {
      setContactMenuOpen(false);
    }
	    if (!messageContextMenu || messageContextMenu.hidden || messageContextMenu.contains(event.target)) {
	      return;
	    }
    closeMessageContextMenu();
  });
  syncComposerInputHeight();
	  document.addEventListener("keydown", function (event) {
	    if (event.key === "Escape") {
	      closeMessageContextMenu();
	      setConversationCreateMenuOpen(false);
      setContactMenuOpen(false);
      setContactPanelOpen(false);
	    }
    if (event.key !== "Escape" || !activeConversationID) {
      return;
    }
    event.preventDefault();
    closeActiveConversation();
  });
  if (composerFileInput) {
    composerFileInput.addEventListener("change", updateSelectedFileState);
    updateSelectedFileState();
  }
  if (attachmentToggle && attachmentMenu) {
    attachmentToggle.addEventListener("click", function () {
      setAttachmentMenuOpen(attachmentMenu.hidden);
    });
    attachmentMenu.querySelectorAll("[data-attachment-kind]").forEach(function (button) {
      button.addEventListener("click", function () {
        chooseAttachment(button.dataset.attachmentKind || "document");
      });
    });
    document.addEventListener("click", function (event) {
      if (attachmentMenu.hidden || messageForm.contains(event.target)) {
        return;
      }
      setAttachmentMenuOpen(false);
    });
    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        setAttachmentMenuOpen(false);
        closeAttachmentDialog(true);
      }
    });
  }
  if (attachmentClose) {
    attachmentClose.addEventListener("click", function () {
      closeAttachmentDialog(true);
    });
  }
  if (attachmentDialog) {
    attachmentDialog.addEventListener("click", function (event) {
      if (event.target === attachmentDialog) {
        closeAttachmentDialog(true);
      }
    });
  }
  if (attachmentSend) {
    attachmentSend.addEventListener("click", sendAttachmentFromDialog);
  }
  if (attachmentCaption) {
    attachmentCaption.addEventListener("input", syncAttachmentCaptionHeight);
    attachmentCaption.addEventListener("keydown", handleAttachmentCaptionKeydown);
    syncAttachmentCaptionHeight();
  }
  if (chatShell && chatDropOverlay) {
    chatShell.addEventListener("dragenter", handleChatDragEnter);
    chatShell.addEventListener("dragover", handleChatDragOver);
    chatShell.addEventListener("dragleave", handleChatDragLeave);
    chatShell.addEventListener("drop", function (event) {
      if (!dragEventHasFiles(event)) {
        return;
      }
      event.preventDefault();
      dragDepth = 0;
      setDropOverlayVisible(false);
    });
    chatDropZones.forEach(function (zone) {
      zone.addEventListener("dragover", handleDropZoneDragOver);
      zone.addEventListener("drop", handleDropZoneDrop);
    });
  }

  connectRealtime();
  loadConversations();
})();
