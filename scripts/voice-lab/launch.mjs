// Disposable tunnel + lab lifecycle. No production ports or database access.
import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";

if (!process.env.ELEVENLABS_API_KEY || !process.env.MISTRAL_API_KEY) {
  console.error("Set ELEVENLABS_API_KEY and MISTRAL_API_KEY in the server environment (never in browser code)."); process.exit(1);
}
const basicEnv = Object.fromEntries(["PATH", "HOME", "TMPDIR"].filter(key => process.env[key]).map(key => [key, process.env[key]]));
const cwd = fileURLToPath(new URL(".", import.meta.url));
const tunnel = spawn(process.env.VOICE_LAB_TUNNEL_BIN ?? "cloudflared", ["tunnel", "--url", "http://127.0.0.1:18082", "--no-autoupdate"], { cwd, env: basicEnv, stdio: ["ignore", "ignore", "pipe"] });
let server;
let starting = false;
let stopped = false;
let logs = "";
const startupDeadline = setTimeout(() => { if (!starting) { console.error("Tunnel startup timed out."); stop(); } }, 45_000);

function stop() {
  if (stopped) return; stopped = true; clearTimeout(startupDeadline);
  if (server && server.exitCode === null) {
    server.kill("SIGTERM");
    server.once("exit", () => tunnel.kill("SIGTERM"));
  } else tunnel.kill("SIGTERM");
}
tunnel.stderr.on("data", chunk => {
  // Read only the generated public URL; don't repeat general environment/log output.
  logs = (logs + chunk.toString()).slice(-8000);
  const url = logs.match(/https:\/\/[a-z0-9-]+\.trycloudflare\.com/);
  if (!url || starting || stopped) return;
  starting = true; clearTimeout(startupDeadline);
  server = spawn(process.execPath, ["server.mjs"], {
    cwd,
    env: { ...basicEnv, ELEVENLABS_API_KEY: process.env.ELEVENLABS_API_KEY, MISTRAL_API_KEY: process.env.MISTRAL_API_KEY, VOICE_LAB_CALLBACK_URL: `${url[0].replace("https:", "wss:")}/ws` },
    stdio: ["ignore", "inherit", "inherit"]
  });
  server.on("error", () => { console.error("Cannot start the voice lab."); stop(); });
  server.on("exit", code => { process.exitCode = code ?? 0; stop(); });
});
tunnel.on("error", () => { console.error("Cannot run cloudflared. Install it or set VOICE_LAB_TUNNEL_BIN to its executable path."); process.exitCode = 1; stop(); });
tunnel.on("exit", () => { if (!stopped) { console.error("Tunnel closed; stopping lab."); stop(); } });
process.once("SIGINT", stop); process.once("SIGTERM", stop);
