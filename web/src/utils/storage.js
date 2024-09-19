/**
 * LocalStorage utility functions for managing persisted data
 */

const USER_STORAGE_KEY = 'met_user'
const WORKSPACE_STORAGE_KEY = 'met_workspace'

/**
 * Load user data from localStorage
 * @returns {Object|null} The stored user object or null if not found/invalid
 */
export const loadUserFromStorage = () => {
  try {
    const stored = localStorage.getItem(USER_STORAGE_KEY)
    return stored ? JSON.parse(stored) : null
  } catch (err) {
    console.error('Failed to load user from localStorage:', err)
    return null
  }
}

/**
 * Save user data to localStorage
 * @param {Object|null} userData - The user object to store, or null to clear
 */
export const saveUserToStorage = (userData) => {
  try {
    if (userData) {
      localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(userData))
    } else {
      localStorage.removeItem(USER_STORAGE_KEY)
    }
  } catch (err) {
    console.error('Failed to save user to localStorage:', err)
  }
}

/**
 * Clear all user-related data from localStorage
 */
export const clearUserStorage = () => {
  try {
    localStorage.removeItem(USER_STORAGE_KEY)
  } catch (err) {
    console.error('Failed to clear user from localStorage:', err)
  }
}

/**
 * Load current workspace from localStorage
 * @returns {Object|null} The stored workspace object or null
 */
export const loadWorkspaceFromStorage = () => {
  try {
    const stored = localStorage.getItem(WORKSPACE_STORAGE_KEY)
    return stored ? JSON.parse(stored) : null
  } catch (err) {
    console.error('Failed to load workspace from localStorage:', err)
    return null
  }
}

/**
 * Save current workspace to localStorage
 * @param {Object|null} workspace - The workspace object to store, or null to clear
 */
export const saveWorkspaceToStorage = (workspace) => {
  try {
    if (workspace) {
      localStorage.setItem(WORKSPACE_STORAGE_KEY, JSON.stringify(workspace))
    } else {
      localStorage.removeItem(WORKSPACE_STORAGE_KEY)
    }
  } catch (err) {
    console.error('Failed to save workspace to localStorage:', err)
  }
}

