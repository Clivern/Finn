<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-semibold text-theme-text">{{ $t('runners.title') }}</h1>
          <p class="text-theme-textLight mt-2">{{ $t('runnersPage.subtitle') }}</p>
        </div>
        <button
          @click="openCreateModal"
          class="btn-primary"
        >
          {{ $t('runnersPage.createRunner') }}
        </button>
      </div>

      <div v-if="successMessage" class="mb-6 rounded-md border border-green-200 bg-green-50 p-4">
        <p class="text-sm text-green-800">{{ successMessage }}</p>
      </div>
      <div v-if="errorMessage" class="mb-6 rounded-md border border-red-200 bg-red-50 p-4">
        <p class="text-sm text-red-800">{{ errorMessage }}</p>
      </div>

      <div class="bg-white rounded-lg border border-theme-border overflow-hidden">
        <div v-if="loading" class="p-8 text-center">
          <svg class="animate-spin h-6 w-6 mx-auto text-theme-text" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <p class="text-theme-textLight mt-2">{{ $t('runnersPage.loading') }}</p>
        </div>

        <div v-else-if="runners.length === 0" class="p-8 text-center">
          <p class="text-theme-textLight">{{ $t('runnersPage.noRunners') }}</p>
          <p class="text-sm text-theme-textLight mt-1">{{ $t('runnersPage.noRunnersHint') }}</p>
        </div>

        <table v-else class="min-w-full divide-y divide-theme-border">
          <thead class="bg-theme-hover">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('runnersPage.name') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('runnersPage.token') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('runnersPage.tags') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('runnersPage.status') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('runnersPage.connection') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('runnersPage.created') }}
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('runnersPage.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-theme-border">
            <tr v-for="runner in runners" :key="runner.id" class="hover:bg-theme-hover">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-theme-text">{{ runner.name || '—' }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <code class="text-xs text-theme-textLight font-mono">{{ maskToken(runner.token) }}</code>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ runner.tags || '—' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="{
                    'bg-green-100 text-green-800': runner.status === 'active',
                    'bg-gray-100 text-gray-800': runner.status === 'inactive',
                    'bg-yellow-100 text-yellow-800': runner.status === 'maintenance'
                  }">
                  {{ runner.status === 'active' ? $t('runnersPage.statusActive') : runner.status === 'inactive' ? $t('runnersPage.statusInactive') : $t('runnersPage.statusMaintenance') }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  v-if="connectionStatus(runner.id) === null && canPingRunner(runner)"
                  class="inline-flex items-center gap-1.5 text-theme-textLight"
                  :title="$t('runnersPage.checkingTitle')"
                >
                  <span class="inline-block w-2.5 h-2.5 rounded-full bg-theme-textLight/40 animate-pulse" aria-hidden="true" />
                  <span class="text-xs">{{ $t('runnersPage.checking') }}</span>
                </span>
                <span
                  v-else-if="connectionStatus(runner.id) === 'ok'"
                  class="inline-flex items-center gap-1.5"
                  :title="$t('runnersPage.connected')"
                >
                  <span class="inline-block w-2.5 h-2.5 rounded-full bg-green-500" aria-hidden="true" />
                  <span class="text-xs text-green-700">{{ $t('runnersPage.connected') }}</span>
                </span>
                <span
                  v-else-if="connectionStatus(runner.id)?.failed"
                  class="inline-flex items-center gap-1.5 cursor-help"
                  :title="connectionStatus(runner.id).error"
                >
                  <span class="inline-block w-2.5 h-2.5 rounded-full bg-red-500" aria-hidden="true" />
                  <span class="text-xs text-red-700">{{ $t('runnersPage.failed') }}</span>
                </span>
                <span v-else class="text-sm text-theme-textLight">—</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatRelative(runner.createdAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  @click="openEditModal(runner)"
                  class="text-blue-600 hover:text-blue-900 mr-4"
                >
                  {{ $t('common.edit') }}
                </button>
                <button
                  @click="openDeleteModal(runner)"
                  class="text-red-600 hover:text-red-900"
                >
                  {{ $t('common.delete') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="runners.length > 0" class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between">
          <div class="text-sm text-theme-textLight">
            {{ $t('runnersPage.showing', { from: showingFrom, to: showingTo, total }) }}
          </div>
          <div class="flex items-center space-x-2">
            <button
              @click="goToPage(offset - limit)"
              :disabled="offset === 0"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('runnersPage.previous') }}
            </button>
            <button
              @click="goToPage(offset + limit)"
              :disabled="offset + limit >= total"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('runnersPage.next') }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Create runner modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeCreateModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeCreateModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <form @submit.prevent="handleCreate">
            <div class="bg-white px-6 pt-6 pb-4">
              <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold text-theme-text">{{ $t('runnersPage.createRunnerTitle') }}</h3>
                <button type="button" @click="closeCreateModal" class="text-theme-textLight hover:text-theme-text">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div class="space-y-4">
                <div>
                  <label for="create-name" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.name') }}</label>
                  <input id="create-name" v-model="createForm.name" type="text" class="input-field" :placeholder="$t('runnersPage.namePlaceholder')" />
                </div>
                <div>
                  <label for="create-tags" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.tags') }}</label>
                  <input id="create-tags" v-model="createForm.tags" type="text" class="input-field" :placeholder="$t('runnersPage.tagsPlaceholder')" />
                </div>
                <div>
                  <label for="create-maxParallel" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.maxParallelTasks') }}</label>
                  <input id="create-maxParallel" v-model.number="createForm.maxParallelTasks" type="number" min="1" class="input-field" />
                </div>
                <div>
                  <label for="create-status" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.status') }}</label>
                  <select id="create-status" v-model="createForm.status" class="input-field">
                    <option value="active">{{ $t('runnersPage.statusActive') }}</option>
                    <option value="maintenance">{{ $t('runnersPage.statusMaintenance') }}</option>
                    <option value="inactive">{{ $t('runnersPage.statusInactive') }}</option>
                  </select>
                </div>
                <div>
                  <label for="create-url" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.url') }}</label>
                  <input id="create-url" v-model="createForm.url" type="url" class="input-field" :placeholder="$t('runnersPage.urlPlaceholder')" />
                </div>
                <div>
                  <label for="create-token" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.token') }}</label>
                  <input id="create-token" v-model="createForm.token" type="text" class="input-field font-mono text-sm" :placeholder="$t('runnersPage.tokenPlaceholder')" />
                </div>
              </div>
            </div>
            <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
              <button type="button" @click="closeCreateModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
              <button type="submit" class="btn-primary" :disabled="createLoading">
                <span v-if="!createLoading">{{ $t('runnersPage.createRunnerSubmit') }}</span>
                <span v-else class="inline-flex items-center gap-2">
                  <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  {{ $t('runnersPage.creating') }}
                </span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Edit runner modal -->
    <div v-if="showEditModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeEditModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeEditModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <form @submit.prevent="handleUpdate">
            <div class="bg-white px-6 pt-6 pb-4">
              <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold text-theme-text">{{ $t('runnersPage.editRunnerTitle') }}</h3>
                <button type="button" @click="closeEditModal" class="text-theme-textLight hover:text-theme-text">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div class="space-y-4">
                <div>
                  <label for="edit-name" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.name') }}</label>
                  <input id="edit-name" v-model="editForm.name" type="text" class="input-field" :placeholder="$t('runnersPage.namePlaceholder')" />
                </div>
                <div>
                  <label for="edit-tags" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.tags') }}</label>
                  <input id="edit-tags" v-model="editForm.tags" type="text" class="input-field" :placeholder="$t('runnersPage.tagsPlaceholder')" />
                </div>
                <div>
                  <label for="edit-maxParallel" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.maxParallelTasks') }}</label>
                  <input id="edit-maxParallel" v-model.number="editForm.maxParallelTasks" type="number" min="1" class="input-field" />
                </div>
                <div>
                  <label for="edit-status" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.status') }}</label>
                  <select id="edit-status" v-model="editForm.status" class="input-field">
                    <option value="active">{{ $t('runnersPage.statusActive') }}</option>
                    <option value="inactive">{{ $t('runnersPage.statusInactive') }}</option>
                    <option value="maintenance">{{ $t('runnersPage.statusMaintenance') }}</option>
                  </select>
                </div>
                <div>
                  <label for="edit-url" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.url') }}</label>
                  <input id="edit-url" v-model="editForm.url" type="url" class="input-field" :placeholder="$t('runnersPage.urlPlaceholder')" />
                </div>
                <div>
                  <label for="edit-token" class="block text-sm font-medium text-theme-text mb-2">{{ $t('runnersPage.token') }}</label>
                  <input id="edit-token" v-model="editForm.token" type="text" class="input-field font-mono text-sm" :placeholder="$t('runnersPage.tokenPlaceholder')" />
                </div>
              </div>
            </div>
            <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
              <button type="button" @click="closeEditModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
              <button type="submit" class="btn-primary" :disabled="updateLoading">
                <span v-if="!updateLoading">{{ $t('common.save') }}</span>
                <span v-else class="inline-flex items-center gap-2">
                  <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  {{ $t('runnersPage.saving') }}
                </span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Delete confirmation modal -->
    <div v-if="showDeleteModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeDeleteModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeDeleteModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div class="bg-white px-6 pt-6 pb-4">
            <div class="flex justify-between items-center mb-4">
              <h3 class="text-lg font-semibold text-theme-text">{{ $t('runnersPage.deleteRunnerTitle') }}</h3>
              <button type="button" @click="closeDeleteModal" class="text-theme-textLight hover:text-theme-text">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <p class="text-sm text-theme-textLight">
              {{ $t('runnersPage.deleteConfirm', { name: runnerToDelete?.name }) }}
            </p>
          </div>
          <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
            <button type="button" @click="closeDeleteModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
            <button
              @click="handleDelete"
              class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700 focus:outline-none transition-all duration-150"
              :disabled="deleteLoading"
            >
              <span v-if="!deleteLoading">{{ $t('common.delete') }}</span>
              <span v-else class="flex items-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                {{ $t('runnersPage.deleting') }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { runnersAPI } from '@/api'
import AppNav from '@/components/AppNav.vue'
import { formatRelative } from '@/utils/date'

const { t } = useI18n()
const loading = ref(false)
const runners = ref([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const successMessage = ref(null)
const errorMessage = ref(null)

const connectionStatusMap = ref({})

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const createLoading = ref(false)
const updateLoading = ref(false)
const deleteLoading = ref(false)
const runnerToDelete = ref(null)
const runnerToEdit = ref(null)

const createForm = reactive({
  token: '',
  name: '',
  tags: '',
  maxParallelTasks: 10,
  status: 'active',
  url: ''
})

const editForm = reactive({
  token: '',
  name: '',
  tags: '',
  maxParallelTasks: 10,
  status: 'active',
  url: ''
})

const showingFrom = computed(() => (total.value === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + runners.value.length, total.value))

function canPingRunner(runner) {
  return runner?.url && runner?.token
}

function connectionStatus(runnerId) {
  return connectionStatusMap.value[runnerId]
}

async function pingRunner(runnerId) {
  connectionStatusMap.value[runnerId] = null

  try {
    const res = await runnersAPI.ping(runnerId)
    const status = res.data?.status
    if (status === 'ok') {
      connectionStatusMap.value[runnerId] = 'ok'
    } else {
      connectionStatusMap.value[runnerId] = {
        failed: true,
        error: res.data?.errorMessage || t('runnersPage.pingFailed')
      }
    }
  } catch (err) {
    connectionStatusMap.value[runnerId] = {
      failed: true,
      error: err.response?.data?.errorMessage || t('runnersPage.connectionFailed')
    }
  }
}

function loadConnectionStatuses() {
  connectionStatusMap.value = {}
  runners.value.forEach((r) => {
    if (canPingRunner(r)) {
      pingRunner(r.id)
    }
  })
}

function maskToken(token) {
  if (!token || typeof token !== 'string') return '—'
  const tkn = token.trim()
  if (tkn.length < 12) return '••••••••'
  if (tkn.length === 36 && /^[0-9a-f-]{36}$/i.test(tkn)) {
    return tkn.slice(0, 8) + '…' + tkn.slice(-4)
  }
  return tkn.slice(0, 4) + '…' + tkn.slice(-8)
}

async function loadRunners(overrideOffset) {
  loading.value = true
  errorMessage.value = null
  const off = overrideOffset !== undefined ? overrideOffset : offset.value

  try {
    const response = await runnersAPI.list({ limit: limit.value, offset: off })
    runners.value = response.data?.runners ?? []
    total.value = response.data?._meta?.total ?? runners.value.length
    offset.value = off
    loadConnectionStatuses()
  } catch (err) {
    console.error('Failed to load runners:', err)
    errorMessage.value = err.response?.data?.errorMessage || t('runnersPage.failedLoad')
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  const maxOffset = Math.max(0, total.value - limit.value)
  const off = Math.max(0, Math.min(newOffset, maxOffset))
  loadRunners(off)
}

function openCreateModal() {
  createForm.token = ''
  createForm.name = ''
  createForm.tags = ''
  createForm.maxParallelTasks = 10
  createForm.status = 'active'
  createForm.url = ''
  showCreateModal.value = true
}

function closeCreateModal() {
  showCreateModal.value = false
}

function openEditModal(runner) {
  runnerToEdit.value = runner
  editForm.token = runner.token
  editForm.name = runner.name
  editForm.tags = runner.tags
  editForm.maxParallelTasks = runner.maxParallelTasks ?? 10
  editForm.status = runner.status
  editForm.url = runner.url
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
  runnerToEdit.value = null
}

function openDeleteModal(runner) {
  runnerToDelete.value = runner
  showDeleteModal.value = true
}

function closeDeleteModal() {
  showDeleteModal.value = false
  runnerToDelete.value = null
}

async function handleCreate() {
  createLoading.value = true
  errorMessage.value = null
  successMessage.value = null

  try {
    await runnersAPI.create({
      token: createForm.token.trim(),
      name: createForm.name,
      tags: createForm.tags,
      maxParallelTasks: createForm.maxParallelTasks,
      status: createForm.status,
      url: createForm.url
    })
    successMessage.value = t('runnersPage.runnerCreated')
    closeCreateModal()
    await loadRunners()
    setTimeout(() => { successMessage.value = null }, 5000)
  } catch (err) {
    console.error('Failed to create runner:', err)
    errorMessage.value = err.response?.data?.errorMessage || t('runnersPage.failedCreate')
  } finally {
    createLoading.value = false
  }
}

async function handleUpdate() {
  if (!runnerToEdit.value?.id) return
  updateLoading.value = true
  errorMessage.value = null

  try {
    await runnersAPI.update(runnerToEdit.value.id, {
      token: editForm.token.trim(),
      name: editForm.name,
      tags: editForm.tags,
      maxParallelTasks: editForm.maxParallelTasks,
      status: editForm.status,
      url: editForm.url
    })
    successMessage.value = t('runnersPage.runnerUpdated')
    closeEditModal()
    await loadRunners()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to update runner:', err)
    errorMessage.value = err.response?.data?.errorMessage || t('runnersPage.failedUpdate')
  } finally {
    updateLoading.value = false
  }
}

async function handleDelete() {
  if (!runnerToDelete.value?.id) return
  deleteLoading.value = true
  errorMessage.value = null

  try {
    await runnersAPI.delete(runnerToDelete.value.id)
    successMessage.value = t('runnersPage.runnerDeleted')
    closeDeleteModal()
    await loadRunners()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to delete runner:', err)
    errorMessage.value = err.response?.data?.errorMessage || t('runnersPage.failedDelete')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(loadRunners)
</script>
