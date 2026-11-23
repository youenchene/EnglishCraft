# Validation réelle — 13 octobre 2026

## Résultats

| Vérification exécutée | Résultat |
| --- | --- |
| `go test -race -count=1 ./...` | 33 cas réussis, incluant sous-tests ; aucune course détectée |
| `TEST_DATABASE_URL=… go test -race -count=1 -tags=integration ./test/integration/...` | 4 tests réussis, aucun skipped |
| `npm --prefix web test` | 5 tests frontend réussis |
| `make lint` | gofmt, go vet, imports interdits et ESLint : succès |
| `make build` | TypeScript, Vite, service worker PWA et binaire Go : succès |
| `docker compose build app` | Image construite, utilisateur `nonroot:nonroot` |
| Playwright contre Go local sur 18081 | 2 tests réussis, aucun skipped |
| Playwright après compilation finale sur 18083 | 2 tests réussis, aucun skipped |
| Playwright contre l'image Docker sur 18082 | 2 tests réussis, aucun skipped |
| Login via proxy Vite avec Origin du frontend | HTTP 200 ; Host préservé |
| Login via proxy Vite avec Origin étrangère | HTTP 403 |
| `npm audit` | 0 vulnérabilité signalée |
| Lecture YAML du contrat OpenAPI 3.1 | 13 chemins analysés ; ce n'est pas une certification exhaustive du schéma |

Les tests PostgreSQL du socle utilisent des schémas jetables et testent huit
retries simultanés du même reçu, puis deux clés distinctes sur la même étape
(un succès, un conflit). Ils vérifient rollback, persistance et révocation des
sessions, lien parent/enfant et refus d'écraser les comptes initiaux.
Un test hors réseau déroule aussi chacune des trois missions complètes et
vérifie les récompenses finales respectives : 65, 60 et 70 XP.

Le parcours navigateur choisit The Lost Diamond, soumet `goed`, voit la correction
`went`, réessaie, termine les cinq étapes, puis vérifie les **65 XP** après logout,
login et reload. Une seconde session parent consulte ces mêmes XP sur un viewport
390 × 844 et vérifie l'absence de débordement horizontal. Le navigateur est Chromium.

## Portée et limites

- Aucune API IA appelée. Aucune conversation vocale.
- Réponses libres persistées et visibles par le parent, sans validation ni XP
  automatiques. La validation humaine avec mutation de progression n'est pas livrée.
- Une seule famille et un enfant pour la commande de provisioning V1.
- Pas de récupération/changement de mot de passe dans l'interface, ni MFA.
- JSONB par élève : pratique en famille, mais journal/reçus non bornés en stockage.
- Pas de soumission hors ligne ; le service worker précache seulement les assets.
- Installation PWA et Safari sur appareils physiques non testés.
- Le workflow CI est ajouté, mais aucun résultat de CI distante n'est revendiqué.
- `.env.example` reste celui du prototype : la permission d'édition de ce fichier
  a été refusée. `docs/local-core.md` donne les variables actives et une procédure
  sans utiliser ses anciens jetons.

Les services déjà présents sur 5432 et 8080 n'ont pas été remplacés. Les essais
du socle ont utilisé des schémas PostgreSQL et des ports distincts. Aucun compte
existant n'a été réinitialisé et aucun volume n'a été supprimé.
