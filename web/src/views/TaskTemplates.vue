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

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" :aria-label="$t('taskTemplatesPage.sectionTitle')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('taskTemplatesPage.sectionTitle') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('taskTemplatesPage.sectionDescription') }}</p>
          </div>
          <button type="button" class="btn-primary text-sm inline-flex items-center gap-2" @click="openCreateModal">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('taskTemplatesPage.addTemplate') }}
          </button>
        </div>

        <div v-if="loading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('taskTemplatesPage.loading') }}</div>
        <div v-else-if="templates.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('taskTemplatesPage.empty') }}</div>
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
          <div class="px-6 py-4 border-t border-theme-border flex flex-wrap items-center justify-between gap-3">
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
                {{ editingTemplate ? $t('taskTemplatesPage.editTemplate') : $t('taskTemplatesPage.createTemplate') }}
              </h2>
              <button type="button" class="text-theme-textLight hover:text-theme-text p-1 rounded shrink-0 -mr-1" :aria-label="$t('common.close')" @click="closeEditModal">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <form class="flex flex-col flex-1 min-h-0" @submit.prevent="handleSaveTemplate">
              <div class="px-6 py-3 border-b border-theme-border flex gap-4 flex-wrap shrink-0" role="tablist">
                <button
                  v-for="tab in modalTabs"
                  :key="tab.id"
                  type="button"
                  role="tab"
                  :aria-selected="modalTab === tab.id"
                  class="pb-2 text-sm font-medium border-b-2 -mb-px transition-colors"
                  :class="modalTab === tab.id ? 'border-primary-600 text-primary-600' : 'border-transparent text-theme-textLight hover:text-theme-text'"
                  @click="modalTab = tab.id"
                >
                  {{ tab.label }}
                </button>
              </div>

              <div class="px-6 py-4 overflow-y-auto flex-1 space-y-4">
                <div v-if="modalError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
                  {{ modalError }}
                </div>

                <div v-show="modalTab === 'general'" class="space-y-4">
                  <div>
                    <label for="tmpl-name" class="form-label">{{ $t('taskTemplatesPage.name') }}</label>
                    <input id="tmpl-name" v-model="form.name" type="text" required maxlength="100" class="input-field" />
                  </div>
                  <div>
                    <label for="tmpl-desc" class="form-label">{{ $t('taskTemplatesPage.description') }}</label>
                    <textarea id="tmpl-desc" v-model="form.description" rows="2" required maxlength="10000" class="input-field" />
                  </div>
                  <div>
                    <label for="tmpl-engine" class="form-label">{{ $t('taskTemplatesPage.engine') }}</label>
                    <select id="tmpl-engine" v-model="form.engine" required class="input-field">
                      <option value="script">{{ $t('taskTemplatesPage.engineScript') }}</option>
                      <option value="iac">{{ $t('taskTemplatesPage.engineIac') }}</option>
                      <option value="ansible">{{ $t('taskTemplatesPage.engineAnsible') }}</option>
                    </select>
                  </div>
                  <div>
                    <label for="tmpl-repo" class="form-label">{{ $t('taskTemplatesPage.repository') }}</label>
                    <select id="tmpl-repo" v-model="form.repositoryId" required class="input-field">
                      <option disabled value="">{{ $t('taskTemplatesPage.selectRepository') }}</option>
                      <option v-for="r in repositories" :key="r.id" :value="String(r.id)">{{ r.name }}</option>
                    </select>
                  </div>
                  <div>
                    <label for="tmpl-wd" class="form-label">{{ $t('taskTemplatesPage.workingDirectory') }}</label>
                    <input id="tmpl-wd" v-model="form.workingDirectory" type="text" maxlength="500" class="input-field" />
                  </div>
                  <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label for="tmpl-branch" class="form-label">{{ $t('taskTemplatesPage.branchOverride') }}</label>
                      <input id="tmpl-branch" v-model="form.branchOverride" type="text" maxlength="100" class="input-field" />
                    </div>
                    <div>
                      <label for="tmpl-tag" class="form-label">{{ $t('taskTemplatesPage.tagOverride') }}</label>
                      <input id="tmpl-tag" v-model="form.tagOverride" type="text" maxlength="100" class="input-field" />
                    </div>
                  </div>
                  <div>
                    <label for="tmpl-gv" class="form-label">{{ $t('taskTemplatesPage.groupVars') }}</label>
                    <select id="tmpl-gv" v-model="form.groupVarsId" class="input-field">
                      <option value="">{{ $t('taskTemplatesPage.groupVarsNone') }}</option>
                      <option v-for="g in groupVarsList" :key="g.id" :value="String(g.id)">{{ g.name }}</option>
                    </select>
                  </div>
                  <div class="flex flex-wrap gap-6 items-center">
                    <label class="inline-flex items-center gap-2 text-sm text-theme-text">
                      <input v-model="form.allowConcurrentRuns" type="checkbox" class="rounded border-theme-border" />
                      {{ $t('taskTemplatesPage.allowConcurrentRuns') }}
                    </label>
                    <label class="inline-flex items-center gap-2 text-sm text-theme-text">
                      <input v-model="form.approvalRequired" type="checkbox" class="rounded border-theme-border" />
                      {{ $t('taskTemplatesPage.approvalRequired') }}
                    </label>
                  </div>
                  <div>
                    <label for="tmpl-max-conc" class="form-label">{{ $t('taskTemplatesPage.maxConcurrentRuns') }}</label>
                    <input id="tmpl-max-conc" v-model.number="form.maxConcurrentRuns" type="number" min="0" class="input-field max-w-xs" />
                  </div>
                  <div>
                    <label for="tmpl-extra" class="form-label">{{ $t('taskTemplatesPage.extraArgs') }}</label>
                    <input id="tmpl-extra" v-model="form.extraArgs" type="text" maxlength="10000" class="input-field" />
                  </div>
                </div>

                <div v-show="modalTab === 'tags'" class="space-y-3">
                  <span class="block text-sm font-medium text-theme-text">{{ $t('runnersPage.tags') }}</span>
                  <p class="text-xs text-theme-textLight">{{ $t('runnersPage.tagsHint') }}</p>
                  <div class="space-y-2">
                    <div class="flex gap-2 text-xs font-medium text-theme-textLight uppercase">
                      <span class="flex-1">{{ $t('runnersPage.tagKey') }}</span>
                      <span class="flex-1">{{ $t('runnersPage.tagValue') }}</span>
                      <span class="w-8 shrink-0" aria-hidden="true" />
                    </div>
                    <div v-for="(row, idx) in runnerTagEntries" :key="idx" class="flex gap-2 items-center">
                      <input v-model="row.key" type="text" class="input-field flex-1 text-sm" :placeholder="$t('runnersPage.tagKeyPlaceholder')" />
                      <input v-model="row.value" type="text" class="input-field flex-1 text-sm" :placeholder="$t('runnersPage.tagValuePlaceholder')" />
                      <button type="button" class="text-red-600 hover:text-red-700 p-1 shrink-0 text-sm" :aria-label="$t('common.delete')" @click="removeTagRow(idx)">
                        ×
                      </button>
                    </div>
                    <button type="button" class="btn-secondary text-sm" @click="addTagRow">{{ $t('taskTemplatesPage.addTagRow') }}</button>
                  </div>
                </div>

                <div v-show="modalTab === 'script'" class="space-y-4">
                  <div>
                    <label for="scr-engine" class="form-label">{{ $t('taskTemplatesPage.scriptEngine') }}</label>
                    <select id="scr-engine" v-model="form.script.engine" class="input-field">
                      <option v-for="e in scriptEngineOptions" :key="e" :value="e">{{ e }}</option>
                    </select>
                  </div>
                  <div>
                    <label for="scr-ver" class="form-label">{{ $t('taskTemplatesPage.scriptEngineVersion') }}</label>
                    <input id="scr-ver" v-model="form.script.engineVersion" type="text" maxlength="100" class="input-field" />
                  </div>
                  <div>
                    <label for="scr-path" class="form-label">{{ $t('taskTemplatesPage.scriptPath') }}</label>
                    <input id="scr-path" v-model="form.script.scriptPath" type="text" maxlength="10000" class="input-field" placeholder="scripts/deploy.sh" />
                  </div>
                </div>

                <div v-show="modalTab === 'iac'" class="space-y-4">
                  <div>
                    <label for="iac-engine" class="form-label">{{ $t('taskTemplatesPage.iacEngine') }}</label>
                    <select id="iac-engine" v-model="form.iac.engine" class="input-field">
                      <option value="terraform">{{ $t('taskTemplatesPage.engineTerraform') }}</option>
                      <option value="opentofu">{{ $t('taskTemplatesPage.engineOpenTofu') }}</option>
                    </select>
                  </div>
                  <div>
                    <label for="iac-ver" class="form-label">{{ $t('taskTemplatesPage.iacEngineVersion') }}</label>
                    <input id="iac-ver" v-model="form.iac.engineVersion" type="text" maxlength="100" class="input-field" />
                  </div>
                  <div>
                    <label class="inline-flex items-center gap-2 text-sm text-theme-text">
                      <input v-model="form.iac.autoApprove" type="checkbox" class="rounded border-theme-border" />
                      {{ $t('taskTemplatesPage.autoApprove') }}
                    </label>
                  </div>
                  <div>
                    <label for="iac-backend" class="form-label">{{ $t('taskTemplatesPage.backend') }}</label>
                    <select id="iac-backend" v-model="form.iac.backendID" class="input-field">
                      <option value="">{{ $t('taskTemplatesPage.backendNone') }}</option>
                      <option v-for="b in backendsList" :key="b.id" :value="String(b.id)">{{ b.name }}</option>
                    </select>
                  </div>
                  <div>
                    <label for="iac-log" class="form-label">{{ $t('taskTemplatesPage.logLevel') }}</label>
                    <select id="iac-log" v-model="form.iac.logLevel" class="input-field">
                      <option value="">{{ $t('taskTemplatesPage.iacLogLevelNone') }}</option>
                      <option v-for="lvl in iacLogLevelOptions" :key="lvl" :value="lvl">{{ lvl }}</option>
                    </select>
                  </div>
                </div>

                <div v-show="modalTab === 'ansible'" class="space-y-4">
                  <div>
                    <label for="ans-ver" class="form-label">{{ $t('taskTemplatesPage.ansibleEngineVersion') }}</label>
                    <input id="ans-ver" v-model="form.ansible.engineVersion" type="text" maxlength="100" class="input-field" />
                  </div>
                  <div>
                    <label for="ans-inv" class="form-label">{{ $t('taskTemplatesPage.inventory') }}</label>
                    <select id="ans-inv" v-model="form.ansible.inventoryID" class="input-field">
                      <option value="">{{ $t('taskTemplatesPage.inventoryNone') }}</option>
                      <option v-for="inv in inventoriesList" :key="inv.id" :value="String(inv.id)">{{ inv.name }}</option>
                    </select>
                  </div>
                  <div>
                    <label for="ans-pb" class="form-label">{{ $t('taskTemplatesPage.playbookPath') }}</label>
                    <input id="ans-pb" v-model="form.ansible.playbookPath" type="text" maxlength="4096" class="input-field" />
                  </div>
                  <div>
                    <label for="ans-vault" class="form-label">{{ $t('taskTemplatesPage.vaultPath') }}</label>
                    <input id="ans-vault" v-model="form.ansible.vaultPath" type="text" maxlength="4096" class="input-field" />
                  </div>
                  <div>
                    <label for="ans-vpw" class="form-label">{{ $t('taskTemplatesPage.vaultPassword') }}</label>
                    <input id="ans-vpw" v-model="form.ansible.vaultPassword" type="password" maxlength="4096" autocomplete="new-password" class="input-field" />
                  </div>
                  <div>
                    <label for="ans-log" class="form-label">{{ $t('taskTemplatesPage.logLevel') }}</label>
                    <select id="ans-log" v-model="form.ansible.logLevel" class="input-field">
                      <option value="">{{ $t('taskTemplatesPage.ansibleLogLevelNone') }}</option>
                      <option v-for="n in ansibleLogLevelOptions" :key="n" :value="String(n)">{{ $t('taskTemplatesPage.ansibleLogLevelN', { n }) }}</option>
                    </select>
                  </div>
                </div>
              </div>

              <div class="px-6 py-4 border-t border-theme-border flex justify-end gap-3 shrink-0 bg-white rounded-b-lg">
                <button type="button" class="btn-secondary" @click="closeEditModal">{{ $t('common.cancel') }}</button>
                <button type="submit" class="btn-primary" :disabled="editSaving">
                  {{ editSaving ? $t('taskTemplatesPage.saving') : $t('common.save') }}
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
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
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

