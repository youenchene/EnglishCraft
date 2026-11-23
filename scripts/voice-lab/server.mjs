// Temporary, standalone adult voice test. Never mounted in the production app.
import { createServer } from "node:http";
import { readFile, writeFile, unlink } from "node:fs/promises";
import { ElevenLabsClient } from "@elevenlabs/elevenlabs-js";
import { localRequest, messages, spokenText, MAX_SECONDS, MAX_TURNS, MODEL } from "./policy.mjs";

const port = 18082;
const callback = process.env.VOICE_LAB_CALLBACK_URL;
if (!process.env.ELEVENLABS_API_KEY || !process.env.MISTRAL_API_KEY || !callback?.startsWith("wss://")) {
  console.error("Required: ELEVENLABS_API_KEY, MISTRAL_API_KEY and VOICE_LAB_CALLBACK_URL=wss://…/ws"); process.exit(1);
}
const elevenlabs = new ElevenLabsClient({ apiKey: process.env.ELEVENLABS_API_KEY });
let admitted = false;
let expectedConversation;
let currentSession;
let attachment;
let engine;
let closing = false;
let totalInput = 0;
let totalOutput = 0;
const timers = new Set();
function json(response, status, data) {
  response.writeHead(status, { "Content-Type": "application/json", "Cache-Control": "no-store" });
  response.end(JSON.stringify(data));
}
const server = createServer(async (request, response) => {
  // The public tunnel exposes only the authenticated /ws callback, not the UI/token endpoint.
  if (!localRequest(request, port)) { response.writeHead(404); response.end(); return; }
  try {
    if (request.method === "POST" && request.url === "/session") {
      if (admitted || closing) { json(response, 409, { message: "Un seul test est autorisé. Relance le serveur pour réessayer." }); return; }
      admitted = true;
      const token = await elevenlabs.conversationalAi.conversations.getWebrtcToken({ agentId: engine.engineId });
      expectedConversation = token.conversationId;
      json(response, 200, { token: token.token, maxSeconds: MAX_SECONDS });
      // A minted but unused token must not leave the lab open indefinitely.
      const timeout = setTimeout(() => void shutdown(), 120_000); timers.add(timeout);
      return;
    }
    const assets = { "/": ["index.html", "text/html"], "/client.js": ["dist/client.js", "text/javascript"], "/style.css": ["style.css", "text/css"] };
    const asset = assets[request.url];
    if (request.method !== "GET" || !asset) { response.writeHead(404); response.end(); return; }
    response.writeHead(200, { "Content-Type": asset[1], "Cache-Control": "no-store", "X-Content-Type-Options": "nosniff", "Permissions-Policy": "microphone=(self)", "Content-Security-Policy": "default-src 'self'; connect-src 'self' https: wss:; media-src 'self' blob:; frame-ancestors 'none'" });
    response.end(await readFile(new URL(asset[0], import.meta.url)));
  } catch { json(response, 502, { message: "Le fournisseur n’a pas accepté la session. Vérifie les permissions et le solde du compte." }); }
});

async function shutdown() {
  if (closing) return; closing = true;
  for (const timer of timers) clearTimeout(timer);
  currentSession?.close();
  await attachment?.close();
  server.close();
  if (engine) {
    try { await elevenlabs.speechEngine.delete(engine.engineId); console.log("Temporary Speech Engine deleted."); }
    catch { console.error("Temporary engine cleanup failed; delete the EnglishCraft voice lab resource in ElevenLabs."); }
  }
  await unlink(new URL(".runtime.json", import.meta.url)).catch(() => {});
  console.log(`Mistral token usage: input=${totalInput}, output=${totalOutput}. No transcripts logged.`);
}

try {
  engine = await elevenlabs.speechEngine.create({
    name: "EnglishCraft adult voice lab (temporary)",
    speechEngine: { wsUrl: callback },
    language: "en",
    conversation: { maxDurationSeconds: MAX_SECONDS, clientEvents: ["audio", "interruption", "agent_response", "user_transcript"] },
    privacy: { recordVoice: false, retentionDays: 1, deleteAudio: true, deleteTranscriptAndPii: true },
    callLimits: { agentConcurrencyLimit: 1, dailyLimit: 1, burstingEnabled: false },
    turn: { silenceEndCallTimeout: 20, speculativeTurn: false }
  });
  let turns = 0;
  attachment = engine.attach(server, "/ws", {
    debug: false,
    onInit(conversationId, session) {
      if (!admitted || conversationId !== expectedConversation || currentSession) { session.close(); return; }
      currentSession = session;
      const timeout = setTimeout(() => session.close(), MAX_SECONDS * 1000); timers.add(timeout);
      console.log("Voice session connected; maximum 60 seconds.");
    },
    async onTranscript(transcript, signal, session) {
      if (session !== currentSession || closing || ++turns > MAX_TURNS) { session.close(); return; }
      try {
        const response = await fetch("https://api.mistral.ai/v1/chat/completions", {
          method: "POST", signal: AbortSignal.any([signal, AbortSignal.timeout(15_000)]),
          headers: { Authorization: `Bearer ${process.env.MISTRAL_API_KEY}`, "Content-Type": "application/json" },
          body: JSON.stringify({ model: MODEL, messages: messages(transcript), reasoning_effort: "none", max_tokens: 120, temperature: .4 })
        });
        if (!response.ok) { console.error(`Mistral response HTTP ${response.status}`); session.close(); return; }
        const result = await response.json();
        totalInput += result.usage?.prompt_tokens ?? 0; totalOutput += result.usage?.completion_tokens ?? 0;
        const text = spokenText(result.choices?.[0]?.message?.content);
        if (typeof text !== "string" || !text.trim()) { session.close(); return; }
        if (!signal.aborted && session.isOpen) session.sendResponse(text.slice(0, 1200));
      } catch { if (!signal.aborted) { console.error("Mistral call failed; transcript omitted."); session.close(); } }
    },
    onClose(session) { if (session === currentSession) { console.log("Voice session closed."); void shutdown(); } },
    onDisconnect(session) { if (session === currentSession) { console.log("Voice session disconnected."); void shutdown(); } },
    onError() { console.error("Speech Engine error; details omitted to protect secrets and transcripts."); }
  });
  const runtime = { engineId: engine.engineId, callback, model: MODEL, port, pid: process.pid, maxSeconds: MAX_SECONDS };
  await writeFile(new URL(".runtime.json", import.meta.url), JSON.stringify(runtime), { mode: 0o600 });
  server.listen(port, "127.0.0.1", () => console.log(`Voice lab ready: http://127.0.0.1:${port}/ — ${MODEL}`));
  // Stop this disposable server even if the browser never starts a session.
  timers.add(setTimeout(() => void shutdown(), 20 * 60_000));
  process.once("SIGINT", () => void shutdown()); process.once("SIGTERM", () => void shutdown());
} catch (error) {
  console.error(`Speech Engine setup failed (HTTP ${error?.statusCode ?? "unknown"}); provider details omitted.`);
  await shutdown(); process.exitCode = 1;
}
