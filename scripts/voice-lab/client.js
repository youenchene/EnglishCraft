import { Conversation } from "@elevenlabs/client";

const start = document.querySelector("#start");
const stop = document.querySelector("#stop");
const status = document.querySelector("#status");
const captions = document.querySelector("#captions");
let conversation;
let timer;
function disconnected() {
  clearTimeout(timer); stop.disabled = true;
  status.textContent = "Session terminée. Un seul essai est autorisé par lancement du serveur.";
}
start.addEventListener("click", async () => {
  start.disabled = true; status.textContent = "Connexion…";
  try {
    const response = await fetch("/session", { method: "POST" });
    const data = await response.json();
    if (!response.ok) throw new Error(data.message ?? "Session indisponible.");
    conversation = await Conversation.startSession({
      conversationToken: data.token,
      connectionType: "webrtc",
      onConnect: () => { document.body.dataset.connected = "yes"; status.textContent = "Micro actif — parle à Alex en anglais (60 secondes maximum)."; stop.disabled = false; },
      onDisconnect: disconnected,
      onError: () => { status.textContent = "Erreur vocale. Vérifie la permission micro et la connexion."; },
      onMessage: message => {
        if (message.source !== "user") document.body.dataset.agentMessages = String(Number(document.body.dataset.agentMessages ?? 0) + 1);
        captions.textContent = `${message.source === "user" ? "Toi" : "Alex"} : ${message.message}`;
      },
      onModeChange: ({ mode }) => {
        if (mode === "speaking") document.body.dataset.speakingEvents = String(Number(document.body.dataset.speakingEvents ?? 0) + 1);
        status.textContent = mode === "speaking" ? "Alex parle…" : "Alex écoute…";
      }
    });
    timer = setTimeout(() => void conversation.endSession(), data.maxSeconds * 1000);
  } catch (error) {
    status.textContent = error instanceof Error ? error.message : "Connexion impossible.";
  }
});
stop.addEventListener("click", async () => { await conversation?.endSession(); disconnected(); });
window.addEventListener("pagehide", () => { clearTimeout(timer); void conversation?.endSession(); });
