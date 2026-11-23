# Modèle de domaine

## Socle actif

- Curriculum : `Skill`, `Prerequisite`, `LearningObjective`, `CurriculumModule`,
  `IrregularVerb` et ses métadonnées pédagogiques. Les valeurs sont construites
  avec validation et les collections sont copiées défensivement.
- Learning : `Student`, `LearningSession`, `Exercise`, `StudentAnswer`,
  `Assessment`, `ReviewSchedule`. `SkillProgress` et `PracticeAttempt` sont
  conservés du prototype ; les routes actives utilisent le scheduler V1 par
  compétence et par verbe décrit dans l'ADR 0003.
- Missions : `Mission`, `MissionStep`, `Challenge`, `Reward`, `MissionProgress`.
  La récompense cumulée doit correspondre exactement aux étapes déjà validées.
  Une étape ne peut être récompensée deux fois, et une mauvaise réponse n'avance
  pas. Les missions finies ne peuvent pas être réinitialisées par le client.
- `StudentAnswer` est non vide, limité à 500 caractères. La normalisation ferme
  se limite aux espaces et à la casse. Variantes britanniques/américaines
  explicitement listées, sans correction approximative.
- Les phrases libres ont l'état `needs_evaluation`, visible par le parent.
  La V1 n'expose pas de validation humaine avec attribution d'XP.

Les sections suivantes décrivent les objets du prototype précédent, conservés
mais non utilisés par le parcours de missions actif.

## Noyau partagé

`SkillID` identifie une compétence de façon stable entre contextes, par exemple `english.grammar.past_simple.irregular_verbs`. Il ne transporte aucune règle propre à un contexte.

## Curriculum

`IrregularVerb` contient les formes canoniques et la traduction française. Sa méthode `Check` normalise espaces et casse, puis produit une correction détaillée. Cette décision ne dépend ni d’une interface ni d’un modèle IA.

## Learning

`PracticeAttempt` est la trace immuable d’une réponse. `SkillProgress` agrège les preuves observées pour un élève et une compétence.

La tranche actuelle mesure la **restitution** :

- une première réussite établit la reconnaissance ;
- cinq réponses avec au moins 80 % de réussite établissent la restitution ;
- l’utilisation en contexte ne peut jamais être accordée par cet exercice isolé.

Une réponse incorrecte planifie une révision à quatre heures, une reconnaissance à un jour et une restitution à sept jours. Ces valeurs sont une politique initiale explicite, appelée à évoluer à partir des retours pédagogiques.

## Invariants

- une progression ne peut pas avoir plus de réponses correctes que de réponses totales ;
- une tentative et la progression résultante sont persistées atomiquement ;
- une mise à jour concurrente de progression est détectée par version et rejouée de manière bornée ;
- un LLM ne peut pas attribuer directement un niveau de validation ;
- un exercice de restitution ne peut pas attribuer le niveau `context`.
