<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('inventoryPage.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('inventoryPage.subtitle') }}
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" :aria-label="$t('inventoryPage.sectionTitle')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('inventoryPage.sectionTitle') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('inventoryPage.sectionDescription') }}</p>
          </div>
          <button type="button" class="btn-primary text-sm inline-flex items-center gap-2" @click="openCreateModal">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('inventoryPage.addInventory') }}
          </button>
        </div>

        <div v-if="loading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('inventoryPage.loading') }}</div>
        <div v-else-if="inventories.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('inventoryPage.empty') }}</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('inventoryPage.name') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('inventoryPage.type') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('inventoryPage.createdAt') }}</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('inventoryPage.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="inv in inventories" :key="inv.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <p class="text-sm font-medium text-theme-text">{{ inv.name }}</p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ typeLabel(inv.type) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ formatCreatedAt(inv.createdAt, locale) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button type="button" class="text-primary-600 hover:text-primary-700 p-1 rounded mr-1" :title="$t('common.edit')" :aria-label="$t('common.edit')" @click="openEditModal(inv)">
                      <svg class="h-4 w-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </button>
                    <button type="button" class="text-red-600 hover:text-red-700 p-1 rounded" :title="$t('common.delete')" :aria-label="$t('common.delete')" @click="openDeleteModal(inv)">
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
                {{ $t('inventoryPage.showing', { from: showingFrom, to: showingTo, total: totalInventories }) }}
              </template>
              <template v-else-if="_meta?.total">
                {{ _meta.total === 1 ? $t('inventoryPage.countOne') : $t('inventoryPage.countMany', { count: _meta.total }) }}
              </template>
            </div>
            <div v-if="hasPagination" class="flex items-center gap-2">
              <button type="button" class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed" :disabled="offset === 0" @click="goToPage(offset - limit)">
                {{ $t('inventoryPage.previous') }}
              </button>
              <button type="button" class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed" :disabled="offset + limit >= totalInventories" @click="goToPage(offset + limit)">
                {{ $t('inventoryPage.next') }}
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
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-2xl max-h-[92vh] flex flex-col">
            <div class="px-6 py-4 border-b border-theme-border flex items-start justify-between gap-3 shrink-0">
              <h2 :id="editModalTitleId" class="text-lg font-semibold text-theme-text pr-8">
                {{ editingInventory ? $t('inventoryPage.editInventory') : $t('inventoryPage.createInventory') }}
              </h2>
              <button type="button" class="text-theme-textLight hover:text-theme-text p-1 rounded shrink-0 -mr-1" :aria-label="$t('common.close')" @click="closeEditModal">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <form class="flex flex-col flex-1 min-h-0" @submit.prevent="handleSave">
              <div class="px-6 py-4 overflow-y-auto flex-1 space-y-4">
                <div v-if="modalError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
                  {{ modalError }}
                </div>

                <div>
                  <label for="inv-name" class="form-label">{{ $t('inventoryPage.name') }}</label>
                  <input id="inv-name" v-model="form.name" type="text" required maxlength="100" class="input-field" />
                </div>

                <div>
                  <label for="inv-type" class="form-label">{{ $t('inventoryPage.type') }}</label>
                  <select id="inv-type" v-model="form.type" class="input-field" @change="onTypeChange">
                    <option value="static">{{ $t('inventoryPage.typeStatic') }}</option>
                    <option value="static_yaml">{{ $t('inventoryPage.typeStaticYaml') }}</option>
                    <option value="file">{{ $t('inventoryPage.typeFile') }}</option>
                  </select>
                </div>

                <div v-if="form.type === 'file'">
                  <label for="inv-path" class="form-label">{{ $t('inventoryPage.inventoryPath') }}</label>
                  <input id="inv-path" v-model="form.inventoryPath" type="text" maxlength="512" class="input-field" :placeholder="$t('inventoryPage.inventoryPathPlaceholder')" />
                  <p class="mt-1 text-xs text-theme-textLight">{{ $t('inventoryPage.inventoryPathHelp') }}</p>
                </div>

                <div v-else>
                  <label for="inv-content" class="form-label">{{ $t('inventoryPage.inventoryContent') }}</label>
                  <textarea
                    id="inv-content"
                    v-model="form.inventory"
                    rows="12"
                    class="input-field font-mono text-sm"
                    :placeholder="form.type === 'static_yaml' ? $t('inventoryPage.yamlPlaceholder') : $t('inventoryPage.iniPlaceholder')"
                  />
                </div>

                <div class="border-t border-theme-border pt-4 space-y-3">
                  <p class="text-sm font-medium text-theme-text">{{ $t('inventoryPage.optionalCredentials') }}</p>

                  <div>
                    <label for="inv-repo" class="form-label">{{ $t('inventoryPage.repository') }}</label>
                    <select id="inv-repo" v-model="form.repositoryId" class="input-field">
                      <option value="">{{ $t('inventoryPage.optionalNone') }}</option>
                      <option v-for="r in repositories" :key="r.id" :value="String(r.id)">{{ r.name }}</option>
                    </select>
                  </div>

                  <div>
                    <label for="inv-ssh" class="form-label">{{ $t('inventoryPage.sshKey') }}</label>
                    <select id="inv-ssh" v-model="form.sshKeyId" class="input-field">
                      <option value="">{{ $t('inventoryPage.optionalNone') }}</option>
                      <option v-for="k in sshKeys" :key="k.id" :value="String(k.id)">{{ k.name }}</option>
                    </select>
                  </div>

                  <div>
                    <label for="inv-cred" class="form-label">{{ $t('inventoryPage.credential') }}</label>
                    <select id="inv-cred" v-model="form.credentialId" class="input-field">
                      <option value="">{{ $t('inventoryPage.optionalNone') }}</option>
                      <option v-for="c in credentials" :key="c.id" :value="String(c.id)">{{ c.name }}</option>
                    </select>
                  </div>

                  <div>
                    <label for="inv-become-key" class="form-label">{{ $t('inventoryPage.becomeKey') }}</label>
                    <select id="inv-become-key" v-model="form.becomeKeyId" class="input-field">
                      <option value="">{{ $t('inventoryPage.optionalNone') }}</option>
                      <option v-for="k in sshKeys" :key="'b-' + k.id" :value="String(k.id)">{{ k.name }}</option>
                    </select>
                  </div>

                  <div>
                    <label for="inv-become-cred" class="form-label">{{ $t('inventoryPage.becomeCredential') }}</label>
                    <select id="inv-become-cred" v-model="form.becomeCredentialId" class="input-field">
                      <option value="">{{ $t('inventoryPage.optionalNone') }}</option>
                      <option v-for="c in credentials" :key="'bc-' + c.id" :value="String(c.id)">{{ c.name }}</option>
                    </select>
                  </div>
                </div>
              </div>

              <div class="px-6 py-4 border-t border-theme-border flex justify-end gap-3 shrink-0 bg-white rounded-b-lg">
                <button type="button" class="btn-secondary" @click="closeEditModal">{{ $t('common.cancel') }}</button>
                <button type="submit" class="btn-primary" :disabled="saveLoading">
                  {{ saveLoading ? $t('inventoryPage.saving') : $t('common.save') }}
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
          v-show="deleteTarget"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-inv-title"
          @click.self="deleteTarget = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-inv-title" class="text-lg font-semibold text-theme-text">{{ $t('inventoryPage.deleteTitle') }}</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('inventoryPage.deleteConfirm', { name: deleteTarget?.name }) }}
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" class="btn-secondary" @click="deleteTarget = null">{{ $t('common.cancel') }}</button>
              <button type="button" class="btn-danger" :disabled="deletingId === deleteTarget?.id" @click="confirmDelete">
                {{ deletingId === deleteTarget?.id ? $t('inventoryPage.deleting') : $t('common.delete') }}
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
import { formatCreatedAt } from '@/utils/date'
import { inventoriesAPI, sshKeysAPI, credentialsAPI, repositoriesAPI } from '@/api'

