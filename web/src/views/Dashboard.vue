<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <!-- Main Content -->
    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <!-- Page Header -->
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('dashboard.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('dashboard.subtitle') }}
        </p>
        <p v-if="user?.email" class="mt-2 text-sm text-theme-text">
          {{ $t('dashboard.welcomeBack') }} <span class="font-medium">{{ user.email }}</span>
        </p>
      </header>

      <!-- Stats Grid -->
      <section class="mb-8" aria-label="Key metrics">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <div class="stat-card group">
            <div class="flex items-start justify-between">
              <div>
                <p class="stat-label">{{ $t('dashboard.status') }}</p>
                <p class="stat-value mt-1 flex items-center gap-2">
                  <span class="inline-block h-2 w-2 rounded-full bg-emerald-500 ring-2 ring-emerald-500/30" aria-hidden="true"></span>
                  {{ stats.status }}
                </p>
              </div>
            </div>
          </div>
          <div class="stat-card group">
            <div>
              <p class="stat-label">{{ $t('dashboard.totalRequests') }}</p>
              <p class="stat-value mt-1">{{ stats.totalRequests.toLocaleString() }}</p>
            </div>
          </div>
          <div class="stat-card group">
            <div>
              <p class="stat-label">{{ $t('dashboard.avgResponse') }}</p>
              <p class="stat-value mt-1">{{ stats.avgResponse }}ms</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Content Grid -->
      <section class="grid grid-cols-1 gap-6 lg:grid-cols-2 mb-8" aria-label="Overview and system info">
        <div class="section">
          <h2 class="section-title">{{ $t('dashboard.overview') }}</h2>
          <p class="text-sm text-theme-textLight leading-relaxed">
            {{ $t('dashboard.overviewDescription') }}
          </p>
          <ul class="mt-4 space-y-2 text-sm">
            <li class="flex items-center gap-2 text-theme-text">
              <span class="text-theme-textLight">•</span>
              {{ $t('dashboard.userManagement') }}
            </li>
            <li class="flex items-center gap-2 text-theme-text">
              <span class="text-theme-textLight">•</span>
              {{ $t('dashboardPage.sessionAuth') }}
            </li>
            <li class="flex items-center gap-2 text-theme-text">
              <span class="text-theme-textLight">•</span>
              {{ $t('dashboardPage.extendFromHere') }}
            </li>
          </ul>
        </div>

        <div class="section">
          <h2 class="section-title">{{ $t('dashboardPage.systemInfo') }}</h2>
          <dl class="space-y-0">
            <div class="flex justify-between py-3 border-b border-theme-border">
              <dt class="text-sm text-theme-textLight">{{ $t('dashboardPage.version') }}</dt>
              <dd class="text-sm font-medium text-theme-text">v0.4.0</dd>
            </div>
            <div class="flex justify-between py-3 border-b border-theme-border">
              <dt class="text-sm text-theme-textLight">{{ $t('dashboardPage.uptime') }}</dt>
              <dd class="text-sm font-medium text-theme-text">{{ systemInfo.uptime }}</dd>
            </div>
            <div class="flex justify-between py-3 border-b border-theme-border">
              <dt class="text-sm text-theme-textLight">{{ $t('dashboardPage.memoryUsage') }}</dt>
              <dd class="text-sm font-medium text-theme-text">{{ systemInfo.memory }}</dd>
            </div>
            <div class="flex justify-between py-3">
              <dt class="text-sm text-theme-textLight">{{ $t('dashboardPage.cpuUsage') }}</dt>
              <dd class="text-sm font-medium text-theme-text">{{ systemInfo.cpu }}</dd>
            </div>
          </dl>
        </div>
      </section>

      <!-- Pending invites (admin): invite links on the dashboard -->
      <section v-if="user?.role === 'admin'" aria-label="Pending invites" class="mb-8">
        <div class="section">
          <div class="flex items-center justify-between mb-4">
            <h2 class="section-title">{{ $t('dashboardPage.pendingInvites') }}</h2>
            <router-link to="/invites" class="text-sm font-medium text-primary-600 hover:text-primary-800">{{ $t('dashboardPage.viewAll') }}</router-link>
          </div>
          <p class="text-sm text-theme-textLight mb-4">
            {{ $t('dashboardPage.invitesDescription') }}
          </p>
          <div v-if="pendingInvitesLoading" class="text-sm text-theme-textLight">{{ $t('dashboardPage.loading') }}</div>
          <div v-else-if="pendingInvites.length === 0" class="text-sm text-theme-textLight">
            {{ $t('dashboardPage.noPendingInvites') }} <router-link to="/invites" class="text-primary-600 hover:text-primary-800">{{ $t('dashboardPage.createInviteLink') }}</router-link>.
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
                {{ copyFeedback === invite.id ? $t('dashboardPage.copied') : $t('dashboardPage.copyInviteLink') }}
              </button>
            </li>
          </ul>
        </div>
      </section>

      <!-- Quick Actions -->
      <section aria-label="Quick actions">
        <div class="section">
          <h2 class="section-title">{{ $t('dashboardPage.quickActions') }}</h2>
          <div class="flex flex-wrap gap-3">
            <router-link to="/users" class="btn-primary">{{ $t('dashboardPage.manageUsers') }}</router-link>
            <router-link to="/invites" class="btn-secondary">{{ $t('nav.invites') }}</router-link>
            <router-link to="/manage" class="btn-secondary">{{ $t('nav.manage') }}</router-link>
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
