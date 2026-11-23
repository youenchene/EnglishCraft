import { FormEvent, useEffect, useRef, useState } from "react";
import { api } from "./coreApi";
import { AlexState, GameEvent } from "./tutorApi";
import { VoiceConnection } from "./voice";
import { Icon } from "./Icon";
import { AlexAvatar } from "./AlexAvatar";

type Message = { id: number; role: "alex" | "student"; text: string; local?: boolean };

export function AlexTutor({ mission, title = "Ton aventure en anglais", onBack, onChanged }: {
  mission: string; title?: string; onBack?: () => void; onChanged: () => Promise<void>;
}) {
  const [state, setState] = useState<AlexState>("idle");
  const [mode, setMode] = useState<"voice" | "text">("voice");
  const [text, setText] = useState("");
  const [messages, setMessages] = useState<Message[]>([{ id: 0, role: "alex", text: "Hello! I'm Alex. Ready for an adventure?" }]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [connected, setConnected] = useState(false);
  const [muted, setMuted] = useState(false);
  const voice = useRef<VoiceConnection>();
  const audio = useRef<HTMLAudioElement>(null);
  const log = useRef<HTMLDivElement>(null);
  const counter = useRef(0);
  const started = useRef(false);
  const changed = useRef(onChanged); changed.current = onChanged;
  useEffect(() => () => { void voice.current?.stop(); }, [mission]);
  useEffect(() => {
    let active = true;
    void api<{text:string}>(`/tutor/${encodeURIComponent(mission)}/intro`).then(result => {
      if (active && !started.current && typeof result.text === "string" && result.text.trim()) {
        setMessages([{ id:0, role:"alex", text:result.text }]);
      }
    }).catch(() => { /* The authenticated start/text endpoints still show actionable errors. */ });
    return () => { active = false; };
  }, [mission]);
  useEffect(() => { if (log.current) log.current.scrollTop = messages.length === 1 ? 0 : log.current.scrollHeight; }, [messages]);

  function say(role: Message["role"], value: string, replace = false, local = false) {
    if (!value.trim()) return;
    const id = ++counter.current;
    setMessages(current => {
      const last = current.at(-1);
      return replace && last?.role === role
        ? [...current.slice(0, -1), { ...last, text: value, local }]
        : [...current.slice(-11), { id, role, text: value, local }];
    });
  }
  function caption(value: string) {
    const student = value.startsWith("Toi : ");
    say(student ? "student" : "alex", student ? value.slice(6) : value, true);
  }
  function events(values: GameEvent[]) {
    for (const event of values) {
      if (event.version !== 1) continue;
      if (event.type === "SHOW_HINT" || event.type === "CHARACTER_SPEAK") say("alex", event.text);
      if (event.type === "MISSION_COMPLETED") { setState("happy"); say("alex", "Well done! You completed the mission."); }
    }
  }
  function stopVoice() {
    void voice.current?.stop(); setConnected(false); setMuted(false); setState("idle");
  }
  function switchMode(next: "voice" | "text") {
    if (next === mode) return;
    stopVoice(); setMode(next); setError("");
  }
  async function submit(event: FormEvent) {
    event.preventDefault();
    const message = text.trim(); if (!message || busy) return;
    started.current = true;
    say("student", message); setText(""); setBusy(true); setState("thinking"); setError("");
    try {
      const result = await api<{ text: string; hint: string; mood: AlexState; mode: string }>(`/tutor/${encodeURIComponent(mission)}/text`, { message });
      say("alex", `${result.text}${result.hint ? ` · ${result.hint}` : ""}`, false, result.mode === "local");
      setState(result.mood);
    } catch (reason) { setError(reason instanceof Error ? reason.message : "Alex est indisponible."); setState("idle"); }
    finally { setBusy(false); }
  }
  async function startVoice() {
    if (!audio.current || busy) return;
    started.current = true;
    setError(""); setBusy(true); setState("thinking");
    const connection = new VoiceConnection(mission, false, {
      state: value => { setState(value); if (value === "idle") setConnected(false); },
      caption, error: setError, events, changed: () => changed.current()
    }, audio.current);
    voice.current = connection;
    try { setConnected(await connection.start()); setMuted(false); }
    catch (reason) { setError(reason instanceof Error ? reason.message : "Microphone indisponible. Tu peux écrire à Alex."); setState("idle"); }
    finally { setBusy(false); }
  }
  const status = muted ? "Micro en pause" : state === "speaking" ? "Alex parle" : state === "listening" ? "Alex t’écoute" : state === "thinking" ? "Un petit instant…" : "Alex est prêt";

  return <section className={`conversation-stage ${state}`} aria-label="Conversation avec Alex">
    <header className="conversation-header">
      {onBack && <button className="conversation-back" onClick={() => { stopVoice(); onBack(); }}><Icon name="arrow" />Retour</button>}
      <div className="conversation-topic"><span>ENGLISHCRAFT · ADVENTURE</span><h1>{title}</h1></div>
      <div className="conversation-mode" aria-label="Mode de conversation">
        <button aria-pressed={mode === "voice"} disabled={busy} onClick={() => switchMode("voice")}><Icon name="mic" />Voix</button>
        <button aria-pressed={mode === "text"} disabled={busy} onClick={() => switchMode("text")}><Icon name="send" />Texte</button>
      </div>
    </header>

    <div className="conversation-body">
      <div className="alex-portrait-stage">
        <div className="portrait-orbit orbit-one" /><div className="portrait-orbit orbit-two" />
        <span className="portrait-spark spark-one" aria-hidden="true">✦</span><span className="portrait-spark spark-two" aria-hidden="true">✧</span>
        <div className="alex-portrait"><AlexAvatar state={muted ? "idle" : state} animation /></div>
        <div className="alex-nameplate"><span className="status-dot" /><strong>Alex</strong><span>Ton guide en blocs · parodie originale</span></div>
      </div>

      <div className="conversation-space">
        <div className="conversation-greeting"><span className="eyebrow">Une aventure. Une conversation.</span><h2>Let’s talk.</h2><p>Pas besoin d’être parfait. On apprend en parlant.</p></div>
        <div className="conversation-log" ref={log} role="log" aria-label="Conversation" aria-live="polite">
          {messages.map(message => <div className={`chat-message from-${message.role}`} key={message.id}><span>{message.role === "alex" ? "ALEX" : "TOI"}</span><p>{message.text}</p>{message.local && <small className="chat-local-label">Accompagnement local, sans IA</small>}</div>)}
        </div>
        <div className="tutor-status conversation-status" role="status"><span className="status-dot" />{status}
          <span className={`voice-wave ${state === "speaking" || state === "listening" ? "active" : ""}`} aria-hidden="true"><i /><i /><i /><i /><i /></span>
        </div>

        {mode === "voice" ? <div className="conversation-controls">
          {!connected ? <button className="start-conversation" disabled={busy} onClick={() => void startVoice()}><Icon name="mic" />{busy ? "Connexion avec Alex…" : "Commencer la conversation"}</button> : <>
            <button className="micro-toggle" aria-pressed={muted} onClick={() => { voice.current?.mute(!muted); setMuted(!muted); }}><Icon name="mic" />{muted ? "Réactiver le micro" : "Couper mon micro"}</button>
            <button className="stop-conversation" onClick={stopVoice}>Terminer la conversation</button>
          </>}
          <p>{connected ? "Pour interrompre Alex, commence à parler. Un casque peut limiter l’écho." : "Alex parlera en premier. Ton micro ne s’active qu’après ton accord."}</p>
        </div> : <form className="conversation-compose" onSubmit={event => void submit(event)}>
          <label className="sr-only" htmlFor="alex-message">Parler à Alex en texte</label>
          <textarea id="alex-message" rows={2} placeholder="Écris ta réponse à Alex…" value={text} maxLength={1000} required disabled={busy} onChange={event => setText(event.target.value)} />
          <button aria-label="Envoyer" disabled={busy || !text.trim()}><Icon name="send" /></button>
        </form>}
        {error && <p className="conversation-error" role="alert">{error}{mode === "voice" && <button onClick={() => switchMode("text")}>Continuer en texte</button>}</p>}
        <p className="conversation-privacy"><Icon name="shield" />Aucun audio enregistré par EnglishCraft. La voix utilise les fournisseurs autorisés par ton parent.</p>
      </div>
    </div>
    <audio ref={audio} controls hidden={!connected || voice.current?.provider === "elevenlabs"} aria-label="Conversation vocale avec Alex" />
  </section>;
}

export { AlexAvatar };
