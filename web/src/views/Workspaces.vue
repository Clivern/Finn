<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('workspaces.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('workspacesPage.subtitle') }}</p>
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
          <p class="text-theme-textLight mt-3">{{ $t('workspacesPage.loading') }}</p>
        </div>

        <div v-else-if="workspaces.length === 0" class="p-12 text-center">
          <div class="flex justify-center mb-4">
            <div class="flex h-12 w-12 items-center justify-center rounded-full bg-primary-100">
              <svg class="h-6 w-6 text-theme-text" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            </div>
          </div>
          <p class="text-theme-textLight">{{ $t('workspacesPage.noWorkspaces') }}</p>
          <p class="text-sm text-theme-textLight mt-1">{{ $t('workspacesPage.noWorkspacesHint') }}</p>
        </div>

        <table v-else class="min-w-full divide-y divide-theme-border">
          <thead class="bg-theme-hover">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('workspacesPage.name') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('workspacesPage.created') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('workspacesPage.members') }}
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('workspacesPage.actions') }}
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
                    {{ $t('workspacesPage.current') }}
                  </span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatDate(workspace.createdAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                {{ workspace.membersCount }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                <span class="inline-flex items-center gap-3">
                  <button
                    type="button"
                    @click="switchToWorkspace(workspace)"
                    class="text-emerald-600 hover:text-emerald-700 hover:underline disabled:opacity-50 text-sm"
                    :disabled="switchLoadingId === workspace.id"
                    title="Switch into this workspace"
                  >
                    {{ switchLoadingId === workspace.id ? $t('workspacesPage.switching') : $t('workspacesPage.switchInto') }}
                  </button>
                  <button
                    type="button"
                    @click="openEditModal(workspace)"
                    class="text-primary-600 hover:text-primary-700 hover:underline disabled:opacity-50"
                    :disabled="updateLoadingId === workspace.id"
                    :title="$t('workspacesPage.editTitle')"
                  >
                    {{ updateLoadingId === workspace.id ? $t('workspacesPage.saving') : $t('workspacesPage.edit') }}
                  </button>
                  <button
                    type="button"
                    @click="openDeleteModal(workspace)"
                    class="text-red-600 hover:text-red-700 hover:underline disabled:opacity-50"
                    :disabled="deletingId === workspace.id"
                    :title="$t('workspacesPage.deleteTitle')"
                  >
                    {{ deletingId === workspace.id ? $t('workspacesPage.deleting') : $t('workspacesPage.delete') }}
                  </button>
                </span>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination footer -->
        <div
          v-if="total > 0"
          class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between"
        >
          <div class="text-sm text-theme-textLight">
            {{ $t('workspacesPage.showing', { from: offset + 1, to: Math.min(offset + limit, total), total }) }}
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="goToPage(offset - limit)"
              :disabled="offset === 0"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('workspacesPage.previous') }}
            </button>
            <button
              type="button"
              @click="goToPage(offset + limit)"
              :disabled="offset + limit >= total"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('workspacesPage.next') }}
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
          @click.self="closeEditModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="edit-workspace-title" class="text-lg font-semibold text-theme-text">
              {{ $t('workspacesPage.editTitle') }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">
              {{ $t('workspacesPage.editDesc') }}
            </p>
            <div v-if="editModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ editModalError }}
            </div>
            <form @submit.prevent="handleUpdate" class="mt-4 space-y-4">
              <div>
                <label for="edit-workspace-name" class="form-label">{{ $t('workspacesPage.name') }}</label>
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
                  @click="closeEditModal"
                  class="btn-secondary"
                  :disabled="updateLoadingId"
                >
                  {{ $t('common.cancel') }}
                </button>
                <button
                  type="submit"
                  :disabled="updateLoadingId || !editWorkspaceName.trim() || editWorkspaceName.trim() === editModalWorkspace?.name"
                  class="btn-primary"
                >
                  {{ updateLoadingId ? $t('workspacesPage.saving') : $t('common.save') }}
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
              {{ $t('workspacesPage.deleteTitle') }}
            </h2>
            <div v-if="deleteModalError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
              {{ deleteModalError }}
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('workspacesPage.deleteConfirm', { name: deleteModalWorkspace?.name }) }}
            </p>
            <div class="mt-4">
              <label for="delete-confirm-name" class="block text-sm font-medium text-theme-text mb-1.5">
                {{ $t('workspacesPage.typeToConfirm', { name: deleteModalWorkspace?.name }) }}
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
                {{ $t('common.cancel') }}
              </button>
              <button
                type="button"
                @click="confirmDelete"
                class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
                :disabled="deleting || !canConfirmDelete"
              >
                {{ deleting ? $t('workspacesPage.deleting') : $t('workspacesPage.delete') }}
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
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { workspaceAPI } from '@/api'
import { loadWorkspaceFromStorage, saveWorkspaceToStorage, removeWorkspaceFromStorage } from '@/utils/storage'

