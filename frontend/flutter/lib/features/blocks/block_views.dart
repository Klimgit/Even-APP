import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:online_cource_app/api/models/learning_dto.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/features/shared/lexeme_utils.dart';
import 'package:online_cource_app/features/shared/study_language_keyboard.dart';
import 'package:online_cource_app/theme/app_theme.dart';
import 'package:video_player/video_player.dart';

/// Renders lesson block content (non-gradable) for the student player.
class BlockContentView extends StatelessWidget {
  final FlowBlockDto block;
  final Map<String, ResolvedLexemeDto> lexemes;
  final Map<String, ResolvedMediaDto> media;

  const BlockContentView({
    super.key,
    required this.block,
    required this.lexemes,
    required this.media,
  });

  @override
  Widget build(BuildContext context) {
    switch (block.blockType) {
      case 'text':
        return _textBlock();
      case 'images_stacked':
        return _itemsBlock();
      case 'audio':
        return _itemsBlock(audioOnly: true);
      case 'video':
        return _videoBlock();
      case 'vocabulary_set':
        return _vocabularySet();
      case 'note':
        return _noteBlock();
      default:
        return Text('Контент: ${block.blockType}');
    }
  }

  Widget _textBlock() {
    final items = block.config['items'] as List<dynamic>? ?? [];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (final raw in items)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              (raw as Map)['text']?.toString() ?? '',
              style: const TextStyle(fontSize: 18, height: 1.5),
            ),
          ),
      ],
    );
  }

  Widget _itemsBlock({bool audioOnly = false}) {
    final items = block.config['items'] as List<dynamic>? ?? [];
    return Column(
      children: [
        for (final raw in items) _promptItem(raw as Map<String, dynamic>),
      ],
    );
  }

  Widget _videoBlock() {
    final mediaId = block.config['media_asset_id'] as String?;
    final url = mediaId != null ? media[mediaId]?.url : null;
    if (url == null) return const Text('Видео недоступно');
    return _VideoPlayer(url: url);
  }

  Widget _vocabularySet() {
    final ids = (block.config['lexeme_ids'] as List<dynamic>? ?? [])
        .map((e) => e.toString())
        .toList();
    final labels = block.config['lexeme_labels'] as Map<String, dynamic>? ?? {};
    return Wrap(
      spacing: 12,
      runSpacing: 12,
      children: [
        for (final id in ids)
          Chip(
            avatar: block.config['show_images'] == true
                ? CircleAvatar(
                    backgroundImage: CachedNetworkImageProvider(
                      media.values
                              .where((m) => m.mediaKind == 'image')
                              .map((m) => m.url)
                              .firstOrNull ??
                          '',
                    ),
                  )
                : null,
            label: Text(
              lexemeDisplayLabel(
                lexemeId: id,
                lexemes: lexemes,
                embeddedLemma: labels[id]?.toString(),
              ),
            ),
          ),
      ],
    );
  }

  Widget _noteBlock() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.primaryColor.withOpacity(0.08),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(block.config['body']?.toString() ?? ''),
    );
  }

  Widget _promptItem(Map<String, dynamic> item) {
    final kind = item['kind'] as String? ?? 'text';
    switch (kind) {
      case 'text':
        return Text(item['text']?.toString() ?? '', style: const TextStyle(fontSize: 18));
      case 'image':
        final id = item['media_asset_id'] as String?;
        final url = id != null ? media[id]?.url : null;
        if (url == null) return const SizedBox.shrink();
        return CachedNetworkImage(imageUrl: url);
      case 'audio':
        final id = item['media_asset_id'] as String?;
        final url = id != null ? media[id]?.url : null;
        if (url == null) return const SizedBox.shrink();
        return _AudioPlayer(url: url);
      case 'lexeme':
        final id = item['lexeme_id'] as String?;
        final lemma = id != null
            ? lexemeDisplayLabel(
                lexemeId: id,
                lexemes: lexemes,
                embeddedLemma: item['lemma'] as String?,
              )
            : null;
        return Text(
          lemma ?? '',
          style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
        );
      default:
        return const SizedBox.shrink();
    }
  }
}

/// Gradable block UI for the student player.
class BlockExerciseView extends StatefulWidget {
  final FlowBlockDto block;
  final Map<String, ResolvedLexemeDto> lexemes;
  final Map<String, ResolvedMediaDto> media;
  final void Function(Map<String, dynamic> response) onSubmit;
  final bool submitting;
  final bool? lastCorrect;
  final List<AlphabetLetterDto> alphabetLetters;
  final String targetLanguageCode;

