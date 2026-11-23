# Socle V1 : contenus, transactions, identité et révisions

## Structure conservée

Le module Go reste à la racine, React dans `web/`, les bounded contexts dans
`internal/<métier>`. Déplacer le travail existant dans `backend/` et `apps/web/`
n'améliorerait pas l'encapsulation. Le contexte Learning contient les cas
d'utilisation orchestrant Curriculum et Missions ; les domaines ne s'importent
pas mutuellement. Les ports sont définis dans les packages consommateurs.

`content/` contient uniquement des contenus JSON versionnés et leur embed Go.
L'adapter `curriculum/adapters/content` les valide à chaque démarrage, reconstruit
les entités et expose les vues nécessaires. L'application reste hors réseau.
La version d'une mission fait partie de sa progression. Une future modification
incompatible doit conserver l'ancienne version ou migrer explicitement ses
progressions : ne jamais remplacer silencieusement les défis d'une mission active.

## Transaction par élève

Le périmètre familial V1 utilise une ligne JSONB `student_states` par profil.
Elle contient les snapshots de missions/révisions, les tentatives et les reçus.
`SELECT FOR UPDATE` sérialise les soumissions de cet élève, puis le cas
d'utilisation applique les règles Go. Un seul commit écrit toutes les conséquences.
Une réponse HTTP perdue peut être rejouée : même clé et même contenu rendent le
reçu original. Une clé réutilisée avec une réponse différente est refusée.
Les XP sont dérivés des étapes validées, jamais d'une valeur client.

Ce choix simple n'est pas un event store. Les identités et sessions sont
relationnelles, avec unicité et clés étrangères. La V1 familiale garde le journal
et les reçus sans pagination de stockage ; avant un usage multi-familles, extraire
tentatives/reçus en tables indexées et prévoir leur rétention. L'historique HTTP
est limité aux vingt dernières tentatives. SQL ne corrige aucune réponse.

## Algorithme de révision

- Nouvelle notion : découverte.
- Réussite à échéance : série +1, intervalle de 1, 3, 7, 14 puis 30 jours.
- Trois réussites espacées : maîtrise ; jamais via un simple spam de réponses.
- Échec : série remise à zéro, apprentissage, prochaine échéance dans 4 heures.
- Réussite avant échéance : tentative enregistrée, sans allonger l'intervalle.
- Entraînement : erreurs arrivées à échéance en priorité, puis verbes nouveaux.
- Aucune notion disponible : pas de répétition automatique excessive.
- Phrase libre : `needs_evaluation`, sans XP, sans révision ou maîtrise accordée.

L'horloge est un port du cas d'utilisation ; le domaine reçoit un `time.Time`.
FSRS pourra remplacer la règle de `ReviewSchedule.Record` derrière les mêmes
opérations métier, avec une migration des snapshots et des tests de simulation.
Il n'y a aucune dépendance à un LLM. Les phrases à remettre en ordre sont fermées
et distinctes des phrases libres : une phrase grammaticalement différente mais
valide n'est pas prétendue validée par le moteur.

## Identité familiale

Une commande administrative crée un parent et son enfant, sans inscription HTTP.
Mots de passe distincts (12–72 octets), empreintes bcrypt coût 12, session aléatoire
de 256 bits, seule l'empreinte SHA-256 de la session est stockée. Durée 12 heures,
révocation au logout, cookie HttpOnly/SameSite=Strict/Secure en HTTPS. Aucun jeton
n'est accessible via localStorage/sessionStorage.

Cinq échecs verrouillent le compte pendant quinze minutes, avec mise à jour
transactionnelle du compteur. Les mutations contrôlent Origin/Sec-Fetch-Site.
Un proxy doit préserver Host. La V1 n'a ni MFA, ni récupération de mot de passe,
ni gestion de plusieurs enfants ; un déploiement Internet doit ajouter une
limitation de débit globale au proxy et des sauvegardes. La création initiale
refuse de remplacer un compte déjà provisionné.

Les adapters du prototype précédent restent conservés mais leurs endpoints et
jetons statiques ne sont plus câblés dans `bootstrap`. Le contrat actif est
`api/openapi.yaml` ; `docs/legacy-practice.openapi.yaml` et `web/api/openapi.yaml`
ne décrivent pas les routes actives.
