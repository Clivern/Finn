<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">Workspaces</h1>
        <p class="text-theme-textLight mt-2">Manage all platform workspaces</p>
      </div>

      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>
      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>

      <div class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div v-if="loading" class="p-12 text-center">
          <svg class="animate-spin h-8 w-8 mx-auto text-theme-text" fill="none" viewBox="0 0 24 24" aria-hidden="true">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <p class="text-theme-textLight mt-3">Loading workspaces…</p>
        </div>

        <div v-else-if="workspaces.length === 0" class="p-12 text-center">
          <div class="flex justify-center mb-4">
            <div class="flex h-12 w-12 items-center justify-center rounded-full bg-primary-100">
              <svg class="h-6 w-6 text-theme-text" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            </div>
          </div>
          <p class="text-theme-textLight">No workspaces yet</p>
          <p class="text-sm text-theme-textLight mt-1">Workspaces are created by users when they switch workspaces.</p>
        </div>

        <table v-else class="min-w-full divide-y divide-theme-border">
          <thead class="bg-theme-hover">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Name
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Created
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Members
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-theme-border">
            <tr
              v-for="workspace in workspaces"
              :key="workspace.id"
              class="hover:bg-theme-hover"
            >
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center gap-3">
                  <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text">
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                    </svg>
                  </span>
                  <span class="text-sm font-medium text-theme-text">{{ workspace.name }}</span>
                  <span
                    v-if="isCurrent(workspace)"
                    class="text-xs text-theme-textLight"
                  >
                    (current)
                  </span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatDate(workspace.createdAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                {{ workspace.membersCount ?? '—' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                <span class="inline-flex gap-3">
                  <button
                    v-if="canEditWorkspace(workspace)"
                    type="button"
                    @click="openEditModal(workspace)"
                    class="text-primary-600 hover:text-primary-700 hover:underline disabled:opacity-50"
                    :disabled="updateLoadingId === workspace.id"
                    title="Edit workspace name"
                  >
                    {{ updateLoadingId === workspace.id ? 'Saving…' : 'Edit' }}
                  </button>
                  <button
                    type="button"
                    @click="openDeleteModal(workspace)"
                    class="text-red-600 hover:text-red-700 hover:underline disabled:opacity-50"
                    :disabled="deletingId === workspace.id"
                    title="Delete workspace"
                  >
                    {{ deletingId === workspace.id ? 'Deleting…' : 'Delete' }}
                  </button>
                </span>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination footer -->
        <div
          v-if="workspaces.length > 0"
          class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between"
        >
          <div class="text-sm text-theme-textLight">
            Showing {{ (page - 1) * limit + 1 }} to {{ Math.min(page * limit, meta.total) }} of {{ meta.total }} workspaces
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="goToPage(page - 1)"
              :disabled="page <= 1 || loading"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              type="button"
              @click="goToPage(page + 1)"
              :disabled="page >= totalPages || loading"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Edit workspace modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="editModalWorkspace"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="edit-workspace-title"
          @click.self="editModalWorkspace = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="edit-workspace-title" class="text-lg font-semibold text-theme-text">
              Edit Workspace Name
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">
              Change the name of this workspace
            </p>
            <form @submit.prevent="handleUpdate" class="mt-4 space-y-4">
              <div>
                <label for="edit-workspace-name" class="form-label">Name</label>
                <input
                  id="edit-workspace-name"
                  v-model="editWorkspaceName"
                  type="text"
                  required
                  class="input-field"
                  placeholder="My Workspace"
                  autofocus
                />
              </div>
              <div class="flex gap-3 justify-end">
                <button
                  type="button"
                  @click="editModalWorkspace = null"
                  class="btn-secondary"
                  :disabled="updateLoadingId"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  :disabled="updateLoadingId || !editWorkspaceName.trim() || editWorkspaceName.trim() === editModalWorkspace?.name"
                  class="btn-primary"
                >
                  {{ updateLoadingId ? 'Saving…' : 'Save' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteModalWorkspace"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-workspace-title"
          @click.self="closeDeleteModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="delete-workspace-title" class="text-lg font-semibold text-theme-text">
              Delete Workspace
            </h2>
            <p class="mt-2 text-sm text-theme-textLight">
              Delete workspace <span class="font-medium text-theme-text">"{{ deleteModalWorkspace?.name }}"</span>? This cannot be undone.
            </p>
            <div class="mt-4">
              <label for="delete-confirm-name" class="block text-sm font-medium text-theme-text mb-1.5">
                Type <span class="font-mono">{{ deleteModalWorkspace?.name }}</span> to confirm
              </label>
              <input
                id="delete-confirm-name"
                v-model="deleteConfirmName"
                type="text"
                class="input-field"
                :placeholder="deleteModalWorkspace?.name"
                autocomplete="off"
                :disabled="deleting"
              />
            </div>
            <div class="mt-6 flex gap-3 justify-end">
              <button
                type="button"
                @click="closeDeleteModal"
                class="btn-secondary"
                :disabled="deleting"
              >
                Cancel
              </button>
              <button
                type="button"
                @click="confirmDelete"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="deleting || !canConfirmDelete"
              >
                {{ deleting ? 'Deleting…' : 'Delete' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import AppNav from '@/components/AppNav.vue'
import { workspaceAPI } from '@/api'
import { loadWorkspaceFromStorage, saveWorkspaceToStorage, removeWorkspaceFromStorage } from '@/utils/storage'
const workspaces = ref([])
const loading = ref(false)
const meta = ref({ limit: 50, offset: 0, total: 0 })
const successMessage = ref(null)
const errorMessage = ref(null)
const page = ref(1)
const limit = ref(10)

const editModalWorkspace = ref(null)
const editWorkspaceName = ref('')
const updateLoadingId = ref(null)
const deleteModalWorkspace = ref(null)
const deleteConfirmName = ref('')
const deleting = ref(false)
const deletingId = ref(null)

function canEditWorkspace(workspace) {
  const role = workspace?.role
  return role === 'admin' || role === 'owner'
}

function openEditModal(workspace) {
  editModalWorkspace.value = workspace
  editWorkspaceName.value = workspace?.name ?? ''
}

async function handleUpdate() {
  const workspace = editModalWorkspace.value
  const name = editWorkspaceName.value?.trim()
  if (!workspace || !name) return
  updateLoadingId.value = workspace.id
  errorMessage.value = null
  try {
    const response = await workspaceAPI.update(workspace.id, { name })
    const updated = response.data
    if (updated) {
      const i = workspaces.value.findIndex((w) => w.id === workspace.id)
      if (i !== -1) workspaces.value[i] = { ...workspaces.value[i], ...updated }
      if (loadWorkspaceFromStorage()?.id === workspace.id) {
        saveWorkspaceToStorage({ ...loadWorkspaceFromStorage(), ...updated })
      }
      editModalWorkspace.value = null
      editWorkspaceName.value = ''
    } else {
      errorMessage.value = 'Failed to update workspace'
    }
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || err.message || 'Failed to update workspace'
  } finally {
    updateLoadingId.value = null
  }
}

const totalPages = computed(() =>
  Math.ceil((meta.value.total || 0) / limit.value) || 1
)

const canConfirmDelete = computed(() => {
  const name = deleteModalWorkspace.value?.name
  return name && deleteConfirmName.value.trim() === name
})

function formatDate(iso) {
  if (!iso) return ''
  try {
    const d = new Date(iso)
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  } catch {
    return iso
  }
}

function isCurrent(workspace) {
  return loadWorkspaceFromStorage()?.id === workspace.id
}

async function loadPage() {
  loading.value = true
  errorMessage.value = null
  try {
    const params = { all: 'true', limit: limit.value, offset: (page.value - 1) * limit.value }
    const response = await workspaceAPI.list(params)
    workspaces.value = response.data?.workspaces ?? []
    const m = response.data?._meta
    if (m) meta.value = { limit: m.limit ?? 50, offset: m.offset ?? 0, total: m.total ?? 0 }
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || err.message || 'Failed to load workspaces'
    workspaces.value = []
  } finally {
    loading.value = false
  }
}

function goToPage(p) {
  page.value = p
  loadPage()
}

function openDeleteModal(workspace) {
  deleteModalWorkspace.value = workspace
  deleteConfirmName.value = ''
}

function closeDeleteModal() {
  if (!deleting.value) {
    deleteModalWorkspace.value = null
    deleteConfirmName.value = ''
  }
}

async function confirmDelete() {
  const workspace = deleteModalWorkspace.value
  if (!workspace) return
  deleting.value = true
  deletingId.value = workspace.id
  errorMessage.value = null
  successMessage.value = null
  try {
    await workspaceAPI.delete(workspace.id)
    if (loadWorkspaceFromStorage()?.id === workspace.id) {
      removeWorkspaceFromStorage()
    }
    deleteModalWorkspace.value = null
    deleteConfirmName.value = ''
    workspaces.value = workspaces.value.filter((w) => w.id !== workspace.id)
    if (workspaces.value.length === 0 && page.value > 1) {
      page.value = page.value - 1
      await loadPage()
    } else {
      meta.value = { ...meta.value, total: Math.max(0, (meta.value.total ?? 0) - 1) }
    }
    successMessage.value = 'Workspace deleted successfully.'
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || err.message || 'Failed to delete workspace'
  } finally {
    deleting.value = false
    deletingId.value = null
  }
}

onMounted(async () => {
  successMessage.value = null
  errorMessage.value = null
  page.value = 1
  await loadPage()
})
</script>
