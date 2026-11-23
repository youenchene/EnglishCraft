import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import { App, AnswerBox, Progress } from "./App";
import { Summary } from "./coreApi";

const summary: Summary = {xp:10,missions:{},reviews:{},due:[],history:[],errors:{}};
const response = (data: unknown, status = 200) => new Response(JSON.stringify(data), {status,headers:{"Content-Type":"application/json"}});

beforeEach(() => {
  sessionStorage.clear();
  vi.restoreAllMocks();
  vi.spyOn(globalThis, "fetch").mockResolvedValue(response({},401));
});

test("asks for the pre-provisioned account credentials", async () => {
  render(<App />);
  expect(await screen.findByRole("heading", { name: /prêt pour l’aventure/i })).toBeInTheDocument();
  expect(screen.getByLabelText(/mot de passe/i)).toHaveAttribute("type", "password");
});

test("loads an exercise after authentication", async () => {
  vi.mocked(globalThis.fetch).mockImplementation(async (url) => {
    const path = String(url);
    if (path.endsWith("/auth/me")) return response({},401);
    if (path.endsWith("/auth/login")) return response({id:"child",role:"child"});
    if (path.endsWith("/progress")) return response(summary);
    if (path.endsWith("/plan")) return response({minutes:18,stages:[]});
    if (path.endsWith("/training")) return response({id:"train-go-past_simple",kind:"past_simple",prompt:"Donne le prétérit de go.",skill:"english.irregular.go",verb:"go"});
    return response([]);
  });
  render(<App />);

  fireEvent.change(await screen.findByLabelText(/identifiant/i), { target: { value: "alex" } });
  fireEvent.change(screen.getByLabelText(/mot de passe/i), { target: { value: "a-long-password" } });
  fireEvent.click(screen.getByRole("button", { name: /commencer/i }));
  fireEvent.click(await screen.findByRole("button", { name: "Quick Training" }));
  await waitFor(() => expect(screen.getByText("Donne le prétérit de go.")).toBeInTheDocument());
  expect(globalThis.fetch).toHaveBeenCalledWith(
    "/api/v1/training",
    expect.objectContaining({ credentials: "same-origin" }),
  );
  expect(sessionStorage.length).toBe(0);
});

test("shows correction and allows mission retry without claiming XP", async () => {
  vi.mocked(globalThis.fetch).mockResolvedValue(response({status:"incorrect",expected:"went",xp:0,completed:false,nextStep:0}));
  render(<AnswerBox exercise={{id:"go",kind:"past_simple",prompt:"Past of go",skill:"english.go",verb:"go"}}
    missionId="diamond" onResult={async () => {}} onNext={async () => {}} />);
  fireEvent.change(screen.getByLabelText("Ta réponse"),{target:{value:"goed"}});
  fireEvent.click(screen.getByRole("button",{name:"Valider ma réponse"}));
  expect(await screen.findByText("Presque !")).toBeInTheDocument();
  expect(screen.getByText("went")).toBeInTheDocument();
  expect(screen.queryByText(/XP enregistrés/)).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button",{name:"Réessayer"}));
  expect(screen.getByLabelText("Ta réponse")).toBeEnabled();
});

test("reuses the idempotency key after a lost network response", async () => {
  vi.mocked(globalThis.fetch).mockRejectedValueOnce(new Error("Network lost"))
    .mockResolvedValueOnce(response({status:"correct",expected:"went",xp:10,completed:false,nextStep:1}));
  render(<AnswerBox exercise={{id:"go",kind:"past_simple",prompt:"Past of go",skill:"english.go",verb:"go"}}
    onResult={async () => {}} onNext={async () => {}} />);
  fireEvent.change(screen.getByLabelText("Ta réponse"),{target:{value:"went"}});
  fireEvent.click(screen.getByRole("button",{name:"Valider ma réponse"}));
  await screen.findByRole("alert");
  fireEvent.click(screen.getByRole("button",{name:"Valider ma réponse"}));
  await screen.findByText("Bien joué !");
  const calls = vi.mocked(globalThis.fetch).mock.calls;
  expect(calls[0]?.[1]?.body).toEqual(calls[1]?.[1]?.body);
});

test("parent sees frequent errors and pending open answers", () => {
  render(<Progress parent summary={{...summary,errors:{"english.irregular.go":3},history:[{exerciseId:"open",skill:"english.go",answer:"I went home.",expected:"",status:"needs_evaluation",at:"2026-01-01T00:00:00Z"}]}} />);
  expect(screen.getByText("Erreurs fréquentes")).toBeInTheDocument();
  expect(screen.getByText("irregular.go : 3 erreurs")).toBeInTheDocument();
  expect(screen.getByText("Évaluation nécessaire")).toBeInTheDocument();
});
