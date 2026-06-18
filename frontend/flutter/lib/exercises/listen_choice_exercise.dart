import 'package:flutter/material.dart';
import 'package:online_cource_app/theme/app_theme.dart';
import 'package:online_cource_app/exercises/audio_service.dart';
import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/exercise_common.dart';

/// Exercise type 1: play a word, then pick the right answer from a 2x2 grid
/// of options, or type it in a text field. Checked with the bottom bar.
class ListenChoiceExercise extends ExerciseData {
  /// Word/phrase that is spoken aloud.
  final String word;
  final String languageCode;

  /// Title shown above the answer area.
  final String prompt;

  /// When non-null, options are shown as a 2x2 grid. When null, a text field
  /// is shown instead.
  final List<String>? options;

  /// The correct answer (must match an option in [options] when in grid mode).
  final String correctAnswer;

  const ListenChoiceExercise({
    required this.word,
    required this.correctAnswer,
    this.options,
    this.prompt = 'Tap to listen, then answer',
    this.languageCode = 'en-US',
  });

  @override
  Widget build({required ValueChanged<bool> onResult}) =>
      _ListenChoiceWidget(data: this, onResult: onResult);
}

class _ListenChoiceWidget extends StatefulWidget {
  final ListenChoiceExercise data;
  final ValueChanged<bool> onResult;

  const _ListenChoiceWidget({required this.data, required this.onResult});

  @override
  State<_ListenChoiceWidget> createState() => _ListenChoiceWidgetState();
}

class _ListenChoiceWidgetState extends State<_ListenChoiceWidget> {
  ExerciseStatus _status = ExerciseStatus.answering;
  String? _selected;
  final TextEditingController _controller = TextEditingController();

  bool get _isGrid => widget.data.options != null;

  @override
  void initState() {
    super.initState();
    // Auto-play the word when the exercise appears.
    WidgetsBinding.instance.addPostFrameCallback((_) => _play());
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _play() {
    AudioService.instance
        .speak(widget.data.word, languageCode: widget.data.languageCode);
  }

  bool get _canCheck =>
      _isGrid ? _selected != null : _controller.text.trim().isNotEmpty;

  void _check() {
    final answer = _isGrid ? (_selected ?? '') : _controller.text.trim();
    final correct = answer.toLowerCase() ==
        widget.data.correctAnswer.trim().toLowerCase();
    setState(() {
      _status = correct ? ExerciseStatus.correct : ExerciseStatus.wrong;
    });
  }

  @override
  Widget build(BuildContext context) {
    final answered = _status != ExerciseStatus.answering;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          widget.data.prompt,
          style: Theme.of(context)
              .textTheme
              .titleLarge
              ?.copyWith(fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 24),
        Center(child: AudioPlayButton(onTap: _play)),
        const SizedBox(height: 32),
        Expanded(child: _isGrid ? _buildGrid() : _buildInput()),
        ExerciseBottomBar(
          status: _status,
          canCheck: _canCheck && !answered,
          onCheck: _check,
          onContinue: () =>
              widget.onResult(_status == ExerciseStatus.correct),
        ),
      ],
    );
  }

  Widget _buildGrid() {
    final options = widget.data.options!;
    final answered = _status != ExerciseStatus.answering;
    return GridView.count(
      crossAxisCount: 2,
      childAspectRatio: 2.4,
      crossAxisSpacing: 12,
      mainAxisSpacing: 12,
      children: [
        for (final option in options)
          _OptionTile(
            label: option,
            selected: _selected == option,
            // After answering, highlight the correct option and a wrong pick.
            correct: answered && option == widget.data.correctAnswer,
            wrong: answered &&
                _selected == option &&
                option != widget.data.correctAnswer,
            onTap: answered ? null : () => setState(() => _selected = option),
          ),
      ],
    );
  }

  Widget _buildInput() {
    final answered = _status != ExerciseStatus.answering;
    return TextField(
      controller: _controller,
      enabled: !answered,
      onChanged: (_) => setState(() {}),
      decoration: InputDecoration(
        hintText: 'Type what you hear',
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: AppTheme.dividerColor),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: AppTheme.dividerColor),
        ),
      ),
    );
  }
}

class _OptionTile extends StatelessWidget {
  final String label;
  final bool selected;
  final bool correct;
  final bool wrong;
  final VoidCallback? onTap;

  const _OptionTile({
    required this.label,
    required this.selected,
    required this.correct,
    required this.wrong,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    Color borderColor = AppTheme.dividerColor;
    Color bgColor = Colors.white;
    if (correct) {
      borderColor = AppTheme.accentColor;
      bgColor = AppTheme.accentColor.withOpacity(0.1);
    } else if (wrong) {
      borderColor = AppTheme.errorColor;
      bgColor = AppTheme.errorColor.withOpacity(0.1);
    } else if (selected) {
      borderColor = AppTheme.accentColor;
      bgColor = AppTheme.accentColor.withOpacity(0.08);
    }

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        alignment: Alignment.center,
        padding: const EdgeInsets.symmetric(horizontal: 8),
        decoration: BoxDecoration(
          color: bgColor,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: borderColor,
            width: selected || correct || wrong ? 1.8 : 1,
          ),
        ),
        child: Text(
          label,
          textAlign: TextAlign.center,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            color: AppTheme.textColor,
          ),
        ),
      ),
    );
  }
}
