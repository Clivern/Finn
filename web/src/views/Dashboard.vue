<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

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
        <p v-if="user?.email" class="mt-2 text-sm text-theme-text">
          Welcome back, <span class="font-medium">{{ user.email }}</span>
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

      <!-- Pending invites (admin): invite links on the dashboard -->
      <section v-if="user?.role === 'admin'" aria-label="Pending invites" class="mb-8">
        <div class="section">
          <div class="flex items-center justify-between mb-4">
            <h2 class="section-title">Pending invites</h2>
            <router-link to="/invites" class="text-sm font-medium text-primary-600 hover:text-primary-800">View all</router-link>
          </div>
          <p class="text-sm text-theme-textLight mb-4">
            Invites send an email to the recipient when SMTP is configured. You can also copy and share the invite link below.
          </p>
          <div v-if="pendingInvitesLoading" class="text-sm text-theme-textLight">Loading…</div>
          <div v-else-if="pendingInvites.length === 0" class="text-sm text-theme-textLight">
            No pending invites. <router-link to="/invites" class="text-primary-600 hover:text-primary-800">Create an invite</router-link>.
          </div>
          <ul v-else class="space-y-3">
            <li
              v-for="invite in pendingInvites"
              :key="invite.id"
              class="flex flex-wrap items-center gap-2 py-3 border-b border-theme-border last:border-0"
            >
              <span class="text-sm font-medium text-theme-text">{{ invite.email }}</span>
              <span class="text-xs text-theme-textLight">({{ invite.role }})</span>
              <button
                type="button"
                @click="copyInviteLink(invite)"
                class="text-sm text-primary-600 hover:text-primary-800"
              >
                {{ copyFeedback === invite.id ? 'Copied!' : 'Copy invite link' }}
              </button>
            </li>
          </ul>
        </div>
      </section>

      <!-- Quick Actions -->
      <section aria-label="Quick actions">
        <div class="section">
          <h2 class="section-title">Quick Actions</h2>
          <div class="flex flex-wrap gap-3">
            <router-link to="/users" class="btn-primary">Manage Users</router-link>
            <router-link to="/invites" class="btn-secondary">Invites</router-link>
            <router-link to="/manage" class="btn-secondary">Manage</router-link>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { healthAPI, invitesAPI } from '@/api'
import AppNav from '@/components/AppNav.vue'
import { user } from '@/lib/auth'
import { loadWorkspaceFromStorage, loadUserFromStorage } from '@/utils/storage'

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

const pendingInvitesLoading = ref(false)
const allInvites = ref([])

const pendingInvites = computed(() => {
  const now = new Date()
  return allInvites.value
    .filter(inv => inv.status === 'pending' && inv.token && new Date(inv.expiresAt) > now)
    .slice(0, 10)
})

const inviteLink = (token) => {
  const base = typeof window !== 'undefined' ? window.location.origin : ''
  return `${base}/register?token=${encodeURIComponent(token)}`
}

const copyInviteLink = async (invite) => {
  const link = inviteLink(invite.token)
  try {
    await navigator.clipboard.writeText(link)
    copyFeedback.value = invite.id
    setTimeout(() => { copyFeedback.value = null }, 2000)
  } catch (_) {}
}

const copyFeedback = ref(null)

const loadPendingInvites = async () => {
  if (user?.role !== 'admin') return
  pendingInvitesLoading.value = true
  try {
    const response = await invitesAPI.list()
    allInvites.value = response.data?.invites ?? []
  } catch (_) {
    allInvites.value = []
  } finally {
    pendingInvitesLoading.value = false
  }
}

onMounted(async () => {
  console.log('workspace', loadWorkspaceFromStorage())
  console.log('user', loadUserFromStorage())
  try {
    await healthAPI.check()
  } catch (error) {
    console.error('Failed to check API health:', error)
  }
  loadPendingInvites()
})
</script>
