import axios from 'axios'
import { removeUserFromStorage, removeWorkspaceFromStorage } from '@/utils/storage'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

api.interceptors.request.use((config) => {
  try {
    const locale = typeof localStorage !== 'undefined' && localStorage.getItem('finn-locale')
    if (locale === 'en' || locale === 'fr') {
      config.headers['Accept-Language'] = locale
    }
  } catch (_) {}
  return config
})

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Only redirect to login if:
    // 1. It's a 401 error
    // 2. NOT from the login endpoint
    // 3. NOT from the profile endpoint (used to check auth on page load)
    // 4. NOT from the logout endpoint (session may already be expired)
    // 5. NOT already on the login or setup page
    const isLoginEndpoint = error.config?.url === 'public/action/login'
    const isProfileEndpoint = error.config?.url === '/action/profile'
    const isLogoutEndpoint = error.config?.url === 'public/action/logout'
    const isForgotPasswordEndpoint = error.config?.url === 'public/action/forgot-password'
    const isResetPasswordEndpoint = error.config?.url === 'public/action/reset-password'
    const isOnPublicPage = window.location.pathname === '/login' ||
                           window.location.pathname === '/setup' ||
                           window.location.pathname.startsWith('/forgot-password') ||
                           window.location.pathname.startsWith('/reset-password') ||
                           window.location.pathname === '/register'

    if (error.response?.status === 401 &&
        !isLoginEndpoint &&
        !isProfileEndpoint &&
        !isLogoutEndpoint &&
        !isForgotPasswordEndpoint &&
        !isResetPasswordEndpoint &&
        !isOnPublicPage) {
      removeUserFromStorage()
      removeWorkspaceFromStorage()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

//export default api

// API endpoints
export const healthAPI = {
  check: () => api.get('public/_health'),
  ready: () => api.get('public/_ready'),
}

export const authAPI = {
  login: (data) => api.post('public/action/login', data),
  logout: () => api.post('public/action/logout'),
  getProfile: () => api.get('/action/profile'),
  updateProfile: (data) => api.put('/action/profile', data),
  forgotPassword: (data) => api.post('public/action/forgot-password', data),
  resetPassword: (data) => api.post('public/action/reset-password', data),
  register: (data) => api.post('public/action/register', data),
  getInviteByToken: (token) => api.get(`public/action/invite/token/${encodeURIComponent(token)}`),
}

export const setupAPI = {
  install: (data) => api.post('public/action/setup', data),
  checkInstalled: () => api.get('public/action/setup/status'),
}

// API endpoints for users
export const userAPI = {
  getUsers: (params = {}) => api.get('/users', { params }),
  getUser: (id) => api.get(`/users/${id}`),
  createUser: (data) => api.post('/users', data),
  updateUser: (id, data) => api.put(`/users/${id}`, data),
  deleteUser: (id) => api.delete(`/users/${id}`),
}

// API endpoints for user invites
export const invitesAPI = {
  list: () => api.get('/invites'),
  get: (id) => api.get(`/invites/${id}`),
  create: (data) => api.post('/invites', data),
  delete: (id) => api.delete(`/invites/${id}`),
}

// API endpoints for runners (platform admin only)
export const runnersAPI = {
  list: (params = {}) => api.get('/runners', { params }),
  get: (id) => api.get(`/runners/${id}`),
  create: (data) => api.post('/runners', data),
  update: (id, data) => api.put(`/runners/${id}`, data),
  delete: (id) => api.delete(`/runners/${id}`),
  ping: (id) => api.post(`/runners/${id}/ping`),
}

// API endpoints for settings
export const settingsAPI = {
  getSettings: () => api.get('/action/settings'),
  updateSettings: (data) => api.put('/action/settings', data),
}

// API endpoints for workspaces (projects)
export const workspaceAPI = {
  list: (params) => api.get('/workspaces', { params }),
  get: (id) => api.get(`/workspaces/${id}`),
  create: (data) => api.post('/workspaces', data),
  update: (id, data) => api.put(`/workspaces/${id}`, data),
  delete: (id) => api.delete(`/workspaces/${id}`),
}

// API endpoints for workspace users (members)
export const workspaceUsersAPI = {
  list: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/users`, { params }),
  get: (workspaceId, userId) => api.get(`/workspaces/${workspaceId}/users/${userId}`),
  add: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/users`, data),
  updateRole: (workspaceId, userId, data) => api.put(`/workspaces/${workspaceId}/users/${userId}`, data),
  remove: (workspaceId, userId) => api.delete(`/workspaces/${workspaceId}/users/${userId}`),
  listAllUsers: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/all-users`, { params }),
}

// API endpoints for current user's API keys
export const apiKeysAPI = {
  list: () => api.get('/apiKeys'),
  create: (data) => api.post('/apiKeys', data),
  delete: (id) => api.delete(`/apiKeys/${id}`),
}

// API endpoints for workspace repositories
export const repositoriesAPI = {
  list: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/repositories`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/repositories/${id}`),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/repositories`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/repositories/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/repositories/${id}`),
  clone: (workspaceId, id) => api.post(`/workspaces/${workspaceId}/repositories/${id}/clone`),
  getCloneTask: (workspaceId, repoId, taskId) => api.get(`/workspaces/${workspaceId}/repositories/${repoId}/tasks/${taskId}`),
}

// API endpoints for workspace SSH keys
export const sshKeysAPI = {
  list: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/keys`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/keys/${id}`),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/keys`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/keys/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/keys/${id}`),
}

// API endpoints for workspace git credentials
export const credentialsAPI = {
  list: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/credentials`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/credentials/${id}`),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/credentials`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/credentials/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/credentials/${id}`),
}

// API endpoints for workspace Ansible inventories
export const inventoriesAPI = {
  list: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/inventories`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/inventories/${id}`),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/inventories`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/inventories/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/inventories/${id}`),
}

// API endpoints for workspace variable groups (extra vars / env / secrets)
export const variableGroupsAPI = {
  list: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/gvars`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/gvars/${id}`),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/gvars`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/gvars/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/gvars/${id}`),
}

// API endpoints for workspace backends (Terraform state backends)
export const backendsAPI = {
  list: (workspaceId, params = {}) => api.get(`/workspaces/${workspaceId}/backends`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/backends/${id}`),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/backends`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/backends/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/backends/${id}`),
  revealPassword: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/backends/${id}/password`),
  unlock: (workspaceId, id) => api.post(`/workspaces/${workspaceId}/backends/${id}/unlock`),
}

// API endpoints for AI assistant
export const aiAPI = {
  getStatus: () => api.get('/ai/status'),
  getMessages: (params = {}) => api.get('/ai/messages', { params }),
  createMessage: (data) => api.post('/ai/messages', data),
}
