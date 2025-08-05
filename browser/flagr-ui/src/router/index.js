import Vue from 'vue'
import Router from 'vue-router'
import Flags from '@/components/Flags'
import Flag from '@/components/Flag'
import Webhooks from '@/components/Webhooks'

Vue.use(Router)

export default new Router({
  mode: 'hash',
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
    },
    {
      path: '/webhooks',
      name: 'webhooks',
      component: Webhooks
    }
  ]
})
