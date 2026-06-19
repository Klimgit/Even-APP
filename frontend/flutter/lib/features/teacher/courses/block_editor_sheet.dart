import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/content_dto.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';
import 'package:online_cource_app/features/shared/lexeme_utils.dart';
import 'package:online_cource_app/features/shared/study_text_field.dart';
import 'package:online_cource_app/features/teacher/pickers/lexeme_picker.dart';
import 'package:online_cource_app/features/teacher/pickers/media_picker.dart';

/// Visual block config editor (replaces raw JSON dialog).
class BlockEditorSheet extends StatefulWidget {
  final LessonBlockDto block;
  final String languageCode;
  final String languageId;

  const BlockEditorSheet({
    super.key,
    required this.block,
    required this.languageCode,
    required this.languageId,
  });

  static Future<Map<String, dynamic>?> show(
    BuildContext context, {
    required LessonBlockDto block,
    required String languageCode,
    required String languageId,
  }) {
    return showModalBottomSheet<Map<String, dynamic>>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(bottom: MediaQuery.of(ctx).viewInsets.bottom),
        child: BlockEditorSheet(
          block: block,
          languageCode: languageCode,
          languageId: languageId,
        ),
      ),
    );
  }

  @override
  State<BlockEditorSheet> createState() => _BlockEditorSheetState();
}

class _BlockEditorSheetState extends State<BlockEditorSheet> {
  late Map<String, dynamic> _config;
  final _titleCtrl = TextEditingController();
  final _correctTextCtrl = TextEditingController();
  final Map<String, String> _lemmaCache = {};
  bool _loadingLexemes = true;

  @override
  void initState() {
    super.initState();
    _config = Map<String, dynamic>.from(widget.block.config);
    _titleCtrl.text = widget.block.title ?? '';
    _correctTextCtrl.text = _config['correct_text']?.toString() ?? '';
    _preloadLexemes();
  }

  Future<void> _preloadLexemes() async {
    final ids = collectLexemeIdsFromConfig(_config);
    if (ids.isNotEmpty) {
      await preloadLemmaCache(
        lexicon: Get.find<LexiconRepository>(),
        languageCode: widget.languageCode,
        ids: ids,
        cache: _lemmaCache,
      );
    }
    if (mounted) setState(() => _loadingLexemes = false);
  }

