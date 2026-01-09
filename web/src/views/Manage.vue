<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <div class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <main>
        <div class="mb-8">
          <h1 class="text-2xl font-semibold text-theme-text">Manage</h1>
          <p class="text-sm text-theme-textLight mt-1">Platform management and configuration</p>
        </div>

        <!-- Manage: quick access to admin areas -->
        <section class="mb-10" aria-label="Manage">
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <router-link
              to="/workspaces"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Workspaces</p>
                <p class="text-xs text-theme-textLight truncate">Manage workspaces</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/users"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Users</p>
                <p class="text-xs text-theme-textLight truncate">User accounts and roles</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/invites"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Invites</p>
                <p class="text-xs text-theme-textLight truncate">User invites and sign-up links</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/integrations"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Integrations</p>
                <p class="text-xs text-theme-textLight truncate">External services and APIs</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/runners"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Runners</p>
                <p class="text-xs text-theme-textLight truncate">Execution runners</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
            <router-link
              to="/settings"
              class="flex items-center gap-4 rounded-xl border border-theme-border bg-white p-5 shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors group"
            >
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text group-hover:bg-primary-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-theme-text">Settings</p>
                <p class="text-xs text-theme-textLight truncate">Platform and email configuration</p>
              </div>
              <svg class="h-4 w-4 text-theme-textLight flex-shrink-0 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </router-link>
          </div>
        </section>

      </main>
    </div>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
</script>
