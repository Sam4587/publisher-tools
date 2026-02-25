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
      // 热点监控API代理到Node.js服务 (3001端口)
      '/api/hot-topics': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      // 平台账号管理API代理到Node.js服务 (3001端口)
      '/api/v1/publisher': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      // AI内容生成API代理到Node.js服务 (3001端口)
      '/api/v1/ai': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,
      },
      // 其他API请求代理到Go服务 (8080端口)
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
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
