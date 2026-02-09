<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-semibold text-theme-text">{{ $t('users.title') }}</h1>
          <p class="text-theme-textLight mt-2">{{ $t('usersPage.subtitle') }}</p>
        </div>
        <button
          @click="openCreateModal"
          class="btn-primary"
        >
          {{ $t('users.addUser') }}
        </button>
      </div>

      <div v-if="successMessage" class="mb-6 rounded-md border border-green-200 bg-green-50 p-4">
        <p class="text-sm text-green-800">{{ successMessage }}</p>
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
          <p class="text-theme-textLight mt-2">{{ $t('usersPage.loading') }}</p>
        </div>

        <div v-else-if="users.length === 0" class="p-8 text-center">
          <p class="text-theme-textLight">{{ $t('usersPage.noUsers') }}</p>
        </div>

        <table v-else class="min-w-full divide-y divide-theme-border">
          <thead class="bg-theme-hover">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('usersPage.user') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('usersPage.role') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('usersPage.status') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('usersPage.lastLogin') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('usersPage.created') }}
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">
                {{ $t('usersPage.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-theme-border">
            <tr v-for="user in users" :key="user.id" class="hover:bg-theme-hover">
              <td class="px-6 py-4 whitespace-nowrap">
                <div>
                  <div class="text-sm font-medium text-theme-text">{{ user.name || '—' }}</div>
                  <div class="text-xs text-theme-textLight">{{ user.email }}</div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="{
                    'bg-purple-100 text-purple-800': user.role === 'admin',
                    'bg-blue-100 text-blue-800': user.role === 'regular'
                  }">
                  {{ user.role.charAt(0).toUpperCase() + user.role.slice(1) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="{
                    'bg-green-100 text-green-800': user.isActive,
                    'bg-red-100 text-red-800': !user.isActive
                  }">
                  {{ user.isActive ? $t('usersPage.active') : $t('usersPage.inactive') }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatRelative(user.lastLoginAt, 'Never') }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                {{ formatRelative(user.createdAt, 'Never') }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  @click="openEditModal(user)"
                  class="text-blue-600 hover:text-blue-900 mr-4"
                >
                  {{ $t('common.edit') }}
                </button>
                <button
                  @click="openDeleteModal(user)"
                  class="text-red-600 hover:text-red-900"
                >
                  {{ $t('common.delete') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="total > 0" class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between">
          <div class="text-sm text-theme-textLight">
            {{ $t('usersPage.showing', { from: offset + 1, to: Math.min(offset + limit, total), total }) }}
          </div>
          <div class="flex items-center space-x-2">
            <button
              @click="goToPage(offset - limit)"
              :disabled="offset === 0"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('usersPage.previous') }}
            </button>
            <button
              @click="goToPage(offset + limit)"
              :disabled="offset + limit >= total"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('usersPage.next') }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <div v-if="showCreateModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeCreateModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeCreateModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <form @submit.prevent="handleCreate">
            <div class="bg-white px-6 pt-6 pb-4">
              <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold text-theme-text">{{ $t('usersPage.createUser') }}</h3>
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
                  <label for="create-name" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('usersPage.name') }} *
                  </label>
                  <input
                    id="create-name"
                    v-model="createForm.name"
                    type="text"
                    required
                    class="input-field"
                    placeholder="Full name"
                  />
                </div>
                <div>
                  <label for="create-email" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('usersPage.email') }} *
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
                  <label for="create-password" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('usersPage.password') }} *
                  </label>
                  <input
                    id="create-password"
                    v-model="createForm.password"
                    type="password"
                    required
                    class="input-field"
                    placeholder="Minimum 8 characters"
                  />
                </div>
                <div>
                  <label for="create-role" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('usersPage.roleRequired') }}
                  </label>
                  <select
                    id="create-role"
                    v-model="createForm.role"
                    required
                    class="input-field"
                  >
                    <option value="admin">{{ $t('usersPage.admin') }}</option>
                    <option value="regular">{{ $t('usersPage.regular') }}</option>
                  </select>
                </div>
                <div class="flex items-start">
                  <input
                    id="create-isActive"
                    v-model="createForm.isActive"
                    type="checkbox"
                    class="h-4 w-4 mt-1 rounded text-theme-text focus:ring-theme-text border-theme-border"
                  />
                  <label for="create-isActive" class="ml-3 block">
                    <span class="text-sm font-medium text-theme-text">{{ $t('usersPage.active') }}</span>
                    <p class="text-xs text-theme-textLight mt-1">{{ $t('usersPage.activeHint') }}</p>
                  </label>
                </div>
              </div>
            </div>
            <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
              <button
                type="button"
                @click="closeCreateModal"
                class="btn-secondary"
              >
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                class="btn-primary"
                :disabled="createLoading"
              >
                <span v-if="!createLoading">{{ $t('common.create') }}</span>
                <span v-else class="flex items-center">
                  <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  {{ $t('usersPage.creating') }}
                </span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <div v-if="showEditModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeEditModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeEditModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <form @submit.prevent="handleUpdate">
            <div class="bg-white px-6 pt-6 pb-4">
              <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold text-theme-text">{{ $t('usersPage.editUser') }}</h3>
                <button
                  type="button"
                  @click="closeEditModal"
                  class="text-theme-textLight hover:text-theme-text"
                >
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div v-if="editModalError" class="mb-4 rounded-md border border-red-200 bg-red-50 p-3" role="alert">
                <p class="text-sm text-red-800">{{ editModalError }}</p>
              </div>
              <div class="space-y-4">
                <div>
                  <label for="edit-email" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('usersPage.email') }} *
                  </label>
                  <input
                    id="edit-email"
                    v-model="editForm.email"
                    type="email"
                    required
                    class="input-field"
                    placeholder="user@example.com"
                  />
                </div>
                <div>
                  <label for="edit-password" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('usersPage.password') }}
                  </label>
                  <input
                    id="edit-password"
                    v-model="editForm.password"
                    type="password"
                    class="input-field"
                    placeholder="Leave empty to keep current password"
                  />
                  <p class="text-xs text-theme-textLight mt-1">{{ $t('usersPage.passwordLeaveBlank') }}</p>
                </div>
                <div>
                  <label for="edit-role" class="block text-sm font-medium text-theme-text mb-2">
                    {{ $t('usersPage.roleRequired') }}
                  </label>
                  <select
                    id="edit-role"
                    v-model="editForm.role"
                    required
                    class="input-field"
                  >
                    <option value="admin">{{ $t('usersPage.admin') }}</option>
                    <option value="regular">{{ $t('usersPage.regular') }}</option>
                  </select>
                </div>
                <div class="flex items-start">
                  <input
                    id="edit-isActive"
                    v-model="editForm.isActive"
                    type="checkbox"
                    class="h-4 w-4 mt-1 rounded text-theme-text focus:ring-theme-text border-theme-border"
                  />
                  <label for="edit-isActive" class="ml-3 block">
                    <span class="text-sm font-medium text-theme-text">{{ $t('usersPage.active') }}</span>
                    <p class="text-xs text-theme-textLight mt-1">{{ $t('usersPage.activeHint') }}</p>
                  </label>
                </div>
              </div>
            </div>
            <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
              <button
                type="button"
                @click="closeEditModal"
                class="btn-secondary"
              >
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                class="btn-primary"
                :disabled="editLoading"
              >
                <span v-if="!editLoading">{{ $t('common.save') }}</span>
                <span v-else class="flex items-center">
                  <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  {{ $t('usersPage.updating') }}
                </span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <div v-if="showDeleteModal" class="fixed inset-0 z-50 overflow-y-auto" @click.self="closeDeleteModal">
      <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" @click="closeDeleteModal"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div class="bg-white px-6 pt-6 pb-4">
            <div class="flex justify-between items-center mb-4">
              <h3 class="text-lg font-semibold text-theme-text">{{ $t('usersPage.deleteUser') }}</h3>
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
              {{ $t('usersPage.deleteUserConfirm', { email: userToDelete?.email }) }}
            </p>
          </div>
          <div class="bg-theme-hover px-6 py-4 flex justify-end space-x-3">
            <button
              type="button"
              @click="closeDeleteModal"
              class="btn-secondary"
            >
              {{ $t('common.cancel') }}
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
                {{ $t('usersPage.deleting') }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { userAPI } from '@/api'
import AppNav from '@/components/AppNav.vue'
import { formatRelative } from '@/utils/date'

const { t } = useI18n()

const loading = ref(false)
const users = ref([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const successMessage = ref(null)
const errorMessage = ref(null)

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const createLoading = ref(false)
const editLoading = ref(false)
const deleteLoading = ref(false)
const userToDelete = ref(null)

const createModalError = ref(null)
const editModalError = ref(null)
const deleteModalError = ref(null)

const createForm = reactive({
  name: '',
  email: '',
  password: '',
  role: 'regular',
  isActive: true
})

const editForm = reactive({
  id: null,
  email: '',
  password: '',
  role: 'regular',
  isActive: true
})

const loadUsers = async () => {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await userAPI.getUsers({
      limit: limit.value,
      offset: offset.value
    })

    users.value = res.data.users
    total.value = res.data._meta.total
  } catch (err) {
    console.error('Failed to load users:', err)
    errorMessage.value = err.response?.data?.errorMessage || t('usersPage.failedLoad')
  } finally {
    loading.value = false
  }
}

const goToPage = (newOffset) => {
  if (newOffset < 0) return
  if (newOffset >= total.value) return
  offset.value = newOffset
  loadUsers()
}

const openCreateModal = () => {
  createForm.name = ''
  createForm.email = ''
  createForm.password = ''
  createForm.role = 'regular'
  createForm.isActive = true
  createModalError.value = null
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
  createModalError.value = null
  createForm.name = ''
  createForm.email = ''
  createForm.password = ''
  createForm.role = 'regular'
  createForm.isActive = true
}

const handleCreate = async () => {
  createLoading.value = true
  errorMessage.value = null
  successMessage.value = null
  createModalError.value = null

  try {
    await userAPI.createUser({
      name: createForm.name,
      email: createForm.email,
      password: createForm.password,
      role: createForm.role,
      isActive: createForm.isActive
    })
    successMessage.value = t('usersPage.userCreated')
    closeCreateModal()
    await loadUsers()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to create user:', err)
    createModalError.value = err.response?.data?.errorMessage || t('usersPage.failedCreate')
  } finally {
    createLoading.value = false
  }
}

const openEditModal = (user) => {
  editForm.id = user.id
  editForm.email = user.email
  editForm.password = ''
  editForm.role = user.role
  editForm.isActive = user.isActive
  editModalError.value = null
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  editModalError.value = null
  editForm.id = null
  editForm.email = ''
  editForm.password = ''
  editForm.role = 'regular'
  editForm.isActive = true
}

const handleUpdate = async () => {
  editLoading.value = true
  errorMessage.value = null
  successMessage.value = null
  editModalError.value = null

  const payload = {
    email: editForm.email,
    role: editForm.role,
    isActive: editForm.isActive
  }
  if (editForm.password) payload.password = editForm.password

  try {
    await userAPI.updateUser(editForm.id, payload)
    successMessage.value = t('usersPage.userUpdated')
    closeEditModal()
    await loadUsers()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to update user:', err)
    editModalError.value = err.response?.data?.errorMessage || t('usersPage.failedUpdate')
  } finally {
    editLoading.value = false
  }
}

const openDeleteModal = (user) => {
  userToDelete.value = user
  deleteModalError.value = null
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  deleteModalError.value = null
  userToDelete.value = null
}

const handleDelete = async () => {
  if (!userToDelete.value) return

  deleteLoading.value = true
  errorMessage.value = null
  successMessage.value = null
  deleteModalError.value = null

  try {
    await userAPI.deleteUser(userToDelete.value.id)
    successMessage.value = t('usersPage.userDeleted')
    closeDeleteModal()
    await loadUsers()
    setTimeout(() => { successMessage.value = null }, 3000)
  } catch (err) {
    console.error('Failed to delete user:', err)
    deleteModalError.value = err.response?.data?.errorMessage || t('usersPage.failedDelete')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadUsers()
})
</script>
