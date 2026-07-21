/**
 * Admin Dashboard Application
 * Main application logic for admin management interface
 */

(function() {
  'use strict';

  const { getSessionToken, getUserId, isAuthenticated, logout, makeAuthenticatedRequest, STORAGE_KEYS } = window.adminAuth;

  // DOM Elements
  let elements = {};
  let editingUserID = null;
  let editingAdminUserID = null;
  let originalEditableUser = null;
  let originalEditableAdmin = null;
  let usersPage = 1;
  let usersPerPage = 10;
  let userInvitationsPage = 1;
  let userInvitationsPerPage = 10;
  let bulkInvitePrecheck = null;
  let adminConversationsPage = 1;
  let adminConversationsPerPage = 10;
  let activeAdminConversationID = null;
  let adminConversationMessagesPage = 1;
  let adminConversationMessagesPerPage = 20;
  let adminsPage = 1;
  let adminsPerPage = 10;
  let modalConfirmHandler = null;

  // Adjust each user-table column here. Use width for a fixed width or
  // minWidth when the column may grow with the available table width.
  const USER_COLUMNS = [
    { key: 'index', title: '#', width: 60 },
    { key: 'avatar', title: '頭像', width: 90 },
    { key: 'account', title: '帳號', minWidth: 160 },
    { key: 'displayName', title: '暱稱', minWidth: 180 },
    { key: 'email', title: 'Email', minWidth: 260 },
    { key: 'status', title: '狀態', width: 100 },
    { key: 'source', title: '來源', width: 100 },
    { key: 'lastOnline', title: '最近在線', minWidth: 180 },
    { key: 'options', title: '選項', width: 260 }
  ];

  const VIEW_ROUTES = {
    dashboard: '/office',
    users: '/office/user',
    userCreate: '/office/user/create',
    userInvitations: '/office/user/invitations',
    userEdit: null,
    conversations: '/office/conversations',
    conversationDetail: null,
    admins: '/office/admins',
    adminCreate: '/office/admins/create',
    adminEdit: null
  };

  const ROUTE_VIEWS = Object.entries(VIEW_ROUTES).reduce((routes, [viewName, path]) => {
    if (path) routes[path] = viewName;
    return routes;
  }, {
    '/admin/dashboard': 'dashboard',
    '/office/user/invite': 'userInvitations'
  });

  function resolveInitialView() {
    const editMatch = window.location.pathname.match(/^\/office\/user\/(\d+)\/edit$/);
    if (editMatch) {
      editingUserID = editMatch[1];
      return 'userEdit';
    }
    const adminEditMatch = window.location.pathname.match(/^\/office\/admins\/(\d+)\/edit$/);
    if (adminEditMatch) {
      editingAdminUserID = adminEditMatch[1];
      return 'adminEdit';
    }
    const conversationDetailMatch = window.location.pathname.match(/^\/office\/conversations\/([^/]+)$/);
    if (conversationDetailMatch) {
      activeAdminConversationID = decodeURIComponent(conversationDetailMatch[1]);
      return 'conversationDetail';
    }
    return ROUTE_VIEWS[window.location.pathname] || 'dashboard';
  }

  /**
   * Initialize the app
   */
  async function initApp() {
    if (!isAuthenticated()) {
      return;
    }

    cacheElements();
    renderTableColumns(elements.usersTable, elements.usersColgroup, elements.usersThead, USER_COLUMNS);
    setupOverlayScrollbar(document.getElementById('users-table-scroll'));
    setupEventListeners();
    setupDateInputPickers();
    switchView(resolveInitialView(), { replace: true });
    await loadInitialData();
  }

  /**
   * Cache DOM elements for performance
   */
  function cacheElements() {
    elements = {
      // Navigation
      appShell: document.querySelector('[data-admin-app]'),
      headerToggle: document.querySelector('.header-toggle'),
      sidebarBackdrop: document.querySelector('[data-sidebar-backdrop]'),
      menu: document.querySelector('[data-sidebar-menu]'),
      mainContent: document.querySelector('[data-main-content]'),
      views: document.querySelectorAll('[data-view]'),
      
      // Dashboard stats
      statUsers: document.getElementById('stat-users'),
      statActive: document.getElementById('stat-active'),
      statConversations: document.getElementById('stat-conversations'),
      statUptime: document.getElementById('stat-uptime'),
      
      // Users
      userFilterForm: document.getElementById('user-filter-form'),
      userSearch: document.getElementById('user-search'),
      userDisplayName: document.getElementById('user-display-name'),
      userEmail: document.getElementById('user-email'),
      userStatus: document.getElementById('user-status'),
      userSource: document.getElementById('user-source'),
      userSort: document.getElementById('user-sort'),
      userCreatedFrom: document.getElementById('user-created-from'),
      userCreatedTo: document.getElementById('user-created-to'),
      usersTable: document.getElementById('users-table'),
      usersColgroup: document.getElementById('users-colgroup'),
      usersThead: document.getElementById('users-thead'),
      usersTbody: document.getElementById('users-tbody'),
      usersLoading: document.getElementById('users-loading'),
      usersEmpty: document.getElementById('users-empty'),
      usersPagination: document.getElementById('users-pagination'),
      usersTotal: document.getElementById('users-total'),
      usersPages: document.getElementById('users-pages'),
      usersPerPage: document.getElementById('users-per-page'),
      refreshUsersBtn: document.getElementById('refresh-users-btn'),
      userCreateForm: document.getElementById('user-create-form'),
      createUserAccount: document.getElementById('create-user-account'),
      createUserDisplayName: document.getElementById('create-user-display-name'),
      createUserEmail: document.getElementById('create-user-email'),
      createUserStatus: document.getElementById('create-user-status'),
      createUserPassword: document.getElementById('create-user-password'),
      submitCreateUser: document.getElementById('submit-create-user'),
      userInvitationsFilterForm: document.getElementById('user-invitations-filter-form'),
      userInvitationEmailFilter: document.getElementById('user-invitation-email-filter'),
      userInvitationStatusFilter: document.getElementById('user-invitation-status-filter'),
      userInvitationSort: document.getElementById('user-invitation-sort'),
      userInvitationsTable: document.getElementById('user-invitations-table'),
      userInvitationsTbody: document.getElementById('user-invitations-tbody'),
      userInvitationsLoading: document.getElementById('user-invitations-loading'),
      userInvitationsEmpty: document.getElementById('user-invitations-empty'),
      userInvitationsPagination: document.getElementById('user-invitations-pagination'),
      userInvitationsTotal: document.getElementById('user-invitations-total'),
      userInvitationsPages: document.getElementById('user-invitations-pages'),
      userInvitationsPerPage: document.getElementById('user-invitations-per-page'),
      refreshUserInvitationsBtn: document.getElementById('refresh-user-invitations-btn'),
      openUserInviteModal: document.getElementById('open-user-invite-modal'),
      openBulkUserInviteModal: document.getElementById('open-bulk-user-invite-modal'),
      userInviteModal: document.getElementById('user-invite-modal'),
      closeUserInviteModal: document.getElementById('close-user-invite-modal'),
      userInviteForm: document.getElementById('user-invite-form'),
      inviteUserEmail: document.getElementById('invite-user-email'),
      submitInviteUser: document.getElementById('submit-invite-user'),
      bulkUserInviteModal: document.getElementById('bulk-user-invite-modal'),
      closeBulkUserInviteModal: document.getElementById('close-bulk-user-invite-modal'),
      bulkUserInviteForm: document.getElementById('bulk-user-invite-form'),
      bulkInviteFile: document.getElementById('bulk-invite-file'),
      bulkInviteSummary: document.getElementById('bulk-invite-summary'),
      bulkInviteResults: document.getElementById('bulk-invite-results'),
      bulkInviteStatus: document.getElementById('bulk-invite-status'),
      submitBulkInvite: document.getElementById('submit-bulk-invite'),
      resetBulkInvite: document.getElementById('reset-bulk-invite'),
      userEditForm: document.getElementById('user-edit-form'),
      editUserAccount: document.getElementById('edit-user-account'),
      editUserPassword: document.getElementById('edit-user-password'),
      editUserDisplayName: document.getElementById('edit-user-display-name'),
      editUserEmail: document.getElementById('edit-user-email'),
      editUserStatus: document.getElementById('edit-user-status'),
      submitEditUser: document.getElementById('submit-edit-user'),
      resetEditUser: document.getElementById('reset-edit-user'),

      // Conversations
      adminConversationFilterForm: document.getElementById('admin-conversation-filter-form'),
      adminConversationType: document.getElementById('admin-conversation-type'),
      adminConversationKeyword: document.getElementById('admin-conversation-keyword'),
      adminConversationLastFrom: document.getElementById('admin-conversation-last-from'),
      adminConversationLastTo: document.getElementById('admin-conversation-last-to'),
      adminConversationsTable: document.getElementById('admin-conversations-table'),
      adminConversationsTbody: document.getElementById('admin-conversations-tbody'),
      adminConversationsLoading: document.getElementById('admin-conversations-loading'),
      adminConversationsEmpty: document.getElementById('admin-conversations-empty'),
      adminConversationsPagination: document.getElementById('admin-conversations-pagination'),
      adminConversationsTotal: document.getElementById('admin-conversations-total'),
      adminConversationsPages: document.getElementById('admin-conversations-pages'),
      adminConversationsPerPage: document.getElementById('admin-conversations-per-page'),
      refreshAdminConversationsBtn: document.getElementById('refresh-admin-conversations-btn'),
      resetAdminConversationsBtn: document.getElementById('reset-admin-conversations-btn'),
      adminConversationBreadcrumbCurrent: document.getElementById('admin-conversation-breadcrumb-current'),
      adminConversationDetailLoading: document.getElementById('admin-conversation-detail-loading'),
      adminConversationDetailSummary: document.getElementById('admin-conversation-detail-summary'),
      adminConversationDetailTitle: document.getElementById('admin-conversation-detail-title'),
      adminConversationDetailType: document.getElementById('admin-conversation-detail-type'),
      adminConversationDetailID: document.getElementById('admin-conversation-detail-id'),
      adminConversationDetailName: document.getElementById('admin-conversation-detail-name'),
      adminConversationDetailCreatedAt: document.getElementById('admin-conversation-detail-created-at'),
      adminConversationDetailLastAt: document.getElementById('admin-conversation-detail-last-at'),
      adminConversationDirectUsers: document.getElementById('admin-conversation-direct-users'),
      adminConversationDetailParticipants: document.getElementById('admin-conversation-detail-participants'),
      adminConversationGroupName: document.getElementById('admin-conversation-group-name'),
      adminConversationDetailGroupName: document.getElementById('admin-conversation-detail-group-name'),
      adminConversationGroupCreator: document.getElementById('admin-conversation-group-creator'),
      adminConversationDetailCreator: document.getElementById('admin-conversation-detail-creator'),
      adminConversationGroupMembers: document.getElementById('admin-conversation-group-members'),
      adminConversationDetailMemberCount: document.getElementById('admin-conversation-detail-member-count'),
      adminConversationMembers: document.getElementById('admin-conversation-members'),
      adminConversationMembersToggle: document.getElementById('admin-conversation-members-toggle'),
      adminConversationMembersList: document.getElementById('admin-conversation-members-list'),
      adminConversationMessageFilterForm: document.getElementById('admin-conversation-message-filter-form'),
      adminConversationMessageKeyword: document.getElementById('admin-conversation-message-keyword'),
      adminConversationMessageSender: document.getElementById('admin-conversation-message-sender'),
      adminConversationMessageType: document.getElementById('admin-conversation-message-type'),
      adminConversationMessageMention: document.getElementById('admin-conversation-message-mention'),
      adminConversationMessageFrom: document.getElementById('admin-conversation-message-from'),
      adminConversationMessageTo: document.getElementById('admin-conversation-message-to'),
      adminConversationMessagesTable: document.getElementById('admin-conversation-messages-table'),
      adminConversationMessagesTbody: document.getElementById('admin-conversation-messages-tbody'),
      adminConversationMessagesEmpty: document.getElementById('admin-conversation-messages-empty'),
      adminConversationMessagesPagination: document.getElementById('admin-conversation-messages-pagination'),
      adminConversationMessagesTotal: document.getElementById('admin-conversation-messages-total'),
      adminConversationMessagesPages: document.getElementById('admin-conversation-messages-pages'),
      adminConversationMessagesPerPage: document.getElementById('admin-conversation-messages-per-page'),
      refreshAdminConversationMessagesBtn: document.getElementById('refresh-admin-conversation-messages-btn'),
      resetAdminConversationMessagesBtn: document.getElementById('reset-admin-conversation-messages-btn'),

      // Admins
      adminFilterForm: document.getElementById('admin-filter-form'),
      adminSearch: document.getElementById('admin-search'),
      adminDisplayName: document.getElementById('admin-display-name'),
      adminStatus: document.getElementById('admin-status'),
      adminsTable: document.getElementById('admins-table'),
      adminsTbody: document.getElementById('admins-tbody'),
      adminsLoading: document.getElementById('admins-loading'),
      adminsEmpty: document.getElementById('admins-empty'),
      adminsPagination: document.getElementById('admins-pagination'),
      adminsTotal: document.getElementById('admins-total'),
      adminsPages: document.getElementById('admins-pages'),
      adminsPerPage: document.getElementById('admins-per-page'),
      refreshAdminsBtn: document.getElementById('refresh-admins-btn'),
      adminCreateForm: document.getElementById('admin-create-form'),
      createAdminAccount: document.getElementById('create-admin-account'),
      createAdminPassword: document.getElementById('create-admin-password'),
      createAdminRole: document.getElementById('create-admin-role'),
      createAdminDisplayName: document.getElementById('create-admin-display-name'),
      createAdminStatus: document.getElementById('create-admin-status'),
      submitCreateAdmin: document.getElementById('submit-create-admin'),
      adminEditForm: document.getElementById('admin-edit-form'),
      editAdminAccount: document.getElementById('edit-admin-account'),
      editAdminPassword: document.getElementById('edit-admin-password'),
      editAdminDisplayName: document.getElementById('edit-admin-display-name'),
      editAdminStatus: document.getElementById('edit-admin-status'),
      submitEditAdmin: document.getElementById('submit-edit-admin'),
      resetEditAdmin: document.getElementById('reset-edit-admin'),
      
      // Devices
      deviceUserId: document.getElementById('device-user-id'),
      devicesTable: document.getElementById('devices-table'),
      devicesTbody: document.getElementById('devices-tbody'),
      devicesLoading: document.getElementById('devices-loading'),
      devicesEmpty: document.getElementById('devices-empty'),
      searchDevicesBtn: document.getElementById('search-devices-btn'),
      refreshDevicesBtn: document.getElementById('refresh-devices-btn'),
      
      // Whitelist
      whitelistUserId: document.getElementById('whitelist-user-id'),
      whitelistLoading: document.getElementById('whitelist-loading'),
      whitelistContent: document.getElementById('whitelist-content'),
      allowAllToggle: document.getElementById('allow-all-toggle'),
      rulesContainer: document.getElementById('rules-container'),
      rulesList: document.getElementById('rules-list'),
      addRuleBtn: document.getElementById('add-rule-btn'),
      loadWhitelistBtn: document.getElementById('load-whitelist-btn'),
      saveWhitelistBtn: document.getElementById('save-whitelist-btn'),
      cancelWhitelistBtn: document.getElementById('cancel-whitelist-btn'),
      
      // Logs
      logFilter: document.getElementById('log-filter'),
      logsList: document.getElementById('logs-list'),
      logsEmpty: document.getElementById('logs-empty'),
      refreshLogsBtn: document.getElementById('refresh-logs-btn'),
      
      // Health
      healthApiTime: document.getElementById('health-api-time'),
      healthDbTime: document.getElementById('health-db-time'),
      healthWsConnections: document.getElementById('health-ws-connections'),
      healthVersion: document.getElementById('health-version'),
      healthUptime: document.getElementById('health-uptime'),
      healthMemory: document.getElementById('health-memory'),
      healthConcurrent: document.getElementById('health-concurrent'),
      
      // Modal
      modal: document.getElementById('admin-modal'),
      modalTitle: document.getElementById('modal-title'),
      modalBody: document.getElementById('modal-body'),
      modalConfirmBtn: document.getElementById('modal-confirm-btn'),
      modalCancelBtn: document.getElementById('modal-cancel-btn'),

      // Header account menu
      adminAccountMenu: document.querySelector('[data-admin-account-menu]'),
      adminAccountTrigger: document.getElementById('admin-account-trigger'),
      adminAccountDropdown: document.querySelector('[data-admin-account-dropdown]'),
      adminProfileAction: document.querySelector('[data-admin-profile-action]'),
      
      // Logout
      logoutButton: document.getElementById('logout-button')
    };
  }

  /**
   * Setup event listeners
   */
  function setupEventListeners() {
    // Menu items
    if (elements.headerToggle) {
      elements.headerToggle.addEventListener('click', toggleAdminSidebar);
    }
    if (elements.sidebarBackdrop) {
      elements.sidebarBackdrop.addEventListener('click', closeAdminSidebar);
    }

    if (elements.menu) {
      elements.menu.addEventListener('click', (e) => {
        const groupToggle = e.target.closest('[data-sidebar-toggle]');
        if (groupToggle) {
          const group = groupToggle.closest('.sidebar-group');
          if (group) {
            setSidebarGroupOpen(group, !group.classList.contains('is-open'));
          }
          return;
        }

        const menuItem = e.target.closest('[data-menu-item]');
        if (menuItem) {
          switchView(menuItem.getAttribute('data-menu-item'));
          closeAdminSidebar();
        }
      });
    }

    if (elements.mainContent) {
      elements.mainContent.addEventListener('click', (e) => {
        const menuItem = e.target.closest('[data-menu-item]');
        if (menuItem) {
          switchView(menuItem.getAttribute('data-menu-item'));
        }
      });
    }

    window.addEventListener('popstate', () => {
      switchView(resolveInitialView(), { skipHistory: true });
      closeAdminSidebar();
    });

    if (elements.adminAccountTrigger) {
      elements.adminAccountTrigger.addEventListener('click', (e) => {
        e.preventDefault();
        toggleAdminAccountMenu();
      });
    }

    if (elements.adminProfileAction) {
      elements.adminProfileAction.addEventListener('click', () => {
        closeAdminAccountMenu();
        openCurrentAdminEdit();
      });
    }

    document.addEventListener('click', (e) => {
      if (!elements.adminAccountMenu || elements.adminAccountMenu.contains(e.target)) {
        return;
      }
      closeAdminAccountMenu();
    });

    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        closeAdminAccountMenu();
        closeAdminSidebar();
      }
    });

    // User management
    if (elements.userFilterForm) {
      elements.userFilterForm.addEventListener('submit', (e) => {
        e.preventDefault();
        usersPage = 1;
        loadUsers();
      });
    }
    if (elements.usersPerPage) {
      elements.usersPerPage.addEventListener('change', () => {
        usersPerPage = Number(elements.usersPerPage.value) || 10;
        usersPage = 1;
        loadUsers();
      });
    }
    if (elements.refreshUsersBtn && !elements.userFilterForm) {
      elements.refreshUsersBtn.addEventListener('click', loadUsers);
    }
    if (elements.userCreateForm) {
      elements.userCreateForm.addEventListener('submit', createUser);
    }
    if (elements.userInvitationsFilterForm) {
      elements.userInvitationsFilterForm.addEventListener('submit', (e) => {
        e.preventDefault();
        userInvitationsPage = 1;
        loadUserInvitations();
      });
    }
    if (elements.userInvitationsPerPage) {
      elements.userInvitationsPerPage.addEventListener('change', () => {
        userInvitationsPerPage = Number(elements.userInvitationsPerPage.value) || 10;
        userInvitationsPage = 1;
        loadUserInvitations();
      });
    }
    if (elements.openUserInviteModal) {
      elements.openUserInviteModal.addEventListener('click', openUserInviteModal);
    }
    if (elements.openBulkUserInviteModal) {
      elements.openBulkUserInviteModal.addEventListener('click', openBulkUserInviteModal);
    }
    if (elements.closeUserInviteModal) {
      elements.closeUserInviteModal.addEventListener('click', closeUserInviteModal);
    }
    if (elements.closeBulkUserInviteModal) {
      elements.closeBulkUserInviteModal.addEventListener('click', closeBulkUserInviteModal);
    }
    if (elements.userInviteModal) {
      elements.userInviteModal.addEventListener('click', (e) => {
        if (e.target === elements.userInviteModal) closeUserInviteModal();
      });
    }
    if (elements.bulkUserInviteModal) {
      elements.bulkUserInviteModal.addEventListener('click', (e) => {
        if (e.target === elements.bulkUserInviteModal) closeBulkUserInviteModal();
      });
    }
    if (elements.userInviteForm) {
      elements.userInviteForm.addEventListener('submit', inviteUser);
    }
    if (elements.bulkInviteFile) {
      elements.bulkInviteFile.addEventListener('change', precheckBulkUserInvitations);
    }
    if (elements.bulkUserInviteForm) {
      elements.bulkUserInviteForm.addEventListener('submit', sendBulkUserInvitations);
    }
    if (elements.resetBulkInvite) {
      elements.resetBulkInvite.addEventListener('click', resetBulkInviteModal);
    }
    if (elements.userEditForm) {
      elements.userEditForm.addEventListener('submit', updateUser);
    }
    if (elements.resetEditUser) {
      elements.resetEditUser.addEventListener('click', resetUserEditForm);
    }

    // Conversation records
    if (elements.adminConversationFilterForm) {
      elements.adminConversationFilterForm.addEventListener('submit', (e) => {
        e.preventDefault();
        adminConversationsPage = 1;
        loadAdminConversations();
      });
    }
    if (elements.adminConversationsPerPage) {
      elements.adminConversationsPerPage.addEventListener('change', () => {
        adminConversationsPerPage = Number(elements.adminConversationsPerPage.value) || 10;
        adminConversationsPage = 1;
        loadAdminConversations();
      });
    }
    if (elements.resetAdminConversationsBtn) {
      elements.resetAdminConversationsBtn.addEventListener('click', resetAdminConversationFilters);
    }
    if (elements.adminConversationMessageFilterForm) {
      elements.adminConversationMessageFilterForm.addEventListener('submit', (e) => {
        e.preventDefault();
        adminConversationMessagesPage = 1;
        loadAdminConversationDetail();
      });
    }
    if (elements.adminConversationMessagesPerPage) {
      elements.adminConversationMessagesPerPage.addEventListener('change', () => {
        adminConversationMessagesPerPage = Number(elements.adminConversationMessagesPerPage.value) || 20;
        adminConversationMessagesPage = 1;
        loadAdminConversationDetail();
      });
    }
    if (elements.resetAdminConversationMessagesBtn) {
      elements.resetAdminConversationMessagesBtn.addEventListener('click', resetAdminConversationMessageFilters);
    }
    if (elements.adminConversationMembersToggle) {
      elements.adminConversationMembersToggle.addEventListener('click', toggleAdminConversationMembers);
    }

    // Admin management
    if (elements.adminFilterForm) {
      elements.adminFilterForm.addEventListener('submit', (e) => {
        e.preventDefault();
        adminsPage = 1;
        loadAdmins();
      });
    }
    if (elements.adminsPerPage) {
      elements.adminsPerPage.addEventListener('change', () => {
        adminsPerPage = Number(elements.adminsPerPage.value) || 10;
        adminsPage = 1;
        loadAdmins();
      });
    }
    if (elements.adminCreateForm) {
      elements.adminCreateForm.addEventListener('submit', createAdmin);
    }
    if (elements.adminEditForm) {
      elements.adminEditForm.addEventListener('submit', updateAdmin);
    }
    if (elements.resetEditAdmin) {
      elements.resetEditAdmin.addEventListener('click', resetAdminEditForm);
    }

    // Devices management
    if (elements.searchDevicesBtn) {
      elements.searchDevicesBtn.addEventListener('click', searchDevices);
    }
    if (elements.deviceUserId) {
      elements.deviceUserId.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') searchDevices();
      });
    }
    if (elements.refreshDevicesBtn) {
      elements.refreshDevicesBtn.addEventListener('click', searchDevices);
    }

    // Whitelist management
    if (elements.loadWhitelistBtn) {
      elements.loadWhitelistBtn.addEventListener('click', loadWhitelist);
    }
    if (elements.whitelistUserId) {
      elements.whitelistUserId.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') loadWhitelist();
      });
    }
    if (elements.allowAllToggle) {
      elements.allowAllToggle.addEventListener('change', toggleRulesContainer);
    }
    if (elements.saveWhitelistBtn) {
      elements.saveWhitelistBtn.addEventListener('click', saveWhitelist);
    }
    if (elements.cancelWhitelistBtn) {
      elements.cancelWhitelistBtn.addEventListener('click', () => switchView('dashboard'));
    }

    // Logs
    if (elements.logFilter) {
      elements.logFilter.addEventListener('change', filterLogs);
    }
    if (elements.refreshLogsBtn) {
      elements.refreshLogsBtn.addEventListener('click', loadLogs);
    }

    // Modal
    if (elements.modalCancelBtn) {
      elements.modalCancelBtn.addEventListener('click', closeModal);
    }
    if (elements.modalConfirmBtn) {
      elements.modalConfirmBtn.addEventListener('click', async () => {
        if (!modalConfirmHandler) {
          closeModal();
          return;
        }

        elements.modalConfirmBtn.disabled = true;
        try {
          await modalConfirmHandler();
          closeModal();
        } catch (err) {
          // Keep the modal open so the admin can retry after the visible error.
        } finally {
          elements.modalConfirmBtn.disabled = false;
        }
      });
    }

    // Logout
    if (elements.logoutButton) {
      elements.logoutButton.addEventListener('click', (e) => {
        e.preventDefault();
        logout();
      });
    }
  }

  function setupDateInputPickers() {
    document.querySelectorAll('input[type="date"]').forEach((input) => {
      if (input.dataset.pickerTriggerBound === 'true') {
        return;
      }
      input.dataset.pickerTriggerBound = 'true';
      updateDateInputPlaceholder(input);
      input.addEventListener('input', () => updateDateInputPlaceholder(input));
      input.addEventListener('change', () => updateDateInputPlaceholder(input));
      input.addEventListener('click', () => {
        if (typeof input.showPicker !== 'function') {
          return;
        }
        try {
          input.showPicker();
        } catch (error) {
          // Some browsers throw when the native picker is already open.
        }
      });
    });
  }

  function updateDateInputPlaceholder(input) {
    const field = input.closest('label');
    if (!field) {
      return;
    }
    field.classList.toggle('date-field-empty', !input.value);
  }

  function toggleAdminSidebar() {
    if (!elements.appShell) return;
    setAdminSidebarOpen(!elements.appShell.classList.contains('is-sidebar-open'));
  }

  function closeAdminSidebar() {
    setAdminSidebarOpen(false);
  }

  function setAdminSidebarOpen(isOpen) {
    if (!elements.appShell) return;
    elements.appShell.classList.toggle('is-sidebar-open', Boolean(isOpen));
    if (elements.headerToggle) {
      elements.headerToggle.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    }
    if (elements.sidebarBackdrop) {
      elements.sidebarBackdrop.hidden = !isOpen;
    }
  }

  function toggleAdminAccountMenu() {
    if (!elements.adminAccountTrigger || !elements.adminAccountDropdown) return;

    const isOpen = elements.adminAccountTrigger.getAttribute('aria-expanded') === 'true';
    setAdminAccountMenuOpen(!isOpen);
  }

  function closeAdminAccountMenu() {
    setAdminAccountMenuOpen(false);
  }

  function setAdminAccountMenuOpen(isOpen) {
    if (!elements.adminAccountTrigger || !elements.adminAccountDropdown) return;

    elements.adminAccountTrigger.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    elements.adminAccountDropdown.hidden = !isOpen;
  }

  function openCurrentAdminEdit() {
    const currentAdminID = getUserId();
    if (!currentAdminID) {
      showError('找不到目前登入管理員');
      return;
    }

    openAdminEdit(currentAdminID);
  }

  function openUserInviteModal() {
    if (!elements.userInviteModal) return;
    elements.userInviteModal.hidden = false;
    requestAnimationFrame(() => elements.inviteUserEmail?.focus());
  }

  function closeUserInviteModal() {
    if (!elements.userInviteModal) return;
    elements.userInviteModal.hidden = true;
    elements.userInviteForm?.reset();
  }

  function openBulkUserInviteModal() {
    if (!elements.bulkUserInviteModal) return;
    resetBulkInviteModal();
    elements.bulkUserInviteModal.hidden = false;
    requestAnimationFrame(() => elements.bulkInviteFile?.focus());
  }

  function closeBulkUserInviteModal() {
    if (!elements.bulkUserInviteModal) return;
    elements.bulkUserInviteModal.hidden = true;
    resetBulkInviteModal();
  }

  function resetBulkInviteModal() {
    bulkInvitePrecheck = null;
    elements.bulkUserInviteForm?.reset();
    setBulkInviteStatus('', false);
    renderBulkInviteSummary(null);
    renderBulkInviteSendResult(null);
    updateBulkInviteSubmit(0, false);
  }

  /**
   * Switch between views
   */
  function switchView(viewName, options = {}) {
    if (!Object.prototype.hasOwnProperty.call(VIEW_ROUTES, viewName)) {
      viewName = 'dashboard';
    }
    if (viewName === 'conversationDetail') {
      const detailPath = options.path || window.location.pathname;
      const match = detailPath.match(/^\/office\/conversations\/([^/]+)$/);
      activeAdminConversationID = match ? decodeURIComponent(match[1]) : activeAdminConversationID;
    }

    if (!options.skipHistory) {
      const nextPath = options.path || VIEW_ROUTES[viewName] || (
        viewName === 'userEdit' && editingUserID ? `/office/user/${editingUserID}/edit` :
        viewName === 'adminEdit' && editingAdminUserID ? `/office/admins/${editingAdminUserID}/edit` :
        viewName === 'conversationDetail' ? window.location.pathname : ''
      );
      if (nextPath && window.location.pathname !== nextPath) {
        const method = options.replace ? 'replaceState' : 'pushState';
        window.history[method]({ adminView: viewName }, '', nextPath);
      }
    }

    // Update menu items
    const activeMenuItem = viewName === 'userCreate' || viewName === 'userEdit'
      ? 'users'
      : viewName === 'adminCreate' || viewName === 'adminEdit'
        ? 'admins'
        : viewName === 'conversationDetail' ? 'conversations' : viewName;
    document.querySelectorAll('[data-menu-item]').forEach(item => {
      item.classList.toggle('is-active', item.getAttribute('data-menu-item') === activeMenuItem);
    });
    updateSidebarGroups(activeMenuItem);

    // Update views
    if (elements.views) {
      elements.views.forEach(view => {
        const isActive = view.getAttribute('data-view') === viewName;
        view.hidden = !isActive;
        if (isActive) {
          view.classList.add('is-active');
        } else {
          view.classList.remove('is-active');
        }
      });
    }

    // Load view-specific data
    switch (viewName) {
      case 'users':
        loadUsers();
        break;
      case 'userCreate':
        elements.createUserAccount?.focus();
        break;
      case 'userInvitations':
        loadUserInvitations();
        elements.inviteUserEmail?.focus();
        break;
      case 'userEdit':
        loadUserForEdit();
        break;
      case 'conversations':
        loadAdminConversations();
        break;
      case 'conversationDetail':
        loadAdminConversationDetail();
        break;
      case 'admins':
        loadAdmins();
        break;
      case 'adminCreate':
        elements.adminCreateForm?.reset();
        elements.createAdminAccount?.focus();
        break;
      case 'adminEdit':
        loadAdminForEdit();
        break;
      case 'devices':
        // Already handled by button
        break;
      case 'whitelist':
        // User needs to enter ID
        break;
      case 'logs':
        loadLogs();
        break;
      case 'health':
        loadHealthStatus();
        break;
    }
  }

  function updateSidebarGroups(activeMenuItem) {
    document.querySelectorAll('.sidebar-group').forEach(group => {
      const containsActiveItem = Boolean(group.querySelector(`[data-menu-item="${activeMenuItem}"]`));
      setSidebarGroupOpen(group, containsActiveItem);
    });
  }

  function setSidebarGroupOpen(group, isOpen) {
    group.classList.toggle('is-open', isOpen);
    const toggle = group.querySelector('[data-sidebar-toggle]');
    if (toggle) {
      toggle.classList.toggle('is-group-open', isOpen);
    }
    const caret = toggle ? toggle.querySelector('.menu-caret') : null;
    if (caret) {
      caret.textContent = isOpen ? '⌃' : '⌄';
    }
  }

  /**
   * Load initial dashboard data
   */
  async function loadInitialData() {
    // For now, set placeholder stats
    updateStats();
    loadHealthStatus();
  }

  /**
   * Update dashboard statistics
   */
  function updateStats() {
    if (elements.statUsers) elements.statUsers.textContent = '計算中...';
    if (elements.statActive) elements.statActive.textContent = '計算中...';
    if (elements.statConversations) elements.statConversations.textContent = '計算中...';
    if (elements.statUptime) elements.statUptime.textContent = '計算中...';
    
    // These would normally be fetched from an API
    setTimeout(() => {
      if (elements.statUsers) elements.statUsers.textContent = '--';
      if (elements.statActive) elements.statActive.textContent = '--';
      if (elements.statConversations) elements.statConversations.textContent = '--';
      if (elements.statUptime) elements.statUptime.textContent = '--';
    }, 500);
  }

  /**
   * Load users list
   */
  async function loadUsers() {
    if (elements.usersLoading) elements.usersLoading.hidden = false;
    if (elements.usersEmpty) elements.usersEmpty.hidden = true;
    if (elements.usersTable) elements.usersTable.hidden = true;
    if (elements.usersPagination) elements.usersPagination.hidden = true;

    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/users?${buildUserQuery().toString()}`);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '載入用戶列表失敗');
      }

      const pageData = data.data || {};
      const users = Array.isArray(pageData.items) ? pageData.items : [];
      const total = Number(pageData.total) || 0;
      const totalPages = Number(pageData.total_pages) || 0;
      usersPage = Number(pageData.page) || usersPage;
      usersPerPage = Number(pageData.per_page) || usersPerPage;
      if (users.length === 0) {
        showUsersEmpty();
        renderUsersPagination(total, totalPages);
        return;
      }

      populateUsersTable(users);
      if (elements.usersTable) elements.usersTable.hidden = false;
      renderUsersPagination(total, totalPages);
    } catch (err) {
      showError('載入用戶列表失敗: ' + err.message);
      showUsersEmpty();
    } finally {
      if (elements.usersLoading) elements.usersLoading.hidden = true;
    }
  }

  function buildUserQuery() {
    const params = new URLSearchParams();
    const filters = [
      ['external_user_id', elements.userSearch?.value],
      ['display_name', elements.userDisplayName?.value],
      ['email', elements.userEmail?.value],
      ['status', elements.userStatus?.value],
      ['source_system', elements.userSource?.value],
      ['sort', elements.userSort?.value],
      ['created_from', elements.userCreatedFrom?.value],
      ['created_to', elements.userCreatedTo?.value]
    ];
    filters.forEach(([key, value]) => {
      const normalized = value?.trim();
      if (normalized) params.set(key, normalized);
    });
    params.set('page', String(usersPage));
    params.set('per_page', String(usersPerPage));
    return params;
  }

  function populateUsersTable(users) {
    if (!elements.usersTbody) return;

    elements.usersTbody.replaceChildren();
    users.forEach((user, index) => {
      const row = document.createElement('tr');
      USER_COLUMNS.forEach(column => {
        row.appendChild(renderUserCell(column.key, user, index));
      });

      elements.usersTbody.appendChild(row);
    });
  }

  function renderTableColumns(table, colgroup, thead, columns) {
    if (!table || !colgroup || !thead) return;

    colgroup.replaceChildren();
    const headerRow = document.createElement('tr');
    let minimumTableWidth = 0;

    columns.forEach(column => {
      const columnWidth = column.width || column.minWidth || 120;
      minimumTableWidth += columnWidth;

      const col = document.createElement('col');
      col.style.width = `${columnWidth}px`;
      colgroup.appendChild(col);

      const header = document.createElement('th');
      header.scope = 'col';
      header.textContent = column.title;
      headerRow.appendChild(header);
    });

    thead.replaceChildren(headerRow);
    table.style.minWidth = `${minimumTableWidth}px`;
  }

  function setupOverlayScrollbar(container) {
    if (!container) return;

    const viewport = container.querySelector('.table-scroll-viewport');
    const track = container.querySelector('.table-scrollbar');
    const thumb = container.querySelector('.table-scrollbar-thumb');
    if (!viewport || !track || !thumb) return;

    const syncThumb = () => {
      const scrollRange = viewport.scrollWidth - viewport.clientWidth;
      if (scrollRange <= 1) {
        track.hidden = true;
        return;
      }

      track.hidden = false;
      const trackWidth = track.clientWidth;
      const thumbWidth = Math.max(48, trackWidth * viewport.clientWidth / viewport.scrollWidth);
      const thumbRange = Math.max(0, trackWidth - thumbWidth);
      const thumbLeft = scrollRange > 0 ? thumbRange * viewport.scrollLeft / scrollRange : 0;
      thumb.style.width = `${thumbWidth}px`;
      thumb.style.transform = `translateX(${thumbLeft}px)`;
    };

    let hideTimer = 0;
    const revealTemporarily = () => {
      container.classList.add('is-scrolling');
      window.clearTimeout(hideTimer);
      hideTimer = window.setTimeout(() => container.classList.remove('is-scrolling'), 700);
    };

    viewport.addEventListener('scroll', () => {
      syncThumb();
      revealTemporarily();
    }, { passive: true });

    track.addEventListener('pointerdown', event => {
      event.preventDefault();
      const trackRect = track.getBoundingClientRect();
      const thumbRect = thumb.getBoundingClientRect();
      const startPointerX = event.clientX;
      const pointerOffset = event.target === thumb
        ? event.clientX - thumbRect.left
        : thumbRect.width / 2;

      if (event.target !== thumb) {
        const ratio = (event.clientX - trackRect.left - pointerOffset) / Math.max(1, trackRect.width - thumbRect.width);
        viewport.scrollLeft = ratio * (viewport.scrollWidth - viewport.clientWidth);
      }

      track.setPointerCapture(event.pointerId);
      thumb.style.cursor = 'grabbing';
      const dragStartScrollLeft = viewport.scrollLeft;
      const dragStartPointerX = event.target === thumb ? startPointerX : event.clientX;

      const move = moveEvent => {
        const thumbRange = Math.max(1, track.clientWidth - thumb.clientWidth);
        const scrollRange = Math.max(0, viewport.scrollWidth - viewport.clientWidth);
        viewport.scrollLeft = dragStartScrollLeft
          + (moveEvent.clientX - dragStartPointerX) * scrollRange / thumbRange;
      };
      const stop = () => {
        thumb.style.cursor = 'grab';
        track.removeEventListener('pointermove', move);
        track.removeEventListener('pointerup', stop);
        track.removeEventListener('pointercancel', stop);
      };
      track.addEventListener('pointermove', move);
      track.addEventListener('pointerup', stop);
      track.addEventListener('pointercancel', stop);
    });

    new ResizeObserver(syncThumb).observe(viewport);
    syncThumb();
  }

  function renderUserCell(key, user, index) {
    const cell = document.createElement('td');

    switch (key) {
      case 'index':
        cell.textContent = String((usersPage - 1) * usersPerPage + index + 1);
        break;
      case 'avatar': {
        const avatar = document.createElement('span');
        avatar.className = 'user-avatar';
        avatar.textContent = userInitial(user.display_name || user.external_user_id);
        cell.appendChild(avatar);
        break;
      }
      case 'account': {
        const account = document.createElement('span');
        account.className = 'account-link';
        account.textContent = user.external_user_id || '--';
        cell.appendChild(account);
        break;
      }
      case 'displayName':
        cell.textContent = user.display_name || '--';
        break;
      case 'email':
        cell.textContent = user.email || '--';
        break;
      case 'status': {
        const status = document.createElement('span');
        status.className = 'status-badge';
        status.textContent = user.status === 'active' ? '啟用' : '停用';
        cell.appendChild(status);
        break;
      }
      case 'source':
        cell.textContent = (user.source_system || '--').toUpperCase();
        break;
      case 'lastOnline':
        cell.textContent = formatDate(user.last_online_at);
        break;
      case 'options': {
        cell.className = 'actions-cell';
        const muteButton = document.createElement('button');
        muteButton.type = 'button';
        muteButton.className = user.is_chat_muted ? 'table-icon-action table-icon-action-danger' : 'table-icon-action';
        muteButton.title = user.is_chat_muted ? '解除禁言' : '禁止發言';
        muteButton.setAttribute('aria-label', muteButton.title);
        muteButton.innerHTML = user.is_chat_muted ? chatBubbleLockedIcon() : chatBubbleIcon();
        muteButton.addEventListener('click', () => toggleUserChatMute(user));
        cell.appendChild(muteButton);

        const temporaryPasswordButton = document.createElement('button');
        temporaryPasswordButton.type = 'button';
        temporaryPasswordButton.className = 'table-action table-action-wide';
        temporaryPasswordButton.textContent = '臨時密碼';
        temporaryPasswordButton.title = user.email ? '寄送臨時密碼' : '使用者沒有 Email，無法寄送臨時密碼';
        temporaryPasswordButton.disabled = !user.email;
        if (user.email) {
          temporaryPasswordButton.addEventListener('click', () => confirmSendTemporaryPassword(user));
        }
        cell.appendChild(temporaryPasswordButton);

        const editButton = document.createElement('button');
        editButton.type = 'button';
        editButton.className = 'table-action';
        editButton.textContent = '編輯';
        editButton.title = '編輯用戶';
        editButton.addEventListener('click', () => openUserEdit(user.id));
        cell.appendChild(editButton);
        break;
      }
      default:
        cell.textContent = '--';
    }

    return cell;
  }

  async function toggleUserChatMute(user) {
    if (!user?.id) return;
    const nextMuted = !Boolean(user.is_chat_muted);
    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/users/${user.id}/chat-mute`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_chat_muted: nextMuted })
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '更新禁言狀態失敗');
      }

      showSuccess(nextMuted ? '已禁止發言' : '已解除禁言');
      await loadUsers();
    } catch (err) {
      showError(err.message || '更新禁言狀態失敗');
    }
  }

  function confirmSendTemporaryPassword(user) {
    if (!user?.id || !user.email) return;
    openConfirmModal('寄送臨時密碼', [
      `將寄送 6 位數臨時密碼至 ${user.email}。`,
      { text: '送出後原密碼將失效，使用者登入後必須立即修改密碼。', className: 'text-danger' }
    ], () => sendTemporaryPassword(user), '寄送');
  }

  async function sendTemporaryPassword(user) {
    if (!user?.id) return;
    if (elements.modalConfirmBtn) {
      elements.modalConfirmBtn.disabled = true;
      elements.modalConfirmBtn.textContent = '寄送中...';
    }
    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/users/${user.id}/temporary-password`, {
        method: 'POST'
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '寄送臨時密碼失敗');
      }

      closeModal();
      showSuccess('臨時密碼已寄送');
      await loadUsers();
    } catch (err) {
      showError(err.message || '寄送臨時密碼失敗');
      if (elements.modalConfirmBtn) {
        elements.modalConfirmBtn.disabled = false;
        elements.modalConfirmBtn.textContent = '寄送';
      }
    }
  }

  function chatBubbleIcon() {
    return `
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="M21 11.5a8.5 8.5 0 0 1-8.5 8.5 8.9 8.9 0 0 1-3.7-.8L3 21l1.8-5.2A8.3 8.3 0 0 1 4 11.5 8.5 8.5 0 0 1 12.5 3 8.5 8.5 0 0 1 21 11.5Z"></path>
      </svg>`;
  }

  function chatBubbleLockedIcon() {
    return `
      <span class="chat-mute-icon">
        ${chatBubbleIcon()}
        <span class="chat-mute-badge" aria-hidden="true">
          <svg viewBox="0 0 24 24">
            <rect x="5" y="10" width="14" height="10" rx="2"></rect>
            <path d="M8 10V7a4 4 0 0 1 8 0v3"></path>
          </svg>
        </span>
      </span>`;
  }

  function renderUsersPagination(total, totalPages) {
    if (!elements.usersPagination || !elements.usersPages) return;

    if (elements.usersTotal) elements.usersTotal.textContent = String(total);
    if (elements.usersPerPage) elements.usersPerPage.value = String(usersPerPage);
    elements.usersPages.replaceChildren();

    const addButton = (label, page, options = {}) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'pagination-button';
      button.textContent = label;
      button.disabled = Boolean(options.disabled);
      if (options.current) {
        button.classList.add('is-current');
        button.setAttribute('aria-current', 'page');
      } else if (!options.disabled) {
        button.addEventListener('click', () => {
          usersPage = page;
          loadUsers();
        });
      }
      elements.usersPages.appendChild(button);
    };

    addButton('‹', usersPage - 1, { disabled: usersPage <= 1 });
    paginationSequence(usersPage, totalPages).forEach(page => {
      if (page === 'ellipsis') {
        const ellipsis = document.createElement('span');
        ellipsis.className = 'pagination-ellipsis';
        ellipsis.textContent = '…';
        elements.usersPages.appendChild(ellipsis);
        return;
      }
      addButton(String(page), page, { current: page === usersPage });
    });
    addButton('›', usersPage + 1, { disabled: usersPage >= totalPages || totalPages === 0 });
    elements.usersPagination.hidden = false;
  }

  async function loadUserInvitations() {
    if (elements.userInvitationsLoading) elements.userInvitationsLoading.hidden = false;
    if (elements.userInvitationsEmpty) elements.userInvitationsEmpty.hidden = true;
    if (elements.userInvitationsTable) elements.userInvitationsTable.hidden = true;
    if (elements.userInvitationsPagination) elements.userInvitationsPagination.hidden = true;

    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/user-invitations?${buildUserInvitationQuery().toString()}`);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '載入註冊邀請列表失敗');
      }

      const pageData = data.data || {};
      const invitations = Array.isArray(pageData.items) ? pageData.items : [];
      const total = Number(pageData.total) || 0;
      const totalPages = Number(pageData.total_pages) || 0;
      userInvitationsPage = Number(pageData.page) || userInvitationsPage;
      userInvitationsPerPage = Number(pageData.per_page) || userInvitationsPerPage;

      if (invitations.length === 0) {
        showUserInvitationsEmpty();
        renderUserInvitationsPagination(total, totalPages);
        return;
      }

      populateUserInvitationsTable(invitations);
      if (elements.userInvitationsTable) elements.userInvitationsTable.hidden = false;
      renderUserInvitationsPagination(total, totalPages);
    } catch (err) {
      showError('載入註冊邀請列表失敗: ' + err.message);
      showUserInvitationsEmpty();
    } finally {
      if (elements.userInvitationsLoading) elements.userInvitationsLoading.hidden = true;
    }
  }

  function buildUserInvitationQuery() {
    const params = new URLSearchParams();
    const email = elements.userInvitationEmailFilter?.value?.trim();
    const status = elements.userInvitationStatusFilter?.value?.trim();
    const sorting = elements.userInvitationSort?.value?.trim();
    if (email) params.set('email', email);
    if (status) params.set('status', status);
    if (sorting) params.set('sorting', sorting);
    params.set('page', String(userInvitationsPage));
    params.set('per_page', String(userInvitationsPerPage));
    return params;
  }

  function populateUserInvitationsTable(invitations) {
    if (!elements.userInvitationsTbody) return;

    elements.userInvitationsTbody.replaceChildren();
    invitations.forEach(invitation => {
      const row = document.createElement('tr');
      [
        invitation.email || '--',
        null,
        formatDate(invitation.sent_at),
        formatDate(invitation.expires_at),
        formatDate(invitation.accepted_at),
        invitationInviter(invitation),
        null
      ].forEach((value, index) => {
        const cell = document.createElement('td');
        if (index === 0) {
          const email = document.createElement('span');
          email.className = 'account-link';
          email.textContent = value;
          cell.appendChild(email);
        } else if (index === 1) {
          const status = document.createElement('span');
          status.className = 'status-badge';
          status.textContent = invitationStatusLabel(invitation.status);
          cell.appendChild(status);
        } else if (index === 6) {
          if (canResendInvitation(invitation)) {
            const resendButton = document.createElement('button');
            resendButton.type = 'button';
            resendButton.className = 'btn invitation-action-btn';
            resendButton.textContent = '重送';
            resendButton.addEventListener('click', () => confirmResendUserInvitation(invitation, resendButton));
            cell.appendChild(resendButton);
          } else {
            cell.textContent = '--';
          }
        } else {
          cell.textContent = value;
        }
        row.appendChild(cell);
      });
      elements.userInvitationsTbody.appendChild(row);
    });
  }

  function canResendInvitation(invitation) {
    return invitation && invitation.email && invitation.status !== 'accepted';
  }

  function invitationStatusLabel(status) {
    switch (status) {
      case 'pending':
        return '待註冊';
      case 'accepted':
        return '已完成';
      case 'expired':
        return '已過期';
      default:
        return status || '--';
    }
  }

  function invitationInviter(invitation) {
    return invitation.invited_by_admin_name
      || invitation.invited_by_admin_account
      || (invitation.invited_by_admin_id ? String(invitation.invited_by_admin_id) : '--');
  }

  function renderUserInvitationsPagination(total, totalPages) {
    if (!elements.userInvitationsPagination || !elements.userInvitationsPages) return;

    if (elements.userInvitationsTotal) elements.userInvitationsTotal.textContent = String(total);
    if (elements.userInvitationsPerPage) elements.userInvitationsPerPage.value = String(userInvitationsPerPage);
    elements.userInvitationsPages.replaceChildren();

    const addButton = (label, page, options = {}) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'pagination-button';
      button.textContent = label;
      button.disabled = Boolean(options.disabled);
      if (options.current) {
        button.classList.add('is-current');
        button.setAttribute('aria-current', 'page');
      } else if (!options.disabled) {
        button.addEventListener('click', () => {
          userInvitationsPage = page;
          loadUserInvitations();
        });
      }
      elements.userInvitationsPages.appendChild(button);
    };

    addButton('‹', userInvitationsPage - 1, { disabled: userInvitationsPage <= 1 });
    paginationSequence(userInvitationsPage, totalPages).forEach(page => {
      if (page === 'ellipsis') {
        const ellipsis = document.createElement('span');
        ellipsis.className = 'pagination-ellipsis';
        ellipsis.textContent = '…';
        elements.userInvitationsPages.appendChild(ellipsis);
        return;
      }
      addButton(String(page), page, { current: page === userInvitationsPage });
    });
    addButton('›', userInvitationsPage + 1, { disabled: userInvitationsPage >= totalPages || totalPages === 0 });
    elements.userInvitationsPagination.hidden = false;
  }

  function showUserInvitationsEmpty() {
    if (elements.userInvitationsLoading) elements.userInvitationsLoading.hidden = true;
    if (elements.userInvitationsEmpty) elements.userInvitationsEmpty.hidden = false;
    if (elements.userInvitationsTable) elements.userInvitationsTable.hidden = true;
  }

  async function loadAdminConversations() {
    if (elements.adminConversationsLoading) elements.adminConversationsLoading.hidden = false;
    if (elements.adminConversationsEmpty) elements.adminConversationsEmpty.hidden = true;
    if (elements.adminConversationsTable) elements.adminConversationsTable.hidden = true;
    if (elements.adminConversationsPagination) elements.adminConversationsPagination.hidden = true;

    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/conversations?${buildAdminConversationQuery().toString()}`);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '載入聊天紀錄失敗');
      }

      const pageData = data.data || {};
      const conversations = Array.isArray(pageData.items) ? pageData.items : [];
      const total = Number(pageData.total) || 0;
      const totalPages = Number(pageData.total_pages) || 0;
      adminConversationsPage = Number(pageData.page) || adminConversationsPage;
      adminConversationsPerPage = Number(pageData.per_page) || adminConversationsPerPage;

      if (conversations.length === 0) {
        showAdminConversationsEmpty();
        renderAdminConversationsPagination(total, totalPages);
        return;
      }

      populateAdminConversationsTable(conversations);
      if (elements.adminConversationsTable) elements.adminConversationsTable.hidden = false;
      renderAdminConversationsPagination(total, totalPages);
    } catch (err) {
      showError('載入聊天紀錄失敗: ' + err.message);
      showAdminConversationsEmpty();
    } finally {
      if (elements.adminConversationsLoading) elements.adminConversationsLoading.hidden = true;
    }
  }

  function buildAdminConversationQuery() {
    const params = new URLSearchParams();
    const type = elements.adminConversationType?.value?.trim();
    const keyword = elements.adminConversationKeyword?.value?.trim();
    const lastFrom = elements.adminConversationLastFrom?.value?.trim();
    const lastTo = elements.adminConversationLastTo?.value?.trim();
    if (type) params.set('type', type);
    if (keyword) params.set('keyword', keyword);
    if (lastFrom) params.set('last_activity_from', lastFrom);
    if (lastTo) params.set('last_activity_to', lastTo);
    params.set('page', String(adminConversationsPage));
    params.set('per_page', String(adminConversationsPerPage));
    return params;
  }

  function populateAdminConversationsTable(conversations) {
    if (!elements.adminConversationsTbody) return;

    elements.adminConversationsTbody.replaceChildren();
    conversations.forEach(conversation => {
      const row = document.createElement('tr');
      [
        String(conversation.id || ''),
        conversationTypeLabel(conversation.type),
        conversation.name || '--',
        String(conversation.member_count || 0),
        conversation.latest_message_available ? (conversation.last_message || '--') : '--',
        conversationLastSender(conversation),
        formatDate(conversation.last_activity_at),
        null
      ].forEach((value, index) => {
        const cell = document.createElement('td');
        if (index === 0) {
          const id = document.createElement('span');
          id.className = 'account-link';
          id.textContent = value;
          cell.appendChild(id);
        } else if (index === 1) {
          const badge = document.createElement('span');
          badge.className = 'status-badge';
          badge.textContent = value;
          cell.appendChild(badge);
        } else if (index === 4) {
          cell.className = 'admin-conversation-message-cell';
          cell.textContent = value;
        } else if (index === 7) {
          const viewButton = document.createElement('button');
          viewButton.type = 'button';
          viewButton.className = 'btn invitation-action-btn';
          viewButton.textContent = '查看';
          viewButton.addEventListener('click', () => {
            switchView('conversationDetail', { path: `/office/conversations/${encodeURIComponent(conversation.id)}` });
          });
          cell.appendChild(viewButton);
        } else {
          cell.textContent = value;
        }
        row.appendChild(cell);
      });
      elements.adminConversationsTbody.appendChild(row);
    });
  }

  function conversationTypeLabel(type) {
    switch (type) {
      case 'direct':
        return '一對一';
      case 'group':
        return '群組';
      default:
        return type || '--';
    }
  }

  function conversationLastSender(conversation) {
    if (!conversation || !conversation.latest_message_available) {
      return '--';
    }
    return conversation.last_sender_name || conversation.last_sender_account || (
      conversation.last_sender_id ? `使用者#${conversation.last_sender_id}` : '--'
    );
  }

  function renderAdminConversationsPagination(total, totalPages) {
    if (!elements.adminConversationsPagination || !elements.adminConversationsPages) return;

    if (elements.adminConversationsTotal) elements.adminConversationsTotal.textContent = String(total);
    if (elements.adminConversationsPerPage) elements.adminConversationsPerPage.value = String(adminConversationsPerPage);
    elements.adminConversationsPages.replaceChildren();

    const addButton = (label, page, options = {}) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'pagination-button';
      button.textContent = label;
      button.disabled = Boolean(options.disabled);
      if (options.current) {
        button.classList.add('is-current');
        button.setAttribute('aria-current', 'page');
      } else if (!options.disabled) {
        button.addEventListener('click', () => {
          adminConversationsPage = page;
          loadAdminConversations();
        });
      }
      elements.adminConversationsPages.appendChild(button);
    };

    addButton('‹', adminConversationsPage - 1, { disabled: adminConversationsPage <= 1 });
    paginationSequence(adminConversationsPage, totalPages).forEach(page => {
      if (page === 'ellipsis') {
        const ellipsis = document.createElement('span');
        ellipsis.className = 'pagination-ellipsis';
        ellipsis.textContent = '…';
        elements.adminConversationsPages.appendChild(ellipsis);
        return;
      }
      addButton(String(page), page, { current: page === adminConversationsPage });
    });
    addButton('›', adminConversationsPage + 1, { disabled: adminConversationsPage >= totalPages || totalPages === 0 });
    elements.adminConversationsPagination.hidden = false;
  }

  function showAdminConversationsEmpty() {
    if (elements.adminConversationsLoading) elements.adminConversationsLoading.hidden = true;
    if (elements.adminConversationsEmpty) elements.adminConversationsEmpty.hidden = false;
    if (elements.adminConversationsTable) elements.adminConversationsTable.hidden = true;
  }

  function resetAdminConversationFilters() {
    if (elements.adminConversationFilterForm) elements.adminConversationFilterForm.reset();
    adminConversationsPage = 1;
    loadAdminConversations();
  }

  async function loadAdminConversationDetail() {
    if (!activeAdminConversationID) {
      showError('缺少聊天室 ID');
      return;
    }

    if (elements.adminConversationDetailLoading) elements.adminConversationDetailLoading.hidden = false;
    if (elements.adminConversationDetailSummary) elements.adminConversationDetailSummary.hidden = true;
    if (elements.adminConversationMessagesEmpty) elements.adminConversationMessagesEmpty.hidden = true;
    if (elements.adminConversationMessagesTable) elements.adminConversationMessagesTable.hidden = true;
    if (elements.adminConversationMessagesPagination) elements.adminConversationMessagesPagination.hidden = true;

    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/conversations/${encodeURIComponent(activeAdminConversationID)}?${buildAdminConversationMessageQuery().toString()}`);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '載入聊天室明細失敗');
      }

      const detail = data.data || {};
      populateAdminConversationSummary(detail.conversation || {}, detail.participants || []);

      const pageData = detail.messages || {};
      const messages = Array.isArray(pageData.items) ? pageData.items : [];
      const total = Number(pageData.total) || 0;
      const totalPages = Number(pageData.total_pages) || 0;
      adminConversationMessagesPage = Number(pageData.page) || adminConversationMessagesPage;
      adminConversationMessagesPerPage = Number(pageData.per_page) || adminConversationMessagesPerPage;

      if (messages.length === 0) {
        showAdminConversationMessagesEmpty();
        renderAdminConversationMessagesPagination(total, totalPages);
        return;
      }

      populateAdminConversationMessagesTable(messages);
      if (elements.adminConversationMessagesTable) elements.adminConversationMessagesTable.hidden = false;
      renderAdminConversationMessagesPagination(total, totalPages);
    } catch (err) {
      showError('載入聊天室明細失敗: ' + err.message);
      showAdminConversationMessagesEmpty();
    } finally {
      if (elements.adminConversationDetailLoading) elements.adminConversationDetailLoading.hidden = true;
    }
  }

  function buildAdminConversationMessageQuery() {
    const params = new URLSearchParams();
    const keyword = elements.adminConversationMessageKeyword?.value?.trim();
    const sender = elements.adminConversationMessageSender?.value?.trim();
    const type = elements.adminConversationMessageType?.value?.trim();
    const mention = elements.adminConversationMessageMention?.value?.trim();
    const sentFrom = elements.adminConversationMessageFrom?.value?.trim();
    const sentTo = elements.adminConversationMessageTo?.value?.trim();
    if (keyword) params.set('keyword', keyword);
    if (sender) params.set('sender', sender);
    if (type) params.set('message_type', type);
    if (mention) params.set('mention_type', mention);
    if (sentFrom) params.set('sent_from', sentFrom);
    if (sentTo) params.set('sent_to', sentTo);
    params.set('page', String(adminConversationMessagesPage));
    params.set('per_page', String(adminConversationMessagesPerPage));
    return params;
  }

  function populateAdminConversationSummary(conversation, participants) {
    const participantLabels = (Array.isArray(participants) ? participants : []).map(adminConversationUserLabel);
    const creatorLabel = conversation.created_by_name || conversation.created_by_account || (
      conversation.created_by_id ? `使用者#${conversation.created_by_id}` : '--'
    );
    const conversationID = String(conversation.id || activeAdminConversationID || '--');
    const conversationName = conversation.name || participantLabels.join(' ↔ ') || '--';

    if (elements.adminConversationDetailTitle) elements.adminConversationDetailTitle.textContent = conversation.name || '聊天室明細';
    if (elements.adminConversationDetailType) elements.adminConversationDetailType.textContent = conversationTypeLabel(conversation.type);
    if (elements.adminConversationBreadcrumbCurrent) elements.adminConversationBreadcrumbCurrent.textContent = `ID:${conversationID} - ${conversationName}`;
    if (elements.adminConversationDetailID) elements.adminConversationDetailID.textContent = conversationID;
    if (elements.adminConversationDetailName) elements.adminConversationDetailName.textContent = conversationName;
    if (elements.adminConversationDetailCreatedAt) elements.adminConversationDetailCreatedAt.textContent = formatDate(conversation.created_at);
    if (elements.adminConversationDetailLastAt) elements.adminConversationDetailLastAt.textContent = formatDate(conversation.last_activity_at);
    if (elements.adminConversationDetailParticipants) elements.adminConversationDetailParticipants.textContent = participantLabels.join(' ↔ ') || '--';
    if (elements.adminConversationDetailGroupName) elements.adminConversationDetailGroupName.textContent = conversation.name || '--';
    if (elements.adminConversationDetailCreator) elements.adminConversationDetailCreator.textContent = creatorLabel;
    if (elements.adminConversationDetailMemberCount) elements.adminConversationDetailMemberCount.textContent = String(conversation.member_count || participants.length || 0);

    const isGroup = conversation.type === 'group';
    if (elements.adminConversationDirectUsers) elements.adminConversationDirectUsers.hidden = isGroup;
    if (elements.adminConversationGroupName) elements.adminConversationGroupName.hidden = !isGroup;
    if (elements.adminConversationGroupCreator) elements.adminConversationGroupCreator.hidden = !isGroup;
    if (elements.adminConversationGroupMembers) elements.adminConversationGroupMembers.hidden = !isGroup;
    renderAdminConversationMembers(isGroup, participants);
    if (elements.adminConversationDetailSummary) elements.adminConversationDetailSummary.hidden = false;
  }

  function renderAdminConversationMembers(isGroup, participants) {
    if (!elements.adminConversationMembers || !elements.adminConversationMembersList || !elements.adminConversationMembersToggle) return;

    const members = Array.isArray(participants) ? participants : [];
    elements.adminConversationMembers.hidden = !isGroup;
    elements.adminConversationMembersList.replaceChildren();
    elements.adminConversationMembersToggle.textContent = `群組人員（${members.length}）`;
    elements.adminConversationMembersToggle.setAttribute('aria-expanded', 'true');
    elements.adminConversationMembersList.hidden = false;

    if (!isGroup) return;

    if (members.length === 0) {
      const empty = document.createElement('span');
      empty.className = 'admin-conversation-member-empty';
      empty.textContent = '尚無成員資料';
      elements.adminConversationMembersList.appendChild(empty);
      return;
    }

    members.forEach(member => {
      const chip = document.createElement('span');
      chip.className = 'admin-conversation-member-chip';
      chip.textContent = adminConversationUserLabel(member);
      elements.adminConversationMembersList.appendChild(chip);
    });
  }

  function toggleAdminConversationMembers() {
    if (!elements.adminConversationMembersToggle || !elements.adminConversationMembersList) return;

    const isExpanded = elements.adminConversationMembersToggle.getAttribute('aria-expanded') === 'true';
    elements.adminConversationMembersToggle.setAttribute('aria-expanded', isExpanded ? 'false' : 'true');
    elements.adminConversationMembersList.hidden = isExpanded;
  }

  function populateAdminConversationMessagesTable(messages) {
    if (!elements.adminConversationMessagesTbody) return;

    elements.adminConversationMessagesTbody.replaceChildren();
    messages.forEach(message => {
      const row = document.createElement('tr');
      [
        String(message.id || ''),
        formatDate(message.sent_at),
        message.sender_account || (message.sender_id ? `使用者#${message.sender_id}` : '--'),
        message.sender_name || '--',
        messageTypeLabel(message.message_type),
        null,
        mentionSummary(message.mentions),
        messageStatusLabel(message)
      ].forEach((value, index) => {
        const cell = document.createElement('td');
        if (index === 0) {
          const id = document.createElement('span');
          id.className = 'account-link';
          id.textContent = value;
          cell.appendChild(id);
        } else if (index === 4 || index === 7) {
          const badge = document.createElement('span');
          badge.className = 'status-badge';
          badge.textContent = value;
          cell.appendChild(badge);
        } else if (index === 5) {
          cell.className = 'admin-message-content-cell';
          renderAdminMessageContent(cell, message);
        } else {
          cell.textContent = value;
        }
        row.appendChild(cell);
      });
      elements.adminConversationMessagesTbody.appendChild(row);
    });
  }

  function renderAdminMessageContent(cell, message) {
    if (message?.is_recalled || message?.status === 'recalled') {
      cell.textContent = '已收回';
      return;
    }
    const type = String(message?.message_type || '').toLowerCase();
    if (type === 'image') {
      const attachments = Array.isArray(message.attachments) ? message.attachments : [];
      appendAdminMessageText(cell, message?.content);
      if (attachments.length === 0) {
        if (!cell.textContent) cell.textContent = '[圖片]';
        return;
      }
      attachments.forEach(attachment => {
        const link = document.createElement('a');
        link.className = 'admin-message-image-link';
        link.href = attachment.url || '#';
        link.target = '_blank';
        link.rel = 'noopener noreferrer';
        const image = document.createElement('img');
        image.src = attachment.url || '';
        image.alt = attachment.original_name || '圖片';
        image.loading = 'lazy';
        link.appendChild(image);
        cell.appendChild(link);
      });
      return;
    }
    if (type === 'file') {
      const attachments = Array.isArray(message.attachments) ? message.attachments : [];
      appendAdminMessageText(cell, message?.content);
      if (attachments.length === 0) {
        if (!cell.textContent) cell.textContent = '[檔案]';
        return;
      }
      attachments.forEach((attachment, index) => {
        if (index > 0) cell.appendChild(document.createTextNode('、'));
        const link = document.createElement('a');
        link.href = attachment.url || '#';
        link.target = '_blank';
        link.rel = 'noopener noreferrer';
        link.textContent = attachment.original_name || '[檔案]';
        cell.appendChild(link);
      });
      return;
    }
    cell.textContent = message?.content || '--';
  }

  function appendAdminMessageText(cell, content) {
    const text = String(content || '').trim();
    if (!text) return;

    const textElement = document.createElement('div');
    textElement.className = 'admin-message-caption';
    textElement.textContent = text;
    cell.appendChild(textElement);
  }

  function messageTypeLabel(type) {
    switch (type) {
      case 'text':
        return '文字';
      case 'image':
        return '圖片';
      case 'file':
        return '檔案';
      default:
        return type || '--';
    }
  }

  function mentionSummary(mentions) {
    if (!Array.isArray(mentions) || mentions.length === 0) {
      return '無標註';
    }
    if (mentions.some(mention => mention.mention_type === 'all')) {
      return '@ALL';
    }
    return mentions.map(mention => '@' + adminConversationUserLabel(mention)).join('、');
  }

  function messageStatusLabel(message) {
    if (message?.is_recalled || message?.status === 'recalled') {
      return '已收回';
    }
    return '正常';
  }

  function adminConversationUserLabel(user) {
    if (!user) return '已刪除使用者';
    return user.display_name || user.account || (user.user_id ? `使用者#${user.user_id}` : '已刪除使用者');
  }

  function renderAdminConversationMessagesPagination(total, totalPages) {
    if (!elements.adminConversationMessagesPagination || !elements.adminConversationMessagesPages) return;

    if (elements.adminConversationMessagesTotal) elements.adminConversationMessagesTotal.textContent = String(total);
    if (elements.adminConversationMessagesPerPage) elements.adminConversationMessagesPerPage.value = String(adminConversationMessagesPerPage);
    elements.adminConversationMessagesPages.replaceChildren();

    const addButton = (label, page, options = {}) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'pagination-button';
      button.textContent = label;
      button.disabled = Boolean(options.disabled);
      if (options.current) {
        button.classList.add('is-current');
        button.setAttribute('aria-current', 'page');
      } else if (!options.disabled) {
        button.addEventListener('click', () => {
          adminConversationMessagesPage = page;
          loadAdminConversationDetail();
        });
      }
      elements.adminConversationMessagesPages.appendChild(button);
    };

    addButton('‹', adminConversationMessagesPage - 1, { disabled: adminConversationMessagesPage <= 1 });
    paginationSequence(adminConversationMessagesPage, totalPages).forEach(page => {
      if (page === 'ellipsis') {
        const ellipsis = document.createElement('span');
        ellipsis.className = 'pagination-ellipsis';
        ellipsis.textContent = '…';
        elements.adminConversationMessagesPages.appendChild(ellipsis);
        return;
      }
      addButton(String(page), page, { current: page === adminConversationMessagesPage });
    });
    addButton('›', adminConversationMessagesPage + 1, { disabled: adminConversationMessagesPage >= totalPages || totalPages === 0 });
    elements.adminConversationMessagesPagination.hidden = false;
  }

  function showAdminConversationMessagesEmpty() {
    if (elements.adminConversationMessagesEmpty) elements.adminConversationMessagesEmpty.hidden = false;
    if (elements.adminConversationMessagesTable) elements.adminConversationMessagesTable.hidden = true;
  }

  function resetAdminConversationMessageFilters() {
    if (elements.adminConversationMessageFilterForm) elements.adminConversationMessageFilterForm.reset();
    adminConversationMessagesPage = 1;
    loadAdminConversationDetail();
  }

  function paginationSequence(current, totalPages) {
    if (totalPages <= 7) {
      return Array.from({ length: totalPages }, (_, index) => index + 1);
    }

    const pages = new Set([1, totalPages, current - 1, current, current + 1]);
    const ordered = [...pages].filter(page => page >= 1 && page <= totalPages).sort((a, b) => a - b);
    const sequence = [];
    ordered.forEach((page, index) => {
      if (index > 0 && page - ordered[index - 1] > 1) sequence.push('ellipsis');
      sequence.push(page);
    });
    return sequence;
  }

  async function loadAdmins() {
    if (elements.adminsLoading) elements.adminsLoading.hidden = false;
    if (elements.adminsEmpty) elements.adminsEmpty.hidden = true;
    if (elements.adminsTable) elements.adminsTable.hidden = true;
    if (elements.adminsPagination) elements.adminsPagination.hidden = true;

    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/admins?${buildAdminQuery().toString()}`);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '載入管理員列表失敗');
      }

      const pageData = data.data || {};
      const admins = Array.isArray(pageData.items) ? pageData.items : [];
      const total = Number(pageData.total) || 0;
      const totalPages = Number(pageData.total_pages) || 0;
      adminsPage = Number(pageData.page) || adminsPage;
      adminsPerPage = Number(pageData.per_page) || adminsPerPage;

      if (admins.length === 0) {
        showAdminsEmpty();
        renderAdminsPagination(total, totalPages);
        return;
      }

      populateAdminsTable(admins);
      if (elements.adminsTable) elements.adminsTable.hidden = false;
      renderAdminsPagination(total, totalPages);
    } catch (err) {
      showError('載入管理員列表失敗: ' + err.message);
      showAdminsEmpty();
    } finally {
      if (elements.adminsLoading) elements.adminsLoading.hidden = true;
    }
  }

  function buildAdminQuery() {
    const params = new URLSearchParams();
    const filters = [
      ['external_user_id', elements.adminSearch?.value],
      ['display_name', elements.adminDisplayName?.value],
      ['status', elements.adminStatus?.value]
    ];
    filters.forEach(([key, value]) => {
      const normalized = value?.trim();
      if (normalized) params.set(key, normalized);
    });
    params.set('page', String(adminsPage));
    params.set('per_page', String(adminsPerPage));
    return params;
  }

  function populateAdminsTable(admins) {
    if (!elements.adminsTbody) return;

    elements.adminsTbody.replaceChildren();
    admins.forEach(admin => {
      const row = document.createElement('tr');
      [
        admin.external_user_id || '--',
        admin.display_name || '--',
        admin.role_name || admin.role || '--',
        null,
        formatDate(admin.last_login_at),
        admin.last_login_ip || '--',
        null
      ].forEach((value, index) => {
        const cell = document.createElement('td');
        if (index === 0) {
          const account = document.createElement('span');
          account.className = 'account-link';
          account.textContent = value;
          cell.appendChild(account);
        } else if (index === 3) {
          const status = document.createElement('span');
          status.className = 'status-badge';
          status.textContent = admin.status === 'active' ? '啟用' : '停用';
          cell.appendChild(status);
        } else if (index === 6) {
          cell.className = 'actions-cell';
          const editButton = document.createElement('button');
          editButton.type = 'button';
          editButton.className = 'table-action';
          editButton.textContent = '編輯';
          editButton.title = '編輯管理員';
          editButton.addEventListener('click', () => openAdminEdit(admin.user_id));
          cell.appendChild(editButton);
        } else {
          cell.textContent = value;
        }
        row.appendChild(cell);
      });
      elements.adminsTbody.appendChild(row);
    });
  }

  function renderAdminsPagination(total, totalPages) {
    if (!elements.adminsPagination || !elements.adminsPages) return;

    if (elements.adminsTotal) elements.adminsTotal.textContent = String(total);
    if (elements.adminsPerPage) elements.adminsPerPage.value = String(adminsPerPage);
    elements.adminsPages.replaceChildren();

    const addButton = (label, page, options = {}) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'pagination-button';
      button.textContent = label;
      button.disabled = Boolean(options.disabled);
      if (options.current) {
        button.classList.add('is-current');
        button.setAttribute('aria-current', 'page');
      } else if (!options.disabled) {
        button.addEventListener('click', () => {
          adminsPage = page;
          loadAdmins();
        });
      }
      elements.adminsPages.appendChild(button);
    };

    addButton('‹', adminsPage - 1, { disabled: adminsPage <= 1 });
    paginationSequence(adminsPage, totalPages).forEach(page => {
      if (page === 'ellipsis') {
        const ellipsis = document.createElement('span');
        ellipsis.className = 'pagination-ellipsis';
        ellipsis.textContent = '…';
        elements.adminsPages.appendChild(ellipsis);
        return;
      }
      addButton(String(page), page, { current: page === adminsPage });
    });
    addButton('›', adminsPage + 1, { disabled: adminsPage >= totalPages || totalPages === 0 });
    elements.adminsPagination.hidden = false;
  }

  function showAdminsEmpty() {
    if (elements.adminsLoading) elements.adminsLoading.hidden = true;
    if (elements.adminsEmpty) elements.adminsEmpty.hidden = false;
    if (elements.adminsTable) elements.adminsTable.hidden = true;
  }

  async function createAdmin(event) {
    event.preventDefault();

    const password = elements.createAdminPassword?.value || '';
    if (/\s/.test(password)) {
      showError('密碼不可包含空白');
      elements.createAdminPassword?.focus();
      return;
    }

    const payload = {
      external_user_id: elements.createAdminAccount?.value?.trim() || '',
      display_name: elements.createAdminDisplayName?.value?.trim() || '',
      role: elements.createAdminRole?.value || 'system_admin',
      password,
      status: elements.createAdminStatus?.value || 'active'
    };
    if (!payload.external_user_id || !payload.display_name || !payload.password) {
      showError('請填寫帳號、名稱與密碼');
      return;
    }

    if (elements.submitCreateAdmin) elements.submitCreateAdmin.disabled = true;
    try {
      const response = await makeAuthenticatedRequest('/api/system-admin/admins', {
        method: 'POST',
        body: JSON.stringify(payload)
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '建立管理員失敗');
      }

      adminsPage = 1;
      showSuccess('管理員已建立');
      switchView('admins');
    } catch (err) {
      showError(err.message || '建立管理員失敗');
    } finally {
      if (elements.submitCreateAdmin) elements.submitCreateAdmin.disabled = false;
    }
  }

  function openAdminEdit(userID) {
    editingAdminUserID = String(userID);
    switchView('adminEdit', { path: `/office/admins/${encodeURIComponent(editingAdminUserID)}/edit` });
  }

  async function loadAdminForEdit() {
    if (!editingAdminUserID || !elements.adminEditForm) return;

    if (elements.submitEditAdmin) elements.submitEditAdmin.disabled = true;
    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/admins/${encodeURIComponent(editingAdminUserID)}`);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '載入管理員資料失敗');
      }

      const admin = data.data || {};
      elements.editAdminAccount.value = admin.external_user_id || '';
      originalEditableAdmin = {
        password: '',
        displayName: admin.display_name || '',
        email: admin.email || '',
        role: admin.role || 'system_admin',
        status: admin.status || 'active'
      };
      resetAdminEditForm();
      elements.editAdminDisplayName.focus();
    } catch (err) {
      showError(err.message || '載入管理員資料失敗');
      switchView('admins');
    } finally {
      if (elements.submitEditAdmin) elements.submitEditAdmin.disabled = false;
    }
  }

  function resetAdminEditForm() {
    if (!originalEditableAdmin) return;

    const editableFields = [
      [elements.editAdminPassword, originalEditableAdmin.password],
      [elements.editAdminDisplayName, originalEditableAdmin.displayName],
      [elements.editAdminStatus, originalEditableAdmin.status]
    ];
    editableFields.forEach(([field, originalValue]) => {
      if (field && field.value !== originalValue) {
        field.value = originalValue;
      }
    });
  }

  async function updateAdmin(event) {
    event.preventDefault();
    if (!editingAdminUserID || !originalEditableAdmin) return;

    const password = elements.editAdminPassword?.value || '';
    if (password && /\s/.test(password)) {
      showError('密碼不可包含空白');
      elements.editAdminPassword?.focus();
      return;
    }

    const payload = {
      password,
      display_name: elements.editAdminDisplayName?.value?.trim() || '',
      role: originalEditableAdmin.role || 'system_admin',
      email: originalEditableAdmin.email,
      status: elements.editAdminStatus?.value || 'active'
    };
    if (!payload.display_name) {
      showError('請填寫名稱');
      return;
    }

    if (elements.submitEditAdmin) elements.submitEditAdmin.disabled = true;
    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/admins/${encodeURIComponent(editingAdminUserID)}`, {
        method: 'PATCH',
        body: JSON.stringify(payload)
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '更新管理員失敗');
      }

      elements.editAdminPassword.value = '';
      showSuccess('管理員資料已更新');
      switchView('admins');
    } catch (err) {
      showError(err.message || '更新管理員失敗');
    } finally {
      if (elements.submitEditAdmin) elements.submitEditAdmin.disabled = false;
    }
  }

  function userInitial(value) {
    const normalized = String(value || '').trim();
    return normalized ? normalized.slice(0, 1).toUpperCase() : '?';
  }

  async function createUser(event) {
    event.preventDefault();

    const password = elements.createUserPassword?.value || '';
    if (/\s/.test(password)) {
      showError('密碼不可包含空白');
      elements.createUserPassword?.focus();
      return;
    }

    const payload = {
      source_system: 'office',
      external_user_id: elements.createUserAccount?.value?.trim() || '',
      display_name: elements.createUserDisplayName?.value?.trim() || '',
      email: elements.createUserEmail?.value?.trim() || '',
      language: 'zh-Hant',
      status: elements.createUserStatus?.value || 'active',
      password
    };

    if (!payload.external_user_id || !payload.display_name || !payload.password) {
      showError('請填寫來源、帳號、暱稱與密碼');
      return;
    }

    if (elements.submitCreateUser) elements.submitCreateUser.disabled = true;
    try {
      const response = await makeAuthenticatedRequest('/api/system-admin/users', {
        method: 'POST',
        body: JSON.stringify(payload)
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '建立用戶失敗');
      }

      elements.userCreateForm?.reset();
      showSuccess('用戶已建立');
      switchView('users');
    } catch (err) {
      showError(err.message || '建立用戶失敗');
    } finally {
      if (elements.submitCreateUser) elements.submitCreateUser.disabled = false;
    }
  }

  async function inviteUser(event) {
    event.preventDefault();

    const payload = {
      email: elements.inviteUserEmail?.value?.trim() || ''
    };
    if (!payload.email) {
      showError('請填寫 Email');
      elements.inviteUserEmail?.focus();
      return;
    }

    if (elements.submitInviteUser) elements.submitInviteUser.disabled = true;
    try {
      const response = await makeAuthenticatedRequest('/api/system-admin/user-invitations', {
        method: 'POST',
        body: JSON.stringify(payload)
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '發送註冊邀請失敗');
      }

      closeUserInviteModal();
      userInvitationsPage = 1;
      await loadUserInvitations();
      showSuccess('註冊邀請已發送');
    } catch (err) {
      showError(err.message || '發送註冊邀請失敗');
    } finally {
      if (elements.submitInviteUser) elements.submitInviteUser.disabled = false;
    }
  }

  async function precheckBulkUserInvitations() {
    const file = elements.bulkInviteFile?.files?.[0];
    bulkInvitePrecheck = null;
    renderBulkInviteSummary(null);
    renderBulkInviteSendResult(null);
    updateBulkInviteSubmit(0, true);

    if (!file) {
      setBulkInviteStatus('', false);
      updateBulkInviteSubmit(0, false);
      return;
    }
    if (!file.name.toLowerCase().endsWith('.txt')) {
      setBulkInviteStatus('僅支援 .txt 檔案', true);
      updateBulkInviteSubmit(0, false);
      return;
    }
    if (file.size > 2 * 1024 * 1024) {
      setBulkInviteStatus('TXT 檔案不可超過 2MB', true);
      updateBulkInviteSubmit(0, false);
      return;
    }

    setBulkInviteStatus('正在預檢 TXT...', false);
    try {
      const response = await postBulkInviteFile('/api/system-admin/user-invitations/bulk-precheck', file);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '批量預檢失敗');
      }

      bulkInvitePrecheck = data.data || {};
      const sendableCount = Number(bulkInvitePrecheck.counts?.sendable) || 0;
      renderBulkInviteSummary(bulkInvitePrecheck);
      setBulkInviteStatus(sendableCount > 0 ? '預檢完成' : '沒有可發送的 Email', false);
      updateBulkInviteSubmit(sendableCount, false);
    } catch (err) {
      bulkInvitePrecheck = null;
      setBulkInviteStatus(err.message || '批量預檢失敗', true);
      updateBulkInviteSubmit(0, false);
    }
  }

  async function sendBulkUserInvitations(event) {
    event.preventDefault();
    const file = elements.bulkInviteFile?.files?.[0];
    const sendableCount = Number(bulkInvitePrecheck?.counts?.sendable) || 0;
    if (!file || sendableCount <= 0) return;

    const preSendPrecheck = bulkInvitePrecheck;
    updateBulkInviteSubmit(sendableCount, true);
    setBulkInviteStatus('正在發送邀請...', false);
    renderBulkInviteSendResult(null);
    try {
      const response = await postBulkInviteFile('/api/system-admin/user-invitations/bulk-send', file);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '批量發送失敗');
      }

      const result = data.data || {};
      bulkInvitePrecheck = preSendPrecheck;
      renderBulkInviteSummary(preSendPrecheck);
      renderBulkInviteSendResult(result, preSendPrecheck);
      setBulkInviteStatus('批量發送完成', false);
      markBulkInviteCompleted();
      userInvitationsPage = 1;
      await loadUserInvitations();
      showSuccess(bulkInviteSendResultMessage(result, preSendPrecheck));
    } catch (err) {
      setBulkInviteStatus(err.message || '批量發送失敗', true);
      updateBulkInviteSubmit(sendableCount, false);
    }
  }

  async function postBulkInviteFile(url, file) {
    const token = getSessionToken();
    if (!token) {
      logout();
      return null;
    }
    const formData = new FormData();
    formData.append('file', file);
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`
      },
      body: formData
    });
    if (response.status === 401) {
      logout();
      return null;
    }
    return response;
  }

  function updateBulkInviteSubmit(count, loading) {
    if (!elements.submitBulkInvite) return;
    elements.submitBulkInvite.disabled = loading || count <= 0;
    elements.submitBulkInvite.textContent = loading ? '處理中...' : `發送 ${count} 封邀請`;
  }

  function markBulkInviteCompleted() {
    if (!elements.submitBulkInvite) return;
    elements.submitBulkInvite.disabled = true;
    elements.submitBulkInvite.textContent = '已完成';
  }

  function setBulkInviteStatus(message, isError) {
    if (!elements.bulkInviteStatus) return;
    elements.bulkInviteStatus.textContent = message || '';
    elements.bulkInviteStatus.classList.toggle('is-error', Boolean(isError));
  }

  function renderBulkInviteSummary(precheck) {
    if (!elements.bulkInviteSummary) return;
    elements.bulkInviteSummary.replaceChildren();
    if (!precheck) {
      elements.bulkInviteSummary.hidden = true;
      return;
    }

    const counts = precheck.counts || {};
    const cards = [
      ['可發送', counts.sendable, 'sendable'],
      ['檔案內重複', counts.duplicate_in_file, 'duplicate_in_file'],
      ['已註冊', counts.already_registered, 'already_registered'],
      ['已有有效邀請', counts.active_invitation, 'active_invitation'],
      ['格式錯誤', counts.invalid_email, 'invalid_email']
    ];
    const grid = document.createElement('div');
    grid.className = 'bulk-invite-count-grid';
    cards.forEach(([label, value, key]) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'bulk-invite-count-card';
      button.textContent = `${label} ${Number(value) || 0}`;
      button.addEventListener('click', () => renderBulkInviteList(label, precheck[key] || []));
      grid.appendChild(button);
    });
    elements.bulkInviteSummary.appendChild(grid);
    renderBulkInviteList('可發送', precheck.sendable || []);
    elements.bulkInviteSummary.hidden = false;
  }

  function renderBulkInviteList(title, items) {
    if (!elements.bulkInviteResults) return;
    elements.bulkInviteResults.replaceChildren();
    const heading = document.createElement('strong');
    heading.textContent = title;
    elements.bulkInviteResults.appendChild(heading);

    const list = document.createElement('div');
    list.className = 'bulk-invite-list';
    const visibleItems = Array.isArray(items) ? items.slice(0, 80) : [];
    if (visibleItems.length === 0) {
      const empty = document.createElement('span');
      empty.className = 'bulk-invite-empty';
      empty.textContent = '沒有資料';
      list.appendChild(empty);
    } else {
      visibleItems.forEach(item => {
        const row = document.createElement('span');
        row.textContent = `第 ${item.line || '--'} 行 ${item.email || item.value || ''}`;
        list.appendChild(row);
      });
      if (items.length > visibleItems.length) {
        const more = document.createElement('span');
        more.className = 'bulk-invite-empty';
        more.textContent = `另有 ${items.length - visibleItems.length} 筆未顯示`;
        list.appendChild(more);
      }
    }
    elements.bulkInviteResults.appendChild(list);
    elements.bulkInviteResults.hidden = false;
  }

  function renderBulkInviteSendResult(result, preSendPrecheck) {
    if (!elements.bulkInviteResults || !result) return;
    elements.bulkInviteResults.replaceChildren();
    const changedBeforeSendCount = bulkInviteChangedBeforeSendCount(result, preSendPrecheck);
    const summary = document.createElement('strong');
    summary.textContent = `本次實際寄出：成功 ${Number(result.success_count) || 0}，失敗 ${Number(result.failure_count) || 0}`;
    elements.bulkInviteResults.appendChild(summary);
    if (changedBeforeSendCount > 0) {
      const note = document.createElement('span');
      note.className = 'bulk-invite-empty';
      note.textContent = `另外 ${changedBeforeSendCount} 筆在送出時已變成不可發送，未寄出邀請。`;
      elements.bulkInviteResults.appendChild(note);
    }
    if (Array.isArray(result.failed) && result.failed.length > 0) {
      const list = document.createElement('div');
      list.className = 'bulk-invite-list';
      result.failed.slice(0, 80).forEach(item => {
        const row = document.createElement('span');
        row.textContent = `第 ${item.line || '--'} 行 ${item.email || ''}：${item.message || item.code || '發送失敗'}`;
        list.appendChild(row);
      });
      elements.bulkInviteResults.appendChild(list);
    }
    elements.bulkInviteResults.hidden = false;
  }

  function bulkInviteSendResultMessage(result, preSendPrecheck) {
    const changedBeforeSendCount = bulkInviteChangedBeforeSendCount(result, preSendPrecheck);
    const suffix = changedBeforeSendCount > 0 ? `，未寄出 ${changedBeforeSendCount}` : '';
    return `批量發送完成：成功 ${Number(result.success_count) || 0}，失敗 ${Number(result.failure_count) || 0}${suffix}`;
  }

  function bulkInviteChangedBeforeSendCount(result, preSendPrecheck) {
    const initialIgnoredCount = preSendPrecheck ? bulkInviteIgnoredCount(preSendPrecheck) : 0;
    const ignoredCount = Number(result?.ignored_count) || 0;
    return ignoredCount > initialIgnoredCount ? ignoredCount - initialIgnoredCount : 0;
  }

  function bulkInviteIgnoredCount(precheck) {
    const counts = precheck?.counts || {};
    return (Number(counts.duplicate_in_file) || 0) +
      (Number(counts.already_registered) || 0) +
      (Number(counts.active_invitation) || 0) +
      (Number(counts.invalid_email) || 0);
  }

  function confirmResendUserInvitation(invitation, button) {
    const email = invitation.email || '';
    if (!email) return;

    openConfirmModal(
      '重新發送註冊邀請',
      [
        '確定要重新發送註冊邀請給 ',
        { text: email, className: 'confirm-modal-email' },
        '？'
      ],
      () => resendUserInvitation(email, button),
      '確認'
    );
  }

  async function resendUserInvitation(email, button) {
    if (button) button.disabled = true;
    try {
      const response = await makeAuthenticatedRequest('/api/system-admin/user-invitations/resend', {
        method: 'POST',
        body: JSON.stringify({ email })
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '重新發送註冊邀請失敗');
      }

      await loadUserInvitations();
      showSuccess('註冊邀請已重新發送');
    } catch (err) {
      showError(err.message || '重新發送註冊邀請失敗');
      throw err;
    } finally {
      if (button) button.disabled = false;
    }
  }

  function openUserEdit(userID) {
    editingUserID = String(userID);
    switchView('userEdit', { path: `/office/user/${encodeURIComponent(editingUserID)}/edit` });
  }

  async function loadUserForEdit() {
    if (!editingUserID || !elements.userEditForm) return;

    if (elements.submitEditUser) elements.submitEditUser.disabled = true;
    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/users/${encodeURIComponent(editingUserID)}`);
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '載入用戶資料失敗');
      }

      const user = data.data || {};
      elements.editUserAccount.value = user.external_user_id || '';
      originalEditableUser = {
        password: '',
        displayName: user.display_name || '',
        email: user.email || '',
        status: user.status || 'active'
      };
      resetUserEditForm();
      elements.editUserDisplayName.focus();
    } catch (err) {
      showError(err.message || '載入用戶資料失敗');
      switchView('users');
    } finally {
      if (elements.submitEditUser) elements.submitEditUser.disabled = false;
    }
  }

  function resetUserEditForm() {
    if (!originalEditableUser) return;

    const editableFields = [
      [elements.editUserPassword, originalEditableUser.password],
      [elements.editUserDisplayName, originalEditableUser.displayName],
      [elements.editUserEmail, originalEditableUser.email],
      [elements.editUserStatus, originalEditableUser.status]
    ];
    editableFields.forEach(([field, originalValue]) => {
      if (field && field.value !== originalValue) {
        field.value = originalValue;
      }
    });
  }

  async function updateUser(event) {
    event.preventDefault();
    if (!editingUserID) return;

    const password = elements.editUserPassword?.value || '';
    if (password && /\s/.test(password)) {
      showError('密碼不可包含空白');
      elements.editUserPassword?.focus();
      return;
    }

    const payload = {
      password,
      display_name: elements.editUserDisplayName?.value?.trim() || '',
      email: elements.editUserEmail?.value?.trim() || '',
      status: elements.editUserStatus?.value || 'active'
    };
    if (!payload.display_name) {
      showError('請填寫暱稱');
      return;
    }

    if (elements.submitEditUser) elements.submitEditUser.disabled = true;
    try {
      const response = await makeAuthenticatedRequest(`/api/system-admin/users/${encodeURIComponent(editingUserID)}`, {
        method: 'PATCH',
        body: JSON.stringify(payload)
      });
      if (!response) return;

      const data = await readJSONResponse(response);
      if (!response.ok || !data.success) {
        throw new Error(data.message || '更新用戶失敗');
      }

      elements.editUserPassword.value = '';
      showSuccess('用戶資料已更新');
      switchView('users');
    } catch (err) {
      showError(err.message || '更新用戶失敗');
    } finally {
      if (elements.submitEditUser) elements.submitEditUser.disabled = false;
    }
  }

  /**
   * Show users empty state
   */
  function showUsersEmpty() {
    if (elements.usersLoading) elements.usersLoading.hidden = true;
    if (elements.usersEmpty) elements.usersEmpty.hidden = false;
    if (elements.usersTable) elements.usersTable.hidden = true;
  }

  /**
   * Search devices for a user
   */
  async function searchDevices() {
    const userId = elements.deviceUserId?.value?.trim();
    if (!userId) {
      showError('請輸入用戶 ID');
      return;
    }

    if (elements.devicesLoading) elements.devicesLoading.hidden = false;
    if (elements.devicesTable) elements.devicesTable.hidden = true;
    if (elements.devicesEmpty) elements.devicesEmpty.hidden = true;

    try {
      const response = await makeAuthenticatedRequest(
        `/api/system-admin/users/${encodeURIComponent(userId)}/devices`
      );

      if (!response) return;

      const data = await readJSONResponse(response);

      if (!data.success) {
        showError(data.message || '載入設備失敗');
        showDevicesEmpty();
        return;
      }

      const devices = data.data || [];
      if (devices.length === 0) {
        showDevicesEmpty();
        return;
      }

      populateDevicesTable(devices);
      if (elements.devicesTable) elements.devicesTable.hidden = false;
    } catch (err) {
      showError('搜尋設備失敗: ' + err.message);
      showDevicesEmpty();
    } finally {
      if (elements.devicesLoading) elements.devicesLoading.hidden = true;
    }
  }

  /**
   * Populate devices table
   */
  function populateDevicesTable(devices) {
    if (!elements.devicesTbody) return;

    elements.devicesTbody.innerHTML = '';
    devices.forEach(device => {
      const row = document.createElement('tr');
      row.innerHTML = `
        <td style="font-size: 0.85rem; font-family: monospace;" title="${device.device_id}">${truncateId(device.device_id)}</td>
        <td>${truncateText(device.user_agent, 40)}</td>
        <td>${device.ip || '--'}</td>
        <td>${formatDate(device.first_login_at)}</td>
        <td>${formatDate(device.last_login_at)}</td>
        <td>
          <span class="status-badge ${device.is_trusted_device ? 'status-trusted' : 'status-pending'}">
            ${device.is_trusted_device ? '已信任' : '待核准'}
          </span>
        </td>
        <td>
          ${!device.is_trusted_device ? `
            <button class="btn btn-small btn-success" onclick="approveDeviceFromTable('${device.device_id}')">
              核准
            </button>
          ` : ''}
        </td>
      `;
      elements.devicesTbody.appendChild(row);
    });
  }

  /**
   * Show devices empty state
   */
  function showDevicesEmpty() {
    if (elements.devicesLoading) elements.devicesLoading.hidden = true;
    if (elements.devicesEmpty) elements.devicesEmpty.hidden = false;
    if (elements.devicesTable) elements.devicesTable.hidden = true;
  }

  /**
   * Load whitelist settings for a user
   */
  async function loadWhitelist() {
    const userId = elements.whitelistUserId?.value?.trim();
    if (!userId) {
      showError('請輸入用戶 ID');
      return;
    }

    if (elements.whitelistLoading) elements.whitelistLoading.hidden = false;
    if (elements.whitelistContent) elements.whitelistContent.hidden = true;

    try {
      // TODO: Get current whitelist settings
      // For now, show a form with sample data
      showWhitelistForm({ allow_all: false, rules: [] });
    } catch (err) {
      showError('載入白名單設定失敗: ' + err.message);
    } finally {
      if (elements.whitelistLoading) elements.whitelistLoading.hidden = true;
    }
  }

  /**
   * Show whitelist edit form
   */
  function showWhitelistForm(whitelist) {
    if (elements.whitelistContent) elements.whitelistContent.hidden = false;
    
    if (elements.allowAllToggle) {
      elements.allowAllToggle.checked = whitelist.allow_all;
    }
    
    toggleRulesContainer();
    
    if (elements.rulesList) {
      elements.rulesList.innerHTML = '';
      (whitelist.rules || []).forEach((rule, index) => {
        addRuleRow(rule, index);
      });
    }
  }

  /**
   * Toggle rules container visibility
   */
  function toggleRulesContainer() {
    const showRules = !elements.allowAllToggle?.checked;
    if (elements.rulesContainer) {
      elements.rulesContainer.hidden = !showRules;
    }
    if (elements.addRuleBtn && showRules) {
      elements.addRuleBtn.hidden = false;
    }
  }

  /**
   * Add a rule row to the form
   */
  function addRuleRow(rule = '', index = null) {
    if (!elements.rulesList) return;

    const row = document.createElement('div');
    row.style.display = 'flex';
    row.style.gap = '8px';
    row.style.marginBottom = '8px';
    row.style.alignItems = 'stretch';

    const input = document.createElement('input');
    input.type = 'text';
    input.placeholder = '例如: 192.168.1.0/24, 2001:db8::0/32';
    input.value = rule;
    input.className = 'rule-input';
    input.style.flex = '1';

    const deleteBtn = document.createElement('button');
    deleteBtn.type = 'button';
    deleteBtn.className = 'btn btn-small btn-danger';
    deleteBtn.textContent = '移除';
    deleteBtn.onclick = () => row.remove();

    row.appendChild(input);
    row.appendChild(deleteBtn);
    elements.rulesList.appendChild(row);
  }

  /**
   * Save whitelist settings
   */
  async function saveWhitelist() {
    const userId = elements.whitelistUserId?.value?.trim();
    if (!userId) {
      showError('請輸入用戶 ID');
      return;
    }

    const allowAll = elements.allowAllToggle?.checked || false;
    const rules = Array.from(document.querySelectorAll('.rule-input'))
      .map(input => input.value.trim())
      .filter(rule => rule);

    if (!allowAll && rules.length === 0) {
      showError('請至少新增一條 IP 規則或啟用「允許所有 IP」');
      return;
    }

    try {
      const response = await makeAuthenticatedRequest(
        `/api/system-admin/users/${encodeURIComponent(userId)}/ip-whitelist`,
        {
          method: 'PUT',
          body: JSON.stringify({
            allow_all: allowAll,
            rules: rules
          })
        }
      );

      if (!response) return;

      const data = await readJSONResponse(response);

      if (!data.success) {
        showError(data.message || '保存失敗');
        return;
      }

      showSuccess('IP 白名單已成功更新');
      setTimeout(() => switchView('dashboard'), 1500);
    } catch (err) {
      showError('保存失敗: ' + err.message);
    }
  }

  /**
   * Load operation logs
   */
  async function loadLogs() {
    if (elements.logsList) {
      elements.logsList.innerHTML = '<div class="is-empty"><p>載入作業紀錄中...</p></div>';
    }

    try {
      // TODO: Implement logs API
      if (elements.logsEmpty) elements.logsEmpty.hidden = false;
    } catch (err) {
      showError('載入日誌失敗: ' + err.message);
    }
  }

  /**
   * Filter logs
   */
  function filterLogs() {
    // TODO: Implement filter logic
    loadLogs();
  }

  /**
   * Load health status
   */
  async function loadHealthStatus() {
    try {
      const response = await fetch('/healthz');
      const now = new Date();

      if (response.ok) {
        if (elements.healthApiTime) {
          elements.healthApiTime.textContent = formatDate(now);
        }
      }

      // Update other health metrics
      if (elements.healthVersion) elements.healthVersion.textContent = '1.0.0';
      if (elements.healthUptime) elements.healthUptime.textContent = '持續運行';
      if (elements.healthMemory) elements.healthMemory.textContent = '--';
      if (elements.healthConcurrent) elements.healthConcurrent.textContent = '--';
    } catch (err) {
      console.error('Health check failed:', err);
    }
  }

  /**
   * Utility: Show error message
   */
  function showError(message) {
    // Create a temporary alert
    const alert = document.createElement('div');
    alert.className = 'alert alert-danger';
    alert.innerHTML = `<strong>錯誤</strong><div>${message}</div>`;
    alert.style.position = 'fixed';
    alert.style.top = '20px';
    alert.style.right = '20px';
    alert.style.maxWidth = '400px';
    alert.style.zIndex = '2000';
    alert.style.animation = 'slideIn 0.3s ease';

    document.body.appendChild(alert);

    setTimeout(() => {
      alert.style.animation = 'slideOut 0.3s ease';
      setTimeout(() => alert.remove(), 300);
    }, 5000);
  }

  /**
   * Utility: Show success message
   */
  function showSuccess(message) {
    const alert = document.createElement('div');
    alert.className = 'alert alert-success';
    alert.innerHTML = `<strong>成功</strong><div>${message}</div>`;
    alert.style.position = 'fixed';
    alert.style.top = '20px';
    alert.style.right = '20px';
    alert.style.maxWidth = '400px';
    alert.style.zIndex = '2000';
    alert.style.animation = 'slideIn 0.3s ease';

    document.body.appendChild(alert);

    setTimeout(() => {
      alert.style.animation = 'slideOut 0.3s ease';
      setTimeout(() => alert.remove(), 300);
    }, 3000);
  }

  /**
   * Utility: Close modal
   */
  function closeModal() {
    if (elements.modal) {
      elements.modal.classList.remove('is-active', 'confirm-modal');
      elements.modal.hidden = true;
    }
    modalConfirmHandler = null;
    if (elements.modalConfirmBtn) {
      elements.modalConfirmBtn.disabled = false;
      elements.modalConfirmBtn.textContent = '確認';
    }
  }

  function openConfirmModal(title, message, onConfirm, confirmText = '確認') {
    modalConfirmHandler = onConfirm;
    if (elements.modalTitle) {
      elements.modalTitle.textContent = title;
    }
    if (elements.modalBody) {
      elements.modalBody.replaceChildren();
      const messageEl = document.createElement('p');
      messageEl.className = 'confirm-modal-message';
      const messageParts = Array.isArray(message) ? message : [message];
      messageParts.forEach((part) => {
        const span = document.createElement('span');
        if (part && typeof part === 'object') {
          span.textContent = part.text || '';
          if (part.className) span.className = part.className;
        } else {
          span.textContent = String(part ?? '');
        }
        messageEl.appendChild(span);
      });
      elements.modalBody.appendChild(messageEl);
    }
    if (elements.modalConfirmBtn) {
      elements.modalConfirmBtn.disabled = false;
      elements.modalConfirmBtn.textContent = confirmText;
    }
    if (elements.modal) {
      elements.modal.hidden = false;
      elements.modal.classList.add('is-active', 'confirm-modal');
    }
  }

  /**
   * Utility: Debounce function
   */
  function debounce(func, delay) {
    let timeoutId;
    return function(...args) {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => func.apply(this, args), delay);
    };
  }

  /**
   * Utility: Format date
   */
  function formatDate(dateStr) {
    if (!dateStr) return '--';
    try {
      const date = new Date(dateStr);
      if (Number.isNaN(date.getTime())) return '--';
      const pad = value => String(value).padStart(2, '0');
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
    } catch {
      return '--';
    }
  }

  async function readJSONResponse(response) {
    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      return response.json();
    }

    const body = await response.text();
    const detail = body.trim() || `HTTP ${response.status}`;
    throw new Error(`API 回應不是 JSON（${detail}）`);
  }

  /**
   * Utility: Truncate ID
   */
  function truncateId(id) {
    return id?.substring(0, 8) + '...' || '--';
  }

  /**
   * Utility: Truncate text
   */
  function truncateText(text, maxLen) {
    if (!text) return '--';
    return text.length > maxLen ? text.substring(0, maxLen) + '...' : text;
  }

  // Add CSS for animations
  const style = document.createElement('style');
  style.textContent = `
    @keyframes slideIn {
      from { transform: translateX(400px); opacity: 0; }
      to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
      from { transform: translateX(0); opacity: 1; }
      to { transform: translateX(400px); opacity: 0; }
    }
  `;
  document.head.appendChild(style);

  // Expose functions to global scope
  window.adminApp = {
    switchView,
    approveDeviceFromTable: () => showError('功能開發中'),
    closeModal
  };

  // Initialize app when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initApp);
  } else {
    initApp();
  }
})();
