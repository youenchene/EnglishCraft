# ElevenLabs vocal, tutorat texte et identité serveur

La voix naturelle validée avec ElevenLabs Speech Engine utilise le fournisseur
texte configuré (Mistral Large 4 pour la démo), plutôt qu'OpenAI Realtime. Son
callback WebSocket signé est traité dans le monolithe Go, sans service Node
supplémentaire : l'identifiant de conversation retrouve une admission PostgreSQL
qui impose l'élève, la mission, le consentement et l'échéance ; chaque réponse
utilise les cas d'usage et budgets du tutorat texte.

Le moteur ElevenLabs est une ressource configurée côté exploitation avec une
adresse WSS joignable, un plafond de durée et une rétention limitée. Les jetons
WebRTC restent en mémoire et ne sont pas persistés. Le transcript est consultatif,
ne soumet pas de réponse et n'accorde jamais de XP. Les outils et formulaires
Learning conservent leurs contrôles existants.

La reconnexion d'une session arrêtée est refusée par suppression de son admission
avant fermeture du callback. Une nouvelle transcription annule l'appel texte
précédent et interdit de réémettre sa réponse obsolète. Le fournisseur gère
l'interruption audio ; sa rapidité perceptible reste à vérifier sur l'appareil.
OpenAI Realtime demeure un autre transport, sélectionné par `VOICE_PROVIDER`.
Changer de fournisseur exige de fermer d'abord les anciennes sessions : les
identifiants d'un transport ne permettent pas le hangup de l'autre.
