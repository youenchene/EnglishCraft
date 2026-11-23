# Architecture

## État implémenté

Le socle actif est décrit dans [la décision V1](decisions/0003-core-state-sessions-and-reviews.md).
Les sections « Parcours verbes irréguliers » et « Sécurité actuelle » ci-dessous
décrivent le prototype conservé, non câblé dans le serveur courant.

EnglishCraft est un monolithe modulaire Go. Une seule image sert l’API JSON et les assets React compilés. PostgreSQL est le seul service de données.

Les dépendances suivent `Adapters → Application → Domain` :

- `internal/*/domain` : entités, value objects et règles déterministes en Go pur ;
- `internal/*/application` : cas d’usage et ports sortants définis par leurs consommateurs ;
- `internal/*/adapters` : HTTP, PostgreSQL et vérification des jetons ;
- `internal/bootstrap` : composition explicite des dépendances ;
- `internal/platform` : configuration, migration, horloge et identifiants techniques ;
- `internal/shared/domain` : petit noyau partagé limité à `SkillID`.

Le test `test/architecture/imports_test.go` refuse les imports des adapters ou de la plateforme depuis l’application, et tout import non partagé depuis le domaine.

## Parcours verbes irréguliers

1. L’adapter HTTP authentifie le jeton et vérifie le rôle enfant.
2. `practice.Service` demande un verbe au port `VerbCatalog`.
3. `IrregularVerb.Check` corrige les deux formes sans IA.
4. `SkillProgress.RecordRecall` applique les seuils et calcule la révision.
5. Le port `AttemptRecorder` persiste tentative et progression dans une transaction.

Les DTO HTTP et les lignes PostgreSQL sont reconstruits en objets du domaine ; ils ne traversent pas ses frontières.

## Contextes

- **Curriculum** : compétences, objectifs, prérequis, modules et bibliothèque de 36 verbes.
- **Learning** : exercices fermés/ouverts, évaluations, révisions, progression et orchestration `application/core`.
- **Identity** : mots de passe bcrypt, sessions PostgreSQL, parent et enfant provisionnés par commande administrative.
- **Missions** : trois missions versionnées, étapes, défis, récompenses et progression.
- **AI Tutoring** : contexte pédagogique, tutorat texte consultatif, outils autorisés via Learning, plan déterministe, imports texte avec validation parentale, ports multi-providers, contrôle WebRTC et ledger de réservations. Voir [ADR 0004](decisions/0004-ai-advice-and-server-authority.md).

La voix ElevenLabs Speech Engine est un second transport WebRTC, avec callback
WebSocket signé traité par le même processus Go et tutorat texte budgété. Voir
[ADR 0006](decisions/0006-elevenlabs-speech-engine-and-text-tutor.md). Le
laboratoire Node sous `scripts/voice-lab/` reste indépendant et n'est pas un
service nécessaire au déploiement.

## Sécurité actuelle

- aucune inscription publique ;
- secrets lus uniquement depuis l’environnement serveur ;
- comparaison des empreintes de jetons en temps constant ;
- limites de taille, timeouts HTTP et erreurs génériques ;
- aucun contenu scolaire ni réponse d’élève dans les logs ;
- PostgreSQL non exposé au réseau public ;
- aucune capture ni conservation audio.

La suppression du profil enfant est parentale, avec confirmation explicite et arrêt des sessions vocales. Récupération/changement de mots de passe et export restent absents. Les plafonds IA sont configurables côté serveur et l'activation par élève est parentale ; l'enfant n'a aucun endpoint d'achat ou de configuration fournisseur.
