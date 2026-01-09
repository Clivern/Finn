<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Integrations
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Connect external services and manage integrations
        </p>
      </header>

      <section class="section" aria-label="Integrations overview">
        <h2 class="section-title">Overview</h2>
        <p class="text-sm text-theme-textLight leading-relaxed mb-4">
          Integrations connect Tivo with third-party tools: version control, notification channels,
          cloud providers, and APIs. Configure credentials and endpoints here.
        </p>
        <div class="rounded-lg border border-theme-border border-dashed bg-theme-hover/30 p-8 text-center">
          <p class="text-sm text-theme-textLight">No integrations configured</p>
          <p class="mt-1 text-xs text-theme-textLight">Add an integration to connect external services</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
</script>
