# Intégration ElevenLabs + Mistral — 14 octobre 2026

## Résultat

Conversation ElevenLabs intégrée au professeur Alex, dans le monolithe Go et
le frontend React. Mistral Large 4 fournit les réponses structurées avec le
contexte de la mission ; OpenAI Realtime reste disponible en configuration
alternative. Aucun service Node additionnel n'est requis pour l'application.

Le GET authentifié de pré-vérification annonce le transport. Le POST réserve
le budget vocal et la durée puis retourne un jeton WebRTC éphémère ; seuls
l'identifiant de conversation, l'élève, la mission et l'échéance sont persistés.
Le callback fournisseur vérifie JWT, admission, consentement et deadline ; les
tours sont limités et audités comme du tutorat texte, sans mutation pédagogique.
Annulation de la réponse précédente lors d'une nouvelle transcription et rejet
des événements dupliqués. Le bouton de lecture locale reste distinct.

## Vérifications observées

- `go test ./...` : 70 tests réussis après ajout des tests ElevenLabs ;
  frontières d'import incluses. `go vet ./...` réussi.
- `go test -race ./internal/tutoring/... ./test/architecture/...` : 34 tests
  réussis, incluant annulation/rejeu/hangup de callbacks WebSocket locaux.
- Tests nouveaux : JWT altéré/expiré/mauvais algorithme, callback anonyme ou
  sans admission, durée/rétention provider non conformes, identité issue du
  stockage, consentement révoqué, deadline expirée, session arrêtée, aucun XP.
- Frontend : 14 tests réussis, TypeScript/Vite/PWA et ESLint réussis. Test
  ElevenLabs : jeton en mémoire, aucune écriture Storage, libération du flux de
  permission, mute/PTT et arrêt idempotent de la session propriétaire.
- `TEST_DATABASE_URL=… go test -tags=integration ./test/integration/...` :
  7 tests PostgreSQL réussis, dont lookup d'admission et rejet d'identifiant inconnu.
- Build Docker réussi. Avertissement Vite : chunk SDK ElevenLabs différé d'environ
  620 Ko (164 Ko gzip). Le bundle initial reste d'environ 193 Ko.
- Famille fraîche dédiée `englishcraft-elevenlabs-e2e`, ports 18081/55434 :
  les 11 E2E existants réussissent, y compris les 2 scénarios avec vrai Go/PostgreSQL
  (entraînement, mission à 65 XP, hors ligne/synchronisation, espace parent).
- `web/e2e/voice-live.mjs` opt-in payant, micro **synthétique** : compte élève
  authentifié → admission Go → vraie connexion ElevenLabs WebRTC → Mistral →
  événement « Alex parle ». XP et historique identiques avant/après. Révocation
  du consentement parent : une nouvelle session est refusée.
- Premier passage du test vocal réel : échec du sélecteur après arrêt, car les
  options vocales s'étaient repliées. Le test a été corrigé pour les rouvrir ;
  second passage réussi. Ce n'était pas un échec du round trip fournisseur.
- Fixture, base et tunnel de test supprimés après vérification ; moteur ElevenLabs
  de test supprimé (HTTP 204). Aucun compte de test ajouté à la famille réelle.
- Démo finale : HTTP 200 ; callback fournisseur sans JWT : HTTP 401.

## Publication et limites

Nouvelle image publiée sur **http://127.0.0.1:18080/** avec clés serveur Mistral
et ElevenLabs, consentements et base existante conservés, sans affichage de
secrets. Sessions limitées à 5 min, durée quotidienne réservée à 15 min par
élève, réserve vocale 0,40 USD/session et texte 0,01 USD/appel. Plafonds locaux :
2 USD/jour par fournisseur, 1 USD/jour par élève tous fournisseurs cumulés.
Ce sont des réservations d'admission, pas une facture ni un plafond fournisseur
garanti. La première session n'est possible qu'après consentement du parent.

Le callback local dépend d'un **tunnel Cloudflare actif**. Une adresse WSS
stable reste nécessaire pour la production. Réglages moteur : enregistrement
audio désactivé, rétention 1 jour, suppressions audio/transcript/PII activées ;
aucune garantie de zéro rétention chez ElevenLabs ou Mistral. La suppression
locale d'un profil ne constitue pas une suppression garantie de traces chez
les fournisseurs ; la conformité pour mineurs reste à valider avant exposition.

La rapidité perçue des interruptions, le vrai micro, Safari/iPhone et le
push-to-talk physique ne sont pas validés par les tests synthétiques. Les
réponses vocales restent des conseils : pas de soumission automatique des défis,
pas d'évaluation fiable de prononciation et aucun XP attribué par l'IA.
