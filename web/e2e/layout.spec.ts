import { expect, test } from "@playwright/test";
import type { Page } from "@playwright/test";

// Explicit UI fixtures: these tests validate layout, not an external LLM.
async function fixture(page: Page, parent = false) {
  await page.route("**/api/v1/**", async route => {
    const path = new URL(route.request().url()).pathname.replace("/api/v1", "");
    const summary = {xp:35,missions:{"lost-diamond":{step:2,completed:false,xp:35}},reviews:{"english.irregular.go":{stage:"learning",attempts:3,failures:1,due:"2026-10-13T12:00:00Z"}},due:["english.irregular.go"],history:[],errors:{"english.irregular.go":1}};
    const data = path === "/auth/me" ? {id:parent ? "parent" : "child",role:parent ? "parent" : "child"}
      : path === "/parent/children" ? [{id:"child",username:"Alex"}]
      : path === "/progress" || path === "/parent/results" ? summary
      : path === "/missions" ? [
        {id:"lost-diamond",title:"The Lost Diamond",theme:"adventure",steps:5,version:1},
        {id:"village",title:"The Mysterious Village",theme:"village",steps:5,version:1},
        {id:"football",title:"The Post-Match Interview",theme:"stadium",steps:5,version:1},
      ]
      : path.endsWith("/plan") ? {minutes:18,stages:[
        {name:"Révisions",minutes:4,skills:["english.irregular.go"],reason:"Échéance de révision atteinte."},
        {name:"Découverte ou consolidation",minutes:4,skills:[],reason:"Une notion à pratiquer."},
        {name:"Aventure ludique",minutes:5,skills:[],reason:"Réutiliser les notions dans une mission."},
        {name:"Évaluation courte",minutes:3,skills:[],reason:"Une correction validée par le serveur."},
        {name:"Bilan",minutes:2,skills:[],reason:"Préparer la prochaine séance."},
      ]}
      : path === "/parent/ai" ? {enabled:true,spending:[]}
       : path.endsWith("/intro") ? {text:"Hi! I'm Alex, your block-world guide. A diamond is missing from our cave. Together, we'll follow clues to find it.\n\nSpeak English; short answers are perfect. I'll ask one question at a time. Say 'help' for a hint, or 'repeat' to hear a question again. You can interrupt me.\n\nWhere should we look first?"}
       : path.includes("/start") || path.includes("/current") ? {missionId:"lost-diamond",title:"The Lost Diamond",theme:"adventure",step:2,total:5,completed:false,dialogue:"Can you help me find the diamond?",exercise:{id:"diamond-go",kind:"past_simple",prompt:"Donne le prétérit de go.",skill:"english.irregular.go",verb:"go"}}
      : [];
    await route.fulfill({json:data});
  });
}
async function noOverflow(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
}

for (const width of [390, 768, 1440]) {
  test(`child dashboard and adventure stay readable at ${width}px`, async ({page}) => {
    await page.setViewportSize({width,height:1000}); await fixture(page); await page.goto("/");
    await expect(page.getByRole("heading",{name:"Une nouvelle quête t’attend."})).toBeVisible();
    await noOverflow(page);
    if (width === 1440) await page.screenshot({path:test.info().outputPath("child-home.png"),fullPage:true});
    await page.locator("section.mission").filter({hasText:"The Lost Diamond"}).getByRole("button").click();
    await expect(page.getByRole("img",{name:"Alex : idle"})).toBeVisible();
    await expect(page.getByLabel("Ta réponse",{exact:true})).toHaveCount(0);
    await expect(page.locator(".challenge-panel, .mission-context, .adventure-progress")).toHaveCount(0);
    await expect(page.getByRole("navigation",{name:"Navigation principale"})).not.toBeVisible();
    await expect(page.getByRole("button",{name:"Commencer la conversation"})).toBeVisible();
    await expect(page.getByRole("log")).toContainText("Speak English");
    const bubble = await page.locator(".chat-message").first().boundingBox();
    expect(bubble!.width).toBeGreaterThan(200); expect(bubble!.height).toBeLessThan(600);
    await page.getByRole("button",{name:"Texte",exact:true}).click();
    await expect(page.getByLabel("Parler à Alex en texte")).toBeVisible();
    await expect(page.getByRole("button",{name:"Texte",exact:true})).toHaveCSS("background-color", "rgb(36, 94, 77)");
    await noOverflow(page);
    await page.screenshot({path:test.info().outputPath("adventure.png"),fullPage:true});
  });
  test(`parent controls and courses stay readable at ${width}px`, async ({page}) => {
    await page.setViewportSize({width,height:1000}); await fixture(page,true); await page.goto("/");
    await expect(page.getByRole("heading",{name:"Chaque effort compte."})).toBeVisible(); await noOverflow(page);
    await page.getByRole("button",{name:"Réglages IA",exact:true}).click();
    const consent=page.getByLabel("Autoriser les fournisseurs IA (désactivés par défaut)");
    await expect(consent).toBeVisible();
    const box=await consent.boundingBox(); expect(box!.width).toBeLessThan(30); expect(box!.height).toBeLessThan(30);
    await noOverflow(page);
    await page.screenshot({path:test.info().outputPath("parent-settings.png"),fullPage:true});
    await page.getByRole("button",{name:"Cours & séance",exact:true}).click();
    await expect(page.getByLabel("Texte du cours")).toBeVisible(); await noOverflow(page);
    await page.getByRole("button",{name:"Ma famille",exact:true}).click();
    await page.locator("details.add-profile > summary").click();
    await expect(page.getByRole("button",{name:"Ajouter l’enfant"})).toBeVisible(); await noOverflow(page);
  });
}

test("conversation portrait respects reduced motion", async ({page}) => {
  await page.emulateMedia({reducedMotion:"reduce"}); await fixture(page); await page.goto("/");
  await page.locator("section.mission").filter({hasText:"The Lost Diamond"}).getByRole("button").click();
  const eyes = page.locator(".alex-eyes");
  await expect(eyes).toBeVisible();
  expect(await eyes.evaluate(element => getComputedStyle(element).animationName)).toBe("none");
});
