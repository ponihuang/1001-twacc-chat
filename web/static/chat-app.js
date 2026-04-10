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

  let conversations = [];
  let activeConversationID = Number(localStorage.getItem(storageKeys.activeConversationID) || 0);

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

  loadConversations();
})();
