<template>
  <div v-if="aiEnabled" class="fixed bottom-6 right-6 z-40">
    <button
      type="button"
      @click="open = true"
      class="flex h-12 w-12 items-center justify-center rounded-full bg-primary-600 text-white shadow-lg hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors"
      :aria-label="$t('aiAssistant.openLabel')"
    >
      <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
      </svg>
    </button>

    <Teleport to="body">
      <Transition name="sidebar">
        <div
          v-show="open"
          class="fixed inset-0 z-50"
          role="dialog"
          aria-modal="true"
          aria-labelledby="ai-assistant-title"
        >
          <!-- Backdrop -->
          <div
            class="absolute inset-0 bg-black/30"
            aria-hidden="true"
            @click="open = false"
          />
          <!-- Sidebar panel -->
          <div class="ai-assistant-panel absolute right-0 top-6 bottom-6 w-full max-w-md rounded-xl bg-white shadow-xl border border-theme-border flex flex-col">
            <!-- Header -->
            <div class="shrink-0 p-4 border-b border-theme-border flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-primary-100 text-primary-600">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
                  </svg>
                </span>
                <h2 id="ai-assistant-title" class="text-lg font-semibold text-theme-text">
                  {{ $t('aiAssistant.title') }}
                </h2>
              </div>
              <button
                type="button"
                @click="open = false"
                class="text-theme-textLight hover:text-theme-text p-1.5 rounded transition-colors"
                :aria-label="$t('common.close')"
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Chat messages -->
            <div
              ref="messagesEl"
              class="flex-1 overflow-y-auto p-4 space-y-4 min-h-0 flex flex-col"
            >
              <div v-if="loadingMessages" class="flex justify-center py-8 text-theme-textLight text-sm">
                {{ $t('aiAssistant.loading') }}
              </div>
              <div
                v-else-if="messages.length === 0"
                class="flex flex-1 flex-col items-center justify-center text-center px-2 py-6 min-h-[12rem]"
              >
                <div
                  class="mb-5 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-50 text-primary-600 ring-1 ring-primary-100"
                  aria-hidden="true"
                >
                  <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM16.5 7.5h.008v.008H16.5V7.5z" />
                  </svg>
                </div>
                <h3 class="text-base font-semibold text-theme-text mb-2">
                  {{ $t('aiAssistant.emptyTitle') }}
                </h3>
                <p class="text-sm text-theme-textLight max-w-sm leading-relaxed mb-6">
                  {{ $t('aiAssistant.description') }}
                </p>
                <p class="text-xs font-medium text-theme-textLight uppercase tracking-wide mb-2 w-full max-w-sm text-left">
                  {{ $t('aiAssistant.tryAsking') }}
                </p>
                <div class="flex w-full max-w-sm flex-col gap-2">
                  <button
                    v-for="(prompt, idx) in starterPrompts"
                    :key="idx"
                    type="button"
                    class="rounded-lg border border-theme-border bg-white px-3 py-2.5 text-left text-sm text-theme-text transition-colors hover:bg-theme-hover hover:border-theme-border focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-1"
                    @click="inputText = prompt"
                  >
                    {{ prompt }}
                  </button>
                </div>
              </div>
              <template v-else>
                <div
                  v-for="msg in messages"
                  :key="msg.id"
                  :class="[
                    'flex',
                    msg.role === 'user' ? 'justify-end' : 'justify-start'
                  ]"
                >
                  <div
                    :class="[
                      'max-w-[85%] rounded-2xl px-4 py-2.5 text-sm whitespace-pre-wrap',
                      msg.role === 'user'
                        ? 'bg-primary-600 text-white rounded-br-md'
                        : 'bg-theme-hover text-theme-text rounded-bl-md'
                    ]"
                  >
                    {{ msg.content }}
                  </div>
                </div>
              </template>
              <div v-if="sending" class="flex justify-start">
                <div class="max-w-[85%] rounded-2xl rounded-bl-md px-4 py-2.5 text-sm bg-theme-hover text-theme-textLight">
                  {{ $t('aiAssistant.thinking') }}
                </div>
              </div>
            </div>

            <!-- Error message -->
            <div v-if="errorMessage" class="shrink-0 px-4 pb-2 text-sm text-red-600">
              {{ errorMessage }}
            </div>

            <!-- Input -->
            <div class="shrink-0 p-4 border-t border-theme-border">
              <form @submit.prevent="send" class="flex gap-2">
                <input
                  v-model="inputText"
                  type="text"
                  class="input-field flex-1 min-w-0"
                  :placeholder="$t('aiAssistant.placeholder')"
                  :disabled="sending"
                  autocomplete="off"
                />
                <button
                  type="submit"
                  class="btn-primary shrink-0 px-4"
                  :disabled="sending || !inputText.trim()"
                  :aria-label="$t('aiAssistant.send')"
                >
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                  </svg>
                </button>
              </form>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { aiAPI } from '@/api'

const { t } = useI18n()

const starterPrompts = computed(() => [
  t('aiAssistant.suggestion1'),
  t('aiAssistant.suggestion2'),
  t('aiAssistant.suggestion3')
])

const open = ref(false)
const inputText = ref('')
const messagesEl = ref(null)
const aiEnabled = ref(false)
const messages = ref([])
const loadingMessages = ref(false)
const sending = ref(false)
const errorMessage = ref('')

async function fetchStatus() {
  try {
    const res = await aiAPI.getStatus()
    aiEnabled.value = res.data?.enabled === true
  } catch (_) {
    aiEnabled.value = false
  }
}

async function loadMessages() {
  if (!aiEnabled.value) return
  loadingMessages.value = true
  errorMessage.value = ''
  try {
    const res = await aiAPI.getMessages({ limit: 50, offset: 0 })
    messages.value = res.data?.messages ?? []
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to load messages'
    messages.value = []
  } finally {
    loadingMessages.value = false
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

async function send() {
  const content = inputText.value.trim()
  if (!content || sending.value) return

  errorMessage.value = ''
  sending.value = true
  inputText.value = ''

  // Optimistically add user message (API persists it and returns assistant reply)
  const userMsg = {
    id: `temp-${Date.now()}`,
    role: 'user',
    content,
    createdAt: new Date().toISOString()
  }
  messages.value.push(userMsg)
  scrollToBottom()

  try {
    const res = await aiAPI.createMessage({ content })
    messages.value.push({
      id: res.data?.id,
      role: res.data?.role || 'assistant',
      content: res.data?.content || '',
      createdAt: res.data?.createdAt
    })
    scrollToBottom()
  } catch (err) {
    errorMessage.value = err.response?.data?.errorMessage || 'Failed to get response'
    messages.value.pop()
  } finally {
    sending.value = false
  }
}

onMounted(fetchStatus)

watch(open, async (isOpen) => {
  if (isOpen) {
    await loadMessages()
    scrollToBottom()
  }
})
</script>

<style scoped>
.sidebar-enter-active,
.sidebar-leave-active {
  transition: opacity 0.2s ease;
}
.sidebar-enter-active .ai-assistant-panel,
.sidebar-leave-active .ai-assistant-panel {
  transition: transform 0.25s ease;
}
.sidebar-enter-from,
.sidebar-leave-to {
  opacity: 0;
}
.sidebar-enter-from .ai-assistant-panel,
.sidebar-leave-to .ai-assistant-panel {
  transform: translateX(100%);
}
</style>
