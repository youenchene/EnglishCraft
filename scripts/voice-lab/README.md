# Laboratoire vocal isolé

Test adulte **ElevenLabs Speech Engine + Mistral Large 4**, distinct de
EnglishCraft : aucun compte élève, aucune base de données, aucun XP. Ce n'est
pas une intégration de production.

## Lancer

Prérequis : Node 22.22+ recommandé, `cloudflared`, clés API déjà exportées
dans l'environnement serveur (`ELEVENLABS_API_KEY`, `MISTRAL_API_KEY`).
Ne pas mettre les clés dans le navigateur, Git ou les commandes affichées.

```bash
cd scripts/voice-lab
npm ci --ignore-scripts
npm run build
npm test
npm start
```

Si cloudflared n'est pas dans le PATH, définir `VOICE_LAB_TUNNEL_BIN` avec
le chemin de son exécutable. Le lanceur ouvre un tunnel temporaire, crée un
Speech Engine, puis expose l'interface **http://127.0.0.1:18082/**.

Cliquer **Démarrer le test**, autoriser le micro et parler en premier, par
exemple : “Hello Alex, can we practise the past tense?” Tester une interruption
et le bouton Arrêter. La qualité sonore se juge sur un véritable appareil.

## Limites et sécurité

- Une seule session par lancement, 60 secondes maximum côté fournisseur et
  serveur, huit réponses Mistral maximum, 120 tokens de sortie par réponse.
- Modèle `mistral-large-4`, `reasoning_effort=none` ; les blocs de raisonnement
  ne sont jamais lus à voix haute.
- Historique limité à huit messages, 1 000 caractères par message ; les
  instructions serveur ne proviennent jamais des rôles reçus dans le transcript.
- Le tunnel n'expose pas l'interface ni le POST `/session` : contrôles Host/Origin.
  Le callback `/ws` vérifie le JWT ElevenLabs via le SDK. Ne pas désactiver cette
  vérification et ne pas publier le jeton WebRTC temporaire.
- Aucune journalisation locale des transcripts ni des secrets ; seuls les
  compteurs de tokens Mistral sont affichés.
- ElevenLabs : enregistrement audio désactivé, rétention 1 jour, suppression
  audio/transcript/PII activée. Ce n'est pas une garantie de zéro rétention chez
  ElevenLabs ou Mistral. Ne pas utiliser de données personnelles ni d'élève.
- Arrêt et suppression du Speech Engine après fermeture de la session, ou
  après 20 minutes sans usage. Le lanceur ferme aussi le tunnel. Un échec de
  nettoyage affiche une instruction de suppression manuelle. Ctrl+C arrête
  l'ensemble. `.runtime.json` contient uniquement des identifiants de test,
  jamais des clés.
- API payantes : environ 0,08 $ par minute Speech Engine au tarif standard,
  plus tokens Mistral, hors taxes et promotions. Ce n'est pas un plafond de facture.

## Test réel automatisé (opt-in payant)

Le navigateur Playwright de `web/` peut fournir un micro synthétique, sans accès
à un vrai micro. Préparer un WAV PCM avec une phrase suivie de silence, puis :

```bash
VOICE_LAB_LIVE_TEST=yes VOICE_LAB_AUDIO_FILE=/chemin/phrase-synthetique.wav node live-smoke.mjs
```

Ce test consomme la session unique : relancer `npm start` avant l'essai humain.
Il vérifie une vraie connexion WebRTC, le tour transcrit→Mistral→ElevenLabs et
l'événement de lecture ; il ne prouve pas la qualité sonore physique.

SDK épinglés et dépendances séparées de l'application principale. Les SDK
officiels gèrent le transport vocal ; corrections et récompenses resteraient
en Go si ce laboratoire était ensuite intégré au produit.
