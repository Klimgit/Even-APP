import 'package:flutter/material.dart';

/// Common contract for every exercise type.
///
/// A lesson is just an ordered list of [ExerciseData]. The lesson runner asks
/// each one to build its widget and listens for the result via [onResult],
/// which fires `true` for a correct answer and `false` otherwise. This keeps
/// exercise types interchangeable and composable into lessons/courses.
abstract class ExerciseData {
  const ExerciseData();

  /// Builds the interactive widget for this exercise.
  ///
  /// The widget owns its "Check" button and reports the outcome through
  /// [onResult] exactly once, after the learner checks their answer.
  Widget build({required ValueChanged<bool> onResult});
}
