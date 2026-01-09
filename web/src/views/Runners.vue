<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-semibold text-theme-text">Runners</h1>
          <p class="text-theme-textLight mt-2">Manage runners that execute your tasks and workflows</p>
        </div>
        <button
          @click="openCreateModal"
          class="btn-primary"
        >
          Create runner
        </button>
      </div>

      <div v-if="successMessage" class="mb-6 rounded-md border border-green-200 bg-green-50 p-4">
        <p class="text-sm text-green-800">{{ successMessage }}</p>
      </div>
      <div v-if="createdRunnerToken" class="mb-6 rounded-lg border border-primary-200 bg-primary-50 p-4">
        <p class="text-sm font-medium text-primary-900 mb-1">Runner created</p>
        <p class="text-xs text-primary-800 mb-2">Use this token when registering the runner. Copy it now — it won’t be shown again in full.</p>
        <div class="flex flex-wrap items-center gap-2">
          <code class="flex-1 min-w-0 text-xs text-primary-900 bg-white/80 px-2 py-1.5 rounded border border-primary-200 truncate font-mono">{{ createdRunnerToken }}</code>
          <button
            type="button"
            @click="copyCreatedToken"
            class="btn-secondary text-sm shrink-0"
          >
            Copy token
          </button>
          <button
            type="button"
            @click="dismissTokenCard"
            class="text-sm text-primary-700 hover:text-primary-900 shrink-0"
          >
            Dismiss
          </button>
        </div>
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
          <p class="text-theme-textLight mt-2">Loading runners...</p>
        </div>

        <div v-else-if="runners.length === 0" class="p-8 text-center">
          <p class="text-theme-textLight">No runners registered</p>
          <p class="text-sm text-theme-textLight mt-1">Create a runner to get a token, then install and register the runner to execute jobs.</p>
        </div>

        <table v-else class="min-w-full divide-y divide-theme-border">
          <thead class="bg-theme-hover">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Name
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Token
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Tags
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Status
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Created
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Actions
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
                <button
                  type="button"
                  @click="copyToken(runner.token)"
                  class="ml-1 text-blue-600 hover:text-blue-900 text-xs"
                >
                  Copy
                </button>
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
                  {{ runner.status === 'active' ? 'Active' : runner.status === 'inactive' ? 'Inactive' : 'Maintenance' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatRelative(runner.createdAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  @click="openEditModal(runner)"
                  class="text-blue-600 hover:text-blue-900 mr-4"
                >
                  Edit
                </button>
                <button
                  @click="openDeleteModal(runner)"
                  class="text-red-600 hover:text-red-900"
                >
                  Delete
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="runners.length > 0" class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between">
          <div class="text-sm text-theme-textLight">
            Showing {{ showingFrom }} to {{ showingTo }} of {{ total }} runners
          </div>
          <div class="flex items-center space-x-2">
            <button
              @click="goToPage(offset - limit)"
              :disabled="offset === 0"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              @click="goToPage(offset + limit)"
              :disabled="offset + limit >= total"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
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
                <h3 class="text-lg font-semibold text-theme-text">Create Runner</h3>
                <button type="button" @click="closeCreateModal" class="text-theme-textLight hover:text-theme-text">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div class="space-y-4">
                <div>
                  <label for="create-name" class="block text-sm font-medium text-theme-text mb-2">Name</label>
                  <input id="create-name" v-model="createForm.name" type="text" class="input-field" placeholder="e.g. prod-runner-1" />
                </div>
                <div>
                  <label for="create-tags" class="block text-sm font-medium text-theme-text mb-2">Tags</label>
                  <input id="create-tags" v-model="createForm.tags" type="text" class="input-field" placeholder="e.g. linux, docker" />
                </div>
                <div>
                  <label for="create-maxParallel" class="block text-sm font-medium text-theme-text mb-2">Max parallel tasks</label>
                  <input id="create-maxParallel" v-model.number="createForm.maxParallelTasks" type="number" min="1" class="input-field" />
                </div>
                <div>
                  <label for="create-status" class="block text-sm font-medium text-theme-text mb-2">Status</label>
                  <select id="create-status" v-model="createForm.status" class="input-field">
                    <option value="active">Active</option>
                    <option value="maintenance">Maintenance</option>
                    <option value="inactive">Inactive</option>
                  </select>
                </div>
                <div>
                  <label for="create-url" class="block text-sm font-medium text-theme-text mb-2">URL</label>
                  <input id="create-url" v-model="createForm.url" type="url" class="input-field" placeholder="https://runner.example.com" />
                </div>
                <div>
                  <label for="create-publicKey" class="block text-sm font-medium text-theme-text mb-2">Public key</label>
                  <textarea id="create-publicKey" v-model="createForm.publicKey" class="input-field min-h-[80px] font-mono text-xs" placeholder="Paste runner public key (e.g. SSH or PEM)" rows="4" />
                </div>
              </div>
            </div>
            <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
              <button type="button" @click="closeCreateModal" class="btn-secondary">Cancel</button>
              <button type="submit" class="btn-primary" :disabled="createLoading">
                <span v-if="!createLoading">Create runner</span>
                <span v-else class="inline-flex items-center gap-2">
                  <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Creating…
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
                <h3 class="text-lg font-semibold text-theme-text">Edit Runner</h3>
                <button type="button" @click="closeEditModal" class="text-theme-textLight hover:text-theme-text">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div class="space-y-4">
                <div>
                  <label for="edit-name" class="block text-sm font-medium text-theme-text mb-2">Name</label>
                  <input id="edit-name" v-model="editForm.name" type="text" class="input-field" placeholder="e.g. prod-runner-1" />
                </div>
                <div>
                  <label for="edit-tags" class="block text-sm font-medium text-theme-text mb-2">Tags</label>
                  <input id="edit-tags" v-model="editForm.tags" type="text" class="input-field" placeholder="e.g. linux, docker" />
                </div>
                <div>
                  <label for="edit-maxParallel" class="block text-sm font-medium text-theme-text mb-2">Max parallel tasks</label>
                  <input id="edit-maxParallel" v-model.number="editForm.maxParallelTasks" type="number" min="1" class="input-field" />
                </div>
                <div>
                  <label for="edit-status" class="block text-sm font-medium text-theme-text mb-2">Status</label>
                  <select id="edit-status" v-model="editForm.status" class="input-field">
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                    <option value="maintenance">Maintenance</option>
                  </select>
                </div>
                <div>
                  <label for="edit-url" class="block text-sm font-medium text-theme-text mb-2">URL</label>
                  <input id="edit-url" v-model="editForm.url" type="url" class="input-field" placeholder="https://runner.example.com" />
                </div>
                <div>
                  <label for="edit-publicKey" class="block text-sm font-medium text-theme-text mb-2">Public key</label>
                  <textarea id="edit-publicKey" v-model="editForm.publicKey" class="input-field min-h-[80px] font-mono text-xs" placeholder="Paste runner public key (e.g. SSH or PEM)" rows="4" />
                </div>
              </div>
            </div>
            <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
              <button type="button" @click="closeEditModal" class="btn-secondary">Cancel</button>
              <button type="submit" class="btn-primary" :disabled="updateLoading">
                <span v-if="!updateLoading">Save</span>
                <span v-else class="inline-flex items-center gap-2">
                  <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Saving…
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
              <h3 class="text-lg font-semibold text-theme-text">Delete Runner</h3>
              <button type="button" @click="closeDeleteModal" class="text-theme-textLight hover:text-theme-text">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <p class="text-sm text-theme-textLight">
              Are you sure you want to delete the runner
              <strong class="text-theme-text">{{ runnerToDelete?.name }}</strong>?
              This action cannot be undone.
            </p>
          </div>
          <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
            <button type="button" @click="closeDeleteModal" class="btn-secondary">Cancel</button>
            <button
              @click="handleDelete"
              class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700 focus:outline-none transition-all duration-150"
              :disabled="deleteLoading"
            >
              <span v-if="!deleteLoading">Delete</span>
              <span v-else class="flex items-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Deleting…
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
import { runnersAPI } from '@/api'
import AppNav from '@/components/AppNav.vue'
import { formatRelative } from '@/utils/date'

const loading = ref(false)
const runners = ref([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const successMessage = ref(null)
const errorMessage = ref(null)

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const createLoading = ref(false)
const updateLoading = ref(false)
const deleteLoading = ref(false)
const runnerToDelete = ref(null)
const runnerToEdit = ref(null)
const createdRunnerToken = ref(null)

const createForm = reactive({
  name: '',
  tags: '',
  maxParallelTasks: 10,
  status: 'active',
  url: '',
  publicKey: ''
})

const editForm = reactive({
  name: '',
  tags: '',
  maxParallelTasks: 10,
  status: 'active',
  url: '',
  publicKey: ''
})

const showingFrom = computed(() => (total.value === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + runners.value.length, total.value))

function maskToken(token) {
  if (!token || token.length < 12) return '••••••••'
  return token.slice(0, 4) + '…' + token.slice(-8)
}

async function loadRunners(overrideOffset) {
  loading.value = true
  errorMessage.value = null
  const off = overrideOffset !== undefined ? overrideOffset : offset.value
  try {
    const response = await runnersAPI.list({ limit: limit.value, offset: off })
    runners.value = response.data?.runners ?? []
    const meta = response.data?._meta
    total.value = meta?.total ?? runners.value.length
    offset.value = off
  } catch (err) {
    console.error('Failed to load runners:', err)
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load runners'
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  const maxOffset = Math.max(0, total.value - limit.value)
  const off = Math.max(0, Math.min(newOffset, maxOffset))
  loadRunners(off)
}

async function copyToken(token) {
  try {
    await navigator.clipboard.writeText(token)
    successMessage.value = 'Token copied to clipboard'
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch {
    errorMessage.value = 'Failed to copy token'
    setTimeout(() => { errorMessage.value = null }, 3000)
  }
}

async function copyCreatedToken() {
  if (!createdRunnerToken.value) return
  try {
    await navigator.clipboard.writeText(createdRunnerToken.value)
    successMessage.value = 'Token copied to clipboard'
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch {
    errorMessage.value = 'Failed to copy token'
    setTimeout(() => { errorMessage.value = null }, 3000)
  }
}

function dismissTokenCard() {
  createdRunnerToken.value = null
}

function openCreateModal() {
  createdRunnerToken.value = null
  createForm.name = ''
  createForm.tags = ''
  createForm.maxParallelTasks = 10
  createForm.status = 'active'
  createForm.url = ''
  createForm.publicKey = ''
  showCreateModal.value = true
}

function closeCreateModal() {
  showCreateModal.value = false
}

function openEditModal(runner) {
  runnerToEdit.value = runner
  editForm.name = runner.name || ''
  editForm.tags = runner.tags || ''
  editForm.maxParallelTasks = runner.maxParallelTasks ?? 10
  editForm.status = runner.status || 'active'
  editForm.url = runner.url || ''
  editForm.publicKey = runner.publicKey || ''
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
  createdRunnerToken.value = null
  try {
    const response = await runnersAPI.create({
      name: createForm.name,
      tags: createForm.tags,
      maxParallelTasks: createForm.maxParallelTasks,
      status: createForm.status,
      url: createForm.url,
      publicKey: createForm.publicKey || undefined
    })
    const data = response.data
    if (data?.token) {
      createdRunnerToken.value = data.token
      successMessage.value = 'Runner created. Copy the token to register the runner.'
    }
    closeCreateModal()
    await loadRunners()
    if (successMessage.value) setTimeout(() => { successMessage.value = null }, 5000)
  } catch (err) {
    console.error('Failed to create runner:', err)
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to create runner'
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
      name: editForm.name || null,
      tags: editForm.tags || null,
      maxParallelTasks: editForm.maxParallelTasks,
      status: editForm.status,
      url: editForm.url || null,
      publicKey: editForm.publicKey || null
    })
    successMessage.value = 'Runner updated'
    closeEditModal()
    await loadRunners()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to update runner:', err)
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to update runner'
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
    successMessage.value = 'Runner deleted'
    closeDeleteModal()
    await loadRunners()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to delete runner:', err)
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to delete runner'
  } finally {
    deleteLoading.value = false
  }
}

onMounted(loadRunners)
</script>
