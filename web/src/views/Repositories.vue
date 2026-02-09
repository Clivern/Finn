<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('repositoriesPage.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('repositoriesPage.subtitle') }}
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" :aria-label="$t('repositoriesPage.sectionTitle')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('repositoriesPage.sectionTitle') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('repositoriesPage.sectionDescription') }}</p>
          </div>
          <button
            type="button"
            @click="openAddModal"
            class="btn-primary text-sm inline-flex items-center gap-2"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('repositoriesPage.addRepository') }}
          </button>
        </div>
        <div v-if="loading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('repositoriesPage.loading') }}</div>
        <div v-else-if="repositories.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('repositoriesPage.empty') }}</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('repositoriesPage.name') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('repositoriesPage.gitUrl') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('repositoriesPage.branch') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('repositoriesPage.tag') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('repositoriesPage.status') }}</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('repositoriesPage.actions') }}</th>
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
                      {{ repo.gitBranch || '-' }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span v-if="repo.gitTag" class="inline-flex rounded-full bg-theme-hover px-2.5 py-0.5 text-xs font-medium text-theme-text">
                      {{ repo.gitTag || '-' }}
                    </span>
                    <span v-else class="text-theme-textLight text-xs">—</span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span
                      :class="[
                        'inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium',
                        repo.latestStatus === 'success' && 'bg-green-100 text-green-800',
                        repo.latestStatus === 'failed' && 'bg-red-100 text-red-800',
                        (repo.latestStatus === 'running' || repo.latestStatus === 'pending') && 'bg-amber-100 text-amber-800',
                        (repo.latestStatus === 'unknown') && 'bg-theme-hover text-theme-textLight'
                      ]"
                    >
                      {{ titleCase(repo.latestStatus) }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      type="button"
                      @click="triggerClone(repo)"
                      :disabled="cloningId === repo.id"
                      class="text-primary-600 hover:text-primary-700 hover:underline mr-3 disabled:opacity-50"
                    >
                      {{ cloningId === repo.id ? $t('repositoriesPage.saving') : $t('repositoriesPage.testClone') }}
                    </button>
                    <button
                      type="button"
                      @click="openEditModal(repo)"
                      class="text-primary-600 hover:text-primary-700 hover:underline mr-3"
                    >
                      {{ $t('common.edit') }}
                    </button>
                    <button
                      type="button"
                      @click="openDeleteModal(repo)"
                      class="text-red-600 hover:text-red-700 hover:underline"
                    >
                      {{ $t('common.delete') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="px-6 py-4 border-t border-theme-border flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm text-theme-textLight">
              <template v-if="hasPagination">
                {{ $t('repositoriesPage.showing', { from: showingFrom, to: showingTo, total: totalRepos }) }}
              </template>
              <template v-else-if="_meta?.total">
                {{ _meta.total === 1 ? $t('repositoriesPage.countOne') : $t('repositoriesPage.countMany', { count: _meta.total }) }}
              </template>
            </div>
            <div v-if="hasPagination" class="flex items-center gap-2">
              <button
                type="button"
                @click="goToPage(offset - limit)"
                :disabled="offset === 0"
                class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ $t('repositoriesPage.previous') }}
              </button>
              <button
                type="button"
                @click="goToPage(offset + limit)"
                :disabled="offset + limit >= totalRepos"
                class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ $t('repositoriesPage.next') }}
              </button>
            </div>
          </div>
        </template>
      </section>
    </main>

    <!-- Add repository modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showAddModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="add-repo-title"
          @click.self="closeAddModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="add-repo-title" class="text-lg font-semibold text-theme-text">{{ $t('repositoriesPage.addRepo') }}</h2>
            <div v-if="addModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ addModalError }}</p>
            </div>
            <form @submit.prevent="handleCreate" class="mt-4 space-y-4">
              <div>
                <label for="add-repo-name" class="form-label">{{ $t('repositoriesPage.name') }}</label>
                <input
                  id="add-repo-name"
                  v-model="addForm.name"
                  type="text"
                  required
                  maxlength="100"
                  class="input-field"
                  placeholder="e.g. my-app"
                />
              </div>
              <div>
                <label for="add-repo-gitUrl" class="form-label">{{ $t('repositoriesPage.gitUrl') }}</label>
                <input
                  id="add-repo-gitUrl"
                  v-model="addForm.gitUrl"
                  type="text"
                  required
                  class="input-field"
                  placeholder="https://github.com/org/repo.git or git@github.com:org/repo.git"
                />
              </div>
              <div>
                <label for="add-repo-gitBranch" class="form-label">{{ $t('repositoriesPage.branch') }}</label>
                <input
                  id="add-repo-gitBranch"
                  v-model="addForm.gitBranch"
                  type="text"
                  maxlength="100"
                  class="input-field"
                  placeholder="main"
                />
              </div>
              <div>
                <label for="add-repo-gitTag" class="form-label">{{ $t('repositoriesPage.tagOptional') }}</label>
                <input
                  id="add-repo-gitTag"
                  v-model="addForm.gitTag"
                  type="text"
                  maxlength="100"
                  class="input-field"
                  placeholder="e.g. v1.0.0"
                />
                <p class="mt-1 text-xs text-theme-textLight">{{ $t('repositoriesPage.tagHint') }}</p>
              </div>
              <div>
                <label for="add-repo-sshKeyId" class="form-label">{{ $t('repositoriesPage.sshKeyOptional') }}</label>
                <select id="add-repo-sshKeyId" v-model="addForm.sshKeyId" class="input-field w-full">
                  <option value="">{{ $t('repositoriesPage.none') }}</option>
                  <option v-for="k in sshKeysList" :key="k.id" :value="k.id">{{ k.name }}</option>
                </select>
              </div>
              <div>
                <label for="add-repo-gitCredentialsId" class="form-label">{{ $t('repositoriesPage.gitCredentialsOptional') }}</label>
                <select id="add-repo-gitCredentialsId" v-model="addForm.gitCredentialsId" class="input-field w-full">
                  <option value="">{{ $t('repositoriesPage.none') }}</option>
                  <option v-for="c in credentialsList" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="closeAddModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="addSaving" class="btn-primary">
                  {{ addSaving ? $t('repositoriesPage.saving') : $t('repositoriesPage.add') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Edit repository modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showEditModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="edit-repo-title"
          @click.self="closeEditModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="edit-repo-title" class="text-lg font-semibold text-theme-text">{{ $t('repositoriesPage.editRepo') }}</h2>
            <div v-if="editModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ editModalError }}</p>
            </div>
            <form @submit.prevent="handleUpdate" class="mt-4 space-y-4">
              <div>
                <label for="edit-repo-name" class="form-label">{{ $t('repositoriesPage.name') }}</label>
                <input
                  id="edit-repo-name"
                  v-model="editForm.name"
                  type="text"
                  required
                  maxlength="100"
                  class="input-field"
                  placeholder="e.g. my-app"
                />
              </div>
              <div>
                <label for="edit-repo-gitUrl" class="form-label">{{ $t('repositoriesPage.gitUrl') }}</label>
                <input
                  id="edit-repo-gitUrl"
                  v-model="editForm.gitUrl"
                  type="text"
                  required
                  class="input-field"
                  placeholder="https://github.com/org/repo.git or git@github.com:org/repo.git"
                />
              </div>
              <div>
                <label for="edit-repo-gitBranch" class="form-label">{{ $t('repositoriesPage.branch') }}</label>
                <input
                  id="edit-repo-gitBranch"
                  v-model="editForm.gitBranch"
                  type="text"
                  maxlength="100"
                  class="input-field"
                  placeholder="main"
                />
              </div>
              <div>
                <label for="edit-repo-gitTag" class="form-label">{{ $t('repositoriesPage.tagOptional') }}</label>
                <input
                  id="edit-repo-gitTag"
                  v-model="editForm.gitTag"
                  type="text"
                  maxlength="100"
                  class="input-field"
                  placeholder="e.g. v1.0.0"
                />
                <p class="mt-1 text-xs text-theme-textLight">{{ $t('repositoriesPage.tagHint') }}</p>
              </div>
              <div>
                <label for="edit-repo-sshKeyId" class="form-label">{{ $t('repositoriesPage.sshKeyOptional') }}</label>
                <select id="edit-repo-sshKeyId" v-model="editForm.sshKeyId" class="input-field w-full">
                  <option value="">{{ $t('repositoriesPage.none') }}</option>
                  <option v-for="k in sshKeysList" :key="k.id" :value="k.id">{{ k.name }}</option>
                </select>
              </div>
              <div>
                <label for="edit-repo-gitCredentialsId" class="form-label">{{ $t('repositoriesPage.gitCredentialsOptional') }}</label>
                <select id="edit-repo-gitCredentialsId" v-model="editForm.gitCredentialsId" class="input-field w-full">
                  <option value="">{{ $t('repositoriesPage.none') }}</option>
                  <option v-for="c in credentialsList" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="closeEditModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="editSaving" class="btn-primary">
                  {{ editSaving ? $t('repositoriesPage.saving') : $t('common.save') }}
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
          @click.self="closeDeleteModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-repo-title" class="text-lg font-semibold text-theme-text">{{ $t('repositoriesPage.deleteRepoTitle') }}</h2>
            <div v-if="deleteModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ deleteModalError }}</p>
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('repositoriesPage.deleteRepoConfirm', { name: deleteModalRepo?.name }) }}
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" @click="closeDeleteModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
              <button
                type="button"
                @click="confirmDelete"
                :disabled="deletingId === deleteModalRepo?.id"
                class="btn-danger"
              >
                {{ deletingId === deleteModalRepo?.id ? $t('repositoriesPage.deleting') : $t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Clone task logs modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="cloneTaskModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="'clone-task-logs-title'"
          @click.self="closeCloneTaskModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-2xl max-h-[90vh] flex flex-col">
            <div class="px-6 py-4 border-b border-theme-border flex items-center justify-between">
              <h2 id="clone-task-logs-title" class="text-lg font-semibold text-theme-text">
                {{ $t('repositoriesPage.cloneTaskLogsTitle') }}
              </h2>
              <button type="button" @click="closeCloneTaskModal" class="text-theme-textLight hover:text-theme-text p-1 rounded">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div class="px-6 py-4 overflow-y-auto flex-1 space-y-4">
              <div v-if="cloneTaskLoadError" class="text-sm text-red-600">{{ cloneTaskLoadError }}</div>
              <template v-else-if="cloneTaskData">
                <div
                :class="[
                  'flex items-center gap-2 rounded-lg px-4 py-2.5 text-sm font-medium',
                  cloneTaskData.status === 'success' && 'bg-green-50 text-green-700 border border-green-200',
                  cloneTaskData.status === 'failed' && 'bg-red-50 text-red-700 border border-red-200',
                  (cloneTaskData.status === 'running' || cloneTaskData.status === 'pending') && 'bg-amber-50 text-amber-700 border border-amber-200'
                ]"
              >
                <svg v-if="cloneTaskData.status === 'success'" class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                <svg v-else-if="cloneTaskData.status === 'failed'" class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
                <svg v-else class="h-4 w-4 shrink-0 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
                </svg>
                <span>{{ $t('repositoriesPage.cloneTaskStatus') }}: {{ titleCase(cloneTaskData.status) || '—' }}</span>
              </div>
                <div>
                  <div
                    v-if="cloneTaskLogLines.length === 0"
                    class="rounded-lg border border-theme-border bg-theme-hover px-4 py-3 text-sm text-theme-textLight"
                  >
                    {{ $t('repositoriesPage.noLogsYet') }}
                  </div>
                  <pre
                    v-else
                    class="rounded-lg border border-theme-border bg-theme-hover p-4 text-xs font-mono text-theme-text overflow-x-auto max-h-80 overflow-y-auto"
                  >{{ cloneTaskLogText }}</pre>
                </div>
              </template>
              <div v-else class="text-sm text-theme-textLight">{{ $t('repositoriesPage.loading') }}</div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { loadWorkspaceFromStorage } from '@/utils/storage'
import { repositoriesAPI, sshKeysAPI, credentialsAPI } from '@/api'

const { t } = useI18n()

const router = useRouter()
const workspace = ref(null)
const repositories = ref([])
const _meta = ref(null)
const loading = ref(true)
const limit = ref(20)
const offset = ref(0)
const errorMessage = ref('')
const successMessage = ref('')

const showAddModal = ref(false)
const showEditModal = ref(false)
const addSaving = ref(false)
const editSaving = ref(false)
const editingRepo = ref(null)
const addForm = reactive({
  name: '',
  gitUrl: '',
  gitBranch: '',
  gitTag: '',
  sshKeyId: '',
  gitCredentialsId: ''
})
const editForm = reactive({
  name: '',
  gitUrl: '',
  gitBranch: '',
  gitTag: '',
  sshKeyId: '',
  gitCredentialsId: ''
})

const sshKeysList = ref([])
const credentialsList = ref([])

const deleteModalRepo = ref(null)
const deletingId = ref(null)
const addModalError = ref(null)
const editModalError = ref(null)
const deleteModalError = ref(null)
const cloningId = ref(null)

const cloneTaskModal = ref(null)
const cloneTaskData = ref(null)
const cloneTaskLoadError = ref('')
let cloneTaskPollTimer = null

function setWorkspace() {
  workspace.value = loadWorkspaceFromStorage()
  if (!workspace.value?.id) {
    router.replace('/select-workspace')
  }
}

async function loadRepositories(overrideOffset) {
  loading.value = true
  errorMessage.value = ''
  const off = overrideOffset !== undefined ? overrideOffset : offset.value

  try {
    const res = await repositoriesAPI.list(workspace.value.id, { limit: limit.value, offset: off })
    repositories.value = res.data.repositories ?? []
    _meta.value = res.data._meta ?? null
    offset.value = off
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('repositoriesPage.failedLoad')
    repositories.value = []
  } finally {
    loading.value = false
  }
}

const totalRepos = computed(() => _meta.value?.total ?? 0)
const showingFrom = computed(() => (totalRepos.value === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + repositories.value.length, totalRepos.value))
const hasPagination = computed(() => totalRepos.value > limit.value)

function goToPage(newOffset) {
  const maxOffset = Math.max(0, totalRepos.value - limit.value)
  const off = Math.max(0, Math.min(newOffset, maxOffset))
  loadRepositories(off)
}

async function loadSshKeysAndCredentials() {
  try {
    const [keysRes, credsRes] = await Promise.all([
      sshKeysAPI.list(workspace.value.id, { limit: 100 }),
      credentialsAPI.list(workspace.value.id, { limit: 100 })
    ])
    sshKeysList.value = keysRes.data.keys ?? []
    credentialsList.value = credsRes.data.credentials ?? []
  } catch {
    sshKeysList.value = []
    credentialsList.value = []
  }
}

function openAddModal() {
  addForm.name = ''
  addForm.gitUrl = ''
  addForm.gitBranch = 'main'
  addForm.gitTag = ''
  addForm.sshKeyId = ''
  addForm.gitCredentialsId = ''
  addModalError.value = null
  showAddModal.value = true
  loadSshKeysAndCredentials()
}

function closeAddModal() {
  showAddModal.value = false
  addModalError.value = null
}

function openEditModal(repo) {
  editingRepo.value = repo
  editForm.name = repo.name
  editForm.gitUrl = repo.gitUrl
  editForm.gitBranch = repo.gitBranch
  editForm.gitTag = repo.gitTag
  editForm.sshKeyId = repo.sshKeyId ?? ''
  editForm.gitCredentialsId = repo.gitCredentialsId ?? ''
  editModalError.value = null
  showEditModal.value = true
  loadSshKeysAndCredentials()
}

function closeEditModal() {
  showEditModal.value = false
  editModalError.value = null
  editingRepo.value = null
}

async function handleCreate() {
  if (!addForm.name?.trim() || !addForm.gitUrl?.trim()) return

  addSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  addModalError.value = null

  try {
    await repositoriesAPI.create(workspace.value.id, {
      name: addForm.name.trim(),
      gitUrl: addForm.gitUrl.trim(),
      gitBranch: addForm.gitBranch.trim(),
      gitTag: addForm.gitTag.trim(),
      sshKeyId: addForm.sshKeyId,
      gitCredentialsId: addForm.gitCredentialsId
    })
    successMessage.value = t('repositoriesPage.repoAdded')
    closeAddModal()
    await loadRepositories()
  } catch (err) {
    addModalError.value = err.response?.data?.errorMessage || t('repositoriesPage.failedAdd')
  } finally {
    addSaving.value = false
  }
}

async function handleUpdate() {
  if (!editForm.name?.trim() || !editForm.gitUrl?.trim()) return

  editSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  editModalError.value = null

  try {
    await repositoriesAPI.update(workspace.value.id, editingRepo.value.id, {
      name: editForm.name.trim(),
      gitUrl: editForm.gitUrl.trim(),
      gitBranch: editForm.gitBranch.trim(),
      gitTag: editForm.gitTag.trim(),
      sshKeyId: editForm.sshKeyId,
      gitCredentialsId: editForm.gitCredentialsId
    })
    successMessage.value = t('repositoriesPage.repoUpdated')
    closeEditModal()
    await loadRepositories()
  } catch (err) {
    editModalError.value = err.response?.data?.errorMessage || t('repositoriesPage.failedUpdate')
  } finally {
    editSaving.value = false
  }
}

function openDeleteModal(repo) {
  deleteModalError.value = null
  deleteModalRepo.value = repo
}

function closeDeleteModal() {
  deleteModalRepo.value = null
  deleteModalError.value = null
}

async function confirmDelete() {
  const repo = deleteModalRepo.value
  if (!repo) return
  deletingId.value = repo.id
  errorMessage.value = ''
  deleteModalError.value = null
  try {
    await repositoriesAPI.delete(workspace.value.id, repo.id)
    successMessage.value = t('repositoriesPage.repoDeleted')
    closeDeleteModal()
    await loadRepositories()
  } catch (err) {
    deleteModalError.value = err.response?.data?.errorMessage || t('repositoriesPage.failedDelete')
  } finally {
    deletingId.value = null
  }
}

async function triggerClone(repo) {
  cloningId.value = repo.id
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const res = await repositoriesAPI.clone(workspace.value.id, repo.id)
    successMessage.value = t('repositoriesPage.cloneStarted')
    await loadRepositories()
    const taskId = res.data?.id
    if (taskId) {
      cloneTaskModal.value = { taskId, repo }
      cloneTaskData.value = null
      cloneTaskLoadError.value = ''
      fetchCloneTask()
    }
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('repositoriesPage.cloneFailed')
  } finally {
    cloningId.value = null
  }
}

