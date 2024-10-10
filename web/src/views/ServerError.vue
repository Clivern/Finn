<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-md w-full text-center">
      <div class="mb-8">
        <router-link to="/">
          <img src="/logo.png" alt="Finn Logo" class="h-24 w-auto mx-auto">
        </router-link>
      </div>
      <p class="text-6xl font-semibold text-theme-textLight mb-2">500</p>
      <h1 class="text-xl font-semibold text-theme-text mb-2">{{ $t('serverErrorPage.title') }}</h1>
      <p class="text-sm text-theme-textLight mb-6">
        {{ $t('serverErrorPage.description') }}
      </p>
      <div class="flex flex-col sm:flex-row gap-3 justify-center">
        <button type="button" class="btn-primary" @click="retry">
          {{ $t('serverErrorPage.tryAgain') }}
        </button>
        <router-link to="/dashboard" class="btn-secondary">
          {{ $t('serverErrorPage.goToDashboard') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
function retry() {
  window.location.reload()
}
</script>
