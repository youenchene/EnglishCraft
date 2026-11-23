import { api } from "./coreApi";

export type GameEvent =
  | { version: 1; type: "CHARACTER_SPEAK" | "SHOW_HINT"; text: string }
  | { version: 1; type: "OPEN_DOOR" | "GIVE_DIAMOND" | "SCORE_GOAL" | "MISSION_COMPLETED" }
  | { version: 1; type: "AWARD_XP"; amount: number };
export type ToolResult = { value: unknown; events: GameEvent[] };
export type VoiceSession = { id: string; sdp: string; expires: string; provider?: "openai" | "elevenlabs"; token?: string; greeting?: string };
export type AlexState = "idle" | "listening" | "thinking" | "speaking" | "happy" | "encouraging";
export type Plan = { minutes: number; stages: { name: string; minutes: number; skills: string[]; reason: string }[] };
export type SchoolImport = { id: string; approved: boolean; source: string; lesson: {
  title: string; explanation: string; vocabulary: string[]; grammar: string[]; verbs: string[]; missions: string[];
}; exercises: { prompt: string; answer: string; skill: string }[] };

export async function tool(mission: string, name: string, args: unknown): Promise<ToolResult> {
  return api<ToolResult>(`/tutor/${encodeURIComponent(mission)}/tools`, { name, arguments: args });
}
