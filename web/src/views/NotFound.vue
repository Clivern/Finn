<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-md w-full text-center">
      <div class="mb-8">
        <router-link to="/">
          <img src="/logo.png" alt="Lapin Logo" class="h-24 w-auto mx-auto">
        </router-link>
      </div>
      <p class="text-6xl font-semibold text-theme-textLight mb-2">404</p>
      <h1 class="text-xl font-semibold text-theme-text mb-2">Page not found</h1>
      <p class="text-sm text-theme-textLight mb-6">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <div class="flex flex-col sm:flex-row gap-3 justify-center">
        <router-link to="/dashboard" class="btn-primary">
          Go to Dashboard
        </router-link>
        <router-link to="/login" class="btn-secondary">
          Sign in
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
// 404 Not Found page
</script>
