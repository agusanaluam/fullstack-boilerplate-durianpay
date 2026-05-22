<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(email.value, password.value)
    router.push({ name: 'dashboard' })
  } catch {
    error.value = 'Invalid email or password. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex h-screen" style="font-family: Arial, sans-serif">
    <!-- Left panel -->
    <div class="w-[45%] flex-shrink-0 bg-gradient-to-br from-indigo-500 via-purple-500 to-purple-400 flex flex-col justify-center items-start p-12 relative overflow-hidden">
      <div class="absolute -top-16 -right-16 w-72 h-72 rounded-full bg-white/[0.07]"></div>
      <div class="absolute -bottom-20 -left-10 w-80 h-80 rounded-full bg-white/[0.05]"></div>
      <div class="relative z-10">
        <div class="w-12 h-12 bg-white rounded-xl flex items-center justify-center text-indigo-600 font-black text-xl mb-8">P</div>
        <h1 class="text-3xl font-bold text-white leading-snug mb-4">Payment<br>Monitor Dashboard</h1>
        <p class="text-white/70 text-sm leading-relaxed max-w-xs">
          Internal tool for monitoring and tracking all incoming payment transactions in real time.
        </p>
        <div class="flex gap-8 mt-10">
          <div>
            <div class="text-white text-2xl font-bold">50+</div>
            <div class="text-white/60 text-xs mt-0.5">Transactions</div>
          </div>
          <div>
            <div class="text-white text-2xl font-bold">2</div>
            <div class="text-white/60 text-xs mt-0.5">User Roles</div>
          </div>
          <div>
            <div class="text-white text-2xl font-bold">99%</div>
            <div class="text-white/60 text-xs mt-0.5">Uptime</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Right form -->
    <div class="flex-1 flex items-center justify-center bg-slate-50 p-8">
      <div class="bg-white rounded-2xl border border-slate-200 p-10 w-full max-w-sm">
        <h2 class="text-xl font-bold text-slate-900 mb-1">Welcome back</h2>
        <p class="text-slate-400 text-sm mb-8">Sign in to your account to continue</p>

        <!-- Error -->
        <div v-if="error" class="mb-4 bg-red-50 border border-red-200 rounded-lg px-4 py-3 text-xs text-red-600 flex items-center gap-2">
          ⚠ {{ error }}
        </div>

        <form @submit.prevent="handleLogin" class="flex flex-col gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Email address</label>
            <input
              v-model="email"
              type="email"
              placeholder="cs@test.com"
              required
              class="w-full border border-slate-200 rounded-lg px-3 py-2.5 text-sm outline-none focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
              style="font-family: Arial, sans-serif"
            />
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Password</label>
            <div class="relative">
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="••••••••"
                required
                class="w-full border border-slate-200 rounded-lg px-3 py-2.5 text-sm outline-none focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 pr-10"
                style="font-family: Arial, sans-serif"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 text-sm"
              >{{ showPassword ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-bold py-2.5 rounded-lg mt-1 disabled:opacity-50"
            style="font-family: Arial, sans-serif"
          >
            {{ loading ? 'Signing in...' : 'Sign in →' }}
          </button>
        </form>

        <div class="flex items-center gap-3 my-5">
          <hr class="flex-1 border-slate-100" /><span class="text-[11px] text-slate-300">test accounts</span><hr class="flex-1 border-slate-100" />
        </div>
        <div class="bg-slate-50 border border-slate-200 rounded-lg px-4 py-3 text-[11px] text-slate-500 leading-relaxed">
          <strong class="text-slate-700">cs@test.com</strong> / password — CS role<br>
          <strong class="text-slate-700">operation@test.com</strong> / password — Operation role
        </div>

        <div class="mt-6 text-center text-[11px] text-slate-400">DurianPay Internal</div>
      </div>
    </div>
  </div>
</template>
