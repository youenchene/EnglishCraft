# EnglishCraft — règles du socle

Ce fichier formalise les frontières décrites dans `docs/architecture.md` ; aucun
AGENTS.md antérieur n'était disponible dans le worktree lors de la reprise.

## Architecture

- Conserver le monolithe modulaire : `internal/<métier>/{domain,application,adapters}`.
- Dépendances : adapters → application → domain. Domain utilise uniquement la
  bibliothèque standard et le petit noyau `internal/shared/domain`.
- Les ports sortants appartiennent à leurs cas d'utilisation consommateurs.
- Composer les dépendances explicitement dans `internal/bootstrap`.
- Séparer les DTO HTTP, contenus JSON, lignes SQL et entités métier.
- Mettre les corrections, récompenses et révisions en Go ; les handlers et SQL
  ne décident pas des résultats pédagogiques.
- Protéger les invariants par constructeurs et opérations métier, sans setters.
- Injecter l'horloge dans les cas d'utilisation. Les tests métier sont hors réseau.

## Intégrité et sécurité

- Persister soumission, progression et récompense dans une transaction.
- L'idempotence est liée à l'élève et au contenu soumis ; un retry ne rapporte
  jamais de deuxième récompense.
- L'identité de l'élève vient de la session serveur, pas du corps JSON.
- Aucune inscription publique ni dépendance IA pour le parcours fonctionnel.
- Les réponses ouvertes restent `needs_evaluation` sans évaluation fiable.
- Sessions opaques, cookies HttpOnly, SameSite, Secure en HTTPS ; aucun secret
  d'authentification dans le stockage JavaScript.

## Vérification

Exécuter tests Go, frontières d'import, frontend, intégration PostgreSQL et E2E.
Documenter les résultats effectivement observés et les commandes impossibles à
exécuter. Lire `docs/decisions/` avant une modification structurante.
