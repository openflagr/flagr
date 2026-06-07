import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/components/Flags.vue')
  },
  {
    path: '/flags/:flagId',
    name: 'flag',
    component: () => import('@/components/Flag.vue')
  }
]

export default createRouter({
  history: createWebHashHistory(),
  routes
})
