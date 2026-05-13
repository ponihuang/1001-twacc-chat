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
  const folderTabs = app.querySelector("[data-folder-tabs]");
  const messageBoard = app.querySelector("[data-message-board]");
  const messageForm = app.querySelector("[data-message-form]");
  const chatShell = app.querySelector(".chat-shell");
  const conversationTitle = app.querySelector("[data-conversation-title]");
  const messageStatus = app.querySelector("[data-message-status]");
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

  let conversations = [];
  let activeConversationID = 0;
  let activeFolderID = "";
  let conversationSearchQuery = "";
  let pendingDirectTarget = null;
  let userSearchRequestToken = 0;
  let messageSearchRequestToken = 0;
  let realtimeSocket = null;
  let reconnectTimer = 0;
  let reconnectAttempts = 0;
  let messageLoadToken = 0;
  let selectedAttachmentKind = "document";
  let selectedAttachmentFiles = [];
  let attachmentPreviewURLs = [];
  let dragDepth = 0;

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
    setSearchMode(false);
    renderDirectSearchResults([], false);
    renderMessageSearchResults([], false);
  }

  function conversationPreview(item) {
    if (item.last_message_type === "image") {
      return "圖片";
    }
    if (item.last_message_type === "file") {
      return "附件";
    }
    return item.last_message_preview || (item.type === "direct" ? "一對一對話" : "群組對話");
  }

  function conversationPreviewExcerpt(item) {
    const preview = conversationPreview(item);
    const limit = 15;
    if (preview.length <= limit) {
      return preview;
    }
    return preview.slice(0, limit) + "...";
  }

  function renderMessages(data) {
    if (!data.messages || !data.messages.length) {
      messageBoard.innerHTML = "";
      scrollMessageBoardToLatest();
      return;
    }

    messageBoard.innerHTML = data.messages.map(function (message) {
      const outgoing = (app.dataset.actor || "").indexOf(message.sender_name) >= 0 ? " outgoing" : "";
      const attachments = messageAttachments(message);
      const photoMessage = allAttachmentsAreImages(attachments);
      const attachment = renderAttachments(message, attachments);
      const contentText = displayMessageContent(message);
      const content = contentText
        ? ("<p>" + linkifyMessageText(contentText) + "</p>")
        : "";
      return [
        '<div class="message-row' + outgoing + '">',
        '<div class="message-bubble' + (photoMessage ? " photo-bubble" : "") + '">',
        "<strong>" + escapeHTML(message.sender_name) + "</strong>",
        content,
        attachment,
        "<small>" + escapeHTML(formatTime(message.created_at)) + "</small>",
        "</div>",
        "</div>"
      ].join("");
    }).join("");
    scrollMessageBoardToLatest();
  }

  function scrollMessageBoardToLatest() {
    window.requestAnimationFrame(function () {
      messageBoard.scrollTop = messageBoard.scrollHeight;
    });
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
    const content = composerInput.value.trim();
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
    connectRealtime();
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
  syncComposerInputHeight();
  document.addEventListener("keydown", function (event) {
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
