import { createApp } from 'vue' // Use createApp for Vue 3
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css' // Import Element Plus styles
import en from 'element-plus/es/locale/lang/en'
import store from './store'; // Ensure you are importing the store correctly
import App from './App.vue'
import router from './router'
import { setupAxiosInstance, setupflagrAxiosInstance } from './utils/apiUtil'

const app = createApp(App) // Create the Vue app instance

app.config.globalProperties.$ELEMENT = { locale: en } // Set locale for ElementPlus

// Autofocus certain fields
app.directive('focus', {
  mounted(el) { // `mounted` is used in Vue 3 instead of `inserted`
    el.focus()
  }
})

// Use ElementPlus and router
app.use(ElementPlus, { locale: en })
app.use(router)
app.use(store);

function initializeApp() {
  console.log("App is initialized!");
  setupAxiosInstance('https://bff.allen-stage.in/');
  setupflagrAxiosInstance();
  // You can do other things here like fetching initial data, setting up listeners, etc.
  // e.g., check if the user is authenticated, load initial data, etc.
  if (localStorage.getItem('tokens')) {
    console.log("User is already authenticated.");
  } else {
    console.log("User is not authenticated.");
  }
}

// Call the function once during app initialization
initializeApp();
// Mount the app to the DOM
app.mount('#app')