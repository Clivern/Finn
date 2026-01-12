<template>
  <div class="rounded-lg border border-theme-border bg-theme-hover/20 p-4">
    <div class="flex flex-wrap items-center justify-between gap-2 mb-3">
      <div class="flex items-center gap-2">
        <h3 class="text-sm font-semibold text-theme-text">{{ title }}</h3>
        <button
          v-if="help"
          type="button"
          class="text-theme-textLight hover:text-theme-text rounded p-0.5"
          :title="help"
          :aria-label="help"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </button>
      </div>
      <div class="flex items-center gap-2">
        <div class="inline-flex rounded-md border border-theme-border bg-white text-xs overflow-hidden">
          <button
            type="button"
            class="px-2.5 py-1 font-medium"
            :class="viewMode === 'table' ? 'bg-theme-hover text-theme-text' : 'text-theme-textLight hover:bg-theme-hover/50'"
            @click="switchToTable"
          >
            {{ $t('groupVarsPage.viewTable') }}
          </button>
          <button
            type="button"
            class="px-2.5 py-1 font-medium border-l border-theme-border"
            :class="viewMode === 'json' ? 'bg-theme-hover text-theme-text' : 'text-theme-textLight hover:bg-theme-hover/50'"
            @click="switchToJson"
          >
            {{ $t('groupVarsPage.viewJson') }}
          </button>
        </div>
        <button
          v-if="viewMode === 'table'"
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
    </div>

    <div v-if="viewMode === 'table'" class="space-y-2">
      <p v-if="rows.length === 0" class="text-sm text-theme-textLight py-2">{{ $t('groupVarsPage.noValues') }}</p>
      <template v-else>
        <div class="grid grid-cols-[1fr_1fr_auto] gap-2 text-xs font-medium text-theme-textLight uppercase tracking-wide">
          <span>{{ $t('groupVarsPage.key') }}</span>
          <span>{{ $t('groupVarsPage.value') }}</span>
          <span class="sr-only">{{ $t('common.delete') }}</span>
        </div>
        <div v-for="(row, idx) in rows" :key="idx" class="grid grid-cols-[1fr_1fr_auto] gap-2 items-center">
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

    <div v-else>
      <textarea
        :value="jsonText"
        rows="10"
        class="input-field font-mono text-sm w-full"
        spellcheck="false"
        @input="onJsonInput($event.target.value)"
      />
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  title: { type: String, required: true },
  help: { type: String, default: '' },
  viewMode: { type: String, required: true },
  rows: { type: Array, required: true },
  jsonText: { type: String, required: true },
  valueSecret: { type: Boolean, default: false },
})

const emit = defineEmits(['update:viewMode', 'update:rows', 'update:jsonText'])

function rowsToObject(rows) {
  const o = {}
  for (const r of rows) {
    const k = r.key?.trim()
    if (!k) continue
    o[k] = r.value ?? ''
  }
  return o
}

function valueToDisplay(v) {
  if (v === null || v === undefined) return ''
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

function jsonObjectToRows(raw) {
  if (!raw || !String(raw).trim()) return []
  try {
    const o = JSON.parse(String(raw))
    if (o === null || typeof o !== 'object' || Array.isArray(o)) return []
    return Object.keys(o).map((k) => ({ key: k, value: valueToDisplay(o[k]) }))
  } catch {
    return []
  }
}

function switchToJson() {
  if (props.viewMode === 'json') return
  const pretty = JSON.stringify(rowsToObject(props.rows), null, 2)
  emit('update:jsonText', pretty)
  emit('update:viewMode', 'json')
}

function switchToTable() {
  if (props.viewMode === 'table') return
  emit('update:rows', jsonObjectToRows(props.jsonText))
  emit('update:viewMode', 'table')
}

function addRow() {
  const next = [...props.rows, { key: '', value: '' }]
  emit('update:rows', next)
}

function removeRow(idx) {
  const next = props.rows.filter((_, i) => i !== idx)
  emit('update:rows', next)
}

function emitRows() {
  emit('update:rows', [...props.rows])
}

function onJsonInput(v) {
  emit('update:jsonText', v)
}
</script>
