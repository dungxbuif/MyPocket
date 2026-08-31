import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const apiProxyTarget = process.env.API_PROXY_TARGET ?? "http://127.0.0.1:8080";

export default defineConfig({
  plugins: [tailwindcss(), react()],
  server: {
    // Keep local development on one origin so OAuth redirects and cookies
    // return to the frontend while the Go API runs on port 8080.
    proxy: {
      "/api": {
        target: apiProxyTarget,
        changeOrigin: true,
      },
      "/docs": {
        target: apiProxyTarget,
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    exclude: ["docs/**", "e2e/**", "node_modules/**"],
  },
});