const editModalOpen = ref(false)
const editModalTitleId = 'tmpl-edit-title'
const editingTemplate = ref(null)
const editSaving = ref(false)
const modalError = ref('')
const modalTab = ref('general')

const deleteModalTemplate = ref(null)
const deletingId = ref(null)

const modalTabs = computed(() => [
  { id: 'general', label: t('taskTemplatesPage.tabGeneral') },
  { id: 'tags', label: t('taskTemplatesPage.tabRunnerTags') },
  { id: 'script', label: t('taskTemplatesPage.tabScript') },
  { id: 'iac', label: t('taskTemplatesPage.tabIac') },
  { id: 'ansible', label: t('taskTemplatesPage.tabAnsible') },
])

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
    ansible: { engineVersion: '', inventoryID: '', playbookPath: '', vaultPath: '', vaultPassword: '', logLevel: '' },
  }
}

const form = reactive(defaultFormState())
const runnerTagEntries = ref([{ key: '', value: '' }])

function resetForm() {
  Object.assign(form, defaultFormState())
  runnerTagEntries.value = [{ key: '', value: '' }]
  modalError.value = ''
  modalTab.value = 'general'
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
  form.ansible.vaultPath = pickConfigField(a, 'vaultPath', 'vault_path')
  form.ansible.vaultPassword = pickConfigField(a, 'vaultPassword', 'vault_password')
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
      vaultPath: form.ansible.vaultPath || '',
      vaultPassword: form.ansible.vaultPassword || '',
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

function openCreateModal() {
  editingTemplate.value = null
  resetForm()
  editModalOpen.value = true
}

async function openEditModal(tmpl) {
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
  editModalOpen.value = true
}

function closeEditModal() {
  editModalOpen.value = false
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
    closeEditModal()
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
  setWorkspace()
  if (workspace.value?.id) {
    await loadAuxiliaryLists()
    loadTemplates()
  }
})
</script>
