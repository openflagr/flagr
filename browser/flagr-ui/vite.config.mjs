import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'url'

const srcDir = fileURLToPath(new URL('src', import.meta.url))

export default defineConfig({
  // Relative base so the built assets load when Flagr is served under a
  // path prefix (FLAGR_WEB_PREFIX), e.g. https://example.com/flagr/.
  // Safe because the router uses hash history, so the document URL always
  // points at the app root.
  base: './',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': srcDir
    },
    extensions: ['.mjs', '.js', '.ts', '.vue', '.json']
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
