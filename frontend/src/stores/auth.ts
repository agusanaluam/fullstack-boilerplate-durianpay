import { defineStore } from 'pinia'
import { ref } from 'vue'
import { postDashboardV1AuthLogin } from '@/api/generated'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const role = ref<string | null>(null)
  const email = ref<string | null>(null)

  async function login(emailInput: string, password: string) {
    const result = await postDashboardV1AuthLogin({
      body: { email: emailInput, password },
    })
    const data = (result as any).data
    const error = (result as any).error
    if (error || !data?.token) {
      throw new Error(error?.message ?? 'Login failed')
    }
    token.value = data.token
    role.value = data.role ?? null
    email.value = data.email ?? emailInput
  }

  function logout() {
    token.value = null
    role.value = null
    email.value = null
  }

  return { token, role, email, login, logout }
}, {
  persist: true,
})