function closeCloneTaskModal() {
  if (cloneTaskPollTimer) {
    clearTimeout(cloneTaskPollTimer)
    cloneTaskPollTimer = null
  }
  cloneTaskModal.value = null
  cloneTaskData.value = null
  cloneTaskLoadError.value = ''
  loadRepositories()
}

async function fetchCloneTask() {
  const modal = cloneTaskModal.value
  if (!modal) return
  const { taskId, repo } = modal

  try {
    const res = await repositoriesAPI.getCloneTask(workspace.value.id, repo.id, taskId)
    cloneTaskData.value = res.data
    cloneTaskLoadError.value = ''
    const status = (res.data?.status || '').toLowerCase()
    if (status === 'pending' || status === 'running') {
      cloneTaskPollTimer = setTimeout(fetchCloneTask, 2000)
    }
  } catch (err) {
    cloneTaskLoadError.value = err.response?.data?.errorMessage || t('repositoriesPage.failedLoad')
    cloneTaskPollTimer = setTimeout(fetchCloneTask, 2000)
  }
}

const cloneTaskLogLines = computed(() => {
  const result = cloneTaskData.value?.result
  if (result == null) return []
  let arr = result
  if (typeof result === 'string') {
    try {
      arr = JSON.parse(result)
    } catch {
      return [{ stream: '', text: result }]
    }
  }
  if (!Array.isArray(arr)) return []
  return arr.map((entry) => {
    const stream = entry?.stream || ''
    const line = entry?.line ?? ''
    const msg = entry?.msg ?? ''
    const text = line || msg || JSON.stringify(entry)
    return { stream, text }
  })
})

const cloneTaskLogText = computed(() => {
  return cloneTaskLogLines.value.map((l) => l.text).join('\n')
})

function titleCase(s) {
  if (!s || typeof s !== 'string') return s
  return s.toLowerCase().replace(/\b\w/g, (c) => c.toUpperCase())
}

onMounted(() => {
  setWorkspace()
  if (workspace.value?.id) {
    loadRepositories()
  }
})

onUnmounted(() => {
  if (cloneTaskPollTimer) {
    clearTimeout(cloneTaskPollTimer)
  }
})

watch(workspace, (w) => {
  if (w?.id) loadRepositories()
}, { immediate: false })
</script>