const { t, locale } = useI18n()
const router = useRouter()

const workspace = ref(null)
const inventories = ref([])
const _meta = ref(null)
const loading = ref(true)
const limit = ref(50)
const offset = ref(0)
const errorMessage = ref('')
const successMessage = ref('')

const editModalOpen = ref(false)
const editModalTitleId = 'inv-edit-title'
const editingInventory = ref(null)
const saveLoading = ref(false)
const modalError = ref('')

const sshKeys = ref([])
const credentials = ref([])
const repositories = ref([])

const deleteTarget = ref(null)
const deletingId = ref(null)

const defaultForm = () => ({
  name: '',
  type: 'static',
  inventory: '',
  inventoryPath: '',
  sshKeyId: '',
  credentialId: '',
  becomeKeyId: '',
  becomeCredentialId: '',
  repositoryId: '',
})

const form = ref(defaultForm())

function idFromApi(v) {
  if (v == null || v === '') return ''
  return String(v)
}

function typeLabel(type) {
  if (type === 'static') return t('inventoryPage.typeStatic')
  if (type === 'static_yaml') return t('inventoryPage.typeStaticYaml')
  if (type === 'file') return t('inventoryPage.typeFile')
  return type || '—'
}

function onTypeChange() {
  if (form.value.type === 'file') {
    form.value.inventory = ''
  } else {
    form.value.inventoryPath = ''
  }
}

function resetForm() {
  form.value = defaultForm()
  modalError.value = ''
}

