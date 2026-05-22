import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getDashboardV1Payments } from '@/api/generated'
import { useAuthStore } from './auth'

export interface Payment {
  id: string
  merchant: string
  amount: number
  status: 'completed' | 'processing' | 'failed'
  created_at: string
}

export const usePaymentsStore = defineStore('payments', () => {
  const payments = ref<Payment[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filterStatus = ref('')
  const filterSort = ref('-created_at')
  const filterSearch = ref('')

  const summary = computed(() => ({
    total: payments.value.length,
    completed: payments.value.filter(p => p.status === 'completed').length,
    processing: payments.value.filter(p => p.status === 'processing').length,
    failed: payments.value.filter(p => p.status === 'failed').length,
  }))

  const filteredPayments = computed(() => {
    const search = filterSearch.value.toLowerCase()
    if (!search) return payments.value
    return payments.value.filter(
      p => p.id.toLowerCase().includes(search) || p.merchant.toLowerCase().includes(search)
    )
  })

  async function fetchPayments() {
    const authStore = useAuthStore()
    loading.value = true
    error.value = null
    try {
      const result = await getDashboardV1Payments({
        query: {
          status: filterStatus.value || undefined,
          sort: filterSort.value || undefined,
        },
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      const data = (result as any).data
      const apiError = (result as any).error
      if (apiError) throw new Error('Failed to fetch payments')
      payments.value = (data?.payments ?? []) as Payment[]
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  return { payments, loading, error, filterStatus, filterSort, filterSearch, summary, filteredPayments, fetchPayments }
})
