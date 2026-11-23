import { FormEvent, useCallback, useEffect, useState } from "react";
import { api, ChildProfile, ExerciseView, Mission, Result, Step, Summary, User, Verb } from "./coreApi";
import { ParentChildren } from "./ParentChildren";
import { AlexTutor } from "./AlexTutor";
import { FamilyLearning } from "./FamilyLearning";
import { OfflineLibrary } from "./OfflineLibrary";
import { downloadVerbs, synchronise } from "./offline";
import { Brand, Icon, IconName } from "./Icon";

export function App() {
  const [user, setUser] = useState<User | null>(null);
  const [checking, setChecking] = useState(true);
  const [summary, setSummary] = useState<Summary | null>(null);
  const [missions, setMissions] = useState<Mission[]>([]);
  const [verbs, setVerbs] = useState<Verb[]>([]);
  const [tab, setTab] = useState("home");
  const [step, setStep] = useState<Step | null>(null);
  const [training, setTraining] = useState<ExerciseView | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [offline, setOffline] = useState(!navigator.onLine);
  const [children, setChildren] = useState<ChildProfile[]>([]);
  const [child, setChild] = useState("");

  const refresh = useCallback(async (activeUser: User, selected = "") => {
    let profile = selected;
    if (activeUser.role === "parent") {
      const profiles = await api<ChildProfile[]>("/parent/children");
      setChildren(profiles);
      profile = profiles.some(value => value.id === selected) ? selected : profiles[0]?.id ?? "";
      setChild(profile);
      if (!profile) { setSummary(null); return; }
    }
    const [progress, available, library] = await Promise.all([
      api<Summary>(activeUser.role === "parent" ? "/parent/results" : "/progress", undefined, undefined, profile),
      api<Mission[]>("/missions"), api<Verb[]>("/verbs"),
    ]);
    setSummary(progress); setMissions(available); setVerbs(library);
  }, []);

  useEffect(() => {
    if (offline) return;
    setChecking(true);
    let active = true;
    api<User>("/auth/me").then(async (current) => {
      if (!active) return;
      setUser(current);
      try { await refresh(current); } catch (reason) { if (active) setError(message(reason)); }
    }).catch(() => {}).finally(() => { if (active) setChecking(false); });
    return () => { active = false; };
  }, [refresh, offline]);

  async function perform(work: () => Promise<void>) {
    setBusy(true); setError("");
    try { await work(); } catch (reason) { setError(message(reason)); }
    finally { setBusy(false); }
  }
  async function connect(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await perform(async () => {
      const current = await api<User>("/auth/login", { username: form.get("username"), password: form.get("password") });
      setUser(current); await refresh(current);
    });
  }
  async function start(id: string) {
    await perform(async () => { setStep(await api<Step>(`/missions/${id}/start`, {})); setTab("adventure"); });
  }
  async function loadTraining() {
    await perform(async () => { setTraining(await api<ExerciseView | null>("/training")); setTab("training"); });
  }
  if (offline) return <OfflineLibrary close={() => setOffline(false)} />;
  if (checking) return <main className="login"><section className="panel loading-panel"><Brand /><div className="loading-dot" /><p>On prépare ton camp de base…</p><button className="secondary" onClick={() => setOffline(true)}>Révisions téléchargées</button></section></main>;
  if (!user) return <main className="login world-grid"><div className="login-layout"><aside className="login-story">
    <Brand /><span className="eyebrow light">L’anglais devient une aventure</span><h2>Small steps.<br />Big adventures.</h2>
    <p>Un diamant à retrouver. Un match à gagner. Et un peu plus de confiance à chaque phrase.</p>
    <Scene theme="adventure" alex /><div className="login-promise"><Icon name="shield" /><span>Un espace familial, pensé pour apprendre à son rythme.</span></div>
  </aside><section className="panel login-panel">
    <span className="eyebrow">Bienvenue dans ton camp de base</span><h1>Prêt pour l’aventure ?</h1>
    <p>Connecte-toi pour retrouver Alex et poursuivre tes progrès.</p>
    <form onSubmit={(event) => void connect(event)}>
      <label>Identifiant<input name="username" required autoComplete="username" maxLength={80} /></label>
      <label>Mot de passe<input name="password" type="password" required autoComplete="current-password" maxLength={72} /></label>
      <button disabled={busy}>Commencer <Icon name="arrow" /></button>
    </form>{error && <p role="alert">{error}</p>}
    <div className="login-note"><Icon name="users" /><small>Comptes créés par ta famille.<br />Aucune inscription publique.</small></div>
    <button className="text-button" onClick={() => setOffline(true)}><Icon name="offline" /> Réviser les contenus téléchargés</button>
  </section></div></main>;

  const parent = user.role === "parent";
  const navigation: { id: string; label: string; icon: IconName; accessible?: string }[] = parent ? [
    {id:"home",label:"Vue d’ensemble",icon:"home",accessible:"Parent Dashboard"},
    {id:"progress",label:"Progression",icon:"chart"},
    {id:"courses",label:"Cours & séance",icon:"book"},
    {id:"family",label:"Ma famille",icon:"users"},
    {id:"settings",label:"Réglages IA",icon:"settings"},
  ] : [
    {id:"home",label:"Mon camp de base",icon:"home",accessible:"Home"},
    {id:"adventure",label:"Mes aventures",icon:"compass",accessible:"Adventure"},
    {id:"training",label:"Révisions",icon:"book",accessible:"Quick Training"},
    {id:"progress",label:"Mes progrès",icon:"chart"},
    {id:"courses",label:"Mes cours",icon:"users"},
  ];
  const currentNav = navigation.find(item => item.id === tab);

  return <div className={`app-layout ${tab === "adventure" && step ? "conversation-focus" : ""}`}>
    <a className="skip-link" href="#main-content">Aller au contenu</a>
    <aside className="sidebar"><Brand /><span className="sidebar-label">{parent ? "L’espace parent" : "Ton espace d’apprentissage"}</span>
      <nav aria-label="Navigation principale">{navigation.map(item => <button key={item.id} className={tab === item.id ? "nav-link active" : "nav-link"}
        aria-label={item.accessible ?? item.label} aria-current={tab === item.id ? "page" : undefined} disabled={busy}
        onClick={() => item.id === "training" ? void loadTraining() : setTab(item.id)}><Icon name={item.icon} /><span>{item.label}</span>{tab === item.id && <span className="nav-dot" />}</button>)}</nav>
      <div className="sidebar-tip"><Icon name={parent ? "shield" : "sparkles"} /><strong>{parent ? "Chaque effort compte" : "Un petit pas, chaque jour"}</strong><p>{parent ? "Une progression réelle, validée par des réponses. Pas par des promesses." : "15 à 20 minutes pour explorer, pratiquer et gagner en confiance."}</p></div>
      <button className="logout-button" disabled={busy} onClick={() => void perform(async () => {
        await api("/auth/logout", {}, "POST"); setUser(null); setSummary(null); setStep(null); setTraining(null); setTab("home");
      })}><Icon name="logout" /> Quitter</button>
    </aside>
    <div className="workspace"><header className="topbar"><div className="breadcrumb">EnglishCraft <span>/</span> <strong>{currentNav?.label}</strong></div><div className="profile-chip"><span className="profile-avatar">{parent ? "P" : "E"}</span><div><strong>{parent ? "Espace parent" : "Explorateur"}</strong><small>{parent ? children.find(profile => profile.id === child)?.username ?? "Famille" : "Prêt pour une nouvelle quête"}</small></div></div></header>
    <main id="main-content" className="app-shell">
    {error && <p className="feedback" role="alert">{error}</p>}
    {parent && tab !== "family" && children.length > 0 && <div className="child-switcher"><span><Icon name="users" /> Suivi de l’élève</span><label className="sr-only" htmlFor="active-child">Profil enfant</label><select id="active-child" value={child} disabled={busy} onChange={event => { const id=event.target.value; void perform(async () => { setSummary(null); await refresh(user,id); }); }}>{children.map(profile => <option key={profile.id} value={profile.id}>{profile.username}</option>)}</select></div>}
    {parent && tab === "family" && <><PageHeading eyebrow="Un espace pour chacun" title="Ma famille" description="Gère les profils et retrouve leurs parcours d’apprentissage." /><ParentChildren children={children} selected={child} onSelect={async id => { setSummary(null); await refresh(user,id); }} /></>}
    {!summary && <button disabled={busy} onClick={() => void perform(() => refresh(user, child))}>Recharger la progression</button>}
    {summary && tab === "home" && <>
      <section className="hero panel"><div className="hero-copy"><span className="eyebrow"><span className="status-dot" /> {parent ? "Ensemble, on avance" : "Une aventure à ton rythme"}</span>
        <h1>{user.role === "parent" ? "Chaque effort compte." : "Une nouvelle quête t’attend."}</h1>
         <p>{parent ? "Une vue claire sur ses réussites, ses difficultés et ses prochaines découvertes." : "Explore de nouveaux mondes avec Alex. Ton anglais grandit, une petite victoire à la fois."}</p>
         {!parent && <button disabled={busy} onClick={() => { const next=missions.find(m=>!summary.missions[m.id]?.completed); if(next) void start(next.id); else void loadTraining(); }}>C’est parti <Icon name="arrow" /></button>}</div>
        <div className="hero-art"><span className="floating-label"><Icon name="sparkles" /> Chaque réponse te fait grandir</span><div className="xp"><Icon name="diamond" /><strong>{summary.xp}</strong><span>XP gagnés</span></div><span className="hero-pixel pixel-one" /><span className="hero-pixel pixel-two" /><span className="hero-pixel pixel-three" /></div></section>
      <div className="stats"><Stat icon="compass" value={Object.values(summary.missions).filter(m => m.completed).length} label="Missions terminées" detail="Des aventures accomplies" />
        <Stat icon="check" value={Object.values(summary.reviews).filter(r => r.stage === "mastery").length} label="Notions maîtrisées" detail="Des réussites dans la durée" />
        <Stat icon="clock" value={summary.due.length} label="Révisions du jour" detail={summary.due.length ? "Un petit rappel pour progresser" : "Tout est à jour, bien joué !"} /></div>
      <div className="section-heading"><div><span className="eyebrow">Apprendre en explorant</span><h2>{parent ? "Ses aventures" : "Choisis ton aventure"}</h2></div><span className="subtle">{missions.length} mondes à découvrir</span></div>
      <div className="mission-grid">{missions.map(m => <section className={`panel mission ${m.theme}`} key={m.id}>
        <Scene theme={m.theme} /><div className="mission-content"><span className="theme-tag">{m.theme === "stadium" ? "Football Academy" : m.theme === "village" ? "Village mystérieux" : "Block Adventure"}</span><h3>{m.title}</h3><p className="mission-description">{m.theme === "stadium" ? "Entre sur le terrain et raconte tes plus belles actions." : m.theme === "village" ? "Rencontre les habitants et résous le mystère du village." : "Pars sur les traces d’un diamant perdu avec Alex."}</p>
        <div className="mission-meta"><span>{m.steps} défis</span><span>Niveau découverte</span></div>
        <div className="mission-track" role="progressbar" aria-label={`Progression ${m.title}`} aria-valuenow={summary.missions[m.id]?.completed ? m.steps : summary.missions[m.id]?.step ?? 0} aria-valuemin={0} aria-valuemax={m.steps}><progress max={m.steps} value={summary.missions[m.id]?.completed ? m.steps : summary.missions[m.id]?.step ?? 0} /></div>
        <p className="mission-completion">{summary.missions[m.id]?.completed ? "Mission terminée ✓" : `${summary.missions[m.id]?.step ?? 0}/${m.steps} étapes validées`}</p>
        {!parent && <button className="secondary" disabled={busy} onClick={() => void start(m.id)}>{summary.missions[m.id] ? "Reprendre" : "Explorer"}<Icon name="arrow" /></button>}</div>
      </section>)}</div>
      <div className="home-bottom"><FamilyLearning key={child} parent={parent} child={child} mode="plan" compact /><section className="panel next-step-card"><span className="icon-tile"><Icon name={parent ? "chart" : "book"} /></span><span className="eyebrow">Le prochain petit pas</span><h2>{parent ? "Comprendre ses progrès" : "Quelques minutes suffisent"}</h2><p>{parent ? "Retrouve les notions travaillées, les réponses récentes et les difficultés à accompagner." : "Révise les verbes du jour. De courtes séances régulières font toute la différence."}</p><button className="secondary" onClick={() => parent ? setTab("progress") : void loadTraining()}>{parent ? "Voir la progression" : "Réviser maintenant"}<Icon name="arrow" /></button></section></div>
      {!parent && <section className="panel offline-card"><div><Icon name="download" /><h3>L’anglais, même sans Wi-Fi</h3><p>Télécharge tes fiches et synchronise tes réponses à ton retour.</p></div><div className="offline-actions">
        <button disabled={busy} onClick={() => void perform(() => downloadVerbs(user.id, verbs))}>Télécharger les verbes</button>
        <button onClick={() => setOffline(true)}>Quiz hors ligne</button>
        <button disabled={busy} onClick={() => void perform(async () => { await synchronise(user.id); await refresh(user); })}>Synchroniser mes réponses locales</button>
      </div></section>}
    </>}
    {tab === "progress" && summary && <><PageHeading eyebrow="Les petites victoires s’additionnent" title={parent ? "Sa progression" : "Mes progrès"} description="Des réponses réelles, des notions travaillées et des prochains pas à préparer." /><Progress summary={summary} parent={parent} /></>}
    {tab === "courses" && <><PageHeading eyebrow={parent ? "Du collège à l’aventure" : "Un peu de pratique chaque jour"} title={parent ? "Cours & séance" : "Mes cours"} description={parent ? "Prépare une séance adaptée et ajoute les cours suivis en classe." : "Retrouve ta séance du jour et les fiches validées par ta famille."} /><FamilyLearning key={child} parent={parent} child={child} mode="courses" /></>}
    {tab === "settings" && parent && <><PageHeading eyebrow="Tu gardes le contrôle" title="Réglages IA" description="Choisis si Alex peut utiliser les fournisseurs IA et consulte les réservations de consommation." /><FamilyLearning key={child} parent child={child} mode="settings" /></>}
    {tab === "adventure" && <section className="adventure">
      {!step ? <><PageHeading eyebrow="Choisis ton prochain monde" title="Mes aventures" description="Un monde à explorer, une conversation en anglais, et Alex à tes côtés." /><section className="panel empty-state"><Icon name="compass" /><strong>Avec Alex, l’aventure commence en parlant</strong><p>Choisis un univers, puis démarre la conversation. Alex prendra la parole en premier.</p><button onClick={() => setTab("home")}>Choisir une aventure<Icon name="arrow" /></button></section></> : <>
          <AlexTutor key={step.missionId} mission={step.missionId} title={step.title} onBack={() => setTab("home")} onChanged={async () => {
            await refresh(user); setStep(await api<Step>(`/missions/${step.missionId}/current`));
          }} />
      </>}
    </section>}
    {tab === "training" && <>
      <PageHeading eyebrow="Un petit entraînement, un grand progrès" title="Quick Training" description="Retrouve les verbes à réviser et consolide tes acquis à ton rythme." />
      <section className="panel training-challenge"><span className="eyebrow"><Icon name="sparkles" /> Le défi du moment</span>
        {training ? <AnswerBox key={training.id} exercise={training} onResult={() => refresh(user)} onNext={async () => setTraining(await api<ExerciseView | null>("/training"))} /> : <p>Tout est à jour ! Reviens à la prochaine révision.</p>}
      </section>
      <div className="section-heading"><div><span className="eyebrow">Ta bibliothèque de poche</span><h2>Cartes de révision</h2></div><span className="count-pill">{verbs.length} verbes</span></div><div className="verb-grid">{verbs.map(v => <details className="panel verb-card" key={v.base}>
        <summary>{v.base} <span>{v.translation}</span></summary><p><strong>{v.past} · {v.participle}</strong></p>
        {v.examples.map(example => <p key={example}>{example}</p>)}<small>Difficulté {v.difficulty}/3 · {v.categories.join(", ")}</small>
        <button className="secondary" onClick={() => setTraining({id:`train-${v.base}-build_sentence`,kind:"build_sentence",prompt:`Écris une phrase au passé avec ${v.base}. Évaluation humaine nécessaire.`,skill:`english.irregular.${v.base}`,verb:v.base})}>Construire une phrase</button>
      </details>)}</div>
    </>}
    <footer><span>EnglishCraft · les petits pas font les grandes aventures.</span><span>Illustrations originales · conçu pour apprendre</span></footer>
    </main></div>
  </div>;
}

