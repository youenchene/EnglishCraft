// Explicit paid integration check. Requires an isolated, provisioned E2E family.
import { chromium } from "@playwright/test";

if (process.env.E2E_VOICE_LIVE !== "yes" || !process.env.E2E_VOICE_AUDIO_FILE || !process.env.E2E_CHILD_PASSWORD || !process.env.E2E_PARENT_PASSWORD) {
  console.error("Requires explicit E2E_VOICE_LIVE=yes, a synthetic audio fixture and isolated E2E family credentials."); process.exit(1);
}
const baseURL = process.env.E2E_BASE_URL;
if (baseURL !== "http://127.0.0.1:18081") { console.error("This live test is restricted to the isolated fixture port 18081."); process.exit(1); }
const browser = await chromium.launch({ args: ["--use-fake-ui-for-media-stream", "--use-fake-device-for-media-stream", `--use-file-for-fake-audio-capture=${process.env.E2E_VOICE_AUDIO_FILE}`, "--autoplay-policy=no-user-gesture-required"] });
const parent = await browser.newContext();
const child = await browser.newContext({ permissions: ["microphone"] });
try {
  const login = async (context, username, password) => {
    const response = await context.request.post(`${baseURL}/api/v1/auth/login`, { data: { username, password } });
    if (!response.ok()) throw new Error(`Test login HTTP ${response.status()}`);
  };
  await login(parent, process.env.E2E_PARENT_USERNAME, process.env.E2E_PARENT_PASSWORD);
  await login(child, process.env.E2E_CHILD_USERNAME, process.env.E2E_CHILD_PASSWORD);
  const enable = await parent.request.put(`${baseURL}/api/v1/parent/ai`, { data: { enabled: true } });
  if (!enable.ok()) throw new Error(`Consent HTTP ${enable.status()}`);
  const before = await (await child.request.get(`${baseURL}/api/v1/progress`)).json();
  const page = await child.newPage();
  await page.addInitScript(() => {
    new MutationObserver(() => {
      if ([...document.querySelectorAll(".tutor-status")].some(node => node.textContent.includes("Alex parle"))) document.documentElement.dataset.voiceSpeaking = "yes";
    }).observe(document, { subtree: true, childList: true, characterData: true });
  });
  await page.goto(baseURL);
  await page.locator("section.mission.stadium").getByRole("button").click();
  await page.getByRole("button", { name: "Commencer la conversation" }).click();
  await page.getByRole("button", { name: "Terminer la conversation" }).waitFor({ timeout: 25_000 });
  await page.waitForFunction(() => document.documentElement.dataset.voiceSpeaking === "yes", null, { timeout: 35_000 });
  if (process.env.E2E_VOICE_SILENT === "yes" && await page.locator(".chat-message.from-student").count() !== 0) throw new Error("Silent fixture produced a student utterance");
  await page.getByRole("button", { name: "Terminer la conversation" }).click();
  await page.getByRole("button", { name: "Commencer la conversation" }).waitFor();
  const after = await (await child.request.get(`${baseURL}/api/v1/progress`)).json();
  if (before.xp !== after.xp || JSON.stringify(before.history) !== JSON.stringify(after.history)) throw new Error("Voice changed learning evidence");
  console.log("PASS: authenticated pupil UI → Go admission → real ElevenLabs WebRTC → speaking event; XP/history unchanged.");
  if (process.env.E2E_VOICE_SILENT === "yes") console.log("PASS: Alex spoke first while the microphone input remained silent.");
  // Revoked consent must fail preflight, before a new microphone/paid session.
  await parent.request.put(`${baseURL}/api/v1/parent/ai`, { data: { enabled: false } });
  await page.getByRole("button", { name: "Commencer la conversation" }).click();
  await page.getByRole("alert").filter({ hasText: "autorisation IA du parent" }).waitFor();
  console.log("PASS: parental consent revocation blocks a subsequent voice session.");
} catch (error) {
  // Never serialize headers, provider URLs/tokens, transcripts or credentials.
  console.error("FAIL: live integrated voice test — " + error.message.split("\n")[0]);
  process.exitCode = 1;
} finally {
  await parent.request.put(`${baseURL}/api/v1/parent/ai`, { data: { enabled: false } }).catch(() => {});
  await browser.close();
}
