# Vérification Alex — 13 octobre 2026

## Statut

**Tranche locale fonctionnelle ; mission globale non déclarée terminée.** Le parcours pédagogique principal fonctionne en texte avec accompagnement local explicitement étiqueté. La conversation vocale réelle et le déploiement Dokploy distant ne sont pas validés. Aucun appel fournisseur réel n'est présenté comme testé.

## Implémenté

- Ports spécialisés et adapters HTTP OpenAI, Anthropic, Mistral ; extension Chat Completions `compatible`. Modèles configurables, annulation, timeout, retry limité et validation JSON locale.
- Contexte pédagogique serveur, tutorat Alex consultatif A1/A2, outils autorisés passant par Learning, corrections/XP/révisions toujours déterministes côté Go.
- Création Realtime serveur selon la documentation officielle consultée, audio direct WebRTC, data channel, micro volontaire, sous-titres, interruption, mute, arrêt, push-to-talk, lecture audio native et fallback texte. Ownership et échéances persistés, arrêt serveur `/hangup`.
- Plan explicable de 18 minutes ; import texte local de notions connues avec brouillons et validation parentale ; source originale non conservée ni transmise aux modèles.
- Avatar pixel-art Alex original, six états, animations désactivables, environnements aventure/stade conservés ; événements d'affichage versionnés.
- PWA avec icônes PNG/iOS, téléchargement volontaire des verbes, quiz locaux, file IndexedDB par profil et synchronisation explicite idempotente. Aucun XP définitif hors ligne.
- Consentement parental IA désactivé par défaut, plafonds/réservations quotidiens persistants, ledger sans contenu sensible, limitation de débit et suppression parentale du profil serveur.
- Docker multi-stage/distroless/nonroot, sondes, migrations contrôlables, exemple Dokploy avec PostgreSQL privé et procédures HTTPS/sauvegarde/restauration/mise à jour.

## Résultats effectivement observés

| Vérification | Résultat |
| --- | --- |
| `go test -race -timeout 60s ./...` | 61 tests réussis |
| `npm test` dans `web` | 8 tests réussis, 3 fichiers |
| `make lint` | gofmt, go vet, frontières d'import AST et ESLint réussis |
| Build React/TypeScript/Vite/PWA | Réussi, également exécuté dans Docker |
| PostgreSQL avec tag `integration` | 6 tests réussis sur base dédiée ; quotas concurrents, ownership, publication et suppression inclus |
| Migrations | Versions 1–4 présentes ; tâche `APP_COMMAND=migrate` rejouée avec succès |
| Build Docker | Réussi ; utilisateur `nonroot:nonroot`, image environ 4,9 Mo, sonde intégrée |
| Compose de vérification | Application et PostgreSQL healthy |
| E2E Chromium | 2 tests réussis en 4,4 s sur famille dédiée |
| Parcours E2E | Login, Alex texte local, indice serveur, erreur corrigée, mission Lost Diamond complète, 65 XP après reconnexion et rechargement |
| E2E hors ligne/parent | Quiz après rechargement hors ligne, double synchronisation sans XP supplémentaire, parent sur viewport 390×844, import validé, consentement IA activé/désactivé |
| Compose Dokploy | `config --quiet` réussi ; aucun port PostgreSQL public dans cet exemple |
| Sauvegarde/restauration | `pg_dump -Fc` puis `pg_restore --exit-on-error` réussis dans une base neuve de test ; 4 migrations, 2 comptes, 1 import approuvé et mission à 65 XP restaurés |
| Test externe vocal | **Ignoré** : appels facturables non autorisés, clé/SDP de test absents |

Les E2E ont reproduit puis permis de corriger deux défauts : reprise d'authentification après démarrage hors ligne ; checkbox de consentement parent avec mise à jour immédiate et rollback sur erreur. Les ressources de test ont été isolées sous le projet Docker `englishcraft-ai-verification` (localhost 18080/55433). Les comptes et la base de développement existants n'ont pas été modifiés.

## Limites / travaux restants

1. **Voix réelle, Safari macOS/iOS, Android/tablettes et Dokploy/Traefik distants non validés.** Exécuter [mobile-voice-checklist.md](mobile-voice-checklist.md). Les serveurs HTTP simulés et Chromium ne prouvent pas l'audio réel.
2. Pas de sideband ; les propositions d'outils transitent par le navigateur, avec autorisation/correction Go systématique. Pas de collecte fiable de coût vocal final. Les réservations sont des plafonds d'admission, pas des factures ; une panne VPS peut empêcher l'arrêt au délai local. Plafonds et alertes fournisseur indispensables.
3. Texte IA sans historique conversationnel multi-tour ; contexte reconstruit par requête. Pas de certification CEFR ni progression B1 durable. Le plan ne couvre pas encore un graphe complet de prérequis ou l'intégration des cours importés dans sa sélection.
4. Générateurs de leçons/exercices et évaluateur : ports/adapters présents mais non intégrés au pipeline de publication métier. Pas d'import PDF, plutôt qu'une extraction prétendument fiable.
5. Exercices importés publiés comme suggestions/fiches, pas comme nouvelles missions jouables récompensées. Édition détaillée et génération de missions restent à implémenter.
6. `GIVE_DIAMOND`/`SCORE_GOAL` : types réservés. Porte/XP/mission affichent un feedback, sans moteur complet d'animations narratives. Pas de lip-sync.
7. Hors ligne limité aux verbes téléchargés et quiz, pas aux missions complètes. Synchronisation manuelle après connexion au même profil. Effacer les téléchargements sur appareils partagés ; aucun secret d'authentification dans IndexedDB.
8. Suppression serveur n'efface ni sauvegardes, ni téléchargements sur d'autres appareils, ni données potentiellement retenues par les fournisseurs. Organiser leur rétention/effacement séparément. Gestion multi-enfants et récupération de mot de passe absentes.
9. `.env.example` ancien conservé après refus d'édition par les permissions ; utiliser [environment.md](environment.md), pas les jetons statiques du prototype.

## Déploiement et activation

[deployment.md](deployment.md) : PostgreSQL privé → build → sauvegarde → tâche `APP_COMMAND=migrate` → provisionnement initial → application `AUTO_MIGRATE=false`, `COOKIE_SECURE=true` → HTTPS Traefik → readiness/login/mission → configuration APIs → consentement parent → validation vocale/appareils.

[environment.md](environment.md) détaille les variables. [ai-providers.md](ai-providers.md) détaille clés, choix des modèles et documentation officielle consultée. [tutoring-api.md](tutoring-api.md) décrit les routes et événements. Ne pas ouvrir la voix à l'enfant sur la seule base de ce rapport local.
