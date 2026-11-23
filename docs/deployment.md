# Déploiement

Image unique Go/React, utilisateur non privilégié, sans GPU. Usage familial : 1–2 vCPU, 2 Go RAM VPS (application limitée à 256 Mo dans l'exemple), PostgreSQL 17, disque sauvegardé. Secrets et limites : [environment.md](environment.md).

## Docker Compose

Le fichier `compose.yaml` démarre PostgreSQL et l’application. Suivre [local-core.md](local-core.md) pour les variables et le provisioning. Les migrations embarquées sont appliquées au démarrage avant l’ouverture du serveur HTTP.

## Dokploy et Traefik

1. Créer une base PostgreSQL privée dans le même réseau Dokploy.
2. Construire le dépôt avec le `Dockerfile` racine.
3. Définir `DATABASE_URL` comme secret et `COOKIE_SECURE=true` ; provisionner les comptes via la commande administrative, pas via des jetons statiques.
4. Exposer uniquement le port applicatif `8080` via Traefik avec HTTPS.
5. Configurer la sonde de vie `GET /healthz`, succès `204`, et readiness `GET /readyz` qui vérifie PostgreSQL. La sonde Docker appelle `/englishcraft health` sur le port 8080.

Exemple Compose : [`deploy/dokploy.compose.yaml`](../deploy/dokploy.compose.yaml). Utiliser un projet Dokploy Compose avec chemin `deploy/dokploy.compose.yaml` ; vérifier le contexte build relatif `..`. La base n'a **aucun port publié** et appartient seulement au réseau `private` interne. L'application rejoint aussi le réseau Traefik `dokploy-network` pour HTTPS et les API sortantes.

### Première installation et migrations contrôlées

1. Définir `POSTGRES_PASSWORD` fort (URL-encoder dans `DATABASE_URL` si nécessaire) et les autres secrets dans Dokploy. Ne pas les inclure dans une image ou dans le frontend.
2. Construire l'image et démarrer PostgreSQL. Avec `AUTO_MIGRATE=false`, exécuter une tâche one-shot de **la nouvelle image**, même réseau et `DATABASE_URL`, `APP_COMMAND=migrate`, avant de lancer l'app.
3. Exécuter ensuite une tâche one-shot `APP_COMMAND=provision` avec les quatre variables `INITIAL_*`. Elle refuse de remplacer une famille existante. Retirer ces mots de passe des variables de l'application après provisionnement.
4. Déployer l'application, vérifier readiness puis le login enfant/parent. Ne pas lancer les tests E2E sur une famille de production : ils modifient la progression.
5. Laisser `AI_PROVIDER=disabled`, `VOICE_ENABLED=false` et les budgets à zéro jusqu'à configuration des comptes fournisseurs et validation du parcours réel. Le parent doit autoriser l'IA même après configuration serveur.

Équivalent CLI dans un clone contrôlé (variables secrètes fournies par l'environnement) :

```bash
docker compose -f deploy/dokploy.compose.yaml build app
docker compose -f deploy/dokploy.compose.yaml up -d --wait postgres
docker compose -f deploy/dokploy.compose.yaml run --rm -e APP_COMMAND=migrate app
docker compose -f deploy/dokploy.compose.yaml run --rm -e APP_COMMAND=provision \
  -e INITIAL_PARENT_USERNAME -e INITIAL_PARENT_PASSWORD \
  -e INITIAL_CHILD_USERNAME -e INITIAL_CHILD_PASSWORD app
docker compose -f deploy/dokploy.compose.yaml up -d --wait app
```

Le réseau externe `dokploy-network` est géré par Dokploy ; ne pas exécuter cet exemple tel quel hors Dokploy sans créer/configurer le réseau.

### Traefik / HTTPS

Dans Domains, sélectionner `app`, port 8080, HTTPS avec Let's Encrypt et redirection HTTP→HTTPS. Préserver `Host`, ne pas réécrire l'Origin. Ne pas ajouter de CORS permissif. Le backend utilise SameSite Strict et contrôle Origin/Sec-Fetch-Site ; `COOKIE_SECURE=true` est obligatoire hors localhost. Seuls 80/443 doivent être exposés publiquement. WebRTC audio sort directement vers OpenAI ; Traefik ne relaie pas l'audio. Pas de WebSocket applicatif à configurer pour cette version.

Limiter le débit au proxy (particulièrement login, tutor, imports), par exemple middleware Traefik rateLimit `average=30`, `burst=60`. Le backend ajoute une limite globale familiale de 120 mutations/minute et des plafonds persistants par élève ; ce n'est pas une protection distribuée multi-tenant.

Ne pas publier le port PostgreSQL. La terminaison TLS appartient à Traefik ; l’application ne doit recevoir aucune clé privée TLS.
Conserver le Host d'origine pour le contrôle CSRF. Configurer une limitation de
débit au proxy, en particulier sur `/api/v1/auth/login`, avant toute exposition publique.

## Sauvegarde et retour arrière

Sauvegarder PostgreSQL avant chaque déploiement contenant une migration. Les migrations sont ascendantes et atomiques ; un retour applicatif doit rester compatible avec le schéma déployé ou être accompagné d’une migration corrective testée.

### Sauvegarde

```bash
docker compose -f deploy/dokploy.compose.yaml exec -T postgres \
  pg_dump -U englishcraft -d englishcraft -Fc > englishcraft-backup.dump
```

Chiffrer le dump, le copier hors VPS, conserver une rétention familiale courte et tester périodiquement la restauration. Il contient mots de passe hachés, sessions, réponses et progression : limiter les accès. Une suppression de profil n'efface pas les sauvegardes ; prévoir leur expiration et ne pas restaurer des données supprimées sans traitement.

### Restauration

Dans une base **neuve et dédiée**, application arrêtée, après vérification du dump :

```bash
docker compose -f deploy/dokploy.compose.yaml exec -T postgres \
  createdb -U englishcraft englishcraft_restore
docker compose -f deploy/dokploy.compose.yaml exec -T postgres \
  pg_restore -U englishcraft -d englishcraft_restore --no-owner --exit-on-error < englishcraft-backup.dump
```

Pointer une instance de vérification sur `englishcraft_restore`, tester readiness/login/progression, puis basculer `DATABASE_URL`. Ne pas utiliser `--clean` sur une base active. Révoquer les sessions restaurées si nécessaire. Conserver l'ancienne base pour un retour contrôlé plutôt que la détruire immédiatement.

### Mise à jour

Sauvegarder → désactiver temporairement voix/IA → arrêter l'ancienne application (prévenir l'enfant) → construire la nouvelle image → tâche `APP_COMMAND=migrate` → démarrer → readiness → login, mission, progression parent → réactiver après validation. Garder l'image précédente. Sur SIGTERM le serveur dispose de 10 secondes pour terminer les requêtes ; les échéances vocales restent persistées et seront reprises au redémarrage. Un arrêt prolongé requiert un contrôle des appels côté fournisseur.

Le build local ne prouve ni le routage Traefik ni la disponibilité réelle Dokploy. La checklist [mobile/voix](mobile-voice-checklist.md) fait partie du go-live.
