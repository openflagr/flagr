import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import './styles/element/index.scss'
import App from './App.vue'
import router from '@/helpers/router'

const app = createApp(App)

app.use(ElementPlus)
app.use(router)

app.directive('focus', {
  mounted(el: HTMLElement) {
    const input = el.querySelector('input') ?? el.querySelector('textarea')
    if (input) input.focus()
  },
})

app.mount('#app')