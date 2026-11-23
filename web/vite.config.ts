import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: ["icon.svg", "apple-touch-icon.png", "icon-192.png", "icon-512.png"],
      manifest: {
        name: "EnglishCraft",
        short_name: "EnglishCraft",
        description: "Ton aventure pour progresser en anglais",
        theme_color: "#183e33",
        background_color: "#f7f8f4",
        display: "standalone",
        start_url: "/",
        icons: [
          { src: "/icon-192.png", sizes: "192x192", type: "image/png", purpose: "any" },
          { src: "/icon-512.png", sizes: "512x512", type: "image/png", purpose: "any maskable" },
        ],
      },
    }),
  ],
  build: {
    outDir: "../internal/learning/adapters/httpapi/web/dist",
    emptyOutDir: true,
  },
  server: {
    // Preserve Host so the API's Origin check also protects Vite development.
    proxy: { "/api": { target: process.env.API_PROXY_TARGET ?? "http://localhost:8080", changeOrigin: false } },
  },
  test: {
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    exclude: ["e2e/**", "node_modules/**"],
  },
});
