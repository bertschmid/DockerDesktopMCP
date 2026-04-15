import { defineConfig } from "vite";

export default defineConfig({
  plugins: [],
  build: {
    outDir: "dist",
    rollupOptions: {
      input: {
        containers: "containers.html",
        images: "images.html",
        volumes: "volumes.html",
        networks: "networks.html",
        compose: "compose.html",
        "disk-usage": "disk-usage.html",
        "system-info": "system-info.html"
      }
    }
  }
});