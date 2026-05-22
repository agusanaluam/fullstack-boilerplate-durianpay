<script setup lang="ts">
defineProps<{
  summary: {
    total: number
    completed: number
    processing: number
    failed: number
  }
}>()

const cards = [
  { key: 'total',      label: 'Total Payments', icon: '💳', bg: 'bg-violet-50',  color: 'text-slate-900' },
  { key: 'completed',  label: 'Completed',       icon: '✅', bg: 'bg-emerald-50', color: 'text-emerald-600' },
  { key: 'processing', label: 'Processing',      icon: '⏳', bg: 'bg-amber-50',   color: 'text-amber-600' },
  { key: 'failed',     label: 'Failed',          icon: '❌', bg: 'bg-red-50',     color: 'text-red-600' },
] as const
</script>

<template>
  <div class="grid grid-cols-4 gap-4">
    <div
      v-for="card in cards"
      :key="card.key"
      class="bg-white rounded-xl border border-slate-200 p-4 flex items-start gap-3"
    >
      <div :class="[card.bg, 'w-10 h-10 rounded-xl flex items-center justify-center text-lg flex-shrink-0']">
        {{ card.icon }}
      </div>
      <div>
        <div :class="[card.color, 'text-2xl font-bold leading-none']">{{ summary[card.key] }}</div>
        <div class="text-xs text-slate-400 mt-1">{{ card.label }}</div>
      </div>
    </div>
  </div>
</template>
