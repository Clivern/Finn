import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Login from '@/views/Login.vue'
import Dashboard from '@/views/Dashboard.vue'
import Setup from '@/views/Setup.vue'
import Settings from '@/views/Settings.vue'
import Users from '@/views/Users.vue'
import SelectWorkspace from '@/views/SelectWorkspace.vue'
import TaskTemplates from '@/views/TaskTemplates.vue'
import Inventory from '@/views/Inventory.vue'
import Environment from '@/views/Environment.vue'
import Repositories from '@/views/Repositories.vue'
import TFBackends from '@/views/TFBackends.vue'
import Runners from '@/views/Runners.vue'
import Integrations from '@/views/Integrations.vue'
import NotFound from '@/views/NotFound.vue'
import ServerError from '@/views/ServerError.vue'

const routes = [
  { path: '/', redirect: '/login' },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true }
  },
  {
    path: '/setup',
    name: 'Setup',
    component: Setup,
    meta: { requiresGuest: true }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/users',
    name: 'Users',
    component: Users,
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: { requiresAuth: true }
  },
  {
    path: '/workspaces',
    name: 'SelectWorkspace',
    component: SelectWorkspace,
    meta: { requiresAuth: true }
  },
  {
    path: '/task-templates',
    name: 'TaskTemplates',
    component: TaskTemplates,
    meta: { requiresAuth: true, title: 'Task Templates', description: 'Manage task templates' }
  },
  {
    path: '/inventory',
    name: 'Inventory',
    component: Inventory,
    meta: { requiresAuth: true, title: 'Inventory', description: 'Manage inventory' }
  },
  {
    path: '/environment',
    name: 'Environment',
    component: Environment,
    meta: { requiresAuth: true, title: 'Environment', description: 'Environment configuration' }
  },
  {
    path: '/repositories',
    name: 'Repositories',
    component: Repositories,
    meta: { requiresAuth: true, title: 'Repositories', description: 'Manage repositories' }
  },
  {
    path: '/tf-backends',
    name: 'TFBackends',
    component: TFBackends,
    meta: { requiresAuth: true, title: 'TF Backends', description: 'Terraform backends' }
  },
  {
    path: '/runners',
    name: 'Runners',
    component: Runners,
    meta: { requiresAuth: true, title: 'Runners', description: 'Manage runners' }
  },
  {
    path: '/integrations',
    name: 'Integrations',
    component: Integrations,
    meta: { requiresAuth: true, title: 'Integrations', description: 'Manage integrations' }
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

// Navigation guard for authentication and setup
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Wait for auth initialization to complete
  while (authStore.isInitializing) {
    await new Promise(resolve => setTimeout(resolve, 50))
  }

  const isAuthenticated = authStore.isAuthenticated

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresGuest && isAuthenticated) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router

