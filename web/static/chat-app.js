(function () {
  const app = document.querySelector("[data-auth-app]");
  if (!app) {
    return;
  }

  const storageKeys = {
    token: "twacc_chat_session_token",
    sourceSystem: "twacc_chat_source_system",
    activeConversationID: "twacc_chat_active_conversation_id"
  };

  const conversationList = app.querySelector("[data-conversation-list]");
  const conversationSummary = app.querySelector("[data-conversation-summary]");
  const directForm = app.querySelector("[data-direct-form]");
  const directStatus = app.querySelector("[data-direct-status]");
  const directExternalUserIDInput = app.querySelector("#direct-external-user-id");
  const messageBoard = app.querySelector("[data-message-board]");
  const realtimeStateNode = app.querySelector("[data-realtime-state]");

  let conversations = [];
  let activeConversationID = Number(localStorage.getItem(storageKeys.activeConversationID) || 0);
  let realtimeSocket = null;
  let reconnectTimer = 0;
  let reconnectAttempts = 0;

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

  function setRealtimeState(text, isConnected) {
    if (!realtimeStateNode) {
      return;
    }
    realtimeStateNode.textContent = text;
    realtimeStateNode.classList.toggle("is-live", Boolean(isConnected));
  }

  function renderConversationList(items) {
    conversations = items;
    if (!items.length) {
      conversationList.innerHTML = [
        '<article class="conversation-card is-placeholder">',
        "<strong>目前沒有對話</strong>",
        "<span>先建立一個一對一對話</span>",
        "<small>等待建立</small>",
        "</article>"
      ].join("");
      conversationSummary.textContent = "目前尚無對話，請先建立一個一對一對話。";
      return;
    }

    conversationList.innerHTML = items.map(function (item) {
      const active = item.conversation_id === activeConversationID ? " is-active" : "";
      const preview = item.last_message_preview || (item.type === "direct" ? "一對一對話" : "群組對話");
      const meta = formatTime(item.last_message_at) || (item.member_count + " 位成員");
      return [
        '<button type="button" class="conversation-card conversation-button' + active + '" data-conversation-id="' + item.conversation_id + '">',
        "<strong>" + escapeHTML(item.title) + "</strong>",
        "<span>" + escapeHTML(preview) + "</span>",
        "<small>" + escapeHTML(meta) + "</small>",
        "</button>"
      ].join("");
    }).join("");

    conversationSummary.textContent = "已載入 " + items.length + " 筆對話。";

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

  function renderMessages(data) {
    if (!data.messages || !data.messages.length) {
      messageBoard.innerHTML = [
        '<div class="message-row incoming">',
        '<div class="message-bubble">',
        "<strong>" + escapeHTML(data.title || "系統提示") + "</strong>",
        "<p>目前還沒有訊息，之後可在這裡開始聊天。</p>",
        "</div>",
        "</div>"
      ].join("");
      return;
    }

    messageBoard.innerHTML = data.messages.map(function (message) {
      return [
        '<div class="message-row incoming">',
        '<div class="message-bubble">',
        "<strong>" + escapeHTML(message.sender_name) + "</strong>",
        "<p>" + escapeHTML(message.content) + "</p>",
        "<small>" + escapeHTML(formatTime(message.created_at)) + "</small>",
        "</div>",
        "</div>"
      ].join("");
    }).join("");
  }

  async function loadConversations() {
    const headers = authHeaders();
    if (!headers) {
      renderConversationList([]);
      conversationSummary.textContent = "請先登入，再載入對話列表。";
      return;
    }

    const response = await fetch("/api/conversations", {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      renderConversationList([]);
      conversationSummary.textContent = result.message || "對話列表讀取失敗。";
      return;
    }

    renderConversationList(result.data || []);
    if (!activeConversationID && result.data && result.data.length) {
      activeConversationID = result.data[0].conversation_id;
      localStorage.setItem(storageKeys.activeConversationID, String(activeConversationID));
    }

    if (activeConversationID) {
      renderConversationList(result.data || []);
      await loadMessages(activeConversationID);
    }
  }

  async function loadMessages(conversationID) {
    const headers = authHeaders();
    if (!headers || !conversationID) {
      return;
    }

    const response = await fetch("/api/conversations/" + conversationID + "/messages", {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      messageBoard.innerHTML = [
        '<div class="message-row incoming">',
        '<div class="message-bubble">',
        "<strong>系統提示</strong>",
        "<p>" + escapeHTML(result.message || "訊息讀取失敗。") + "</p>",
        "</div>",
        "</div>"
      ].join("");
      return;
    }

    renderMessages(result.data || {});
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

  document.addEventListener("twacc:session-changed", function () {
    connectRealtime();
    loadConversations();
  });

  window.addEventListener("storage", function (event) {
    if (!event.key || event.key.indexOf("twacc_chat_") !== 0) {
      return;
    }
    loadConversations();
  });

  if (directForm) {
    directForm.addEventListener("submit", createDirectConversation);
  }

  connectRealtime();
  loadConversations();
})();
