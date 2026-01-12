<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('groupVarsPage.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('groupVarsPage.subtitle') }}
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" :aria-label="$t('groupVarsPage.sectionTitle')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('groupVarsPage.sectionTitle') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('groupVarsPage.sectionDescription') }}</p>
          </div>
          <button type="button" class="btn-primary text-sm inline-flex items-center gap-2" @click="openCreateModal">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('groupVarsPage.addGroup') }}
          </button>
        </div>

        <div v-if="loading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('groupVarsPage.loading') }}</div>
        <div v-else-if="groups.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('groupVarsPage.empty') }}</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('groupVarsPage.groupName') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('groupVarsPage.createdAt') }}</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('groupVarsPage.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="g in groups" :key="g.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <p class="text-sm font-medium text-theme-text">{{ g.name }}</p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ formatCreatedAt(g.createdAt) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button type="button" class="text-primary-600 hover:text-primary-700 p-1 rounded mr-1" :title="$t('common.edit')" :aria-label="$t('common.edit')" @click="openEditModal(g)">
                      <svg class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </button>
                    <button type="button" class="text-red-600 hover:text-red-700 p-1 rounded" :title="$t('common.delete')" :aria-label="$t('common.delete')" @click="openDeleteModal(g)">
                      <svg class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="px-6 py-4 border-t border-theme-border flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm text-theme-textLight">
              <template v-if="hasPagination">
                {{ $t('groupVarsPage.showing', { from: showingFrom, to: showingTo, total: totalGroups }) }}
              </template>
              <template v-else-if="_meta?.total">
                {{ _meta.total === 1 ? $t('groupVarsPage.countOne') : $t('groupVarsPage.countMany', { count: _meta.total }) }}
              </template>
            </div>
            <div v-if="hasPagination" class="flex items-center gap-2">
              <button type="button" class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed" :disabled="offset === 0" @click="goToPage(offset - limit)">
                {{ $t('groupVarsPage.previous') }}
              </button>
              <button type="button" class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed" :disabled="offset + limit >= totalGroups" @click="goToPage(offset + limit)">
                {{ $t('groupVarsPage.next') }}
              </button>
            </div>
          </div>
        </template>
      </section>
    </main>

    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="editModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="editModalTitleId"
          @click.self="closeEditModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-3xl max-h-[92vh] flex flex-col">
            <div class="px-6 py-4 border-b border-theme-border flex items-start justify-between gap-3 shrink-0">
              <h2 :id="editModalTitleId" class="text-lg font-semibold text-theme-text pr-8">
                {{ editingGroup ? $t('groupVarsPage.editGroup') : $t('groupVarsPage.createGroup') }}
              </h2>
              <button type="button" class="text-theme-textLight hover:text-theme-text p-1 rounded shrink-0 -mr-1" :aria-label="$t('common.close')" @click="closeEditModal">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <form class="flex flex-col flex-1 min-h-0" @submit.prevent="handleSaveGroup">
              <div class="px-6 py-4 overflow-y-auto flex-1 space-y-4">
                <div v-if="modalError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
                  {{ modalError }}
                </div>

                <div>
                  <label for="gvar-group-name" class="form-label">{{ $t('groupVarsPage.groupName') }}</label>
                  <input
                    id="gvar-group-name"
                    v-model="editName"
                    type="text"
                    required
                    maxlength="255"
                    class="input-field"
                  />
                </div>

                <div class="border-b border-theme-border flex gap-6" role="tablist">
                  <button
                    type="button"
                    role="tab"
                    :aria-selected="modalTab === 'variables'"
                    class="pb-2 text-sm font-medium border-b-2 -mb-px transition-colors"
                    :class="modalTab === 'variables' ? 'border-primary-600 text-primary-600' : 'border-transparent text-theme-textLight hover:text-theme-text'"
                    @click="modalTab = 'variables'"
                  >
                    {{ $t('groupVarsPage.tabVariables') }}
                  </button>
                  <button
                    type="button"
                    role="tab"
                    :aria-selected="modalTab === 'secrets'"
                    class="pb-2 text-sm font-medium border-b-2 -mb-px transition-colors"
                    :class="modalTab === 'secrets' ? 'border-primary-600 text-primary-600' : 'border-transparent text-theme-textLight hover:text-theme-text'"
                    @click="modalTab = 'secrets'"
                  >
                    {{ $t('groupVarsPage.tabSecrets') }}
                  </button>
                </div>

                <div v-show="modalTab === 'variables'" class="space-y-6">
                  <VarSection
                    :title="$t('groupVarsPage.extraVars')"
                    :help="$t('groupVarsPage.extraVarsHelp')"
                    :view-mode="plainExtraViewMode"
                    :rows="plainExtraRows"
                    :json-text="plainExtraJson"
                    :value-secret="false"
                    @update:view-mode="plainExtraViewMode = $event"
                    @update:rows="plainExtraRows = $event"
                    @update:json-text="plainExtraJson = $event"
                  />
                  <EnvSection
                    :title="$t('groupVarsPage.envVars')"
                    :rows="plainEnvRows"
                    :value-secret="false"
                    @update:rows="plainEnvRows = $event"
                  />
                </div>

                <div v-show="modalTab === 'secrets'" class="space-y-6">
                  <VarSection
                    :title="$t('groupVarsPage.extraVars')"
                    :help="$t('groupVarsPage.extraVarsHelp')"
                    :view-mode="secretExtraViewMode"
                    :rows="secretExtraRows"
                    :json-text="secretExtraJson"
                    :value-secret="true"
                    @update:view-mode="secretExtraViewMode = $event"
                    @update:rows="secretExtraRows = $event"
                    @update:json-text="secretExtraJson = $event"
                  />
                  <EnvSection
                    :title="$t('groupVarsPage.envVars')"
                    :rows="secretEnvRows"
                    :value-secret="true"
                    @update:rows="secretEnvRows = $event"
                  />
                </div>
              </div>

              <div class="px-6 py-4 border-t border-theme-border flex justify-end gap-3 shrink-0 bg-white rounded-b-lg">
                <button type="button" class="btn-secondary" @click="closeEditModal">{{ $t('common.cancel') }}</button>
                <button type="submit" class="btn-primary" :disabled="editSaving">
                  {{ editSaving ? $t('groupVarsPage.saving') : $t('common.save') }}
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
          v-show="deleteModalGroup"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-gvar-title"
          @click.self="deleteModalGroup = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-gvar-title" class="text-lg font-semibold text-theme-text">{{ $t('groupVarsPage.deleteTitle') }}</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('groupVarsPage.deleteConfirm', { name: deleteModalGroup?.name }) }}
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" class="btn-secondary" @click="deleteModalGroup = null">{{ $t('common.cancel') }}</button>
              <button type="button" class="btn-danger" :disabled="deletingId === deleteModalGroup?.id" @click="confirmDelete">
                {{ deletingId === deleteModalGroup?.id ? $t('groupVarsPage.deleting') : $t('common.delete') }}
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
import { variableGroupsAPI } from '@/api'
import VarSection from '@/components/groupVars/VarSection.vue'
import EnvSection from '@/components/groupVars/EnvSection.vue'

