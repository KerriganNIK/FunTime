import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';
import { getLanOrigins } from '../scripts/dev-network.mjs';

export default defineConfig(({ command, mode }) => ({
  plugins: [react()],
  define: { 'import.meta.env.VITE_INVITE_ORIGIN': JSON.stringify(loadEnv(mode, fileURLToPath(new URL('.', import.meta.url))).VITE_INVITE_ORIGIN || (command === 'serve' ? getLanOrigins()[0] ?? '' : '')) },
  resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    proxy: { '/api': { target: process.env.GATEWAY_URL || 'http://127.0.0.1:8080' } },
  },
}));
