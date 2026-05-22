<script setup lang="ts">
import { onMounted, watch } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'
import SummaryCards from '@/components/SummaryCards.vue'
import PaymentTable from '@/components/PaymentTable.vue'
import { usePaymentsStore } from '@/stores/payments'

const paymentsStore = usePaymentsStore()

onMounted(() => paymentsStore.fetchPayments())

watch([() => paymentsStore.filterStatus, () => paymentsStore.filterSort], () => {
  paymentsStore.fetchPayments()
})

const today = new Date().toLocaleDateString('en-US', { weekday: 'short', day: 'numeric', month: 'long', year: 'numeric' })
</script>

<template>
  <div class="flex h-screen bg-slate-50" style="font-family: Arial, sans-serif">
    <AppSidebar />

    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Topbar -->
      <div class="bg-white border-b border-slate-200 px-7 py-4 flex items-center justify-between flex-shrink-0">
        <div>
          <h1 class="text-base font-bold text-slate-900">Payment Monitor</h1>
          <div class="text-xs text-slate-400 mt-0.5">Showing all incoming transactions</div>
        </div>
        <div class="text-xs text-slate-400">{{ today }}</div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-7 flex flex-col gap-5">
        <SummaryCards :summary="paymentsStore.summary" />

        <PaymentTable
          :payments="paymentsStore.filteredPayments"
          :loading="paymentsStore.loading"
          v-model:filterStatus="paymentsStore.filterStatus"
          v-model:filterSort="paymentsStore.filterSort"
          v-model:filterSearch="paymentsStore.filterSearch"
        />
      </div>
    </div>
  </div>
</template>
