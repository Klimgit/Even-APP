import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/content_dto.dart';
import 'package:online_cource_app/api/repositories/auth_repository.dart';
import 'package:online_cource_app/api/repositories/content_repository.dart';
import 'package:online_cource_app/features/teacher/courses/block_editor_sheet.dart';
import 'package:online_cource_app/theme/app_theme.dart';

class LessonEditorScreen extends StatefulWidget {
  final String lessonId;
  final String title;
  final String languageCode;
  final String languageId;

  const LessonEditorScreen({
    super.key,
    required this.lessonId,
    required this.title,
    required this.languageCode,
    required this.languageId,
  });

  @override
  State<LessonEditorScreen> createState() => _LessonEditorScreenState();
}

class _LessonEditorScreenState extends State<LessonEditorScreen> {
  final _content = Get.find<ContentRepository>();
  LessonDto? _lesson;
  List<BlockTypeInfoDto> _blockTypes = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      _lesson = await _content.getLesson(widget.lessonId);
      _blockTypes = await _content.listBlockTypes();
    } catch (e) {
      _error = e.toString();
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _showError(Object e) {
    final message = e is ApiException ? e.message : e.toString();
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Ошибка: $message')),
    );
  }

  Future<void> _addSection() async {
    try {
      await _content.createSection(lessonId: widget.lessonId, title: 'Секция');
      await _load();
    } catch (e) {
      _showError(e);
    }
  }

  int _nextSortOrder(LessonSectionDto section) {
    if (section.blocks.isEmpty) return 0;
    return section.blocks.map((b) => b.sortOrder).reduce((a, b) => a > b ? a : b) + 1;
  }

  Future<void> _addBlock(LessonSectionDto section) async {
    final type = await showModalBottomSheet<BlockTypeInfoDto>(
      context: context,
      builder: (ctx) => ListView(
        children: [
          for (final t in _blockTypes)
            ListTile(
              title: Text(t.title),
              subtitle: Text(t.blockType),
              trailing: t.isGradable ? const Icon(Icons.quiz, size: 18) : null,
              onTap: () => Navigator.pop(ctx, t),
            ),
        ],
      ),
    );
    if (type == null) return;

    var config = _defaultConfig(type.blockType);
    var title = type.title;

    if (type.isGradable || _needsEditorBeforeCreate(type.blockType)) {
      final draft = LessonBlockDto(
        id: '',
        sectionId: section.id,
        sortOrder: _nextSortOrder(section),
        title: title,
        blockType: type.blockType,
        config: config,
        isHomework: false,
        isGradable: type.isGradable,
      );
      final edited = await BlockEditorSheet.show(
        context,
        block: draft,
        languageCode: widget.languageCode,
        languageId: widget.languageId,
      );
      if (edited == null) return;
      edited.remove('_media_name');
      config = edited;
      title = draft.title ?? type.title;
    }

    try {
      await _content.createBlock(
        lessonId: widget.lessonId,
        sectionId: section.id,
        blockType: type.blockType,
        config: config,
        sortOrder: _nextSortOrder(section),
        title: title,
      );
      await _load();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Задание создано')),
        );
      }
    } catch (e) {
      _showError(e);
    }
  }

  bool _needsEditorBeforeCreate(String blockType) {
    return blockType == 'vocabulary_set' || blockType == 'video' || blockType == 'audio';
  }

  Map<String, dynamic> _defaultConfig(String blockType) {
    switch (blockType) {
      case 'text':
        return {
          'items': [
            {'kind': 'text', 'text': 'Текст урока'}
          ]
        };
      case 'note':
        return {'body': 'Заметка'};
      case 'vocabulary_set':
        return {'lexeme_ids': [], 'show_images': true, 'show_audio': true};
      case 'video':
        return {'media_asset_id': '', 'caption': ''};
      case 'audio':
        return {
          'items': [
            {'kind': 'audio', 'media_asset_id': ''}
          ]
        };
      case 'images_stacked':
        return {
          'items': [
            {'kind': 'text', 'text': 'Подпись'}
          ]
        };
      case 'prompt_choose_word':
      case 'listen_choose_word':
        return {
          'prompt': {
            'items': [
              {'kind': 'text', 'text': 'Вопрос'}
            ]
          },
          'choices': [
            {'text': 'A'},
            {'text': 'B'}
          ],
          'correct_index': 0,
        };
      case 'word_choose_image':
        return {
          'prompt': {
            'items': [
              {'kind': 'text', 'text': 'Выберите изображение'}
            ]
          },
          'choices': [
            {'media_asset_id': ''},
            {'media_asset_id': ''},
          ],
          'correct_index': 0,
        };
      case 'word_choose_translation':
        return {
          'prompt': {
            'items': [
              {'kind': 'text', 'text': 'Выберите перевод'}
            ]
          },
          'choices': [
            {'text': 'A'},
            {'text': 'B'}
          ],
          'correct_index': 0,
        };
      case 'prompt_type_word':
      case 'listen_type_word':
        return {
          'prompt': {
            'items': [
              {'kind': 'text', 'text': 'Введите слово'}
            ]
          },
          'correct_lexeme_id': '',
          'correct_text': '',
        };
      case 'prompt_sentence_type':
      case 'listen_sentence_type':
        return {
          'prompt': {
            'items': [
              {'kind': 'text', 'text': 'Введите предложение'}
            ]
          },
          'correct_text': ' ',
        };
      case 'gap_sentence_choose_word':
        return {
          'body': 'Это {{gap:0}} предложение.',
          'gaps': [
            {'id': 0, 'correct_lexeme_id': ''}
          ],
          'choices': [''],
        };
      case 'prompt_sentence_word_order':
      case 'listen_sentence_word_order':
        return {
          'sub_items': [
            {
              'prompt': {
                'items': [
                  {'kind': 'text', 'text': 'Соберите предложение'}
                ]
              },
              'tokens': ['слово1', 'слово2'],
              'correct_order': [0, 1],
            }
          ],
        };
      default:
        return {};
    }
  }

  Future<void> _editBlock(LessonBlockDto block) async {
    final config = await BlockEditorSheet.show(
      context,
      block: block,
      languageCode: widget.languageCode,
      languageId: widget.languageId,
    );
    if (config == null) return;
    config.remove('_media_name');
    try {
      await _content.patchBlock(blockId: block.id, config: config);
      await _load();
    } catch (e) {
      _showError(e);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        actions: [
          if (_lesson?.status != 'published')
            TextButton(
              onPressed: () async {
                try {
                  await _content.publishLesson(widget.lessonId);
                  await _load();
                } catch (e) {
                  _showError(e);
                }
              },
              child: const Text('Опубликовать'),
            ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!))
              : ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    OutlinedButton.icon(
                      onPressed: _addSection,
                      icon: const Icon(Icons.add),
                      label: const Text('Добавить секцию'),
                    ),
                    for (final section in _lesson?.sections ?? []) ...[
                      const SizedBox(height: 16),
                      Row(
                        children: [
                          Text(section.title, style: Theme.of(context).textTheme.titleLarge),
                          const Spacer(),
                          Text(section.sectionKind,
                              style: TextStyle(color: AppTheme.secondaryTextColor)),
                          IconButton(
                            icon: const Icon(Icons.add_box_outlined),
                            tooltip: 'Добавить задание',
                            onPressed: () => _addBlock(section),
                          ),
                        ],
                      ),
                      for (final block in section.blocks)
                        Card(
                          child: ListTile(
                            leading: Icon(
                              block.isGradable ? Icons.quiz : Icons.article_outlined,
                            ),
                            title: Text(block.title ?? block.blockType),
                            subtitle: Text(block.displayLabel ?? block.blockType),
                            trailing: IconButton(
                              icon: const Icon(Icons.delete_outline),
                              onPressed: () async {
                                try {
                                  await _content.deleteBlock(block.id);
                                  await _load();
                                } catch (e) {
                                  _showError(e);
                                }
                              },
                            ),
                            onTap: () => _editBlock(block),
                          ),
                        ),
                    ],
                  ],
                ),
    );
  }
}
