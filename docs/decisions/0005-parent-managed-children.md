# Plusieurs enfants gérés par le parent

La création initiale du parent reste administrative, sans inscription publique.
Un parent authentifié peut lister et créer ses enfants depuis son tableau de bord.
Chaque compte utilise un identifiant unique, une empreinte bcrypt et un état
pédagogique vide, créés dans la même transaction. Aucun mot de passe n'est renvoyé.

Les routes parent acceptent un paramètre `child` ; le serveur vérifie toujours
que ce profil appartient au parent de la session. Sans paramètre, un enfant est
choisi de manière déterministe pour préserver les clients existants. Le frontend
sélectionne explicitement le profil pour les résultats, plans, consentements IA,
imports et suppressions. Le choix reste en mémoire, sans secret stocké en JavaScript.

Cette décision remplace la limitation à un enfant de l'ADR 0003. Le schéma
relationnel existant autorise déjà plusieurs enfants, sans migration.

La lecture « Écouter Alex » utilise la synthèse vocale anglaise du navigateur,
sur clic explicite, indépendamment du micro et des fournisseurs IA. Elle n'est
pas une conversation : la voix interactive cloud conserve ses exigences de
configuration, consentement parental, Internet et budget. Le navigateur peut
utiliser une voix locale ou distante selon le système ; aucune garantie hors ligne.
