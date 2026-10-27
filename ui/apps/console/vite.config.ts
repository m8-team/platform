import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/resource-manager/v1': 'http://127.0.0.1:8080',
      // Resource Manager serves probes separately (M8_HEALTH_HTTP_ADDR; default :8081).
      '/health/resource-manager': {
        target: process.env.M8_RESOURCE_MANAGER_HEALTH_PROXY_TARGET || 'http://127.0.0.1:8081',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/health\/resource-manager(?=\/|$)/, ''),
      },
    },
  },
})
