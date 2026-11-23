# EnglishCraft

EnglishCraft est une PWA familiale pour apprendre l'anglais par des aventures textuelles. Le socle fonctionne localement, sans API IA.

## Fonctionnalités disponibles

- comptes parent/enfant, mots de passe bcrypt et sessions HttpOnly, sans inscription publique ;
- 36 verbes fréquents avec traductions, exemples et catégories ;
- trois missions : The Lost Diamond, The Mysterious Village, The Post-Match Interview ;
- prétérit, participe passé, phrases à compléter, formes incorrectes, phrases guidées ;
- phrases libres enregistrées comme nécessitant une évaluation, jamais validées automatiquement ;
- XP serveur, soumissions idempotentes, progression persistante et révisions espacées ;
- Home, Adventure, Quick Training, cartes et tableau de bord parent ;
- API REST documentée dans [`api/openapi.yaml`](api/openapi.yaml) ;
- interface React responsive et installable en PWA ;
- persistance PostgreSQL et migrations versionnées ;
- image Docker unique servant le backend et le frontend compilé.

Les illustrations sont originales. Une extension Alex propose tutorat texte, adapters OpenAI/Anthropic/Mistral, voix WebRTC optionnelle, outils serveur, import texte avec validation parentale, plan de 18 minutes et quiz hors ligne synchronisables. Les providers sont désactivés par défaut. La voix réelle et le déploiement Dokploy nécessitent une validation externe : voir [les limites](docs/ai-providers.md) et [la vérification](docs/validation-ai.md).

## Démarrage local

La procédure complète (variables, création initiale des comptes, HTTPS et Docker)
est dans **[docs/local-core.md](docs/local-core.md)**. Résumé après export des variables :

```bash
make db
make migrate
make provision         # une seule fois ; mots de passe fournis via environnement
make build
make api               # http://localhost:8080
make web               # autre terminal, http://localhost:5173
```

`DATABASE_URL` est obligatoire ; `COOKIE_SECURE=false` uniquement pour localhost HTTP.

PostgreSQL écoute uniquement sur `127.0.0.1` en développement. Il n’est pas publié par le conteneur applicatif.

## Commandes

```bash
make build             # frontend puis binaire Go
make test              # tests unitaires, architecture et frontend
make test-integration  # nécessite PostgreSQL local
make lint
make audit
```

## Documentation

- [Architecture](docs/architecture.md)
- [Modèle de domaine](docs/domain-model.md)
- [Déploiement](docs/deployment.md)
- [Fournisseurs IA](docs/ai-providers.md)
- [Stratégie et résultats réels des tests](docs/testing.md)
- [Décisions d’architecture](docs/decisions/)
