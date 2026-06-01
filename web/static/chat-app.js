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
  const groupInfoOpenButton = app.querySelector("[data-group-info-open]");
  const groupInfoPanel = app.querySelector("[data-group-info-panel]");
  const groupInfoClose = app.querySelector("[data-group-info-close]");
  const groupInfoHeading = app.querySelector("[data-group-info-heading]");
  const groupInfoEditButton = app.querySelector("[data-group-info-edit]");
  const groupInfoBody = app.querySelector("[data-group-info-body]");
  const groupInfoAvatar = app.querySelector("[data-group-info-avatar]");
  const groupInfoTitle = app.querySelector("[data-group-info-title]");
  const groupInfoCount = app.querySelector("[data-group-info-count]");
  const groupInfoDescription = app.querySelector("[data-group-info-description]");
  const groupInfoDescriptionText = app.querySelector("[data-group-info-description-text]");
  const groupInfoMembers = app.querySelector("[data-group-info-members]");
  const groupInfoAddButton = app.querySelector("[data-group-info-add]");
  const groupAddMembers = app.querySelector("[data-group-add-members]");
  const groupAddSelectedNode = app.querySelector("[data-group-add-selected]");
  const groupAddSearchInput = app.querySelector("[data-group-add-search]");
  const groupAddResultsNode = app.querySelector("[data-group-add-results]");
  const groupAddSubmitButton = app.querySelector("[data-group-add-submit]");
  const groupEditPanel = app.querySelector("[data-group-edit-panel]");
  const groupEditAvatar = app.querySelector("[data-group-edit-avatar]");
  const groupEditNameInput = app.querySelector("[data-group-edit-name]");
  const groupEditDescriptionInput = app.querySelector("[data-group-edit-description]");
  const groupEditSubmitButton = app.querySelector("[data-group-edit-submit]");
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
  const mobileChatBackButton = app.querySelector("[data-mobile-chat-back]");
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
  const photoEditor = app.querySelector("[data-photo-editor]");
  const photoEditorCanvas = app.querySelector("[data-photo-editor-canvas]");
  const photoEditorClose = app.querySelector("[data-photo-editor-close]");
  const photoEditorApply = app.querySelector("[data-photo-editor-apply]");
  const photoEditorReset = app.querySelector("[data-photo-editor-reset]");
  const photoEditorUndo = app.querySelector("[data-photo-editor-undo]");
  const photoEditorRedo = app.querySelector("[data-photo-editor-redo]");
  const photoEditorConfirm = app.querySelector("[data-photo-editor-confirm]");
  const photoEditorConfirmCancel = app.querySelector("[data-photo-editor-confirm-cancel]");
  const photoEditorConfirmDiscard = app.querySelector("[data-photo-editor-confirm-discard]");
  const photoEditorModeButtons = app.querySelectorAll("[data-photo-editor-mode]");
  const photoEditorPanels = app.querySelectorAll("[data-photo-editor-panel]");
  const photoRatioButtons = app.querySelectorAll("[data-photo-ratio]");
  const photoColorButtons = app.querySelectorAll("[data-photo-color]");
  const photoCustomColorButton = app.querySelector("[data-photo-color-custom]");
  const photoCustomColorPanel = app.querySelector("[data-photo-custom-color]");
  const photoColorHueInput = app.querySelector("[data-photo-color-hue]");
  const photoColorSV = app.querySelector("[data-photo-color-sv]");
  const photoColorSVPointer = app.querySelector("[data-photo-color-sv-pointer]");
  const photoColorHexInput = app.querySelector("[data-photo-color-hex]");
  const photoColorRGBInput = app.querySelector("[data-photo-color-rgb]");
  const photoToolButtons = app.querySelectorAll("[data-photo-tool]");
  const photoBrushSizeInput = app.querySelector("[data-photo-brush-size]");
  const chatDropOverlay = app.querySelector("[data-chat-drop-overlay]");
  const chatDropZones = app.querySelectorAll("[data-drop-kind]");
  const messageContextMenu = app.querySelector("[data-message-context-menu]");
  const replyPreview = app.querySelector("[data-reply-preview]");
  const replySenderNode = app.querySelector("[data-reply-sender]");
  const replyExcerptNode = app.querySelector("[data-reply-excerpt]");
  const replyCancelButton = app.querySelector("[data-reply-cancel]");
  const groupMemberMenu = document.createElement("div");
  groupMemberMenu.className = "group-member-menu";
  groupMemberMenu.hidden = true;
  groupMemberMenu.innerHTML = [
    '<button type="button" data-group-member-action="message">發送訊息</button>',
    '<button type="button" data-group-member-action="remove">從群組中移除</button>'
  ].join("");
  document.body.appendChild(groupMemberMenu);
  const composerMentionMenu = document.createElement("div");
  composerMentionMenu.className = "composer-mention-menu";
  composerMentionMenu.hidden = true;
  composerMentionMenu.setAttribute("role", "listbox");
  messageForm.querySelector(".composer-fields").appendChild(composerMentionMenu);

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
  let groupInfoLoadToken = 0;
  let selectedAttachmentKind = "document";
  let selectedAttachmentFiles = [];
  let attachmentPreviewURLs = [];
  let photoEditorState = null;
  let dragDepth = 0;
  let contextMessage = null;
  let replyMessage = null;
  let newGroupSearchQuery = "";
  let newGroupSelectedMembers = [];
  let groupInfoMemberKeys = new Set();
  let groupInfoActorRole = "";
  let groupInfoMode = "info";
  let groupInfoData = { title: "", description: "", memberCount: 0 };
  let groupEditOriginal = { name: "", description: "" };
  let groupAddSearchQuery = "";
  let groupAddSelectedMembers = [];
  let groupMemberMenuTarget = null;
  let contactsLoadToken = 0;
  let contactKeys = new Set();
  let contactExternalIDs = new Set();
  let contactItems = [];
  let contactsLoaded = false;
  let contactsIndexLoading = false;
  let mentionMembersCache = new Map();
  let mentionMenuItems = [];
  let mentionMenuActiveIndex = 0;
  let mentionQueryRange = null;
  let mentionLoadToken = 0;
  let readUpdateTimer = 0;
  let readUpdateInFlight = false;
  let pendingReadMessageID = 0;
  let lastReportedReadMessageID = 0;
  const photoToolDefaultColors = {
    pen: "#ff8a0a",
    arrow: "#ffd21f",
    marker: "#ff4b40",
    blur: "#38d8d3",
    eraser: "#ff9aa6"
  };

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
    }).replace(/(^|\s)(@(ALL|[^\s@<]+))/g, function (_, prefix, mention) {
      return prefix + '<span class="message-mention">' + mention + '</span>';
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
    app.classList.toggle("is-mobile-chat-open", Boolean(isActive));
    if (!isActive) {
      setContactMenuOpen(false);
      setContactPanelOpen(false);
      setGroupInfoPanelOpen(false);
    }
    updateContactActions();
    updateGroupInfoAction();
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

  function setGroupMemberMenuOpen(isOpen, x, y, target) {
    if (!isOpen || !target) {
      groupMemberMenu.hidden = true;
      groupMemberMenuTarget = null;
      return;
    }
    groupMemberMenuTarget = target;
    const removeButton = groupMemberMenu.querySelector('[data-group-member-action="remove"]');
    const canRemove = (groupInfoActorRole === "owner" || groupInfoActorRole === "admin")
      && target.role !== "owner"
      && !target.isSelf;
    if (removeButton) {
      removeButton.hidden = !canRemove;
    }
    groupMemberMenu.hidden = false;
    const rect = groupMemberMenu.getBoundingClientRect();
    const left = Math.min(Math.max(8, x), window.innerWidth - rect.width - 8);
    const top = Math.min(Math.max(8, y), window.innerHeight - rect.height - 8);
    groupMemberMenu.style.left = left + "px";
    groupMemberMenu.style.top = top + "px";
  }

  function setGroupInfoPanelOpen(isOpen) {
    if (!groupInfoPanel) {
      return;
    }
    groupInfoPanel.hidden = !isOpen;
    const stage = groupInfoPanel.closest(".desktop-stage");
    if (stage) {
      stage.classList.toggle("is-group-info-open", Boolean(isOpen));
    }
    if (!isOpen) {
      groupInfoLoadToken += 1;
      setGroupInfoMode("info");
      setGroupMemberMenuOpen(false);
    }
  }

  function setGroupInfoMode(mode) {
    groupInfoMode = mode === "add" || mode === "edit" ? mode : "info";
    const isAdd = groupInfoMode === "add";
    const isEdit = groupInfoMode === "edit";
    const isInfo = groupInfoMode === "info";
    if (groupInfoClose) {
      groupInfoClose.textContent = isInfo ? "×" : "‹";
      groupInfoClose.setAttribute("aria-label", isInfo ? "關閉群組資訊" : "返回群組資訊");
    }
    if (groupInfoBody) {
      groupInfoBody.hidden = !isInfo;
    }
    if (groupAddMembers) {
      groupAddMembers.hidden = !isAdd;
    }
    if (groupEditPanel) {
      groupEditPanel.hidden = !isEdit;
    }
    if (groupInfoHeading) {
      groupInfoHeading.textContent = isAdd ? "新增成員" : isEdit ? "編輯" : "群組資訊";
    }
    if (groupInfoEditButton) {
      groupInfoEditButton.hidden = !isInfo || !canManageActiveGroup();
    }
    if (!isAdd) {
      groupAddSearchQuery = "";
      groupAddSelectedMembers = [];
      if (groupAddSearchInput) {
        groupAddSearchInput.value = "";
      }
      renderGroupAddSelectedMembers();
    }
  }

  function setGroupAddMode(isOpen) {
    setGroupInfoMode(isOpen ? "add" : "info");
  }

  function setGroupEditMode(isOpen) {
    setGroupInfoMode(isOpen ? "edit" : "info");
  }

  function canManageActiveGroup() {
    return groupInfoActorRole === "owner" || groupInfoActorRole === "admin";
  }

  function updateGroupInfoAction() {
    if (!groupInfoOpenButton) {
      return;
    }
    const group = activeGroupConversation();
    groupInfoOpenButton.disabled = !group;
    groupInfoOpenButton.classList.toggle("is-clickable", Boolean(group));
    groupInfoOpenButton.setAttribute("aria-label", group ? "開啟群組資訊" : "目前對話");
    if (!group) {
      setGroupInfoPanelOpen(false);
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

  function renderGroupInfoLoading(group) {
    const title = group ? group.title || "群組" : "群組";
    groupInfoData = { title: title, description: "", memberCount: 0 };
    if (groupInfoAvatar) {
      groupInfoAvatar.textContent = title.slice(0, 1).toUpperCase() || "G";
    }
    if (groupInfoTitle) {
      groupInfoTitle.textContent = title;
    }
    if (groupInfoCount) {
      groupInfoCount.textContent = "載入成員中...";
    }
    if (groupInfoDescription) {
      groupInfoDescription.hidden = true;
    }
    if (groupInfoDescriptionText) {
      groupInfoDescriptionText.textContent = "";
    }
    if (groupInfoMembers) {
      groupInfoMembers.innerHTML = '<div class="group-info-empty">載入成員中...</div>';
    }
    if (groupInfoEditButton) {
      groupInfoEditButton.hidden = true;
    }
  }

  function memberRoleLabel(role) {
    if (role === "owner") {
      return "擁有者";
    }
    if (role === "admin") {
      return "管理員";
    }
    return "成員";
  }

  function renderGroupInfo(data) {
    const title = data.title || activeConversationTitle() || "群組";
    const description = data.description || "";
    const members = Array.isArray(data.members) ? data.members : [];
    const memberCount = data.member_count || members.length || 0;
    groupInfoData = { title: title, description: description, memberCount: memberCount };
    if (groupInfoAvatar) {
      groupInfoAvatar.textContent = title.slice(0, 1).toUpperCase() || "G";
    }
    if (groupInfoTitle) {
      groupInfoTitle.textContent = title;
    }
    if (groupInfoCount) {
      groupInfoCount.textContent = memberCount + " 位成員";
    }
    if (groupInfoDescription) {
      groupInfoDescription.hidden = !description;
    }
    if (groupInfoDescriptionText) {
      groupInfoDescriptionText.textContent = description;
    }
    if (!groupInfoMembers) {
      return;
    }
    if (!members.length) {
      groupInfoMemberKeys = new Set();
      groupInfoActorRole = "";
      groupInfoMembers.innerHTML = '<div class="group-info-empty">目前沒有成員資料</div>';
      if (groupInfoEditButton) {
        groupInfoEditButton.hidden = true;
      }
      return;
    }
    const actorSource = String(readSourceSystem() || "").trim();
    const actorExternalID = String(readExternalUserID() || "").trim().toLowerCase();
    groupInfoActorRole = "";
    groupInfoMemberKeys = new Set(members.map(function (member) {
      const memberSource = String(member.source_system || readSourceSystem()).trim();
      const memberExternalID = String(member.external_user_id || "").trim().toLowerCase();
      if (memberSource === actorSource && memberExternalID === actorExternalID) {
        groupInfoActorRole = member.role || "";
      }
      return memberKey({
        sourceSystem: memberSource,
        externalUserID: member.external_user_id || ""
      });
    }));
    if (groupInfoEditButton) {
      groupInfoEditButton.hidden = groupInfoMode !== "info" || !canManageActiveGroup();
    }
    groupInfoMembers.innerHTML = members.map(function (member) {
      const name = escapeHTML(member.display_name || member.external_user_id || "成員");
      const account = member.external_user_id ? "@" + member.external_user_id : memberRoleLabel(member.role);
      const initial = (member.display_name || member.external_user_id || "?").slice(0, 1).toUpperCase();
      const source = member.source_system || readSourceSystem();
      const externalID = member.external_user_id || "";
      const isSelf = String(source || "").trim() === actorSource && String(externalID || "").trim().toLowerCase() === actorExternalID;
      const roleBadge = member.role === "owner"
        ? '<span class="group-info-member-role">' + escapeHTML(memberRoleLabel(member.role)) + '</span>'
        : "";
      return '<button type="button" class="group-info-member" data-group-member-source="' + escapeHTML(source) + '" data-group-member-external="' + escapeHTML(externalID) + '" data-group-member-title="' + escapeHTML(member.display_name || member.external_user_id || "成員") + '" data-group-member-role="' + escapeHTML(member.role || "") + '" data-group-member-self="' + (isSelf ? "1" : "0") + '">'
        + '<div class="group-info-member-avatar">' + escapeHTML(initial) + '</div>'
        + '<div class="group-info-member-text">'
        + '<strong>' + name + '</strong>'
        + '<small>' + escapeHTML(account) + '</small>'
        + '</div>'
        + roleBadge
        + '</button>';
    }).join("");
  }

  function fillGroupEditPanel() {
    const title = groupInfoData.title || activeConversationTitle() || "群組";
    const description = groupInfoData.description || "";
    groupEditOriginal = { name: title, description: description };
    if (groupEditAvatar) {
      groupEditAvatar.textContent = title.slice(0, 1).toUpperCase() || "G";
    }
    if (groupEditNameInput) {
      groupEditNameInput.value = title;
    }
    if (groupEditDescriptionInput) {
      groupEditDescriptionInput.value = description;
    }
    updateGroupEditSubmitState();
  }

  function updateGroupEditSubmitState() {
    if (!groupEditSubmitButton) {
      return;
    }
    const name = groupEditNameInput ? groupEditNameInput.value.trim() : "";
    const description = groupEditDescriptionInput ? groupEditDescriptionInput.value.trim() : "";
    const dirty = name !== groupEditOriginal.name || description !== groupEditOriginal.description;
    groupEditSubmitButton.hidden = !dirty;
    groupEditSubmitButton.disabled = !dirty || !name;
  }

  function openGroupEditPanel() {
    if (!activeGroupConversation() || !canManageActiveGroup()) {
      return;
    }
    fillGroupEditPanel();
    setGroupEditMode(true);
    if (groupEditNameInput) {
      window.requestAnimationFrame(function () {
        groupEditNameInput.focus();
      });
    }
  }

  async function saveGroupEdit() {
    const headers = authHeaders();
    const group = activeGroupConversation();
    const name = groupEditNameInput ? groupEditNameInput.value.trim() : "";
    const description = groupEditDescriptionInput ? groupEditDescriptionInput.value.trim() : "";
    if (!headers || !group || !name || (groupEditSubmitButton && groupEditSubmitButton.disabled)) {
      return;
    }
    if (groupEditSubmitButton) {
      groupEditSubmitButton.disabled = true;
    }
    setMessageStatus("更新群組資訊中...", false);
    const response = await fetch("/api/conversations/" + group.conversation_id, {
      method: "PATCH",
      headers: headers,
      body: JSON.stringify({
        name: name,
        description: description
      })
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "群組資訊更新失敗。", true);
      updateGroupEditSubmitState();
      return;
    }
    renderGroupInfo(result.data || {});
    if (result.data && result.data.title) {
      const active = activeConversation();
      if (active && active.type === "group") {
        active.title = result.data.title;
      }
      if (conversationTitle) {
        conversationTitle.textContent = result.data.title;
      }
      renderConversationList(conversations);
    }
    setGroupEditMode(false);
    setMessageStatus("群組資訊已更新。", false);
    refreshConversationListOnly();
  }

  function openGroupMemberConversation(button) {
    openGroupMemberDirect({
      sourceSystem: button.dataset.groupMemberSource || readSourceSystem(),
      externalUserID: button.dataset.groupMemberExternal || "",
      title: button.dataset.groupMemberTitle || button.dataset.groupMemberExternal || "成員"
    });
  }

  function openGroupMemberDirect(member) {
    const sourceSystem = String(member.sourceSystem || readSourceSystem()).trim();
    const externalUserID = String(member.externalUserID || "").trim();
    const title = String(member.title || externalUserID || "成員").trim();
    if (!externalUserID) {
      return;
    }
    const currentSource = String(readSourceSystem() || "").trim();
    const currentExternalID = String(readExternalUserID() || "").trim().toLowerCase();
    if (sourceSystem === currentSource && externalUserID.toLowerCase() === currentExternalID) {
      setMessageStatus("這是你目前登入的帳號。", false);
      return;
    }
    setGroupInfoPanelOpen(false);
    openPendingDirectConversation({
      sourceSystem: sourceSystem,
      externalUserID: externalUserID,
      title: title
    });
  }

  async function removeGroupMember(member) {
    const headers = authHeaders();
    const group = activeGroupConversation();
    if (!headers || !group || !member || !member.externalUserID) {
      return;
    }
    setMessageStatus("移除成員中...", false);
    const response = await fetch("/api/conversations/" + group.conversation_id + "/members", {
      method: "DELETE",
      headers: headers,
      body: JSON.stringify({
        source_system: member.sourceSystem || readSourceSystem(),
        external_user_id: member.externalUserID
      })
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "移除成員失敗。", true);
      return;
    }
    renderGroupInfo(result.data || {});
    setMessageStatus("成員已移除。", false);
    refreshConversationListOnly();
  }

  async function loadGroupInfo() {
    const headers = authHeaders();
    const group = activeGroupConversation();
    if (!headers || !group) {
      return;
    }
    const requestToken = ++groupInfoLoadToken;
    renderGroupInfoLoading(group);
    let response;
    let result = {};
    try {
      response = await fetch("/api/conversations/" + group.conversation_id + "/members", {
        headers: { Authorization: headers.Authorization }
      });
      result = await parseJSON(response);
    } catch (_) {
      if (requestToken === groupInfoLoadToken && groupInfoMembers) {
        groupInfoMembers.innerHTML = '<div class="group-info-empty">群組資訊載入失敗</div>';
      }
      return;
    }
    if (requestToken !== groupInfoLoadToken || group.conversation_id !== activeConversationID) {
      return;
    }
    if (!response.ok || !result.success) {
      if (groupInfoMembers) {
        groupInfoMembers.innerHTML = '<div class="group-info-empty">' + escapeHTML(result.message || "群組資訊載入失敗") + '</div>';
      }
      return;
    }
    renderGroupInfo(result.data || {});
  }

  function isExistingGroupMember(member) {
    return groupInfoMemberKeys.has(memberKey(member));
  }

  function isGroupAddMemberSelected(member) {
    const key = memberKey(member);
    return groupAddSelectedMembers.some(function (item) {
      return memberKey(item) === key;
    });
  }

  function updateGroupAddSelectedMember(member, selected) {
    const key = memberKey(member);
    groupAddSelectedMembers = groupAddSelectedMembers.filter(function (item) {
      return memberKey(item) !== key;
    });
    if (selected) {
      groupAddSelectedMembers.push(member);
    }
    renderGroupAddSelectedMembers();
  }

  function renderGroupAddSelectedMembers() {
    if (!groupAddSelectedNode || !groupAddSubmitButton) {
      return;
    }
    groupAddSelectedNode.hidden = !groupAddSelectedMembers.length;
    groupAddSubmitButton.disabled = !groupAddSelectedMembers.length;
    groupAddSelectedNode.innerHTML = groupAddSelectedMembers.map(function (member) {
      const title = member.title || member.externalUserID || "使用者";
      const initial = title.slice(0, 1).toUpperCase();
      return [
        '<span class="new-group-selected-chip" data-group-add-selected-key="' + escapeHTML(memberKey(member)) + '">',
        '<span class="new-group-selected-avatar">' + escapeHTML(initial) + "</span>",
        "<span>" + escapeHTML(title) + "</span>",
        '<button type="button" aria-label="移除 ' + escapeHTML(title) + '">×</button>',
        "</span>"
      ].join("");
    }).join("");
  }

  function renderGroupAddResults(items, loading) {
    if (!groupAddResultsNode) {
      return;
    }
    if (loading) {
      groupAddResultsNode.innerHTML = '<div class="group-info-empty">載入聯絡人中...</div>';
      return;
    }
    const query = String(groupAddSearchQuery || "").replace(/^@+/, "").trim().toLowerCase();
    const visibleItems = items.filter(function (item) {
      const member = contactItemToGroupMember(item);
      const haystack = [member.title, member.externalUserID].join(" ").toLowerCase();
      return !isExistingGroupMember(member) && (!query || haystack.includes(query));
    });
    if (!visibleItems.length) {
      groupAddResultsNode.innerHTML = '<div class="group-info-empty">沒有可新增的聯絡人</div>';
      return;
    }
    groupAddResultsNode.innerHTML = visibleItems.map(function (item) {
      const member = contactItemToGroupMember(item);
      const selected = isGroupAddMemberSelected(member);
      const initial = (member.title || member.externalUserID || "U").slice(0, 1).toUpperCase();
      return [
        '<button type="button" class="new-group-member' + (selected ? " is-selected" : "") + '"',
        ' data-group-add-source="' + escapeHTML(member.sourceSystem) + '"',
        ' data-group-add-id="' + escapeHTML(member.externalUserID) + '"',
        ' data-group-add-title="' + escapeHTML(member.title) + '">',
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

  async function openGroupAddMembers() {
    if (!activeGroupConversation()) {
      return;
    }
    setGroupAddMode(true);
    renderGroupAddSelectedMembers();
    renderGroupAddResults(contactItems, !contactsLoaded);
    const contacts = await fetchContacts(0);
    if (!groupAddMembers || groupAddMembers.hidden) {
      return;
    }
    renderGroupAddResults(contacts || contactItems, false);
    if (groupAddSearchInput) {
      window.requestAnimationFrame(function () {
        groupAddSearchInput.focus();
      });
    }
  }

  async function addSelectedMembersToActiveGroup() {
    const headers = authHeaders();
    const group = activeGroupConversation();
    if (!headers || !group || !groupAddSelectedMembers.length) {
      return;
    }
    if (groupAddSubmitButton) {
      groupAddSubmitButton.disabled = true;
    }
    setMessageStatus("新增成員中...", false);
    const response = await fetch("/api/conversations/" + group.conversation_id + "/members", {
      method: "POST",
      headers: headers,
      body: JSON.stringify({
        members: groupAddSelectedMembers.map(function (member) {
          return {
            source_system: member.sourceSystem,
            external_user_id: member.externalUserID
          };
        })
      })
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "新增成員失敗。", true);
      renderGroupAddSelectedMembers();
      return;
    }
    groupAddSelectedMembers = [];
    renderGroupAddSelectedMembers();
    setGroupAddMode(false);
    renderGroupInfo(result.data || {});
    setMessageStatus("成員已新增。", false);
    refreshConversationListOnly();
  }

  function activeConversationTitle() {
    if (pendingDirectTarget) {
      return pendingDirectTarget.title || pendingDirectTarget.externalUserID || "";
    }
    const active = activeConversation();
    return active ? active.title : "";
  }

  function activeConversation() {
    return conversations.find(function (item) {
      return item.conversation_id === activeConversationID;
    }) || null;
  }

  function activeGroupConversation() {
    const active = activeConversation();
    return active && active.type === "group" ? active : null;
  }

  function activeDirectPeer() {
    if (pendingDirectTarget) {
      return {
        sourceSystem: pendingDirectTarget.sourceSystem || readSourceSystem(),
        externalUserID: pendingDirectTarget.externalUserID || "",
        title: pendingDirectTarget.title || pendingDirectTarget.externalUserID || ""
      };
    }
    const active = activeConversation();
    if (!active || active.type !== "direct" || !active.direct_external_user_id) {
      return null;
    }
    return {
      sourceSystem: active.direct_source_system || readSourceSystem(),
      externalUserID: active.direct_external_user_id || "",
      title: active.title || active.direct_external_user_id || ""
    };
  }

  function clearMentionMenu() {
    composerMentionMenu.hidden = true;
    composerMentionMenu.innerHTML = "";
    mentionMenuItems = [];
    mentionMenuActiveIndex = 0;
    mentionQueryRange = null;
  }

  function mentionTitle(member) {
    return String(member.display_name || member.external_user_id || "").trim();
  }

  async function loadMentionMembers(conversationID) {
    const id = Number(conversationID || 0);
    const headers = authHeaders();
    if (!id || !headers) {
      return [];
    }
    if (mentionMembersCache.has(id)) {
      return mentionMembersCache.get(id);
    }
    const response = await fetch("/api/conversations/" + id + "/members", {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (!response.ok || !result.success) {
      return [];
    }
    const members = Array.isArray(result.data && result.data.members) ? result.data.members : [];
    mentionMembersCache.set(id, members);
    return members;
  }

  function currentMentionQuery() {
    if (!composerInput || composerInput.selectionStart !== composerInput.selectionEnd) {
      return null;
    }
    const caret = composerInput.selectionStart;
    const beforeCaret = composerInput.value.slice(0, caret);
    const match = beforeCaret.match(/(^|\s)@([^\s@]*)$/);
    if (!match) {
      return null;
    }
    return {
      start: caret - match[2].length - 1,
      end: caret,
      query: match[2] || ""
    };
  }

  function renderMentionMenu(items) {
    mentionMenuItems = items;
    mentionMenuActiveIndex = Math.min(mentionMenuActiveIndex, Math.max(items.length - 1, 0));
    if (!items.length) {
      clearMentionMenu();
      return;
    }
    composerMentionMenu.innerHTML = items.map(function (item, index) {
      const active = index === mentionMenuActiveIndex ? " is-active" : "";
      const initial = item.kind === "all" ? "@" : (item.title || "?").slice(0, 1).toUpperCase();
      const subtitle = item.kind === "all" ? "標註所有成員" : "@" + item.externalUserID;
      return [
        '<button type="button" class="composer-mention-item' + active + '" data-mention-index="' + index + '" role="option">',
        '<span class="composer-mention-avatar">' + escapeHTML(initial) + "</span>",
        "<span>",
        "<strong>" + escapeHTML(item.title) + "</strong>",
        "<small>" + escapeHTML(subtitle) + "</small>",
        "</span>",
        "</button>"
      ].join("");
    }).join("");
    composerMentionMenu.hidden = false;
  }

  async function updateMentionMenu() {
    const queryInfo = currentMentionQuery();
    if (!queryInfo || !activeConversationID) {
      clearMentionMenu();
      return;
    }
    mentionQueryRange = queryInfo;
    const token = ++mentionLoadToken;
    const members = await loadMentionMembers(activeConversationID);
    if (token !== mentionLoadToken || !mentionQueryRange) {
      return;
    }
    const query = queryInfo.query.toLowerCase();
    const allItem = { kind: "all", title: "ALL", value: "@ALL" };
    const memberItems = members.map(function (member) {
      const title = mentionTitle(member);
      return {
        kind: "member",
        title: title || member.external_user_id || "使用者",
        externalUserID: member.external_user_id || "",
        value: "@" + (title || member.external_user_id || "")
      };
    }).filter(function (item) {
      const text = (item.title + " " + item.externalUserID).toLowerCase();
      return !query || text.indexOf(query) >= 0;
    });
    const items = [allItem].concat(memberItems).filter(function (item) {
      const text = (item.title + " " + (item.externalUserID || "")).toLowerCase();
      return !query || text.indexOf(query) >= 0;
    }).slice(0, 8);
    mentionMenuActiveIndex = 0;
    renderMentionMenu(items);
  }

  function insertMentionItem(item) {
    if (!composerInput || !mentionQueryRange || !item) {
      return;
    }
    const before = composerInput.value.slice(0, mentionQueryRange.start);
    const after = composerInput.value.slice(mentionQueryRange.end);
    const insertion = item.value + " ";
    composerInput.value = before + insertion + after;
    const caret = before.length + insertion.length;
    composerInput.setSelectionRange(caret, caret);
    clearMentionMenu();
    syncComposerInputHeight();
    composerInput.focus();
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
    clearMentionMenu();
    lastReportedReadMessageID = 0;
    pendingReadMessageID = 0;
    const key = accountStorageKey(storageKeys.activeConversationID);
    if (key) {
      localStorage.setItem(key, String(activeConversationID));
    }
    closeSearchMode();
    renderConversationList(conversations);
    updateContactActions();
    updateGroupInfoAction();
    setGroupInfoPanelOpen(false);
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
    clearMentionMenu();
    pendingDirectTarget = null;
    activeConversationID = 0;
    lastReportedReadMessageID = 0;
    pendingReadMessageID = 0;
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

  function contactItemToGroupMember(item) {
    const title = contactListTitle(item);
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
      renderNewGroupContactResults(contactItems, contactsIndexLoading);
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

  function renderNewGroupContactResults(items, loading) {
    if (!newGroupResultsNode) {
      return;
    }
    if (loading) {
      newGroupResultsNode.innerHTML = '<div class="new-group-empty">載入聯絡人中...</div>';
      return;
    }
    if (!items.length) {
      newGroupResultsNode.innerHTML = '<div class="new-group-empty">尚未加入聯絡人，可輸入名稱或帳號搜尋成員</div>';
      return;
    }
    newGroupResultsNode.innerHTML = items.map(function (item) {
      const member = contactItemToGroupMember(item);
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

  async function renderNewGroupDefaultContacts() {
    if (newGroupSearchQuery) {
      return;
    }
    if (contactsLoaded) {
      renderNewGroupContactResults(contactItems, false);
      return;
    }
    renderNewGroupContactResults([], true);
    const contacts = await fetchContacts(0);
    if (newGroupSearchQuery) {
      return;
    }
    if (contacts) {
      renderNewGroupContactResults(contacts, false);
      return;
    }
    renderNewGroupContactResults([], false);
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
    rememberContactItem(result.data || {
      source_system: peer.sourceSystem,
      external_user_id: peer.externalUserID,
      display_name: peer.title,
      alias_name: contactAliasName ? contactAliasName.value.trim() : "",
      status: "active"
    });
    setContactPanelOpen(false);
    updateContactActions();
    const refreshedContacts = await fetchContacts(0);
    const latestContacts = refreshedContacts || contactItems;
    if (sidebarShell && sidebarShell.dataset.sidebarView === "contacts") {
      renderContacts(latestContacts, false);
    }
    if (groupAddMembers && !groupAddMembers.hidden) {
      renderGroupAddResults(latestContacts, false);
    }
    setMessageStatus("已加入聯絡人。", false);
  }

  function contactListTitle(item) {
    return item.alias_name || item.display_name || item.external_user_id || "聯絡人";
  }

  function rememberContactItem(item) {
    if (!item || !item.external_user_id) {
      return;
    }
    const key = contactKey(item.source_system, item.external_user_id);
    contactItems = contactItems.filter(function (contact) {
      return contactKey(contact.source_system, contact.external_user_id) !== key;
    });
    contactItems.push(item);
    contactItems.sort(function (a, b) {
      return contactListTitle(a).localeCompare(contactListTitle(b), "zh-Hant");
    });
    contactKeys.add(key);
    contactExternalIDs.add(String(item.external_user_id || "").trim().toLowerCase());
    contactsLoaded = true;
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
      contactItems = [];
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
      contactItems = [];
      if (contactsList && requestToken === contactsLoadToken) {
        contactsList.innerHTML = '<div class="contacts-empty is-error">' + escapeHTML(result.message || "聯絡人載入失敗") + '</div>';
      }
      return null;
    }
    const data = result.data || {};
    const contacts = data.contacts || [];
    contactItems = contacts;
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

  function renderMessages(data, options) {
    options = options || {};
    if (!data.messages || !data.messages.length) {
      messageBoard.innerHTML = "";
      if (!options.preserveScroll) {
        scrollMessageBoardToLatest();
      }
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
    if (options.preserveScroll && options.scrollAnchor) {
      restoreMessageScrollAnchor(options.scrollAnchor);
    } else if (options.forceBottom || options.stickToBottom) {
      scrollMessageBoardToLatest();
    } else if (!scrollMessageBoardToUnreadDivider() && !scrollMessageBoardToMessage(firstUnreadMessageID)) {
      scrollMessageBoardToLatest();
    }
    scheduleVisibleReadUpdate();
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
      scheduleVisibleReadUpdate();
    });
  }

  function isMessageBoardNearBottom() {
    if (!messageBoard) {
      return true;
    }
    return messageBoard.scrollHeight - messageBoard.scrollTop - messageBoard.clientHeight < 72;
  }

  function captureMessageScrollAnchor() {
    if (!messageBoard) {
      return null;
    }
    const boardRect = messageBoard.getBoundingClientRect();
    const rows = Array.from(messageBoard.querySelectorAll("[data-message-id]"));
    for (const row of rows) {
      const rect = row.getBoundingClientRect();
      if (rect.bottom >= boardRect.top + 8) {
        return {
          messageID: row.dataset.messageId || "",
          offset: rect.top - boardRect.top
        };
      }
    }
    return null;
  }

  function restoreMessageScrollAnchor(anchor) {
    if (!messageBoard || !anchor || !anchor.messageID) {
      return;
    }
    window.requestAnimationFrame(function () {
      const target = messageBoard.querySelector('[data-message-id="' + CSS.escape(String(anchor.messageID)) + '"]');
      if (!target) {
        return;
      }
      const boardRect = messageBoard.getBoundingClientRect();
      const rect = target.getBoundingClientRect();
      messageBoard.scrollTop += rect.top - boardRect.top - anchor.offset;
      scheduleVisibleReadUpdate();
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
      scheduleVisibleReadUpdate();
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
      scheduleVisibleReadUpdate();
    });
    return true;
  }

  function highestVisibleIncomingMessageID() {
    if (!messageBoard || !activeConversationID) {
      return 0;
    }
    const boardRect = messageBoard.getBoundingClientRect();
    let highest = 0;
    messageBoard.querySelectorAll(".message-bubble[data-message-id]").forEach(function (bubble) {
      if (bubble.dataset.outgoing === "true") {
        return;
      }
      const rect = bubble.getBoundingClientRect();
      const visibleHeight = Math.min(rect.bottom, boardRect.bottom) - Math.max(rect.top, boardRect.top);
      const requiredHeight = Math.min(48, Math.max(20, rect.height * 0.45));
      if (visibleHeight >= requiredHeight) {
        highest = Math.max(highest, Number(bubble.dataset.messageId || 0));
      }
    });
    return highest;
  }

  function scheduleVisibleReadUpdate() {
    if (readUpdateTimer) {
      window.clearTimeout(readUpdateTimer);
    }
    readUpdateTimer = window.setTimeout(reportVisibleReadMessage, 180);
  }

  async function reportVisibleReadMessage() {
    readUpdateTimer = 0;
    const messageID = highestVisibleIncomingMessageID();
    if (!messageID || messageID <= lastReportedReadMessageID) {
      return;
    }
    pendingReadMessageID = Math.max(pendingReadMessageID, messageID);
    if (readUpdateInFlight) {
      return;
    }
    const conversationID = activeConversationID;
    const headers = authHeaders();
    if (!headers || !conversationID) {
      return;
    }

    readUpdateInFlight = true;
    const readMessageID = pendingReadMessageID;
    pendingReadMessageID = 0;
    try {
      const response = await fetch("/api/conversations/" + conversationID + "/read", {
        method: "POST",
        headers: headers,
        body: JSON.stringify({ last_read_message_id: readMessageID })
      });
      const result = await parseJSON(response);
      if (response.ok && result.success && conversationID === activeConversationID) {
        lastReportedReadMessageID = Math.max(lastReportedReadMessageID, readMessageID);
        refreshConversationListOnly();
      }
    } catch (_) {
      pendingReadMessageID = Math.max(pendingReadMessageID, readMessageID);
    } finally {
      readUpdateInFlight = false;
      if (pendingReadMessageID > lastReportedReadMessageID) {
        scheduleVisibleReadUpdate();
      }
    }
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

  function imageExtensionForMime(type) {
    const normalized = String(type || "").toLowerCase();
    if (normalized === "image/jpeg") {
      return "jpg";
    }
    if (normalized === "image/png") {
      return "png";
    }
    return "";
  }

  function normalizePastedImageFile(file, index) {
    if (!file || !isImageFile(file)) {
      return null;
    }
    const name = String(file.name || "").trim();
    if (/\.(jpg|jpeg|png)$/i.test(name)) {
      return file;
    }
    const extension = imageExtensionForMime(file.type);
    if (!extension || typeof File !== "function") {
      return file;
    }
    const timestamp = new Date().toISOString().replace(/[-:]/g, "").replace(/\..+$/, "");
    return new File([file], "pasted-image-" + timestamp + "-" + (index + 1) + "." + extension, {
      type: file.type || "image/" + extension,
      lastModified: Date.now()
    });
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

  function renderEditablePhoto(file, index) {
    return [
      '<span class="attachment-photo-editable">',
      '<canvas class="attachment-photo-canvas" data-attachment-photo-canvas="' + index + '" aria-label="' + escapeHTML(file.name || "photo") + '"></canvas>',
      '<button type="button" class="attachment-photo-edit-button" data-photo-edit-index="' + index + '" aria-label="編輯圖片">',
      '<svg viewBox="0 0 24 24" aria-hidden="true">',
      '<path d="M4 7h10"></path>',
      '<path d="M18 7h2"></path>',
      '<path d="M4 17h2"></path>',
      '<path d="M10 17h10"></path>',
      '<path d="M7 4v6"></path>',
      '<path d="M17 14v6"></path>',
      '</svg>',
      '</button>',
      '</span>'
    ].join("");
  }

  function drawAttachmentPhotoCanvas(canvas, file) {
    if (!canvas || !file) {
      return;
    }
    loadImageFromFile(file).then(function (image) {
      const ctx = canvas.getContext("2d");
      canvas.width = image.naturalWidth;
      canvas.height = image.naturalHeight;
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.drawImage(image, 0, 0);
    }).catch(function () {
      canvas.hidden = true;
    });
  }

  function renderAttachmentPhotoCanvases() {
    if (!attachmentPreview) {
      return;
    }
    attachmentPreview.querySelectorAll("[data-attachment-photo-canvas]").forEach(function (canvas) {
      const index = Number(canvas.dataset.attachmentPhotoCanvas || 0);
      drawAttachmentPhotoCanvas(canvas, selectedAttachmentFiles[index]);
    });
  }

  function renderAttachmentPhotoGrid(files) {
    return [
      '<div class="' + photoGridClass(files.length) + '">',
      files.map(function (file, index) {
        return renderEditablePhoto(file, index);
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
      attachmentPreview.innerHTML = renderEditablePhoto(selectedAttachmentFiles[0], 0);
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
    renderAttachmentPhotoCanvases();
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

  function clipboardImageFiles(event) {
    const clipboard = event.clipboardData;
    if (!clipboard) {
      return [];
    }
    const itemFiles = clipboard.items ? Array.from(clipboard.items).map(function (item) {
      if (item.kind !== "file" || String(item.type || "").indexOf("image/") !== 0) {
        return null;
      }
      return item.getAsFile();
    }).filter(Boolean) : [];
    const files = itemFiles.length ? itemFiles : Array.from(clipboard.files || []).filter(isImageFile);
    return files.map(normalizePastedImageFile).filter(Boolean);
  }

  function isEditableTarget(target) {
    if (!target || target === document.body) {
      return false;
    }
    if (target.isContentEditable) {
      return true;
    }
    const tagName = String(target.tagName || "").toLowerCase();
    return tagName === "input" || tagName === "textarea" || tagName === "select";
  }

  function shouldHandlePasteUpload(target) {
    if (!activeConversationID || attachmentDialog && !attachmentDialog.hidden) {
      return false;
    }
    if (!app.contains(target)) {
      return false;
    }
    if (isEditableTarget(target) && target !== composerInput) {
      return false;
    }
    return true;
  }

  function handlePasteUpload(event) {
    if (!shouldHandlePasteUpload(event.target)) {
      return;
    }
    const files = clipboardImageFiles(event);
    if (!files.length) {
      return;
    }
    event.preventDefault();
    if (composerFileInput) {
      composerFileInput.value = "";
    }
    setAttachmentMenuOpen(false);
    openAttachmentDialog(files, "photo");
  }

  function clamp(value, min, max) {
    return Math.min(Math.max(value, min), max);
  }

  function loadImageFromFile(file) {
    return new Promise(function (resolve, reject) {
      const url = URL.createObjectURL(file);
      const image = new Image();
      image.onload = function () {
        URL.revokeObjectURL(url);
        resolve(image);
      };
      image.onerror = function () {
        URL.revokeObjectURL(url);
        reject(new Error("image load failed"));
      };
      image.src = url;
    });
  }

  function resetPhotoEditorAnnotation(state) {
    state.annotationCanvas = document.createElement("canvas");
    state.annotationCanvas.width = state.image.naturalWidth;
    state.annotationCanvas.height = state.image.naturalHeight;
    state.markerCanvas = document.createElement("canvas");
    state.markerCanvas.width = state.image.naturalWidth;
    state.markerCanvas.height = state.image.naturalHeight;
  }

  function photoEditorCanvasImageData(canvas) {
    return canvas.getContext("2d").getImageData(0, 0, canvas.width, canvas.height);
  }

  function photoEditorSnapshot() {
    const state = photoEditorState;
    if (!state) {
      return null;
    }
    return {
      crop: Object.assign({}, state.crop),
      ratio: state.ratio,
      annotation: photoEditorCanvasImageData(state.annotationCanvas),
      marker: state.markerCanvas ? photoEditorCanvasImageData(state.markerCanvas) : null
    };
  }

  function restorePhotoEditorSnapshot(snapshot) {
    const state = photoEditorState;
    if (!state || !snapshot) {
      return;
    }
    state.crop = Object.assign({}, snapshot.crop);
    state.ratio = snapshot.ratio;
    state.annotationCanvas.getContext("2d").putImageData(snapshot.annotation, 0, 0);
    if (snapshot.marker && state.markerCanvas) {
      state.markerCanvas.getContext("2d").putImageData(snapshot.marker, 0, 0);
    }
    photoRatioButtons.forEach(function (button) {
      button.classList.toggle("is-active", button.dataset.photoRatio === state.ratio);
    });
    renderPhotoEditor();
    updatePhotoEditorHistoryButtons();
  }

  function updatePhotoEditorHistoryButtons() {
    if (photoEditorUndo) {
      photoEditorUndo.disabled = !photoEditorState || !photoEditorState.undoStack.length;
    }
    if (photoEditorRedo) {
      photoEditorRedo.disabled = !photoEditorState || !photoEditorState.redoStack.length;
    }
  }

  function updatePhotoToolIcons() {
    if (!photoEditor || !photoEditorState) {
      return;
    }
    photoEditor.style.setProperty("--photo-tool-color", photoEditorState.color || "#ff4b40");
  }

  function updatePhotoBrushRange() {
    if (!photoBrushSizeInput) {
      return;
    }
    const min = Number(photoBrushSizeInput.min || 0);
    const max = Number(photoBrushSizeInput.max || 100);
    const value = Number(photoBrushSizeInput.value || min);
    const progress = max > min ? ((value - min) / (max - min)) * 100 : 0;
    photoBrushSizeInput.style.setProperty("--range-progress", Math.min(Math.max(progress, 0), 100) + "%");
  }

  function hsvToRGB(hue, saturation, value) {
    const chroma = value * saturation;
    const section = hue / 60;
    const x = chroma * (1 - Math.abs(section % 2 - 1));
    let red = 0;
    let green = 0;
    let blue = 0;
    if (section >= 0 && section < 1) {
      red = chroma;
      green = x;
    } else if (section < 2) {
      red = x;
      green = chroma;
    } else if (section < 3) {
      green = chroma;
      blue = x;
    } else if (section < 4) {
      green = x;
      blue = chroma;
    } else if (section < 5) {
      red = x;
      blue = chroma;
    } else {
      red = chroma;
      blue = x;
    }
    const match = value - chroma;
    return {
      red: Math.round((red + match) * 255),
      green: Math.round((green + match) * 255),
      blue: Math.round((blue + match) * 255)
    };
  }

  function componentToHex(value) {
    return value.toString(16).padStart(2, "0");
  }

  function rgbToHex(rgb) {
    return "#" + componentToHex(rgb.red) + componentToHex(rgb.green) + componentToHex(rgb.blue);
  }

  function hexToRGB(hex) {
    const normalized = String(hex || "").trim().replace(/^#/, "");
    if (!/^[0-9a-fA-F]{6}$/.test(normalized)) {
      return null;
    }
    return {
      red: parseInt(normalized.slice(0, 2), 16),
      green: parseInt(normalized.slice(2, 4), 16),
      blue: parseInt(normalized.slice(4, 6), 16)
    };
  }

  function rgbToHSV(rgb) {
    const red = rgb.red / 255;
    const green = rgb.green / 255;
    const blue = rgb.blue / 255;
    const max = Math.max(red, green, blue);
    const min = Math.min(red, green, blue);
    const delta = max - min;
    let hue = 0;
    if (delta) {
      if (max === red) {
        hue = 60 * (((green - blue) / delta) % 6);
      } else if (max === green) {
        hue = 60 * ((blue - red) / delta + 2);
      } else {
        hue = 60 * ((red - green) / delta + 4);
      }
    }
    if (hue < 0) {
      hue += 360;
    }
    return {
      hue: Math.round(hue),
      saturation: max ? delta / max : 0,
      value: max
    };
  }

  function setPhotoEditorColor(color, options) {
    if (!photoEditorState) {
      return;
    }
    photoEditorState.color = color;
    if (photoEditorState.toolColors && photoEditorState.tool) {
      photoEditorState.toolColors[photoEditorState.tool] = color;
    }
    updatePhotoToolIcons();
    photoColorButtons.forEach(function (item) {
      item.classList.toggle("is-active", item.dataset.photoColor === color && !(options && options.custom));
    });
    if (photoCustomColorButton) {
      photoCustomColorButton.classList.toggle("is-active", Boolean(options && options.custom));
      photoCustomColorButton.style.setProperty("--swatch", color);
    }
  }

  function updateCustomColorUI(color) {
    const rgb = hexToRGB(color);
    if (!rgb) {
      return;
    }
    const hsv = rgbToHSV(rgb);
    if (photoColorHueInput) {
      photoColorHueInput.value = String(hsv.hue);
      photoColorHueInput.style.setProperty("--photo-color-hue", hsv.hue);
    }
    if (photoColorSV) {
      photoColorSV.style.setProperty("--photo-color-hue", hsv.hue);
      photoColorSV.style.setProperty("--photo-color-saturation", Math.round(hsv.saturation * 100));
      photoColorSV.style.setProperty("--photo-color-value", Math.round(hsv.value * 100));
    }
    if (photoColorHexInput) {
      photoColorHexInput.value = color;
    }
    if (photoColorRGBInput) {
      photoColorRGBInput.value = rgb.red + ", " + rgb.green + ", " + rgb.blue;
    }
  }

  function applyCustomHSV(saturation, value) {
    const hue = Number(photoColorHueInput ? photoColorHueInput.value : 0);
    const rgb = hsvToRGB(hue, saturation, value);
    const color = rgbToHex(rgb);
    updateCustomColorUI(color);
    setPhotoEditorColor(color, { custom: true });
  }

  function pushPhotoEditorHistory() {
    const state = photoEditorState;
    const snapshot = photoEditorSnapshot();
    if (!state || !snapshot) {
      return;
    }
    state.undoStack.push(snapshot);
    if (state.undoStack.length > 30) {
      state.undoStack.shift();
    }
    state.redoStack = [];
    state.hasChanges = true;
    updatePhotoEditorHistoryButtons();
  }

  function undoPhotoEditorChange() {
    const state = photoEditorState;
    if (!state || !state.undoStack.length) {
      return;
    }
    const current = photoEditorSnapshot();
    const previous = state.undoStack.pop();
    if (current) {
      state.redoStack.push(current);
    }
    restorePhotoEditorSnapshot(previous);
  }

  function redoPhotoEditorChange() {
    const state = photoEditorState;
    if (!state || !state.redoStack.length) {
      return;
    }
    const current = photoEditorSnapshot();
    const next = state.redoStack.pop();
    if (current) {
      state.undoStack.push(current);
    }
    restorePhotoEditorSnapshot(next);
  }

  function isPhotoEditorOpen() {
    return Boolean(photoEditor && !photoEditor.hidden && photoEditorState);
  }

  function ratioNumber(ratio) {
    if (!photoEditorState || ratio === "free") {
      return 0;
    }
    if (ratio === "original") {
      return photoEditorState.image.naturalWidth / photoEditorState.image.naturalHeight;
    }
    const parts = ratio.split(":");
    return Number(parts[0]) / Number(parts[1]);
  }

  function setPhotoEditorCropForRatio(ratio) {
    const state = photoEditorState;
    if (!state) {
      return;
    }
    state.ratio = ratio;
    const width = state.image.naturalWidth;
    const height = state.image.naturalHeight;
    const aspect = ratioNumber(ratio);
    let cropWidth = width;
    let cropHeight = height;
    if (aspect > 0) {
      if (cropWidth / cropHeight > aspect) {
        cropWidth = cropHeight * aspect;
      } else {
        cropHeight = cropWidth / aspect;
      }
    }
    state.crop = {
      x: (width - cropWidth) / 2,
      y: (height - cropHeight) / 2,
      width: cropWidth,
      height: cropHeight
    };
    renderPhotoEditor();
    updatePhotoEditorHistoryButtons();
  }

  function setPhotoEditorMode(mode) {
    const state = photoEditorState;
    if (!state) {
      return;
    }
    state.mode = mode === "draw" ? "draw" : "crop";
    photoEditorModeButtons.forEach(function (button) {
      button.classList.toggle("is-active", button.dataset.photoEditorMode === state.mode);
    });
    photoEditorPanels.forEach(function (panel) {
      panel.hidden = panel.dataset.photoEditorPanel !== state.mode;
    });
    if (photoEditorCanvas) {
      photoEditorCanvas.style.cursor = state.mode === "draw" ? "crosshair" : "move";
    }
    renderPhotoEditor();
  }

  function imagePointFromEvent(event) {
    const rect = photoEditorCanvas.getBoundingClientRect();
    return {
      x: clamp((event.clientX - rect.left) * (photoEditorCanvas.width / rect.width), 0, photoEditorCanvas.width),
      y: clamp((event.clientY - rect.top) * (photoEditorCanvas.height / rect.height), 0, photoEditorCanvas.height)
    };
  }

  function drawCropOverlay(ctx, crop) {
    ctx.save();
    ctx.fillStyle = "rgba(0, 0, 0, 0.5)";
    ctx.beginPath();
    ctx.rect(0, 0, photoEditorCanvas.width, photoEditorCanvas.height);
    ctx.rect(crop.x, crop.y, crop.width, crop.height);
    ctx.fill("evenodd");
    ctx.strokeStyle = "rgba(255, 255, 255, 0.95)";
    ctx.lineWidth = Math.max(2, photoEditorCanvas.width / 900);
    ctx.strokeRect(crop.x, crop.y, crop.width, crop.height);
    ctx.strokeStyle = "rgba(255, 255, 255, 0.32)";
    ctx.lineWidth = 1;
    for (let i = 1; i < 3; i += 1) {
      const x = crop.x + crop.width * i / 3;
      const y = crop.y + crop.height * i / 3;
      ctx.beginPath();
      ctx.moveTo(x, crop.y);
      ctx.lineTo(x, crop.y + crop.height);
      ctx.moveTo(crop.x, y);
      ctx.lineTo(crop.x + crop.width, y);
      ctx.stroke();
    }
    ctx.fillStyle = "#ffffff";
    [[crop.x, crop.y], [crop.x + crop.width, crop.y], [crop.x, crop.y + crop.height], [crop.x + crop.width, crop.y + crop.height]].forEach(function (point) {
      ctx.beginPath();
      ctx.arc(point[0], point[1], Math.max(5, photoEditorCanvas.width / 220), 0, Math.PI * 2);
      ctx.fill();
    });
    ctx.restore();
  }

  function renderPhotoEditor() {
    const state = photoEditorState;
    if (!state || !photoEditorCanvas) {
      return;
    }
    photoEditorCanvas.width = state.image.naturalWidth;
    photoEditorCanvas.height = state.image.naturalHeight;
    const ctx = photoEditorCanvas.getContext("2d");
    ctx.clearRect(0, 0, photoEditorCanvas.width, photoEditorCanvas.height);
    ctx.drawImage(state.image, 0, 0);
    ctx.drawImage(state.annotationCanvas, 0, 0);
    if (state.markerCanvas) {
      ctx.save();
      ctx.globalAlpha = 0.34;
      ctx.drawImage(state.markerCanvas, 0, 0);
      ctx.restore();
    }
    if (state.mode === "crop") {
      drawCropOverlay(ctx, state.crop);
    }
  }

  function cropContainsPoint(crop, point) {
    return point.x >= crop.x && point.x <= crop.x + crop.width && point.y >= crop.y && point.y <= crop.y + crop.height;
  }

  function cropCoversFullImage(state) {
    if (!state || !state.crop || !state.image) {
      return false;
    }
    return state.crop.x <= 0
      && state.crop.y <= 0
      && state.crop.width >= state.image.naturalWidth
      && state.crop.height >= state.image.naturalHeight;
  }

  function moveCropBy(deltaX, deltaY) {
    const state = photoEditorState;
    const crop = state.crop;
    crop.x = clamp(crop.x + deltaX, 0, state.image.naturalWidth - crop.width);
    crop.y = clamp(crop.y + deltaY, 0, state.image.naturalHeight - crop.height);
  }

  function updateFreeCrop(start, point) {
    const state = photoEditorState;
    const minSize = 32;
    const x = Math.min(start.x, point.x);
    const y = Math.min(start.y, point.y);
    const width = Math.max(minSize, Math.abs(point.x - start.x));
    const height = Math.max(minSize, Math.abs(point.y - start.y));
    state.crop = {
      x: clamp(x, 0, state.image.naturalWidth - minSize),
      y: clamp(y, 0, state.image.naturalHeight - minSize),
      width: Math.min(width, state.image.naturalWidth - x),
      height: Math.min(height, state.image.naturalHeight - y)
    };
  }

  function drawBlurLine(from, to) {
    const state = photoEditorState;
    const brushWidth = state.brushSize;
    const blurAmount = Math.max(4, state.brushSize * 0.42);
    const padding = brushWidth + blurAmount * 2;
    const left = Math.max(0, Math.floor(Math.min(from.x, to.x) - padding));
    const top = Math.max(0, Math.floor(Math.min(from.y, to.y) - padding));
    const right = Math.min(state.image.naturalWidth, Math.ceil(Math.max(from.x, to.x) + padding));
    const bottom = Math.min(state.image.naturalHeight, Math.ceil(Math.max(from.y, to.y) + padding));
    const width = Math.max(1, right - left);
    const height = Math.max(1, bottom - top);

    const blurred = document.createElement("canvas");
    blurred.width = width;
    blurred.height = height;
    const blurredCtx = blurred.getContext("2d");
    blurredCtx.filter = "blur(" + blurAmount + "px)";
    blurredCtx.drawImage(state.image, left, top, width, height, 0, 0, width, height);

    const mask = document.createElement("canvas");
    mask.width = width;
    mask.height = height;
    const maskCtx = mask.getContext("2d");
    maskCtx.lineCap = "round";
    maskCtx.lineJoin = "round";
    maskCtx.lineWidth = brushWidth;
    maskCtx.strokeStyle = "#fff";
    maskCtx.beginPath();
    maskCtx.moveTo(from.x - left, from.y - top);
    maskCtx.lineTo(to.x - left, to.y - top);
    maskCtx.stroke();

    blurredCtx.globalCompositeOperation = "destination-in";
    blurredCtx.drawImage(mask, 0, 0);

    const ctx = state.annotationCanvas.getContext("2d");
    ctx.save();
    ctx.globalCompositeOperation = "source-over";
    ctx.drawImage(blurred, left, top);
    ctx.restore();
  }

  function drawEditorLine(from, to) {
    const state = photoEditorState;
    if (state.tool === "blur") {
      drawBlurLine(from, to);
      return;
    }
    const targetCanvas = state.tool === "marker" && state.markerCanvas ? state.markerCanvas : state.annotationCanvas;
    const ctx = targetCanvas.getContext("2d");
    ctx.save();
    ctx.lineCap = "round";
    ctx.lineJoin = "round";
    ctx.lineWidth = state.brushSize;
    if (state.tool === "eraser") {
      ctx.globalCompositeOperation = "destination-out";
      ctx.strokeStyle = "rgba(0, 0, 0, 1)";
    } else if (state.tool === "marker") {
      ctx.globalCompositeOperation = "source-over";
      ctx.globalAlpha = 1;
      ctx.strokeStyle = state.color;
    } else {
      ctx.globalCompositeOperation = "source-over";
      ctx.strokeStyle = state.color;
    }
    ctx.beginPath();
    ctx.moveTo(from.x, from.y);
    ctx.lineTo(to.x, to.y);
    ctx.stroke();
    ctx.restore();

    if (state.tool === "eraser" && state.markerCanvas) {
      const markerCtx = state.markerCanvas.getContext("2d");
      markerCtx.save();
      markerCtx.lineCap = "round";
      markerCtx.lineJoin = "round";
      markerCtx.lineWidth = state.brushSize;
      markerCtx.globalCompositeOperation = "destination-out";
      markerCtx.beginPath();
      markerCtx.moveTo(from.x, from.y);
      markerCtx.lineTo(to.x, to.y);
      markerCtx.stroke();
      markerCtx.restore();
    }
  }

  function drawContinuousEditorLine(from, to) {
    const state = photoEditorState;
    if (state.tool === "blur") {
      drawEditorLine(from, to);
      return;
    }
    const distance = Math.hypot(to.x - from.x, to.y - from.y);
    const step = Math.max(4, state.brushSize * 0.45);
    const segments = Math.max(1, Math.ceil(distance / step));
    let previous = from;
    for (let index = 1; index <= segments; index += 1) {
      const point = {
        x: from.x + (to.x - from.x) * index / segments,
        y: from.y + (to.y - from.y) * index / segments
      };
      drawEditorLine(previous, point);
      previous = point;
    }
  }

  function drawArrowHeadOnContext(ctx, from, to, color, size) {
    const distance = Math.hypot(to.x - from.x, to.y - from.y);
    if (distance < 4) {
      return;
    }
    const angle = Math.atan2(to.y - from.y, to.x - from.x);
    const headLength = Math.min(Math.max(size * 3.5, 30), Math.max(distance * 0.65, size * 2.4));
    const wingAngle = Math.PI / 4.2;
    ctx.save();
    ctx.lineCap = "round";
    ctx.lineJoin = "round";
    ctx.lineWidth = size;
    ctx.strokeStyle = color;
    ctx.beginPath();
    ctx.moveTo(to.x, to.y);
    ctx.lineTo(to.x - headLength * Math.cos(angle - wingAngle), to.y - headLength * Math.sin(angle - wingAngle));
    ctx.moveTo(to.x, to.y);
    ctx.lineTo(to.x - headLength * Math.cos(angle + wingAngle), to.y - headLength * Math.sin(angle + wingAngle));
    ctx.stroke();
    ctx.restore();
  }

  function drawEditorArrow(from, to) {
    const state = photoEditorState;
    const ctx = state.annotationCanvas.getContext("2d");
    drawArrowHeadOnContext(ctx, from, to, state.color, state.brushSize);
  }

  function arrowHeadDirectionPoint(points, endPoint) {
    for (let index = points.length - 2; index >= 0; index -= 1) {
      const point = points[index];
      if (Math.hypot(endPoint.x - point.x, endPoint.y - point.y) >= Math.max(12, photoEditorState.brushSize * 1.4)) {
        return point;
      }
    }
    return points[0] || endPoint;
  }

  async function openPhotoEditor(index) {
    const file = selectedAttachmentFiles[index];
    if (!file || !isImageFile(file) || !photoEditor || !photoEditorCanvas) {
      return;
    }
    try {
      const image = await loadImageFromFile(file);
      photoEditorState = {
        fileIndex: index,
        file: file,
        image: image,
        mode: "crop",
        ratio: "free",
        crop: { x: 0, y: 0, width: image.naturalWidth, height: image.naturalHeight },
        color: photoToolDefaultColors.pen,
        toolColors: Object.assign({}, photoToolDefaultColors),
        brushSize: 18,
        tool: "pen",
        pointerMode: "",
        pointerStart: null,
        previewPoint: null,
        arrowPoints: [],
        lastPoint: null,
        undoStack: [],
        redoStack: [],
        hasChanges: false
      };
      resetPhotoEditorAnnotation(photoEditorState);
      photoRatioButtons.forEach(function (button) {
        button.classList.toggle("is-active", button.dataset.photoRatio === "free");
      });
      photoColorButtons.forEach(function (button) {
        button.classList.toggle("is-active", button.dataset.photoColor === photoEditorState.color);
      });
      if (photoCustomColorButton) {
        photoCustomColorButton.classList.remove("is-active");
      }
      if (photoCustomColorPanel) {
        photoCustomColorPanel.hidden = true;
      }
      updateCustomColorUI("#ffd21f");
      photoToolButtons.forEach(function (button) {
        button.classList.toggle("is-active", button.dataset.photoTool === photoEditorState.tool);
      });
      if (photoBrushSizeInput) {
        photoBrushSizeInput.value = String(photoEditorState.brushSize);
        updatePhotoBrushRange();
      }
      if (photoEditor.parentElement !== document.body) {
        document.body.appendChild(photoEditor);
      }
      photoEditor.hidden = false;
      updatePhotoToolIcons();
      setPhotoEditorMode("crop");
      setPhotoEditorCropForRatio("free");
      updatePhotoEditorHistoryButtons();
    } catch (_) {
      setMessageStatus("圖片讀取失敗，請重新選擇圖片。", true);
    }
  }

  function hidePhotoEditorConfirm() {
    if (photoEditorConfirm) {
      photoEditorConfirm.hidden = true;
    }
  }

  function showPhotoEditorConfirm() {
    if (photoEditorConfirm) {
      photoEditorConfirm.hidden = false;
    }
  }

  function closePhotoEditor() {
    if (photoEditor) {
      photoEditor.hidden = true;
    }
    hidePhotoEditorConfirm();
    if (photoEditorState) {
      photoEditorState.pointerMode = "";
      photoEditorState.pointerStart = null;
      photoEditorState.previewPoint = null;
      photoEditorState.arrowPoints = [];
      photoEditorState.lastPoint = null;
      photoEditorState.cropStart = null;
    }
    photoEditorState = null;
    updatePhotoEditorHistoryButtons();
  }

  function requestClosePhotoEditor() {
    if (photoEditorState && photoEditorState.hasChanges) {
      showPhotoEditorConfirm();
      return;
    }
    closePhotoEditor();
  }

  function resetPhotoEditor() {
    if (!photoEditorState) {
      return;
    }
    pushPhotoEditorHistory();
    resetPhotoEditorAnnotation(photoEditorState);
    setPhotoEditorCropForRatio(photoEditorState.ratio);
  }

  function handlePhotoEditorPointerDown(event) {
    if (!photoEditorState || !photoEditorCanvas) {
      return;
    }
    event.preventDefault();
    photoEditorCanvas.setPointerCapture(event.pointerId);
    const point = imagePointFromEvent(event);
    if (photoEditorState.mode === "draw") {
      pushPhotoEditorHistory();
      photoEditorState.pointerMode = "draw";
      photoEditorState.pointerStart = point;
      photoEditorState.previewPoint = null;
      photoEditorState.arrowPoints = photoEditorState.tool === "arrow" ? [point] : [];
      photoEditorState.lastPoint = point;
      if (photoEditorState.tool !== "arrow") {
        drawEditorLine(point, point);
      }
      renderPhotoEditor();
      return;
    }
    pushPhotoEditorHistory();
    if (photoEditorState.ratio === "free" && (cropCoversFullImage(photoEditorState) || !cropContainsPoint(photoEditorState.crop, point))) {
      photoEditorState.pointerMode = "select-crop";
      photoEditorState.cropStart = point;
      updateFreeCrop(point, point);
    } else {
      photoEditorState.pointerMode = "move-crop";
      photoEditorState.lastPoint = point;
      if (!cropContainsPoint(photoEditorState.crop, point)) {
        moveCropBy(point.x - (photoEditorState.crop.x + photoEditorState.crop.width / 2), point.y - (photoEditorState.crop.y + photoEditorState.crop.height / 2));
      }
    }
    renderPhotoEditor();
  }

  function handlePhotoEditorPointerMove(event) {
    if (!photoEditorState || !photoEditorState.pointerMode) {
      return;
    }
    event.preventDefault();
    const point = imagePointFromEvent(event);
    if (photoEditorState.pointerMode === "draw") {
      if (photoEditorState.tool === "arrow") {
        drawContinuousEditorLine(photoEditorState.lastPoint, point);
        photoEditorState.arrowPoints.push(point);
        photoEditorState.lastPoint = point;
        renderPhotoEditor();
      } else {
        drawContinuousEditorLine(photoEditorState.lastPoint, point);
        photoEditorState.lastPoint = point;
        renderPhotoEditor();
      }
      return;
    } else if (photoEditorState.pointerMode === "select-crop") {
      updateFreeCrop(photoEditorState.cropStart, point);
    } else if (photoEditorState.pointerMode === "move-crop") {
      moveCropBy(point.x - photoEditorState.lastPoint.x, point.y - photoEditorState.lastPoint.y);
      photoEditorState.lastPoint = point;
    }
    renderPhotoEditor();
  }

  function handlePhotoEditorPointerUp(event) {
    if (!photoEditorState || !photoEditorCanvas) {
      return;
    }
    const point = imagePointFromEvent(event);
    if (photoEditorState.pointerMode === "draw" && photoEditorState.tool === "arrow" && photoEditorState.arrowPoints.length) {
      const from = arrowHeadDirectionPoint(photoEditorState.arrowPoints, point);
      drawEditorArrow(from, point);
      photoEditorState.previewPoint = null;
      renderPhotoEditor();
    }
    photoEditorCanvas.releasePointerCapture(event.pointerId);
    photoEditorState.pointerMode = "";
    photoEditorState.pointerStart = null;
    photoEditorState.previewPoint = null;
    photoEditorState.arrowPoints = [];
    photoEditorState.lastPoint = null;
    photoEditorState.cropStart = null;
  }

  function canvasToBlob(canvas, type, quality) {
    return new Promise(function (resolve) {
      canvas.toBlob(resolve, type, quality);
    });
  }

  function editedFileName(file, type) {
    const extension = type === "image/png" ? "png" : "jpg";
    const base = String(file.name || "image").replace(/\.[^.]+$/, "");
    return base + "-edited." + extension;
  }

  async function applyPhotoEditor() {
    const state = photoEditorState;
    if (!state) {
      return;
    }
    const crop = {
      x: Math.round(state.crop.x),
      y: Math.round(state.crop.y),
      width: Math.max(1, Math.round(state.crop.width)),
      height: Math.max(1, Math.round(state.crop.height))
    };
    const output = document.createElement("canvas");
    output.width = crop.width;
    output.height = crop.height;
    const ctx = output.getContext("2d");
    ctx.drawImage(state.image, crop.x, crop.y, crop.width, crop.height, 0, 0, crop.width, crop.height);
    ctx.drawImage(state.annotationCanvas, crop.x, crop.y, crop.width, crop.height, 0, 0, crop.width, crop.height);
    if (state.markerCanvas) {
      ctx.save();
      ctx.globalAlpha = 0.34;
      ctx.drawImage(state.markerCanvas, crop.x, crop.y, crop.width, crop.height, 0, 0, crop.width, crop.height);
      ctx.restore();
    }
    const type = state.file.type === "image/png" ? "image/png" : "image/jpeg";
    const blob = await canvasToBlob(output, type, 0.92);
    if (!blob) {
      setMessageStatus("圖片編輯套用失敗。", true);
      return;
    }
    selectedAttachmentFiles[state.fileIndex] = new File([blob], editedFileName(state.file, type), {
      type: type,
      lastModified: Date.now()
    });
    closePhotoEditor();
    openAttachmentDialog(selectedAttachmentFiles, "photo");
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
    updateGroupInfoAction();
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

  async function handleConversationMembersUpdated(conversationID) {
    const updatedConversationID = Number(conversationID || 0);
    if (updatedConversationID) {
      mentionMembersCache.delete(updatedConversationID);
    }
    const wasActive = updatedConversationID && updatedConversationID === activeConversationID;
    const items = await fetchConversationItems();
    if (!items) {
      return;
    }
    renderConversationList(items);
    const stillVisible = !updatedConversationID || items.some(function (item) {
      return Number(item.conversation_id || 0) === updatedConversationID;
    });
    if (wasActive && !stillVisible) {
      closeActiveConversation();
      setMessageStatus("你已不在此群組。", false);
      return;
    }
    if (wasActive) {
      await loadMessages(activeConversationID, false);
    }
    if (wasActive && groupInfoPanel && !groupInfoPanel.hidden) {
      loadGroupInfo();
    }
  }

  async function loadMessages(conversationID, refreshListAfterRead, options) {
    options = options || {};
    const headers = authHeaders();
    if (!headers || !conversationID) {
      return;
    }

    const scrollAnchor = options.preserveScroll ? captureMessageScrollAnchor() : null;
    const requestToken = ++messageLoadToken;
    setConversationUIActive(true);
    setMessageStatus("正在載入訊息...", false);
    setConversationTitle(activeConversationTitle());
    if (!options.preserveScroll) {
      renderLoadingConversationState();
    }
    const response = await fetch("/api/conversations/" + conversationID + "/messages", {
      headers: { Authorization: headers.Authorization }
    });
    const result = await parseJSON(response);
    if (requestToken !== messageLoadToken || conversationID !== activeConversationID) {
      return;
    }
    if (!response.ok || !result.success) {
      setMessageStatus(result.message || "訊息讀取失敗。", true);
      if (result.code === "CONVERSATION_NOT_FOUND") {
        closeActiveConversation();
        setMessageStatus("你已不在此群組。", false);
      }
      return;
    }

    const data = result.data || {};
    setConversationTitle(data.title || activeConversationTitle() || "目前對話");
    renderMessages(data, {
      forceBottom: Boolean(options.forceBottom),
      preserveScroll: Boolean(options.preserveScroll),
      scrollAnchor: scrollAnchor,
      stickToBottom: Boolean(options.stickToBottom)
    });
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

    if (event.event_type === "conversation.members.updated") {
      handleConversationMembersUpdated(event.conversation_id);
      return;
    }

    if (event.event_type === "message.created") {
      refreshConversationListOnly();
      if (event.conversation_id && Number(event.conversation_id) === activeConversationID) {
        const stickToBottom = isMessageBoardNearBottom();
        loadMessages(activeConversationID, false, {
          stickToBottom: stickToBottom,
          preserveScroll: !stickToBottom
        });
      }
      return;
    }

    if (event.event_type === "message.recalled" || event.event_type === "message.deleted") {
      refreshConversationListOnly();
      if (event.conversation_id && Number(event.conversation_id) === activeConversationID) {
        loadMessages(activeConversationID, false, { preserveScroll: true });
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
    lastReportedReadMessageID = 0;
    pendingReadMessageID = 0;
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
    clearMentionMenu();
    if (composerFileInput) {
      composerFileInput.value = "";
    }
    selectedAttachmentFiles = [];
    setReplyMessage(null);
    updateSelectedFileState();
    closeAttachmentDialog(false);
    await loadMessages(activeConversationID, false, { forceBottom: true });
    await refreshConversationListOnly();
    setMessageStatus("訊息已送出。", false);
  }

  function handleComposerKeydown(event) {
    if (!composerMentionMenu.hidden) {
      if (event.key === "ArrowDown") {
        event.preventDefault();
        mentionMenuActiveIndex = (mentionMenuActiveIndex + 1) % mentionMenuItems.length;
        renderMentionMenu(mentionMenuItems);
        return;
      }
      if (event.key === "ArrowUp") {
        event.preventDefault();
        mentionMenuActiveIndex = (mentionMenuActiveIndex - 1 + mentionMenuItems.length) % mentionMenuItems.length;
        renderMentionMenu(mentionMenuItems);
        return;
      }
      if ((event.key === "Enter" || event.key === "Tab") && !event.isComposing) {
        event.preventDefault();
        insertMentionItem(mentionMenuItems[mentionMenuActiveIndex]);
        return;
      }
      if (event.key === "Escape") {
        event.preventDefault();
        clearMentionMenu();
        return;
      }
    }

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

  if (groupInfoOpenButton) {
    groupInfoOpenButton.addEventListener("click", function () {
      if (!activeGroupConversation()) {
        return;
      }
      setContactMenuOpen(false);
      setContactPanelOpen(false);
      setGroupInfoPanelOpen(true);
      loadGroupInfo();
    });
  }

  if (groupInfoClose) {
    groupInfoClose.addEventListener("click", function () {
      if (groupInfoMode !== "info") {
        setGroupInfoMode("info");
        return;
      }
      setGroupInfoPanelOpen(false);
    });
  }

  if (groupInfoEditButton) {
    groupInfoEditButton.addEventListener("click", openGroupEditPanel);
  }

  if (groupInfoAddButton) {
    groupInfoAddButton.addEventListener("click", openGroupAddMembers);
  }

  if (groupEditNameInput) {
    groupEditNameInput.addEventListener("input", updateGroupEditSubmitState);
  }

  if (groupEditDescriptionInput) {
    groupEditDescriptionInput.addEventListener("input", updateGroupEditSubmitState);
  }

  if (groupEditSubmitButton) {
    groupEditSubmitButton.addEventListener("click", saveGroupEdit);
  }

  if (groupAddSearchInput) {
    groupAddSearchInput.addEventListener("input", function () {
      groupAddSearchQuery = groupAddSearchInput.value.trim();
      renderGroupAddResults(contactItems, !contactsLoaded);
    });
  }

  if (groupAddResultsNode) {
    groupAddResultsNode.addEventListener("click", function (event) {
      const memberButton = event.target.closest("[data-group-add-id]");
      if (!memberButton) {
        return;
      }
      const member = {
        sourceSystem: memberButton.dataset.groupAddSource || readSourceSystem(),
        externalUserID: memberButton.dataset.groupAddId || "",
        title: memberButton.dataset.groupAddTitle || ""
      };
      updateGroupAddSelectedMember(member, !isGroupAddMemberSelected(member));
      renderGroupAddResults(contactItems, !contactsLoaded);
    });
  }

  if (groupAddSelectedNode) {
    groupAddSelectedNode.addEventListener("click", function (event) {
      const removeButton = event.target.closest("button");
      const chip = event.target.closest("[data-group-add-selected-key]");
      if (!removeButton || !chip) {
        return;
      }
      groupAddSelectedMembers = groupAddSelectedMembers.filter(function (member) {
        return memberKey(member) !== chip.dataset.groupAddSelectedKey;
      });
      renderGroupAddSelectedMembers();
      renderGroupAddResults(contactItems, !contactsLoaded);
    });
  }

  if (groupAddSubmitButton) {
    groupAddSubmitButton.addEventListener("click", addSelectedMembersToActiveGroup);
  }

  if (groupInfoMembers) {
    groupInfoMembers.addEventListener("click", function (event) {
      const memberButton = event.target.closest("[data-group-member-external]");
      if (!memberButton || !groupInfoMembers.contains(memberButton)) {
        return;
      }
      setGroupMemberMenuOpen(false);
      openGroupMemberConversation(memberButton);
    });
    groupInfoMembers.addEventListener("contextmenu", function (event) {
      const memberButton = event.target.closest("[data-group-member-external]");
      if (!memberButton || !groupInfoMembers.contains(memberButton)) {
        return;
      }
      event.preventDefault();
      setGroupMemberMenuOpen(true, event.clientX + 8, event.clientY + 8, {
        sourceSystem: memberButton.dataset.groupMemberSource || readSourceSystem(),
        externalUserID: memberButton.dataset.groupMemberExternal || "",
        title: memberButton.dataset.groupMemberTitle || memberButton.dataset.groupMemberExternal || "成員",
        role: memberButton.dataset.groupMemberRole || "",
        isSelf: memberButton.dataset.groupMemberSelf === "1"
      });
    });
  }

  groupMemberMenu.addEventListener("click", function (event) {
    const actionButton = event.target.closest("[data-group-member-action]");
    const target = groupMemberMenuTarget;
    if (!actionButton || !target) {
      return;
    }
    const action = actionButton.dataset.groupMemberAction;
    setGroupMemberMenuOpen(false);
    if (action === "message") {
      openGroupMemberDirect(target);
    }
    if (action === "remove") {
      removeGroupMember(target);
    }
  });

  document.addEventListener("click", function (event) {
    if (!groupMemberMenu.hidden && !groupMemberMenu.contains(event.target)) {
      setGroupMemberMenuOpen(false);
    }
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape") {
      setGroupMemberMenuOpen(false);
    }
  });

  if (mobileChatBackButton) {
    mobileChatBackButton.addEventListener("click", function () {
      app.classList.remove("is-mobile-chat-open");
      setContactMenuOpen(false);
      setContactPanelOpen(false);
      setGroupInfoPanelOpen(false);
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
        setGroupInfoPanelOpen(false);
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
      renderNewGroupDefaultContacts();
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
  composerMentionMenu.addEventListener("mousedown", function (event) {
    event.preventDefault();
  });
  composerMentionMenu.addEventListener("click", function (event) {
    const button = event.target.closest("[data-mention-index]");
    if (!button) {
      return;
    }
    insertMentionItem(mentionMenuItems[Number(button.dataset.mentionIndex || 0)]);
  });
  composerInput.addEventListener("input", function () {
    syncComposerInputHeight();
    updateMentionMenu();
  });
  composerInput.addEventListener("click", updateMentionMenu);
  composerInput.addEventListener("blur", function () {
    window.setTimeout(clearMentionMenu, 120);
  });
  composerInput.addEventListener("keydown", handleComposerKeydown);
  messageBoard.addEventListener("contextmenu", function (event) {
    const bubble = bubbleFromContextEvent(event);
    if (!bubble) {
      return;
    }
    openMessageContextMenu(event, bubble);
  }, true);
  messageBoard.addEventListener("scroll", function () {
    closeMessageContextMenu();
    scheduleVisibleReadUpdate();
  }, { passive: true });
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
    if (isPhotoEditorOpen()) {
      const key = String(event.key || "").toLowerCase();
      if ((event.metaKey || event.ctrlKey) && key === "z" && !event.shiftKey) {
        event.preventDefault();
        undoPhotoEditorChange();
        return;
      }
      if ((event.metaKey || event.ctrlKey) && (key === "y" || key === "z" && event.shiftKey)) {
        event.preventDefault();
        redoPhotoEditorChange();
        return;
      }
      if (event.key === "Escape") {
        event.preventDefault();
        if (photoEditorConfirm && !photoEditorConfirm.hidden) {
          hidePhotoEditorConfirm();
        } else {
          requestClosePhotoEditor();
        }
        return;
      }
    }
	    if (event.key === "Escape") {
      closeMessageContextMenu();
      setConversationCreateMenuOpen(false);
      setContactMenuOpen(false);
      setContactPanelOpen(false);
      setGroupInfoPanelOpen(false);
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
      if (isPhotoEditorOpen()) {
        return;
      }
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
  if (attachmentPreview) {
    attachmentPreview.addEventListener("click", function (event) {
      const button = event.target.closest("[data-photo-edit-index]");
      if (!button) {
        return;
      }
      openPhotoEditor(Number(button.dataset.photoEditIndex || 0));
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
  if (photoEditorClose) {
    photoEditorClose.addEventListener("click", requestClosePhotoEditor);
  }
  if (photoEditorUndo) {
    photoEditorUndo.addEventListener("click", undoPhotoEditorChange);
  }
  if (photoEditorRedo) {
    photoEditorRedo.addEventListener("click", redoPhotoEditorChange);
  }
  if (photoEditorConfirmCancel) {
    photoEditorConfirmCancel.addEventListener("click", hidePhotoEditorConfirm);
  }
  if (photoEditorConfirmDiscard) {
    photoEditorConfirmDiscard.addEventListener("click", closePhotoEditor);
  }
  if (photoEditorApply) {
    photoEditorApply.addEventListener("click", applyPhotoEditor);
  }
  if (photoEditorReset) {
    photoEditorReset.addEventListener("click", resetPhotoEditor);
  }
  if (photoEditorCanvas) {
    photoEditorCanvas.addEventListener("pointerdown", handlePhotoEditorPointerDown);
    photoEditorCanvas.addEventListener("pointermove", handlePhotoEditorPointerMove);
    photoEditorCanvas.addEventListener("pointerup", handlePhotoEditorPointerUp);
    photoEditorCanvas.addEventListener("pointercancel", handlePhotoEditorPointerUp);
  }
  photoEditorModeButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      setPhotoEditorMode(button.dataset.photoEditorMode || "crop");
    });
  });
  photoRatioButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      pushPhotoEditorHistory();
      photoRatioButtons.forEach(function (item) {
        item.classList.toggle("is-active", item === button);
      });
      setPhotoEditorCropForRatio(button.dataset.photoRatio || "free");
    });
  });
  photoColorButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      if (!photoEditorState) {
        return;
      }
      if (photoCustomColorPanel) {
        photoCustomColorPanel.hidden = true;
      }
      setPhotoEditorColor(button.dataset.photoColor || "#ff4b40");
    });
  });
  if (photoCustomColorButton) {
    photoCustomColorButton.addEventListener("click", function () {
      if (!photoEditorState) {
        return;
      }
      if (photoCustomColorPanel) {
        photoCustomColorPanel.hidden = !photoCustomColorPanel.hidden;
      }
      const color = photoColorHexInput && hexToRGB(photoColorHexInput.value) ? photoColorHexInput.value : "#ffd21f";
      updateCustomColorUI(color);
      setPhotoEditorColor(color, { custom: true });
    });
  }
  if (photoColorHueInput) {
    photoColorHueInput.addEventListener("input", function () {
      const saturation = Number(photoColorSV ? photoColorSV.style.getPropertyValue("--photo-color-saturation") : 100) / 100 || 1;
      const value = Number(photoColorSV ? photoColorSV.style.getPropertyValue("--photo-color-value") : 100) / 100 || 1;
      applyCustomHSV(saturation, value);
    });
  }
  if (photoColorSV) {
    photoColorSV.addEventListener("pointerdown", function (event) {
      event.preventDefault();
      photoColorSV.setPointerCapture(event.pointerId);
      const updateFromEvent = function (pointerEvent) {
        const rect = photoColorSV.getBoundingClientRect();
        const saturation = Math.min(Math.max((pointerEvent.clientX - rect.left) / rect.width, 0), 1);
        const value = 1 - Math.min(Math.max((pointerEvent.clientY - rect.top) / rect.height, 0), 1);
        applyCustomHSV(saturation, value);
      };
      updateFromEvent(event);
      const handleMove = function (moveEvent) {
        updateFromEvent(moveEvent);
      };
      const handleUp = function (upEvent) {
        updateFromEvent(upEvent);
        photoColorSV.releasePointerCapture(upEvent.pointerId);
        photoColorSV.removeEventListener("pointermove", handleMove);
        photoColorSV.removeEventListener("pointerup", handleUp);
        photoColorSV.removeEventListener("pointercancel", handleUp);
      };
      photoColorSV.addEventListener("pointermove", handleMove);
      photoColorSV.addEventListener("pointerup", handleUp);
      photoColorSV.addEventListener("pointercancel", handleUp);
    });
  }
  if (photoColorHexInput) {
    photoColorHexInput.addEventListener("change", function () {
      const rgb = hexToRGB(photoColorHexInput.value);
      if (!rgb) {
        updateCustomColorUI(photoEditorState ? photoEditorState.color : "#ffd21f");
        return;
      }
      const color = rgbToHex(rgb);
      updateCustomColorUI(color);
      setPhotoEditorColor(color, { custom: true });
    });
  }
  photoToolButtons.forEach(function (button) {
    button.addEventListener("click", function () {
      if (!photoEditorState) {
        return;
      }
      const tool = button.dataset.photoTool || "pen";
      photoToolButtons.forEach(function (item) {
        item.classList.toggle("is-active", item === button);
      });
      photoEditorState.tool = tool;
      photoEditorState.color = photoEditorState.toolColors && photoEditorState.toolColors[tool]
        ? photoEditorState.toolColors[tool]
        : photoToolDefaultColors[tool] || photoEditorState.color || photoToolDefaultColors.pen;
      setPhotoEditorColor(photoEditorState.color);
    });
  });
  if (photoBrushSizeInput) {
    updatePhotoBrushRange();
    photoBrushSizeInput.addEventListener("input", function () {
      if (photoEditorState) {
        photoEditorState.brushSize = Number(photoBrushSizeInput.value || 18);
      }
      updatePhotoBrushRange();
    });
  }
  if (chatShell && chatDropOverlay) {
    chatShell.addEventListener("paste", handlePasteUpload);
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
