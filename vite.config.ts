import path from "node:path"
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  root: "resources/js",
  base: "/dashboard/",
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "resources/js/src"),
    },
  },
  build: {
    outDir: "../../public/dashboard",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("recharts") || id.includes("d3-") || id.includes("victory-vendor")) return "charts"
          if (id.includes("@radix-ui")) return "radix"
          if (id.includes("react")) return "react"
        },
      },
    },
  },
})