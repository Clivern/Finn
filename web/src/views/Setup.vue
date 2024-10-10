<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4 py-12">
    <div class="max-w-md w-full">
      <div class="text-center mb-10">
        <div class="flex justify-center mb-6">
          <img src="/logo.png" :alt="$t('common.logoAlt')" class="h-24 w-auto">
        </div>
        <h1 class="text-2xl font-semibold text-theme-text mb-2">{{ $t('setup.title') }}</h1>
        <p class="text-sm text-theme-textLight">{{ $t('setup.subtitle') }}</p>
      </div>

      <div class="bg-white rounded-lg border border-theme-border p-8 shadow-sm">
        <form class="space-y-5" @submit.prevent="handleSetup">
          <div>
            <label for="platform-name" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('setup.platformName') }}
            </label>
            <input
              id="platform-name"
              v-model="form.platformName"
              type="text"
              required
              minlength="2"
              maxlength="50"
              class="input-field"
              :placeholder="$t('setup.platformNamePlaceholder')"
              :disabled="loading"
            >
            <p class="text-xs text-theme-textLight mt-1.5">{{ $t('setup.platformNameHint') }}</p>
          </div>

          <div>
            <label for="platform-url" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('setup.platformURL') }}
            </label>
            <input
              id="platform-url"
              v-model="form.platformURL"
              type="url"
              required
              class="input-field"
              :placeholder="$t('setup.platformURLPlaceholder')"
              :disabled="loading"
            >
            <p class="text-xs text-theme-textLight mt-1.5">{{ $t('setup.platformURLHint') }}</p>
          </div>

          <div>
            <label for="platform-email" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('setup.platformEmail') }}
            </label>
            <input
              id="platform-email"
              v-model="form.platformEmail"
              type="email"
              required
              class="input-field"
              :placeholder="$t('setup.platformEmailPlaceholder')"
              :disabled="loading"
            >
            <p class="text-xs text-theme-textLight mt-1.5">{{ $t('setup.platformEmailHint') }}</p>
          </div>

          <div class="relative py-3">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-theme-border"></div>
            </div>
            <div class="relative flex justify-center text-xs">
              <span class="px-3 bg-white text-theme-textLight font-medium">{{ $t('setup.adminAccount') }}</span>
            </div>
          </div>

          <div>
            <label for="admin-email" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('setup.adminEmail') }}
            </label>
            <input
              id="admin-email"
              v-model="form.adminEmail"
              type="email"
              required
              class="input-field"
              :placeholder="$t('setup.adminEmailPlaceholder')"
              :disabled="loading"
            >
            <p class="text-xs text-theme-textLight mt-1.5">{{ $t('setup.adminEmailHint') }}</p>
          </div>

          <div>
            <label for="admin-password" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('setup.adminPassword') }}
            </label>
            <input
              id="admin-password"
              v-model="form.adminPassword"
              type="password"
              required
              minlength="8"
              class="input-field"
              :placeholder="$t('setup.adminPasswordPlaceholder')"
              :disabled="loading"
            >
            <p class="text-xs text-theme-textLight mt-1.5">{{ $t('setup.adminPasswordHint') }}</p>
          </div>

          <div>
            <label for="confirm-password" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('setup.confirmPassword') }}
            </label>
            <input
              id="confirm-password"
              v-model="form.confirmPassword"
              type="password"
              required
              minlength="8"
              class="input-field"
              :placeholder="$t('setup.confirmPasswordPlaceholder')"
              :disabled="loading"
            >
          </div>

          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3">
            <p class="text-sm text-red-800">
              {{ error }}
            </p>
          </div>

          <div v-if="success" class="rounded-md border border-green-200 bg-green-50 p-3">
            <p class="text-sm text-green-800">
              {{ success }}
            </p>
          </div>

          <div>
            <button
              type="submit"
              class="w-full btn-primary py-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
            >
              <span v-if="!loading">{{ $t('setup.submit') }}</span>
              <span v-else class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                {{ $t('setup.submitting') }}
              </span>
            </button>
          </div>
        </form>
      </div>

      <p class="text-center text-xs text-theme-textLight mt-8">
        {{ $t('common.copyright') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { setupAPI } from '@/api'

const { t } = useI18n()
const router = useRouter()
const form = reactive({
  platformURL: '',
  platformEmail: '',
  platformName: '',
  adminEmail: '',
  adminPassword: '',
  confirmPassword: ''
})

const loading = ref(false)
const error = ref(null)
const success = ref(null)

onMounted(async () => {
  try {
    const response = await setupAPI.checkInstalled()
    if (response.data.installed) {
      router.push('/login')
    }
  } catch (err) {
    console.error('Failed to check setup status:', err)
  }
})

const validateForm = () => {
  if (form.adminPassword !== form.confirmPassword) {
    error.value = t('setup.errorPasswordsMatch')
    return false
  }

  if (form.adminPassword.length < 8) {
    error.value = t('setup.errorPasswordLength')
    return false
  }

  try {
    new URL(form.platformURL)
  } catch {
    error.value = t('setup.errorInvalidURL')
    return false
  }

  return true
}

const handleSetup = async () => {
  loading.value = true
  error.value = null
  success.value = null

  if (!validateForm()) {
    loading.value = false
    return
  }

  try {
    await setupAPI.install({
      platformURL: form.platformURL,
      platformEmail: form.platformEmail,
      platformName: form.platformName,
      adminEmail: form.adminEmail,
      adminPassword: form.adminPassword
    })

    success.value = t('setup.success')

    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (err) {
    error.value = err.response?.data?.errorMessage || t('setup.errorDefault')
  } finally {
    loading.value = false
  }
}
</script>

