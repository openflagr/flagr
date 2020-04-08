// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'

import ElementUI from 'element-ui'
import locale from 'element-ui/lib/locale/lang/en'
import 'element-ui/lib/theme-chalk/index.css'

import App from './App.vue'
import router from './router'

Vue.config.productionTip = false
Vue.use(ElementUI, { locale })

// Autofocus certain fields
Vue.directive('focus', {
  inserted: function (el) {
    el.__vue__.focus()
  }
})

/* eslint-disable no-new */

new Vue({
  render: h => h(App),
  router
}).$mount('#app')
