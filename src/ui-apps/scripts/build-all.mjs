import { build } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

const entries = [
  "containers.html",
  "images.html",
  "volumes.html",
  "networks.html",
  "compose.html",
  "disk-usage.html",
  "system-info.html"
];

async function run() {
  for (let i = 0; i < entries.length; i += 1) {
    const entry = entries[i];
    // Build each app separately because vite-plugin-singlefile disables code splitting.
    await build({
      plugins: [viteSingleFile()],
      build: {
        emptyOutDir: i === 0,
        outDir: "dist",
        rollupOptions: {
          input: entry
        }
      }
    });
  }
}

run().catch((err) => {
  console.error(err);
  process.exit(1);
});
