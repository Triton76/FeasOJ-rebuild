import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'
import compression from 'vite-plugin-compression';

export default defineConfig({
  base: './',
  server: {
    host: '0.0.0.0',
    port: 5173,
    cors: true,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8082',
        changeOrigin: true,
      },
    },
  },
  plugins: [
    compression({
      algorithm: 'brotliCompress',
      threshold: 10240
    }),
    vue(),
    vuetify({
      autoImport: true,
    }),
  ]
})
