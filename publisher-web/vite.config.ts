import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    host: true,
    allowedHosts: ['.monkeycode-ai.online'],
    proxy: {
      '/api/v1/publisher': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      '/api/platforms': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      '/api/tasks': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      '/api/publish': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      '/api/health': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      }
    },
    open: false,
  },
  preview: {
    port: 4173,
    host: true,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
