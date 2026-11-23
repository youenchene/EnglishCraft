import { expect, test } from "@playwright/test";

// Browser UI contracts use mocked HTTP and speech. Real persistence and ownership
// are covered by PostgreSQL integration tests; these never change the local family.
test("parent adds and switches child profiles", async ({ page }) => {
  const children = [{ id: "first", username: "alex" }];
  await page.route("**/api/v1/**", async route => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace("/api/v1", "");
    let body: unknown = [];
    let status = 200;
    if (path === "/auth/me") body = { id: "parent", role: "parent" };
    if (path === "/parent/children") {
      if (route.request().method() === "POST") {
        const input = route.request().postDataJSON();
        children.push({ id: "second", username: input.username });
        body = children[1]; status = 201;
      } else body = children;
    }
    if (path === "/parent/results") body = { xp: url.searchParams.get("child") === "second" ? 0 : 65, missions: {}, reviews: {}, due: [], history: [], errors: {} };
    if (path === "/parent/plan") body = { minutes: 18, stages: [] };
    if (path === "/parent/ai") body = { enabled: false, spending: [] };
    await route.fulfill({ status, json: body });
  });
  await page.goto("/");
  await expect(page.locator(".xp strong")).toHaveText("65");
  await page.getByRole("button", {name:"Ma famille",exact:true}).click();
  await page.locator("details.add-profile > summary").click();
  await page.getByLabel("Identifiant de l’enfant").fill("sam");
  await page.getByLabel("Mot de passe de l’enfant").fill("long-child-password");
  await page.getByRole("button", { name: "Ajouter l’enfant" }).click();
  await expect(page.getByLabel("Profil enfant")).toHaveValue("second");
  await page.getByRole("button",{name:"Parent Dashboard",exact:true}).click();
  await expect(page.locator(".xp strong")).toHaveText("0");
  await page.getByLabel("Profil enfant").selectOption("first");
  await expect(page.locator(".xp strong")).toHaveText("65");
});

test("adventure opens with Alex's English greeting and explicit voice/text conversation choices", async ({ page }) => {
  await page.route("**/api/v1/**", async route => {
    const path = new URL(route.request().url()).pathname;
    let body: unknown = [];
    if (path.endsWith("/auth/me")) body = { id: "child", role: "child" };
    if (path.endsWith("/progress")) body = { xp: 0, missions: {}, reviews: {}, due: [], history: [], errors: {} };
    if (path.endsWith("/missions")) body = [{ id: "diamond", title: "Diamond", theme: "village", steps: 1, version: 1 }];
    if (path.endsWith("/start")) body = { missionId: "diamond", title: "Diamond", theme: "village", step: 0, total: 1, completed: false, dialogue: "Hello", exercise: null };
    if (path.endsWith("/plan")) body = { minutes: 18, stages: [] };
    await route.fulfill({ json: body });
  });
  await page.goto("/");
  await page.getByRole("button", { name: "Explorer", exact: true }).click();
  await expect(page.getByText("Hello! I'm Alex. Ready for an adventure?")).toBeVisible();
  await expect(page.getByRole("button", { name:"Commencer la conversation" })).toBeVisible();
  await expect(page.getByText(/Alex parlera en premier/)).toBeVisible();
  await page.getByRole("button", { name:"Texte", exact:true }).click();
  await expect(page.getByLabel("Parler à Alex en texte")).toBeVisible();
});
