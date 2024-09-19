<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Inventory
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Manage inventory and resource tracking for your workspaces
        </p>
      </header>

      <section class="section" aria-label="Inventory overview">
        <h2 class="section-title">Overview</h2>
        <p class="text-sm text-theme-textLight leading-relaxed mb-4">
          Track hosts, groups, and variables used by your automation. Inventory can be static or dynamic
          and is used to target runs and pass configuration to tasks.
        </p>
        <div class="rounded-lg border border-theme-border border-dashed bg-theme-hover/30 p-8 text-center">
          <p class="text-sm text-theme-textLight">No inventory defined</p>
          <p class="mt-1 text-xs text-theme-textLight">Add inventory sources to get started</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
</script>
