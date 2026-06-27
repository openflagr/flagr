import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import Flags from '@/components/Flags.vue'
import Flag from '@/components/Flag.vue'

export type AppRouteName = 'home' | 'flag'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: Flags },
  { path: '/flags/:flagId', name: 'flag', component: Flag },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})