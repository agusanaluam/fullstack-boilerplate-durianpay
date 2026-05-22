import { setActivePinia, createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useAuthStore } from './auth'

vi.mock('@/api/generated', () => ({
  postDashboardV1AuthLogin: vi.fn(),
}))

import { postDashboardV1AuthLogin } from '@/api/generated'

describe('authStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('login stores token and role on success', async () => {
    vi.mocked(postDashboardV1AuthLogin).mockResolvedValueOnce({
      data: { token: 'jwt-token-123', role: 'cs', email: 'cs@test.com' },
      error: undefined,
    } as any)

    const store = useAuthStore()
    await store.login('cs@test.com', 'password')

    expect(store.token).toBe('jwt-token-123')
    expect(store.role).toBe('cs')
  })

  it('logout clears token and role', async () => {
    const store = useAuthStore()
    store.token = 'some-token'
    store.role = 'cs'
    store.logout()

    expect(store.token).toBeNull()
    expect(store.role).toBeNull()
  })

  it('login throws on API error', async () => {
    vi.mocked(postDashboardV1AuthLogin).mockResolvedValueOnce({
      data: undefined,
      error: { message: 'invalid credentials' },
    } as any)

    const store = useAuthStore()
    await expect(store.login('bad@test.com', 'wrong')).rejects.toThrow()
  })
})
