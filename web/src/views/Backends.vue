<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('backendsPage.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('backendsPage.subtitle') }}
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" :aria-label="$t('backendsPage.sectionTitle')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('backendsPage.sectionTitle') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('backendsPage.sectionDescription') }}</p>
          </div>
          <button
            type="button"
            @click="openAddModal"
            class="btn-primary text-sm inline-flex items-center gap-2"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('backendsPage.addBackend') }}
          </button>
        </div>
        <div v-if="loading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('backendsPage.loading') }}</div>
        <div v-else-if="backends.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('backendsPage.empty') }}</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('backendsPage.name') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('backendsPage.status') }}</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('backendsPage.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="backend in backends" :key="backend.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <p class="text-sm font-medium text-theme-text">{{ backend.name }}</p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span
                      v-if="backend.locked"
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800"
                      :title="$t('backendsPage.lockedHint')"
                    >
                      <svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                        <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm2-2v2H8V7a3 3 0 016 0v2h-1z" clip-rule="evenodd" />
                      </svg>
                      {{ $t('backendsPage.locked') }}
                    </span>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-theme-hover text-theme-textLight"
                    >
                      <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                      </svg>
                      {{ $t('backendsPage.unlocked') }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      v-if="backend.locked && isPlatformAdmin"
                      type="button"
                      @click="unlockBackend(backend)"
                      :disabled="unlockingId === backend.id"
                      class="text-amber-700 hover:text-amber-800 hover:underline mr-3 font-medium"
                    >
                      {{ unlockingId === backend.id ? $t('backendsPage.unlocking') : $t('backendsPage.forceUnlock') }}
                    </button>
                    <button
                      type="button"
                      @click="openUsageModal(backend)"
                      class="text-primary-600 hover:text-primary-700 hover:underline mr-3"
                    >
                      {{ $t('backendsPage.howToUse') }}
                    </button>
                    <button
                      type="button"
                      @click="openRevealModal(backend)"
                      class="text-primary-600 hover:text-primary-700 hover:underline mr-3"
                    >
                      {{ $t('backendsPage.revealCredentials') }}
                    </button>
                    <button
                      type="button"
                      @click="openEditModal(backend)"
                      class="text-primary-600 hover:text-primary-700 hover:underline mr-3"
                    >
                      {{ $t('backendsPage.edit') }}
                    </button>
                    <button
                      type="button"
                      @click="openDeleteModal(backend)"
                      class="text-red-600 hover:text-red-700 hover:underline"
                    >
                      {{ $t('backendsPage.delete') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="px-6 py-4 border-t border-theme-border flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm text-theme-textLight">
              <template v-if="hasPagination">
                {{ $t('backendsPage.showing', { from: showingFrom, to: showingTo, total: totalBackends }) }}
              </template>
              <template v-else-if="_meta?.total">
                {{ _meta.total === 1 ? $t('backendsPage.countOne') : $t('backendsPage.countMany', { count: _meta.total }) }}
              </template>
            </div>
            <div v-if="hasPagination" class="flex items-center gap-2">
              <button
                type="button"
                @click="goToPage(offset - limit)"
                :disabled="offset === 0"
                class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ $t('backendsPage.previous') }}
              </button>
              <button
                type="button"
                @click="goToPage(offset + limit)"
                :disabled="offset + limit >= totalBackends"
                class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ $t('backendsPage.next') }}
              </button>
            </div>
          </div>
        </template>
      </section>
    </main>

    <!-- Add backend modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="addModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="add-backend-title"
          @click.self="addModalOpen = false"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="add-backend-title" class="text-lg font-semibold text-theme-text">{{ $t('backendsPage.addBackend') }}</h2>
            <form @submit.prevent="saveAddBackend" class="mt-4 space-y-4">
              <div>
                <label for="add-backend-name" class="form-label">{{ $t('backendsPage.name') }}</label>
                <input
                  id="add-backend-name"
                  v-model="addForm.name"
                  type="text"
                  required
                  maxlength="255"
                  class="input-field"
                  placeholder="e.g. production"
                />
              </div>
              <div>
                <label for="add-backend-username" class="form-label">{{ $t('backendsPage.username') }}</label>
                <div class="flex gap-2">
                  <input
                    id="add-backend-username"
                    v-model="addForm.username"
                    type="text"
                    required
                    maxlength="255"
                    class="input-field flex-1"
                    placeholder="e.g. tf-state-user"
                  />
                  <button type="button" @click="generateUsername(addForm)" class="btn-secondary shrink-0" :title="$t('backendsPage.generateUsername')">
                    {{ $t('backendsPage.generate') }}
                  </button>
                </div>
              </div>
              <div>
                <label for="add-backend-password" class="form-label">{{ $t('backendsPage.password') }}</label>
                <div class="flex gap-2">
                  <input
                    id="add-backend-password"
                    v-model="addForm.password"
                    :type="addShowPassword ? 'text' : 'password'"
                    required
                    maxlength="255"
                    class="input-field flex-1"
                  />
                  <button
                    type="button"
                    @click="addShowPassword = !addShowPassword"
                    class="btn-secondary shrink-0 px-3"
                    :title="addShowPassword ? $t('backendsPage.hidePassword') : $t('backendsPage.showPassword')"
                  >
                    <svg v-if="!addShowPassword" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  </button>
                  <button type="button" @click="generatePassword(addForm)" class="btn-secondary shrink-0" :title="$t('backendsPage.generatePassword')">
                    {{ $t('backendsPage.generate') }}
                  </button>
                </div>
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="addModalOpen = false" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="addSaving" class="btn-primary">
                  {{ addSaving ? $t('backendsPage.saving') : $t('backendsPage.addBackend') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Edit backend modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="editModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="edit-backend-title"
          @click.self="editModalOpen = false"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="edit-backend-title" class="text-lg font-semibold text-theme-text">{{ $t('backendsPage.editBackend') }}</h2>
            <form @submit.prevent="saveEditBackend" class="mt-4 space-y-4">
              <div>
                <label for="edit-backend-name" class="form-label">{{ $t('backendsPage.name') }}</label>
                <input
                  id="edit-backend-name"
                  v-model="editForm.name"
                  type="text"
                  required
                  maxlength="255"
                  class="input-field"
                  placeholder="e.g. production"
                />
              </div>
              <div>
                <label for="edit-backend-username" class="form-label">{{ $t('backendsPage.username') }}</label>
                <div class="flex gap-2">
                  <input
                    id="edit-backend-username"
                    v-model="editForm.username"
                    type="text"
                    required
                    maxlength="255"
                    class="input-field flex-1"
                    placeholder="e.g. tf-state-user"
                  />
                  <button type="button" @click="generateUsername(editForm)" class="btn-secondary shrink-0" :title="$t('backendsPage.generateUsername')">
                    {{ $t('backendsPage.generate') }}
                  </button>
                </div>
              </div>
              <div>
                <label for="edit-backend-password" class="form-label">{{ $t('backendsPage.password') }}</label>
                <div class="flex flex-wrap gap-2">
                  <input
                    id="edit-backend-password"
                    v-model="editForm.password"
                    :type="editShowPassword ? 'text' : 'password'"
                    maxlength="255"
                    class="input-field flex-1 min-w-0"
                    :placeholder="$t('backendsPage.passwordOptional')"
                  />
                  <button
                    type="button"
                    @click="editShowPassword = !editShowPassword"
                    class="btn-secondary shrink-0 px-3"
                    :title="editShowPassword ? $t('backendsPage.hidePassword') : $t('backendsPage.showPassword')"
                  >
                    <svg v-if="!editShowPassword" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  </button>
                  <button type="button" @click="generatePassword(editForm)" class="btn-secondary shrink-0" :title="$t('backendsPage.generatePassword')">
                    {{ $t('backendsPage.generate') }}
                  </button>
                </div>
                <p class="mt-1 text-xs text-theme-textLight">{{ $t('backendsPage.passwordOptional') }}</p>
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="editModalOpen = false" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="editSaving" class="btn-primary">
                  {{ editSaving ? $t('backendsPage.saving') : $t('common.save') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- How to use backend modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="usageModalBackend"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="usage-backend-title"
          @click.self="usageModalBackend = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div class="p-6 border-b border-theme-border flex items-center justify-between">
              <h2 id="usage-backend-title" class="text-lg font-semibold text-theme-text">
                {{ $t('backendsPage.howToUse') }}: {{ usageModalBackend?.name }}
              </h2>
              <button
                type="button"
                @click="usageModalBackend = null"
                class="text-theme-textLight hover:text-theme-text p-1 rounded"
                :aria-label="$t('common.cancel')"
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div v-if="usageModalBackend" class="p-6 space-y-6">
              <p class="text-sm text-theme-textLight">{{ $t('backendsPage.usageIntro') }}</p>

              <div>
                <p class="text-sm font-medium text-theme-text mb-1">{{ $t('backendsPage.usageBackendTf') }}</p>
                <p class="text-xs text-theme-textLight mb-2">{{ $t('backendsPage.usageBackendTfNote') }}</p>
                <pre class="bg-theme-hover border border-theme-border rounded-lg p-4 text-xs overflow-x-auto"><code>terraform {
  backend "http" {
    address        = "{{ backendBaseUrl }}/client/{{ usageModalBackend.slug }}/state"
    lock_address   = "{{ backendBaseUrl }}/client/{{ usageModalBackend.slug }}/lock"
    unlock_address = "{{ backendBaseUrl }}/client/{{ usageModalBackend.slug }}/unlock"
    lock_method    = "POST"
    unlock_method  = "POST"
  }
}</code></pre>
              </div>

              <div>
                <p class="text-sm font-medium text-theme-text mb-1">{{ $t('backendsPage.usageEnvVars') }}</p>
                <p class="text-xs text-theme-textLight mb-2">{{ $t('backendsPage.usageEnvVarsNote') }}</p>
                <pre class="bg-theme-hover border border-theme-border rounded-lg p-4 text-xs overflow-x-auto"><code>export TF_HTTP_USERNAME="{{ usageModalBackend.username }}"
export TF_HTTP_PASSWORD="&lt;your-backend-password&gt;"</code></pre>
              </div>

              <div class="flex justify-end pt-2">
                <button type="button" @click="usageModalBackend = null" class="btn-secondary">{{ $t('common.close') }}</button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Reveal credentials modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="revealModalBackend"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="reveal-credentials-title"
          @click.self="revealModalBackend = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="reveal-credentials-title" class="text-lg font-semibold text-theme-text">
              {{ $t('backendsPage.revealCredentialsTitle', { name: revealModalBackend?.name }) }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('backendsPage.revealCredentialsIntro') }}</p>
            <div v-if="revealModalBackend" class="mt-4 space-y-4">
              <div v-if="revealModalPassword === null && !revealModalError">
                <button
                  type="button"
                  @click="fetchRevealPassword"
                  :disabled="revealModalLoading"
                  class="btn-primary w-full"
                >
                  {{ revealModalLoading ? $t('backendsPage.revealing') : $t('backendsPage.revealPassword') }}
                </button>
              </div>
              <p v-if="revealModalError" class="text-sm text-red-600">{{ revealModalError }}</p>
              <template v-else-if="revealModalPassword !== null">
                <div>
                  <label class="form-label">{{ $t('backendsPage.username') }}</label>
                  <div class="flex gap-2">
                    <input
                      :value="revealModalBackend.username"
                      type="text"
                      readonly
                      class="input-field flex-1 bg-theme-hover font-mono text-sm"
                    />
                    <button type="button" @click="copyToClipboard(revealModalBackend.username, 'username')" class="btn-secondary shrink-0">
                      {{ copiedField === 'username' ? $t('backendsPage.copied') : $t('backendsPage.copy') }}
                    </button>
                  </div>
                </div>
                <div>
                  <label class="form-label">{{ $t('backendsPage.password') }}</label>
                  <div class="flex gap-2">
                    <input
                      :value="revealModalPassword"
                      :type="revealShowPassword ? 'text' : 'password'"
                      readonly
                      class="input-field flex-1 bg-theme-hover font-mono text-sm"
                    />
                    <button type="button" @click="revealShowPassword = !revealShowPassword" class="btn-secondary shrink-0 px-3" :title="revealShowPassword ? $t('backendsPage.hidePassword') : $t('backendsPage.showPassword')">
                      <svg v-if="!revealShowPassword" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                      <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                    </button>
                    <button type="button" @click="copyToClipboard(revealModalPassword, 'password')" class="btn-secondary shrink-0">
                      {{ copiedField === 'password' ? $t('backendsPage.copied') : $t('backendsPage.copy') }}
                    </button>
                  </div>
                </div>
              </template>
            </div>
            <div class="flex justify-end mt-6 pt-4 border-t border-theme-border">
              <button type="button" @click="closeRevealModal" class="btn-secondary">{{ $t('common.close') }}</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Delete backend modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteModalBackend"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-backend-title"
          @click.self="deleteModalBackend = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-backend-title" class="text-lg font-semibold text-theme-text">{{ $t('backendsPage.deleteTitle') }}</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('backendsPage.deleteConfirm', { name: deleteModalBackend?.name }) }}
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" @click="deleteModalBackend = null" class="btn-secondary">{{ $t('common.cancel') }}</button>
              <button
                type="button"
                @click="confirmDelete"
                :disabled="deletingId === deleteModalBackend?.id"
                class="btn-danger"
              >
                {{ deletingId === deleteModalBackend?.id ? $t('backendsPage.deleting') : $t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { loadWorkspaceFromStorage } from '@/utils/storage'
import { backendsAPI } from '@/api'
import { user } from '@/lib/auth'

const { t } = useI18n()
const router = useRouter()

const workspace = ref(null)
const backends = ref([])
const _meta = ref(null)
const loading = ref(true)
const limit = ref(20)
const offset = ref(0)
const errorMessage = ref('')
const successMessage = ref('')

const addModalOpen = ref(false)
const addForm = ref({ name: '', username: '', password: '' })
const addSaving = ref(false)
const addShowPassword = ref(false)

const editModalOpen = ref(false)
const editingBackend = ref(null)
const editForm = ref({ name: '', username: '', password: '' })
const editSaving = ref(false)
const editShowPassword = ref(false)

const revealModalBackend = ref(null)
const revealModalPassword = ref(null)
const revealModalLoading = ref(false)
const revealModalError = ref('')
const revealShowPassword = ref(false)
const copiedField = ref(null)

const deleteModalBackend = ref(null)
const deletingId = ref(null)
const unlockingId = ref(null)

const isPlatformAdmin = computed(() => user.value?.role === 'admin')

const usageModalBackend = ref(null)
const backendBaseUrl = typeof window !== 'undefined' ? window.location.origin : ''


function secureRandom(charset, len) {
  const arr = new Uint8Array(len)
  crypto.getRandomValues(arr)
  let s = ''
  for (let i = 0; i < len; i++) s += charset[arr[i] % charset.length]
  return s
}

function generateUsername(formRef) {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  const form = formRef?.value ?? formRef
  if (form) form.username = 'tf-' + secureRandom(chars, 14)
}

function generatePassword(formRef) {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*'
  const form = formRef?.value ?? formRef
  if (form) form.password = secureRandom(chars, 24)
}

function setWorkspace() {
  workspace.value = loadWorkspaceFromStorage()
  if (!workspace.value?.id) {
    router.replace('/select-workspace')
  }
}

async function loadBackends(overrideOffset) {
  if (!workspace.value?.id) return
  loading.value = true
  errorMessage.value = ''
  const off = overrideOffset !== undefined ? overrideOffset : offset.value

  try {
    const res = await backendsAPI.list(
      workspace.value.id,
      { limit: limit.value, offset: off },
    )
    backends.value = res.data?.backends ?? []
    _meta.value = res.data?._meta ?? null
    offset.value = off
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('backendsPage.failedLoad')
    backends.value = []
  } finally {
    loading.value = false
  }
}

const totalBackends = computed(() => _meta.value?.total ?? 0)
const showingFrom = computed(() => (totalBackends.value === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + backends.value.length, totalBackends.value))
const hasPagination = computed(() => totalBackends.value > limit.value)

function goToPage(newOffset) {
  const maxOffset = Math.max(0, totalBackends.value - limit.value)
  const off = Math.max(0, Math.min(newOffset, maxOffset))
  loadBackends(off)
}

function openAddModal() {
  addForm.value = {
    name: '',
    username: '',
    password: '',
  }
  addShowPassword.value = false
  addModalOpen.value = true
}

function openEditModal(backend) {
  editingBackend.value = backend
  editForm.value = {
    name: backend.name,
    username: backend.username,
    password: ''
  }
  editShowPassword.value = false
  editModalOpen.value = true
}

function openRevealModal(backend) {
  revealModalBackend.value = backend
  revealModalPassword.value = null
  revealModalError.value = ''
  revealShowPassword.value = false
  copiedField.value = null
}

function closeRevealModal() {
  revealModalBackend.value = null
  revealModalPassword.value = null
  revealModalError.value = ''
}

async function fetchRevealPassword() {
  if (!workspace.value?.id || !revealModalBackend.value?.id) return
  revealModalLoading.value = true
  revealModalError.value = ''

  try {
    const res = await backendsAPI.revealPassword(
      workspace.value.id,
      revealModalBackend.value.id,
    )
    const pwd = res.data?.password
    revealModalPassword.value = pwd != null ? pwd : ''
  } catch (err) {
    revealModalError.value = err.response?.data?.errorMessage || t('backendsPage.failedReveal')
  } finally {
    revealModalLoading.value = false
  }
}

async function copyToClipboard(text, field) {
  try {
    await navigator.clipboard.writeText(text)
    copiedField.value = field
    setTimeout(() => { copiedField.value = null }, 2000)
  } catch (_) {}
}

async function saveAddBackend() {
  if (!workspace.value?.id) return
  addSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    await backendsAPI.create(workspace.value.id, {
      name: addForm.value.name.trim(),
      username: addForm.value.username.trim(),
      password: addForm.value.password.trim(),
    })
    successMessage.value = t('backendsPage.backendAdded')
    addModalOpen.value = false
    await loadBackends()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('backendsPage.failedAdd')
  } finally {
    addSaving.value = false
  }
}

async function saveEditBackend() {
  if (!workspace.value?.id || !editingBackend.value?.id) return
  editSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const payload = {
      name: editForm.value.name.trim(),
      username: editForm.value.username.trim(),
    }
    if (editForm.value.password?.trim()) { payload.password = editForm.value.password.trim() }
    await backendsAPI.update(
      workspace.value.id,
      editingBackend.value.id,
      payload,
    )
    successMessage.value = t('backendsPage.backendUpdated')
    editModalOpen.value = false
    await loadBackends()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('backendsPage.failedUpdate')
  } finally {
    editSaving.value = false
  }
}

function openDeleteModal(backend) {
  deleteModalBackend.value = backend
}

async function unlockBackend(backend) {
  if (!workspace.value?.id || !backend?.id || !isPlatformAdmin.value) return
  unlockingId.value = backend.id
  errorMessage.value = ''
  successMessage.value = ''

  try {
    await backendsAPI.unlock(workspace.value.id, backend.id)
    successMessage.value = t('backendsPage.unlockSuccess')
    await loadBackends()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('backendsPage.failedUnlock')
  } finally {
    unlockingId.value = null
  }
}

function openUsageModal(backend) {
  usageModalBackend.value = backend
}

async function confirmDelete() {
  const backend = deleteModalBackend.value
  if (!workspace.value?.id || !backend?.id) return
  deletingId.value = backend.id
  errorMessage.value = ''

  try {
    await backendsAPI.delete(workspace.value.id, backend.id)
    successMessage.value = t('backendsPage.backendDeleted')
    deleteModalBackend.value = null
    await loadBackends()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('backendsPage.failedDelete')
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  setWorkspace()
  if (workspace.value?.id) {
    loadBackends()
  }
})

watch(workspace, (w) => {
  if (w?.id) loadBackends()
}, { immediate: false })
</script>