const { t } = useI18n()
const router = useRouter()
const workspaces = ref([])
const loading = ref(false)
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const successMessage = ref(null)
const errorMessage = ref(null)

const editModalWorkspace = ref(null)
const editWorkspaceName = ref('')
const updateLoadingId = ref(null)
const deleteModalWorkspace = ref(null)
const deleteConfirmName = ref('')
const deleting = ref(false)
const deletingId = ref(null)
const switchLoadingId = ref(null)

const editModalError = ref(null)
const deleteModalError = ref(null)

async function switchToWorkspace(workspace) {
  if (!workspace?.id) return
  switchLoadingId.value = workspace.id
  errorMessage.value = null

  try {
    const res = await workspaceAPI.get(workspace.id)
    saveWorkspaceToStorage({ ...workspace, ...res.data, role: res.data.role })
    router.push('/dashboard')
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('workspacesPage.failedSwitch')
  } finally {
    switchLoadingId.value = null
  }
}

function openEditModal(workspace) {
  editModalWorkspace.value = workspace
  editWorkspaceName.value = workspace.name
  editModalError.value = null
}

function closeEditModal() {
  if (updateLoadingId.value) return
  editModalWorkspace.value = null
  editWorkspaceName.value = ''
  editModalError.value = null
}

async function handleUpdate() {
  const workspace = editModalWorkspace.value
  const name = editWorkspaceName.value.trim()

  if (!workspace || !name) return
  updateLoadingId.value = workspace.id
  errorMessage.value = null
  editModalError.value = null

  try {
    const res = await workspaceAPI.update(workspace.id, { name })
    if (loadWorkspaceFromStorage()?.id === workspace.id) {
      saveWorkspaceToStorage({ ...loadWorkspaceFromStorage(), ...res.data })
    }
    editModalWorkspace.value = null
    editWorkspaceName.value = ''
    editModalError.value = null
    await loadPage()
    successMessage.value = t('workspacesPage.workspaceUpdated')
  } catch (err) {
    editModalError.value = err.response?.data?.errorMessage || t('workspacesPage.failedUpdate')
  } finally {
    updateLoadingId.value = null
  }
}

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
    const res = await workspaceAPI.list({
      all: 'true',
      limit: limit.value,
      offset: offset.value
    })
    workspaces.value = res.data.workspaces
    total.value = res.data._meta.total
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('workspacesPage.failedLoad')
    workspaces.value = []
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0) return
  if (newOffset >= total.value) return
  offset.value = newOffset
  loadPage()
}

function openDeleteModal(workspace) {
  deleteModalWorkspace.value = workspace
  deleteConfirmName.value = ''
  deleteModalError.value = null
}

function closeDeleteModal() {
  if (!deleting.value) {
    deleteModalWorkspace.value = null
    deleteConfirmName.value = ''
    deleteModalError.value = null
  }
}

async function confirmDelete() {
  const workspace = deleteModalWorkspace.value
  if (!workspace) return
  deleting.value = true
  deletingId.value = workspace.id
  errorMessage.value = null
  successMessage.value = null
  deleteModalError.value = null

  try {
    await workspaceAPI.delete(workspace.id)
    if (loadWorkspaceFromStorage()?.id === workspace.id) {
      removeWorkspaceFromStorage()
    }
    deleteModalWorkspace.value = null
    deleteConfirmName.value = ''
    deleteModalError.value = null
    await loadPage()
    successMessage.value = t('workspacesPage.workspaceDeleted')
  } catch (err) {
    deleteModalError.value = err.response?.data?.errorMessage || t('workspacesPage.failedDelete')
  } finally {
    deleting.value = false
    deletingId.value = null
  }
}

onMounted(() => {
  successMessage.value = null
  errorMessage.value = null
  offset.value = 0
  loadPage()
})
</script>
