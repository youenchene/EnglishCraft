export type Exercise = {
  verbId: string;
  baseForm: string;
  translation: string;
  skillId: string;
};

export type AnswerInput = { pastSimple: string; pastParticiple: string };

export type AnswerResult = {
  correct: boolean;
  pastSimpleCorrect: boolean;
  pastParticipleCorrect: boolean;
  expectedPastSimple: string;
  expectedPastParticiple: string;
  feedback: string;
  validationLevel: "not_started" | "recognition" | "recall" | "context";
  correctAnswers: number;
  totalAnswers: number;
  nextReviewAt: string;
};

type APIError = { error?: { message?: string } };

async function request<T>(path: string, token: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as APIError;
    throw new Error(payload.error?.message ?? "Impossible de contacter EnglishCraft.");
  }
  return (await response.json()) as T;
}

export function getNextExercise(token: string): Promise<Exercise> {
  return request<Exercise>("/api/v1/practice/irregular-verbs/next", token);
}

export function submitAnswer(token: string, verbId: string, input: AnswerInput): Promise<AnswerResult> {
  return request<AnswerResult>(`/api/v1/practice/irregular-verbs/${encodeURIComponent(verbId)}/answers`, token, {
    method: "POST",
    body: JSON.stringify(input),
  });
}
