<!--
  Flows – Visual flow builder for chaining tasks.
  Example: run Terraform task → parse outputs → run Ansible task with inventory from Terraform output.
  Backend and flow execution to be implemented later.
-->
<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('flowsPage.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('flowsPage.subtitle') }}
        </p>
      </header>

      <section class="section" :aria-label="$t('flowsPage.overview')">
        <h2 class="section-title">{{ $t('flowsPage.overview') }}</h2>
        <p class="text-sm text-theme-textLight leading-relaxed mb-4">
          {{ $t('flowsPage.overviewDescription') }}
        </p>
        <div class="rounded-lg border border-theme-border border-dashed bg-theme-hover/30 p-8 text-center">
          <p class="text-sm text-theme-textLight">{{ $t('flowsPage.empty') }}</p>
          <p class="mt-1 text-xs text-theme-textLight">{{ $t('flowsPage.emptyHint') }}</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
</script>
