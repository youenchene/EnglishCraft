import { api, Result, Verb } from "./coreApi";

export type OfflineAnswer = { exerciseId: string; answer: string; idempotencyKey: string };
export type Download = { student: string; verbs: Verb[]; pending: OfflineAnswer[] };
function database(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open("englishcraft-downloads", 1);
    request.onupgradeneeded = () => request.result.createObjectStore("downloads", { keyPath: "student" });
    request.onsuccess = () => resolve(request.result); request.onerror = () => reject(request.error);
  });
}
export async function downloads(): Promise<Download[]> {
  const db = await database();
  try { return await new Promise((resolve, reject) => {
    const request = db.transaction("downloads").objectStore("downloads").getAll();
    request.onsuccess = () => resolve(request.result as Download[]); request.onerror = () => reject(request.error);
  }); } finally { db.close(); }
}
async function update(student: string, mutate: (value: Download | undefined) => Download): Promise<void> {
  const db = await database();
  try { await new Promise<void>((resolve, reject) => {
    const tx = db.transaction("downloads", "readwrite"); const store = tx.objectStore("downloads");
    const read = store.get(student);
    read.onsuccess = () => { try { store.put(mutate(read.result as Download | undefined)); } catch (error) { tx.abort(); reject(error); } };
    tx.oncomplete = () => resolve(); tx.onerror = () => reject(tx.error); tx.onabort = () => reject(tx.error);
  }); } finally { db.close(); }
}
export async function downloadVerbs(student: string, verbs: Verb[]) {
  await update(student, value => ({ student, verbs, pending: value?.pending ?? [] }));
}
export async function enqueue(student: string, answer: OfflineAnswer) {
  await update(student, value => {
    if (!value || value.pending.length >= 100) throw new Error("Téléchargement absent ou limite de 100 réponses atteinte.");
    return { ...value, pending: [...value.pending, answer] };
  });
}
export async function synchronise(student: string): Promise<number> {
  const record = (await downloads()).find(value => value.student === student);
  if (!record) return 0;
  let synced = 0;
  for (const answer of record.pending) {
    // The live HttpOnly session supplies identity. Never synchronise another
    // profile's queue, nor accept offline correctness as authoritative.
    await api<Result>("/answers", { ...answer, missionId: "", step: 0 });
    await update(student, value => ({ ...value!, pending: value!.pending.filter(item => item.idempotencyKey !== answer.idempotencyKey) }));
    synced++;
  }
  return synced;
}
