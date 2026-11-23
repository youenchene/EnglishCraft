import { afterEach, expect, test, vi } from "vitest";
import { BrowserSpeech } from "./speech";

afterEach(() => vi.unstubAllGlobals());
test("prefers a natural English voice over the first legacy British voice and keeps it between replies", async () => {
  const legacy = { name: "Daniel", lang: "en-GB", voiceURI: "legacy", localService: true };
  const natural = { name: "Google UK English Male", lang: "en-GB", voiceURI: "natural", localService: false };
  let voices = [legacy, natural];
  const speak = vi.fn();
  vi.stubGlobal("speechSynthesis", { cancel: vi.fn(), getVoices: () => voices, speak });
  vi.stubGlobal("SpeechSynthesisUtterance", class { constructor(public text: string) {} });
  const speech = new BrowserSpeech();
  await speech.speak("Hello", () => {}, () => {});
  voices = [legacy, { name: "Enhanced English", lang: "en-GB", voiceURI: "another", localService: true }, natural];
  await speech.speak("Try again", () => {}, () => {});
  expect(speak.mock.calls.map(([utterance]) => utterance.voice)).toEqual([natural, natural]);
});
test("waits for voices to load instead of using a default legacy voice on the first reply", async () => {
  const natural = { name: "Siri English", lang: "en-GB", voiceURI: "siri", localService: true };
  let voices: object[] = [];
  const synthesis = Object.assign(new EventTarget(), { cancel: vi.fn(), getVoices: () => voices, speak: vi.fn() });
  vi.stubGlobal("speechSynthesis", synthesis);
  vi.stubGlobal("SpeechSynthesisUtterance", class { constructor(public text: string) {} });
  const speech = new BrowserSpeech();
  const pending = speech.speak("Hello", () => {}, () => {});
  expect(synthesis.speak).not.toHaveBeenCalled();
  voices = [natural]; synthesis.dispatchEvent(new Event("voiceschanged"));
  await pending;
  expect(synthesis.speak).toHaveBeenCalledWith(expect.objectContaining({ voice: natural }));
});