function PageHeading({eyebrow,title,description}: {eyebrow:string;title:string;description:string}) {
  return <div className="page-heading"><span className="eyebrow">{eyebrow}</span><h1>{title}</h1><p>{description}</p></div>;
}
function Stat({icon,value,label,detail}: {icon:IconName;value:number;label:string;detail:string}) {
  return <section className="panel stat-card"><span className="icon-tile"><Icon name={icon} /></span><div><span className="stat-label">{label}</span><strong>{value}</strong><small>{detail}</small></div></section>;
}

export function AnswerBox({ exercise, missionId = "", step = 0, onResult, onNext }: {
  exercise: ExerciseView; missionId?: string; step?: number; onResult: () => Promise<void>; onNext: () => Promise<void>;
}) {
  const [answer, setAnswer] = useState("");
  const [result, setResult] = useState<Result | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [receipt, setReceipt] = useState(() => ({ key: newKey(), answer: "" }));
  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setError("");
    // Keep the same key and payload after a network error, including when the
    // server committed but its response was lost. Edited answers use a new key.
    const current = receipt.answer === answer ? receipt : { key: newKey(), answer };
    setReceipt(current);
    try {
      const outcome = await api<Result>("/answers", { missionId, step, exerciseId: exercise.id, answer, idempotencyKey: current.key });
      setResult(outcome); await onResult();
    } catch (reason) { setError(message(reason)); }
    finally { setBusy(false); }
  }
  async function next() {
    setBusy(true); setError("");
    try { await onNext(); } catch (reason) { setError(message(reason)); } finally { setBusy(false); }
  }
  return <div className="answer-box"><h2>{exercise.prompt}</h2>
    <form onSubmit={(event) => void submit(event)}><label>Ta réponse<input required maxLength={500} autoComplete="off" value={answer}
      disabled={busy || Boolean(result)} onChange={event => setAnswer(event.target.value)} /></label>
      <button disabled={busy || Boolean(result)}>Valider ma réponse</button></form>
    {error && <p role="alert">{error}</p>}
    {result && <section className="feedback" aria-live="polite">
      <h3>{result.status === "correct" ? "Bien joué !" : result.status === "needs_evaluation" ? "À évaluer avec ton parent" : "Presque !"}</h3>
      {result.expected && <p>Réponse : <strong>{result.expected}</strong></p>}
      {result.status === "needs_evaluation" && <p>Phrase enregistrée, sans validation automatique ni XP.</p>}
      {result.xp > 0 && <p>+ {result.xp} XP enregistrés</p>}
      {missionId && result.status === "incorrect" ? <button disabled={busy} onClick={() => { setResult(null); setAnswer(""); setReceipt({key:newKey(),answer:""}); }}>Réessayer</button> :
        <button disabled={busy} onClick={() => void next()}>{missionId ? "Continuer" : "Question suivante"}</button>}
    </section>}
  </div>;
}