function fillForm(inv) {
  form.value = {
    name: inv.name || '',
    type: inv.type || 'static',
    inventory: inv.inventory || '',
    inventoryPath: inv.inventoryPath || '',
    sshKeyId: idFromApi(inv.sshKeyId),
    credentialId: idFromApi(inv.credentialId),
    becomeKeyId: idFromApi(inv.becomeKeyId),
    becomeCredentialId: idFromApi(inv.becomeCredentialId),
    repositoryId: idFromApi(inv.repositoryId),
  }
}

async function loadRefs() {
  if (!workspace.value?.id) return
  const wid = workspace.value.id
  try {
    const [keysRes, credRes, repoRes] = await Promise.all([
      sshKeysAPI.list(wid, { limit: 100, offset: 0 }),
      credentialsAPI.list(wid, { limit: 100, offset: 0 }),
      repositoriesAPI.list(wid, { limit: 100, offset: 0 }),
    ])
    sshKeys.value = keysRes.data?.keys ?? []
    credentials.value = credRes.data?.credentials ?? []
    repositories.value = repoRes.data?.repositories ?? []
  } catch {
    sshKeys.value = []
    credentials.value = []
    repositories.value = []
  }
}

function setWorkspace() {
  workspace.value = loadWorkspaceFromStorage()
  if (!workspace.value?.id) {
    router.replace('/select-workspace')
  }
}

async function loadInventories(overrideOffset) {
  if (!workspace.value?.id) return
  loading.value = true
  errorMessage.value = ''
  const off = overrideOffset !== undefined ? overrideOffset : offset.value

  try {
    const res = await inventoriesAPI.list(workspace.value.id, { limit: limit.value, offset: off })
    inventories.value = res.data?.inventories ?? []
    _meta.value = res.data?._meta ?? null
    offset.value = off
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('inventoryPage.failedLoad')
    inventories.value = []
  } finally {
    loading.value = false
  }
}

const totalInventories = computed(() => _meta.value?.total ?? 0)
const showingFrom = computed(() => (totalInventories.value === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + inventories.value.length, totalInventories.value))
const hasPagination = computed(() => totalInventories.value > limit.value)

function goToPage(newOffset) {
  const maxOffset = Math.max(0, totalInventories.value - limit.value)
  const off = Math.max(0, Math.min(newOffset, maxOffset))
  loadInventories(off)
}

function buildPayload() {
  const f = form.value
  return {
    name: f.name.trim(),
    type: f.type,
    inventory: f.type === 'file' ? '' : f.inventory,
    inventoryPath: f.type === 'file' ? f.inventoryPath.trim() : '',
    sshKeyId: f.sshKeyId || '',
    credentialId: f.credentialId || '',
    becomeKeyId: f.becomeKeyId || '',
    becomeCredentialId: f.becomeCredentialId || '',
    repositoryId: f.repositoryId || '',
  }
}

async function openCreateModal() {
  editingInventory.value = null
  resetForm()
  await loadRefs()
  editModalOpen.value = true
}

async function openEditModal(inv) {
  editingInventory.value = inv
  resetForm()
  await loadRefs()
  fillForm(inv)
  editModalOpen.value = true
}

function closeEditModal() {
  editModalOpen.value = false
  editingInventory.value = null
  resetForm()
}

async function handleSave() {
  if (!workspace.value?.id) return
  modalError.value = ''
  const payload = buildPayload()

  if (payload.type === 'file' && !payload.inventoryPath) {
    modalError.value = t('inventoryPage.pathRequired')
    return
  }
  if (payload.type !== 'file' && !payload.inventory?.trim()) {
    modalError.value = t('inventoryPage.contentRequired')
    return
  }

  saveLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    if (editingInventory.value?.id) {
      await inventoriesAPI.update(workspace.value.id, editingInventory.value.id, payload)
      successMessage.value = t('inventoryPage.inventoryUpdated')
    } else {
      await inventoriesAPI.create(workspace.value.id, payload)
      successMessage.value = t('inventoryPage.inventoryCreated')
    }
    closeEditModal()
    await loadInventories()
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) {
    const msg = err.response?.data?.errorMessage || err.response?.data?.message
    if (msg) {
      modalError.value = msg
    } else {
      errorMessage.value = t('inventoryPage.failedSave')
    }
  } finally {
    saveLoading.value = false
  }
}

function openDeleteModal(inv) {
  deleteTarget.value = inv
}

async function confirmDelete() {
  const inv = deleteTarget.value
  if (!workspace.value?.id || !inv?.id) return
  deletingId.value = inv.id
  errorMessage.value = ''

  try {
    await inventoriesAPI.delete(workspace.value.id, inv.id)
    successMessage.value = t('inventoryPage.inventoryDeleted')
    deleteTarget.value = null
    await loadInventories()
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || t('inventoryPage.failedDelete')
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  setWorkspace()
  if (workspace.value?.id) {
    loadInventories()
  }
})

watch(workspace, (w) => {
  if (w?.id) loadInventories()
}, { immediate: false })
</script>
