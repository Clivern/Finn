<template>
  <div class="rounded-lg border border-theme-border bg-theme-hover/20 p-4">
    <div class="flex flex-wrap items-center justify-between gap-2 mb-3">
      <h3 class="text-sm font-semibold text-theme-text">{{ title }}</h3>
      <button
        type="button"
        class="p-1 rounded text-primary-600 hover:bg-primary-50"
        :title="$t('groupVarsPage.addRow')"
        :aria-label="$t('groupVarsPage.addRow')"
        @click="addRow"
      >
        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    </div>

    <div v-if="rows.length === 0" class="text-sm text-theme-textLight py-2">{{ $t('groupVarsPage.noValues') }}</div>
    <template v-else>
      <div class="grid grid-cols-[1fr_1fr_auto] gap-2 text-xs font-medium text-theme-textLight uppercase tracking-wide mb-2">
        <span>{{ $t('groupVarsPage.key') }}</span>
        <span>{{ $t('groupVarsPage.value') }}</span>
        <span class="sr-only">{{ $t('common.delete') }}</span>
      </div>
      <div v-for="(row, idx) in rows" :key="idx" class="grid grid-cols-[1fr_1fr_auto] gap-2 items-center mb-2 last:mb-0">
        <input v-model="row.key" type="text" class="input-field text-sm" :placeholder="$t('groupVarsPage.key')" @input="emitRows" />
        <input v-model="row.value" :type="valueSecret ? 'password' : 'text'" class="input-field text-sm font-mono" :placeholder="$t('groupVarsPage.value')" @input="emitRows" />
        <button type="button" class="p-1 text-red-600 hover:bg-red-50 rounded shrink-0" :aria-label="$t('common.delete')" @click="removeRow(idx)">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </template>
  </div>
</template>

<script setup>
const props = defineProps({
  title: { type: String, required: true },
  rows: { type: Array, required: true },
  valueSecret: { type: Boolean, default: false },
})

const emit = defineEmits(['update:rows'])

function addRow() {
  emit('update:rows', [...props.rows, { key: '', value: '' }])
}

function removeRow(idx) {
  emit('update:rows', props.rows.filter((_, i) => i !== idx))
}

function emitRows() {
  emit('update:rows', [...props.rows])
}
</script>
