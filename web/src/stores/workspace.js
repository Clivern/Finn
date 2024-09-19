import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { loadWorkspaceFromStorage, saveWorkspaceToStorage } from '@/utils/storage'
import { workspaceAPI } from '@/api'

export const useWorkspaceStore = defineStore('workspace', () => {
  const current = ref(loadWorkspaceFromStorage())
  const list = ref([])
  const loading = ref(false)
  const error = ref(null)

  const currentWorkspace = computed(() => current.value)
  const workspaces = computed(() => list.value)

  async function fetchWorkspaces() {
    loading.value = true
    error.value = null
    try {
      const response = await workspaceAPI.list()
      list.value = response.data?.workspaces ?? []
      if (list.value.length === 0) {
        list.value = getDefaultWorkspaces()
      }
      return list.value
    } catch (err) {
      error.value = err.response?.data?.errorMessage || err.message || 'Failed to load workspaces'
      list.value = getDefaultWorkspaces()
      return list.value
    } finally {
      loading.value = false
    }
  }

  function getDefaultWorkspaces() {
    return [
      { id: 'default-1', name: 'Default', createdAt: new Date().toISOString(), status: 'active' }
    ]
  }

  function setCurrentWorkspace(workspace) {
    current.value = workspace ?? null
    saveWorkspaceToStorage(current.value)
  }

  async function createWorkspace(name) {
    loading.value = true
    error.value = null
    try {
      const response = await workspaceAPI.create({ name })
      const created = response.data?.workspace
      if (created) {
        list.value = [...list.value, created]
        setCurrentWorkspace(created)
        return { success: true, workspace: created }
      }
      throw new Error('Invalid response')
    } catch (err) {
      error.value = err.response?.data?.errorMessage || err.message || 'Failed to create workspace'
      const localWorkspace = {
        id: `local-${Date.now()}`,
        name: name.trim(),
        createdAt: new Date().toISOString(),
        status: 'active'
      }
      list.value = [...list.value, localWorkspace]
      setCurrentWorkspace(localWorkspace)
      return { success: true, workspace: localWorkspace }
    } finally {
      loading.value = false
    }
  }

  return {
    current,
    list,
    loading,
    error,
    currentWorkspace,
    workspaces,
    fetchWorkspaces,
    setCurrentWorkspace,
    createWorkspace
  }
})
