# Refonte interface — 13 octobre 2026

## Changements

- Nouvelle identité visuelle crème/vert profond, navigation latérale sur ordinateur et navigation compacte sur mobile, icônes SVG originales et écran de connexion repensé.
- Tableau de bord avec statistiques, cartes de missions/progression et séance du jour. Historiques, cours, gestion familiale et réglages IA ont leurs propres espaces : ils ne s'empilent plus sur l'accueil.
- Missions : défi et contexte à gauche, Alex à droite ; formulaire texte, lecture, indices et conversation micro clairement séparés. Sur petit écran, les colonnes s'empilent.
- Réglages IA : contrôle parent correctement dimensionné, statut lisible et consommation séparée de l'import. Les champs texte et les checkboxes ont désormais des styles distincts.
- Formulaire d'ajout d'enfant repliable ; interface de cours avec états vides utiles et plan pédagogique structuré.
- Sélecteurs clavier/focus visibles, lien d'évitement, animations désactivables et respect reduced-motion. Pas de dépendance à une police distante ou à des images sous licence.

## Vérifications observées

- 10 tests frontend réussis ; TypeScript/Vite/PWA et build Docker réussis.
- ESLint réussi et 62 tests Go réussis, dont contrôle des frontières d'import.
- 6 nouveaux tests de présentation réussis aux largeurs 390, 768 et 1440 px : accueil enfant, mission, réglages parent, cours et gestion familiale sans débordement horizontal. Largeur/hauteur de checkbox vérifiées inférieures à 30 px.
- Deux contrats navigateur existants (ajout/sélection d'enfant, lecture anglaise locale) adaptés à la nouvelle navigation puis réussis : 8 tests de présentation/contrats réussis ensemble.
- Deux E2E avec véritable serveur Go/PostgreSQL réussis sur famille de régression dédiée : entraînement, mission Lost Diamond avec 65 XP, reconnexion, hors ligne/synchronisation et parcours parent. Aucun appel API LLM réel pendant ces tests.
- Captures des vues ordinateur/mobile examinées ; connexion parent et ouverture des réglages également vérifiées sur la démo réelle en lecture seule.

Un test de contrat familial cherchait encore le formulaire d'ajout directement sur l'accueil. Il a été corrigé pour ouvrir « Ma famille » et déplier le formulaire ; le comportement existant n'a pas été supprimé.

## Publication locale

Nouvelle image publiée sur **http://localhost:18080**. Comptes, progression, consentement et environnement serveur (y compris les clés/providers) conservés : comparaison des variables avant/après réussie, aucune clé affichée. L'autre instance sur 8080 n'a pas été modifiée.

La fixture `englishcraft-ui-regression` utilise `compose.e2e.yaml` avec l'overlay `compose.ui-tests.yaml` (ports localhost 18081/55434) : aucun reset de la famille de démonstration. Les scénarios de présentation emploient des réponses HTTP explicitement simulées ; ils ne prétendent pas vérifier un provider IA.

Cette vérification responsive utilise Chromium. Elle ne valide pas un iPhone physique ou la conversation vocale cloud. Si une ancienne interface PWA reste affichée, fermer/réouvrir l'onglet ou forcer le rechargement ; ne pas effacer les données locales si des réponses attendent leur synchronisation.
