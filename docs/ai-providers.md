# Fournisseurs IA

Les adapters HTTP OpenAI, Anthropic et Mistral sont implémentés dans `internal/tutoring/adapters/models`. Aucun SDK ni type fournisseur ne traverse un port applicatif. Le fournisseur `compatible` accepte un endpoint HTTPS compatible Chat Completions pour une extension Qwen/DeepSeek ; la compatibilité réelle doit être vérifiée avec le modèle choisi.

Les ports `LessonGenerator`, `ExerciseGenerator`, `AnswerEvaluator`, `TutorConversation`, `SpeechSessionGateway` exposent des données applicatives. Les trois premiers ont des implémentations HTTP mais ne sont **pas encore intégrés au parcours de publication** ; l'import courant reste déterministe. L'évaluation IA est consultative, exige `needsHumanReview=true` et n'accorde aucun XP.

Les instructions d'Alex restent dans l'adapter, jamais dans le domaine. JSON strict (champs inconnus, taille, valeurs, refus et sorties tronquées rejetés), timeout total de 10 secondes, annulation propagée, un retry sur 429/503 uniquement. Les erreurs exposées ne contiennent ni corps fournisseur ni clé. Les tests HTTP utilisent des serveurs locaux explicitement simulés.

## Activation

1. Créer une clé projet séparée, avec budget et alertes chez le fournisseur. Vérifier ses règles d'utilisation pour les mineurs, sa rétention et les exigences de consentement ; EnglishCraft ne garantit pas l'absence de rétention côté fournisseur.
2. Choisir `AI_PROVIDER=openai|anthropic|mistral|compatible`, `AI_TEXT_MODEL`, la clé correspondante et éventuellement `AI_BASE_URL` (HTTPS uniquement, sauf serveurs de test loopback).
3. Définir les budgets et réservations décrits dans [environment.md](environment.md). Zéro interdit les appels ; jamais de paiement côté enfant.
4. Connecter le parent et activer « Autoriser les fournisseurs IA ». La désactivation empêche de nouvelles réservations et ferme les sessions vocales connues.
5. Pour la voix : `VOICE_ENABLED=true`, `OPENAI_API_KEY`, `AI_VOICE_MODEL=gpt-realtime-2.1` si disponible sur votre compte. Tester avec une véritable connexion navigateur avant toute exposition à l'enfant.

### ElevenLabs Speech Engine + Mistral

Alternative à la voix OpenAI : `AI_PROVIDER=mistral`,
`AI_TEXT_MODEL=mistral-large-4`, `MISTRAL_API_KEY`, `VOICE_ENABLED=true`,
`VOICE_PROVIDER=elevenlabs`, `ELEVENLABS_API_KEY`, `ELEVENLABS_ENGINE_ID`.
L'overlay `compose.elevenlabs.yaml` fournit des budgets d'admission conservateurs
à adapter à vos tarifs. Aucune clé OpenAI n'est nécessaire. Le tutorat texte
Mistral utilise `reasoning_effort=none` pour obtenir directement le JSON final.

1. Avec la même clé ElevenLabs que celle du serveur, configurer la ressource :
   `ELEVENLABS_CALLBACK_URL=wss://votre-domaine/api/v1/tutor/voice/elevenlabs/upstream node scripts/configure-elevenlabs.mjs`.
   Le script affiche l'identifiant public `seng_…` à passer au serveur.
   Fournir `ELEVENLABS_ENGINE_ID` au script pour mettre à jour uniquement une
   ressource existante ; sinon il en crée une nouvelle.
2. L'adresse doit être joignable par ElevenLabs. Sur localhost, garder un
   Cloudflare Tunnel temporaire actif vers le port de l'app et utiliser son
   domaine dans le callback. En production, préférer un domaine WSS stable ;
   un proxy doit autoriser les upgrades WebSocket et la durée des sessions.
3. La durée du moteur ne doit pas dépasser `VOICE_MAX_SECONDS`. Le script règle
   cette durée (300 s par défaut), désactive l'enregistrement audio, limite la
   rétention à 1 jour et active suppression audio/transcript/PII. Le serveur
   refuse de créer un jeton si ces limites ne sont pas respectées. Cela ne
   garantit pas le zéro rétention des fournisseurs.
4. Activer le consentement du profil dans « Réglages IA ». Aucune activation
   automatique des profils existants. Chaque tour revérifie le consentement et
   reserve/audite le budget du fournisseur texte ; la durée vocale est réservée
   lors de l'admission, séparément sous `elevenlabs`.

