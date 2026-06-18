import 'package:flutter/material.dart';
import 'package:online_cource_app/theme/app_theme.dart';
import 'package:online_cource_app/exercises/exercise.dart';

/// Plays an ordered list of [ExerciseData] as a single lesson: a progress bar,
/// one exercise at a time, and a summary at the end. This is how exercises are
/// composed into lessons/courses.
class LessonRunner extends StatefulWidget {
  final String title;
  final List<ExerciseData> exercises;

  const LessonRunner({
    super.key,
    required this.title,
    required this.exercises,
  });

  @override
  State<LessonRunner> createState() => _LessonRunnerState();
}

class _LessonRunnerState extends State<LessonRunner> {
  int _index = 0;
  int _correct = 0;
  bool _finished = false;

  void _onResult(bool correct) {
    if (correct) _correct++;
    if (_index + 1 >= widget.exercises.length) {
      setState(() => _finished = true);
    } else {
      setState(() => _index++);
    }
  }

  @override
  Widget build(BuildContext context) {
    final progress = widget.exercises.isEmpty
        ? 0.0
        : (_finished ? 1.0 : _index / widget.exercises.length);

    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.close, color: AppTheme.textColor),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
        title: ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: LinearProgressIndicator(
            value: progress,
            minHeight: 10,
            backgroundColor: AppTheme.dividerColor,
            valueColor:
                const AlwaysStoppedAnimation<Color>(AppTheme.accentColor),
          ),
        ),
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: _finished
              ? _buildSummary()
              : KeyedSubtree(
                  // Fresh state for each exercise.
                  key: ValueKey(_index),
                  child: widget.exercises[_index].build(onResult: _onResult),
                ),
        ),
      ),
    );
  }

  Widget _buildSummary() {
    final total = widget.exercises.length;
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.emoji_events_rounded,
              size: 72, color: AppTheme.accentColor),
          const SizedBox(height: 16),
          Text(
            'Lesson complete!',
            style: Theme.of(context)
                .textTheme
                .displaySmall
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Text(
            '$_correct / $total correct',
            style: TextStyle(
              fontSize: 18,
              color: AppTheme.secondaryTextColor,
            ),
          ),
          const SizedBox(height: 32),
          SizedBox(
            width: 200,
            height: 50,
            child: ElevatedButton(
              onPressed: () => Navigator.of(context).maybePop(),
              style: ElevatedButton.styleFrom(
                backgroundColor: AppTheme.accentColor,
                foregroundColor: Colors.white,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
              child: const Text(
                'Done',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