  const BlockExerciseView({
    super.key,
    required this.block,
    required this.lexemes,
    required this.media,
    required this.onSubmit,
    this.submitting = false,
    this.lastCorrect,
    this.alphabetLetters = const [],
    this.targetLanguageCode = 'evn',
  });

  @override
  State<BlockExerciseView> createState() => _BlockExerciseViewState();
}

class _BlockExerciseViewState extends State<BlockExerciseView> {
  int? _selectedIndex;
  final _textController = TextEditingController();
  List<int> _order = [];

  @override
  void dispose() {
    _textController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final type = widget.block.blockType;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (_hasPrompt(type)) _buildPrompt(),
        const SizedBox(height: 16),
        _buildInput(type),
        if (widget.lastCorrect != null) ...[
          const SizedBox(height: 12),
          Text(
            widget.lastCorrect! ? 'Верно!' : 'Неверно, попробуйте ещё',
            style: TextStyle(
              color: widget.lastCorrect! ? AppTheme.successColor : AppTheme.errorColor,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
        const SizedBox(height: 16),
        ElevatedButton(
          onPressed: widget.submitting ? null : _submit,
          child: widget.submitting
              ? const SizedBox(
                  height: 20,
                  width: 20,
                  child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                )
              : const Text('Проверить'),
        ),
      ],
    );
  }

  bool _hasPrompt(String type) =>
      type.contains('prompt') || type.contains('listen') || type.contains('word_choose');

  Widget _buildPrompt() {
    final prompt = widget.block.config['prompt'] as Map<String, dynamic>?;
    final items = prompt?['items'] as List<dynamic>? ?? [];
    return Column(
      children: [
        for (final raw in items)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: BlockContentView(
              block: FlowBlockDto(
                id: widget.block.id,
                blockType: 'text',
                config: {'items': [raw]},
                isGradable: false,
              ),
              lexemes: widget.lexemes,
              media: widget.media,
            ),
          ),
      ],
    );
  }

  Widget _buildInput(String type) {
    if (type.contains('choose_word') ||
        type.contains('choose_translation') ||
        type.contains('choose_image') ||
        type.contains('listen_choose')) {
      return _choicesList();
    }
    if (type.contains('type_word') || type.contains('sentence_type')) {
      return _studyTextInput();
    }
    if (type.contains('word_order')) {
      return _wordOrder();
    }
    if (type == 'gap_sentence_choose_word') {
      final choices = widget.block.config['choices'] as List<dynamic>? ?? [];
      if (choices.isNotEmpty) return _choicesList();
      return _gapSentence();
    }
    return TextField(
      controller: _textController,
      decoration: const InputDecoration(hintText: 'Ответ'),
    );
  }

  bool get _usesStudyKeyboard => widget.alphabetLetters.isNotEmpty;

