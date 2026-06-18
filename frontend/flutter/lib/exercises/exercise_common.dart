import 'package:flutter/material.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Lifecycle of a single exercise attempt.
enum ExerciseStatus { answering, correct, wrong }

/// Shared bottom bar for every exercise: a full-width "Check" button while the
/// learner is answering, then a result banner + "Continue" once checked.
///
/// Keeps the look and interaction identical across all exercise types.
class ExerciseBottomBar extends StatelessWidget {
  final ExerciseStatus status;

  /// Whether enough has been entered/selected to allow checking.
  final bool canCheck;
  final VoidCallback onCheck;
  final VoidCallback onContinue;

  const ExerciseBottomBar({
    super.key,
    required this.status,
    required this.canCheck,
    required this.onCheck,
    required this.onContinue,
  });

  @override
  Widget build(BuildContext context) {
    final bool answered = status != ExerciseStatus.answering;
    final bool correct = status == ExerciseStatus.correct;
    final Color accent = correct ? AppTheme.accentColor : AppTheme.errorColor;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (answered) ...[
          Row(
            children: [
              Icon(
                correct ? Icons.check_circle : Icons.cancel,
                color: accent,
              ),
              const SizedBox(width: 8),
              Text(
                correct ? 'Correct!' : 'Not quite',
                style: TextStyle(
                  color: accent,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
        ],
        SizedBox(
          height: 50,
          child: ElevatedButton(
            onPressed: answered
                ? onContinue
                : (canCheck ? onCheck : null),
            style: ElevatedButton.styleFrom(
              backgroundColor: answered ? accent : AppTheme.accentColor,
              foregroundColor: Colors.white,
              disabledBackgroundColor: AppTheme.dividerColor,
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: Text(
              answered ? 'Continue' : 'Check',
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ),
      ],
    );
  }
}

/// Round "play audio" button shared by listening exercises.
class AudioPlayButton extends StatelessWidget {
  final VoidCallback onTap;
  final double size;

  const AudioPlayButton({super.key, required this.onTap, this.size = 64});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(size),
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          color: AppTheme.accentColor,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Icon(Icons.volume_up_rounded, color: Colors.white, size: size * 0.5),
      ),
    );
  }
}
