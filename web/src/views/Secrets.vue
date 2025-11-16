<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          Secrets
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          Manage SSH keys and Git credentials for this workspace
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <!-- SSH Keys section -->
      <section class="mb-10 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" aria-label="SSH keys">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">SSH keys</h2>
            <p class="text-sm text-theme-textLight mt-0.5">Add, edit, or remove SSH keys for cloning and pushing over SSH</p>
          </div>
          <button
            type="button"
            @click="openAddKeyModal"
            class="btn-primary text-sm inline-flex items-center gap-2"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Add SSH key
          </button>
        </div>
        <div v-if="keysLoading" class="p-8 text-center text-theme-textLight text-sm">Loading SSH keys…</div>
        <div v-else-if="sshKeys.length === 0" class="p-8 text-center text-theme-textLight text-sm">No SSH keys yet. Add one above.</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Name</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Committer</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="k in sshKeys" :key="k.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <p class="text-sm font-medium text-theme-text">{{ k.name }}</p>
                  </td>
                  <td class="px-6 py-4">
                    <p class="text-sm text-theme-textLight">{{ k.committerName }} &lt;{{ k.committerEmail }}&gt;</p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button type="button" @click="openEditKeyModal(k)" class="text-primary-600 hover:text-primary-700 hover:underline mr-3">Edit</button>
                    <button type="button" @click="openDeleteKeyModal(k)" class="text-red-600 hover:text-red-700 hover:underline">Delete</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </section>

      <!-- Git credentials section -->
      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" aria-label="Git credentials">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">Git credentials</h2>
            <p class="text-sm text-theme-textLight mt-0.5">Add, edit, or remove username/password or token credentials for HTTPS Git access</p>
          </div>
          <button
            type="button"
            @click="openAddCredModal"
            class="btn-primary text-sm inline-flex items-center gap-2"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Add Git credentials
          </button>
        </div>
        <div v-if="credsLoading" class="p-8 text-center text-theme-textLight text-sm">Loading credentials…</div>
        <div v-else-if="credentials.length === 0" class="p-8 text-center text-theme-textLight text-sm">No Git credentials yet. Add one above.</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Name</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Committer / Username</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="c in credentials" :key="c.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <p class="text-sm font-medium text-theme-text">{{ c.name }}</p>
                  </td>
                  <td class="px-6 py-4">
                    <p class="text-sm text-theme-textLight">
                      {{ c.committerName }} &lt;{{ c.committerEmail }}&gt;
                      <span v-if="c.username" class="block text-xs mt-0.5">User: {{ c.username }}</span>
                    </p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button type="button" @click="openEditCredModal(c)" class="text-primary-600 hover:text-primary-700 hover:underline mr-3">Edit</button>
                    <button type="button" @click="openDeleteCredModal(c)" class="text-red-600 hover:text-red-700 hover:underline">Delete</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </section>
    </main>

    <!-- SSH key Add/Edit modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="keyFormModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="editingKey ? 'edit-key-title' : 'add-key-title'"
          @click.self="keyFormModalOpen = false"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 :id="editingKey ? 'edit-key-title' : 'add-key-title'" class="text-lg font-semibold text-theme-text">
              {{ editingKey ? 'Edit SSH key' : 'Add SSH key' }}
            </h2>
            <form @submit.prevent="saveKey" class="mt-4 space-y-4">
              <div>
                <label for="key-name" class="form-label">Name</label>
                <input id="key-name" v-model="keyForm.name" type="text" required maxlength="100" class="input-field" placeholder="e.g. GitHub deploy key" />
              </div>
              <div>
                <label for="key-committerName" class="form-label">Committer name</label>
                <input id="key-committerName" v-model="keyForm.committerName" type="text" required maxlength="100" class="input-field" placeholder="Your Name" />
              </div>
              <div>
                <label for="key-committerEmail" class="form-label">Committer email</label>
                <input id="key-committerEmail" v-model="keyForm.committerEmail" type="email" required maxlength="100" class="input-field" placeholder="you@example.com" />
              </div>
              <div>
                <label for="key-privateKey" class="form-label">Private key {{ editingKey ? '(leave blank to keep current)' : '' }}</label>
                <textarea
                  id="key-privateKey"
                  v-model="keyForm.privateKey"
                  :required="!editingKey"
                  rows="6"
                  class="input-field font-mono text-sm"
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                />
              </div>
              <div>
                <label for="key-publicKey" class="form-label">Public key {{ editingKey ? '(optional)' : '(optional, e.g. ssh-rsa AAAA...)' }}</label>
                <input id="key-publicKey" v-model="keyForm.publicKey" type="text" class="input-field font-mono text-sm" placeholder="ssh-rsa AAAA..." />
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="keyFormModalOpen = false" class="btn-secondary">Cancel</button>
                <button type="submit" :disabled="keyFormSaving" class="btn-primary">
                  {{ keyFormSaving ? 'Saving…' : (editingKey ? 'Save' : 'Add') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- SSH key Delete modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteKeyModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-key-title"
          @click.self="deleteKeyModal = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-key-title" class="text-lg font-semibold text-theme-text">Delete SSH key</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              Delete SSH key <span class="font-medium text-theme-text">"{{ deleteKeyModal?.name }}"</span>? This cannot be undone.
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" @click="deleteKeyModal = null" class="btn-secondary">Cancel</button>
              <button type="button" @click="confirmDeleteKey" :disabled="deletingKeyId === deleteKeyModal?.id" class="btn-danger">
                {{ deletingKeyId === deleteKeyModal?.id ? 'Deleting…' : 'Delete' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Git credential Add/Edit modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="credFormModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="editingCred ? 'edit-cred-title' : 'add-cred-title'"
          @click.self="credFormModalOpen = false"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 :id="editingCred ? 'edit-cred-title' : 'add-cred-title'" class="text-lg font-semibold text-theme-text">
              {{ editingCred ? 'Edit Git credentials' : 'Add Git credentials' }}
            </h2>
            <form @submit.prevent="saveCred" class="mt-4 space-y-4">
              <div>
                <label for="cred-name" class="form-label">Name</label>
                <input id="cred-name" v-model="credForm.name" type="text" required maxlength="100" class="input-field" placeholder="e.g. GitHub token" />
              </div>
              <div>
                <label for="cred-committerName" class="form-label">Committer name</label>
                <input id="cred-committerName" v-model="credForm.committerName" type="text" required maxlength="100" class="input-field" placeholder="Your Name" />
              </div>
              <div>
                <label for="cred-committerEmail" class="form-label">Committer email</label>
                <input id="cred-committerEmail" v-model="credForm.committerEmail" type="email" required maxlength="100" class="input-field" placeholder="you@example.com" />
              </div>
              <div>
                <label for="cred-username" class="form-label">Username (optional)</label>
                <input id="cred-username" v-model="credForm.username" type="text" maxlength="100" class="input-field" placeholder="git or your username" />
              </div>
              <div>
                <label for="cred-passwordOrToken" class="form-label">Password or token {{ editingCred ? '(leave blank to keep current)' : '' }}</label>
                <input
                  id="cred-passwordOrToken"
                  v-model="credForm.passwordOrToken"
                  :required="!editingCred"
                  type="password"
                  autocomplete="off"
                  class="input-field"
                  placeholder="••••••••"
                />
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="credFormModalOpen = false" class="btn-secondary">Cancel</button>
                <button type="submit" :disabled="credFormSaving" class="btn-primary">
                  {{ credFormSaving ? 'Saving…' : (editingCred ? 'Save' : 'Add') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Git credential Delete modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteCredModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-cred-title"
          @click.self="deleteCredModal = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-cred-title" class="text-lg font-semibold text-theme-text">Delete Git credentials</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              Delete credentials <span class="font-medium text-theme-text">"{{ deleteCredModal?.name }}"</span>? This cannot be undone.
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" @click="deleteCredModal = null" class="btn-secondary">Cancel</button>
              <button type="button" @click="confirmDeleteCred" :disabled="deletingCredId === deleteCredModal?.id" class="btn-danger">
                {{ deletingCredId === deleteCredModal?.id ? 'Deleting…' : 'Delete' }}
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
import { sshKeysAPI, credentialsAPI } from '@/api'

const router = useRouter()
const workspace = ref(null)
const sshKeys = ref([])
const credentials = ref([])
const keysLoading = ref(true)
const credsLoading = ref(true)
const errorMessage = ref('')
const successMessage = ref('')

const keyFormModalOpen = ref(false)
const keyFormSaving = ref(false)
const editingKey = ref(null)
const keyForm = ref({
  name: '',
  committerName: '',
  committerEmail: '',
  privateKey: '',
  publicKey: ''
})
const deleteKeyModal = ref(null)
const deletingKeyId = ref(null)

const credFormModalOpen = ref(false)
const credFormSaving = ref(false)
const editingCred = ref(null)
const credForm = ref({
  name: '',
  committerName: '',
  committerEmail: '',
  username: '',
  passwordOrToken: ''
})
const deleteCredModal = ref(null)
const deletingCredId = ref(null)

function setWorkspace() {
  workspace.value = loadWorkspaceFromStorage()
  if (!workspace.value?.id) {
    router.replace('/select-workspace')
  }
}

async function loadSshKeys() {
  if (!workspace.value?.id) return
  keysLoading.value = true
  try {
    const res = await sshKeysAPI.list(workspace.value.id, { limit: 100 })
    sshKeys.value = res.data?.keys ?? []
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load SSH keys'
    sshKeys.value = []
  } finally {
    keysLoading.value = false
  }
}

async function loadCredentials() {
  if (!workspace.value?.id) return
  credsLoading.value = true
  try {
    const res = await credentialsAPI.list(workspace.value.id, { limit: 100 })
    credentials.value = res.data?.credentials ?? []
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load credentials'
    credentials.value = []
  } finally {
    credsLoading.value = false
  }
}

function openAddKeyModal() {
  editingKey.value = null
  keyForm.value = { name: '', committerName: '', committerEmail: '', privateKey: '', publicKey: '' }
  keyFormModalOpen.value = true
}

function openEditKeyModal(k) {
  editingKey.value = k
  keyForm.value = {
    name: k.name,
    committerName: k.committerName,
    committerEmail: k.committerEmail,
    privateKey: '',
    publicKey: k.publicKey ?? ''
  }
  keyFormModalOpen.value = true
}

async function saveKey() {
  if (!workspace.value?.id) return
  const { name, committerName, committerEmail, privateKey, publicKey } = keyForm.value
  if (!name?.trim() || !committerName?.trim() || !committerEmail?.trim()) return
  if (!editingKey.value && !privateKey?.trim()) return
  keyFormSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const payload = {
      name: name.trim(),
      committerName: committerName.trim(),
      committerEmail: committerEmail.trim()
    }
    if (privateKey?.trim()) payload.privateKey = privateKey.trim()
    if (publicKey?.trim()) payload.publicKey = publicKey.trim()
    if (editingKey.value) {
      await sshKeysAPI.update(workspace.value.id, editingKey.value.id, payload)
      successMessage.value = 'SSH key updated'
    } else {
      await sshKeysAPI.create(workspace.value.id, payload)
      successMessage.value = 'SSH key added'
    }
    keyFormModalOpen.value = false
    await loadSshKeys()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || (editingKey.value ? 'Failed to update SSH key' : 'Failed to add SSH key')
  } finally {
    keyFormSaving.value = false
  }
}

function openDeleteKeyModal(k) {
  deleteKeyModal.value = k
}

async function confirmDeleteKey() {
  const k = deleteKeyModal.value
  if (!workspace.value?.id || !k?.id) return
  deletingKeyId.value = k.id
  errorMessage.value = ''
  try {
    await sshKeysAPI.delete(workspace.value.id, k.id)
    successMessage.value = 'SSH key deleted'
    deleteKeyModal.value = null
    await loadSshKeys()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to delete SSH key'
  } finally {
    deletingKeyId.value = null
  }
}

function openAddCredModal() {
  editingCred.value = null
  credForm.value = { name: '', committerName: '', committerEmail: '', username: '', passwordOrToken: '' }
  credFormModalOpen.value = true
}

function openEditCredModal(c) {
  editingCred.value = c
  credForm.value = {
    name: c.name,
    committerName: c.committerName ?? '',
    committerEmail: c.committerEmail ?? '',
    username: c.username ?? '',
    passwordOrToken: ''
  }
  credFormModalOpen.value = true
}

async function saveCred() {
  if (!workspace.value?.id) return
  const { name, committerName, committerEmail, username, passwordOrToken } = credForm.value
  if (!name?.trim() || !committerName?.trim() || !committerEmail?.trim()) return
  if (!editingCred.value && !passwordOrToken) return
  credFormSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const payload = {
      name: name.trim(),
      committerName: committerName.trim(),
      committerEmail: committerEmail.trim(),
      username: username?.trim() || undefined
    }
    if (passwordOrToken) payload.passwordOrToken = passwordOrToken
    if (editingCred.value) {
      await credentialsAPI.update(workspace.value.id, editingCred.value.id, payload)
      successMessage.value = 'Git credentials updated'
    } else {
      await credentialsAPI.create(workspace.value.id, payload)
      successMessage.value = 'Git credentials added'
    }
    credFormModalOpen.value = false
    await loadCredentials()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || (editingCred.value ? 'Failed to update credentials' : 'Failed to add credentials')
  } finally {
    credFormSaving.value = false
  }
}

function openDeleteCredModal(c) {
  deleteCredModal.value = c
}

async function confirmDeleteCred() {
  const c = deleteCredModal.value
  if (!workspace.value?.id || !c?.id) return
  deletingCredId.value = c.id
  errorMessage.value = ''
  try {
    await credentialsAPI.delete(workspace.value.id, c.id)
    successMessage.value = 'Git credentials deleted'
    deleteCredModal.value = null
    await loadCredentials()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to delete credentials'
  } finally {
    deletingCredId.value = null
  }
}

onMounted(() => {
  setWorkspace()
  if (workspace.value?.id) {
    loadSshKeys()
    loadCredentials()
  }
})

watch(workspace, (w) => {
  if (w?.id) {
    loadSshKeys()
    loadCredentials()
  }
}, { immediate: false })
</script>
