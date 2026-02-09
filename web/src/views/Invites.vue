<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-semibold text-theme-text">{{ $t('invites.title') }}</h1>
          <p class="text-theme-textLight mt-2">{{ $t('invitesPage.subtitle') }}</p>
        </div>
        <button
          @click="openCreateModal"
          class="btn-primary"
        >
          {{ $t('invitesPage.createInvite') }}
        </button>
      </div>

      <div v-if="successMessage" class="mb-6 rounded-md border border-green-200 bg-green-50 p-4">
        <p class="text-sm text-green-800">{{ successMessage }}</p>
      </div>
      <div v-if="createdInviteLink" class="mb-6 rounded-lg border border-primary-200 bg-primary-50 p-4">
        <p class="text-sm font-medium text-primary-900 mb-1">{{ $t('invitesPage.inviteCreated') }}</p>
        <p class="text-xs text-primary-800 mb-2">{{ $t('invitesPage.inviteCreatedHint') }}</p>
        <div class="flex flex-wrap items-center gap-2">
          <code class="flex-1 min-w-0 text-xs text-primary-900 bg-white/80 px-2 py-1.5 rounded border border-primary-200 truncate">{{ createdInviteLink }}</code>
          <button
            type="button"
            @click="copyCreatedInviteLink"
            class="btn-secondary text-sm shrink-0"
          >
            {{ $t('invitesPage.copyLink') }}
          </button>
          <button
            type="button"
            @click="dismissInviteLinkCard"
            class="text-sm text-primary-700 hover:text-primary-900 shrink-0"
          >
            {{ $t('invitesPage.dismiss') }}
          </button>
        </div>
      </div>
      <div v-if="errorMessage" class="mb-6 rounded-md border border-red-200 bg-red-50 p-4">
        <p class="text-sm text-red-800">{{ errorMessage }}</p>
      </div>

      <div class="bg-white rounded-lg border border-theme-border overflow-hidden">
        <div v-if="loading" class="p-8 text-center">
          <svg class="animate-spin h-6 w-6 mx-auto text-theme-text" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <p class="text-theme-textLight mt-2">{{ $t('invitesPage.loading') }}</p>
        </div>

        <div v-else-if="invites.length === 0" class="p-8 text-center">
          <p class="text-theme-textLight">{{ $t('invitesPage.noInvites') }}</p>
          <p class="text-sm text-theme-textLight mt-1">{{ $t('invitesPage.noInvitesHint') }}</p>
        </div>

        <table v-else class="min-w-full divide-y divide-theme-border">
          <thead class="bg-theme-hover">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('invitesPage.user') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('invitesPage.role') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('invitesPage.status') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('invitesPage.expires') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('invitesPage.created') }}
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('invitesPage.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-theme-border">
            <tr v-for="invite in paginatedInvites" :key="invite.id" class="hover:bg-theme-hover">
              <td class="px-6 py-4 whitespace-nowrap">
                <div>
                  <div class="text-sm font-medium text-theme-text">{{ invite.email }}</div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="{
                    'bg-purple-100 text-purple-800': invite.role === 'admin',
                    'bg-blue-100 text-blue-800': invite.role === 'regular'
                  }">
                  {{ invite.role.charAt(0).toUpperCase() + invite.role.slice(1) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="{
                    'bg-amber-100 text-amber-800': invite.status === 'pending',
                    'bg-green-100 text-green-800': invite.status === 'accepted',
                    'bg-red-100 text-red-800': invite.status === 'expired'
                  }">
                  {{ invite.status.charAt(0).toUpperCase() + invite.status.slice(1) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatExpires(invite.expiresAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatRelative(invite.createdAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  v-if="invite.token && invite.status === 'pending'"
                  @click="copyInviteLink(invite)"
                  class="text-blue-600 hover:text-blue-900 mr-4"
                >
                  {{ $t('invitesPage.copyLink') }}
                </button>
                <button
                  @click="openDeleteModal(invite)"
                  class="text-red-600 hover:text-red-900"
                >
                  {{ $t('invitesPage.delete') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div v-if="invites.length > 0" class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between">
          <div class="text-sm text-theme-textLight">
            {{ $t('invitesPage.showing', { from: showingFrom, to: showingTo, count: invites.length }) }}
          </div>
          <div class="flex items-center space-x-2">
            <button
              @click="goToPage(offset - limit)"
              :disabled="offset === 0"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('invitesPage.previous') }}
            </button>
            <button
              @click="goToPage(offset + limit)"
              :disabled="offset + limit >= invites.length"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('invitesPage.next') }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Create invite modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeCreateModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeCreateModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <form @submit.prevent="handleCreate">
            <div class="bg-white px-6 pt-6 pb-4">
              <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold text-theme-text">{{ $t('invitesPage.createModalTitle') }}</h3>
                <button
                  type="button"
                  @click="closeCreateModal"
                  class="text-theme-textLight hover:text-theme-text"
                >
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div v-if="createModalError" class="mb-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
                <p class="text-sm text-red-800">{{ createModalError }}</p>
              </div>
              <div class="space-y-4">
                <div>
                  <label for="create-email" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('invitesPage.emailRequired') }}
                  </label>
                  <input
                    id="create-email"
                    v-model="createForm.email"
                    type="email"
                    required
                    class="input-field"
                    placeholder="user@example.com"
                  />
                </div>
                <div>
                  <label for="create-role" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('invitesPage.roleRequired') }}
                  </label>
                  <select
                    id="create-role"
                    v-model="createForm.role"
                    required
                    class="input-field"
                  >
                    <option value="regular">Regular</option>
                    <option value="admin">Admin</option>
                  </select>
                </div>
              </div>
            </div>
            <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
              <button
                type="button"
                @click="closeCreateModal"
                class="btn-secondary"
              >
                {{ $t('invitesPage.cancel') }}
              </button>
              <button
                type="submit"
                class="btn-primary"
                :disabled="createLoading"
              >
                <span v-if="!createLoading">{{ $t('invitesPage.createInviteSubmit') }}</span>
                <span v-else class="inline-flex items-center gap-2">
                  <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  {{ $t('invitesPage.creating') }}
                </span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Delete confirmation modal -->
    <div v-if="showDeleteModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeDeleteModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeDeleteModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div class="bg-white px-6 pt-6 pb-4">
            <div class="flex justify-between items-center mb-4">
              <h3 class="text-lg font-semibold text-theme-text">{{ $t('invitesPage.deleteModalTitle') }}</h3>
              <button
                type="button"
                @click="closeDeleteModal"
                class="text-theme-textLight hover:text-theme-text"
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div v-if="deleteModalError" class="mb-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
              <p class="text-sm text-red-800">{{ deleteModalError }}</p>
            </div>
            <p class="text-sm text-theme-textLight">
              {{ $t('invitesPage.deleteConfirm', { email: inviteToDelete?.email }) }}
            </p>
          </div>
          <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
            <button
              type="button"
              @click="closeDeleteModal"
              class="btn-secondary"
            >
              {{ $t('invitesPage.cancel') }}
            </button>
            <button
              @click="handleDelete"
              class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700 focus:outline-none transition-all duration-150"
              :disabled="deleteLoading"
            >
              <span v-if="!deleteLoading">{{ $t('common.delete') }}</span>
              <span v-else class="flex items-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                {{ $t('invitesPage.deleting') }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { invitesAPI } from '@/api'
