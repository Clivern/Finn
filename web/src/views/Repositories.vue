<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Repositories
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Manage Git repositories connected to this workspace
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" aria-label="Repositories overview">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">Repositories</h2>
            <p class="text-sm text-theme-textLight mt-0.5">Add, edit, or remove Git repositories. Connect repos to enable runs, pull changes, and manage access via SSH keys.</p>
          </div>
          <button
            type="button"
            @click="openAddModal"
            class="btn-primary text-sm inline-flex items-center gap-2"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Add repository
          </button>
        </div>
        <div v-if="loading" class="p-8 text-center text-theme-textLight text-sm">Loading repositories…</div>
        <div v-else-if="repositories.length === 0" class="p-8 text-center text-theme-textLight text-sm">No repositories yet. Add one above.</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Name</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Git URL</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Branch</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Tag</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="repo in repositories" :key="repo.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <p class="text-sm font-medium text-theme-text">{{ repo.name }}</p>
                  </td>
                  <td class="px-6 py-4">
                    <p class="text-sm text-theme-textLight truncate max-w-xs" :title="repo.gitUrl">{{ repo.gitUrl }}</p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex rounded-full bg-primary-100 px-2.5 py-0.5 text-xs font-medium text-theme-text">
                      {{ repo.gitBranch || 'main' }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span v-if="repo.gitTag" class="inline-flex rounded-full bg-theme-hover px-2.5 py-0.5 text-xs font-medium text-theme-text">
                      {{ repo.gitTag }}
                    </span>
                    <span v-else class="text-theme-textLight text-xs">—</span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      type="button"
                      @click="openEditModal(repo)"
                      class="text-primary-600 hover:text-primary-700 hover:underline mr-3"
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      @click="openDeleteModal(repo)"
                      class="text-red-600 hover:text-red-700 hover:underline"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="_meta?.total" class="px-6 py-4 border-t border-theme-border text-sm text-theme-textLight">
            {{ _meta.total }} {{ _meta.total === 1 ? 'repository' : 'repositories' }}
          </div>
        </template>
      </section>
    </main>

    <!-- Add / Edit repository modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="formModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="editingRepo ? 'edit-repo-title' : 'add-repo-title'"
          @click.self="formModalOpen = false"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 :id="editingRepo ? 'edit-repo-title' : 'add-repo-title'" class="text-lg font-semibold text-theme-text">
              {{ editingRepo ? 'Edit repository' : 'Add repository' }}
            </h2>
            <form @submit.prevent="saveRepo" class="mt-4 space-y-4">
              <div>
                <label for="repo-name" class="form-label">Name</label>
                <input
                  id="repo-name"
                  v-model="form.name"
                  type="text"
                  required
                  maxlength="100"
                  class="input-field"
                  placeholder="e.g. my-app"
                />
              </div>
              <div>
                <label for="repo-gitUrl" class="form-label">Git URL</label>
                <input
                  id="repo-gitUrl"
                  v-model="form.gitUrl"
                  type="text"
                  required
                  class="input-field"
                  placeholder="https://github.com/org/repo.git or git@github.com:org/repo.git"
                />
              </div>
              <div>
                <label for="repo-gitBranch" class="form-label">Branch</label>
                <input
                  id="repo-gitBranch"
                  v-model="form.gitBranch"
                  type="text"
                  maxlength="100"
                  class="input-field"
                  placeholder="main"
                />
              </div>
              <div>
                <label for="repo-gitTag" class="form-label">Tag (optional)</label>
                <input
                  id="repo-gitTag"
                  v-model="form.gitTag"
                  type="text"
                  maxlength="100"
                  class="input-field"
                  placeholder="e.g. v1.0.0"
                />
                <p class="mt-1 text-xs text-theme-textLight">If set, clone will use this tag instead of the branch.</p>
              </div>
              <div>
                <label for="repo-sshKeyId" class="form-label">SSH key (optional)</label>
                <select id="repo-sshKeyId" v-model="form.sshKeyId" class="input-field w-full">
                  <option value="">None</option>
                  <option v-for="k in sshKeysList" :key="k.id" :value="k.id">{{ k.name }}</option>
                </select>
              </div>
              <div>
                <label for="repo-gitCredentialsId" class="form-label">Git credentials (optional)</label>
                <select id="repo-gitCredentialsId" v-model="form.gitCredentialsId" class="input-field w-full">
                  <option value="">None</option>
                  <option v-for="c in credentialsList" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="formModalOpen = false" class="btn-secondary">Cancel</button>
                <button type="submit" :disabled="formSaving" class="btn-primary">
                  {{ formSaving ? 'Saving…' : (editingRepo ? 'Save' : 'Add') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Delete repository modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteModalRepo"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-repo-title"
          @click.self="deleteModalRepo = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-repo-title" class="text-lg font-semibold text-theme-text">Delete repository</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              Delete repository <span class="font-medium text-theme-text">"{{ deleteModalRepo?.name }}"</span>? This cannot be undone.
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" @click="deleteModalRepo = null" class="btn-secondary">Cancel</button>
              <button
                type="button"
                @click="confirmDelete"
                :disabled="deletingId === deleteModalRepo?.id"
                class="btn-danger"
              >
                {{ deletingId === deleteModalRepo?.id ? 'Deleting…' : 'Delete' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppNav from '@/components/AppNav.vue'
import { loadWorkspaceFromStorage } from '@/utils/storage'
import { repositoriesAPI, sshKeysAPI, credentialsAPI } from '@/api'

const router = useRouter()
const workspace = ref(null)
const repositories = ref([])
const _meta = ref(null)
const loading = ref(true)
const errorMessage = ref('')
const successMessage = ref('')

const formModalOpen = ref(false)
const formSaving = ref(false)
const editingRepo = ref(null)
const form = ref({
  name: '',
  gitUrl: '',
  gitBranch: 'main',
  gitTag: '',
  sshKeyId: '',
  gitCredentialsId: ''
})

const sshKeysList = ref([])
const credentialsList = ref([])

const deleteModalRepo = ref(null)
const deletingId = ref(null)

function setWorkspace() {
  workspace.value = loadWorkspaceFromStorage()
  if (!workspace.value?.id) {
    router.replace('/select-workspace')
  }
}

async function loadRepositories() {
  if (!workspace.value?.id) return
  loading.value = true
  errorMessage.value = ''
  try {
    const res = await repositoriesAPI.list(workspace.value.id, { limit: 100, offset: 0 })
    repositories.value = res.data?.repositories ?? []
    _meta.value = res.data?._meta ?? null
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load repositories'
    repositories.value = []
  } finally {
    loading.value = false
  }
}

async function loadSshKeysAndCredentials() {
  if (!workspace.value?.id) return
  try {
    const [keysRes, credsRes] = await Promise.all([
      sshKeysAPI.list(workspace.value.id, { limit: 100 }),
      credentialsAPI.list(workspace.value.id, { limit: 100 })
    ])
    sshKeysList.value = keysRes.data?.keys ?? []
    credentialsList.value = credsRes.data?.credentials ?? []
  } catch {
    sshKeysList.value = []
    credentialsList.value = []
  }
}

function openAddModal() {
  editingRepo.value = null
  form.value = {
    name: '',
    gitUrl: '',
    gitBranch: 'main',
    gitTag: '',
    sshKeyId: '',
    gitCredentialsId: ''
  }
  formModalOpen.value = true
  loadSshKeysAndCredentials()
}

function openEditModal(repo) {
  editingRepo.value = repo
  form.value = {
    name: repo.name,
    gitUrl: repo.gitUrl,
    gitBranch: repo.gitBranch || 'main',
    gitTag: repo.gitTag ?? '',
    sshKeyId: repo.sshKeyId ?? '',
    gitCredentialsId: repo.gitCredentialsId ?? ''
  }
  formModalOpen.value = true
  loadSshKeysAndCredentials()
}

async function saveRepo() {
  if (!workspace.value?.id) return
  const { name, gitUrl, gitBranch, gitTag, sshKeyId, gitCredentialsId } = form.value
  if (!name?.trim() || !gitUrl?.trim()) return
  formSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const payload = {
      name: name.trim(),
      gitUrl: gitUrl.trim(),
      gitBranch: (gitBranch || 'main').trim(),
      gitTag: (gitTag || '').trim() || undefined,
      sshKeyId: sshKeyId || undefined,
      gitCredentialsId: gitCredentialsId || undefined
    }
    if (editingRepo.value) {
      await repositoriesAPI.update(workspace.value.id, editingRepo.value.id, payload)
      successMessage.value = 'Repository updated'
    } else {
      await repositoriesAPI.create(workspace.value.id, payload)
      successMessage.value = 'Repository added'
    }
    formModalOpen.value = false
    await loadRepositories()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || (editingRepo.value ? 'Failed to update repository' : 'Failed to add repository')
  } finally {
    formSaving.value = false
  }
}

function openDeleteModal(repo) {
  deleteModalRepo.value = repo
}

async function confirmDelete() {
  const repo = deleteModalRepo.value
  if (!workspace.value?.id || !repo?.id) return
  deletingId.value = repo.id
  errorMessage.value = ''
  try {
    await repositoriesAPI.delete(workspace.value.id, repo.id)
    successMessage.value = 'Repository deleted'
    deleteModalRepo.value = null
    await loadRepositories()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to delete repository'
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  setWorkspace()
  if (workspace.value?.id) {
    loadRepositories()
  }
})

watch(workspace, (w) => {
  if (w?.id) loadRepositories()
}, { immediate: false })
</script>
