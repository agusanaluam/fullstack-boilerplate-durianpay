<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentsStore } from '@/stores/payments'

const router = useRouter()
const authStore = useAuthStore()
const paymentsStore = usePaymentsStore()

function logout() {
  authStore.logout()
  router.push({ name: 'login' })
}

const initials = authStore.role?.slice(0, 2).toUpperCase() ?? 'U'
</script>

<template>
  <aside class="w-[220px] flex-shrink-0 bg-white border-r border-slate-200 flex flex-col h-screen">
    <!-- Brand -->
    <div class="p-5 border-b border-slate-100 flex items-center gap-3">
      <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white font-bold text-sm">
        P
      </div>
      <div>
        <div class="font-bold text-sm text-slate-900">PayDash</div>
        <div class="text-[11px] text-slate-400">Internal Dashboard</div>
      </div>
    </div>

    <!-- Nav -->
    <nav class="flex-1 p-3 flex flex-col gap-1">
      <div class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider px-3 mb-1 mt-2">Main</div>
      <router-link
        to="/dashboard"
        class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold text-indigo-600 bg-indigo-50"
      >
        <span>📊</span>
        <span>Payments</span>
        <span class="ml-auto bg-indigo-600 text-white text-[10px] font-bold px-2 py-0.5 rounded-full">
          {{ paymentsStore.summary.total }}
        </span>
      </router-link>
    </nav>

    <!-- User -->
    <div class="p-4 border-t border-slate-100">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-cyan-500 flex items-center justify-center text-white text-[11px] font-bold flex-shrink-0">
          {{ initials }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-xs font-bold text-slate-900 truncate">{{ authStore.email }}</div>
          <div class="text-[10px] text-slate-400 capitalize">{{ authStore.role }} role</div>
        </div>
        <button @click="logout" class="text-slate-400 hover:text-slate-600 text-sm">↩</button>
      </div>
    </div>
  </aside>
</template>
