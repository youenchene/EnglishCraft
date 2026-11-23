# Diagnostic voix — 13 octobre 2026

## Causes et corrections

- Instance `http://127.0.0.1:18080` : `AI_PROVIDER=disabled`, `VOICE_ENABLED` absent. Aucun fournisseur vocal n'était composé ; le refus venait du serveur, pas d'une permission micro. Aucune clé OpenAI disponible dans l'environnement inspecté.
- Avant accès micro, un GET authentifié vérifie désormais fournisseur, consentement et contexte. Les réponses distinguent configuration manquante, consentement absent et échec fournisseur. Le POST garde ses contrôles et réservations.
- La lecture locale sélectionnait simplement la première voix en-GB. Elle attend maintenant le chargement des voix, privilégie les voix identifiées comme naturelles/améliorées ou distantes, garde le même identifiant de voix entre les lectures et utilise une vitesse normale. Cette sélection est une heuristique, pas une garantie de qualité audio.

## Vérifications

- Régression serveur : test initial en échec avec `AI unavailable`, puis raison explicite fournisseur désactivé.
- Régression navigateur : test initial demandant le micro malgré un fournisseur désactivé ; corrigé, sans demande de permission. Tests de chargement et de stabilité de voix.
- `npm test` : 13 tests réussis. `npm run build`, `npm run lint` et build Docker réussis.
- `go test ./...` : 65 tests réussis, y compris frontières d'import ; `go vet ./...` réussi.
- `E2E_BASE_URL=http://127.0.0.1:18080 npm run test:e2e` : 9 tests réussis, 2 scénarios réels ignorés faute d'identifiants E2E. HTTP et micro simulés dans les contrats UI ; aucun appel vocal cloud.
- Première invocation E2E sans URL explicite : 8 échecs sur l'autre instance 8080 ; relance réussie contre 18080.
- Intégration PostgreSQL : `TEST_DATABASE_URL=… go test -tags=integration ./test/integration/...` : 7 tests réussis contre la base de test du port 55433.
- Nouvelle image publiée sur 18080 en réutilisant l'environnement du conteneur, sans afficher de secrets ni recréer PostgreSQL. Instance 8080 inchangée.

## Limites

La conversation micro reste désactivée tant que `VOICE_ENABLED=true`, une clé
OpenAI serveur, un modèle accessible, des budgets adaptés et l'autorisation
parentale ne sont pas configurés. Aucune écoute physique, validation Safari ou
conversation WebRTC fournisseur réelle n'a été effectuée. La qualité finale
dépend des voix installées/disponibles sur l'appareil.
