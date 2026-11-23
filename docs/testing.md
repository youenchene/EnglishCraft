# Tests

## Niveaux disponibles

- domaine : invariants, corrections strictes, récompenses et jours simulés 0, 1, 4, 11, 25 et 26 (échéances jusqu'au jour 55), sans réseau ni base ;
- application : cas d’usage testés avec ports simulés ;
- HTTP : cookies, CSRF, rôles parent/enfant, DTO anti-triche, idempotence et révocation ;
- architecture : imports interdits vérifiés par analyse AST ;
- PostgreSQL : transactions/rollback, huit retries concurrents, comptes, sessions, redémarrage et repositories sous le tag `integration` ;
- frontend : parcours de connexion et chargement d’exercice avec Vitest/Testing Library.
- end-to-end : entraînement, mission complète avec erreur corrigée, 65 XP après reconnexion et suivi parent sur viewport mobile, dans Chromium.
- tutoring : contexte/limites/fallback, paramètres et identité des outils, échéances vocales avec fakes explicites ; adapters HTTP simulés multi-provider, JSON invalide, retry, annulation et timeout ; aucun appel fournisseur réel par défaut.
- PostgreSQL tutoring : vingt réservations concurrentes, quota, ownership vocal/import, validation parentale et suppression des données.
- E2E étendu : accompagnement texte local, indice serveur, installation du service worker, quiz hors ligne, double synchronisation et import parent validé.

## Exécution

```bash
make test
docker compose up -d postgres
make test-integration
# Exporter les identifiants de famille E2E et l'URL ; voir local-core.md.
make test-e2e
make lint
make audit
```

Les tests métier ne nécessitent ni PostgreSQL, ni réseau, ni API IA. Le test d’intégration utilise `TEST_DATABASE_URL` et génère un identifiant élève unique pour rester isolé.

Voir [validation-core.md](validation-core.md) pour les résultats réellement obtenus.

La tranche Alex est documentée dans [validation-ai.md](validation-ai.md) : tests de ports avec fakes explicites, adapters avec serveurs HTTP simulés, JSON invalide, timeout/annulation/retry, outils autorisés, quotas PostgreSQL concurrents et E2E hors ligne/import parent.

Les tests facturables sont sous le tag `external`. Ils exigent `ALLOW_BILLABLE_TESTS=true`, une clé OpenAI, `AI_VOICE_MODEL` et `TEST_REALTIME_SDP_FILE` issu d'un navigateur réel. Un test ignoré ne valide pas l'intégration réelle. Vérifications Safari/iOS : [mobile-voice-checklist.md](mobile-voice-checklist.md).
Voir [validation-ai.md](validation-ai.md) pour la tranche Alex, ses résultats et ses limites. Les tests facturables sont exclusivement sous le tag `external` : `ALLOW_BILLABLE_TESTS=true`, clé OpenAI, modèle voix et `TEST_REALTIME_SDP_FILE` provenant d'un vrai navigateur sont requis. Un test ignoré ne valide pas l'intégration réelle. Checklist appareils : [mobile-voice-checklist.md](mobile-voice-checklist.md).
