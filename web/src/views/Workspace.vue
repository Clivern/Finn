<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div v-if="!workspace" class="flex items-center justify-center py-24">
        <div class="text-center">
          <svg class="animate-spin h-10 w-10 mx-auto text-theme-text" fill="none" viewBox="0 0 24 24" aria-hidden="true">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <p class="text-theme-textLight mt-3">{{ $t('common.loading') }}</p>
        </div>
      </div>

      <template v-else>
        <div class="mb-8">
          <h1 class="text-3xl font-semibold text-theme-text">{{ $t('workspacePage.title') }}</h1>
          <p class="text-theme-textLight mt-2">{{ $t('workspacePage.subtitle') }}</p>
        </div>

        <div v-if="successMessage" class="mb-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
          {{ successMessage }}
        </div>
        <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {{ errorMessage }}
        </div>

        <div class="grid grid-cols-1 gap-6 xl:grid-cols-3">
          <div class="xl:col-span-1 space-y-6">
            <section class="rounded-xl bg-white border border-theme-border shadow-sm overflow-hidden">
              <form @submit.prevent="saveName" class="block">
                <div class="px-6 py-4 border-b border-theme-border bg-primary-50/50">
                  <h2 class="text-base font-semibold text-theme-text">{{ $t('workspacePage.workspaceName') }}</h2>
                  <p class="text-xs text-theme-textLight mt-0.5">{{ $t('workspacePage.workspaceNameDesc') }}</p>
                </div>
                <div class="p-6 space-y-5">
                  <div>
                    <label for="workspace-name" class="block text-sm font-medium text-theme-text mb-1.5">{{ $t('workspacePage.name') }}</label>
                    <input
                      id="workspace-name"
                      v-model="editName"
                      type="text"
                      required
                      class="input-field max-w-md"
                      placeholder="Workspace name"
                      :disabled="nameSaving || !canManageWorkspace"
                    />
                  </div>
                  <button
                    v-if="canManageWorkspace"
                    type="submit"
                    class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
                    :disabled="nameSaving || !editName.trim() || editName.trim() === workspace.name"
                  >
                    {{ nameSaving ? $t('workspacePage.saving') : $t('common.save') }}
                  </button>
                </div>
              </form>
            </section>

            <!-- Danger zone: delete workspace -->
            <section v-if="canManageWorkspace" class="rounded-xl border border-red-200 bg-red-50/30 overflow-hidden">
              <div class="px-6 py-4 border-b border-red-200">
                <h2 class="text-lg font-semibold text-red-800">{{ $t('workspacePage.deleteWorkspace') }}</h2>
                <p class="text-sm text-red-700/90 mt-0.5">{{ $t('workspacePage.deleteWorkspaceDesc') }}</p>
              </div>
              <div class="px-6 py-4">
                <button
                  type="button"
                  @click="openDeleteModal"
                  class="rounded-lg border border-red-300 bg-white px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1"
                >
                  {{ $t('workspacePage.deleteWorkspaceButton') }}
                </button>
              </div>
            </section>
          </div>

          <!-- Right column: members (full width on mobile, 2/3 on xl) -->
          <section v-if="canManageWorkspace" class="xl:col-span-2 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="px-6 py-4 border-b border-theme-border flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">{{ $t('workspacePage.members') }}</h2>
                <p class="text-sm text-theme-textLight mt-0.5">{{ $t('workspacePage.membersDesc') }}</p>
              </div>
              <button type="button" @click="openAddUserModal" class="btn-primary text-sm inline-flex items-center gap-2">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                {{ $t('workspacePage.addMember') }}
              </button>
            </div>
            <div v-if="usersLoading" class="p-8 text-center text-theme-textLight text-sm">{{ $t('workspacePage.loadingMembers') }}</div>
            <div v-else-if="workspaceUsers.length === 0" class="p-8 text-center text-theme-textLight text-sm">{{ $t('workspacePage.noMembers') }}</div>
            <template v-else>
              <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-theme-border">
                  <thead class="bg-theme-hover">
                    <tr>
                      <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">User</th>
                      <th class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider">Role</th>
                      <th class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider">Actions</th>
                    </tr>
                  </thead>
                  <tbody class="bg-white divide-y divide-theme-border">
                    <tr v-for="wu in workspaceUsers" :key="wu.id" class="hover:bg-theme-hover">
                      <td class="px-6 py-4 whitespace-nowrap">
                        <div>
                          <p class="text-sm font-medium text-theme-text">{{ wu.name || wu.email }}</p>
                          <p v-if="wu.email" class="text-xs text-theme-textLight">{{ wu.email }}</p>
                        </div>
                      </td>
                      <td class="px-6 py-4 whitespace-nowrap">
                        <span class="inline-flex rounded-full bg-primary-100 px-2.5 py-0.5 text-xs font-medium text-theme-text">
                          {{ wu.role }}
                        </span>
                      </td>
                      <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                        <button
                          v-if="canEditOrRemoveUser(wu)"
                          type="button"
                          @click="openEditRoleModal(wu)"
                          class="text-primary-600 hover:text-primary-700 hover:underline mr-3"
                        >
                          Edit role
                        </button>
                        <button
                          v-if="canEditOrRemoveUser(wu)"
                          type="button"
                          @click="openRemoveUserModal(wu)"
                          class="text-red-600 hover:text-red-700 hover:underline"
                        >
                          Remove
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <!-- Pagination -->
              <div
                v-if="membersTotal > membersLimit"
                class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between flex-wrap gap-2"
              >
                <p class="text-sm text-theme-textLight">
                  Showing {{ (membersPage - 1) * membersLimit + 1 }} to {{ Math.min(membersPage * membersLimit, membersTotal) }} of {{ membersTotal }} members
                </p>
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    @click="goToMembersPage(membersPage - 1)"
                    :disabled="membersPage <= 1 || usersLoading"
                    class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Previous
                  </button>
                  <button
                    type="button"
                    @click="goToMembersPage(membersPage + 1)"
                    :disabled="membersPage >= membersTotalPages || usersLoading"
                    class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Next
                  </button>
                </div>
              </div>
            </template>
          </section>
        </div>
      </template>
    </main>

    <!-- Add user modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="addUserModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="add-user-title"
          @click.self="addUserModalOpen = false"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-md p-6 max-h-[90vh] flex flex-col">
            <h2 id="add-user-title" class="text-lg font-semibold text-theme-text">Add member</h2>
            <p class="mt-1 text-sm text-theme-textLight">Choose a user and role for this workspace</p>
            <form @submit.prevent="addUser" class="mt-4 flex flex-col flex-1 min-h-0 space-y-4">
              <div>
                <label for="add-user-select" class="form-label">User</label>
                <select
                  id="add-user-select"
                  v-model="addUserUserId"
                  required
                  class="input-field w-full"
                >
                  <option value="">Select a user…</option>
                  <option v-for="u in allUsersList" :key="u.id" :value="u.id">{{ u.name || u.email }} ({{ u.email }})</option>
                </select>
                <p v-if="allUsersLoading" class="text-xs text-theme-textLight mt-1">Loading users…</p>
              </div>
              <div>
                <label for="add-user-role" class="form-label">Role</label>
                <select id="add-user-role" v-model="addUserRole" required class="input-field w-full">
                  <option value="readonly">Readonly</option>
                  <option value="admin">Admin</option>
                  <option value="owner">Owner</option>
                </select>
              </div>
              <div class="flex gap-3 justify-end pt-2">
                <button type="button" @click="addUserModalOpen = false" class="btn-secondary">Cancel</button>
                <button type="submit" :disabled="addUserSaving || !addUserUserId || !addUserRole" class="btn-primary">
                  {{ addUserSaving ? 'Adding…' : 'Add' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Edit role modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="editRoleModalUser"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="edit-role-title"
          @click.self="editRoleModalUser = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="edit-role-title" class="text-lg font-semibold text-theme-text">Edit role</h2>
            <p class="mt-1 text-sm text-theme-textLight">
              {{ editRoleModalUser?.name || editRoleModalUser?.email }}
            </p>
            <form @submit.prevent="saveRole" class="mt-4 space-y-4">
              <div>
                <label for="edit-role-select" class="form-label">Role</label>
                <select id="edit-role-select" v-model="editRoleValue" required class="input-field w-full">
                  <option value="readonly">Readonly</option>
                  <option value="admin">Admin</option>
                  <option value="owner">Owner</option>
                </select>
              </div>
              <div class="flex gap-3 justify-end">
                <button type="button" @click="editRoleModalUser = null" class="btn-secondary">Cancel</button>
                <button type="submit" :disabled="editRoleSaving" class="btn-primary">
                  {{ editRoleSaving ? 'Saving…' : 'Save' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Remove user modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="removeUserModalUser"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="remove-user-title"
          @click.self="removeUserModalUser = null"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="remove-user-title" class="text-lg font-semibold text-theme-text">Remove member</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              Remove <span class="font-medium text-theme-text">{{ removeUserModalUser?.name || removeUserModalUser?.email }}</span> from this workspace?
            </p>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" @click="removeUserModalUser = null" class="btn-secondary" :disabled="removeUserSaving">Cancel</button>
              <button type="button" @click="confirmRemoveUser" class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500" :disabled="removeUserSaving">
                {{ removeUserSaving ? 'Removing…' : 'Remove' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Delete workspace modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-show="deleteModalOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-workspace-title"
          @click.self="closeDeleteModal"
        >
          <div class="bg-white rounded-lg shadow-xl border border-theme-border w-full max-w-sm p-6">
            <h2 id="delete-workspace-title" class="text-lg font-semibold text-theme-text">Delete workspace</h2>
            <p class="mt-2 text-sm text-theme-textLight">
              Delete workspace <span class="font-medium text-theme-text">"{{ workspace?.name }}"</span>? This cannot be undone.
            </p>
            <div class="mt-4">
              <label for="delete-confirm-name" class="block text-sm font-medium text-theme-text mb-1.5">
                Type <span class="font-mono">{{ workspace?.name }}</span> to confirm
              </label>
              <input
                id="delete-confirm-name"
                v-model="deleteConfirmName"
                type="text"
                class="input-field"
                :placeholder="workspace?.name"
                autocomplete="off"
                :disabled="deleteSaving"
              />
            </div>
            <div class="mt-6 flex gap-3 justify-end">
              <button type="button" @click="closeDeleteModal" class="btn-secondary" :disabled="deleteSaving">Cancel</button>
              <button type="button" @click="confirmDelete" class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500" :disabled="deleteSaving || !canConfirmDelete">
                {{ deleteSaving ? 'Deleting…' : 'Delete' }}
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
import AppNav from '@/components/AppNav.vue'
import { workspaceAPI, workspaceUsersAPI } from '@/api'
import { loadWorkspaceFromStorage, saveWorkspaceToStorage, removeWorkspaceFromStorage } from '@/utils/storage'

const router = useRouter()
const workspace = ref(null)
const errorMessage = ref(null)
const successMessage = ref(null)

const editName = ref('')
const nameSaving = ref(false)

const workspaceUsers = ref([])
const usersLoading = ref(false)
const membersPage = ref(1)
const membersLimit = ref(10)
const membersTotal = ref(0)

const membersTotalPages = computed(() => Math.ceil((membersTotal.value || 0) / membersLimit.value) || 1)

const canManageWorkspace = computed(() => {
  const role = workspace.value?.role
  return role === 'admin' || role === 'owner'
})

async function saveName() {
  const name = editName.value?.trim()
  if (!workspace.value || !name || name === workspace.value.name) return
  nameSaving.value = true
  errorMessage.value = null

  try {
    const res = await workspaceAPI.update(workspace.value.id, { name })
    workspace.value = { ...workspace.value, ...res.data }
    saveWorkspaceToStorage(workspace.value)
    successMessage.value = 'Workspace name updated.'
    clearSuccessAfter()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to update workspace name'
  } finally {
    nameSaving.value = false
  }
}

async function loadWorkspaceUsers() {
  if (!workspace.value?.id) return
  usersLoading.value = true

  try {
    const offset = (membersPage.value - 1) * membersLimit.value
    const res = await workspaceUsersAPI.list(workspace.value.id, { limit: membersLimit.value, offset })
    workspaceUsers.value = res.data.users ?? []
    membersTotal.value = res.data._meta.total ?? 0
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load workspace users'
    workspaceUsers.value = []
    membersTotal.value = 0
  } finally {
    usersLoading.value = false
  }
}

function goToMembersPage(page) {
  const p = Math.max(1, Math.min(page, membersTotalPages.value))
  if (p === membersPage.value) return
  membersPage.value = p
  loadWorkspaceUsers()
}

function canEditOrRemoveUser(wu) {
  if (!canManageWorkspace.value) return false
  const myRole = workspace.value?.role
  const theirRole = wu.role
  if (myRole === 'owner') return theirRole !== 'owner' || wu.userId === currentUserId.value
  if (myRole === 'admin') return theirRole === 'readonly'
  return false
}

const currentUserId = ref(null)
function setCurrentUserId() {
  try {
    const user = JSON.parse(localStorage.getItem('finn_user') || '{}')
    currentUserId.value = user?.id ?? null
  } catch {
    currentUserId.value = null
  }
}

const addUserModalOpen = ref(false)
const addUserUserId = ref('')
const addUserRole = ref('readonly')
const addUserSaving = ref(false)
const allUsersList = ref([])
const allUsersLoading = ref(false)

async function loadAllUsers() {
  if (!workspace.value?.id) return
  allUsersLoading.value = true

  try {
    const res = await workspaceUsersAPI.listAllUsers(
      workspace.value.id,
      { excludeMembers: 'true', limit: 200 }
    )
    allUsersList.value = res.data.users
  } catch {
    allUsersList.value = []
  } finally {
    allUsersLoading.value = false
  }
}

watch(addUserModalOpen, (open) => {
  if (open) {
    addUserUserId.value = ''
    addUserRole.value = 'readonly'
    loadAllUsers()
  }
})

function openAddUserModal() {
  addUserModalOpen.value = true
}

async function addUser() {
  if (!workspace.value?.id || !addUserUserId.value || !addUserRole.value) return
  addUserSaving.value = true
  errorMessage.value = null

  try {
    await workspaceUsersAPI.add(workspace.value.id, { userId: addUserUserId.value, role: addUserRole.value })
    addUserModalOpen.value = false
    await loadWorkspaceUsers()
    successMessage.value = 'Member added.'
    clearSuccessAfter()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to add member'
  } finally {
    addUserSaving.value = false
  }
}

const editRoleModalUser = ref(null)
const editRoleValue = ref('readonly')
const editRoleSaving = ref(false)

function openEditRoleModal(wu) {
  editRoleModalUser.value = wu
  editRoleValue.value = wu.role || 'readonly'
}

async function saveRole() {
  const wu = editRoleModalUser.value
  if (!workspace.value?.id || !wu?.userId) return
  editRoleSaving.value = true
  errorMessage.value = null

  try {
    await workspaceUsersAPI.updateRole(workspace.value.id, wu.userId, { role: editRoleValue.value })
    editRoleModalUser.value = null
    await loadWorkspaceUsers()
    successMessage.value = 'Role updated.'
    clearSuccessAfter()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to update role'
  } finally {
    editRoleSaving.value = false
  }
}

const removeUserModalUser = ref(null)
const removeUserSaving = ref(false)

function openRemoveUserModal(wu) {
  removeUserModalUser.value = wu
}

async function confirmRemoveUser() {
  const wu = removeUserModalUser.value
  if (!workspace.value?.id || !wu?.userId) return
  removeUserSaving.value = true
  errorMessage.value = null

  try {
    await workspaceUsersAPI.remove(workspace.value.id, wu.userId)
    removeUserModalUser.value = null
    await loadWorkspaceUsers()
    successMessage.value = 'Member removed.'
    clearSuccessAfter()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to remove member'
  } finally {
    removeUserSaving.value = false
  }
}

const deleteModalOpen = ref(false)
const deleteConfirmName = ref('')
const deleteSaving = ref(false)
const canConfirmDelete = computed(() => workspace.value?.name && deleteConfirmName.value.trim() === workspace.value.name)

function openDeleteModal() {
  deleteModalOpen.value = true
  deleteConfirmName.value = ''
}

function closeDeleteModal() {
  if (!deleteSaving.value) {
    deleteModalOpen.value = false
    deleteConfirmName.value = ''
  }
}

async function confirmDelete() {
  if (!workspace.value || !canConfirmDelete.value) return
  deleteSaving.value = true
  errorMessage.value = null

  try {
    await workspaceAPI.delete(workspace.value.id)
    if (loadWorkspaceFromStorage()?.id === workspace.value.id) {
      removeWorkspaceFromStorage()
    }
    closeDeleteModal()
    router.push('/select-workspace')
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to delete workspace'
  } finally {
    deleteSaving.value = false
  }
}

function clearSuccessAfter() {
  setTimeout(() => { successMessage.value = null }, 4000)
}

watch(workspace, (w) => {
  if (w?.name != null) editName.value = w.name
}, { immediate: true })

onMounted(async () => {
  setCurrentUserId()
  const stored = loadWorkspaceFromStorage()
  if (!stored?.id) {
    router.replace('/select-workspace')
    return
  }
  workspace.value = { ...stored }
  errorMessage.value = null

  try {
    const res = await workspaceAPI.get(stored.id)
    workspace.value = { ...workspace.value, ...res.data }
    saveWorkspaceToStorage(workspace.value)
  } catch (err) {
    if (err.response?.status === 403 || err.response?.status === 404) {
      removeWorkspaceFromStorage()
      router.replace('/select-workspace')
      return
    }
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load workspace'
  }
  await loadWorkspaceUsers()
})
</script>
