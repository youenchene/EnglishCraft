// Explicit opt-in paid test, synthetic microphone only. No pupil credentials.
import { createRequire } from "node:module";
const require = createRequire(new URL("../../web/package.json", import.meta.url));
const { chromium } = require("@playwright/test");
if (!process.env.VOICE_LAB_AUDIO_FILE || process.env.VOICE_LAB_LIVE_TEST !== "yes") {
  console.error("Set VOICE_LAB_LIVE_TEST=yes and VOICE_LAB_AUDIO_FILE to a synthetic WAV file. This consumes paid API usage."); process.exit(1);
}
const browser = await chromium.launch({ args: ["--use-fake-ui-for-media-stream", "--use-fake-device-for-media-stream", `--use-file-for-fake-audio-capture=${process.env.VOICE_LAB_AUDIO_FILE}`, "--autoplay-policy=no-user-gesture-required"] });
const context = await browser.newContext({ permissions: ["microphone"] });
const page = await context.newPage();
try {
  await page.goto("http://127.0.0.1:18082/");
  await page.locator("#start").click();
  await page.waitForFunction(() => document.body.dataset.connected === "yes", null, { timeout: 25_000 });
  await page.waitForFunction(() => Number(document.body.dataset.agentMessages ?? 0) > 0 && Number(document.body.dataset.speakingEvents ?? 0) > 0, null, { timeout: 30_000 });
  console.log("PASS: real ElevenLabs WebRTC session, synthetic microphone transcript, Mistral answer and ElevenLabs speaking event.");
  console.log("No physical-device sound quality assessment performed.");
} catch {
  console.error("FAIL: live speech round trip incomplete. UI status: " + await page.locator("#status").textContent());
  process.exitCode = 1;
} finally {
  if (await page.locator("#stop").isEnabled()) await page.locator("#stop").click();
  await browser.close();
}
