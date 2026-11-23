import { FormEvent, useState } from "react";
import { api, ChildProfile } from "./coreApi";
import { Icon } from "./Icon";

export function ParentChildren({ children, selected, onSelect }: { children: ChildProfile[]; selected: string; onSelect: (id: string) => Promise<void> }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  async function select(id: string) {
    setBusy(true); setError("");
    try { await onSelect(id); } catch { setError("Chargement du profil impossible."); }
    finally { setBusy(false); }
  }
  async function add(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    setBusy(true); setError(""); setNotice("");
    try {
      const child = await api<ChildProfile>("/parent/children", { username: data.get("username"), password: data.get("password") });
      form.reset(); setNotice(`Compte créé pour ${child.username}.`);
      await onSelect(child.id);
    } catch (reason) { setError(reason instanceof Error ? reason.message : "Création impossible."); }
    finally { setBusy(false); }
  }
  return <section className="panel family-profiles"><div className="card-heading"><div><span className="eyebrow">Chacun son parcours</span><h2>Mes enfants</h2></div><span className="icon-tile"><Icon name="users" /></span></div>
    {children.length ? <label>Profil enfant<select value={selected} disabled={busy} onChange={event => void select(event.target.value)}>
      {children.map(child => <option key={child.id} value={child.id}>{child.username}</option>)}
    </select></label> : <p>Ajoute un enfant pour commencer.</p>}
    <details className="add-profile"><summary>Ajouter un enfant</summary><form onSubmit={event => void add(event)}>
      <label>Identifiant de l’enfant<input name="username" required minLength={3} maxLength={80} pattern="[A-Za-z0-9_.-]+" autoComplete="off" /></label>
      <label>Mot de passe de l’enfant<input name="password" type="password" required minLength={12} maxLength={72} autoComplete="new-password" /></label>
      <small>Au moins 12 caractères. Conserve ces identifiants pour la connexion de l’enfant.</small>
      <button disabled={busy}>Ajouter l’enfant</button>
    </form></details>{error && <p role="alert">{error}</p>}{notice && <p role="status">{notice}</p>}
  </section>;
}
