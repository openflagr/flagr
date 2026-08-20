import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import './styles/element/index.scss'
import App from './App.vue'
import router from '@/helpers/router'
import { initServerMode } from '@/helpers/serverMode'

/**
 * Upper bound on waiting for /health before first paint. Past it the app
 * mounts anyway and stays fail-open (editable UI; backend 403s are the
 * backstop for read-only deployments).
 */
const SERVER_MODE_TIMEOUT_MS = 1500

const app = createApp(App)

app.use(ElementPlus)
app.use(router)

app.directive('focus', {
  mounted(el: HTMLElement) {
    const input = el.querySelector('input') ?? el.querySelector('textarea')
    if (input) input.focus()
  },
})

async function bootstrap(): Promise<void> {
  // Resolve the server mode before first paint so a read-only (eval-only)
  // deployment never flashes editable controls. Aborting the health fetch
  // at the bound fail-opens and prevents a late response from flipping
  // the read source after the first paint already chose a path.
  const ac = new AbortController()
  const timer = window.setTimeout(() => ac.abort(), SERVER_MODE_TIMEOUT_MS)
  try {
    await initServerMode(ac.signal)
  } finally {
    window.clearTimeout(timer)
  }
  app.mount('#app')
}

void bootstrap()