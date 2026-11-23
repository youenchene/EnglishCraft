import { api } from "./coreApi";
import { AlexState, GameEvent, tool, VoiceSession } from "./tutorApi";
import type { Conversation } from "@elevenlabs/client";

type VoiceEvent = { type: string; delta?: string; transcript?: string; name?: string; arguments?: string; call_id?: string };
type Callbacks = { state: (state: AlexState) => void; caption: (text: string) => void;
  error: (text: string) => void; events: (events: GameEvent[]) => void; changed: () => Promise<void> };

// Audio and captions remain in memory. No recording or provider key is stored.
export class VoiceConnection {
  provider: "openai" | "elevenlabs" = "openai";
  private eleven?: Conversation;
  private peer?: RTCPeerConnection;
  private channel?: RTCDataChannel;
  private microphone?: MediaStream;
  private session?: VoiceSession;
  private timer?: ReturnType<typeof setTimeout>;
  private stopped = false;
  private calls = new Map<string, Promise<unknown>>();
  private captions = "";
  private audio: HTMLAudioElement;
  constructor(private mission: string, private pushToTalk: boolean, private callbacks: Callbacks, audio: HTMLAudioElement) {
    this.audio = audio;
  }
  async start() {
    if (!navigator.onLine) throw new Error("La voix cloud est indisponible hors ligne.");
    if (!navigator.mediaDevices?.getUserMedia || !globalThis.RTCPeerConnection) throw new Error("Ce navigateur ne permet pas la voix. Utilise le mode texte.");
    this.callbacks.state("thinking");
    try {
      const ready = await api<{ provider?: string }>(`/tutor/${encodeURIComponent(this.mission)}/voice`);
      if (this.stopped) return false;
      const stream = await navigator.mediaDevices.getUserMedia({ audio: { echoCancellation: true, noiseSuppression: true } });
      if (this.stopped) { stream.getTracks().forEach(t => t.stop()); return false; }
      if (ready.provider === "elevenlabs") {
        // Validate permission before spending admission budget. The SDK owns
        // its own microphone stream; don't leave a second capture running.
        stream.getTracks().forEach(t => t.stop());
        this.provider = "elevenlabs";
        await this.startEleven();
        return !this.stopped;
      }
      this.microphone = stream;
      const peer = new RTCPeerConnection(); this.peer = peer;
      this.audio.autoplay = true;
      peer.ontrack = event => {
        this.audio.srcObject = new MediaStream([event.track]);
        void this.audio.play().catch(() => this.callbacks.error("Appuie sur Lecture pour écouter Alex (autorisation audio Safari)."));
      };
      for (const track of stream.getAudioTracks()) { track.enabled = !this.pushToTalk; peer.addTrack(track, stream); }
      peer.onconnectionstatechange = () => {
        if (peer.connectionState === "failed" || peer.connectionState === "disconnected") {
          this.callbacks.error("Connexion vocale perdue. Tu peux continuer en texte."); void this.stop();
        }
      };
      const channel = peer.createDataChannel("oai-events"); this.channel = channel;
      channel.onmessage = event => { void this.receive(event.data); };
      channel.onclose = () => { if (!this.stopped) { this.callbacks.error("La session vocale est terminée."); void this.stop(); } };
      channel.onopen = () => { this.callbacks.state("listening"); this.send({ type: "response.create" }); };
      await peer.setLocalDescription(await peer.createOffer());
      if (peer.iceGatheringState !== "complete") {
        await new Promise<void>((resolve, reject) => {
          const timeout = setTimeout(() => { peer.removeEventListener("icegatheringstatechange", check); reject(new Error("Connexion trop lente. Réessaie en texte.")); }, 8000);
          function check() { if (peer.iceGatheringState === "complete") { clearTimeout(timeout); peer.removeEventListener("icegatheringstatechange", check); resolve(); } }
          peer.addEventListener("icegatheringstatechange", check); check();
        });
      }
      if (this.stopped) return false;
      const result = await api<VoiceSession>(`/tutor/${encodeURIComponent(this.mission)}/voice`, { sdp: peer.localDescription?.sdp, pushToTalk: this.pushToTalk });
      this.session = result;
      if (this.stopped) { await this.endServer(); return false; }
      await peer.setRemoteDescription({ type: "answer", sdp: result.sdp });
      this.timer = setTimeout(() => { this.callbacks.error("Temps de voix atteint. Continue en texte."); void this.stop(); }, Math.max(0, Date.parse(result.expires) - Date.now()));
      return true;
    } catch (error) {
      await this.stop();
      throw error;
    }
  }
  private async startEleven() {
    const result = await api<VoiceSession>(`/tutor/${encodeURIComponent(this.mission)}/voice`, { pushToTalk: this.pushToTalk });
    this.session = result;
    if (this.stopped) { await this.endServer(); return; }
    if (!result.token) throw new Error("Le fournisseur n’a pas autorisé la session vocale.");
    const { Conversation } = await import("@elevenlabs/client");
    if (this.stopped) { await this.endServer(); return; }
    const conversation = await Conversation.startSession({
      conversationToken: result.token, connectionType: "webrtc",
      overrides: { agent: { firstMessage: result.greeting ?? "Hello! I'm Alex. Ready for an adventure?" } },
      onConnect: () => { if (!this.stopped) this.callbacks.state("listening"); },
      onDisconnect: () => { if (!this.stopped) void this.stop(); },
      onError: () => { if (!this.stopped) { this.callbacks.error("Connexion vocale indisponible. Réessaie ou continue en texte."); void this.stop(); } },
      onModeChange: ({ mode }) => { if (!this.stopped) this.callbacks.state(mode === "speaking" ? "speaking" : "listening"); },
      onMessage: message => { if (!this.stopped) this.callbacks.caption(`${message.source === "user" ? "Toi : " : ""}${message.message}`); },
      onInterruption: () => { if (!this.stopped) this.callbacks.state("listening"); }
    });
    this.eleven = conversation;
    if (this.stopped) { await conversation.endSession(); await this.endServer(); return; }
    if (this.pushToTalk) conversation.setMicMuted(true);
    this.timer = setTimeout(() => { this.callbacks.error("Temps de voix atteint. Continue en texte."); void this.stop(); }, Math.max(0, Date.parse(result.expires) - Date.now()));
  }
  mute(muted: boolean) { this.eleven?.setMicMuted(muted); this.microphone?.getAudioTracks().forEach(track => { track.enabled = !muted; }); }
  interrupt() { this.send({ type: "response.cancel" }); this.send({ type: "output_audio_buffer.clear" }); this.callbacks.state("listening"); }
  talk(pressed: boolean) {
    if (this.provider === "elevenlabs") { this.mute(!pressed); return; }
    if (pressed) { this.interrupt(); this.send({ type: "input_audio_buffer.clear" }); this.mute(false); }
    else { this.mute(true); this.send({ type: "input_audio_buffer.commit" }); this.send({ type: "response.create" }); }
  }
  private send(event: unknown) { if (this.channel?.readyState === "open") this.channel.send(JSON.stringify(event)); }
  private async receive(raw: string) {
    let event: VoiceEvent;
    try { event = JSON.parse(raw) as VoiceEvent; } catch { return; }
    if (this.stopped) return;
    switch (event.type) {
      case "input_audio_buffer.speech_started": this.callbacks.state("listening"); break;
      case "response.created": this.captions = ""; this.callbacks.state("thinking"); break;
      case "output_audio_buffer.started": this.callbacks.state("speaking"); break;
      case "output_audio_buffer.stopped": this.callbacks.state("listening"); break;
      case "response.output_audio_transcript.delta": this.captions = (this.captions + (event.delta ?? "")).slice(-4000); this.callbacks.caption(this.captions); break;
      case "conversation.item.input_audio_transcription.completed": this.callbacks.caption(`Toi : ${event.transcript ?? ""}`); break;
      case "response.function_call_arguments.done": await this.execute(event); break;
      case "error": this.callbacks.error("Alex n'a pas pu traiter cette demande. Réessaie ou utilise le texte."); break;
    }
  }
  private async execute(event: VoiceEvent) {
    if (!event.call_id || !event.name || !event.arguments || event.arguments.length > 2048) return;
    const id = event.call_id;
    let work = this.calls.get(id);
    if (!work) {
      if (this.calls.size >= 50) { this.callbacks.error("Limite d'actions atteinte."); await this.stop(); return; }
      work = (async () => {
        try {
          const args: unknown = JSON.parse(event.arguments!);
          const result = await tool(this.mission, event.name!, args);
          this.callbacks.events(result.events); await this.callbacks.changed();
          return result.value;
        } catch { return { error: "Action refusée par le serveur. Relis le défi courant ; aucun progrès n'est accordé sans réponse valide." }; }
      })();
      this.calls.set(id, work);
    }
    const output = await work;
    this.send({ type: "conversation.item.create", item: { type: "function_call_output", call_id: id, output: JSON.stringify(output) } });
    this.send({ type: "response.create" });
  }
  private async endServer() {
    const session = this.session; this.session = undefined;
    if (session) await api(`/tutor/voice/${encodeURIComponent(session.id)}`, undefined, "DELETE").catch(() => {
      // The server's persistent deadline still closes the call if offline.
    });
  }
  async stop() {
    if (this.stopped) return;
    this.stopped = true; clearTimeout(this.timer);
    this.microphone?.getTracks().forEach(track => track.stop());
    this.channel?.close(); this.peer?.close(); this.audio.pause(); this.audio.srcObject = null;
    await this.eleven?.endSession().catch(() => {}); this.eleven = undefined;
    this.callbacks.state("idle"); await this.endServer();
  }
}
