# Accueil anglais, Alex parodique et plafond — 14 octobre 2026

## Diagnostic du blocage

Lecture du ledger de l'élève de démonstration : deux réservations vocales à
0,40 USD (600 s réservées) et vingt réservations texte à 0,01 USD, soit 1 USD
réservé pour un plafond élève de 1 USD. Une troisième voix demandait 0,40 USD
supplémentaire ; le contrôle de durée permettait encore 600 + 300 ≤ 900 s.
La boucle reproduisant le contrôle d'admission échouait sur le budget élève,
pas sur le quota de durée. Ce montant n'était pas une mesure de facture réelle.

Après **accord explicite de l'utilisateur**, plafond élève porté à 2 USD/jour.
Plafonds fournisseur et durée inchangés. Aucune réservation, donnée pédagogique,
session de connexion ou clé n'a été effacée/remplacée. Après publication :
1 000 000 + 400 000 ≤ 2 000 000 micros, et 600 + 300 ≤ 900 s : contrôles réussis
sans ouvrir de session payante supplémentaire. Le message de limite distingue
désormais réservation et facture et précise le renouvellement à 00:00 UTC.

## Accueil et personnage

- Cause du mélange : le welcome vocal préfixait directement le dialogue du
  contenu scolaire, majoritairement français ou bilingue. Test initial en échec
  reproduisant cette lecture.
- Script serveur désormais entièrement anglais, propre à chaque univers :
  situation/objectifs, rôle, règles (`help`, `repeat`, réponses courtes,
  une question à la fois, interruptions), puis une seule question d'ouverture.
- Même script dans l'aperçu local authentifié `/intro` et le greeting WebRTC.
  Aperçu sans LLM, budget ou micro ; fallback local également en anglais et
  indication « sans IA » séparée de la phrase d'Alex dans l'interface.
- Instructions Mistral : anglais par défaut, français seulement sur demande
  explicite ; ne pas lire le contenu bilingue comme une introduction.
- Portrait SVG original de parodie Minecraft Alex : cheveux orange et mèche
  latérale, yeux verts, tenue verte, volumes cubiques, animations conservées et
  respect de la réduction des mouvements. Aucun skin/image/texture officiel.

## Vérifications observées

- `go test ./...` : 73 tests réussis ; `go vet ./...` réussi.
- `go test -race ./internal/tutoring/... ./test/architecture/...` : 37 tests
  réussis, frontières d'import conservées.
- PostgreSQL : 7 tests d'intégration réussis sur schémas isolés du port 55433.
- Frontend : 16 tests réussis, TypeScript/Vite/PWA et ESLint réussis.
- Build Docker réussi ; avertissement inchangé sur le chunk SDK ElevenLabs différé.
- E2E UI sur la démo avec HTTP simulé : 10 tests réussis, deux scénarios métier
  ignorés volontairement faute de famille E2E isolée pour cette passe. Aucun
  compte réel/progression réelle modifié par ces contrats UI.
- Captures à 390, 768 et 1440 px relues : portrait en blocs et accueil anglais
  avec paragraphes histoire/règles/question, bulle sans collision de position.
- API `/intro` **réelle**, authentifiée avec le compte élève de démonstration :
  histoire diamant + règles anglaises + une question, sans dialogue français.
  Cookie de vérification révoqué par logout ; aucun appel LLM/micro lancé.

La première passe frontend réutilisait un même objet `Response` pour deux GET/POST
après ajout de l'aperçu. Mocks corrigés pour fournir une réponse fraîche par
requête ; les deux échecs de test ne reproduisaient pas un comportement serveur.

## Publication et limites

Nouvelle image publiée sur **http://127.0.0.1:18080/** ; environnement serveur,
comptes, mots de passe, consentements, historique et progression conservés, sauf
la hausse autorisée du plafond. Le tunnel WSS local reste nécessaire.
Le test vocal réel avec micro silencieux a validé précédemment le démarrage
d'Alex ; pour cette passe, le nouveau contenu a été vérifié via tests/aperçu,
sans consommer la dernière réservation vocale disponible de l'élève réel.
Qualité sonore/interruptions sur appareil physique restent à tester.
