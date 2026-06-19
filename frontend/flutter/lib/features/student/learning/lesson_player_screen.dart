import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/learning_dto.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/api/repositories/learning_repository.dart';
import 'package:online_cource_app/features/blocks/block_views.dart';
import 'package:online_cource_app/features/shared/alphabet_controller.dart';
import 'package:online_cource_app/features/shared/lexeme_utils.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Stepik-style lesson player: one block at a time with progress bar.
class LessonPlayerScreen extends StatefulWidget {
  final String lessonId;
  final String title;
  final String courseId;
  final String targetLanguageCode;

  const LessonPlayerScreen({
    super.key,
    required this.lessonId,
    required this.title,
    required this.courseId,
    this.targetLanguageCode = 'evn',
  });

  @override
  State<LessonPlayerScreen> createState() => _LessonPlayerScreenState();
}

class _LessonPlayerScreenState extends State<LessonPlayerScreen> {
  final _learning = Get.find<LearningRepository>();
  LessonFlowDto? _flow;
  StudentLessonDto? _lesson;
  List<AlphabetLetterDto> _alphabet = [];
  int _step = 0;
  bool _loading = true;
  bool _submitting = false;
  bool? _lastCorrect;
  DateTime? _stepStarted;

  @override
  void initState() {
    super.initState();
    if (!Get.isRegistered<AlphabetController>()) {
      Get.put(AlphabetController(), permanent: true);
    }
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      _flow = await _learning.getLessonFlow(widget.lessonId);
      _lesson = await _learning.getLesson(widget.lessonId);
      _alphabet = await Get.find<AlphabetController>().lettersFor(widget.targetLanguageCode);
      _stepStarted = DateTime.now();
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  List<FlowBlockDto> get _blocks {
    return (_flow?.items ?? [])
        .where((i) => i.kind == 'lesson_block' && i.block != null)
        .map((i) => i.block!)
        .toList();
  }

  int get _timeSpent {
    if (_stepStarted == null) return 0;
    return DateTime.now().difference(_stepStarted!).inSeconds;
  }

  Future<void> _submit(Map<String, dynamic> response) async {
    final block = _blocks[_step];
    setState(() {
      _submitting = true;
      _lastCorrect = null;
    });
    try {
      final result = await _learning.submitBlockAttempt(
        blockId: block.id,
        response: response,
        timeSpentSeconds: _timeSpent,
      );
      setState(() => _lastCorrect = result.isCorrect);
    } finally {
      setState(() => _submitting = false);
    }
  }

  void _next() {
    if (_step < _blocks.length - 1) {
      setState(() {
        _step++;
        _lastCorrect = null;
        _stepStarted = DateTime.now();
      });
    } else {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return Scaffold(
        appBar: AppBar(title: Text(widget.title)),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    final blocks = _blocks;
    if (blocks.isEmpty) {
      return Scaffold(
        appBar: AppBar(title: Text(widget.title)),
        body: const Center(child: Text('Урок пуст')),
      );
    }

    final block = blocks[_step];
    final progress = (_step + 1) / blocks.length;
    final lexemes = buildLexemeDisplayMap(
      resolved: _lesson?.resolvedLexemes ?? {},
      blockConfigs: blocks.map((b) => b.config),
    );
    final media = _lesson?.resolvedMedia ?? {};
    final needsKeyboard = block.isGradable &&
        (block.blockType.contains('type_word') || block.blockType.contains('sentence_type'));

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(4),
          child: LinearProgressIndicator(value: progress),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (block.displayLabel != null)
              Text(block.displayLabel!, style: TextStyle(color: AppTheme.secondaryTextColor)),
            if (block.title != null)
              Text(block.title!, style: Theme.of(context).textTheme.headlineMedium),
            if (needsKeyboard && _alphabet.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                'Используйте клавиатуру языка для ввода ответа',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppTheme.accentColor,
                      fontWeight: FontWeight.w500,
                    ),
              ),
            ],
            const SizedBox(height: 24),
            if (block.isGradable)
              BlockExerciseView(
                key: ValueKey(block.id),
                block: block,
                lexemes: lexemes,
                media: media,
                alphabetLetters: _alphabet,
                targetLanguageCode: widget.targetLanguageCode,
                onSubmit: _submit,
                submitting: _submitting,
                lastCorrect: _lastCorrect,
              )
            else
              BlockContentView(block: block, lexemes: lexemes, media: media),
            const SizedBox(height: 32),
            Row(
              children: [
                Text('Шаг ${_step + 1} из ${blocks.length}'),
                const Spacer(),
                if (!block.isGradable || _lastCorrect == true)
                  ElevatedButton(
                    onPressed: _next,
                    child: Text(_step < blocks.length - 1 ? 'Далее' : 'Завершить'),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
