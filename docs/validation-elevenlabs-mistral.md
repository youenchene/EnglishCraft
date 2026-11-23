# Test ElevenLabs Speech Engine + Mistral Large 4 — 14 octobre 2026

## Portée

Laboratoire temporaire `scripts/voice-lab/`, sur **http://127.0.0.1:18082/**,
séparé du monolithe et de la démonstration du port 18080. Aucun accès aux comptes,
à la base PostgreSQL, aux corrections ou aux XP. Ce n'est pas encore une
intégration du fournisseur vocal dans le parcours élève.

## Résultats effectivement observés

- Accès ElevenLabs Speech Engine et liste des modèles Mistral : HTTP 200 avec
  les clés de l'environnement, sans affichage de ces clés.
- Le compte Mistral expose `mistral-large-4`. Premier essai avec raisonnement
  par défaut et petit plafond de sortie : uniquement un bloc de raisonnement,
  sans réponse finale. Avec `reasoning_effort=none` : HTTP 200, réponse courte
  présente (124 tokens d'entrée, 22 de sortie sur le test texte synthétique).
- Création effective d'un Speech Engine via le SDK officiel ; callback WSS
  via Cloudflare Tunnel temporaire, JWT vérifié par le SDK.
- POST public `/session` : HTTP 404 ; seules les requêtes locales autorisées
  peuvent obtenir un jeton WebRTC.
- Test Playwright **réel et payant**, micro synthétique fourni par fichier WAV :
  connexion WebRTC ElevenLabs, transcription, réponse Mistral et événement de
  lecture vocale reçus. Réponse Mistral de ce tour : 123 tokens d'entrée,
  15 de sortie. Ce ne sont pas des réponses HTTP simulées.
- Moteur du premier essai supprimé après arrêt, tunnel initial fermé. Un
  nouveau lancement a été préparé pour le test humain.
- `npm test` dans le laboratoire : 3 tests réussis (Host/Origin, rôles et taille
  de l'historique, exclusion des blocs de raisonnement). Build esbuild et
  vérifications de syntaxe Node réussis ; audit npm : aucune vulnérabilité signalée.
- `go test ./...` : 65 tests réussis, dont frontières d'import. Frontend principal :
  `npm test`, 13 tests réussis. Pas de nouvelle intégration PostgreSQL ou de
  nouveaux E2E métier exécutés pour ce laboratoire sans base de données.

## Restrictions

- Une session par lancement, 60 secondes maximum, huit réponses Mistral au
  maximum et 120 tokens de sortie par réponse. Aucun budget en dollars garanti :
  la consommation réelle dépend des tarifs du compte et des tokens.
- Aucun transcript enregistré localement. Paramètres ElevenLabs :
  `record_voice=false`, `retention_days=1`, `delete_audio=true`,
  `delete_transcript_and_pii=true`. Pas de promesse de zéro rétention chez les
  fournisseurs ; ne pas utiliser des données personnelles ni de compte élève.
- Qualité sonore physique, Safari/iPhone et interruptions sur vrai micro restent
  à vérifier par l'utilisateur. Le test automatisé constate un événement de
  lecture, pas une écoute humaine ni une mesure de qualité sonore.
- Node local 22.21.1 : avertissement d'installation pour une dépendance demandant
  Node 22.22+. Les builds et le test réel ont néanmoins réussi ; version plus
  récente recommandée pour un essai reproductible.

## Sécurité du shell

Trois commandes `export` seules dans `~/.zshenv` affichaient involontairement
l'environnement au démarrage. Retirées avec l'accord de l'utilisateur ; aucune
définition de clé supprimée ou modifiée. Les valeurs précédemment exposées ne
sont pas reproduites dans ce document. Renouvellement recommandé, mais non
effectué car l'utilisateur a demandé uniquement de corriger l'affichage.

## Références officielles

- [Speech Engine](https://elevenlabs.io/docs/overview/capabilities/speech-engine)
- [Quickstart](https://elevenlabs.io/docs/eleven-api/guides/cookbooks/speech-engine)
- [SDK JavaScript et authentification callback](https://elevenlabs.io/docs/eleven-api/resources/libraries/speech-engine/javascript-sdk-reference)
- [Mistral Large 4](https://docs.mistral.ai/models/mistral-large-4-0)
- [Mistral reasoning_effort](https://docs.mistral.ai/studio/conversations/reasoning)
