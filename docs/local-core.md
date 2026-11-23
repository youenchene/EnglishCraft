# Exécution du socle

Prérequis : Go 1.27, Node 22, Docker Compose. PostgreSQL 17 est utilisé en local.

## Variables

| Variable | Usage |
| --- | --- |
| `DATABASE_URL` | URL PostgreSQL, obligatoire pour Go |
| `POSTGRES_PASSWORD` | Mot de passe de PostgreSQL dans Compose |
| `APP_ADDRESS` | Adresse Go, `:8080` par défaut |
| `API_PROXY_TARGET` | Cible du proxy Vite, `http://localhost:8080` par défaut |
| `COOKIE_SECURE` | `true` par défaut ; `false` uniquement pour localhost HTTP |
| `INITIAL_PARENT_USERNAME` | Identifiant parent, 3–80 caractères ASCII lettres/chiffres/`_.-`, création initiale uniquement |
| `INITIAL_PARENT_PASSWORD` | Mot de passe parent de 12 à 72 octets |
| `INITIAL_CHILD_USERNAME` | Identifiant enfant distinct |
| `INITIAL_CHILD_PASSWORD` | Mot de passe enfant distinct, 12 à 72 octets |
| `APP_COMMAND` | `migrate` ou `provision` pour commande administrative |
| `TEST_DATABASE_URL` | URL de la base de tests PostgreSQL |
| `E2E_BASE_URL` | Serveur existant utilisé par Playwright |
| `E2E_CHILD_USERNAME/PASSWORD` | Compte enfant du jeu de tests E2E |
| `E2E_PARENT_USERNAME/PASSWORD` | Compte parent du jeu de tests E2E |

Les anciennes variables `APP_PARENT_TOKEN` et `APP_CHILD_TOKEN` ne sont plus
utilisées. Le `.env.example` hérité les contient encore : sa modification a été
refusée par les permissions du harness pendant cette implémentation. Utiliser
les exports ci-dessous plutôt que ses anciens codes d'accès.

## PostgreSQL et comptes

Générer localement un secret, le saisir sans l'inscrire dans l'historique shell :

```bash
read -s POSTGRES_PASSWORD
export POSTGRES_PASSWORD
export DATABASE_URL="postgres://englishcraft:${POSTGRES_PASSWORD}@127.0.0.1:5432/englishcraft?sslmode=disable"
export COOKIE_SECURE=false
make db
make migrate
export INITIAL_PARENT_USERNAME=parent INITIAL_CHILD_USERNAME=alex
read -s INITIAL_PARENT_PASSWORD
read -s INITIAL_CHILD_PASSWORD
export INITIAL_PARENT_PASSWORD INITIAL_CHILD_PASSWORD
make provision
unset INITIAL_PARENT_PASSWORD INITIAL_CHILD_PASSWORD
```

Utiliser un mot de passe généré sans caractères réservés d'URL ou l'encoder dans
`DATABASE_URL`. Si le volume PostgreSQL existait déjà, son ancien mot de passe
reste en vigueur ; changer une variable ne change pas le volume. Ne pas supprimer
un volume pour contourner ce problème.

## Mode développement

```bash
make build         # frontend incorporé au binaire + compilation Go
make api           # terminal 1, avec DATABASE_URL et COOKIE_SECURE exportés
make web           # terminal 2 ; http://localhost:5173
```

Le proxy Vite transmet `/api` à Go sur 8080 ; les cookies restent same-origin.
Le serveur Go sert aussi directement la PWA compilée à `http://localhost:8080`.
Connexion avec les identifiants parent/enfant créés ci-dessus.

## Image unique

```bash
docker compose up -d --build --wait
docker compose run --rm -e APP_COMMAND=provision \
  -e INITIAL_PARENT_USERNAME -e INITIAL_PARENT_PASSWORD \
  -e INITIAL_CHILD_USERNAME -e INITIAL_CHILD_PASSWORD app
```

L'image incorpore les contenus et migrations ; aucune API IA ni fichier externe
n'est nécessaire. Un compte déjà créé fait échouer la commande de provisioning
plutôt que d'en remplacer le mot de passe. Les migrations sont appliquées au
démarrage et par `make migrate`. Sauvegarder la base avant une migration réelle.

En HTTPS : définir `COOKIE_SECURE=true` dans Compose et conserver Host au proxy.
La PWA précache uniquement les assets ; elle n'enregistre pas les réponses ni
les données privées hors ligne. Un exercice et sa soumission nécessitent le réseau.

## Vérifications

```bash
make test
go test -race ./...
make lint
make test-integration
make test-e2e
make build
```

Les nouveaux tests d'intégration créent et nettoient leurs propres schémas. Les
deux tests du prototype restent présents et écrivent des identifiants de test
uniques : fournir une base de tests dédiée, pas une base de production.

E2E exige une famille fraîche et un serveur déjà lancé : une mission terminée
ne peut pas être réinitialisée depuis le client. Exporter les quatre variables
E2E de compte et `E2E_BASE_URL`, puis exécuter `make test-e2e`. Les tests exigent
Chromium Playwright (`cd web && npx playwright install chromium`). Un environnement
sans ces variables affiche explicitement des tests skipped, pas un succès E2E.