export function Progress({ summary, parent }: { summary: Summary; parent: boolean }) {
  return <section className="panel progress"><h2>{parent ? "Compétences travaillées" : "Tes progrès"}</h2>
    {Object.entries(summary.reviews).length === 0 && <p>Les premières réponses feront apparaître ta progression.</p>}
    <ul>{Object.entries(summary.reviews).map(([skill, r]) => <li key={skill}><strong>{skill.replace("english.", "")}</strong><span>{r.stage === "mastery" ? "Maîtrise" : "Apprentissage"} · {r.attempts} tentatives · révision {new Date(r.due).toLocaleString("fr-FR")}</span></li>)}</ul>
    <h3>Révisions à effectuer</h3><p>{summary.due.length ? summary.due.map(skill => skill.replace("english.", "")).join(" · ") : "Aucune révision en retard."}</p>
    {parent && <><h3>Erreurs fréquentes</h3><ul>{Object.entries(summary.errors).sort((a,b) => b[1]-a[1]).map(([skill,count]) => <li key={skill}>{skill.replace("english.", "")} : {count} erreurs</li>)}</ul></>}
    <h3>Historique récent</h3><ul>{summary.history.map((a,i) => <li key={`${a.at}-${i}`}><strong>{a.answer}</strong><span>{a.status === "correct" ? "✓ Correct" : a.status === "needs_evaluation" ? "Évaluation nécessaire" : `À revoir : ${a.expected}`}</span></li>)}</ul>
  </section>;
}
function Scene({ theme, alex = false }: { theme: string; alex?: boolean }) {
  return <div className={`scene ${theme}`} role="img" aria-label={`Illustration originale : ${theme}${alex ? ", Alex l'explorateur" : ""}`}>
    <div className="sun" /><div className="hill hill-one" /><div className="hill hill-two" />
    {theme === "village" && <div className="house" />}{theme === "stadium" && <div className="goal" />}
    {alex && <div className="alex"><span>Alex</span><div className="head" /><div className="body" /></div>}
  </div>;
}
function message(reason: unknown) { return reason instanceof Error ? reason.message : "Une erreur est survenue."; }
function newKey() { return Array.from(crypto.getRandomValues(new Uint8Array(16)), byte => byte.toString(16).padStart(2, "0")).join(""); }
