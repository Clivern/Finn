<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />
    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ title }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ description }}
        </p>
      </header>
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppNav from '@/components/AppNav.vue'

const route = useRoute()

const title = computed(() => route.meta?.title ?? route.name ?? 'Page')
const description = computed(() => route.meta?.description ?? '')
</script>
