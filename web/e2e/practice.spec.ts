import { expect, test } from "@playwright/test";

test("child completes an irregular verb practice round", async ({ page }) => {
  const username = process.env.E2E_CHILD_USERNAME;
  const password = process.env.E2E_CHILD_PASSWORD;
  test.skip(!username || !password, "Provision an isolated E2E family and supply credentials");

  await page.goto("/");
  await page.getByLabel("Identifiant").fill(username!);
  await page.getByLabel("Mot de passe").fill(password!);
  await page.getByRole("button", { name: "Commencer" }).click();
  await page.getByRole("button", { name: "Quick Training" }).click();
  await page.getByLabel("Ta réponse").fill("test-answer");
  await page.getByRole("button", { name: "Valider ma réponse" }).click();

  await expect(page.getByRole("heading", { name: "Presque !" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Question suivante" })).toBeVisible();
});

test("conversation-only adventure, deterministic server correction, persisted XP, and parent results", async ({ page, browser }) => {
  const child = process.env.E2E_CHILD_USERNAME;
  const password = process.env.E2E_CHILD_PASSWORD;
  const parent = process.env.E2E_PARENT_USERNAME;
  const parentPassword = process.env.E2E_PARENT_PASSWORD;
  test.skip(!child || !password || !parent || !parentPassword, "Provision an isolated E2E family");
  await page.goto("/");
  await page.getByLabel("Identifiant").fill(child!);
  await page.getByLabel("Mot de passe").fill(password!);
  await page.getByRole("button", {name:"Commencer"}).click();
  const mission = page.locator("section.mission").filter({hasText:"The Lost Diamond"});
  await mission.getByRole("button").click();
  await expect(page.getByLabel("Ta réponse", {exact:true})).toHaveCount(0);
  await page.getByRole("button", {name:"Texte",exact:true}).click();
  await page.getByLabel("Parler à Alex en texte").fill("Can you help me?");
  await page.getByRole("button",{name:"Envoyer",exact:true}).click();
  await expect(page.getByText(/Accompagnement local, sans IA/)).toBeVisible();
  // Adventure no longer renders an exercise panel. Server learning contracts
  // remain tested explicitly, without reintroducing grading into the chat UI.
  const submit = async (answer: string) => {
    const step = await (await page.request.get("/api/v1/missions/lost-diamond/current")).json();
    const response = await page.request.post("/api/v1/answers", {data:{missionId:"lost-diamond",step:step.step,exerciseId:step.exercise.id,answer,idempotencyKey:crypto.randomUUID()}});
    expect(response.ok()).toBe(true); return response.json();
  };
  expect((await submit("goed")).status).toBe("incorrect");
  for (const answer of ["went","found","builded","taken","I found a diamond."]) {
    expect((await submit(answer)).status).toBe("correct");
  }
  await page.getByRole("button", {name:"Retour",exact:true}).click();
  await page.reload();
  await expect(page.locator(".xp strong")).toHaveText("65");
  await page.getByRole("button",{name:"Quitter"}).click();
  await page.getByLabel("Identifiant").fill(child!);
  await page.getByLabel("Mot de passe").fill(password!);
  await page.getByRole("button",{name:"Commencer"}).click();
  await expect(page.locator(".xp strong")).toHaveText("65");
  await page.getByRole("button",{name:"Télécharger les verbes"}).click();
  await expect(page.getByRole("button",{name:"Télécharger les verbes"})).toBeEnabled();
  await page.evaluate(() => navigator.serviceWorker.ready.then(() => undefined));
  await page.context().setOffline(true);
  await page.reload();
  await expect(page.getByRole("heading",{name:"Révisions hors ligne"})).toBeVisible();
  await page.getByLabel(/^Prétérit de/).fill("offline-response");
  await page.getByRole("button",{name:"Répondre",exact:true}).click();
  await expect(page.getByText(/Résultat provisoire/)).toBeVisible();
  await page.context().setOffline(false);
  await page.getByRole("button",{name:"Retour",exact:true}).click();
  await page.getByRole("button",{name:"Synchroniser mes réponses locales"}).click();
  await expect(page.getByRole("button",{name:"Synchroniser mes réponses locales"})).toBeEnabled();
  await expect(page.locator(".xp strong")).toHaveText("65");
  await page.getByRole("button",{name:"Synchroniser mes réponses locales"}).click();
  await expect(page.getByRole("button",{name:"Synchroniser mes réponses locales"})).toBeEnabled();
  await expect(page.locator(".xp strong")).toHaveText("65");
  await expect(page.locator("section.mission").filter({hasText:"The Lost Diamond"})).toContainText("Mission terminée ✓");
  // A reload exercises restoration via the HttpOnly cookie, not JS storage.
  await page.reload();
  await expect(page.locator(".xp strong")).toHaveText("65");
  const context = await browser.newContext({viewport:{width:390,height:844}});
  const dashboard = await context.newPage();
  await dashboard.goto(process.env.E2E_BASE_URL ?? "http://127.0.0.1:8080");
  await dashboard.getByLabel("Identifiant").fill(parent!);
  await dashboard.getByLabel("Mot de passe").fill(parentPassword!);
  await dashboard.getByRole("button",{name:"Commencer"}).click();
  await expect(dashboard.getByRole("heading",{name:"Chaque effort compte."})).toBeVisible();
  await expect(dashboard.locator(".xp strong")).toHaveText("65");
  await dashboard.getByRole("button",{name:"Progression",exact:true}).click();
  await expect(dashboard.getByText("Erreurs fréquentes")).toBeVisible();
  await dashboard.getByRole("button",{name:"Cours & séance",exact:true}).click();
  await dashboard.getByLabel("Texte du cours").fill("Past simple: I went home. I found a diamond.");
  await dashboard.getByRole("button",{name:"Analyser le texte"}).click();
  const imported=dashboard.locator("details").filter({hasText:"Cours du collège"});
  await imported.locator("summary").click();
  await imported.getByRole("button",{name:"Valider et publier les fiches"}).click();
  await expect(imported.locator("summary")).toContainText("Validé");
  await dashboard.getByRole("button",{name:"Réglages IA",exact:true}).click();
  await dashboard.getByLabel("Autoriser les fournisseurs IA (désactivés par défaut)").check();
  await expect(dashboard.getByLabel("Autoriser les fournisseurs IA (désactivés par défaut)")).toBeChecked();
  await dashboard.getByLabel("Autoriser les fournisseurs IA (désactivés par défaut)").uncheck();
  expect(await dashboard.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await context.close();
});
