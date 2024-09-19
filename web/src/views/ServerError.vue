<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-md w-full text-center">
      <div class="mb-8">
        <img src="/logo.png" alt="Met Logo" class="h-16 w-auto mx-auto">
      </div>
      <p class="text-6xl font-semibold text-theme-textLight mb-2">500</p>
      <h1 class="text-xl font-semibold text-theme-text mb-2">Something went wrong</h1>
      <p class="text-sm text-theme-textLight mb-6">
        We're sorry, but something went wrong on our end. Please try again later.
      </p>
      <div class="flex flex-col sm:flex-row gap-3 justify-center">
        <button type="button" class="btn-primary" @click="retry">
          Try again
        </button>
        <router-link to="/dashboard" class="btn-secondary">
          Go to Dashboard
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
