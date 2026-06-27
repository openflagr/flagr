import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'url'

const srcDir = fileURLToPath(new URL('src', import.meta.url))

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': srcDir },
    extensions: ['.mjs', '.js', '.ts', '.vue', '.json'],
  },
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts'],
  },
})