const { t, locale } = useI18n()
const router = useRouter()

const workspace = ref(null)
const groups = ref([])
const _meta = ref(null)
const loading = ref(true)
const limit = ref(50)
const offset = ref(0)
const errorMessage = ref('')
const successMessage = ref('')

const editModalOpen = ref(false)
const editModalTitleId = 'gvar-edit-title'
const editingGroup = ref(null)
const editSaving = ref(false)
const modalError = ref('')
const modalTab = ref('variables')
const editName = ref('')

const plainExtraViewMode = ref('table')
const plainExtraJson = ref('{}')
const plainExtraRows = ref([])
const plainEnvRows = ref([])

const secretExtraViewMode = ref('table')
const secretExtraJson = ref('{}')
const secretExtraRows = ref([])
const secretEnvRows = ref([])

const deleteModalGroup = ref(null)
const deletingId = ref(null)

function formatCreatedAt(iso) {
  if (!iso) return '—'
  try {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return '—'
    return d.toLocaleDateString(locale.value || 'en', { month: 'short', day: 'numeric', year: 'numeric' })
  } catch {
    return '—'
  }
}

function valueToDisplay(v) {
  if (v === null || v === undefined) return ''
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

function jsonObjectToRows(raw) {
  if (!raw || !String(raw).trim()) return []
  try {
    const o = JSON.parse(String(raw))
    if (o === null || typeof o !== 'object' || Array.isArray(o)) return []
    return Object.keys(o).map((k) => ({ key: k, value: valueToDisplay(o[k]) }))
  } catch {
    return []
  }
}

function rowsToObject(rows) {
  const o = {}
  for (const r of rows) {
    const k = r.key?.trim()
    if (!k) continue
    o[k] = r.value ?? ''
  }
  return o
}

function prettyJsonString(raw) {
  if (!raw || !String(raw).trim()) return '{\n}'
  try {
    const o = JSON.parse(String(raw))
    if (o === null || typeof o !== 'object' || Array.isArray(o)) return '{\n}'
    return JSON.stringify(o, null, 2)
  } catch {
    return '{\n}'
  }
}

function serializeExtraField(viewMode, rows, jsonText) {
  if (viewMode === 'json') {
    const raw = (jsonText || '').trim() || '{}'
    let parsed
    try {
      parsed = JSON.parse(raw)
    } catch {
      throw new Error(t('groupVarsPage.invalidJson'))
    }
    if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
      throw new Error(t('groupVarsPage.invalidJson'))
    }
    return JSON.stringify(parsed)
  }
  return JSON.stringify(rowsToObject(rows))
}