  Widget _studyTextInput() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _textController,
          readOnly: _usesStudyKeyboard,
          showCursor: true,
          decoration: InputDecoration(
            hintText: _usesStudyKeyboard ? 'Наберите ответ на клавиатуре ниже' : 'Введите ответ',
            suffixIcon: _textController.text.isNotEmpty
                ? IconButton(
                    icon: const Icon(Icons.clear),
                    onPressed: () => setState(() => _textController.clear()),
                  )
                : null,
          ),
          onChanged: (_) => setState(() {}),
        ),
        if (_usesStudyKeyboard) ...[
          const SizedBox(height: 12),
          StudyLanguageKeyboard(
            title: widget.targetLanguageCode == 'evn'
                ? 'Эвенская клавиатура'
                : 'Клавиатура языка',
            letters: widget.alphabetLetters,
            onCharacter: (ch) => setState(() {
              _textController.text += ch;
              _textController.selection = TextSelection.collapsed(
                offset: _textController.text.length,
              );
            }),
            onBackspace: () => setState(() {
              final text = _textController.text;
              if (text.isEmpty) return;
              _textController.text = text.substring(0, text.length - 1);
              _textController.selection = TextSelection.collapsed(
                offset: _textController.text.length,
              );
            }),
            onSpace: () => setState(() {
              _textController.text += ' ';
              _textController.selection = TextSelection.collapsed(
                offset: _textController.text.length,
              );
            }),
          ),
        ],
      ],
    );
  }

  Widget _choicesList() {
    final choices = widget.block.config['choices'] as List<dynamic>? ?? [];
    final isImage = widget.block.blockType.contains('image');
    return Column(
      children: [
        for (var i = 0; i < choices.length; i++)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: OutlinedButton(
              onPressed: () => setState(() => _selectedIndex = i),
              style: OutlinedButton.styleFrom(
                backgroundColor:
                    _selectedIndex == i ? AppTheme.primaryColor.withOpacity(0.1) : null,
                side: BorderSide(
                  color: _selectedIndex == i ? AppTheme.primaryColor : AppTheme.dividerColor,
                ),
              ),
              child: Align(
                alignment: Alignment.centerLeft,
                child: isImage
                    ? _choiceImage(choices[i] as Map<String, dynamic>)
                    : Text(_choiceLabel(choices[i] as Map<String, dynamic>)),
              ),
            ),
          ),
      ],
    );
  }

  Widget _choiceImage(Map<String, dynamic> choice) {
    final id = choice['media_asset_id'] as String?;
    final url = id != null ? widget.media[id]?.url : null;
    if (url == null) return const Text('—');
    return CachedNetworkImage(imageUrl: url, height: 80);
  }

  String _choiceLabel(Map<String, dynamic> choice) {
    if (choice['text'] != null) return choice['text'].toString();
    final lexId = choice['lexeme_id'] as String?;
    if (lexId != null) {
      return lexemeDisplayLabel(
        lexemeId: lexId,
        lexemes: widget.lexemes,
        embeddedLemma: choice['lemma'] as String?,
      );
    }
    return '—';
  }

  Widget _gapSentence() {
    final body = widget.block.config['body']?.toString() ?? '';
    return Text(body.replaceAllMapped(RegExp(r'\{\{gap:\d+\}\}'), (_) => '_____'));
  }

  Widget _wordOrder() {
    final tokens = (widget.block.config['tokens'] as List<dynamic>? ??
            widget.block.config['sub_items']?[0]?['tokens'] as List<dynamic>? ??
            [])
        .map((e) => e.toString())
        .toList();
    if (_order.length != tokens.length) {
      _order = List.generate(tokens.length, (i) => i);
    }
    return ReorderableListView(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      onReorder: (oldIndex, newIndex) {
        setState(() {
          if (newIndex > oldIndex) newIndex--;
          final item = _order.removeAt(oldIndex);
          _order.insert(newIndex, item);
        });
      },
      children: [
        for (var i = 0; i < _order.length; i++)
          ListTile(
            key: ValueKey('token_$i'),
            title: Text(_tokenLabel(tokens[_order[i]])),
            leading: const Icon(Icons.drag_handle),
          ),
      ],
    );
  }

  String _tokenLabel(String token) {
    if (widget.lexemes[token] != null) return widget.lexemes[token]!.lemma;
    if (_looksLikeUuid(token)) return 'Слово';
    return token;
  }

  bool _looksLikeUuid(String value) {
    return RegExp(
      r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$',
    ).hasMatch(value);
  }

  void _submit() {
    final type = widget.block.blockType;
    Map<String, dynamic> response;
    if (_selectedIndex != null &&
        (type.contains('choose') || type.contains('listen_choose'))) {
      response = {'selected_index': _selectedIndex};
    } else if (type.contains('type') || type.contains('sentence_type')) {
      response = {'text': _textController.text.trim()};
    } else if (type.contains('word_order')) {
      response = {'order': _order};
    } else {
      response = {'text': _textController.text.trim()};
    }
    widget.onSubmit(response);
  }
}

class _AudioPlayer extends StatefulWidget {
  final String url;
  const _AudioPlayer({required this.url});

  @override
  State<_AudioPlayer> createState() => _AudioPlayerState();
}

class _AudioPlayerState extends State<_AudioPlayer> {
  late VideoPlayerController _controller;

  @override
  void initState() {
    super.initState();
    _controller = VideoPlayerController.networkUrl(Uri.parse(widget.url))
      ..initialize().then((_) => setState(() {}));
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (!_controller.value.isInitialized) {
      return const CircularProgressIndicator();
    }
    return IconButton(
      icon: Icon(_controller.value.isPlaying ? Icons.pause : Icons.play_arrow),
      onPressed: () {
        setState(() {
          _controller.value.isPlaying ? _controller.pause() : _controller.play();
        });
      },
    );
  }
}

class _VideoPlayer extends StatefulWidget {
  final String url;
  const _VideoPlayer({required this.url});

  @override
  State<_VideoPlayer> createState() => _VideoPlayerState();
}

class _VideoPlayerState extends State<_VideoPlayer> {
  late VideoPlayerController _controller;

  @override
  void initState() {
    super.initState();
    _controller = VideoPlayerController.networkUrl(Uri.parse(widget.url))
      ..initialize().then((_) {
        setState(() {});
        _controller.play();
      });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (!_controller.value.isInitialized) {
      return const AspectRatio(
        aspectRatio: 16 / 9,
        child: Center(child: CircularProgressIndicator()),
      );
    }
    return AspectRatio(
      aspectRatio: _controller.value.aspectRatio,
      child: VideoPlayer(_controller),
    );
  }
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull => isEmpty ? null : first;
}
