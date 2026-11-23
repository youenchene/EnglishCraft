import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { FamilyLearning } from "./FamilyLearning";

test("parent AI consent toggles immediately and reverts on rejected server request", async () => {
  vi.restoreAllMocks();
  const fetch=vi.spyOn(globalThis,"fetch").mockImplementation(async (url,input) => {
    if (input?.method==="PUT") return new Response(JSON.stringify({error:{message:"Not authorised"}}),{status:403});
    const path=String(url);
    return new Response(JSON.stringify(path.endsWith("/ai") ? {enabled:false,spending:[]} : path.endsWith("/plan") ? {minutes:18,stages:[]} : []),{status:200});
  });
  render(<FamilyLearning parent />);
  const consent=await screen.findByLabelText("Autoriser les fournisseurs IA (désactivés par défaut)");
  await waitFor(() => expect(fetch).toHaveBeenCalledWith("/api/v1/parent/ai",expect.anything()));
  fireEvent.click(consent);
  expect(consent).toBeChecked();
  expect(await screen.findByRole("alert")).toHaveTextContent("Contrôle IA indisponible.");
  expect(consent).not.toBeChecked();
});
