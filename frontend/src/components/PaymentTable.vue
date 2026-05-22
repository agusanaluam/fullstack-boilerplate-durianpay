<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Payment } from '@/stores/payments'

const props = defineProps<{
  payments: Payment[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'update:filterStatus', val: string): void
  (e: 'update:filterSort', val: string): void
  (e: 'update:filterSearch', val: string): void
}>()

const search = ref('')
const statusFilter = ref('')
const sortFilter = ref('-created_at')

const PAGE_SIZE = 10
const currentPage = ref(1)

const displayed = computed(() => {
  const s = search.value.toLowerCase()
  let list = props.payments
  if (s) list = list.filter(p => p.id.toLowerCase().includes(s) || p.merchant.toLowerCase().includes(s))
  const start = (currentPage.value - 1) * PAGE_SIZE
  return list.slice(start, start + PAGE_SIZE)
})

const totalPages = computed(() => {
  const s = search.value.toLowerCase()
  let list = props.payments
  if (s) list = list.filter(p => p.id.toLowerCase().includes(s) || p.merchant.toLowerCase().includes(s))
  return Math.max(1, Math.ceil(list.length / PAGE_SIZE))
})

function onStatusChange(val: string) {
  statusFilter.value = val
  currentPage.value = 1
  emit('update:filterStatus', val)
}

function onSortChange(val: string) {
  sortFilter.value = val
  emit('update:filterSort', val)
}

function onSearch(val: string) {
  search.value = val
  currentPage.value = 1
  emit('update:filterSearch', val)
}

const statusBadge: Record<string, string> = {
  completed: 'bg-emerald-100 text-emerald-700',
  processing: 'bg-amber-100 text-amber-700',
  failed: 'bg-red-100 text-red-700',
}

function formatAmount(amount: number) {
  return `Rp ${amount.toLocaleString('id-ID')}`
}

function formatDate(dt: string) {
  return new Date(dt).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<template>
  <div class="bg-white rounded-xl border border-slate-200 overflow-hidden">
    <!-- Header + Filters -->
    <div class="px-5 py-4 border-b border-slate-100 flex items-center justify-between gap-4">
      <h2 class="text-sm font-bold text-slate-900">All Transactions</h2>
      <div class="flex gap-2 items-center">
        <input
          :value="search"
          @input="onSearch(($event.target as HTMLInputElement).value)"
          placeholder="Search by ID or merchant..."
          class="border border-slate-200 rounded-lg px-3 py-1.5 text-xs w-48 outline-none focus:border-indigo-400"
          style="font-family: Arial, sans-serif"
        />
        <select
          :value="statusFilter"
          @change="onStatusChange(($event.target as HTMLSelectElement).value)"
          class="border border-slate-200 rounded-lg px-2 py-1.5 text-xs outline-none appearance-none pr-6"
          style="font-family: Arial, sans-serif"
        >
          <option value="">All Status</option>
          <option value="completed">Completed</option>
          <option value="processing">Processing</option>
          <option value="failed">Failed</option>
        </select>
        <select
          :value="sortFilter"
          @change="onSortChange(($event.target as HTMLSelectElement).value)"
          class="border border-slate-200 rounded-lg px-2 py-1.5 text-xs outline-none appearance-none pr-6"
          style="font-family: Arial, sans-serif"
        >
          <option value="-created_at">Newest First</option>
          <option value="created_at">Oldest First</option>
          <option value="-amount">Amount ↓</option>
          <option value="amount">Amount ↑</option>
        </select>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="py-16 text-center text-sm text-slate-400">Loading...</div>

    <!-- Table -->
    <table v-else class="w-full border-collapse">
      <thead>
        <tr class="bg-slate-50 border-b border-slate-100">
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Payment ID</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Merchant</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Date</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Amount</th>
          <th class="px-5 py-3 text-left text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in displayed" :key="p.id" class="border-b border-slate-50 hover:bg-slate-50/50">
          <td class="px-5 py-3 text-xs text-indigo-600 font-bold">{{ p.id }}</td>
          <td class="px-5 py-3 text-xs font-semibold text-slate-900">{{ p.merchant }}</td>
          <td class="px-5 py-3 text-xs text-slate-500">{{ formatDate(p.created_at) }}</td>
          <td class="px-5 py-3 text-xs font-bold text-slate-900">{{ formatAmount(p.amount) }}</td>
          <td class="px-5 py-3">
            <span :class="[statusBadge[p.status], 'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-bold']">
              <span class="w-1.5 h-1.5 rounded-full bg-current"></span>
              {{ p.status }}
            </span>
          </td>
        </tr>
        <tr v-if="displayed.length === 0">
          <td colspan="5" class="px-5 py-12 text-center text-sm text-slate-400">No payments found</td>
        </tr>
      </tbody>
    </table>

    <!-- Pagination -->
    <div class="px-5 py-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-400">
      <span>Showing {{ displayed.length }} of {{ payments.length }} payments</span>
      <div class="flex gap-1">
        <button
          v-for="page in totalPages"
          :key="page"
          @click="currentPage = page"
          :class="['w-7 h-7 rounded-lg border text-xs flex items-center justify-center',
            page === currentPage
              ? 'bg-indigo-600 text-white border-indigo-600'
              : 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50']"
        >{{ page }}</button>
      </div>
    </div>
  </div>
</template>
