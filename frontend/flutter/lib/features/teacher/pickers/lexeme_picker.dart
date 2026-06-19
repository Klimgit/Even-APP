import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Searchable lexeme picker (teacher + platform scope via API).
Future<LexemeDto?> showLexemePicker({
  required BuildContext context,
  required String languageCode,
  bool platformScope = false,
}) async {
  return showModalBottomSheet<LexemeDto>(
    context: context,
    isScrollControlled: true,
    builder: (ctx) => _LexemePickerSheet(
      languageCode: languageCode,
      platformScope: platformScope,
    ),
  );
}

Future<List<LexemeDto>> showLexemeMultiPicker({
  required BuildContext context,
  required String languageCode,
  List<String> initialIds = const [],
}) async {
  final result = await showModalBottomSheet<List<LexemeDto>>(
    context: context,
    isScrollControlled: true,
    builder: (ctx) => _LexemePickerSheet(
      languageCode: languageCode,
      multiSelect: true,
      initialIds: initialIds,
    ),
  );
  return result ?? [];
}

class _LexemePickerSheet extends StatefulWidget {
  final String languageCode;
  final bool platformScope;
  final bool multiSelect;
  final List<String> initialIds;

  const _LexemePickerSheet({
    required this.languageCode,
    this.platformScope = false,
    this.multiSelect = false,
    this.initialIds = const [],
  });

  @override
  State<_LexemePickerSheet> createState() => _LexemePickerSheetState();
}

class _LexemePickerSheetState extends State<_LexemePickerSheet> {
  final _lexicon = Get.find<LexiconRepository>();
  final _search = TextEditingController();
  List<LexemeDto> _items = [];
  final Set<String> _selected = {};
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _selected.addAll(widget.initialIds);
    _load();
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final resp = widget.platformScope
          ? await _lexicon.listPlatformLexicon(
              languageCode: widget.languageCode,
              query: _search.text.trim(),
            )
          : await _lexicon.listTeacherLexicon(
              languageCode: widget.languageCode,
              query: _search.text.trim(),
            );
      setState(() => _items = resp.items);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return DraggableScrollableSheet(
      expand: false,
      initialChildSize: 0.7,
      maxChildSize: 0.95,
      builder: (ctx, scrollController) => Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: TextField(
              controller: _search,
              decoration: const InputDecoration(
                hintText: 'Поиск слова…',
                prefixIcon: Icon(Icons.search),
              ),
              onSubmitted: (_) => _load(),
            ),
          ),
          if (widget.multiSelect)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                children: [
                  Text('Выбрано: ${_selected.length}'),
                  const Spacer(),
                  ElevatedButton(
                    onPressed: () {
                      final picked = _items.where((l) => _selected.contains(l.id)).toList();
                      Navigator.pop(context, picked);
                    },
                    child: const Text('Готово'),
                  ),
                ],
              ),
            ),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : ListView.builder(
                    controller: scrollController,
                    itemCount: _items.length,
                    itemBuilder: (ctx, i) {
                      final lex = _items[i];
                      final trans = lex.translations.map((t) => t.text).join(', ');
                      if (widget.multiSelect) {
                        return CheckboxListTile(
                          value: _selected.contains(lex.id),
                          title: Text(lex.lemma),
                          subtitle: Text(trans.isEmpty ? '—' : trans),
                          onChanged: (v) {
                            setState(() {
                              if (v == true) {
                                _selected.add(lex.id);
                              } else {
                                _selected.remove(lex.id);
                              }
                            });
                          },
                        );
                      }
                      return ListTile(
                        title: Text(lex.lemma, style: const TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: Text(trans.isEmpty ? '—' : trans),
                        onTap: () => Navigator.pop(context, lex),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

/// Compact chip showing selected lexeme with pick action.
class LexemePickerField extends StatelessWidget {
  final String? lexemeId;
  final String? lemma;
  final String label;
  final VoidCallback onPick;

  const LexemePickerField({
    super.key,
    required this.lexemeId,
    required this.lemma,
    required this.label,
    required this.onPick,
  });

  @override
  Widget build(BuildContext context) {
    return InputDecorator(
      decoration: InputDecoration(labelText: label),
      child: InkWell(
        onTap: onPick,
        child: Row(
          children: [
            Expanded(
              child: Text(
                lemma ??
                    (lexemeId != null && lexemeId!.isNotEmpty && !_looksLikeUuid(lexemeId!)
                        ? lexemeId!
                        : (lexemeId != null && lexemeId!.isNotEmpty ? 'Слово' : 'Выбрать…')),
                style: TextStyle(
                  color: lemma != null || (lexemeId != null && lexemeId!.isNotEmpty)
                      ? AppTheme.textColor
                      : AppTheme.secondaryTextColor,
                ),
              ),
            ),
            const Icon(Icons.search, size: 20),
          ],
        ),
      ),
    );
  }
}

bool _looksLikeUuid(String value) {
  return RegExp(
    r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$',
  ).hasMatch(value);
}
