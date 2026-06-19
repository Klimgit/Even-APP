import 'package:flutter/material.dart';

import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/exercise_catalog.dart';
import 'package:online_cource_app/exercises/listen_choice_exercise.dart';
import 'package:online_cource_app/exercises/sentence_builder_exercise.dart';
import 'package:online_cource_app/exercises/word_match_exercise.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Form for creating/editing a single lesson step (exercise). Returns the
/// built [ExerciseData] via `Navigator.pop` when saved.
class StepEditorScreen extends StatefulWidget {
  final String typeId;
  final ExerciseData? initial;

  const StepEditorScreen({super.key, required this.typeId, this.initial});

  @override
  State<StepEditorScreen> createState() => _StepEditorScreenState();
}

class _StepEditorScreenState extends State<StepEditorScreen> {
  // Shared
  final _lang = TextEditingController(text: 'en-US');

  // Listen & choose
  final _lcPrompt = TextEditingController();
  final _lcWord = TextEditingController();
  bool _lcGrid = true;
  final List<TextEditingController> _lcOptions =
      List.generate(4, (_) => TextEditingController());
  int _lcCorrect = 0;
  final _lcAnswer = TextEditingController();

  // Build a sentence
  final _sbPrompt = TextEditingController();
  final _sbSpeak = TextEditingController();
  final _sbSentence = TextEditingController();

  // Match words
  final _wmPrompt = TextEditingController();
  final List<List<TextEditingController>> _wmPairs = [];

  @override
  void initState() {
    super.initState();
    _prefill();
    if (_wmPairs.isEmpty) {
      for (var i = 0; i < 4; i++) {
        _wmPairs.add([TextEditingController(), TextEditingController()]);
      }
    }
  }

  void _prefill() {
    final initial = widget.initial;
    if (initial is ListenChoiceExercise) {
      _lcPrompt.text = initial.prompt;
      _lcWord.text = initial.word;
      _lang.text = initial.languageCode;
      _lcGrid = initial.options != null;
      if (initial.options != null) {
        for (var i = 0; i < 4 && i < initial.options!.length; i++) {
          _lcOptions[i].text = initial.options![i];
        }
        _lcCorrect = initial.options!.indexOf(initial.correctAnswer).clamp(0, 3);
      } else {
        _lcAnswer.text = initial.correctAnswer;
      }
    } else if (initial is SentenceBuilderExercise) {
      _sbPrompt.text = initial.promptText;
      _sbSpeak.text = initial.speakText;
      _lang.text = initial.languageCode;
      _sbSentence.text = initial.correctWords.join(' ');
    } else if (initial is WordMatchExercise) {
      _wmPrompt.text = initial.prompt;
      for (final p in initial.pairs) {
        _wmPairs.add([
          TextEditingController(text: p.known),
          TextEditingController(text: p.target),
        ]);
      }
    }
  }

  @override
  void dispose() {
    for (final c in [
      _lang, _lcPrompt, _lcWord, _lcAnswer, _sbPrompt, _sbSpeak,
      _sbSentence, _wmPrompt, ..._lcOptions,
    ]) {
      c.dispose();
    }
    for (final pair in _wmPairs) {
      for (final c in pair) c.dispose();
    }
    super.dispose();
  }

  String? _validateAndBuild(void Function(ExerciseData) onOk) {
    switch (widget.typeId) {
      case ListenChoiceExercise.typeId:
        if (_lcWord.text.trim().isEmpty) return 'Enter the word to play';
        if (_lcGrid) {
          final opts = _lcOptions.map((c) => c.text.trim()).toList();
          if (opts.any((o) => o.isEmpty)) return 'Fill in all 4 options';
          onOk(ListenChoiceExercise(
            word: _lcWord.text.trim(),
            options: opts,
            correctAnswer: opts[_lcCorrect],
            prompt: _lcPrompt.text.trim().isEmpty
                ? 'Which word did you hear?'
                : _lcPrompt.text.trim(),
            languageCode: _lang.text.trim(),
          ));
        } else {
          if (_lcAnswer.text.trim().isEmpty) return 'Enter the correct answer';
          onOk(ListenChoiceExercise(
            word: _lcWord.text.trim(),
            correctAnswer: _lcAnswer.text.trim(),
            prompt: _lcPrompt.text.trim().isEmpty
                ? 'Type what you hear'
                : _lcPrompt.text.trim(),
            languageCode: _lang.text.trim(),
          ));
        }
        return null;
      case SentenceBuilderExercise.typeId:
        final words = _sbSentence.text.trim().split(RegExp(r'\s+'))
          ..removeWhere((w) => w.isEmpty);
        if (words.length < 2) return 'Enter a sentence (2+ words)';
        onOk(SentenceBuilderExercise(
          promptText: _sbPrompt.text.trim(),
          speakText: _sbSpeak.text.trim().isEmpty
              ? _sbSentence.text.trim()
              : _sbSpeak.text.trim(),
          correctWords: words,
          languageCode: _lang.text.trim(),
        ));
        return null;
      case WordMatchExercise.typeId:
        final pairs = <WordPair>[];
        for (final p in _wmPairs) {
          final known = p[0].text.trim();
          final target = p[1].text.trim();
          if (known.isEmpty && target.isEmpty) continue;
          if (known.isEmpty || target.isEmpty) {
            return 'Fill both sides of each pair';
          }
          pairs.add(WordPair(known: known, target: target));
        }
        if (pairs.length < 2) return 'Add at least 2 complete pairs';
        onOk(WordMatchExercise(
          prompt: _wmPrompt.text.trim().isEmpty
              ? 'Match the pairs'
              : _wmPrompt.text.trim(),
          pairs: pairs,
        ));
        return null;
      default:
        return 'Unknown type';
    }
  }

