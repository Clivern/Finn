<template>
  <div class="min-h-screen bg-theme-bg flex flex-col items-center justify-center px-4 py-12">
    <div class="w-full max-w-md">
      <!-- Icon -->
      <div class="flex justify-center mb-6">
        <div class="flex h-14 w-14 items-center justify-center rounded-full bg-primary-200">
          <svg class="h-8 w-8 text-theme-text" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
        </div>
      </div>

      <h1 class="text-xl font-semibold text-theme-text text-center">
        Select a Workspace
      </h1>
      <p class="mt-2 text-sm text-theme-textLight text-center">
        Choose a workspace to continue or create a new one
      </p>

      <!-- Workspace list -->
      <div class="mt-6 space-y-3">
        <button
          v-for="workspace in workspaceStore.workspaces"
          :key="workspace.id"
          type="button"
          @click="selectWorkspace(workspace)"
          class="w-full flex items-center justify-between rounded-lg border border-theme-border bg-white px-4 py-3 text-left shadow-sm hover:border-primary-400 hover:bg-theme-hover transition-colors focus:outline-none focus:ring-2 focus:ring-primary-800 focus:ring-offset-1"
        >
          <div class="flex min-w-0 flex-1 items-center gap-3">
            <div class="min-w-0 flex-1">
              <p class="text-sm font-semibold text-theme-text truncate">
                {{ workspace.name }}
              </p>
              <p class="text-xs text-theme-textLight mt-0.5">
                {{ formatDate(workspace.createdAt) }}
              </p>
            </div>
            <span
              v-if="workspace.status === 'active' || isCurrent(workspace)"
              class="inline-flex flex-shrink-0 rounded-full bg-primary-200 px-2.5 py-0.5 text-xs font-medium text-theme-text"
            >
              Active
            </span>
          </div>
          <svg class="ml-3 h-5 w-5 flex-shrink-0 text-theme-textLight" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      <!-- Create New Workspace -->
      <div class="mt-6">
        <button
          type="button"
          @click="showCreateModal = true"
          class="w-full flex items-center justify-center gap-2 rounded-lg border border-theme-border bg-white px-4 py-3 text-sm font-medium text-theme-text hover:bg-theme-hover focus:outline-none focus:ring-2 focus:ring-primary-800 focus:ring-offset-1 transition-colors"
        >
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Create New Workspace
        </button>
      </div>

      <!-- Error message -->
      <p v-if="workspaceStore.error" class="mt-4 text-sm text-red-600 text-center">
        {{ workspaceStore.error }}
      </p>
    </div>

    <!-- Create workspace modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showCreateModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="create-workspace-title"
          @click.self="showCreateModal = false"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="create-workspace-title" class="text-lg font-semibold text-theme-text">
              Create New Workspace
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">
              Enter a name for your new workspace
            </p>
            <form @submit.prevent="handleCreate" class="mt-4 space-y-4">
              <div>
                <label for="workspace-name" class="form-label">Name</label>
                <input
                  id="workspace-name"
                  v-model="newWorkspaceName"
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
                  @click="showCreateModal = false; newWorkspaceName = ''"
                  class="btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  :disabled="createLoading || !newWorkspaceName.trim()"
                  class="btn-primary"
                >
                  {{ createLoading ? 'Creating…' : 'Create' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWorkspaceStore } from '@/stores/workspace'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const showCreateModal = ref(false)
const newWorkspaceName = ref('')
const createLoading = ref(false)

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
  return workspaceStore.currentWorkspace?.id === workspace.id
}

function selectWorkspace(workspace) {
  workspaceStore.setCurrentWorkspace(workspace)
  router.push('/dashboard')
}

async function handleCreate() {
  const name = newWorkspaceName.value?.trim()
  if (!name) return
  createLoading.value = true
  try {
    const result = await workspaceStore.createWorkspace(name)
    if (result?.success) {
      showCreateModal.value = false
      newWorkspaceName.value = ''
      router.push('/dashboard')
    }
  } finally {
    createLoading.value = false
  }
}

onMounted(async () => {
  await workspaceStore.fetchWorkspaces()
})
</script>
