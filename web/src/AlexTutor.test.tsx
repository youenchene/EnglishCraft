import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import { AlexTutor, AlexAvatar } from "./AlexTutor";

beforeEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });
test("adventure shows only Alex and conversation controls without silently activating the microphone", () => {
  const getUserMedia = vi.fn();
  Object.defineProperty(navigator, "mediaDevices", { configurable: true, value: { getUserMedia } });
  render(<AlexTutor mission="diamond" onChanged={async () => {}} />);
  expect(screen.getByRole("img", { name: "Alex : idle" })).toHaveAttribute("viewBox", "0 0 360 420");
  expect(screen.getByRole("img", { name: "Alex : idle" })).toHaveClass("block-alex");
  expect(screen.getByRole("button", { name: "Commencer la conversation" })).toBeEnabled();
  expect(screen.getByText(/Alex parlera en premier/)).toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "Un indice" })).not.toBeInTheDocument();
  expect(screen.queryByLabelText("Ta réponse")).not.toBeInTheDocument();
  expect(getUserMedia).not.toHaveBeenCalled();
});

test("the preview explains the story and rules without starting paid voice or speaking French", async () => {
  const fetch = vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(JSON.stringify({text:"Hi! I'm Alex. A diamond is missing. Speak English. I'll ask one question at a time. Say help for a hint."})));
  render(<AlexTutor mission="lost-diamond" onChanged={async () => {}} />);
  expect(await screen.findByText(/A diamond is missing/)).toBeInTheDocument();
  expect(fetch).toHaveBeenCalledOnce();
  expect(fetch).toHaveBeenCalledWith("/api/v1/tutor/lost-diamond/intro", expect.objectContaining({method:"GET"}));
});
test("text remains an explicit conversation fallback using the authenticated backend", async () => {
  vi.spyOn(globalThis, "fetch").mockImplementation(async () => new Response(JSON.stringify({ text:"Let's try.", hint:"Think about yesterday.", mood:"encouraging", mode:"local" }), { status:200 }));
  render(<AlexTutor mission="diamond" onChanged={async () => {}} />);
  fireEvent.click(screen.getByRole("button", { name:"Texte" }));
  fireEvent.change(screen.getByLabelText("Parler à Alex en texte"), { target:{ value:"Help me" } });
  fireEvent.click(screen.getByRole("button", { name:"Envoyer" }));
  expect(await screen.findByText(/Accompagnement local, sans IA/)).toBeInTheDocument();
  expect(screen.getByText("Help me")).toBeInTheDocument();
  expect(globalThis.fetch).toHaveBeenCalledWith("/api/v1/tutor/diamond/text", expect.objectContaining({ credentials:"same-origin", body:JSON.stringify({ message:"Help me" }) }));
});
test("microphone denial leaves a clear text conversation fallback", async () => {
  vi.spyOn(globalThis, "fetch").mockImplementation(async () => new Response("{}", { status:200 }));
  const getUserMedia = vi.fn().mockRejectedValue(new Error("Micro refusé"));
  Object.defineProperty(navigator, "mediaDevices", { configurable:true, value:{ getUserMedia } });
  vi.stubGlobal("RTCPeerConnection", vi.fn());
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => {});
  render(<AlexTutor mission="diamond" onChanged={async () => {}} />);
  fireEvent.click(screen.getByRole("button", { name:"Commencer la conversation" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("Micro refusé");
  fireEvent.click(screen.getByRole("button", { name:"Continuer en texte" }));
  expect(screen.getByLabelText("Parler à Alex en texte")).toBeEnabled();
});
test("Alex has distinct talking, listening and still animation states", () => {
  const view = render(<AlexAvatar state="speaking" animation />);
  expect(screen.getByRole("img", { name:"Alex : speaking" })).toHaveClass("speaking");
  expect(view.container.querySelector(".alex-talking-mouth")).toBeInTheDocument();
  view.rerender(<AlexAvatar state="listening" animation />);
  expect(screen.getByRole("img", { name:"Alex : listening" })).toHaveClass("listening");
  view.rerender(<AlexAvatar state="idle" animation={false} />);
  expect(screen.getByRole("img", { name:"Alex : idle" })).toHaveClass("still");
});
