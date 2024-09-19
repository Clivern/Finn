<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Repositories
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Manage Git repositories connected to this workspace
        </p>
      </header>

      <section class="section" aria-label="Repositories overview">
        <h2 class="section-title">Overview</h2>
        <p class="text-sm text-theme-textLight leading-relaxed mb-4">
          Repositories store your code, Terraform configs, and playbooks. Connect repositories to enable
          runs, pull latest changes, and manage access via SSH keys.
        </p>
        <div class="rounded-lg border border-theme-border border-dashed bg-theme-hover/30 p-8 text-center">
          <p class="text-sm text-theme-textLight">No repositories linked</p>
          <p class="mt-1 text-xs text-theme-textLight">Add a repository to start running workflows</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
</script>
