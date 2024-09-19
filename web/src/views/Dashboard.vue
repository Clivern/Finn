<template>
  <div class="min-h-screen bg-theme-bg">
    <!-- Navigation -->
    <nav class="bg-white border-b border-theme-border">
      <div class="w-full px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-14">
          <div class="flex items-center">
            <div class="flex-shrink-0 flex items-center">
              <img src="/logo.png" alt="Met Logo" class="h-8 w-auto">
            </div>
            <div class="hidden md:ml-8 md:flex md:space-x-1">
              <router-link
                to="/dashboard"
                class="bg-theme-hover text-theme-text inline-flex items-center px-3 py-2 rounded-md text-sm font-medium"
              >
                Dashboard
              </router-link>
              <router-link
                to="/users"
                class="text-theme-textLight hover:bg-theme-hover hover:text-theme-text inline-flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Users
              </router-link>
              <router-link
                to="/settings"
                class="text-theme-textLight hover:bg-theme-hover hover:text-theme-text inline-flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Settings
              </router-link>
            </div>
          </div>
          <div class="flex items-center">
            <UserAvatarDropdown />
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <!-- Page Header -->
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Dashboard
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Overview of your platform status and quick actions
        </p>
        <p v-if="authStore.currentUser?.email" class="mt-2 text-sm text-theme-text">
          Welcome back, <span class="font-medium">{{ authStore.currentUser.email }}</span>
        </p>
      </header>

      <!-- Stats Grid -->
      <section class="mb-8" aria-label="Key metrics">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <div class="stat-card group">
            <div class="flex items-start justify-between">
              <div>
                <p class="stat-label">Status</p>
                <p class="stat-value mt-1 flex items-center gap-2">
                  <span class="inline-block h-2 w-2 rounded-full bg-emerald-500 ring-2 ring-emerald-500/30" aria-hidden="true"></span>
                  {{ stats.status }}
                </p>
              </div>
            </div>
          </div>
          <div class="stat-card group">
            <div>
              <p class="stat-label">Total Requests</p>
              <p class="stat-value mt-1">{{ stats.totalRequests.toLocaleString() }}</p>
            </div>
          </div>
          <div class="stat-card group">
            <div>
              <p class="stat-label">Avg Response</p>
              <p class="stat-value mt-1">{{ stats.avgResponse }}ms</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Content Grid -->
      <section class="grid grid-cols-1 gap-6 lg:grid-cols-2 mb-8" aria-label="Overview and system info">
        <div class="section">
          <h2 class="section-title">Overview</h2>
          <p class="text-sm text-theme-textLight leading-relaxed">
            Your platform is running with users and sessions. Use the links below to manage users, adjust settings, or explore the API.
          </p>
          <ul class="mt-4 space-y-2 text-sm">
            <li class="flex items-center gap-2 text-theme-text">
              <span class="text-theme-textLight">•</span>
              User management and roles
            </li>
            <li class="flex items-center gap-2 text-theme-text">
              <span class="text-theme-textLight">•</span>
              Session-based authentication
            </li>
            <li class="flex items-center gap-2 text-theme-text">
              <span class="text-theme-textLight">•</span>
              Extend from here with your own features
            </li>
          </ul>
        </div>

        <div class="section">
          <h2 class="section-title">System Information</h2>
          <dl class="space-y-0">
            <div class="flex justify-between py-3 border-b border-theme-border">
              <dt class="text-sm text-theme-textLight">Version</dt>
              <dd class="text-sm font-medium text-theme-text">v0.4.0</dd>
            </div>
            <div class="flex justify-between py-3 border-b border-theme-border">
              <dt class="text-sm text-theme-textLight">Uptime</dt>
              <dd class="text-sm font-medium text-theme-text">{{ systemInfo.uptime }}</dd>
            </div>
            <div class="flex justify-between py-3 border-b border-theme-border">
              <dt class="text-sm text-theme-textLight">Memory Usage</dt>
              <dd class="text-sm font-medium text-theme-text">{{ systemInfo.memory }}</dd>
            </div>
            <div class="flex justify-between py-3">
              <dt class="text-sm text-theme-textLight">CPU Usage</dt>
              <dd class="text-sm font-medium text-theme-text">{{ systemInfo.cpu }}</dd>
            </div>
          </dl>
        </div>
      </section>

      <!-- Quick Actions -->
      <section aria-label="Quick actions">
        <div class="section">
          <h2 class="section-title">Quick Actions</h2>
          <div class="flex flex-wrap gap-3">
            <router-link to="/users" class="btn-primary">Manage Users</router-link>
            <router-link to="/settings" class="btn-secondary">Settings</router-link>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { healthAPI } from '@/api'
import UserAvatarDropdown from '@/components/UserAvatarDropdown.vue'

const authStore = useAuthStore()

const stats = ref({
  status: 'Online',
  totalRequests: 42853,
  avgResponse: 145
})

const systemInfo = ref({
  uptime: '7 days, 14 hours',
  memory: '245 MB / 2 GB',
  cpu: '12%'
})

onMounted(async () => {
  try {
    await healthAPI.check()
  } catch (error) {
    console.error('Failed to check API health:', error)
  }
})
</script>
