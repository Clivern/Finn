import { createRouter, createWebHistory } from 'vue-router'
import { loadUserFromStorage, loadWorkspaceFromStorage } from '@/utils/storage'

const routes = [
  { path: '/', redirect: '/setup' },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresGuest: true, title: 'Login', description: 'Login to platform' }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('@/views/ForgotPassword.vue'),
    meta: { requiresGuest: true, title: 'Forgot password', description: 'Request password reset' }
  },
  {
    path: '/reset-password/:token',
    name: 'ResetPassword',
    component: () => import('@/views/ResetPassword.vue'),
    meta: { requiresGuest: true, title: 'Reset password', description: 'Set new password' }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { requiresGuest: true, title: 'Register', description: 'Create an account' }
  },
  {
    path: '/setup',
    name: 'Setup',
    component: () => import('@/views/Setup.vue'),
    meta: { requiresGuest: true, title: 'Setup', description: 'Setup platform' }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Dashboard', description: 'Platform dashboard' }
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/Users.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Users', description: 'Manage users' }
  },
  {
    path: '/invites',
    name: 'Invites',
    component: () => import('@/views/Invites.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Invites', description: 'Manage user invites' }
  },
  {
    path: '/manage',
    name: 'Manage',
    component: () => import('@/views/Manage.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Manage', description: 'Manage platform' }
  },
  {
    path: '/select-workspace',
    name: 'SwitchWorkspaces',
    component: () => import('@/views/SelectWorkspace.vue'),
    meta: { requiresAuth: true, title: 'Switch Workspaces', description: 'Choose or create a workspace' }
  },
  {
    path: '/workspaces',
    name: 'Workspaces',
    component: () => import('@/views/Workspaces.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Workspaces', description: 'Manage all platform workspaces' }
  },
  {
    path: '/workspace',
    name: 'Workspace',
    component: () => import('@/views/Workspace.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Workspace settings', description: 'Workspace name, members, and delete' }
  },
  {
    path: '/templates',
    name: 'TaskTemplates',
    component: () => import('@/views/TaskTemplates.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Task Templates', description: 'Manage task templates' }
  },
  {
    path: '/flows',
    name: 'Flows',
    component: () => import('@/views/Flows.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Flows', description: 'Build flows visually' }
  },
  {
    path: '/inventory',
    name: 'Inventory',
    component: () => import('@/views/Inventory.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Inventory', description: 'Manage inventory' }
  },
  {
    path: '/group-vars',
    name: 'GroupVars',
    component: () => import('@/views/GroupVars.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Group vars', description: 'Group vars configuration' }
  },
  {
    path: '/repositories',
    name: 'Repositories',
    component: () => import('@/views/Repositories.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Repositories', description: 'Manage repositories' }
  },
  {
    path: '/secrets',
    name: 'Secrets',
    component: () => import('@/views/Secrets.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Secrets', description: 'SSH keys and Git credentials' }
  },
  {
    path: '/backends',
    name: 'Backends',
    component: () => import('@/views/Backends.vue'),
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'TF Backends', description: 'Terraform backends' }
  },
  {
    path: '/runners',
    name: 'Runners',
    component: () => import('@/views/Runners.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Runners', description: 'Manage runners' }
  },
  {
    path: '/integrations',
    name: 'Integrations',
    component: () => import('@/views/Integrations.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Integrations', description: 'Manage integrations' }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/Profile.vue'),
    meta: { requiresAuth: true, title: 'Profile', description: 'Your account profile' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/Settings.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Settings', description: 'Platform settings' }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: 'Page not found' }
  },
  {
    path: '/500',
    name: 'ServerError',
    component: () => import('@/views/ServerError.vue'),
    meta: { title: 'Server error' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFoundCatchAll',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: 'Page not found' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard - reads auth state directly from storage
router.beforeEach((to, from, next) => {
  const currentUser = loadUserFromStorage()
  const isAuthenticated = !!currentUser

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresGuest && isAuthenticated) {
    next('/select-workspace')
  } else if (to.meta.requiresAdmin && currentUser?.role !== 'admin') {
    next('/select-workspace')
  } else if (to.meta.requiresWorkspace && !loadWorkspaceFromStorage()) {
    next('/select-workspace')
  } else {
    next()
  }
})

export default router
