import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// ✅ Vite configuration
export default defineConfig({
  plugins: [react()],

  // 🚀 Development server configuration
  server: {
    port: 5173, // Keep your dev frontend on this port
    strictPort: true, // Prevent random fallback ports
    host: "localhost",

    // 🔄 Fix "WebSocket closed before established" (Vite HMR interference)
    hmr: {
      protocol: "ws",
      host: "localhost",
      overlay: false, // disable error overlay (optional but cleaner)
    },

    // ✅ Allow API and WebSocket requests to backend on :9090
    proxy: {
      "/ws": {
        target: "ws://localhost:9090",
        ws: true,
      },
      "/leaderboard": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
    },
  },

  // ⚙️ Build config (optional, helps with deployment)
  build: {
    outDir: "dist",
    sourcemap: true,
  },
});
