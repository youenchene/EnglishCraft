import { FormEvent, useEffect, useState } from "react";
import { Download, downloads, enqueue } from "./offline";

export function OfflineLibrary({ close }: { close: () => void }) {
  const [library, setLibrary] = useState<Download[]>([]);
  const [student, setStudent] = useState(""); const [index, setIndex] = useState(0); const [answer, setAnswer] = useState(""); const [feedback, setFeedback] = useState("");
  useEffect(() => { void downloads().then(values => { setLibrary(values); setStudent(values[0]?.student ?? ""); }).catch(() => setFeedback("Stockage local indisponible.")); }, []);
  const record = library.find(value => value.student === student); const verb = record?.verbs[index % record.verbs.length];
  async function submit(event: FormEvent) {
    event.preventDefault(); if (!verb) return;
    try {
      await enqueue(student, { exerciseId: `train-${verb.base}-past_simple`, answer, idempotencyKey: crypto.randomUUID() });
      const correct = answer.trim().toLowerCase() === verb.past.toLowerCase();
      setFeedback(`${correct ? "Bien joué !" : `À revoir : ${verb.past}`} Résultat provisoire : à synchroniser après connexion. Aucun XP accordé hors ligne.`);
      setIndex(index + 1); setAnswer("");
    } catch (reason) { setFeedback(reason instanceof Error ? reason.message : "Stockage indisponible."); }
  }
  return <main className="panel app-shell"><h1>Révisions hors ligne</h1><button onClick={close}>Retour</button>
    <p>Fiches téléchargées uniquement · voix cloud indisponible · progression définitive après validation serveur.</p>
    {library.length === 0 ? <p>Connecte-toi en ligne et télécharge les verbes depuis Home.</p> : <>
      <label>Profil local<select value={student} onChange={event => { setStudent(event.target.value); setIndex(0); }}>{library.map(value => <option key={value.student}>{value.student}</option>)}</select></label>
      {verb && <form onSubmit={event => void submit(event)}><label>Prétérit de {verb.base}<input required maxLength={500} value={answer} onChange={event => setAnswer(event.target.value)} /></label><button>Répondre</button></form>}
      <div className="verb-grid">{record?.verbs.map(value => <details className="panel" key={value.base}><summary>{value.base} · {value.translation}</summary><p>{value.past} · {value.participle}</p>{value.examples.map(example => <p key={example}>{example}</p>)}</details>)}</div>
    </>}{feedback && <p role="status">{feedback}</p>}
  </main>;
}
