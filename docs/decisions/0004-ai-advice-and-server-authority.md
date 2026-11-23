# IA consultative, progression serveur et voix WebRTC

Le contexte Tutoring consomme les cas d'utilisation Learning, sans déplacer les corrections ni les transactions. Les providers renvoient des conseils non fiables ; les évaluations ouvertes restent soumises à une validation humaine. Les outils ne reçoivent aucune identité élève : elle vient des sessions HttpOnly. `complete_mission_step` lit seulement une progression déjà validée par `Submit`.

Le flux Realtime unifié garde la clé sur Go et transporte l'audio directement par WebRTC. Le navigateur transmet les propositions d'outils à l'API authentifiée ; ni ces propositions ni les événements fournisseur ne créent une récompense. Une sideband pourra remplacer cette transmission, mais ne remplace pas l'autorisation et la correction Go.

Les réservations quotidiennes sont conservatrices et persistantes, y compris après erreur ambiguë ou arrêt du processus. Elles ne sont pas des factures : prix et plafond fournisseur doivent être configurés séparément. L'import texte est d'abord local et ne conserve pas le document original ; publier une fiche exige la validation parentale. Les sessions IA sont désactivées par défaut et exigent un consentement parental.
