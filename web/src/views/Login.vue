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
        <h2 class="text-2xl font-semibold text-theme-text mb-6 text-center">Welcome back</h2>
        <form class="space-y-5" @submit.prevent="handleLogin">
          <div>
            <label for="email" class="block text-sm font-medium text-theme-text mb-2">
              Email
            </label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              required
              class="input-field"
              placeholder="Enter your email"
              :disabled="loading"
            >
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-theme-text mb-2">
              Password
            </label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              required
              class="input-field"
              placeholder="Enter your password"
              :disabled="loading"
            >
          </div>

          <div class="flex items-center justify-between">
            <div class="flex items-center">
              <input
                id="remember-me"
                v-model="form.rememberMe"
                type="checkbox"
                class="h-4 w-4 rounded text-theme-text focus:ring-theme-text border-theme-border"
              >
              <label for="remember-me" class="ml-2 block text-sm text-theme-textLight">
                Remember me
              </label>
            </div>
            <router-link to="/forgot-password" class="text-sm text-primary-800 hover:text-primary-900 font-medium">
              Forgot password?
            </router-link>
          </div>

          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3">
            <p class="text-sm text-red-800">
              {{ error }}
            </p>
          </div>

          <div class="space-y-3">
            <button
              type="submit"
              class="w-full btn-primary py-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
            >
              <span v-if="!loading">Sign in</span>
              <span v-else class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Signing in...
              </span>
            </button>
            <p class="text-center text-sm text-theme-textLight">
              Don't have an account?
              <router-link to="/register" class="font-medium text-primary-800 hover:text-primary-900">Register</router-link>
            </p>
          </div>
        </form>
      </div>

      <p class="text-center text-xs text-theme-textLight mt-8">
        Copyright © 2025 Lapin. All rights reserved.
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '@/api'
import { saveUser } from '@/lib/auth'

const router = useRouter()
const form = reactive({
  email: '',
  password: '',
  rememberMe: false
})

const loading = ref(false)
const error = ref(null)

const handleLogin = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await authAPI.login({
      email: form.email,
      password: form.password,
      rememberMe: form.rememberMe
    })

    saveUser(response.data?.user)
    router.push('/select-workspace')
  } catch (err) {
    error.value = err.response?.data?.errorMessage || 'Login failed. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

