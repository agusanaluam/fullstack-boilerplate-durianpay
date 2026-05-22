import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PaymentTable from './PaymentTable.vue'
import type { Payment } from '@/stores/payments'

const mockPayments: Payment[] = [
  { id: 'PAY-001', merchant: 'Tokopedia', amount: 150000, status: 'completed', created_at: '2025-05-01 10:00:00' },
  { id: 'PAY-002', merchant: 'Shopee',    amount: 75000,  status: 'processing', created_at: '2025-05-02 11:00:00' },
  { id: 'PAY-003', merchant: 'Gojek',     amount: 50000,  status: 'failed',    created_at: '2025-05-03 12:00:00' },
]

describe('PaymentTable', () => {
  it('renders all payment rows', () => {
    const wrapper = mount(PaymentTable, {
      props: { payments: mockPayments, loading: false },
    })
    expect(wrapper.text()).toContain('PAY-001')
    expect(wrapper.text()).toContain('PAY-002')
    expect(wrapper.text()).toContain('PAY-003')
  })

  it('renders merchant names', () => {
    const wrapper = mount(PaymentTable, {
      props: { payments: mockPayments, loading: false },
    })
    expect(wrapper.text()).toContain('Tokopedia')
    expect(wrapper.text()).toContain('Shopee')
  })

  it('shows loading state', () => {
    const wrapper = mount(PaymentTable, {
      props: { payments: [], loading: true },
    })
    expect(wrapper.text()).toContain('Loading')
  })
})