  void _save() {
    final error = _validateAndBuild((built) => Navigator.of(context).pop(built));
    if (error != null) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(error)));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        backgroundColor: AppTheme.backgroundColor,
        elevation: 0,
        foregroundColor: AppTheme.textColor,
        title: Text(exerciseTypeLabel(widget.typeId)),
        actions: [
          TextButton(
            onPressed: _save,
            child: const Text('Save',
                style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: _buildForm(),
      ),
    );
  }

  List<Widget> _buildForm() {
    switch (widget.typeId) {
      case ListenChoiceExercise.typeId:
        return _listenChoiceForm();
      case SentenceBuilderExercise.typeId:
        return _sentenceForm();
      case WordMatchExercise.typeId:
        return _matchForm();
      default:
        return const [Text('Unknown exercise type')];
    }
  }

  // --- Listen & choose ---
  List<Widget> _listenChoiceForm() {
    return [
      _field(_lcWord, 'Word/phrase to play (spoken aloud)'),
      _field(_lcPrompt, 'Prompt (optional)'),
      _field(_lang, 'Language code (e.g. en-US)'),
      const SizedBox(height: 8),
      Row(
        children: [
          const Text('Answer mode:'),
          const SizedBox(width: 12),
          ChoiceChip(
            label: const Text('2×2 grid'),
            selected: _lcGrid,
            onSelected: (_) => setState(() => _lcGrid = true),
          ),
          const SizedBox(width: 8),
          ChoiceChip(
            label: const Text('Text input'),
            selected: !_lcGrid,
            onSelected: (_) => setState(() => _lcGrid = false),
          ),
        ],
      ),
      const SizedBox(height: 8),
      if (_lcGrid) ...[
        const Text('Options (select the correct one):',
            style: TextStyle(fontWeight: FontWeight.w600)),
        for (var i = 0; i < 4; i++)
          Row(
            children: [
              Radio<int>(
                value: i,
                groupValue: _lcCorrect,
                onChanged: (v) => setState(() => _lcCorrect = v ?? 0),
              ),
              Expanded(child: _field(_lcOptions[i], 'Option ${i + 1}')),
            ],
          ),
      ] else
        _field(_lcAnswer, 'Correct answer'),
    ];
  }

  // --- Build a sentence ---
  List<Widget> _sentenceForm() {
    return [
      _field(_sbPrompt, 'Prompt (e.g. translation shown to the learner)'),
      _field(_sbSentence, 'Correct sentence (words, space-separated)'),
      _field(_sbSpeak, 'Spoken text (optional; defaults to the sentence)'),
      _field(_lang, 'Language code (e.g. en-US)'),
      const SizedBox(height: 8),
      Text(
        'The learner reassembles these words in order.',
        style: TextStyle(color: AppTheme.secondaryTextColor, fontSize: 12),
      ),
    ];
  }

  // --- Match words ---
  List<Widget> _matchForm() {
    return [
      _field(_wmPrompt, 'Prompt (optional)'),
      const SizedBox(height: 8),
      const Text('Pairs (known ↔ target):',
          style: TextStyle(fontWeight: FontWeight.w600)),
      for (var i = 0; i < _wmPairs.length; i++)
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 4),
          child: Row(
            children: [
              Expanded(child: _field(_wmPairs[i][0], 'Known')),
              const SizedBox(width: 8),
              Expanded(child: _field(_wmPairs[i][1], 'Target')),
              IconButton(
                icon: const Icon(Icons.remove_circle_outline),
                onPressed: _wmPairs.length <= 2
                    ? null
                    : () => setState(() {
                          for (final c in _wmPairs[i]) c.dispose();
                          _wmPairs.removeAt(i);
                        }),
              ),
            ],
          ),
        ),
      TextButton.icon(
        onPressed: () => setState(() => _wmPairs
            .add([TextEditingController(), TextEditingController()])),
        icon: const Icon(Icons.add),
        label: const Text('Add pair'),
      ),
    ];
  }

  Widget _field(TextEditingController c, String label) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: TextField(
        controller: c,
        decoration: InputDecoration(
          labelText: label,
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide(color: AppTheme.dividerColor),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide(color: AppTheme.dividerColor),
          ),
        ),
      ),
    );
  }
}
