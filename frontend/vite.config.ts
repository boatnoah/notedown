// vite.config.ts
import { defineConfig } from "vite";

export default defineConfig({
  server: {
    proxy: {
      // WebSockets
      "/ws": {
        target: "ws://localhost:3000",
        ws: true,
      },
      // HTTP (OAuth, APIs, etc.)
      "/auth": {
        target: "http://localhost:3000",
        changeOrigin: true,
      },
    },
  },
});
