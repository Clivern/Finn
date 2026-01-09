<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Terraform Backends
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Configure and manage Terraform state backends
        </p>
      </header>

      <section class="section" aria-label="TF backends overview">
        <h2 class="section-title">Overview</h2>
        <p class="text-sm text-theme-textLight leading-relaxed mb-4">
          Terraform backends store state and lock it during operations. Configure remote backends
          (e.g. S3, GCS, Azure) so state is shared and protected across runs and team members.
        </p>
        <div class="rounded-lg border border-theme-border border-dashed bg-theme-hover/30 p-8 text-center">
          <p class="text-sm text-theme-textLight">No backends configured</p>
          <p class="mt-1 text-xs text-theme-textLight">Add a backend to store Terraform state remotely</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
</script>
