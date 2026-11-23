# Vérification manuelle voix et PWA

**Non exécutée sur appareils physiques.** Les tests jsdom et Playwright ne valident pas un micro Safari réel. Avant activation pour l'enfant, vérifier Chrome macOS/Android, Safari macOS, Safari iOS et PWA installée iPad/iPhone, via HTTPS.

- Installation : icônes PNG, ajout à l'écran d'accueil iOS, lancement autonome, portrait/paysage, clavier et aucune largeur débordante.
- Aucun prompt microphone avant clic volontaire. Refus, permission précédemment refusée, retrait dans les réglages : message compréhensible et texte utilisable.
- Lecture audio : Safari peut refuser `play()` après le handshake réseau ; bouton Lecture natif présent. Tester haut-parleur, casque Bluetooth et changement de sortie.
- Écho/bruit : vérifier la capture et l'intelligibilité ; la négociation codec appartient à WebRTC, ne pas imposer PCM navigateur.
- Indicateurs écoute/parole, sous-titres lorsque disponibles, interruption d'Alex, mute, démute, arrêt : vérifier extinction de l'indicateur micro OS.
- Push-to-talk : pression tactile/relâchement/cancel, touche Espace, capture de pointeur, réponse précise courte et longue.
- Réseau : mode avion, changement Wi-Fi/4G, fermeture du data channel, délais ICE. Plus d'audio après arrêt ; fallback texte et défis locaux.
- Lifecycle : quitter la mission, déconnexion, onglet arrière-plan, verrouillage écran, retour après suspension iOS. Un navigateur suspendu peut dépasser le timer local : confirmer l'arrêt serveur et le quota réservé.
- Comptes : aucune clé API dans le navigateur, cookie HttpOnly, impossible d'activer IA depuis l'enfant, parent peut désactiver les sessions actives.
- Outils : bonne réponse orale réellement retranscrite, correction `incorrect` puis retry, XP idempotents serveur, étape suivante et mission complète, progression sur un second appareil.
- Hors ligne : télécharger en ligne, relancer sans réseau, consulter verbes, quiz provisoires, reconnecter le **même** profil, synchroniser deux fois sans double preuve ni récompense. Ne jamais synchroniser la file d'un autre profil.
- Animation : préférences reduced-motion et checkbox Animations ; toutes les actions restent lisibles sans mouvement.
- Budget/durée : attendre la durée maximale réelle, tester quota quotidien atteint et fournisseur indisponible. Les plafonds fournisseur sont indispensables si le VPS tombe.

La voix n'est pas « validée » avant un rapport daté de ces tests avec navigateur/version/appareil et clés de test autorisées. Aucun audio n'est conservé par EnglishCraft ; vérifier séparément les politiques de rétention du fournisseur.
