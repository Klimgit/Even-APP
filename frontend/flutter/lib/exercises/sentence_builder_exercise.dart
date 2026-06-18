import 'package:flutter/material.dart';
import 'package:online_cource_app/theme/app_theme.dart';
import 'package:online_cource_app/exercises/audio_service.dart';
import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/exercise_common.dart';

/// Exercise type 2: a prompt with a listen icon, and the words of the target
/// sentence shuffled below. Tapping words places them in order onto the answer
/// lines; tapping a placed word returns it to the bank. Checked at the bottom.
class SentenceBuilderExercise extends ExerciseData {
  /// Prompt shown in the main area (e.g. the sentence in the known language).
  final String promptText;

  /// Text spoken by the listen icon (the target sentence).
  final String speakText;

  final String languageCode;

  /// The correct sentence as an ordered list of words.
  final List<String> correctWords;

  const SentenceBuilderExercise({
    required this.promptText,
    required this.speakText,
    required this.correctWords,
    this.languageCode = 'en-US',
  });

  @override
  Widget build({required ValueChanged<bool> onResult}) =>
      _SentenceBuilderWidget(data: this, onResult: onResult);
}

class _Token {
  final int id;
  final String text;
  const _Token(this.id, this.text);
}

class _SentenceBuilderWidget extends StatefulWidget {
  final SentenceBuilderExercise data;
  final ValueChanged<bool> onResult;

  const _SentenceBuilderWidget({required this.data, required this.onResult});

  @override
  State<_SentenceBuilderWidget> createState() => _SentenceBuilderWidgetState();
}

class _SentenceBuilderWidgetState extends State<_SentenceBuilderWidget> {
  ExerciseStatus _status = ExerciseStatus.answering;
  late List<_Token> _bank;
  final List<_Token> _answer = [];

  @override
  void initState() {
    super.initState();
    _bank = [
      for (var i = 0; i < widget.data.correctWords.length; i++)
        _Token(i, widget.data.correctWords[i]),
    ]..shuffle();
  }

  void _play() {
    AudioService.instance
        .speak(widget.data.speakText, languageCode: widget.data.languageCode);
  }

  void _place(_Token token) {
    setState(() {
      _bank.remove(token);
      _answer.add(token);
    });
  }

  void _remove(_Token token) {
    setState(() {
      _answer.remove(token);
      _bank.add(token);
    });
  }

  bool get _canCheck => _bank.isEmpty && _answer.isNotEmpty;

  void _check() {
    final assembled = _answer.map((t) => t.text).toList();
    final correct = assembled.length == widget.data.correctWords.length &&
        List.generate(assembled.length,
                (i) => assembled[i] == widget.data.correctWords[i])
            .every((ok) => ok);
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
        // Prompt + listen icon.
        Row(
          children: [
            IconButton(
              onPressed: _play,
              icon: Icon(Icons.volume_up_rounded,
                  color: AppTheme.accentColor, size: 28),
            ),
            const SizedBox(width: 4),
            Expanded(
              child: Text(
                widget.data.promptText,
                style: Theme.of(context)
                    .textTheme
                    .titleLarge
                    ?.copyWith(fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        const SizedBox(height: 24),
        // Answer lines.
        Container(
          width: double.infinity,
          constraints: const BoxConstraints(minHeight: 64),
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(color: AppTheme.dividerColor),
              top: BorderSide(color: AppTheme.dividerColor),
            ),
          ),
          child: Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final token in _answer)
                _WordChip(
                  label: token.text,
                  onTap: answered ? null : () => _remove(token),
                ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        // Word bank.
        Expanded(
          child: SingleChildScrollView(
            child: Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final token in _bank)
                  _WordChip(
                    label: token.text,
                    onTap: answered ? null : () => _place(token),
                  ),
              ],
            ),
          ),
        ),
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
}

class _WordChip extends StatelessWidget {
  final String label;
  final VoidCallback? onTap;

  const _WordChip({required this.label, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppTheme.dividerColor),
        ),
        child: Text(
          label,
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
