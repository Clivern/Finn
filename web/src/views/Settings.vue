<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <!-- Settings content -->
    <div class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <main>
        <div class="mb-8">
          <h1 class="text-2xl font-semibold text-theme-text">Settings</h1>
          <p class="text-sm text-theme-textLight mt-1">System-wide settings and platform configuration</p>
        </div>

        <!-- Settings dashboard: quick access to admin areas -->
        <section class="mb-10" aria-label="System settings">
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <router-link
              to="/workspaces"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Workspaces</p>
                <p class="text-xs text-theme-textLight truncate">Manage workspaces</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/users"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Users</p>
                <p class="text-xs text-theme-textLight truncate">User accounts and roles</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/integrations"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Integrations</p>
                <p class="text-xs text-theme-textLight truncate">External services and APIs</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/runners"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Runners</p>
                <p class="text-xs text-theme-textLight truncate">Execution runners</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
          </div>
        </section>

        <!-- Success / Error -->
        <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
          {{ successMessage }}
        </div>
        <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {{ errorMessage }}
        </div>

        <form @submit.prevent="handleSave" class="space-y-8">
          <!-- General -->
          <section class="rounded-xl bg-white border border-theme-border shadow-sm overflow-hidden">
              <div class="px-6 py-4 border-b border-theme-border bg-primary-50/50">
                <h2 class="text-base font-semibold text-theme-text">General</h2>
                <p class="text-xs text-theme-textLight mt-0.5">Platform identity and contact</p>
              </div>
              <div class="p-6 space-y-5">
                <div>
                  <label for="platformName" class="block text-sm font-medium text-theme-text mb-1.5">Platform Name</label>
                  <input
                    id="platformName"
                    v-model="form.platformName"
                    type="text"
                    required
                    class="input-field max-w-md"
                    placeholder="Met"
                    :disabled="loading"
                  >
                </div>
                <div>
                  <label for="platformEmail" class="block text-sm font-medium text-theme-text mb-1.5">Platform Email</label>
                  <input
                    id="platformEmail"
                    v-model="form.platformEmail"
                    type="email"
                    required
                    class="input-field max-w-md"
                    placeholder="admin@met.example.com"
                    :disabled="loading"
                  >
                </div>
                <div>
                  <label for="platformURL" class="block text-sm font-medium text-theme-text mb-1.5">Platform URL</label>
                  <input
                    id="platformURL"
                    v-model="form.platformURL"
                    type="url"
                    required
                    class="input-field max-w-md"
                    placeholder="https://met.example.com"
                    :disabled="loading"
                  >
                </div>
              </div>
            </section>

          <!-- Email (SMTP) -->
          <section class="rounded-xl bg-white border border-theme-border shadow-sm overflow-hidden">
              <div class="px-6 py-4 border-b border-theme-border bg-primary-50/50">
                <h2 class="text-base font-semibold text-theme-text">Email</h2>
                <p class="text-xs text-theme-textLight mt-0.5">SMTP server for outgoing mail</p>
              </div>
              <div class="p-6 space-y-5">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
                  <div>
                    <label for="smtpServer" class="block text-sm font-medium text-theme-text mb-1.5">SMTP Server</label>
                    <input
                      id="smtpServer"
                      v-model="form.smtpServer"
                      type="text"
                      class="input-field"
                      placeholder="smtp.example.com"
                      :disabled="loading"
                    >
                  </div>
                  <div>
                    <label for="smtpPort" class="block text-sm font-medium text-theme-text mb-1.5">Port</label>
                    <input
                      id="smtpPort"
                      v-model="form.smtpPort"
                      type="text"
                      class="input-field"
                      placeholder="587"
                      :disabled="loading"
                    >
                  </div>
                </div>
                <div>
                  <label for="smtpFromEmail" class="block text-sm font-medium text-theme-text mb-1.5">From Email</label>
                  <input
                    id="smtpFromEmail"
                    v-model="form.smtpFromEmail"
                    type="email"
                    class="input-field max-w-md"
                    placeholder="noreply@example.com"
                    :disabled="loading"
                  >
                </div>
                <div>
                  <label for="smtpUsername" class="block text-sm font-medium text-theme-text mb-1.5">Username</label>
                  <input
                    id="smtpUsername"
                    v-model="form.smtpUsername"
                    type="text"
                    class="input-field max-w-md"
                    placeholder="SMTP username"
                    :disabled="loading"
                  >
                </div>
                <div>
                  <label for="smtpPassword" class="block text-sm font-medium text-theme-text mb-1.5">Password</label>
                  <input
                    id="smtpPassword"
                    v-model="form.smtpPassword"
                    type="password"
                    class="input-field max-w-md"
                    placeholder="SMTP password"
                    :disabled="loading"
                  >
                </div>
                <label class="flex items-start gap-3 cursor-pointer">
                  <input
                    id="smtpUseTLS"
                    v-model="form.smtpUseTLS"
                    type="checkbox"
                    class="mt-1 h-4 w-4 rounded border-theme-border text-primary-800 focus:ring-primary-500"
                    :disabled="loading"
                  >
                  <span class="text-sm text-theme-text">
                    <span class="font-medium">Use TLS</span>
                    <span class="text-theme-textLight ml-1">— secure connection</span>
                  </span>
                </label>
              </div>
            </section>

          <!-- Maintenance -->
          <section class="rounded-xl bg-white border border-theme-border shadow-sm overflow-hidden">
              <div class="px-6 py-4 border-b border-theme-border bg-primary-50/50">
                <h2 class="text-base font-semibold text-theme-text">Maintenance</h2>
                <p class="text-xs text-theme-textLight mt-0.5">Control application availability</p>
              </div>
              <div class="p-6">
                <label class="flex items-start gap-3 cursor-pointer">
                  <input
                    id="maintenanceMode"
                    v-model="form.maintenanceMode"
                    type="checkbox"
                    class="mt-1 h-4 w-4 rounded border-theme-border text-primary-800 focus:ring-primary-500"
                    :disabled="loading"
                  >
                  <span class="text-sm text-theme-text">
                    <span class="font-medium">Maintenance mode</span>
                    <span class="text-theme-textLight ml-1">— disable access for non-admins</span>
                  </span>
                </label>
              </div>
            </section>

          <!-- Save -->
          <div class="flex items-center justify-end pt-2 pb-4">
            <button
              type="submit"
              class="btn-primary"
              :disabled="loading"
            >
              <span v-if="!loading">Save settings</span>
              <span v-else class="inline-flex items-center gap-2">
                <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24" aria-hidden="true">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Saving…
              </span>
            </button>
          </div>
        </form>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { settingsAPI } from '@/api'
