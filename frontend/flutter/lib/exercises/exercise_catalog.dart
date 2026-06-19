import 'package:flutter/material.dart';

import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/listen_choice_exercise.dart';
import 'package:online_cource_app/exercises/material_exercise.dart';
import 'package:online_cource_app/exercises/sentence_builder_exercise.dart';
import 'package:online_cource_app/exercises/word_match_exercise.dart';

/// Rebuilds an [ExerciseData] from its stored JSON (see each type's `toJson`).
ExerciseData exerciseFromJson(Map<String, dynamic> json) {
  switch (json['type'] as String?) {
    case ListenChoiceExercise.typeId:
      return ListenChoiceExercise.fromJson(json);
    case SentenceBuilderExercise.typeId:
      return SentenceBuilderExercise.fromJson(json);
    case WordMatchExercise.typeId:
      return WordMatchExercise.fromJson(json);
    case MaterialExercise.typeId:
      return MaterialExercise.fromJson(json);
    default:
      throw ArgumentError('Unknown exercise type: ${json['type']}');
  }
}

/// Catalog entry describing an exercise type for the teacher's lesson builder.
class ExerciseType {
  final String id;
  final String label;
  final String description;
  final IconData icon;

  const ExerciseType({
    required this.id,
    required this.label,
    required this.description,
    required this.icon,
  });
}

/// The exercise types a teacher can add as lesson steps.
const List<ExerciseType> exerciseTypes = [
  ExerciseType(
    id: ListenChoiceExercise.typeId,
    label: 'Listen & choose',
    description: 'Play a word, pick from a grid or type the answer',
    icon: Icons.hearing_rounded,
  ),
  ExerciseType(
    id: SentenceBuilderExercise.typeId,
    label: 'Build a sentence',
    description: 'Reorder shuffled words into the correct sentence',
    icon: Icons.reorder_rounded,
  ),
  ExerciseType(
    id: WordMatchExercise.typeId,
    label: 'Match words',
    description: 'Match words in two columns by meaning',
    icon: Icons.compare_arrows_rounded,
  ),
  ExerciseType(
    id: MaterialExercise.typeId,
    label: 'Study material',
    description: 'Text, images, audio and video to read/watch',
    icon: Icons.article_rounded,
  ),
];

/// Human-readable label for a stored exercise (by its `type`).
String exerciseTypeLabel(String typeId) {
  return exerciseTypes
      .firstWhere(
        (t) => t.id == typeId,
        orElse: () => const ExerciseType(
          id: '',
          label: 'Exercise',
          description: '',
          icon: Icons.extension_rounded,
        ),
      )
      .label;
}