  @override
  void dispose() {
    _titleCtrl.dispose();
    _correctTextCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return DraggableScrollableSheet(
      expand: false,
      initialChildSize: 0.85,
      maxChildSize: 0.95,
      builder: (ctx, scrollController) => Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    'Редактор: ${widget.block.blockType}',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                ),
                TextButton(onPressed: () => Navigator.pop(context), child: const Text('Отмена')),
                ElevatedButton(
                  onPressed: () => Navigator.pop(context, _config),
                  child: const Text('Сохранить'),
                ),
              ],
            ),
          ),
          Expanded(
            child: ListView(
              controller: scrollController,
              padding: const EdgeInsets.all(16),
              children: [
                TextField(
                  controller: _titleCtrl,
                  decoration: const InputDecoration(labelText: 'Заголовок блока'),
                ),
                const SizedBox(height: 16),
                if (_loadingLexemes)
                  const LinearProgressIndicator(minHeight: 2)
                else
                  ..._fieldsForType(widget.block.blockType),
              ],
            ),
          ),
        ],
      ),
    );
  }

  List<Widget> _fieldsForType(String type) {
    switch (type) {
      case 'text':
        return [_textItemsEditor()];
      case 'note':
        return [_noteEditor()];
      case 'vocabulary_set':
        return [_vocabularySetEditor()];
      case 'video':
        return [_videoEditor()];
      case 'audio':
        return [_audioEditor()];
      case 'images_stacked':
        return [_stackedEditor()];
      case 'prompt_choose_word':
      case 'listen_choose_word':
      case 'word_choose_translation':
        return [_chooseEditor(useLexeme: type == 'prompt_choose_word' || type == 'listen_choose_word')];
      case 'word_choose_image':
        return [_chooseImageEditor()];
      case 'prompt_type_word':
      case 'listen_type_word':
      case 'prompt_sentence_type':
      case 'listen_sentence_type':
        return [_typeAnswerEditor()];
      case 'gap_sentence_choose_word':
        return [_gapEditor()];
      case 'prompt_sentence_word_order':
      case 'listen_sentence_word_order':
        return [_wordOrderEditor()];
      default:
        return [
          TextField(
            maxLines: 8,
            decoration: const InputDecoration(labelText: 'Config (JSON вручную — нестандартный тип)'),
            controller: TextEditingController(text: _config.toString()),
            onChanged: (_) {},
          ),
        ];
    }
  }

  Widget _textItemsEditor() {
    final items = (_config['items'] as List<dynamic>? ?? [])
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList();
    if (items.isEmpty) items.add({'kind': 'text', 'text': ''});

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Текстовые блоки', style: TextStyle(fontWeight: FontWeight.w600)),
        for (var i = 0; i < items.length; i++)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: TextFormField(
              initialValue: items[i]['text']?.toString() ?? '',
              decoration: InputDecoration(labelText: 'Абзац ${i + 1}'),
              maxLines: 3,
              onChanged: (v) {
                items[i]['kind'] = 'text';
                items[i]['text'] = v;
                _config['items'] = items;
              },
            ),
          ),
        TextButton.icon(
          onPressed: () => setState(() {
            items.add({'kind': 'text', 'text': ''});
            _config['items'] = items;
          }),
          icon: const Icon(Icons.add),
          label: const Text('Добавить абзац'),
        ),
      ],
    );
  }

  Widget _noteEditor() {
    return TextFormField(
      initialValue: _config['body']?.toString() ?? '',
      decoration: const InputDecoration(labelText: 'Текст заметки'),
      maxLines: 5,
      onChanged: (v) => _config['body'] = v,
    );
  }

  Widget _vocabularySetEditor() {
    final ids = (_config['lexeme_ids'] as List<dynamic>? ?? []).map((e) => e.toString()).toList();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text('Слова (${ids.length})'),
            const Spacer(),
            ElevatedButton.icon(
              onPressed: () async {
                final picked = await showLexemeMultiPicker(
                  context: context,
                  languageCode: widget.languageCode,
                  initialIds: ids,
                );
                setState(() {
                  _config['lexeme_ids'] = picked.map((l) => l.id).toList();
                  final labels = <String, String>{};
                  for (final l in picked) {
                    _lemmaCache[l.id] = l.lemma;
                    labels[l.id] = l.lemma;
                  }
                  _config['lexeme_labels'] = labels;
                });
              },
              icon: const Icon(Icons.menu_book),
              label: const Text('Выбрать слова'),
            ),
          ],
        ),
        SwitchListTile(
          title: const Text('Показывать картинки'),
          value: _config['show_images'] as bool? ?? true,
          onChanged: (v) => setState(() => _config['show_images'] = v),
        ),
        SwitchListTile(
          title: const Text('Показывать аудио'),
          value: _config['show_audio'] as bool? ?? true,
          onChanged: (v) => setState(() => _config['show_audio'] = v),
        ),
      ],
    );
  }

  Widget _videoEditor() {
    final mediaId = _config['media_asset_id'] as String? ?? '';
    return Column(
      children: [
        MediaPickerField(
          mediaId: mediaId,
          displayName: _config['_media_name'] as String?,
          label: 'Видео',
          onPick: () async {
            final m = await showMediaPicker(
              context: context,
              languageCode: widget.languageCode,
              languageId: widget.languageId,
              kind: 'video',
            );
            if (m != null) {
              setState(() {
                _config['media_asset_id'] = m.id;
                _config['_media_name'] = m.displayName;
              });
            }
          },
        ),
        TextFormField(
          initialValue: _config['caption']?.toString() ?? '',
          decoration: const InputDecoration(labelText: 'Подпись'),
          onChanged: (v) => _config['caption'] = v,
        ),
      ],
    );
  }

  Widget _audioEditor() {
    return ElevatedButton.icon(
      onPressed: () async {
        final m = await showMediaPicker(
          context: context,
          languageCode: widget.languageCode,
          languageId: widget.languageId,
          kind: 'audio',
        );
        if (m != null) {
          setState(() {
            _config['items'] = [
              {'kind': 'audio', 'media_asset_id': m.id},
            ];
          });
        }
      },
      icon: const Icon(Icons.audiotrack),
      label: const Text('Выбрать аудио'),
    );
  }

  Widget _stackedEditor() {
    return TextFormField(
      initialValue: (_config['items'] as List?)?.firstWhere(
            (e) => (e as Map)['kind'] == 'text',
            orElse: () => {'text': ''},
          )['text']?.toString() ??
          '',
      decoration: const InputDecoration(labelText: 'Подпись к изображению'),
      onChanged: (v) => setState(() {
        _config['items'] = [
          {'kind': 'text', 'text': v},
        ];
      }),
    );
  }

  Widget _promptField() {
    final prompt = _config['prompt'] as Map<String, dynamic>? ?? {'items': []};
    final items = prompt['items'] as List<dynamic>? ?? [];
    final text = items.isNotEmpty ? (items.first as Map)['text']?.toString() ?? '' : '';
    return TextFormField(
      initialValue: text,
      decoration: const InputDecoration(labelText: 'Вопрос / задание'),
      maxLines: 2,
      onChanged: (v) {
        _config['prompt'] = {
          'items': [{'kind': 'text', 'text': v}],
        };
      },
    );
  }

  Widget _chooseEditor({required bool useLexeme}) {
    final choices = (_config['choices'] as List<dynamic>? ?? [])
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList();
    while (choices.length < 2) {
      choices.add(useLexeme ? {'lexeme_id': ''} : {'text': ''});
    }
    var correctIndex = _config['correct_index'] as int? ?? 0;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _promptField(),
        const SizedBox(height: 12),
        const Text('Варианты ответа', style: TextStyle(fontWeight: FontWeight.w600)),
        for (var i = 0; i < choices.length; i++)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Row(
              children: [
                Radio<int>(
                  value: i,
                  groupValue: correctIndex,
                  onChanged: (v) => setState(() {
                    correctIndex = v ?? 0;
                    _config['correct_index'] = correctIndex;
                  }),
                ),
                Expanded(
                  child: useLexeme
                      ? LexemePickerField(
                          lexemeId: choices[i]['lexeme_id'] as String?,
                          lemma: _lemmaCache[choices[i]['lexeme_id'] as String?] ??
                              choices[i]['lemma'] as String?,
                          label: 'Вариант ${i + 1}',
                          onPick: () async {
                            final lex = await showLexemePicker(
                              context: context,
                              languageCode: widget.languageCode,
                            );
                            if (lex != null) {
                              setState(() {
                                choices[i]['lexeme_id'] = lex.id;
                                choices[i]['lemma'] = lex.lemma;
                                _lemmaCache[lex.id] = lex.lemma;
                                _config['choices'] = choices;
                              });
                            }
                          },
                        )
                      : TextFormField(
                          initialValue: choices[i]['text']?.toString() ?? '',
                          decoration: InputDecoration(labelText: 'Вариант ${i + 1}'),
                          onChanged: (v) {
                            choices[i]['text'] = v;
                            _config['choices'] = choices;
                          },
                        ),
                ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _chooseImageEditor() {
    final choices = (_config['choices'] as List<dynamic>? ?? [])
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList();
    while (choices.length < 2) choices.add({'media_asset_id': ''});
    var correctIndex = _config['correct_index'] as int? ?? 0;

    return Column(
      children: [
        _promptField(),
        for (var i = 0; i < choices.length; i++)
          ListTile(
            leading: Radio<int>(
              value: i,
              groupValue: correctIndex,
              onChanged: (v) => setState(() => _config['correct_index'] = v ?? 0),
            ),
            title: MediaPickerField(
              mediaId: choices[i]['media_asset_id'] as String?,
              displayName: null,
              label: 'Изображение ${i + 1}',
              onPick: () async {
                final m = await showMediaPicker(
                  context: context,
                  languageCode: widget.languageCode,
                  languageId: widget.languageId,
                  kind: 'image',
                );
                if (m != null) {
                  setState(() {
                    choices[i]['media_asset_id'] = m.id;
                    _config['choices'] = choices;
                  });
                }
              },
            ),
          ),
      ],
    );
  }

  Widget _typeAnswerEditor() {
    final isWordType =
        widget.block.blockType == 'prompt_type_word' ||
        widget.block.blockType == 'listen_type_word';
    final lexId = _config['correct_lexeme_id'] as String?;
    return Column(
      children: [
        _promptField(),
        if (isWordType)
          LexemePickerField(
            lexemeId: lexId != null && lexId.isNotEmpty ? lexId : null,
            lemma: lexId != null ? _lemmaCache[lexId] : null,
            label: 'Правильное слово',
            onPick: () async {
              final lex = await showLexemePicker(
                context: context,
                languageCode: widget.languageCode,
              );
              if (lex != null) {
                setState(() {
                  _config['correct_lexeme_id'] = lex.id;
                  _config['correct_text'] = lex.lemma;
                  _correctTextCtrl.text = lex.lemma;
                  _lemmaCache[lex.id] = lex.lemma;
                });
              }
            },
          ),
        StudyTextField(
          controller: _correctTextCtrl,
          languageCode: widget.languageCode,
          hintText: 'правильный ответ',
          onChanged: (v) => _config['correct_text'] = v,
        ),
      ],
    );
  }

  Widget _gapEditor() {
    final gaps = (_config['gaps'] as List<dynamic>? ?? []);
    final gapLexId = gaps.isNotEmpty ? (gaps.first as Map)['correct_lexeme_id'] as String? : null;
    return Column(
      children: [
        TextFormField(
          initialValue: _config['body']?.toString() ?? '',
          decoration: const InputDecoration(
            labelText: 'Предложение',
            hintText: 'Используйте {{gap:0}} для пропуска',
          ),
          onChanged: (v) => _config['body'] = v,
        ),
        LexemePickerField(
          lexemeId: gapLexId,
          lemma: gapLexId != null ? _lemmaCache[gapLexId] : null,
          label: 'Правильное слово',
          onPick: () async {
            final lex = await showLexemePicker(
              context: context,
              languageCode: widget.languageCode,
            );
            if (lex != null) {
              setState(() {
                _config['gaps'] = [{'id': 0, 'correct_lexeme_id': lex.id}];
                _lemmaCache[lex.id] = lex.lemma;
              });
            }
          },
        ),
      ],
    );
  }

  Widget _wordOrderEditor() {
    final subItems = _config['sub_items'] as List<dynamic>? ?? [];
    final sub = subItems.isNotEmpty
        ? Map<String, dynamic>.from(subItems.first as Map)
        : {'tokens': <String>[], 'correct_order': <int>[]};
    final tokensCtrl = TextEditingController(
      text: (sub['tokens'] as List<dynamic>? ?? []).join(', '),
    );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _promptField(),
        TextField(
          controller: tokensCtrl,
          decoration: const InputDecoration(
            labelText: 'Слова (через запятую)',
            hintText: 'слово1, слово2, слово3',
          ),
          onChanged: (v) {
            final tokens = v.split(',').map((s) => s.trim()).where((s) => s.isNotEmpty).toList();
            sub['tokens'] = tokens;
            sub['correct_order'] = List.generate(tokens.length, (i) => i);
            _config['sub_items'] = [sub];
          },
        ),
      ],
    );
  }
}
