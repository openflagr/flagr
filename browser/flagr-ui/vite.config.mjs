import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'url'

const srcDir = fileURLToPath(new URL('src', import.meta.url))

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': srcDir
    },
    extensions: ['.mjs', '.js', '.vue', '.json']
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: 'modern-compiler'
      }
    }
  },
  server: {
    port: 8080,
    watch: {
      usePolling: true,
      interval: 1000,
    },
    proxy: {
      '/api/v1': {
        target: 'http://127.0.0.1:18000',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'static',
    chunkSizeWarningLimit: 1500,
  }
})
