import { createRouter, createWebHashHistory } from 'vue-router'
import Flags from '@/components/Flags'
import Flag from '@/components/Flag'

const routes = [
  {
    path: '/',
    name: 'home',
    component: Flags
  },
  {
    path: '/flags/:flagId',
    name: 'flag',
    component: Flag
  }
]

export default createRouter({
  history: createWebHashHistory(),
  routes
})
