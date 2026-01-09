<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-md w-full">
      <div class="text-center mb-10">
        <div class="flex justify-center mb-8">
          <img src="/logo.png" alt="Tivo Logo" class="h-24 w-auto">
        </div>
      </div>

      <div class="bg-white rounded-lg border border-theme-border p-8 shadow-sm">
        <!-- No token: invite-only message -->
        <template v-if="!token">
          <h2 class="text-2xl font-semibold text-theme-text mb-2 text-center">Create an account</h2>
          <div class="rounded-md border border-primary-200 bg-primary-50 p-4 mb-6">
            <p class="text-sm text-primary-800">
              <strong>Registration is by invite only.</strong> New accounts are created by invite only. If you received an invite, use the link in your email to set up your account.
            </p>
          </div>
          <p class="text-sm text-theme-textLight text-center mb-6">
            Already have an account?
          </p>
          <div>
            <router-link
              to="/login"
              class="w-full btn-primary py-2.5 inline-block text-center"
            >
              Sign in
            </router-link>
          </div>
        </template>

        <!-- Token present: loading invite -->
        <template v-else-if="inviteLoading">
          <h2 class="text-2xl font-semibold text-theme-text mb-6 text-center">Create an account</h2>
          <div class="flex justify-center py-8">
            <svg class="animate-spin h-8 w-8 text-theme-text" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
          </div>
        </template>

        <!-- Token invalid or expired -->
        <template v-else-if="inviteError">
          <h2 class="text-2xl font-semibold text-theme-text mb-2 text-center">Create an account</h2>
          <div class="rounded-md border border-red-200 bg-red-50 p-4 mb-6">
            <p class="text-sm text-red-800">
              {{ inviteError }}
            </p>
          </div>
          <p class="text-center text-sm text-theme-textLight mb-6">
            Please request a new invite or contact your administrator.
          </p>
          <router-link
            to="/login"
            class="w-full btn-primary py-2.5 inline-block text-center"
          >
            Sign in
          </router-link>
        </template>

        <!-- Token valid: registration form -->
        <template v-else>
          <h2 class="text-2xl font-semibold text-theme-text mb-2 text-center">Create an account</h2>
          <p class="text-sm text-theme-textLight text-center mb-6">
            Complete the form below to set up your account.
          </p>
          <form class="space-y-5" @submit.prevent="handleRegister">
            <div>
              <label for="email" class="block text-sm font-medium text-theme-text mb-2">Email</label>
              <input
                id="email"
                :value="invite?.email"
                type="email"
                readonly
                class="input-field bg-theme-bg cursor-not-allowed"
              >
            </div>
            <div>
              <label for="name" class="block text-sm font-medium text-theme-text mb-2">Name</label>
              <input
                id="name"
                v-model="form.name"
                type="text"
                required
                class="input-field"
                placeholder="Your full name"
                :disabled="loading"
              >
            </div>
            <div>
              <label for="password" class="block text-sm font-medium text-theme-text mb-2">Password</label>
              <input
                id="password"
                v-model="form.password"
                type="password"
                required
                minlength="8"
                class="input-field"
                placeholder="At least 8 characters, with uppercase, lowercase, number and special character"
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
                placeholder="Confirm your password"
                :disabled="loading"
              >
            </div>

            <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3">
              <p class="text-sm text-red-800">{{ error }}</p>
            </div>

            <div>
              <button
                type="submit"
                class="w-full btn-primary py-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="loading"
              >
                <span v-if="!loading">Create account</span>
                <span v-else class="flex items-center justify-center">
                  <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Creating account...
                </span>
              </button>
            </div>
            <p class="text-center text-sm text-theme-textLight">
              Already have an account?
              <router-link to="/login" class="font-medium text-primary-800 hover:text-primary-900">Sign in</router-link>
            </p>
          </form>
        </template>
      </div>

      <p class="text-center text-xs text-theme-textLight mt-8">
        Copyright © 2025 Tivo. All rights reserved.
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authAPI } from '@/api'

const route = useRoute()
const router = useRouter()
const token = computed(() => route.query.token || '')

const invite = ref(null)
const inviteLoading = ref(false)
const inviteError = ref(null)

const form = reactive({
  name: '',
  password: '',
  confirmPassword: ''
})
const loading = ref(false)
const error = ref(null)

async function fetchInvite() {
  if (!token.value) return
  inviteLoading.value = true
  inviteError.value = null
  try {
    const res = await authAPI.getInviteByToken(token.value)
    invite.value = res.data
  } catch (err) {
    inviteError.value = err.response?.data?.errorMessage || 'This invite link is invalid or has expired.'
  } finally {
    inviteLoading.value = false
  }
}

const handleRegister = async () => {
  if (form.password !== form.confirmPassword) {
    error.value = 'Passwords do not match.'
    return
  }
  if (form.password.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return
  }
  loading.value = true
  error.value = null
  try {
    await authAPI.register({
      token: token.value,
      name: form.name.trim(),
      password: form.password
    })
    router.push({ path: '/login', query: { registered: '1' } })
  } catch (err) {
    error.value = err.response?.data?.errorMessage || 'Registration failed. Please try again.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (token.value) fetchInvite()
})
</script>
