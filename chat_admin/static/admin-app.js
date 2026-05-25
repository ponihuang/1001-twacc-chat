/**
 * Admin Dashboard Application
 * Main application logic for admin management interface
 */

(function() {
  'use strict';

  const { getSessionToken, getUserId, logout, makeAuthenticatedRequest, STORAGE_KEYS } = window.adminAuth;

  // DOM Elements
  let elements = {};

  /**
   * Initialize the app
   */
  async function initApp() {
    cacheElements();
    setupEventListeners();
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
      userSearch: document.getElementById('user-search'),
      usersTable: document.getElementById('users-table'),
      usersTbody: document.getElementById('users-tbody'),
      usersLoading: document.getElementById('users-loading'),
      usersEmpty: document.getElementById('users-empty'),
      refreshUsersBtn: document.getElementById('refresh-users-btn'),
      
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
        const menuItem = e.target.closest('[data-menu-item]');
        if (menuItem) {
          switchView(menuItem.getAttribute('data-menu-item'));
        }
      });
    }

    // User management
    if (elements.userSearch) {
      elements.userSearch.addEventListener('input', debounce(searchUsers, 300));
    }
    if (elements.refreshUsersBtn) {
      elements.refreshUsersBtn.addEventListener('click', loadUsers);
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

    // Logout
    if (elements.logoutButton) {
      elements.logoutButton.addEventListener('click', (e) => {
        e.preventDefault();
        logout();
      });
    }
  }

  /**
   * Switch between views
   */
  function switchView(viewName) {
    // Update menu items
    document.querySelectorAll('[data-menu-item]').forEach(item => {
      item.classList.toggle('is-active', item.getAttribute('data-menu-item') === viewName);
    });

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
    if (elements.usersTable) elements.usersTable.hidden = true;
    if (elements.usersEmpty) elements.usersEmpty.hidden = true;

    try {
      // TODO: Implement actual API call to fetch users
      // This would call an endpoint like GET /api/system-admin/users
      
      // For now, show empty state
      showUsersEmpty();
    } catch (err) {
      showError('載入使用者列表失敗: ' + err.message);
      showUsersEmpty();
    }
  }

  /**
   * Search users
   */
  async function searchUsers() {
    const query = elements.userSearch?.value?.trim() || '';
    if (!query) {
      loadUsers();
      return;
    }

    if (elements.usersLoading) elements.usersLoading.hidden = false;
    if (elements.usersTable) elements.usersTable.hidden = true;

    try {
      // TODO: Implement actual search API call
      showUsersEmpty();
    } catch (err) {
      showError('搜尋使用者失敗: ' + err.message);
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
      showError('請輸入使用者 ID');
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

      const data = await response.json();

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
      showError('請輸入使用者 ID');
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
      showError('請輸入使用者 ID');
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

      const data = await response.json();

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
          elements.healthApiTime.textContent = now.toLocaleTimeString();
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
      return new Date(dateStr).toLocaleString('zh-TW');
    } catch {
      return '--';
    }
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
