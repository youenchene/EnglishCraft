import { expect, test } from "@playwright/test";

test("disabled cloud voice explains configuration without requesting the student's microphone", async ({ page }) => {
  await page.addInitScript(() => {
    Object.defineProperty(navigator.mediaDevices, "getUserMedia", { value: async () => {
      document.body.dataset.microphoneRequested = "yes";
      throw new Error("Unexpected microphone request");
    } });
  });
  await page.route("**/api/v1/**", async route => {
    const path = new URL(route.request().url()).pathname;
    let body: unknown = [];
    if (path.endsWith("/auth/me")) body = { id: "child", role: "child" };
    if (path.endsWith("/progress")) body = { xp: 0, missions: {}, reviews: {}, due: [], history: [], errors: {} };
    if (path.endsWith("/missions")) body = [{ id: "diamond", title: "Diamond", theme: "village", steps: 1, version: 1 }];
    if (path.endsWith("/start")) body = { missionId: "diamond", title: "Diamond", theme: "village", step: 0, total: 1, completed: false, dialogue: "Hello", exercise: null };
    if (path.endsWith("/plan")) body = { minutes: 18, stages: [] };
    if (path.endsWith("/voice")) {
      await route.fulfill({ status: 503, json: { error: { code: "voice_disabled", message: "Conversation micro non configurée sur le serveur." } } });
      return;
    }
    await route.fulfill({ json: body });
  });
  await page.goto("/");
  await page.getByRole("button", { name: "Explorer", exact: true }).click();
  await page.getByRole("button", { name: "Commencer la conversation" }).click();
  await expect(page.getByRole("alert")).toContainText("Conversation micro non configurée sur le serveur.");
  expect(await page.locator("body").getAttribute("data-microphone-requested")).toBeNull();
  await page.getByRole("button", { name: "Continuer en texte" }).click();
  await expect(page.getByLabel("Parler à Alex en texte")).toBeEnabled();
});
