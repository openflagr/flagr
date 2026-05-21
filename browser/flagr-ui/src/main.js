import { createApp } from 'vue'

import ElementPlus from 'element-plus'
import locale from 'element-plus/dist/locale/en.mjs'
import 'element-plus/theme-chalk/index.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(ElementPlus, { locale })
app.use(router)

// Autofocus certain fields
app.directive('focus', {
  mounted(el) {
    const input = el.querySelector('input') || el.querySelector('textarea')
    if (input) input.focus()
  }
})

app.mount('#app')
