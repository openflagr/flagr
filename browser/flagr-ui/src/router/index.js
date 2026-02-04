import { createRouter, createWebHashHistory } from 'vue-router'
import Flags from '@/components/Flags'
import Flag from '@/components/Flag'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
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
})
