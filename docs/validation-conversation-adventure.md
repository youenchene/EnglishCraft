# Adventure conversationnelle et Alex animé — 14 octobre 2026

## Changements

- Adventure devient un canevas plein écran : portrait d'Alex, dialogue,
  boutons voix/texte et retour. Suppression de la scène secondaire, du panneau
  de défi, de la progression de mission et des contrôles d'indice/lecture locale.
  Navigation de tableau de bord et footer masqués uniquement dans la conversation ;
  accueil, espace parent, réglages et révisions conservés.
- Nouveau portrait vectoriel original, grand et responsive : visage expressif,
  clignement, respiration, léger mouvement de tête, gestes et bouche animée
  durant l'état parlant. Écoute/réflexion ont leurs poses. Respect de
  `prefers-reduced-motion` ; pas de photo, animation ou police distante.
- Historique affiché uniquement en mémoire, jusqu'à 12 messages. Le mode texte
  est un choix explicite et reste disponible si le micro/fournisseur échoue.
- Alex prend la parole en premier après démarrage volontaire du micro.
  `greeting` vient du contexte de mission serveur ; l'override ElevenLabs du
  premier message est activé par configuration, et vérifié avant émission du
  jeton. Aucun micro ni consommation déclenché à l'ouverture de la vue.

## Vérifications observées

- Frontend : 15 tests réussis, TypeScript/Vite/PWA et ESLint réussis.
- `go test ./...` : 70 tests réussis ; `go vet ./...` réussi.
- `go test -race ./internal/tutoring/... ./test/architecture/...` : 34 tests
  réussis ; frontières d'import et annulation des réponses obsolètes conservées.
- PostgreSQL : 7 tests d'intégration réussis sur schémas isolés du port 55433.
- Build Docker réussi ; avertissement préexistant sur la taille du chunk SDK
  ElevenLabs différé (environ 620 Ko / 164 Ko gzip).
- Famille fraîche de fixture sur 18081/55434 : 12 E2E réussis, avec les deux
  parcours Go/PostgreSQL réels. Le contrat pédagogique de mission est testé par
  API authentifiée, plutôt que de réintroduire un formulaire retiré de l'UI.
- Présentation vérifiée à 390, 768 et 1440 px : pas de débordement horizontal,
  bulle de conversation suffisamment large, avatar visible, aucun panneau de
  défi ni navigation de dashboard dans l'aventure. Captures relues.
- Relance ciblée des 7 tests de présentation après ajout d'une assertion sur
  l'état visuel du sélecteur voix/texte : tous réussis. Fixture PostgreSQL,
  moteur ElevenLabs et tunnel de test nettoyés après vérification, démo conservée.
- Test réel ElevenLabs avec **micro synthétique silencieux** : événement
  « Alex parle » reçu sans aucune phrase de l'élève. XP et historique inchangés ;
  révocation parentale bloque la session suivante. L'accueil ne nécessite pas
  un appel Mistral : il est fourni par le serveur et lu par ElevenLabs.

Lors de la première vérification visuelle, la classe `alex` d'une bulle entrait
en collision avec l'ancien personnage de scène (position absolue, largeur 60 px).
Classes de locuteurs désormais préfixées `from-…`, et test de dimensions ajouté.
Un premier E2E attendait uniquement le texte d'erreur sans le nouveau bouton de
fallback ; assertion adaptée à la présence de ce bouton, puis suite complète
réussie sur une fixture fraîche.

## Limites et publication

Publication sur **http://127.0.0.1:18080/** avec comptes, mots de passe,
consentements, budgets et progression existants conservés. Moteur ElevenLabs
mis à jour avec cette même clé serveur pour autoriser l'accueil initial.
Le tunnel WSS local doit rester actif.

Les conversations ne valident pas automatiquement de défis ou missions et ne
rapportent pas de XP : aucune maîtrise n'est inventée à partir du dialogue.
Les opérations Learning existantes et les révisions restent déterministes.
L'animation de bouche suit l'état de parole, pas une synchronisation phonémique.
Qualité sonore réelle, interruptions sur vrai micro et Safari/iPhone restent
à vérifier sur l'appareil ; les tests automatisés ne prétendent pas les mesurer.
