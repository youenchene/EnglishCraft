// Creates/updates only the explicitly named Speech Engine; never logs API keys.
const key = process.env.ELEVENLABS_API_KEY;
const callback = process.env.ELEVENLABS_CALLBACK_URL;
const seconds = Number(process.env.VOICE_MAX_SECONDS ?? 300);
const existing = process.env.ELEVENLABS_ENGINE_ID;
if (!key || !callback?.startsWith("wss://") || !callback.endsWith("/api/v1/tutor/voice/elevenlabs/upstream") || !Number.isInteger(seconds) || seconds < 1 || seconds > 1200 || (existing && !/^seng_[A-Za-z0-9_-]+$/.test(existing))) {
  console.error("Required: ELEVENLABS_API_KEY and ELEVENLABS_CALLBACK_URL=wss://PUBLIC_HOST/api/v1/tutor/voice/elevenlabs/upstream; optional VOICE_MAX_SECONDS and existing ELEVENLABS_ENGINE_ID."); process.exit(1);
}
const response = await fetch(`https://api.elevenlabs.io/v1/speech-engine${existing ? `/${existing}` : ""}`, {
  method: existing ? "PATCH" : "POST",
  headers: { "xi-api-key": key, "Content-Type": "application/json" },
  body: JSON.stringify({
    name: "EnglishCraft Alex",
    speech_engine: { ws_url: callback },
    language: "en",
    overrides: { first_message: true },
    conversation: { max_duration_seconds: seconds, client_events: ["audio", "interruption", "agent_response", "user_transcript"] },
    privacy: { record_voice: false, retention_days: 1, delete_audio: true, delete_transcript_and_pii: true },
    call_limits: { agent_concurrency_limit: 2, daily_limit: 50, bursting_enabled: false },
    turn: { silence_end_call_timeout: 30, speculative_turn: false }
  })
});
if (!response.ok) { console.error(`Speech Engine configuration failed: HTTP ${response.status}. Provider details omitted.`); process.exit(1); }
const result = await response.json();
if (!/^seng_[A-Za-z0-9_-]+$/.test(result.speech_engine_id)) { console.error("Invalid Speech Engine response."); process.exit(1); }
console.log(`ELEVENLABS_ENGINE_ID=${result.speech_engine_id}`);
