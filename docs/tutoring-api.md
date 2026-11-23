# Extension API Alex (v1)

Toutes les routes sont sous `/api/v1`, session HttpOnly obligatoire, vérification same-origin pour mutations. Aucun `studentId`, `xp`, rôle ou clé fournisseur accepté dans les DTO. Erreurs au format du socle ; 400 paramètres invalides, 401 session absente, 403 rôle/origine, 409 défi périmé, 429 admission/budget, 503 fournisseur indisponible.

| Méthode et route | Rôle / contrat |
| --- | --- |
| `GET /tutor/{mission}/intro` | Enfant ; accueil anglais local : histoire, rôle, règles et une question ; `{text}` ; aucun appel fournisseur ni réservation ; aucune identité élève dans les paramètres |
| `POST /tutor/{mission}/text` | Enfant ; `{message}` ≤1000 octets ; réponse `{text,mood,hint,mode,events}` ; `mode=local` signale le fallback non-IA |
| `POST /tutor/{mission}/tools` | Enfant ; `{name,arguments}` ; réponse `{value,events}` |
| `GET /tutor/{mission}/voice` | Enfant ; vérifie fournisseur, consentement et contexte avant accès micro ; `{available:true,provider}` ; ne réserve aucun budget, admission budgétaire au POST |
| `POST /tutor/{mission}/voice` | Enfant, consentement parent et budget ; OpenAI `{sdp,pushToTalk}` → 201 `{id,sdp,expires,provider,greeting}` ; ElevenLabs `{pushToTalk}` → 201 `{id,token,expires,provider,greeting}` ; accueil issu du contexte serveur, jeton WebRTC éphémère seulement, `Cache-Control:no-store`, aucune clé fournisseur |
| `DELETE /tutor/voice/{id}` | Enfant propriétaire uniquement ; 204 après hangup serveur |
| `GET /plan` | Enfant ; plan de 18 minutes avec raisons |
| `GET /parent/plan` | Parent ; plan de son enfant seulement |
| `GET /parent/ai` | Parent ; `{enabled,spending:[{provider,reservedMicros,calls,voiceSeconds}]}` ; date UTC ; réservations, pas facture |
| `PUT /parent/ai` | Parent ; `{enabled:boolean}` ; 204 ; false ferme les voix connues |
| `POST /parent/imports` | Parent ; `{text}` ≤20 000 octets UTF-8 ; enveloppe JSON ≤64 Ko ; 201 brouillon ; texte source non conservé |
| `GET /parent/imports` | Parent ; brouillons et fiches validées de son enfant |
| `POST /parent/imports/{id}/approve` | Parent propriétaire ; 204 ; publication des fiches uniquement |
| `GET /school-lessons` | Enfant ; uniquement les fiches approuvées de son profil |
| `DELETE /parent/child` | Parent ; `{confirmation:"DELETE STUDENT"}` ; suppression définitive du compte enfant et données associées ; 204 |

La pré-vérification vocale distingue `503 voice_disabled` (fournisseur non configuré)
et `403 ai_consent_required` (autorisation parent absente). Les erreurs du
fournisseur restent `503 unavailable`. Le navigateur ne demande le micro qu'après
une pré-vérification réussie ; le POST revérifie consentement et budget.

`GET /tutor/voice/elevenlabs/upstream` est un upgrade WebSocket **réservé au
fournisseur**, sans session navigateur : JWT HS256 ElevenLabs obligatoire, puis
admission PostgreSQL par identifiant de conversation. Il ne reçoit aucun
`studentId`. Le transcript passe par le cas d'usage tutorat texte, sans soumission
de réponse ni XP. Une nouvelle transcription annule la réponse en cours ; les
identifiants d'événements rejoués ne relancent pas le LLM. Cette route ne doit
pas être utilisée comme endpoint public de conversation ni avec une clé depuis
JavaScript.

Outils autorisés : `get_current_objective`, `get_current_challenge`, `request_hint`, `get_review_items`, `complete_mission_step` avec `{}` exclusivement. `complete_mission_step` lit l'étape, **sans mutation**. `submit_student_answer` exige `{step,exerciseId,answer,idempotencyKey}` et appelle `core.Submit` : correction, progression et récompense dans la transaction existante. La mission vient du chemin ; la session donne l'élève. Aucune requête SQL ou action générale n'est exposée.

## Événements d'affichage

Protocole `version:1` : `CHARACTER_SPEAK`/`SHOW_HINT` avec `text`, `AWARD_XP` avec `amount`, `OPEN_DOOR`, `GIVE_DIAMOND`, `SCORE_GOAL`, `MISSION_COMPLETED` sans payload. Le frontend a une union discriminée ; le serveur émet actuellement `CHARACTER_SPEAK`, `SHOW_HINT`, `OPEN_DOOR`, `AWARD_XP`, `MISSION_COMPLETED`. Diamant et football restent des types réservés, sans moteur d'animation spécifique. Un événement ne modifie jamais la progression, et les retries peuvent renvoyer les mêmes événements associés au reçu initial.

La voix retourne les résultats d'outils au fournisseur via le data channel. Le navigateur est un transport non fiable : le backend valide toujours identité, paramètres et défi actuel. Aucune décision de maîtrise n'est extraite des propos d'Alex ou des sous-titres. Les modèles générateurs/évaluateurs ne sont pas exposés par un endpoint général permettant des appels non budgétés.
# Profils enfants

Toutes les routes `/parent/` concernant un profil acceptent `?child=<id>`.
Le serveur vérifie son appartenance au parent authentifié avant toute lecture
ou mutation, y compris consentement IA, import, approbation et suppression.
`GET /parent/children` liste les profils ; `POST /parent/children` crée un compte
et son état pédagogique vide (contrat détaillé dans `api/openapi.yaml`).
