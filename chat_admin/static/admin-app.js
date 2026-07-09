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
  let adminsPage = 1;
  let adminsPerPage = 10;

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
    { key: 'options', title: '選項', width: 120 }
  ];

  const VIEW_ROUTES = {
    dashboard: '/office',
    users: '/office/user',
    userCreate: '/office/user/create',
    userInvitations: '/office/user/invitations',
    userEdit: null,
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
    switchView(resolveInitialView(), { replace: true });
    await loadInitialData();
  }

  /**
   * Cache DOM elements for performance
   */
  function cacheElements() {
    elements = {
      // Navigation
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
      userInviteModal: document.getElementById('user-invite-modal'),
      closeUserInviteModal: document.getElementById('close-user-invite-modal'),
      userInviteForm: document.getElementById('user-invite-form'),
      inviteUserEmail: document.getElementById('invite-user-email'),
      submitInviteUser: document.getElementById('submit-invite-user'),
      userEditForm: document.getElementById('user-edit-form'),
      editUserAccount: document.getElementById('edit-user-account'),
      editUserPassword: document.getElementById('edit-user-password'),
      editUserDisplayName: document.getElementById('edit-user-display-name'),
      editUserEmail: document.getElementById('edit-user-email'),
      editUserStatus: document.getElementById('edit-user-status'),
      submitEditUser: document.getElementById('submit-edit-user'),
      resetEditUser: document.getElementById('reset-edit-user'),

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
      
      // Logout
      logoutButton: document.getElementById('logout-button')
    };
  }

  /**
   * Setup event listeners
   */
  function setupEventListeners() {
    // Menu items
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
    if (elements.closeUserInviteModal) {
      elements.closeUserInviteModal.addEventListener('click', closeUserInviteModal);
    }
    if (elements.userInviteModal) {
      elements.userInviteModal.addEventListener('click', (e) => {
        if (e.target === elements.userInviteModal) closeUserInviteModal();
      });
    }
    if (elements.userInviteForm) {
      elements.userInviteForm.addEventListener('submit', inviteUser);
    }
    if (elements.userEditForm) {
      elements.userEditForm.addEventListener('submit', updateUser);
    }
    if (elements.resetEditUser) {
      elements.resetEditUser.addEventListener('click', resetUserEditForm);
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
      elements.modalConfirmBtn.addEventListener('click', closeModal);
    }

    // Logout
    if (elements.logoutButton) {
      elements.logoutButton.addEventListener('click', (e) => {
        e.preventDefault();
        logout();
      });
    }
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

  /**
   * Switch between views
   */
  function switchView(viewName, options = {}) {
    if (!Object.prototype.hasOwnProperty.call(VIEW_ROUTES, viewName)) {
      viewName = 'dashboard';
    }

    if (!options.skipHistory) {
      const nextPath = options.path || VIEW_ROUTES[viewName] || (
        viewName === 'userEdit' && editingUserID ? `/office/user/${editingUserID}/edit` :
        viewName === 'adminEdit' && editingAdminUserID ? `/office/admins/${editingAdminUserID}/edit` : ''
      );
      if (nextPath && window.location.pathname !== nextPath) {
        const method = options.replace ? 'replaceState' : 'pushState';
        window.history[method]({ adminView: viewName }, '', nextPath);
      }
    }

    // Update menu items
    const activeMenuItem = viewName === 'userCreate' || viewName === 'userEdit'
      ? 'users'
      : viewName === 'adminCreate' || viewName === 'adminEdit' ? 'admins' : viewName;
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
    if (email) params.set('email', email);
    if (status) params.set('status', status);
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
        invitationInviter(invitation)
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
        } else {
          cell.textContent = value;
        }
        row.appendChild(cell);
      });
      elements.userInvitationsTbody.appendChild(row);
    });
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
      elements.modal.classList.remove('is-active');
      elements.modal.hidden = true;
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
