import { createRouter, createWebHistory } from 'vue-router'
import { loadUserFromStorage, loadWorkspaceFromStorage } from '@/utils/storage'
import Login from '@/views/Login.vue'
import ForgotPassword from '@/views/ForgotPassword.vue'
import ResetPassword from '@/views/ResetPassword.vue'
import Register from '@/views/Register.vue'
import Dashboard from '@/views/Dashboard.vue'
import Setup from '@/views/Setup.vue'
import Manage from '@/views/Manage.vue'
import Users from '@/views/Users.vue'
import Invites from '@/views/Invites.vue'
import SelectWorkspace from '@/views/SelectWorkspace.vue'
import Workspaces from '@/views/Workspaces.vue'
import Workspace from '@/views/Workspace.vue'
import TaskTemplates from '@/views/TaskTemplates.vue'
import Inventory from '@/views/Inventory.vue'
import Environment from '@/views/Environment.vue'
import Repositories from '@/views/Repositories.vue'
import Backends from '@/views/Backends.vue'
import Runners from '@/views/Runners.vue'
import Integrations from '@/views/Integrations.vue'
import Profile from '@/views/Profile.vue'
import Settings from '@/views/Settings.vue'
import NotFound from '@/views/NotFound.vue'
import ServerError from '@/views/ServerError.vue'

const routes = [
  { path: '/', redirect: '/setup' },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true, title: 'Login', description: 'Login to platform' }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: ForgotPassword,
    meta: { requiresGuest: true, title: 'Forgot password', description: 'Request password reset' }
  },
  {
    path: '/reset-password/:token',
    name: 'ResetPassword',
    component: ResetPassword,
    meta: { requiresGuest: true, title: 'Reset password', description: 'Set new password' }
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { requiresGuest: true, title: 'Register', description: 'Create an account' }
  },
  {
    path: '/setup',
    name: 'Setup',
    component: Setup,
    meta: { requiresGuest: true, title: 'Setup', description: 'Setup platform' }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Dashboard', description: 'Platform dashboard' }
  },
  {
    path: '/users',
    name: 'Users',
    component: Users,
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Users', description: 'Manage users' }
  },
  {
    path: '/invites',
    name: 'Invites',
    component: Invites,
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Invites', description: 'Manage user invites' }
  },
  {
    path: '/manage',
    name: 'Manage',
    component: Manage,
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Manage', description: 'Manage platform' }
  },
  {
    path: '/select-workspace',
    name: 'SwitchWorkspaces',
    component: SelectWorkspace,
    meta: { requiresAuth: true, title: 'Switch Workspaces', description: 'Choose or create a workspace' }
  },
  {
    path: '/workspaces',
    name: 'Workspaces',
    component: Workspaces,
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Workspaces', description: 'Manage all platform workspaces' }
  },
  {
    path: '/workspace',
    name: 'Workspace',
    component: Workspace,
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Workspace settings', description: 'Workspace name, members, and delete' }
  },
  {
    path: '/templates',
    name: 'TaskTemplates',
    component: TaskTemplates,
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Task Templates', description: 'Manage task templates' }
  },
  {
    path: '/inventory',
    name: 'Inventory',
    component: Inventory,
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Inventory', description: 'Manage inventory' }
  },
  {
    path: '/environment',
    name: 'Environment',
    component: Environment,
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Environment', description: 'Environment configuration' }
  },
  {
    path: '/repositories',
    name: 'Repositories',
    component: Repositories,
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'Repositories', description: 'Manage repositories' }
  },
  {
    path: '/backends',
    name: 'Backends',
    component: Backends,
    meta: { requiresAuth: true, requiresWorkspace: true, title: 'TF Backends', description: 'Terraform backends' }
  },
  {
    path: '/runners',
    name: 'Runners',
    component: Runners,
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Runners', description: 'Manage runners' }
  },
  {
    path: '/integrations',
    name: 'Integrations',
    component: Integrations,
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Integrations', description: 'Manage integrations' }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: { requiresAuth: true, title: 'Profile', description: 'Your account profile' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: { requiresAuth: true, requiresAdmin: true, requiresWorkspace: true, title: 'Settings', description: 'Platform settings' }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: NotFound,
    meta: { title: 'Page not found' }
  },
  {
    path: '/500',
    name: 'ServerError',
    component: ServerError,
    meta: { title: 'Server error' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFoundCatchAll',
    component: NotFound,
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

