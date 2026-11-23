export type User = { id: string; role: "parent" | "child" };
export type ChildProfile = { id: string; username: string };
export type Mission = { id: string; title: string; theme: string; steps: number; version: number };
export type ExerciseView = { id: string; kind: string; prompt: string; skill: string; verb: string };
export type Step = { missionId: string; title: string; theme: string; step: number; total: number; completed: boolean; dialogue: string; exercise: ExerciseView | null };
export type Result = { status: "correct" | "incorrect" | "needs_evaluation"; expected: string; xp: number; completed: boolean; nextStep: number };
export type Verb = { base: string; past: string; participle: string; translation: string; difficulty: number; categories: string[]; examples: string[] };
export type Review = { stage: string; streak: number; attempts: number; failures: number; due: string; last: string };
export type Attempt = { exerciseId: string; skill: string; answer: string; expected: string; status: string; at: string };
export type Summary = { xp: number; missions: Record<string, { step: number; completed: boolean; xp: number }>; reviews: Record<string, Review>; due: string[]; history: Attempt[]; errors: Record<string, number> };

export async function api<T>(path: string, body?: unknown, method?: string, child?: string): Promise<T> {
  const response = await fetch(`/api/v1${path}${child ? `?child=${encodeURIComponent(child)}` : ""}`, {
    method: method ?? (body === undefined ? "GET" : "POST"),
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  });
  if (!response.ok) {
    const payload = await response.json().catch(() => ({})) as { error?: { message?: string } };
    throw new Error(payload.error?.message ?? "Impossible de contacter EnglishCraft.");
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
