/**
 * Admin Authentication Module
 * Handles authentication checks and redirects for admin pages
 */

(function() {
  'use strict';

  const STORAGE_KEYS = {
    SESSION_TOKEN: 'admin_session_token',
    USER_ID: 'admin_user_id',
    EXTERNAL_USER_ID: 'admin_external_user_id'
  };

  /**
   * Get the current session token
   */
  function getSessionToken() {
    return localStorage.getItem(STORAGE_KEYS.SESSION_TOKEN);
  }

  /**
   * Get the current user ID
   */
  function getUserId() {
    return localStorage.getItem(STORAGE_KEYS.USER_ID);
  }

  /**
   * Check if user is authenticated
   */
  function isAuthenticated() {
    return !!getSessionToken() && !!getUserId();
  }

  /**
   * Logout the current user
   */
  function logout() {
    // Clear all admin-related session data
    Object.values(STORAGE_KEYS).forEach(key => {
      localStorage.removeItem(key);
    });
    window.location.href = '/admin/login';
  }

  /**
   * Make an authenticated API request
   */
  async function makeAuthenticatedRequest(url, options = {}) {
    const token = getSessionToken();
    if (!token) {
      throw new Error('Not authenticated');
    }

    const headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options.headers
    };

    const response = await fetch(url, {
      ...options,
      headers
    });

    if (response.status === 401) {
      // Session expired
      logout();
      return null;
    }

    return response;
  }

  /**
   * Initialize auth guard
   */
  function initAuthGuard() {
    // Check if page requires authentication
    const pageElement = document.querySelector('[data-admin-app]');
    if (!pageElement) return;

    const redirectUrl = pageElement.getAttribute('data-require-auth-redirect');

    if (!isAuthenticated()) {
      if (redirectUrl) {
        window.location.href = redirectUrl;
      }
      return;
    }

    // Set up logout button
    const logoutButton = document.getElementById('logout-button');
    if (logoutButton) {
      logoutButton.addEventListener('click', (e) => {
        e.preventDefault();
        logout();
      });
    }

    // Update header with admin name
    const adminNameEl = document.getElementById('header-admin-name');
    if (adminNameEl) {
      const externalUserId = localStorage.getItem(STORAGE_KEYS.EXTERNAL_USER_ID);
      if (externalUserId) {
        adminNameEl.textContent = externalUserId;
      }
    }
  }

  /**
   * Set up global error handler for API responses
   */
  function setupGlobalErrorHandler() {
    // This will be enhanced in admin-app.js
  }

  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initAuthGuard);
  } else {
    initAuthGuard();
  }

  // Expose to global scope for other modules
  window.adminAuth = {
    getSessionToken,
    getUserId,
    isAuthenticated,
    logout,
    makeAuthenticatedRequest,
    STORAGE_KEYS
  };
})();
