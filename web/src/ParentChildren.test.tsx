import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ParentChildren } from "./ParentChildren";

test("parent creates a child and selects the new profile without exposing its password", async () => {
  const fetch = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({ id: "new", username: "alex" }), { status: 201 }));
  const onSelect = vi.fn().mockResolvedValue(undefined);
  render(<ParentChildren children={[]} selected="" onSelect={onSelect} />);
  fireEvent.change(screen.getByLabelText("Identifiant de l’enfant"), { target: { value: "alex" } });
  fireEvent.change(screen.getByLabelText("Mot de passe de l’enfant"), { target: { value: "a-long-password" } });
  fireEvent.click(screen.getByRole("button", { name: "Ajouter l’enfant" }));
  await waitFor(() => expect(onSelect).toHaveBeenCalledWith("new"));
  expect(fetch).toHaveBeenCalledWith("/api/v1/parent/children", expect.objectContaining({ body: JSON.stringify({ username: "alex", password: "a-long-password" }) }));
  expect(screen.getByLabelText("Mot de passe de l’enfant")).toHaveValue("");
  fetch.mockRestore();
});
