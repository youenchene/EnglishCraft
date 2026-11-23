import { FormEvent, useEffect, useState } from "react";
import { api as request } from "./coreApi";
import { Plan, SchoolImport } from "./tutorApi";
import { Icon } from "./Icon";

export function FamilyLearning({ parent, child, mode = "all", compact = false }: { parent: boolean; child?: string; mode?: "all" | "plan" | "courses" | "settings"; compact?: boolean }) {
  function api<T>(path: string, body?: unknown, method?: string) { return request<T>(path, body, method, child); }
  const [plan, setPlan] = useState<Plan>();
  const [imports, setImports] = useState<SchoolImport[]>([]);
  const [spending, setSpending] = useState<{provider: string; calls: number; reservedMicros: number; voiceSeconds: number}[]>([]);
  const [enabled,setEnabled] = useState(false);
  const [text, setText] = useState(""); const [error, setError] = useState(""); const [busy, setBusy] = useState(false);
  async function load() {
    setPlan(await api<Plan>(parent ? "/parent/plan" : "/plan"));
    setImports(await api<SchoolImport[]>(parent ? "/parent/imports" : "/school-lessons"));
    if (parent) { const value=await api<{enabled:boolean;spending:typeof spending}>("/parent/ai"); setSpending(value.spending); setEnabled(value.enabled); }
  }
  useEffect(() => { void load().catch(() => setError("Plan pédagogique indisponible.")); }, [parent]); // eslint-disable-line react-hooks/exhaustive-deps
  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setError("");
    try { await api("/parent/imports", { text }); setText(""); await load(); }
    catch (reason) { setError(reason instanceof Error ? reason.message : "Import impossible."); }
    finally { setBusy(false); }
  }
  async function toggleAI(value: boolean) {
    const previous=enabled;
    setEnabled(value); setBusy(true); setError("");
    try { await api("/parent/ai",{enabled:value},"PUT"); await load(); }
    catch { setEnabled(previous); setError("Contrôle IA indisponible."); }
    finally { setBusy(false); }
  }
  const showPlan = mode !== "settings";
  const showCourses = mode === "all" || mode === "courses";
  const showSettings = parent && (mode === "all" || mode === "settings");
  return <div className={compact ? "family-learning compact-plan" : "family-learning"}>
    {showPlan && <section className="panel session-plan"><div className="card-heading"><div><span className="eyebrow">Un rythme qui fait du bien</span><h2>{parent ? "Sa séance du jour" : "Ta séance du jour"}</h2></div><span className="duration-pill"><Icon name="clock" />{plan?.minutes ?? 18} min</span></div>
    <ol className="plan-stages">{plan?.stages.map((stage,index) => <li key={stage.name}><span className="stage-number">{index + 1}</span><div><strong>{stage.name}</strong>{!compact && <p>{stage.reason} {stage.skills.join(", ")}</p>}</div><span className="stage-duration">{stage.minutes} min</span></li>)}</ol>
    {!plan && <p className="subtle">Chargement de la séance…</p>}
    {compact && <small>Un plan adapté aux révisions et aux réponses observées.</small>}</section>}
    {showSettings && <section className="panel ai-settings"><div className="card-heading"><div><span className="eyebrow">Des limites, pas des surprises</span><h2>Utilisation IA aujourd’hui (UTC)</h2></div><span className="icon-tile"><Icon name="shield" /></span></div>
      <div className="consent-row"><div><strong>Alex et les fournisseurs IA</strong><p>Tu peux autoriser ou désactiver les appels aux modèles pour cet enfant.</p></div><label className="toggle-label"><input type="checkbox" checked={enabled} disabled={busy} onChange={event => {
        void toggleAI(event.target.checked);
      }} /><span>Autoriser les fournisseurs IA (désactivés par défaut)</span></label></div>
      <div className={`permission-status ${enabled ? "enabled" : ""}`}><span className="status-dot" />{enabled ? "Autorisation activée" : "Fournisseurs IA désactivés"}</div>
      {spending.length === 0 && <div className="empty-state"><Icon name="chart" /><strong>Aucun appel.</strong><p>Les réservations apparaîtront ici après la première utilisation.</p></div>}
      {spending.map(value => <div className="spending-card" key={value.provider}><strong>{value.provider}</strong><div><span>{value.calls} réservations</span><span>{(value.reservedMicros / 1_000_000).toFixed(2)} USD réservés</span><span>{value.voiceSeconds} s de voix réservées</span></div></div>)}
      <p className="fine-print">L’autorisation ne configure pas de clé API. Les réservations affichées sont des plafonds de consommation, pas une facture fournisseur.</p></section>}
    {showCourses && <div className="course-layout">{parent && <section className="panel import-panel"><span className="icon-tile"><Icon name="book" /></span><span className="eyebrow">Le collège rejoint l’aventure</span><h2>Importer un cours (texte)</h2><p>Ajoute une leçon pour retrouver les verbes et préparer des fiches à valider.</p>
      <form onSubmit={event => void submit(event)}><label>Texte du cours<textarea placeholder="Colle ici le contenu de la leçon, sans informations personnelles…" required maxLength={20000} value={text} onChange={event => setText(event.target.value)} /></label><p className="fine-print"><Icon name="shield" /> Retire noms et coordonnées. L’extraction locale ne transmet pas le texte à un fournisseur.</p><button disabled={busy}>Analyser le texte<Icon name="arrow" /></button></form></section>}
    <section className="panel course-library"><div className="card-heading"><div><span className="eyebrow">Les fiches de la famille</span><h2>{parent ? "Contenus à valider" : "Mes cours validés"}</h2></div><span className="count-pill">{imports.length}</span></div>
    {imports.length === 0 && <div className="empty-state"><Icon name="book" /><strong>{parent ? "La bibliothèque attend son premier cours" : "Tes prochains cours arrivent bientôt"}</strong><p>{parent ? "Importe une leçon puis valide les fiches avant de les publier." : "Les cours validés par ta famille apparaîtront ici."}</p></div>}
    {imports.map(value => <details className="imported-lesson" key={value.id}><summary>{value.lesson.title} · {value.approved ? "Validé" : "Brouillon"}</summary>
      <p>{value.lesson.explanation}</p><p>{value.lesson.grammar.join(" · ")}</p><p>{value.lesson.vocabulary.join(" · ")}</p>
      <ul>{value.exercises.map(exercise => <li key={exercise.skill}>{exercise.prompt} : {exercise.answer}</li>)}</ul>
      <p>{value.lesson.missions.join(" · ")}</p>
      {parent && !value.approved && <button onClick={() => { void api(`/parent/imports/${value.id}/approve`, {}).then(load).catch(() => setError("Validation impossible.")); }}>Valider et publier les fiches</button>}
    </details>)}</section></div>}
    {error && <p role="alert">{error}</p>}
    {showSettings && <details className="panel danger-zone"><summary>Suppression du profil enfant</summary><p>Supprime définitivement comptes, sessions, progression, imports et usage du profil. Les sauvegardes doivent être purgées séparément. Efface aussi les téléchargements sur chaque appareil.</p>
      <button className="danger-button" onClick={() => { if (window.confirm("Supprimer définitivement le profil enfant ?")) {
        void api("/parent/child",{confirmation:"DELETE STUDENT"},"DELETE").then(() => window.location.reload()).catch(() => setError("Suppression impossible."));
      } }}>Supprimer le profil enfant</button></details>}
  </div>;
}
