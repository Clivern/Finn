<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="emitClose"
      >
        <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-3xl max-h-[92vh] flex flex-col">
          <div class="px-6 py-4 border-b border-theme-border flex items-start justify-between gap-3 shrink-0">
            <h2 :id="titleId" class="text-lg font-semibold text-theme-text pr-8">
              {{ modalHeading }}
            </h2>
            <button type="button" class="text-theme-textLight hover:text-theme-text p-1 rounded shrink-0 -mr-1" :aria-label="$t('common.close')" @click="emitClose">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <form class="flex flex-col flex-1 min-h-0" @submit.prevent="$emit('save')">
            <div class="px-6 py-4 overflow-y-auto flex-1 space-y-8">
              <div v-if="modalError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
                {{ modalError }}
              </div>

              <section class="space-y-4" :aria-label="$t('taskTemplatesPage.tabGeneral')">
                <h3 class="text-sm font-semibold text-theme-text">{{ $t('taskTemplatesPage.tabGeneral') }}</h3>
                <div>
                  <label :for="`${idPrefix}-name`" class="form-label">{{ $t('taskTemplatesPage.name') }}</label>
                  <input :id="`${idPrefix}-name`" v-model="form.name" type="text" required maxlength="100" class="input-field" />
                </div>
                <div>
                  <label :for="`${idPrefix}-desc`" class="form-label">{{ $t('taskTemplatesPage.description') }}</label>
                  <textarea :id="`${idPrefix}-desc`" v-model="form.description" rows="2" required maxlength="10000" class="input-field" />
                </div>
                <p class="text-sm text-theme-textLight">
                  <span class="font-medium text-theme-text">{{ $t('taskTemplatesPage.engine') }}:</span>
                  {{ engineLabel }}
                </p>
                <div>
                  <label :for="`${idPrefix}-repo`" class="form-label">{{ $t('taskTemplatesPage.repository') }}</label>
                  <select :id="`${idPrefix}-repo`" v-model="form.repositoryId" required class="input-field">
                    <option disabled value="">{{ $t('taskTemplatesPage.selectRepository') }}</option>
                    <option v-for="r in repositories" :key="r.id" :value="String(r.id)">{{ r.name }}</option>
                  </select>
                </div>
                <div>
                  <label :for="`${idPrefix}-wd`" class="form-label">{{ $t('taskTemplatesPage.workingDirectory') }}</label>
                  <input :id="`${idPrefix}-wd`" v-model="form.workingDirectory" type="text" maxlength="500" class="input-field" />
                </div>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label :for="`${idPrefix}-branch`" class="form-label">{{ $t('taskTemplatesPage.branchOverride') }}</label>
                    <input :id="`${idPrefix}-branch`" v-model="form.branchOverride" type="text" maxlength="100" class="input-field" />
                  </div>
                  <div>
                    <label :for="`${idPrefix}-tagOv`" class="form-label">{{ $t('taskTemplatesPage.tagOverride') }}</label>
                    <input :id="`${idPrefix}-tagOv`" v-model="form.tagOverride" type="text" maxlength="100" class="input-field" />
                  </div>
                </div>
                <div>
                  <label :for="`${idPrefix}-gv`" class="form-label">{{ $t('taskTemplatesPage.groupVars') }}</label>
                  <select :id="`${idPrefix}-gv`" v-model="form.groupVarsId" class="input-field">
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
                  <label :for="`${idPrefix}-maxConc`" class="form-label">{{ $t('taskTemplatesPage.maxConcurrentRuns') }}</label>
                  <input :id="`${idPrefix}-maxConc`" v-model.number="form.maxConcurrentRuns" type="number" min="0" class="input-field max-w-xs" />
                </div>
                <div>
                  <label :for="`${idPrefix}-extra`" class="form-label">{{ $t('taskTemplatesPage.extraArgs') }}</label>
                  <input :id="`${idPrefix}-extra`" v-model="form.extraArgs" type="text" maxlength="10000" class="input-field" />
                </div>
              </section>

              <section class="space-y-3" :aria-label="$t('taskTemplatesPage.tabRunnerTags')">
                <h3 class="text-sm font-semibold text-theme-text">{{ $t('taskTemplatesPage.tabRunnerTags') }}</h3>
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
                    <button type="button" class="text-red-600 hover:text-red-700 p-1 shrink-0 text-sm" :aria-label="$t('common.delete')" @click="emitRemoveRunnerTag(idx)">
                      ×
                    </button>
                  </div>
                  <button type="button" class="btn-secondary text-sm" @click="emitAddRunnerTag">{{ $t('taskTemplatesPage.addTagRow') }}</button>
                </div>
              </section>

              <section v-if="engine === 'script'" class="space-y-4" :aria-label="$t('taskTemplatesPage.tabScript')">
                <h3 class="text-sm font-semibold text-theme-text">{{ $t('taskTemplatesPage.tabScript') }}</h3>
                <div>
                  <label :for="`${idPrefix}-scr-engine`" class="form-label">{{ $t('taskTemplatesPage.scriptEngine') }}</label>
                  <select :id="`${idPrefix}-scr-engine`" v-model="form.script.engine" class="input-field">
                    <option v-for="e in scriptEngineOptions" :key="e" :value="e">{{ e }}</option>
                  </select>
                </div>
                <div>
                  <label :for="`${idPrefix}-scr-ver`" class="form-label">{{ $t('taskTemplatesPage.scriptEngineVersion') }}</label>
                  <input :id="`${idPrefix}-scr-ver`" v-model="form.script.engineVersion" type="text" maxlength="100" class="input-field" />
                </div>
                <div>
                  <label :for="`${idPrefix}-scr-path`" class="form-label">{{ $t('taskTemplatesPage.scriptPath') }}</label>
                  <input :id="`${idPrefix}-scr-path`" v-model="form.script.scriptPath" type="text" maxlength="10000" class="input-field" placeholder="scripts/deploy.sh" />
                </div>
              </section>

              <section v-else-if="engine === 'iac'" class="space-y-4" :aria-label="$t('taskTemplatesPage.tabIac')">
                <h3 class="text-sm font-semibold text-theme-text">{{ $t('taskTemplatesPage.tabIac') }}</h3>
                <div>
                  <label :for="`${idPrefix}-iac-engine`" class="form-label">{{ $t('taskTemplatesPage.iacEngine') }}</label>
                  <select :id="`${idPrefix}-iac-engine`" v-model="form.iac.engine" class="input-field">
                    <option value="terraform">{{ $t('taskTemplatesPage.engineTerraform') }}</option>
                    <option value="opentofu">{{ $t('taskTemplatesPage.engineOpenTofu') }}</option>
                  </select>
                </div>
                <div>
                  <label :for="`${idPrefix}-iac-ver`" class="form-label">{{ $t('taskTemplatesPage.iacEngineVersion') }}</label>
                  <input :id="`${idPrefix}-iac-ver`" v-model="form.iac.engineVersion" type="text" maxlength="100" class="input-field" />
                </div>
                <div>
                  <label class="inline-flex items-center gap-2 text-sm text-theme-text">
                    <input v-model="form.iac.autoApprove" type="checkbox" class="rounded border-theme-border" />
                    {{ $t('taskTemplatesPage.autoApprove') }}
                  </label>
                </div>
                <div>
                  <label :for="`${idPrefix}-iac-backend`" class="form-label">{{ $t('taskTemplatesPage.backend') }}</label>
                  <select :id="`${idPrefix}-iac-backend`" v-model="form.iac.backendID" class="input-field">
                    <option value="">{{ $t('taskTemplatesPage.backendNone') }}</option>
                    <option v-for="b in backendsList" :key="b.id" :value="String(b.id)">{{ b.name }}</option>
                  </select>
                </div>
                <div>
                  <label :for="`${idPrefix}-iac-log`" class="form-label">{{ $t('taskTemplatesPage.logLevel') }}</label>
                  <select :id="`${idPrefix}-iac-log`" v-model="form.iac.logLevel" class="input-field">
                    <option value="">{{ $t('taskTemplatesPage.iacLogLevelNone') }}</option>
                    <option v-for="lvl in iacLogLevelOptions" :key="lvl" :value="lvl">{{ lvl }}</option>
                  </select>
                </div>
              </section>

              <section v-else-if="engine === 'ansible'" class="space-y-4" :aria-label="$t('taskTemplatesPage.tabAnsible')">
                <h3 class="text-sm font-semibold text-theme-text">{{ $t('taskTemplatesPage.tabAnsible') }}</h3>
                <div>
                  <label :for="`${idPrefix}-ans-ver`" class="form-label">{{ $t('taskTemplatesPage.ansibleEngineVersion') }}</label>
                  <input :id="`${idPrefix}-ans-ver`" v-model="form.ansible.engineVersion" type="text" maxlength="100" class="input-field" />
                </div>
                <div>
                  <label :for="`${idPrefix}-ans-inv`" class="form-label">{{ $t('taskTemplatesPage.inventory') }}</label>
                  <select :id="`${idPrefix}-ans-inv`" v-model="form.ansible.inventoryID" class="input-field">
                    <option value="">{{ $t('taskTemplatesPage.inventoryNone') }}</option>
                    <option v-for="inv in inventoriesList" :key="inv.id" :value="String(inv.id)">{{ inv.name }}</option>
                  </select>
                </div>
                <div>
                  <label :for="`${idPrefix}-ans-pb`" class="form-label">{{ $t('taskTemplatesPage.playbookPath') }}</label>
                  <input :id="`${idPrefix}-ans-pb`" v-model="form.ansible.playbookPath" type="text" maxlength="4096" class="input-field" />
                </div>
                <div>
                  <label :for="`${idPrefix}-ans-log`" class="form-label">{{ $t('taskTemplatesPage.logLevel') }}</label>
                  <select :id="`${idPrefix}-ans-log`" v-model="form.ansible.logLevel" class="input-field">
                    <option value="">{{ $t('taskTemplatesPage.ansibleLogLevelNone') }}</option>
                    <option v-for="n in ansibleLogLevelOptions" :key="n" :value="String(n)">{{ $t('taskTemplatesPage.ansibleLogLevelN', { n }) }}</option>
                  </select>
                </div>
              </section>
            </div>

            <div class="px-6 py-4 border-t border-theme-border flex justify-end gap-3 shrink-0 bg-white rounded-b-lg">
              <button type="button" class="btn-secondary" @click="emitClose">{{ $t('common.cancel') }}</button>
              <button type="submit" class="btn-primary" :disabled="editSaving">
                {{ editSaving ? $t('taskTemplatesPage.saving') : $t('common.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  engine: {
    type: String,
    required: true,
    validator: (v) => v === 'script' || v === 'iac' || v === 'ansible',
  },
  editing: { type: Boolean, default: false },
  titleId: { type: String, required: true },
  form: { type: Object, required: true },
  runnerTagEntries: { type: Array, required: true },
  modalError: { type: String, default: '' },
  editSaving: { type: Boolean, default: false },
  repositories: { type: Array, default: () => [] },
  groupVarsList: { type: Array, default: () => [] },
  backendsList: { type: Array, default: () => [] },
  inventoriesList: { type: Array, default: () => [] },
  scriptEngineOptions: { type: Array, default: () => [] },
  iacLogLevelOptions: { type: Array, default: () => [] },
  ansibleLogLevelOptions: { type: Array, default: () => [] },
  addRunnerTagRow: { type: Function, required: true },
  removeRunnerTagRow: { type: Function, required: true },
})

const emit = defineEmits(['close', 'save'])

const { t } = useI18n()

const idPrefix = computed(() => props.engine)

const engineLabel = computed(() => {
  if (props.engine === 'script') return t('taskTemplatesPage.engineScript')
  if (props.engine === 'iac') return t('taskTemplatesPage.engineIac')
  return t('taskTemplatesPage.engineAnsible')
})

const TITLE_KEYS = {
  script: { create: 'taskTemplatesPage.createScriptTemplate', edit: 'taskTemplatesPage.editScriptTemplate' },
  iac: { create: 'taskTemplatesPage.createIacTemplate', edit: 'taskTemplatesPage.editIacTemplate' },
  ansible: { create: 'taskTemplatesPage.createAnsibleTemplate', edit: 'taskTemplatesPage.editAnsibleTemplate' },
}

const modalHeading = computed(() => {
  const bucket = TITLE_KEYS[props.engine]
  const key = props.editing ? bucket.edit : bucket.create
  return t(key)
})

function emitClose() {
  emit('close')
}

function emitAddRunnerTag() {
  props.addRunnerTagRow()
}

function emitRemoveRunnerTag(idx) {
  props.removeRunnerTagRow(idx)
}
</script>
