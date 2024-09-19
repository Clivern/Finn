<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Task Templates
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Create and manage reusable task templates for your workflows
        </p>
      </header>

      <section class="section" aria-label="Task templates overview">
        <h2 class="section-title">Overview</h2>
        <p class="text-sm text-theme-textLight leading-relaxed mb-4">
          Task templates define reusable steps, inputs, and configurations that can be applied across runs.
          Use them to standardize common operations and keep your workflows consistent.
        </p>
        <div class="rounded-lg border border-theme-border border-dashed bg-theme-hover/30 p-8 text-center">
          <p class="text-sm text-theme-textLight">No task templates yet</p>
          <p class="mt-1 text-xs text-theme-textLight">Task templates you create will appear here</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
</script>
