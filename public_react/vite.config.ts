import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    hmr: {
      host: "localhost",
      protocol: "ws",
    },
    proxy: {
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
      "/file": {
        target: "http://localhost:8080",
      },
    },
  },
});
