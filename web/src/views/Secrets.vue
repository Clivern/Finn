<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('secretsPage.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('secretsPage.subtitle') }}
        </p>
      </header>

      <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
        {{ successMessage }}
      </div>

      <!-- SSH Keys section -->
      <section class="mb-10 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" :aria-label="$t('secretsPage.sshKeys')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.sshKeys') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('secretsPage.sshKeysDescription') }}</p>
          </div>
          <button
            type="button"
            @click="openAddKeyModal"
            class="btn-primary text-sm inline-flex items-center gap-2"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('secretsPage.addSshKey') }}
          </button>
        </div>
        <div v-if="keysLoading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('secretsPage.loadingKeys') }}</div>
        <div v-else-if="sshKeys.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('secretsPage.noKeys') }}</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('secretsPage.name') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('secretsPage.committer') }}</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('secretsPage.actions') }}</th>
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
                    <button type="button" @click="openEditKeyModal(k)" class="text-primary-600 hover:text-primary-700 hover:underline mr-3">{{ $t('common.edit') }}</button>
                    <button type="button" @click="openDeleteKeyModal(k)" class="text-red-600 hover:text-red-700 hover:underline">{{ $t('common.delete') }}</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </section>

      <!-- Credentials section -->
      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden" :aria-label="$t('secretsPage.credentials')">
        <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.credentials') }}</h2>
            <p class="text-sm text-theme-textLight mt-0.5">{{ $t('secretsPage.gitCredentialsDescription') }}</p>
          </div>
          <button
            type="button"
            @click="openAddCredModal"
            class="btn-primary text-sm inline-flex items-center gap-2"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ $t('secretsPage.addCredential') }}
          </button>
        </div>
        <div v-if="credsLoading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('secretsPage.loadingCreds') }}</div>
        <div v-else-if="credentials.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('secretsPage.noCreds') }}</div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('secretsPage.name') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('secretsPage.credType') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('secretsPage.committerUsername') }}</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">{{ $t('secretsPage.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="c in credentials" :key="c.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <p class="text-sm font-medium text-theme-text">{{ c.name }}</p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span class="text-sm text-theme-textLight">{{ credTypeLabel(c.type) }}</span>
                  </td>
                  <td class="px-6 py-4">
                    <p class="text-sm text-theme-textLight">
                      <template v-if="c.type === 'git'">
                        {{ c.committerName || '—' }} &lt;{{ c.committerEmail || '—' }}&gt;
                        <span v-if="c.username" class="block text-xs mt-0.5">{{ $t('secretsPage.user') }}: {{ c.username }}</span>
                      </template>
                      <template v-else-if="c.type === 'credential'">
                        {{ $t('secretsPage.user') }}: {{ c.username || '—' }}
                      </template>
                      <template v-else>—</template>
                    </p>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button type="button" @click="openEditCredModal(c)" class="text-primary-600 hover:text-primary-700 hover:underline mr-3">{{ $t('common.edit') }}</button>
                    <button type="button" @click="openDeleteCredModal(c)" class="text-red-600 hover:text-red-700 hover:underline">{{ $t('common.delete') }}</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </section>
    </main>

    <!-- SSH key Add modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showAddKeyModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="add-key-title"
          @click.self="closeAddKeyModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="add-key-title" class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.addKey') }}</h2>
            <div v-if="addKeyModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ addKeyModalError }}</p>
            </div>
            <form @submit.prevent="handleCreateKey" class="mt-4 space-y-4">
              <div>
                <label for="add-key-name" class="form-label">{{ $t('secretsPage.name') }}</label>
                <input id="add-key-name" v-model="keyForm.name" type="text" required maxlength="100" class="input-field" placeholder="e.g. GitHub deploy key" />
              </div>
              <div>
                <label for="add-key-committerName" class="form-label">{{ $t('secretsPage.committerName') }}</label>
                <input id="add-key-committerName" v-model="keyForm.committerName" type="text" required maxlength="100" class="input-field" placeholder="Your Name" />
              </div>
              <div>
                <label for="add-key-committerEmail" class="form-label">{{ $t('secretsPage.committerEmail') }}</label>
                <input id="add-key-committerEmail" v-model="keyForm.committerEmail" type="email" required maxlength="100" class="input-field" placeholder="you@example.com" />
              </div>
              <div>
                <label for="add-key-privateKey" class="form-label">{{ $t('secretsPage.privateKey') }}</label>
                <textarea
                  id="add-key-privateKey"
                  v-model="keyForm.privateKey"
                  required
                  rows="6"
                  class="input-field font-mono text-sm"
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                />
              </div>
              <div>
                <label for="add-key-publicKey" class="form-label">{{ $t('secretsPage.publicKey') }} {{ $t('secretsPage.publicKeyOptional') }}</label>
                <input id="add-key-publicKey" v-model="keyForm.publicKey" type="text" class="input-field font-mono text-sm" :placeholder="$t('secretsPage.publicKeyPlaceholder')" />
              </div>
              <div>
                <label for="add-key-passphrase" class="form-label">{{ $t('secretsPage.passphrase') }}</label>
                <input
                  id="add-key-passphrase"
                  v-model="keyForm.passphrase"
                  type="password"
                  autocomplete="off"
                  class="input-field"
                  :placeholder="$t('secretsPage.passphraseOptional')"
                />
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="closeAddKeyModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="keyFormSaving" class="btn-primary">
                  {{ keyFormSaving ? $t('secretsPage.saving') : $t('secretsPage.add') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- SSH key Edit modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showEditKeyModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="edit-key-title"
          @click.self="closeEditKeyModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="edit-key-title" class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.editKey') }}</h2>
            <div v-if="editKeyModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ editKeyModalError }}</p>
            </div>
            <form @submit.prevent="handleUpdateKey" class="mt-4 space-y-4">
              <div>
                <label for="edit-key-name" class="form-label">{{ $t('secretsPage.name') }}</label>
                <input id="edit-key-name" v-model="keyForm.name" type="text" required maxlength="100" class="input-field" placeholder="e.g. GitHub deploy key" />
              </div>
              <div>
                <label for="edit-key-committerName" class="form-label">{{ $t('secretsPage.committerName') }}</label>
                <input id="edit-key-committerName" v-model="keyForm.committerName" type="text" required maxlength="100" class="input-field" placeholder="Your Name" />
              </div>
              <div>
                <label for="edit-key-committerEmail" class="form-label">{{ $t('secretsPage.committerEmail') }}</label>
                <input id="edit-key-committerEmail" v-model="keyForm.committerEmail" type="email" required maxlength="100" class="input-field" placeholder="you@example.com" />
              </div>
              <div>
                <label for="edit-key-privateKey" class="form-label">{{ $t('secretsPage.privateKey') }} {{ $t('secretsPage.privateKeyLeaveBlank') }}</label>
                <textarea
                  id="edit-key-privateKey"
                  v-model="keyForm.privateKey"
                  rows="6"
                  class="input-field font-mono text-sm"
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                />
              </div>
              <div>
                <label for="edit-key-publicKey" class="form-label">{{ $t('secretsPage.publicKey') }} {{ $t('secretsPage.publicKeyOptional') }}</label>
                <input id="edit-key-publicKey" v-model="keyForm.publicKey" type="text" class="input-field font-mono text-sm" :placeholder="$t('secretsPage.publicKeyPlaceholder')" />
              </div>
              <div>
                <label for="edit-key-passphrase" class="form-label">{{ $t('secretsPage.passphrase') }} {{ $t('secretsPage.passphraseLeaveBlank') }}</label>
                <input
                  id="edit-key-passphrase"
                  v-model="keyForm.passphrase"
                  type="password"
                  autocomplete="off"
                  class="input-field"
                  :placeholder="$t('secretsPage.passphraseOptional')"
                />
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="closeEditKeyModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="keyFormSaving" class="btn-primary">
                  {{ keyFormSaving ? $t('secretsPage.saving') : $t('common.save') }}
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
          @click.self="closeDeleteKeyModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-key-title" class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.deleteKeyTitle') }}</h2>
            <div v-if="deleteKeyModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ deleteKeyModalError }}</p>
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('secretsPage.deleteKeyConfirm', { name: deleteKeyModal?.name }) }}
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" @click="closeDeleteKeyModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
              <button type="button" @click="confirmDeleteKey" :disabled="deletingKeyId === deleteKeyModal?.id" class="btn-danger">
                {{ deletingKeyId === deleteKeyModal?.id ? $t('secretsPage.deleting') : $t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Credential Add modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showAddCredModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="add-cred-title"
          @click.self="closeAddCredModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="add-cred-title" class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.addCredTitle') }}</h2>
            <div v-if="addCredModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ addCredModalError }}</p>
            </div>
            <form @submit.prevent="handleCreateCred" class="mt-4 space-y-4">
              <div>
                <label for="add-cred-name" class="form-label">{{ $t('secretsPage.name') }}</label>
                <input id="add-cred-name" v-model="credForm.name" type="text" required maxlength="100" class="input-field" placeholder="e.g. GitHub token" />
              </div>
              <div>
                <label for="add-cred-type" class="form-label">{{ $t('secretsPage.credType') }}</label>
                <select id="add-cred-type" v-model="credForm.type" class="input-field">
                  <option value="git">{{ $t('secretsPage.credTypeGit') }}</option>
                  <option value="credential">{{ $t('secretsPage.credTypeCredential') }}</option>
                  <option value="none">{{ $t('secretsPage.credTypeNone') }}</option>
                </select>
              </div>
              <template v-if="credForm.type === 'git'">
                <div>
                  <label for="add-cred-committerName" class="form-label">{{ $t('secretsPage.committerName') }}</label>
                  <input id="add-cred-committerName" v-model="credForm.committerName" type="text" required maxlength="100" class="input-field" placeholder="Your Name" />
                </div>
                <div>
                  <label for="add-cred-committerEmail" class="form-label">{{ $t('secretsPage.committerEmail') }}</label>
                  <input id="add-cred-committerEmail" v-model="credForm.committerEmail" type="email" required maxlength="100" class="input-field" placeholder="you@example.com" />
                </div>
              </template>
              <template v-if="credForm.type === 'git' || credForm.type === 'credential'">
                <div>
                  <label for="add-cred-username" class="form-label">{{ $t('secretsPage.username') }}</label>
                  <input id="add-cred-username" v-model="credForm.username" type="text" :required="credForm.type === 'git' || credForm.type === 'credential'" maxlength="100" class="input-field" placeholder="git or your username" />
                </div>
                <div>
                  <label for="add-cred-passwordOrToken" class="form-label">{{ $t('secretsPage.passwordOrToken') }}</label>
                  <input
                    id="add-cred-passwordOrToken"
                    v-model="credForm.passwordOrToken"
                    type="password"
                    :required="credForm.type === 'git' || credForm.type === 'credential'"
                    autocomplete="off"
                    class="input-field"
                    placeholder="••••••••"
                  />
                </div>
              </template>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="closeAddCredModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="credFormSaving" class="btn-primary">
                  {{ credFormSaving ? $t('secretsPage.saving') : $t('secretsPage.add') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Credential Edit modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="showEditCredModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="edit-cred-title"
          @click.self="closeEditCredModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
            <h2 id="edit-cred-title" class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.editCredTitle') }}</h2>
            <div v-if="editCredModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ editCredModalError }}</p>
            </div>
            <form @submit.prevent="handleUpdateCred" class="mt-4 space-y-4">
              <div>
                <label for="edit-cred-name" class="form-label">{{ $t('secretsPage.name') }}</label>
                <input id="edit-cred-name" v-model="credForm.name" type="text" required maxlength="100" class="input-field" placeholder="e.g. GitHub token" />
              </div>
              <div>
                <label for="edit-cred-type" class="form-label">{{ $t('secretsPage.credType') }}</label>
                <select id="edit-cred-type" v-model="credForm.type" class="input-field">
                  <option value="git">{{ $t('secretsPage.credTypeGit') }}</option>
                  <option value="credential">{{ $t('secretsPage.credTypeCredential') }}</option>
                  <option value="none">{{ $t('secretsPage.credTypeNone') }}</option>
                </select>
              </div>
              <template v-if="credForm.type === 'git'">
                <div>
                  <label for="edit-cred-committerName" class="form-label">{{ $t('secretsPage.committerName') }}</label>
                  <input id="edit-cred-committerName" v-model="credForm.committerName" type="text" required maxlength="100" class="input-field" placeholder="Your Name" />
                </div>
                <div>
                  <label for="edit-cred-committerEmail" class="form-label">{{ $t('secretsPage.committerEmail') }}</label>
                  <input id="edit-cred-committerEmail" v-model="credForm.committerEmail" type="email" required maxlength="100" class="input-field" placeholder="you@example.com" />
                </div>
              </template>
              <template v-if="credForm.type === 'git' || credForm.type === 'credential'">
                <div>
                  <label for="edit-cred-username" class="form-label">{{ $t('secretsPage.username') }}</label>
                  <input id="edit-cred-username" v-model="credForm.username" type="text" :required="credForm.type === 'git' || credForm.type === 'credential'" maxlength="100" class="input-field" placeholder="git or your username" />
                </div>
                <div>
                  <label for="edit-cred-passwordOrToken" class="form-label">{{ $t('secretsPage.passwordOrToken') }} {{ $t('secretsPage.passwordLeaveBlank') }}</label>
                  <input
                    id="edit-cred-passwordOrToken"
                    v-model="credForm.passwordOrToken"
                    type="password"
                    autocomplete="off"
                    class="input-field"
                    placeholder="••••••••"
                  />
                </div>
              </template>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="closeEditCredModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
                <button type="submit" :disabled="credFormSaving" class="btn-primary">
                  {{ credFormSaving ? $t('secretsPage.saving') : $t('common.save') }}
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
          @click.self="closeDeleteCredModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6">
            <h2 id="delete-cred-title" class="text-lg font-semibold text-theme-text">{{ $t('secretsPage.deleteCredTitle') }}</h2>
            <div v-if="deleteCredModalError" class="mt-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ deleteCredModalError }}</p>
            </div>
            <p class="mt-2 text-sm text-theme-textLight">
              {{ $t('secretsPage.deleteCredConfirm', { name: deleteCredModal?.name }) }}
            </p>
            <div class="flex gap-3 justify-end mt-6">
              <button type="button" @click="closeDeleteCredModal" class="btn-secondary">{{ $t('common.cancel') }}</button>
              <button type="button" @click="confirmDeleteCred" :disabled="deletingCredId === deleteCredModal?.id" class="btn-danger">
                {{ deletingCredId === deleteCredModal?.id ? $t('secretsPage.deleting') : $t('common.delete') }}
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
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { loadWorkspaceFromStorage } from '@/utils/storage'
import { sshKeysAPI, credentialsAPI } from '@/api'

const { t } = useI18n()

const router = useRouter()
const workspace = ref(null)
const sshKeys = ref([])
const credentials = ref([])
const keysLoading = ref(true)
const credsLoading = ref(true)
const errorMessage = ref('')
const successMessage = ref('')

const showAddKeyModal = ref(false)
const showEditKeyModal = ref(false)
const keyFormSaving = ref(false)
const editingKey = ref(null)
const keyForm = ref({
  name: '',
  committerName: '',
  committerEmail: '',
  privateKey: '',
  publicKey: '',
  passphrase: ''
})
const deleteKeyModal = ref(null)
const deletingKeyId = ref(null)

const showAddCredModal = ref(false)
const showEditCredModal = ref(false)
const credFormSaving = ref(false)
const editingCred = ref(null)
const credForm = ref({
  type: 'git',
  name: '',
  committerName: '',
  committerEmail: '',
  username: '',
  passwordOrToken: ''
})
const deleteCredModal = ref(null)
const deletingCredId = ref(null)

const addKeyModalError = ref(null)
const editKeyModalError = ref(null)
const deleteKeyModalError = ref(null)
const addCredModalError = ref(null)
const editCredModalError = ref(null)
const deleteCredModalError = ref(null)

function closeAddKeyModal() {
  showAddKeyModal.value = false
  addKeyModalError.value = null
}

function closeEditKeyModal() {
  showEditKeyModal.value = false
  editKeyModalError.value = null
}

function closeDeleteKeyModal() {
  deleteKeyModal.value = null
  deleteKeyModalError.value = null
}

function closeAddCredModal() {
  showAddCredModal.value = false
  addCredModalError.value = null
}

function closeEditCredModal() {
  showEditCredModal.value = false
  editCredModalError.value = null
}

function closeDeleteCredModal() {
  deleteCredModal.value = null
  deleteCredModalError.value = null
}

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
    errorMessage.value = err.response?.data?.errorMessage || t('secretsPage.failedLoadKeys')
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
    errorMessage.value = err.response?.data?.errorMessage || t('secretsPage.failedLoadCreds')
    credentials.value = []
  } finally {
    credsLoading.value = false
  }
}

function openAddKeyModal() {
  editingKey.value = null
  keyForm.value = {
    name: '',
    committerName: '',
    committerEmail: '',
    privateKey: '',
    publicKey: '',
    passphrase: ''
  }
  addKeyModalError.value = null
  showAddKeyModal.value = true
  showEditKeyModal.value = false
}

function openEditKeyModal(k) {
  editingKey.value = k
  keyForm.value = {
    name: k.name,
    committerName: k.committerName,
    committerEmail: k.committerEmail,
    privateKey: '',
    publicKey: k.publicKey ?? '',
    passphrase: ''
  }
  editKeyModalError.value = null
  showEditKeyModal.value = true
  showAddKeyModal.value = false
}

async function handleCreateKey() {
  keyFormSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  addKeyModalError.value = null

  try {
    const payload = {
      name: keyForm.value.name.trim(),
      committerName: keyForm.value.committerName.trim(),
      committerEmail: keyForm.value.committerEmail.trim()
    }
    if (keyForm.value.privateKey?.trim()) { payload.privateKey = keyForm.value.privateKey.trim() }
    if (keyForm.value.publicKey?.trim()) { payload.publicKey = keyForm.value.publicKey.trim() }
    if (keyForm.value.passphrase?.trim()) { payload.passphrase = keyForm.value.passphrase.trim() }
    await sshKeysAPI.create(workspace.value.id, payload)
    successMessage.value = t('secretsPage.sshKeyAdded')
    closeAddKeyModal()
    await loadSshKeys()
  } catch (err) {
    addKeyModalError.value = err.response?.data?.errorMessage || t('secretsPage.failedAddKey')
  } finally {
    keyFormSaving.value = false
  }
}

async function handleUpdateKey() {
  keyFormSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  editKeyModalError.value = null

  try {
    const payload = {
      name: keyForm.value.name.trim(),
      committerName: keyForm.value.committerName.trim(),
      committerEmail: keyForm.value.committerEmail.trim()
    }
    if (keyForm.value.privateKey?.trim()) { payload.privateKey = keyForm.value.privateKey.trim() }
    if (keyForm.value.publicKey?.trim()) { payload.publicKey = keyForm.value.publicKey.trim() }
    if (keyForm.value.passphrase?.trim()) { payload.passphrase = keyForm.value.passphrase.trim() }
    await sshKeysAPI.update(workspace.value.id, editingKey.value.id, payload)
    successMessage.value = t('secretsPage.sshKeyUpdated')
    closeEditKeyModal()
    await loadSshKeys()
  } catch (err) {
    editKeyModalError.value = err.response?.data?.errorMessage || t('secretsPage.failedUpdateKey')
  } finally {
    keyFormSaving.value = false
  }
}

function openDeleteKeyModal(k) {
  deleteKeyModalError.value = null
  deleteKeyModal.value = k
}

async function confirmDeleteKey() {
  if (!workspace.value?.id || !deleteKeyModal.value?.id) return
  deletingKeyId.value = deleteKeyModal.value.id
  errorMessage.value = ''
  deleteKeyModalError.value = null

  try {
    await sshKeysAPI.delete(
      workspace.value.id,
      deletingKeyId.value
    )
    successMessage.value = t('secretsPage.sshKeyDeleted')
    closeDeleteKeyModal()
    await loadSshKeys()
  } catch (err) {
    deleteKeyModalError.value = err.response?.data?.errorMessage || t('secretsPage.failedDeleteKey')
  } finally {
    deletingKeyId.value = null
  }
}

function credTypeLabel(type) {
  if (type === 'git') return t('secretsPage.credTypeGit')
  if (type === 'credential') return t('secretsPage.credTypeCredential')
  if (type === 'none') return t('secretsPage.credTypeNone')
  return type || '—'
}

function openAddCredModal() {
  editingCred.value = null
  credForm.value = {
    type: 'git',
    name: '',
    committerName: '',
    committerEmail: '',
    username: '',
    passwordOrToken: ''
  }
  addCredModalError.value = null
  showAddCredModal.value = true
  showEditCredModal.value = false
}

function openEditCredModal(c) {
  editingCred.value = c
  credForm.value = {
    type: c.type || 'git',
    name: c.name,
    committerName: c.committerName ?? '',
    committerEmail: c.committerEmail ?? '',
    username: c.username ?? '',
    passwordOrToken: ''
  }
  editCredModalError.value = null
  showEditCredModal.value = true
  showAddCredModal.value = false
}

async function handleCreateCred() {
  credFormSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  addCredModalError.value = null

  try {
    const type = credForm.value.type || 'git'
    const payload = {
      type,
      name: credForm.value.name.trim()
    }
    if (type === 'git') {
      payload.committerName = credForm.value.committerName.trim()
      payload.committerEmail = credForm.value.committerEmail.trim()
      payload.username = credForm.value.username.trim()
      payload.passwordOrToken = credForm.value.passwordOrToken
    } else if (type === 'credential') {
      payload.username = credForm.value.username.trim()
      payload.passwordOrToken = credForm.value.passwordOrToken
    }
    await credentialsAPI.create(workspace.value.id, payload)
    successMessage.value = t('secretsPage.credAdded')
    closeAddCredModal()
    await loadCredentials()
  } catch (err) {
    addCredModalError.value = err.response?.data?.errorMessage || t('secretsPage.failedAddCred')
  } finally {
    credFormSaving.value = false
  }
}

async function handleUpdateCred() {
  credFormSaving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  editCredModalError.value = null

  try {
    const type = credForm.value.type || 'git'
    const payload = {
      type,
      name: credForm.value.name.trim()
    }
    if (type === 'git') {
      payload.committerName = credForm.value.committerName.trim()
      payload.committerEmail = credForm.value.committerEmail.trim()
      payload.username = credForm.value.username.trim()
    } else if (type === 'credential') {
      payload.username = credForm.value.username.trim()
    }
    if (credForm.value.passwordOrToken) payload.passwordOrToken = credForm.value.passwordOrToken
    await credentialsAPI.update(workspace.value.id, editingCred.value.id, payload)
    successMessage.value = t('secretsPage.credUpdated')
    closeEditCredModal()
    await loadCredentials()
  } catch (err) {
    editCredModalError.value = err.response?.data?.errorMessage || t('secretsPage.failedUpdateCred')
  } finally {
    credFormSaving.value = false
  }
}

function openDeleteCredModal(c) {
  deleteCredModalError.value = null
  deleteCredModal.value = c
}

async function confirmDeleteCred() {
  if (!workspace.value?.id || !deleteCredModal.value?.id) return
  deletingCredId.value = deleteCredModal.value.id
  errorMessage.value = ''
  deleteCredModalError.value = null

  try {
    await credentialsAPI.delete(
      workspace.value.id,
      deletingCredId.value
    )
    successMessage.value = t('secretsPage.credDeleted')
    closeDeleteCredModal()
    await loadCredentials()
  } catch (err) {
    deleteCredModalError.value = err.response?.data?.errorMessage || t('secretsPage.failedDeleteCred')
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
