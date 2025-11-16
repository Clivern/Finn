<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-md w-full">
      <div class="text-center mb-10">
        <div class="flex justify-center mb-8">
          <router-link to="/">
            <img src="/logo.png" alt="Lapin Logo" class="h-24 w-auto">
          </router-link>
        </div>
      </div>

      <div class="bg-white rounded-lg border border-theme-border p-8 shadow-sm">
        <h2 class="text-2xl font-semibold text-theme-text mb-2 text-center">Reset password</h2>
        <p class="text-sm text-theme-textLight text-center mb-6">
          Enter your new password below.
        </p>
        <form class="space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label for="password" class="block text-sm font-medium text-theme-text mb-2">New password</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              required
              minlength="8"
              class="input-field"
              placeholder="Enter new password"
              :disabled="loading"
            >
          </div>
          <div>
            <label for="confirmPassword" class="block text-sm font-medium text-theme-text mb-2">Confirm password</label>
            <input
              id="confirmPassword"
              v-model="form.confirmPassword"
              type="password"
              required
              minlength="8"
              class="input-field"
              placeholder="Confirm new password"
              :disabled="loading"
            >
          </div>

          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3">
            <p class="text-sm text-red-800">{{ error }}</p>
          </div>
          <div v-if="success" class="rounded-md border border-green-200 bg-green-50 p-3">
            <p class="text-sm text-green-800">{{ success }}</p>
          </div>

          <div>
            <button
              type="submit"
              class="w-full btn-primary py-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
            >
              <span v-if="!loading">Reset password</span>
              <span v-else class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Resetting...
              </span>
            </button>
          </div>
        </form>

        <p class="text-center mt-6">
          <router-link to="/login" class="text-sm text-primary-800 hover:text-primary-900 font-medium">
            Back to login
          </router-link>
        </p>
      </div>

      <p class="text-center text-xs text-theme-textLight mt-8">
        Copyright © 2025 Lapin. All rights reserved.
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authAPI } from '@/api'

const route = useRoute()
const router = useRouter()
const token = computed(() => route.params.token || '')

const form = reactive({
  password: '',
  confirmPassword: ''
})
const loading = ref(false)
const error = ref(null)
const success = ref(null)

const handleSubmit = async () => {
  if (form.password !== form.confirmPassword) {
    error.value = 'Passwords do not match.'
    return
  }
  if (form.password.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return
  }
  if (!token.value) {
    error.value = 'Invalid or missing reset link.'
    return
  }

  loading.value = true
  error.value = null
  success.value = null
  try {
    await authAPI.resetPassword({ token: token.value, password: form.password })
    success.value = 'Your password has been reset. Redirecting to login...'
    setTimeout(() => router.push('/login'), 2000)
  } catch (err) {
    error.value = err.response?.data?.errorMessage || 'Failed to reset password. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>
