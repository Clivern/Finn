<template>
  <div class="relative" ref="dropdownRef">
    <button
      type="button"
      @click.stop="open = !open"
      class="flex items-center gap-2 rounded-full border border-theme-border bg-white pl-1 pr-2.5 py-1 text-left hover:bg-theme-hover focus:outline-none focus:ring-2 focus:ring-primary-800 focus:ring-offset-1 transition-colors"
      :aria-expanded="open"
      aria-haspopup="true"
      aria-label="User menu"
    >
      <span
        class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary-200 text-sm font-medium text-theme-text"
        aria-hidden="true"
      >
        {{ userInitials }}
      </span>
      <svg
        class="h-4 w-4 text-theme-textLight transition-transform"
        :class="{ 'rotate-180': open }"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        aria-hidden="true"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <Transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-show="open"
        class="absolute right-0 z-50 mt-2 w-56 origin-top-right rounded-lg border border-theme-border bg-white py-1 shadow-lg focus:outline-none"
        role="menu"
      >
        <div class="border-b border-theme-border px-4 py-3">
          <p class="text-sm font-semibold text-theme-text truncate">
            {{ displayName }}
          </p>
          <p class="text-sm text-theme-textLight truncate">
            {{ authStore.currentUser?.email ?? '' }}
          </p>
        </div>
        <div class="py-1">
          <router-link
            to="/settings"
            class="block px-4 py-2 text-sm text-theme-text hover:bg-theme-hover"
            role="menuitem"
            @click="open = false"
          >
            Profile
          </router-link>
          <router-link
            to="/workspaces"
            class="block px-4 py-2 text-sm text-theme-text hover:bg-theme-hover"
            role="menuitem"
            @click="open = false"
          >
            Switch Workspace
          </router-link>
          <button
            type="button"
            class="block w-full px-4 py-2 text-left text-sm text-theme-text hover:bg-theme-hover"
            role="menuitem"
            @click="handleLogout"
          >
            Logout
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const dropdownRef = ref(null)
const open = ref(false)

const displayName = computed(() => {
  const u = authStore.currentUser
  if (!u) return ''
  if (u.name) return u.name
  if (u.email) {
    const local = u.email.split('@')[0]
    return local.charAt(0).toUpperCase() + local.slice(1).replace(/[._]/g, ' ')
  }
  return 'User'
})

const userInitials = computed(() => {
  const name = displayName.value
  if (!name) return '?'
  const parts = name.trim().split(/\s+/)
  if (parts.length >= 2) {
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase().slice(0, 2)
  }
  return name.slice(0, 2).toUpperCase()
})

function handleLogout() {
  open.value = false
  authStore.logout()
  router.push('/login')
}

function onClickOutside(event) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    open.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
})
</script>