import AppNav from '@/components/AppNav.vue'

const authStore = useAuthStore()

const loading = ref(false)
const successMessage = ref(null)
const errorMessage = ref(null)

const form = reactive({
  platformName: '',
  platformURL: '',
  platformEmail: '',
  maintenanceMode: false,
  smtpServer: '',
  smtpPort: '',
  smtpFromEmail: '',
  smtpUsername: '',
  smtpPassword: '',
  smtpUseTLS: true
})

const loadSettings = async () => {
  loading.value = true
  errorMessage.value = null
  try {
    const response = await settingsAPI.getSettings()
    const settings = response.data?.settings
    if (settings) {
      form.platformName = settings.platform_name || ''
      form.platformURL = settings.platform_url || ''
      form.platformEmail = settings.platform_email || ''
      form.maintenanceMode = settings.maintenance_mode === '1'
      form.smtpServer = settings.smtp_server || ''
      form.smtpPort = settings.smtp_port || ''
      form.smtpFromEmail = settings.smtp_from_email || ''
      form.smtpUsername = settings.smtp_username || ''
      form.smtpPassword = settings.smtp_password || ''
      form.smtpUseTLS = settings.smtp_use_tls === '1'
    }
  } catch (err) {
    console.error('Failed to load settings:', err)
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load settings'
  } finally {
    loading.value = false
  }
}

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

watch([successMessage, errorMessage], () => {
  if (successMessage.value || errorMessage.value) {
    nextTick(scrollToTop)
  }
})

const handleSave = async () => {
  loading.value = true
  successMessage.value = null
  errorMessage.value = null
  try {
    const payload = {
      platformName: form.platformName,
      platformUrl: form.platformURL,
      platformEmail: form.platformEmail,
      maintenanceMode: form.maintenanceMode,
      smtpUseTLS: form.smtpUseTLS
    }
    if (form.smtpServer) payload.smtpServer = form.smtpServer
    if (form.smtpPort) payload.smtpPort = parseInt(form.smtpPort, 10) || 587
    if (form.smtpFromEmail) payload.smtpFromEmail = form.smtpFromEmail
    if (form.smtpUsername) payload.smtpUsername = form.smtpUsername
    if (form.smtpPassword) payload.smtpPassword = form.smtpPassword

    const response = await settingsAPI.updateSettings(payload)
    if (response.data?.successMessage) {
      successMessage.value = response.data.successMessage
      setTimeout(() => { successMessage.value = null }, 3000)
    }
  } catch (err) {
    console.error('Failed to save settings:', err)
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to save settings. Please try again.'
  } finally {
    loading.value = false
  }
}

onMounted(loadSettings)
</script>
