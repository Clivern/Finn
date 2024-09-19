<template>
  <nav class="bg-white border-b border-theme-border">
    <div class="w-full px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-14">
        <div class="flex items-center">
          <div class="flex-shrink-0 flex items-center">
            <img src="/logo.png" alt="Met Logo" class="h-8 w-auto">
          </div>
          <div class="hidden md:ml-8 md:flex md:space-x-1">
            <router-link
              v-for="item in navItems"
              :key="item.to"
              :to="item.to"
              :class="[isActive(item.to) ? 'nav-link-active' : 'nav-link']"
            >
              {{ item.label }}
            </router-link>
          </div>
        </div>
        <div class="flex items-center">
          <UserAvatarDropdown />
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { useRoute } from 'vue-router'
import UserAvatarDropdown from '@/components/UserAvatarDropdown.vue'

const route = useRoute()

const navItems = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/task-templates', label: 'Task Templates' },
  { to: '/inventory', label: 'Inventory' },
  { to: '/environment', label: 'Environment' },
  { to: '/repositories', label: 'Repositories' },
  { to: '/tf-backends', label: 'Terraform Backends' }
]

function isActive(path) {
  return route.path === path || (path !== '/dashboard' && route.path.startsWith(path + '/'))
}
</script>