La conversation reprend le contexte de la mission et jusqu'à 12 messages
transitoires, sans persistance de transcript. La vue Adventure est désormais
entièrement conversationnelle : plus de formulaire de défi, de scène séparée
ou de barre de progression. Elle ne transforme pas automatiquement la discussion
en preuves pédagogiques ni en XP. Les révisions et les cas d'usage déterministes
Learning restent disponibles hors de cette vue. Le mode texte reste un choix
explicite de conversation ; la lecture de synthèse navigateur a été retirée
de la vue Adventure.

Alex prend la parole en premier **après** le clic sur « Commencer la conversation »
et l'autorisation micro. Le serveur fournit l'accueil depuis le contexte de la
mission. `scripts/configure-elevenlabs.mjs` active `overrides.first_message` ;
le serveur refuse de créer un jeton si ce réglage manque. Aucun micro ni appel
payant n'est déclenché simplement en naviguant vers l'aventure.

L'accueil n'est plus une concaténation du dialogue bilingue du contenu scolaire.
Un script anglais dédié pose le décor pour le diamant, le village ou l'interview
de football, puis explique réponses courtes, une question à la fois, `help`,
`repeat` et interruptions. Son aperçu authentifié `/tutor/{mission}/intro` est
local et gratuit. Les réponses Mistral sont en anglais par défaut ; le français
n'est utilisé que si l'élève le demande explicitement. Le personnage affiché est
une parodie originale d'Alex de Minecraft, dessinée en SVG sans asset officiel.

Le preset ElevenLabs réserve désormais jusqu'à **2 USD/jour/élève** (tous
fournisseurs cumulés), avec 2 USD/jour par fournisseur, 5 min/session et
15 min de durée **réservée** quotidienne. Les réservations ambiguës restent
comptées : le ledger n'est pas une facture et ne garantit pas un plafond externe.
Pour interrompre la voix ElevenLabs, commencer à parler ; un casque limite
l'écho. Le serveur annule les réponses texte obsolètes, le fournisseur gère
l'interruption audio. Push-to-talk utilise le mute du SDK et la détection de fin
de tour ; sa fluidité sur un appareil physique reste à tester.

Après un redémarrage, les admissions et échéances survivent dans PostgreSQL.
Un callback signé ne peut reprendre une session arrêtée/sans consentement/
expirée. Fermer toutes les sessions avant de changer `VOICE_PROVIDER`.
La réservation des tokens texte reste conservée après une interruption ambiguë.

Scopes ElevenLabs pour le provisionnement : Agents/Conversational AI lecture
et écriture ; Voix lecture uniquement si vous souhaitez sélectionner une voix.
Ne pas autoriser clonage ou administration globale. La clé du créateur du
moteur doit être celle utilisée pour vérifier le callback JWT.

Le mode texte IA traite chaque message avec le contexte courant ; il ne conserve **pas d'historique conversationnel multi-tour**. Le fallback `mode=local` est un accompagnement statique explicitement étiqueté, pas un faux appel API. Les défis restent utilisables sans fournisseur.

## Contrat vocal vérifié dans la documentation officielle le 13 octobre 2026

- [WebRTC](https://developers.openai.com/api/docs/guides/realtime-webrtc) : section **Realtime**, et non l'API GPT-Live distincte ; multipart `sdp` + `session`, `POST /v1/realtime/calls`, HTTP 201, SDP texte, ID dans `Location`.
- [Conversations](https://developers.openai.com/api/docs/guides/realtime-conversations) : `response.create`, `response.cancel`, `output_audio_buffer.clear`, `response.function_call_arguments.done`, `conversation.item.create`/`function_call_output`, `response.output_audio_transcript.delta`, événements speech/output buffer. Push-to-talk avec `turn_detection=null`, clear/commit/create.
- [Contrôle serveur](https://developers.openai.com/api/docs/guides/realtime-server-controls) et [hangup](https://developers.openai.com/api/reference/resources/realtime/subresources/calls/methods/hangup) : arrêt serveur du call WebRTC ; sideband disponible mais non implémentée ici.
- [Anthropic Messages](https://platform.claude.com/docs/en/api/messages), [Mistral Chat](https://docs.mistral.ai/api/endpoint/chat).

EnglishCraft ne stocke aucun flux audio. La durée est contrôlée par un nettoyeur serveur toutes les secondes et des échéances PostgreSQL, ainsi que par le navigateur. **Un VPS indisponible ne peut pas garantir l'arrêt au délai local** ; prévoir les plafonds du fournisseur. La facturation vocale précise et l'usage final ne sont pas collectés sans sideband : les réservations ne doivent pas être affichées comme coût réel.