function resetEditForm() {
  modalTab.value = 'variables'
  editName.value = ''
  plainExtraViewMode.value = 'table'
  plainExtraJson.value = '{\n}'
  plainExtraRows.value = []
  plainEnvRows.value = []
  secretExtraViewMode.value = 'table'
  secretExtraJson.value = '{\n}'
  secretExtraRows.value = []
  secretEnvRows.value = []
  modalError.value = ''
}

function fillFormFromGroup(g) {
  editName.value = g.name || ''
  plainExtraRows.value = jsonObjectToRows(g.extraVariables)
  plainExtraJson.value = prettyJsonString(g.extraVariables)
  plainEnvRows.value = jsonObjectToRows(g.environmentVariables)
  secretExtraRows.value = jsonObjectToRows(g.secretExtraVariables)
  secretExtraJson.value = prettyJsonString(g.secretExtraVariables)
  secretEnvRows.value = jsonObjectToRows(g.secretEnvironmentVariables)
  plainExtraViewMode.value = 'table'
  secretExtraViewMode.value = 'table'
}

function setWorkspace() {
  workspace.value = loadWorkspaceFromStorage()
  if (!workspace.value?.id) {
    router.replace('/select-workspace')
  }
}

async function loadGroups(overrideOffset) {
  if (!workspace.value?.id) return
  loading.value = true
  errorMessage.value = ''
  const off = overrideOffset !== undefined ? overrideOffset : offset.value

  try {
    const res = await variableGroupsAPI.list(workspace.value.id, { limit: limit.value, offset: off })
    groups.value = res.data?.groups ?? []
    _meta.value = res.data?._meta ?? null
    offset.value = off
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('groupVarsPage.failedLoad')
    groups.value = []
  } finally {
    loading.value = false
  }
}

const totalGroups = computed(() => _meta.value?.total ?? 0)
const showingFrom = computed(() => (totalGroups.value === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + groups.value.length, totalGroups.value))
const hasPagination = computed(() => totalGroups.value > limit.value)

function goToPage(newOffset) {
  const maxOffset = Math.max(0, totalGroups.value - limit.value)
  const off = Math.max(0, Math.min(newOffset, maxOffset))
  loadGroups(off)
}

function openCreateModal() {
  editingGroup.value = null
  resetEditForm()
  editModalOpen.value = true
}

function openEditModal(g) {
  editingGroup.value = g
  resetEditForm()
  fillFormFromGroup(g)
  editModalOpen.value = true
}

function closeEditModal() {
  editModalOpen.value = false
  editingGroup.value = null
  resetEditForm()
}

async function handleSaveGroup() {
  if (!workspace.value?.id) return
  modalError.value = ''
  let extraVariables
  let environmentVariables
  let secretExtraVariables
  let secretEnvironmentVariables
  try {
    extraVariables = serializeExtraField(plainExtraViewMode.value, plainExtraRows.value, plainExtraJson.value)
    environmentVariables = JSON.stringify(rowsToObject(plainEnvRows.value))
    secretExtraVariables = serializeExtraField(secretExtraViewMode.value, secretExtraRows.value, secretExtraJson.value)
    secretEnvironmentVariables = JSON.stringify(rowsToObject(secretEnvRows.value))
  } catch (e) {
    modalError.value = e.message || t('groupVarsPage.modalError')
    return
  }

  const payload = {
    name: editName.value.trim(),
    extraVariables,
    environmentVariables,
    secretExtraVariables,
    secretEnvironmentVariables,
  }

  editSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    if (editingGroup.value?.id) {
      await variableGroupsAPI.update(workspace.value.id, editingGroup.value.id, payload)
      successMessage.value = t('groupVarsPage.groupUpdated')
    } else {
      await variableGroupsAPI.create(workspace.value.id, payload)
      successMessage.value = t('groupVarsPage.groupCreated')
    }
    closeEditModal()
    await loadGroups()
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) {
    const msg = err.response?.data?.errorMessage || err.response?.data?.message
    if (msg) {
      modalError.value = msg
    } else {
      errorMessage.value = t('groupVarsPage.failedSave')
    }
  } finally {
    editSaving.value = false
  }
}

function openDeleteModal(g) {
  deleteModalGroup.value = g
}

async function confirmDelete() {
  const g = deleteModalGroup.value
  if (!workspace.value?.id || !g?.id) return
  deletingId.value = g.id
  errorMessage.value = ''

  try {
    await variableGroupsAPI.delete(workspace.value.id, g.id)
    successMessage.value = t('groupVarsPage.groupDeleted')
    deleteModalGroup.value = null
    await loadGroups()
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('groupVarsPage.failedDelete')
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  setWorkspace()
  if (workspace.value?.id) {
    loadGroups()
  }
})

watch(workspace, (w) => {
  if (w?.id) loadGroups()
}, { immediate: false })
</script>
