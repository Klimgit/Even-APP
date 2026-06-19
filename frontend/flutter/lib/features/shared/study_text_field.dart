import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/features/shared/alphabet_controller.dart';
import 'package:online_cource_app/features/shared/study_language_keyboard.dart';

/// Text field with optional on-screen alphabet keyboard for a study language.
class StudyTextField extends StatefulWidget {
  final TextEditingController controller;
  final String languageCode;
  final String hintText;
  final int maxLines;
  final ValueChanged<String>? onChanged;

  const StudyTextField({
    super.key,
    required this.controller,
    required this.languageCode,
    this.hintText = 'Введите текст',
    this.maxLines = 1,
    this.onChanged,
  });

  @override
  State<StudyTextField> createState() => _StudyTextFieldState();
}

class _StudyTextFieldState extends State<StudyTextField> {
  List<AlphabetLetterDto> _letters = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    widget.controller.addListener(_onTextChanged);
    _loadAlphabet();
  }

  @override
  void didUpdateWidget(covariant StudyTextField oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.languageCode != widget.languageCode) {
      _loadAlphabet();
    }
  }

  @override
  void dispose() {
    widget.controller.removeListener(_onTextChanged);
    super.dispose();
  }

  void _onTextChanged() {
    if (mounted) setState(() {});
  }

  Future<void> _loadAlphabet() async {
    setState(() => _loading = true);
    if (!Get.isRegistered<AlphabetController>()) {
      Get.put(AlphabetController(), permanent: true);
    }
    final letters = await Get.find<AlphabetController>().lettersFor(widget.languageCode);
    if (mounted) {
      setState(() {
        _letters = letters;
        _loading = false;
      });
    }
  }

  bool get _useKeyboard => _letters.isNotEmpty;

  void _insert(String ch) {
    final text = widget.controller.text;
    final sel = widget.controller.selection;
    final start = sel.start >= 0 ? sel.start : text.length;
    final end = sel.end >= 0 ? sel.end : text.length;
    final next = text.replaceRange(start, end, ch);
    widget.controller.text = next;
    widget.controller.selection = TextSelection.collapsed(offset: start + ch.length);
    widget.onChanged?.call(next);
    setState(() {});
  }

  void _backspace() {
    final text = widget.controller.text;
    final sel = widget.controller.selection;
    if (text.isEmpty) return;
    if (sel.start != sel.end) {
      _insert('');
      return;
    }
    final pos = sel.start >= 0 ? sel.start : text.length;
    if (pos == 0) return;
    final next = text.replaceRange(pos - 1, pos, '');
    widget.controller.text = next;
    widget.controller.selection = TextSelection.collapsed(offset: pos - 1);
    widget.onChanged?.call(next);
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final keyboardTitle = widget.languageCode == 'evn'
        ? 'Эвенская клавиатура'
        : 'Клавиатура языка';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: widget.controller,
          readOnly: _useKeyboard,
          showCursor: true,
          maxLines: widget.maxLines,
          decoration: InputDecoration(
            hintText: _useKeyboard ? 'Наберите ${widget.hintText.toLowerCase()} на клавиатуре ниже' : widget.hintText,
            suffixIcon: widget.controller.text.isNotEmpty
                ? IconButton(
                    icon: const Icon(Icons.clear),
                    onPressed: () {
                      widget.controller.clear();
                      widget.onChanged?.call('');
                      setState(() {});
                    },
                  )
                : null,
          ),
          onChanged: widget.onChanged,
        ),
        if (_loading)
          const Padding(
            padding: EdgeInsets.only(top: 8),
            child: LinearProgressIndicator(minHeight: 2),
          ),
        if (_useKeyboard) ...[
          const SizedBox(height: 12),
          StudyLanguageKeyboard(
            title: keyboardTitle,
            letters: _letters,
            onCharacter: _insert,
            onBackspace: _backspace,
            onSpace: () => _insert(' '),
          ),
        ],
      ],
    );
  }
}
