import { createRouter, createWebHistory } from 'vue-router'
import Flags from '@/components/Flags'
// import Flags2 from '@/components/Flags2'

import Flag from '@/components/Flag'
import Login from '@/components/Login'
import Callback from '@/components/Callback'
import { isAuthenticated } from '../utils/apiUtil'

// Create the router instance using Vue 3's createRouter function
const router = createRouter({
  history: createWebHistory(), // Use hash mode similar to Vue 2
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: Login,
    },
    {
      path: '/callback',
      name: 'Callback',
      component: Callback,
    },
    {
      path: '/',
      name: 'home',
      component: Flags,
      meta: { requiresAuth: true },
    },
    {
      path: '/flags',
      name: 'Flags',
      component: Flags,
      meta: { requiresAuth: true },
    },
    {
      path: '/flags/:flagId',
      name: 'flag',
      component: Flag,
      meta: { requiresAuth: true },
    },
    {
      path: '/:catchAll(.*)',
      redirect: '/login', // Redirect unknown routes to login
      meta: { requiresAuth: true },
    }
  ]
})

router.beforeEach((to, from, next) => {
  console.log("to from", to, from, to.meta.requiresAuth, isAuthenticated())
  if (to.meta.requiresAuth && !isAuthenticated()) {
    console.log("sdss")
    next({ path: '/login', replace: true });
  } else {
    next();
  }
});

export default router