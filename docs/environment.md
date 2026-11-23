# Variables d'environnement

Les secrets sont uniquement serveur, jamais `VITE_*`, jamais committés. `.env.example` contient encore les anciennes variables de jetons du prototype : **ne pas utiliser `APP_PARENT_TOKEN` ou `APP_CHILD_TOKEN`**, elles ne sont pas câblées dans le serveur actif.

| Variable | Usage / valeur par défaut |
| --- | --- |
| `DATABASE_URL` | Connexion PostgreSQL obligatoire ; encoder les caractères spéciaux du mot de passe dans l'URL |
| `APP_ADDRESS` | `:8080` ; la sonde Docker intégrée suppose ce port |
| `COOKIE_SECURE` | `true` ; `false` uniquement sur localhost HTTP |
| `AUTO_MIGRATE` | `true` pour conserver le démarrage local ; `false` en production |
| `APP_COMMAND` | `migrate`, `provision`, ou absent pour le serveur |
| `INITIAL_PARENT_USERNAME/PASSWORD` | Provisionnement initial uniquement |
| `INITIAL_CHILD_USERNAME/PASSWORD` | Provisionnement initial uniquement ; mots de passe distincts, 12–72 octets |
| `AI_PROVIDER` | Absent/`disabled`, `openai`, `anthropic`, `mistral`, `compatible` |
| `AI_TEXT_MODEL` | Obligatoire si provider activé ; nom validé auprès du compte fournisseur |
| `OPENAI_API_KEY` | Clé projet OpenAI, texte et/ou voix |
| `ANTHROPIC_API_KEY` | Clé Anthropic |
| `MISTRAL_API_KEY` | Clé Mistral |
| `AI_API_KEY` | Clé du provider `compatible` |
| `AI_BASE_URL` | Endpoint `/v1` optionnel ; requis pour `compatible` ; ne peut pas être fourni par un client HTTP |
| `VOICE_ENABLED` | `false` (seul `true` active) |
| `VOICE_PROVIDER` | `openai` par défaut, ou `elevenlabs` ; indépendant du fournisseur texte |
| `AI_VOICE_MODEL` | Obligatoire pour voix OpenAI ; exemple documenté `gpt-realtime-2.1` |
| `ELEVENLABS_API_KEY` | Clé serveur ElevenLabs, utilisée aussi pour vérifier les JWT du callback |
| `ELEVENLABS_ENGINE_ID` | Ressource Speech Engine `seng_…` configurée avec cette même clé |
| `ELEVENLABS_CALLBACK_URL` | Utilisée uniquement par `scripts/configure-elevenlabs.mjs` : WSS public se terminant par `/api/v1/tutor/voice/elevenlabs/upstream` |
| `VOICE_MAX_SECONDS` | `300`, maximum configurable 1200 |
| `VOICE_DAILY_SECONDS` | `900`, maximum configurable 3600 ; quota réservé par élève, jour UTC |
| `AI_PROVIDER_DAILY_MICROS` | `0` : plafond quotidien par fournisseur en millionièmes USD |
| `AI_STUDENT_DAILY_MICROS` | `0` : plafond quotidien par élève, tous fournisseurs cumulés |
| `AI_TEXT_RESERVE_MICROS` | `0` : plafond conservateur par appel texte, doit couvrir deux tentatives, 24 Ko d'entrée et 1200 tokens de sortie par tentative |
| `AI_VOICE_RESERVE_MICROS` | `0` : plafond conservateur par session complète, durée maximale, entrée/sortie audio/texte, transcription et outils compris |

1 000 000 micros = 1 USD. Les valeurs non nulles doivent être déterminées avec les prix du modèle choisi. Elles restent réservées après une erreur potentiellement facturée ; le ledger conserve seulement fournisseur, date, tokens, plafond et statut, pas le contenu. Les limites locales contrôlent l'admission, **pas la facture externe** : configurer aussi plafonds et alertes fournisseur. Les cookies et consentements sont requis indépendamment des clés.
