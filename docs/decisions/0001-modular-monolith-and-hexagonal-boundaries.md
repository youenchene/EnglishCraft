# Monolithe modulaire avec frontières hexagonales

EnglishCraft est déployé comme un monolithe modulaire plutôt que comme plusieurs services. Le domaine reste isolé par des frontières d’import testées, tandis qu’un seul processus et une seule image réduisent fortement la complexité opérationnelle sur Dokploy ; une extraction future reste possible si une contrainte mesurée la justifie.
