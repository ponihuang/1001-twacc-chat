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
  const directForm = app.querySelector("[data-direct-form]");
  const directStatus = app.querySelector("[data-direct-status]");
  const directExternalUserIDInput = app.querySelector("#direct-external-user-id");
  const messageBoard = app.querySelector("[data-message-board]");
  const messageForm = app.querySelector("[data-message-form]");
  const conversationTitle = app.querySelector("[data-conversation-title]");
  const messageStatus = app.querySelector("[data-message-status]");
  const realtimeStateNode = app.querySelector("[data-realtime-state]");
  const composerInput = messageForm.querySelector(".composer-input");
  const composerFileInput = messageForm.querySelector(".composer-file-input");
  const fileStateNode = messageForm.querySelector("[data-file-state]");

  let conversations = [];
  let activeConversationID = 0;
  let activeFolderID = localStorage.getItem(storageKeys.activeFolderID) || "";
  let realtimeSocket = null;
  let reconnectTimer = 0;
  let reconnectAttempts = 0;
  let messageLoadToken = 0;

  function readSessionToken() {
    return localStorage.getItem(storageKeys.token) || "";
  }

  function readSourceSystem() {
    return localStorage.getItem(storageKeys.sourceSystem) || "erp";
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
    const active = conversations.find(function (item) {
      return item.conversation_id === activeConversationID;
    });
    return active ? active.title : "";
  }

  function readFolderCategories() {
    try {
      const raw = localStorage.getItem(storageKeys.folderCategories);
      return raw ? JSON.parse(raw) : [];
    } catch (_) {
      return [];
    }
  }

  function filteredConversations(items) {
    if (!activeFolderID) {
      return items;
    }
    const folder = readFolderCategories().find(function (item) {
      return String(item.id || "") === activeFolderID;
    });
    if (!folder || !Array.isArray(folder.conversation_ids)) {
      return items;
    }
    const allowed = new Set(folder.conversation_ids.map(String));
    return items.filter(function (item) {
      return allowed.has(String(item.conversation_id));
    });
  }

  function syncConversationCatalog(items) {
    localStorage.setItem(storageKeys.conversationCatalog, JSON.stringify(items.map(function (item) {
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
    setConversationUIActive(false);
    setConversationTitle(stateTitle);
    setMessageStatus(stateDetail, false);
    messageBoard.innerHTML = [
      '<div class="empty-chat-state">',
      "<strong>" + escapeHTML(stateTitle) + "</strong>",
      "</div>"
    ].join("");
  }

  function renderLoadingConversationState() {
    messageBoard.innerHTML = [
      '<div class="empty-chat-state">',
      "<strong>載入中...</strong>",
      "</div>"
    ].join("");
  }

  function closeActiveConversation() {
    if (!activeConversationID) {
      return;
    }
    activeConversationID = 0;
    localStorage.removeItem(storageKeys.activeConversationID);
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
      conversationList.innerHTML = [
        '<article class="conversation-card is-placeholder">',
        "<strong>" + (items.length ? "此分類尚無對話" : "目前沒有對話") + "</strong>",
        "<span>" + (items.length ? "請到聊天室分類加入對話" : "先建立一個一對一對話") + "</span>",
        "<small>" + (items.length ? "等待分類" : "等待建立") + "</small>",
        "</article>"
      ].join("");
      conversationSummary.textContent = items.length ? "此聊天室分類目前沒有對應對話。" : "目前尚無對話，請先建立一個一對一對話。";
      renderEmptyConversationState("請選擇對話對象，開始傳訊息");
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
        activeConversationID = Number(button.dataset.conversationId || 0);
        if (!activeConversationID) {
          return;
        }
        localStorage.setItem(storageKeys.activeConversationID, String(activeConversationID));
        renderConversationList(conversations);
        loadMessages(activeConversationID);
      });
    });
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
      const attachment = renderAttachment(message.attachment);
      const contentText = displayMessageContent(message);
      const content = contentText
        ? ("<p>" + escapeHTML(contentText) + "</p>")
        : "";
      return [
        '<div class="message-row' + outgoing + '">',
        '<div class="message-bubble">',
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
    const file = composerFileInput.files && composerFileInput.files[0];
    fileStateNode.textContent = file ? ("已選擇：" + file.name + " (" + formatBytes(file.size) + ")") : "尚未選擇檔案";
  }

  async function loadConversations() {
    const items = await fetchConversationItems();
    if (!items) {
      return;
    }

    renderConversationList(items);
    if (activeConversationID) {
      await loadMessages(activeConversationID);
      return;
    }

    renderEmptyConversationState();
  }

  async function fetchConversationItems() {
    const headers = authHeaders();
    if (!headers) {
      renderConversationList([]);
      conversationSummary.textContent = "請先登入，再載入對話列表。";
      renderEmptyConversationState("請選擇對話對象，開始傳訊息");
      return null;
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
      if (event.conversation_id) {
        activeConversationID = Number(event.conversation_id);
        localStorage.setItem(storageKeys.activeConversationID, String(activeConversationID));
        loadMessages(activeConversationID);
      }
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

  async function createDirectConversation(event) {
    event.preventDefault();
    const headers = authHeaders();
    if (!headers) {
      directStatus.textContent = "請先登入。";
      directStatus.classList.add("is-error");
      return;
    }

    directStatus.textContent = "建立對話中...";
    directStatus.classList.remove("is-error");
    directStatus.classList.remove("is-success");

    const response = await fetch("/api/conversations/direct", {
      method: "POST",
      headers: headers,
      body: JSON.stringify({
        source_system: readSourceSystem(),
        external_user_id: directExternalUserIDInput.value.trim()
      })
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      directStatus.textContent = result.message || "建立對話失敗";
      directStatus.classList.add("is-error");
      return;
    }

    activeConversationID = result.data.conversation_id;
    localStorage.setItem(storageKeys.activeConversationID, String(activeConversationID));
    directStatus.textContent = "對話已準備完成：" + (result.data.title || directExternalUserIDInput.value.trim());
    directStatus.classList.add("is-success");
    await loadConversations();
  }

  async function sendMessage(event) {
    event.preventDefault();
    const headers = authHeaders();
    const content = composerInput.value.trim();
    const file = composerFileInput && composerFileInput.files ? composerFileInput.files[0] : null;
    if (!headers) {
      setMessageStatus("請先登入。", true);
      return;
    }
    if (!activeConversationID) {
      setMessageStatus("請先建立或選擇對話。", true);
      return;
    }
    if (!content && !file) {
      setMessageStatus("請輸入訊息內容或選擇檔案。", true);
      return;
    }

    setMessageStatus("訊息送出中...", false);
    let response;
    if (file) {
      const formData = new FormData();
      formData.set("content", content);
      formData.set("file", file);
      response = await fetch("/api/conversations/" + activeConversationID + "/messages", {
        method: "POST",
        headers: { Authorization: headers.Authorization },
        body: formData
      });
    } else {
      response = await fetch("/api/conversations/" + activeConversationID + "/messages", {
        method: "POST",
        headers: headers,
        body: JSON.stringify({ type: "text", content: content })
      });
    }
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "送出失敗。", true);
      return;
    }

    composerInput.value = "";
    if (composerFileInput) {
      composerFileInput.value = "";
    }
    updateSelectedFileState();
    await loadMessages(activeConversationID);
    await loadConversations();
    setMessageStatus("訊息已送出。", false);
  }

  function handleComposerKeydown(event) {
    if (event.key !== "Enter" || event.shiftKey || event.isComposing) {
      return;
    }

    event.preventDefault();
    sendMessage(event);
  }

  document.addEventListener("twacc:session-changed", function () {
    connectRealtime();
    loadConversations();
  });

  document.addEventListener("twacc:conversation-filter-changed", function (event) {
    const detail = event.detail || {};
    activeFolderID = detail.folderID || "";
    if (activeFolderID) {
      localStorage.setItem(storageKeys.activeFolderID, activeFolderID);
    } else {
      localStorage.removeItem(storageKeys.activeFolderID);
    }
    renderConversationList(conversations);
    if (!activeConversationID) {
      return;
    }
    if (!detail.folderID || !Array.isArray(detail.conversationIDs) || detail.conversationIDs.indexOf(String(activeConversationID)) < 0) {
      closeActiveConversation();
    }
  });

  window.addEventListener("storage", function (event) {
    if (!event.key || event.key.indexOf("twacc_chat_") !== 0) {
      return;
    }
    if (event.key === storageKeys.activeFolderID) {
      activeFolderID = localStorage.getItem(storageKeys.activeFolderID) || "";
    }
    loadConversations();
  });

  if (directForm) {
    directForm.addEventListener("submit", createDirectConversation);
  }
  messageForm.addEventListener("submit", sendMessage);
  composerInput.addEventListener("keydown", handleComposerKeydown);
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

  connectRealtime();
  loadConversations();
})();
