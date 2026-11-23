import { AlexState } from "./tutorApi";

// Keep one voice for Alex, including when the browser loads more voices later.
export class BrowserSpeech {
  private voiceURI?: string;
  private generation = 0;
  cancel() { this.generation++; globalThis.speechSynthesis?.cancel(); }
  async speak(text: string, state: (value: AlexState) => void, error: (message: string) => void) {
    this.cancel();
    const generation = this.generation;
    const synthesis = globalThis.speechSynthesis;
    if (!synthesis || !globalThis.SpeechSynthesisUtterance) {
      error("Lecture vocale indisponible dans ce navigateur. Essaie Safari, Chrome ou Edge."); return;
    }
    if (!synthesis.getVoices().length && synthesis.addEventListener) {
      await new Promise<void>(resolve => {
        const finish = () => { clearTimeout(timer); synthesis.removeEventListener("voiceschanged", finish); resolve(); };
        const timer = setTimeout(finish, 1200);
        synthesis.addEventListener("voiceschanged", finish);
        if (synthesis.getVoices().length) finish();
      });
    }
    if (generation !== this.generation) return;
    const english = synthesis.getVoices().filter(voice => /^en(?:[-_]|$)/i.test(voice.lang));
    const quality = (voice: SpeechSynthesisVoice) =>
      (/natural|neural|enhanced|premium|siri/i.test(voice.name) ? 100 : 0) +
      (!voice.localService ? 50 : 0) + (/^en[-_]GB$/i.test(voice.lang) ? 10 : 0);
    const selected = english.find(voice => voice.voiceURI === this.voiceURI) ?? english.sort((a, b) => quality(b) - quality(a))[0];
    if (selected) this.voiceURI = selected.voiceURI;
    const utterance = new SpeechSynthesisUtterance(text.replace(" (Accompagnement local, sans IA)", ""));
    utterance.lang = selected?.lang ?? "en-GB";
    utterance.rate = 1;
    if (selected) utterance.voice = selected;
    utterance.onstart = () => { if (generation === this.generation) state("speaking"); };
    utterance.onend = () => { if (generation === this.generation) state("idle"); };
    utterance.onerror = event => {
      if (generation !== this.generation) return;
      state("idle");
      if (event.error !== "canceled" && event.error !== "interrupted") error("Lecture impossible. Vérifie le son de l’appareil et la disponibilité d’une voix anglaise.");
    };
    synthesis.speak(utterance);
  }
}
