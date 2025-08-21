<template>
  <div id="app" class="min-h-screen">
    <router-view />
    <Assistant v-if="showAssistant" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Assistant from '@/components/Assistant.vue'

const route = useRoute()
const showAssistant = computed(() => !!route.meta?.requiresAuth)
</script>

<style scoped>
/* Additional app-level styles can go here */
</style>