import AppNav from '@/components/AppNav.vue'
import { formatRelative, formatExpires } from '@/utils/date'

const { t } = useI18n()

const loading = ref(false)
const invites = ref([])
const successMessage = ref(null)
const errorMessage = ref(null)
const limit = ref(50)
const offset = ref(0)

const showCreateModal = ref(false)
const showDeleteModal = ref(false)
const createLoading = ref(false)
const deleteLoading = ref(false)
const inviteToDelete = ref(null)
const createdInviteLink = ref(null)

const createModalError = ref(null)
const deleteModalError = ref(null)

const createForm = reactive({
  email: '',
  role: 'regular'
})

const paginatedInvites = computed(() => {
  const end = offset.value + limit.value
  return invites.value.slice(offset.value, end)
})

const showingFrom = computed(() => (invites.value.length === 0 ? 0 : offset.value + 1))
const showingTo = computed(() => Math.min(offset.value + limit.value, invites.value.length))

const loadInvites = async () => {
  loading.value = true
  errorMessage.value = null
  try {
    const res = await invitesAPI.list()
    invites.value = res.data.invites
    offset.value = 0
  } catch (err) {
    console.error('Failed to load invites:', err)
    errorMessage.value = err.response?.data?.errorMessage || t('invitesPage.failedLoad')
  } finally {
    loading.value = false
  }
}

const goToPage = (newOffset) => {
  const maxOffset = Math.max(0, invites.value.length - limit.value)
  offset.value = Math.max(0, Math.min(newOffset, maxOffset))
}

const inviteLink = (token) => {
  const base = typeof window !== 'undefined' ? window.location.origin : ''
  return `${base}/register?token=${encodeURIComponent(token)}`
}

const copyInviteLink = async (invite) => {
  try {
    await navigator.clipboard.writeText(inviteLink(invite.token))
    successMessage.value = t('invitesPage.inviteLinkCopied')
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch {
    errorMessage.value = 'Failed to copy link'
    setTimeout(() => { errorMessage.value = null }, 3000)
  }
}

const openCreateModal = () => {
  createdInviteLink.value = null
  createForm.email = ''
  createForm.role = 'regular'
  createModalError.value = null
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
  createModalError.value = null
}

const copyCreatedInviteLink = async () => {
  if (!createdInviteLink.value) return

  try {
    await navigator.clipboard.writeText(createdInviteLink.value)
    successMessage.value = t('invitesPage.inviteLinkCopied')
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch {
    errorMessage.value = 'Failed to copy link'
    setTimeout(() => { errorMessage.value = null }, 3000)
  }
}

const dismissInviteLinkCard = () => {
  createdInviteLink.value = null
}

const handleCreate = async () => {
  createLoading.value = true
  errorMessage.value = null
  successMessage.value = null
  createdInviteLink.value = null
  createModalError.value = null

  try {
    const res = await invitesAPI.create({
      email: createForm.email,
      role: createForm.role
    })

    createdInviteLink.value = inviteLink(res.data.token)
    successMessage.value = t('invitesPage.inviteCreated')
    closeCreateModal()
    await loadInvites()
    setTimeout(() => { successMessage.value = null }, 5000)
  } catch (err) {
    console.error('Failed to create invite:', err)
    createModalError.value = err.response?.data?.errorMessage || t('invitesPage.failedCreate')
  } finally {
    createLoading.value = false
  }
}

const openDeleteModal = (invite) => {
  inviteToDelete.value = invite
  deleteModalError.value = null
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  deleteModalError.value = null
  inviteToDelete.value = null
}

const handleDelete = async () => {
  if (!inviteToDelete.value?.id) return
  deleteLoading.value = true
  errorMessage.value = null
  deleteModalError.value = null

  try {
    await invitesAPI.delete(inviteToDelete.value.id)
    successMessage.value = t('invitesPage.inviteDeleted')
    closeDeleteModal()
    await loadInvites()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to delete invite:', err)
    deleteModalError.value = err.response?.data?.errorMessage || t('invitesPage.failedDelete')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(loadInvites)
</script>
