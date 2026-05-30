<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('taskTemplatesPage.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('taskTemplatesPage.subtitle') }}
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-visible" :aria-label="$t('taskTemplatesPage.sectionTitle')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('taskTemplatesPage.sectionTitle') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('taskTemplatesPage.sectionDescription') }}</p>
          </div>
          <div ref="addTemplateDropdownRef" class="relative shrink-0">
            <button
              type="button"
              class="btn-primary text-sm inline-flex items-center gap-2"
              :aria-expanded="addTemplateMenuOpen"
              aria-haspopup="menu"
              :aria-label="$t('taskTemplatesPage.addTemplateMenuAria')"
              @click.stop="toggleAddTemplateMenu"
            >
              <svg class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              {{ $t('taskTemplatesPage.addTemplate') }}
              <svg
                class="h-4 w-4 shrink-0 text-white/90 transition-transform"
                :class="{ 'rotate-180': addTemplateMenuOpen }"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <Transition
              enter-active-class="transition ease-out duration-100"
              enter-from-class="opacity-0 scale-95"
              enter-to-class="opacity-100 scale-100"
              leave-active-class="transition ease-in duration-75"
              leave-from-class="opacity-100 scale-100"
              leave-to-class="opacity-0 scale-95"
            >
              <div
                v-show="addTemplateMenuOpen"
                class="absolute right-0 top-full z-40 mt-2 w-56 min-w-max rounded-lg border border-theme-border bg-white py-1 shadow-lg origin-top-right"
                role="menu"
                :aria-label="$t('taskTemplatesPage.addTemplateMenuAria')"
              >
                <button
                  type="button"
                  role="menuitem"
                  class="flex w-full items-center px-4 py-2.5 text-left text-sm text-theme-text hover:bg-theme-hover"
                  @click="openCreateModal('ansible')"
                >
                  {{ $t('taskTemplatesPage.addAnsibleTemplate') }}
                </button>
                <button
                  type="button"
                  role="menuitem"
                  class="flex w-full items-center px-4 py-2.5 text-left text-sm text-theme-text hover:bg-theme-hover"
                  @click="openCreateModal('iac')"
                >
                  {{ $t('taskTemplatesPage.addIacTemplate') }}
                </button>
                <button
                  type="button"
                  role="menuitem"
                  class="flex w-full items-center px-4 py-2.5 text-left text-sm text-theme-text hover:bg-theme-hover"
                  @click="openCreateModal('script')"
                >
                  {{ $t('taskTemplatesPage.addScriptTemplate') }}
                </button>
              </div>
            </Transition>
          </div>
        </div>

        <div v-if="loading" class="p-8 text-center text-theme-textLight text-sm rounded-b-xl">{{ $t('taskTemplatesPage.loading') }}</div>
        <div v-else-if="templates.length === 0" class="p-8 text-center text-theme-textLight text-sm rounded-b-xl">{{ $t('taskTemplatesPage.empty') }}</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('taskTemplatesPage.name') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('taskTemplatesPage.engine') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('taskTemplatesPage.repository') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('taskTemplatesPage.updatedAt') }}</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('taskTemplatesPage.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="tmpl in templates" :key="tmpl.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4">
                    <p class="text-sm font-medium text-theme-text">{{ tmpl.name }}</p>
                    <p v-if="tmpl.description" class="text-xs text-theme-textLight mt-0.5 line-clamp-2">{{ tmpl.description }}</p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">{{ formatTemplateEngine(tmpl.engine) }}</td>
                  <td class="px-6 py-4 text-sm text-theme-textLight">
                    <span :title="String(tmpl.repositoryId ?? '')">{{ repositoryLabel(tmpl.repositoryId) }}</span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ formatCreatedAt(tmpl.updatedAt, locale) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button type="button" class="text-primary-600 hover:text-primary-700 p-1 rounded mr-1" :title="$t('common.edit')" :aria-label="$t('common.edit')" @click="openEditModal(tmpl)">
                      <svg class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </button>
                    <button type="button" class="text-red-600 hover:text-red-700 p-1 rounded" :title="$t('common.delete')" :aria-label="$t('common.delete')" @click="openDeleteModal(tmpl)">
                      <svg class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="px-6 py-4 border-t border-theme-border flex flex-wrap items-center justify-between gap-3 rounded-b-xl">
            <div class="text-sm text-theme-textLight">
              <template v-if="hasPagination">
                {{ $t('taskTemplatesPage.showing', { from: showingFrom, to: showingTo, total: totalTemplates }) }}
              </template>
              <template v-else-if="_meta?.total">
                {{ _meta.total === 1 ? $t('taskTemplatesPage.countOne') : $t('taskTemplatesPage.countMany', { count: _meta.total }) }}
              </template>
            </div>
            <div v-if="hasPagination" class="flex items-center gap-2">
              <button type="button" class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed" :disabled="offset === 0" @click="goToPage(offset - limit)">
                {{ $t('taskTemplatesPage.previous') }}
              </button>
              <button type="button" class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed" :disabled="offset + limit >= totalTemplates" @click="goToPage(offset + limit)">
                {{ $t('taskTemplatesPage.next') }}
              </button>
            </div>
          </div>
        </template>
      </section>
    </main>

    <template v-for="eng in templateModalEngines" :key="eng">
      <TaskTemplateEngineModal
        v-if="templateModalEngine === eng"
        :engine="eng"
        :editing="!!editingTemplate"
        :title-id="`tmpl-modal-${eng}-title`"
        :form="form"
        :runner-tag-entries="runnerTagEntries"
        :modal-error="modalError"
        :edit-saving="editSaving"
        :repositories="repositories"
        :group-vars-list="groupVarsList"
        :backends-list="backendsList"
        :inventories-list="inventoriesList"
        :script-engine-options="scriptEngineOptions"
        :iac-log-level-options="iacLogLevelOptions"
        :ansible-log-level-options="ansibleLogLevelOptions"
        :add-runner-tag-row="addTagRow"
        :remove-runner-tag-row="removeTagRow"
        @close="closeTemplateModal"
        @save="handleSaveTemplate"
      />
    </template>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteModalTemplate"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="'delete-tmpl-title'"
          @click.self="deleteModalTemplate = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-tmpl-title" class="text-lg font-semibold text-theme-text">{{ $t('taskTemplatesPage.deleteTitle') }}</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('taskTemplatesPage.deleteConfirm', { name: deleteModalTemplate?.name }) }}
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" class="btn-secondary" @click="deleteModalTemplate = null">{{ $t('common.cancel') }}</button>
              <button type="button" class="btn-danger" :disabled="deletingId === deleteModalTemplate?.id" @click="confirmDelete">
                {{ deletingId === deleteModalTemplate?.id ? $t('taskTemplatesPage.deleting') : $t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import TaskTemplateEngineModal from '@/components/TaskTemplateEngineModal.vue'
import { loadWorkspaceFromStorage } from '@/utils/storage'
import { formatCreatedAt } from '@/utils/date'
import {
  workspaceTemplatesAPI,
  repositoriesAPI,
  variableGroupsAPI,
  backendsAPI,
  inventoriesAPI,
} from '@/api'

const { t, locale } = useI18n()
const router = useRouter()

const workspace = ref(null)
const templates = ref([])
const repositories = ref([])
const groupVarsList = ref([])
const backendsList = ref([])
const inventoriesList = ref([])
const _meta = ref(null)
const loading = ref(true)
const limit = ref(50)
const offset = ref(0)
const errorMessage = ref('')
const successMessage = ref('')

const templateModalEngines = ['ansible', 'iac', 'script']

const addTemplateDropdownRef = ref(null)
const addTemplateMenuOpen = ref(false)
/** Which template modal is open: 'script' | 'iac' | 'ansible' | null */
const templateModalEngine = ref(null)
const editingTemplate = ref(null)
const editSaving = ref(false)
const modalError = ref('')

const deleteModalTemplate = ref(null)
const deletingId = ref(null)

/** Matches API `Script.engine` oneof */
const scriptEngineOptions = ['bash', 'sh', 'python3', 'node']
/** Matches API `IaC.logLevel` oneof */
const iacLogLevelOptions = ['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR']
/** Matches API `Ansible.logLevel` oneof */
const ansibleLogLevelOptions = [0, 1, 2, 3, 4, 5]

function pickConfigField(obj, camelKey, snakeKey) {
  if (!obj || typeof obj !== 'object') return ''
  if (Object.prototype.hasOwnProperty.call(obj, camelKey)) {
    const v = obj[camelKey]
    return v == null ? '' : String(v)
  }
  if (snakeKey && Object.prototype.hasOwnProperty.call(obj, snakeKey)) {
    const v = obj[snakeKey]
    return v == null ? '' : String(v)
  }
  return ''
}

function formatTemplateEngine(engine) {
  const e = engine != null && engine !== '' ? String(engine) : ''
  if (!e) return '—'
  if (e === 'script') return t('taskTemplatesPage.engineScript')
  if (e === 'iac') return t('taskTemplatesPage.engineIac')
  if (e === 'ansible') return t('taskTemplatesPage.engineAnsible')
  if (e === 'terraform') return t('taskTemplatesPage.engineTerraform')
  if (e === 'opentofu') return t('taskTemplatesPage.engineOpenTofu')
  return e
}

function defaultFormState() {
  return {
    name: '',
    description: '',
    engine: 'script',
    repositoryId: '',
    workingDirectory: '',
    branchOverride: '',
    tagOverride: '',
    allowConcurrentRuns: false,
    maxConcurrentRuns: 0,
    approvalRequired: false,
    extraArgs: '',
    groupVarsId: '',
    script: { engine: 'bash', engineVersion: '', scriptPath: '' },
    iac: { engine: 'terraform', engineVersion: '', autoApprove: false, backendID: '', logLevel: '' },
    ansible: { engineVersion: '', inventoryID: '', playbookPath: '', logLevel: '' },
  }
}

const form = reactive(defaultFormState())
const runnerTagEntries = ref([{ key: '', value: '' }])

function resetForm() {
  Object.assign(form, defaultFormState())
  runnerTagEntries.value = [{ key: '', value: '' }]
  modalError.value = ''
}

function tagsObjectFromEntries(entries) {
  const o = {}
  for (const { key, value } of entries) {
    const k = (key || '').trim()
    if (!k) continue
    o[k] = (value ?? '').trim()
  }
  return o
}

function entriesFromTags(tags) {
  if (!tags || typeof tags !== 'object') {
    return [{ key: '', value: '' }]
  }
  const pairs = Object.entries(tags).map(([key, value]) => ({
    key,
    value: String(value ?? ''),
  }))
  return pairs.length ? pairs : [{ key: '', value: '' }]
}

function addTagRow() {
  runnerTagEntries.value.push({ key: '', value: '' })
}

function removeTagRow(idx) {
  runnerTagEntries.value.splice(idx, 1)
  if (runnerTagEntries.value.length === 0) {
    runnerTagEntries.value.push({ key: '', value: '' })
  }
}

/** Normalize API `engine` to modal kind (script | iac | ansible). */
function templateModalKindFromEngine(raw) {
  const e = raw != null && raw !== '' ? String(raw).toLowerCase() : ''
  if (e === 'terraform' || e === 'opentofu') return 'iac'
  if (e === 'script' || e === 'iac' || e === 'ansible') return e
  return 'script'
}

function fillFormFromTemplate(tmpl) {
  const rawEngine = tmpl.engine != null ? String(tmpl.engine) : 'script'
  let engine = rawEngine
  if (rawEngine === 'terraform' || rawEngine === 'opentofu') {
    engine = 'iac'
  }

  Object.assign(form, {
    name: tmpl.name || '',
    description: tmpl.description || '',
    engine,
    repositoryId: tmpl.repositoryId != null ? String(tmpl.repositoryId) : '',
    workingDirectory: tmpl.workingDirectory || '',
    branchOverride: tmpl.branchOverride || '',
    tagOverride: tmpl.tagOverride || '',
    allowConcurrentRuns: !!tmpl.allowConcurrentRuns,
    maxConcurrentRuns: Number(tmpl.maxConcurrentRuns) || 0,
    approvalRequired: !!tmpl.approvalRequired,
    extraArgs: tmpl.extraArgs || '',
    groupVarsId: tmpl.groupVarsId ? String(tmpl.groupVarsId) : '',
  })

  const s = tmpl.script && typeof tmpl.script === 'object' ? tmpl.script : {}
  form.script.engine = pickConfigField(s, 'engine') || 'bash'
  if (!scriptEngineOptions.includes(form.script.engine)) {
    form.script.engine = 'bash'
  }
  form.script.engineVersion = pickConfigField(s, 'engineVersion', 'engine_version')
  form.script.scriptPath = pickConfigField(s, 'scriptPath', 'script_path')

  const i = tmpl.iac && typeof tmpl.iac === 'object' ? tmpl.iac : {}
  if (rawEngine === 'terraform' || rawEngine === 'opentofu') {
    form.iac.engine = rawEngine
  } else {
    const prov = pickConfigField(i, 'engine')
    form.iac.engine = prov === 'opentofu' ? 'opentofu' : 'terraform'
  }
  form.iac.engineVersion = pickConfigField(i, 'engineVersion', 'engine_version')
  form.iac.autoApprove = !!(i.autoApprove ?? i.auto_approve)
  form.iac.backendID = pickConfigField(i, 'backendID', 'backend_id')
  form.iac.logLevel = pickConfigField(i, 'logLevel', 'log_level')

  const a = tmpl.ansible && typeof tmpl.ansible === 'object' ? tmpl.ansible : {}
  form.ansible.engineVersion = pickConfigField(a, 'engineVersion', 'engine_version')
  form.ansible.inventoryID = pickConfigField(a, 'inventoryID', 'inventory_id')
  form.ansible.playbookPath = pickConfigField(a, 'playbookPath', 'playbook_path')
  form.ansible.logLevel = pickConfigField(a, 'logLevel', 'log_level')

  runnerTagEntries.value = entriesFromTags(tmpl.runnerTags)
}

function buildPayload() {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    engine: form.engine.trim(),
    repositoryId: form.repositoryId,
    workingDirectory: form.workingDirectory.trim(),
    branchOverride: form.branchOverride.trim(),
    tagOverride: form.tagOverride.trim(),
    runnerTags: tagsObjectFromEntries(runnerTagEntries.value),
    allowConcurrentRuns: form.allowConcurrentRuns,
    maxConcurrentRuns: Number(form.maxConcurrentRuns) || 0,
    approvalRequired: form.approvalRequired,
    extraArgs: form.extraArgs,
    groupVarsId: form.groupVarsId || '',
    script: {
      engine: form.script.engine || '',
      engineVersion: form.script.engineVersion || '',
      scriptPath: form.script.scriptPath || '',
    },
    iac: {
      engine: form.iac.engine || '',
      engineVersion: form.iac.engineVersion || '',
      autoApprove: form.iac.autoApprove,
      backendID: form.iac.backendID || '',
      logLevel: form.iac.logLevel || '',
    },
    ansible: {
      engineVersion: form.ansible.engineVersion || '',
      inventoryID: form.ansible.inventoryID || '',
      playbookPath: form.ansible.playbookPath || '',
      logLevel: form.ansible.logLevel || '',
    },
  }
}

function repositoryLabel(repoId) {
  if (repoId == null || repoId === '') return '—'
  const id = String(repoId)
  const r = repositories.value.find((x) => String(x.id) === id)
  return r?.name ? r.name : id
}

function setWorkspace() {
  workspace.value = loadWorkspaceFromStorage()
  if (!workspace.value?.id) {
    router.replace('/select-workspace')
  }
}

async function loadAuxiliaryLists() {
  if (!workspace.value?.id) return
  const ws = workspace.value.id
  try {
    const [reposRes, gvRes, beRes, invRes] = await Promise.all([
      repositoriesAPI.list(ws, { limit: 100, offset: 0 }),
      variableGroupsAPI.list(ws, { limit: 100, offset: 0 }),
      backendsAPI.list(ws, { limit: 100, offset: 0 }),
      inventoriesAPI.list(ws, { limit: 100, offset: 0 }),
    ])
    repositories.value = reposRes.data?.repositories ?? []
    groupVarsList.value = gvRes.data?.groups ?? []
    backendsList.value = beRes.data?.backends ?? []
    inventoriesList.value = invRes.data?.inventories ?? []
  } catch (err) {
    console.error('Failed to load template form lookups:', err)
  }
}

async function loadTemplates(overrideOffset) {
  if (!workspace.value?.id) return
  loading.value = true
  errorMessage.value = ''
  const off = overrideOffset !== undefined ? overrideOffset : offset.value

  try {
    const res = await workspaceTemplatesAPI.list(workspace.value.id, { limit: limit.value, offset: off })
    templates.value = res.data?.templates ?? []
    _meta.value = res.data?._meta ?? null
    offset.value = off
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('taskTemplatesPage.failedLoad')
    templates.value = []
  } finally {
    loading.value = false
  }
}

const totalTemplates = computed(() => _meta.value?.total ?? 0)
const showingFrom = computed(() => (totalTemplates.value === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + templates.value.length, totalTemplates.value))
const hasPagination = computed(() => totalTemplates.value > limit.value)

function goToPage(newOffset) {
  const maxOffset = Math.max(0, totalTemplates.value - limit.value)
  const off = Math.max(0, Math.min(newOffset, maxOffset))
  loadTemplates(off)
}

function toggleAddTemplateMenu() {
  addTemplateMenuOpen.value = !addTemplateMenuOpen.value
}

function onClickOutsideAddTemplate(event) {
  if (addTemplateDropdownRef.value && !addTemplateDropdownRef.value.contains(event.target)) {
    addTemplateMenuOpen.value = false
  }
}

/** @param {'script' | 'iac' | 'ansible'} engine */
function openCreateModal(engine) {
  editingTemplate.value = null
  resetForm()
  if (engine === 'script' || engine === 'iac' || engine === 'ansible') {
    form.engine = engine
  }
  addTemplateMenuOpen.value = false
  templateModalEngine.value = templateModalKindFromEngine(form.engine)
}

async function openEditModal(tmpl) {
  addTemplateMenuOpen.value = false
  editingTemplate.value = tmpl
  resetForm()
  try {
    const res = await workspaceTemplatesAPI.get(workspace.value.id, tmpl.id)
    const full = res.data
    if (full && typeof full === 'object') {
      fillFormFromTemplate(full)
    } else {
      fillFormFromTemplate(tmpl)
    }
  } catch (err) {
    console.error('Failed to load template:', err)
    fillFormFromTemplate(tmpl)
  }
  templateModalEngine.value = templateModalKindFromEngine(form.engine)
}

function closeTemplateModal() {
  templateModalEngine.value = null
  editingTemplate.value = null
  resetForm()
}

async function handleSaveTemplate() {
  if (!workspace.value?.id) return
  if (!form.description.trim()) {
    modalError.value = t('taskTemplatesPage.descriptionRequired')
    return
  }
  if (!form.repositoryId) {
    modalError.value = t('taskTemplatesPage.repositoryRequired')
    return
  }
  modalError.value = ''
  editSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''

  const payload = buildPayload()

  try {
    if (editingTemplate.value?.id) {
      await workspaceTemplatesAPI.update(workspace.value.id, editingTemplate.value.id, payload)
      successMessage.value = t('taskTemplatesPage.templateUpdated')
    } else {
      await workspaceTemplatesAPI.create(workspace.value.id, payload)
      successMessage.value = t('taskTemplatesPage.templateCreated')
    }
    closeTemplateModal()
    await loadTemplates()
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) {
    const msg = err.response?.data?.errorMessage || err.response?.data?.message
    if (msg) {
      modalError.value = msg
    } else {
      errorMessage.value = t('taskTemplatesPage.failedSave')
    }
  } finally {
    editSaving.value = false
  }
}

function openDeleteModal(tmpl) {
  deleteModalTemplate.value = tmpl
}

async function confirmDelete() {
  const tmpl = deleteModalTemplate.value
  if (!workspace.value?.id || !tmpl?.id) return
  deletingId.value = tmpl.id
  errorMessage.value = ''

  try {
    await workspaceTemplatesAPI.delete(workspace.value.id, tmpl.id)
    successMessage.value = t('taskTemplatesPage.templateDeleted')
    deleteModalTemplate.value = null
    await loadTemplates()
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('taskTemplatesPage.failedDelete')
  } finally {
    deletingId.value = null
  }
}

onMounted(async () => {
  document.addEventListener('click', onClickOutsideAddTemplate)
  setWorkspace()
  if (workspace.value?.id) {
    await loadAuxiliaryLists()
    loadTemplates()
  }
})

onUnmounted(() => {
  document.removeEventListener('click', onClickOutsideAddTemplate)
})
</script>
