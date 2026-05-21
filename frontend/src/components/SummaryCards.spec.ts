import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import SummaryCards from './SummaryCards.vue'

describe('SummaryCards', () => {
  const summary = { total: 50, completed: 30, processing: 12, failed: 8 }

  it('renders all four counts', () => {
    const wrapper = mount(SummaryCards, { props: { summary } })
    expect(wrapper.text()).toContain('50')
    expect(wrapper.text()).toContain('30')
    expect(wrapper.text()).toContain('12')
    expect(wrapper.text()).toContain('8')
  })

  it('renders correct labels', () => {
    const wrapper = mount(SummaryCards, { props: { summary } })
    expect(wrapper.text()).toContain('Total Payments')
    expect(wrapper.text()).toContain('Completed')
    expect(wrapper.text()).toContain('Processing')
    expect(wrapper.text()).toContain('Failed')
  })
})
