import { afterEach, expect, test, vi } from "vitest";
import { VoiceConnection } from "./voice";
const sdk = vi.hoisted(() => ({ startSession: vi.fn(), endSession: vi.fn(), setMicMuted: vi.fn() }));
vi.mock("@elevenlabs/client", () => ({ Conversation: { startSession: sdk.startSession } }));

afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });
test("checks server availability before asking a student for microphone access", async () => {
  const getUserMedia = vi.fn().mockRejectedValue(new Error("Micro demandé malgré une voix désactivée"));
  Object.defineProperty(navigator, "mediaDevices", { configurable: true, value: { getUserMedia } });
  vi.stubGlobal("RTCPeerConnection", vi.fn());
  vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ error: { message: "Voix non configurée sur le serveur." } }), { status: 503 }));
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => {});
  const connection = new VoiceConnection("diamond", false, { state: vi.fn(), caption: vi.fn(), error: vi.fn(), events: vi.fn(), changed: async () => {} }, document.createElement("audio"));
  await expect(connection.start()).rejects.toThrow("Voix non configurée sur le serveur.");
  expect(getUserMedia).not.toHaveBeenCalled();
});

test("ElevenLabs uses a memory-only token, frees permission stream, and stops its owned server session", async () => {
  sdk.startSession.mockReset(); sdk.endSession.mockReset(); sdk.setMicMuted.mockReset();
  sdk.endSession.mockResolvedValue(undefined);
  sdk.startSession.mockResolvedValue({ endSession: sdk.endSession, setMicMuted: sdk.setMicMuted });
  const stop = vi.fn();
  Object.defineProperty(navigator, "mediaDevices", { configurable: true, value: { getUserMedia: vi.fn().mockResolvedValue({ getTracks: () => [{ stop }] }) } });
  vi.stubGlobal("RTCPeerConnection", vi.fn());
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => {});
  const fetch = vi.spyOn(globalThis, "fetch")
    .mockResolvedValueOnce(new Response(JSON.stringify({ available: true, provider: "elevenlabs" })))
    .mockResolvedValueOnce(new Response(JSON.stringify({ id: "conv_owned", provider: "elevenlabs", token: "ephemeral", greeting: "Hello! Let's explore the village.", expires: new Date(Date.now()+60000).toISOString() }), { status: 201 }))
    .mockResolvedValue(new Response(null, { status: 204 }));
  const setItem = vi.spyOn(Storage.prototype, "setItem");
  const connection = new VoiceConnection("diamond", true, { state: vi.fn(), caption: vi.fn(), error: vi.fn(), events: vi.fn(), changed: async () => {} }, document.createElement("audio"));
  expect(await connection.start()).toBe(true);
  expect(stop).toHaveBeenCalledOnce();
  expect(sdk.startSession).toHaveBeenCalledWith(expect.objectContaining({ conversationToken: "ephemeral", connectionType: "webrtc", overrides: { agent: { firstMessage: "Hello! Let's explore the village." } } }));
  expect(setItem).not.toHaveBeenCalled();
  expect(sdk.setMicMuted).toHaveBeenCalledWith(true);
  connection.talk(true); expect(sdk.setMicMuted).toHaveBeenLastCalledWith(false);
  await connection.stop(); await connection.stop();
  expect(sdk.endSession).toHaveBeenCalledOnce();
  expect(fetch).toHaveBeenLastCalledWith("/api/v1/tutor/voice/conv_owned", expect.objectContaining({ method: "DELETE", credentials: "same-origin" }));
});
