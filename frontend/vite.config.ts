import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/dummyLogin': 'http://localhost:8080',
      '/register': 'http://localhost:8080',
      '/login': 'http://localhost:8080',
      '/rooms': 'http://localhost:8080',
      '/bookings': 'http://localhost:8080',
      '/_info': 'http://localhost:8080',
    },
  },
})
