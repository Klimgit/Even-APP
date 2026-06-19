import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';
import 'package:online_cource_app/api/token_storage.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/features/shared/alphabet_controller.dart';
import 'package:online_cource_app/features/shared/media_url.dart';
import 'package:online_cource_app/features/shared/study_text_field.dart';
import 'package:online_cource_app/features/teacher/pickers/media_picker.dart';
import 'package:video_player/video_player.dart';

/// Platform (admin) + teacher personal lexicon management.
class KnowledgeBaseScreen extends StatefulWidget {
  const KnowledgeBaseScreen({super.key});

  @override
  State<KnowledgeBaseScreen> createState() => _KnowledgeBaseScreenState();
}

class _KnowledgeBaseScreenState extends State<KnowledgeBaseScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabs;
  final _lexicon = Get.find<LexiconRepository>();
  final _search = TextEditingController();
  String? _languageCode;
  List<LexemeDto> _items = [];
  bool _loading = false;
  bool _platformTab = false;

  bool get _canEditPlatform => Get.find<ApiAuthController>().isAdmin;

  @override
  void initState() {
    super.initState();
    if (!Get.isRegistered<AlphabetController>()) {
      Get.put(AlphabetController(), permanent: true);
    }
    _tabs = TabController(length: _canEditPlatform ? 2 : 1, vsync: this);
    _platformTab = _canEditPlatform;
    _tabs.addListener(() {
      if (!_tabs.indexIsChanging) {
        setState(() => _platformTab = _tabs.index == 0 && _canEditPlatform);
        _load();
      }
    });
    _load();
  }

  @override
  void dispose() {
    _search.dispose();
    _tabs.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final langs = Get.find<LanguagesController>().languages;
    if (langs.isEmpty) return;
    _languageCode ??= langs.first.code;
    setState(() => _loading = true);
    try {
      final resp = _platformTab && _canEditPlatform
          ? await _lexicon.listPlatformLexicon(
              languageCode: _languageCode!,
              query: _search.text.trim(),
            )
          : await _lexicon.listTeacherLexicon(
              languageCode: _languageCode!,
              query: _search.text.trim(),
            );
      setState(() => _items = resp.items);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _addLexeme() async {
    if (_platformTab && !_canEditPlatform) return;
    final lemmaCtrl = TextEditingController();
    final transCtrl = TextEditingController();
    final langs = Get.find<LanguagesController>().languages;
    var uiLangId = langs.isNotEmpty ? langs.first.id : '';
    final studyCode = _languageCode ?? 'evn';

    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Новое слово'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              StudyTextField(
                controller: lemmaCtrl,
                languageCode: studyCode,
                hintText: 'лемму на языке',
              ),
              const SizedBox(height: 12),
              TextField(
                controller: transCtrl,
                decoration: const InputDecoration(hintText: 'Перевод (русский)'),
              ),
              if (langs.isNotEmpty) ...[
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: uiLangId,
                  decoration: const InputDecoration(labelText: 'Язык перевода'),
                  items: langs
                      .map((l) => DropdownMenuItem(value: l.id, child: Text(l.nativeName)))
                      .toList(),
                  onChanged: (v) => uiLangId = v ?? uiLangId,
                ),
              ],
            ],
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Создать')),
        ],
      ),
    );
    if (ok != true || lemmaCtrl.text.trim().isEmpty) return;

    final lexeme = _platformTab && _canEditPlatform
        ? await _lexicon.createPlatformLexeme(
            languageCode: _languageCode!,
            lemma: lemmaCtrl.text.trim(),
          )
        : await _lexicon.createTeacherLexeme(
            languageCode: _languageCode!,
            lemma: lemmaCtrl.text.trim(),
          );

    if (transCtrl.text.trim().isNotEmpty && uiLangId.isNotEmpty) {
      if (_platformTab && _canEditPlatform) {
        await _lexicon.addPlatformTranslation(
          lexemeId: lexeme.id,
          targetLanguageId: uiLangId,
          text: transCtrl.text.trim(),
        );
      } else {
        await _lexicon.addTeacherTranslation(
          lexemeId: lexeme.id,
          targetLanguageId: uiLangId,
          text: transCtrl.text.trim(),
        );
      }
    }
    await _load();
  }

  Future<Map<String, String>> _authHeaders() async {
    final token = await Get.find<TokenStorage>().readAccessToken();
    if (token == null || token.isEmpty) return const {};
    return {'Authorization': 'Bearer $token'};
  }

  Future<void> _showLexemeDetail(LexemeDto lexeme) async {
    final langs = Get.find<LanguagesController>().languages;
    final studyCode = _languageCode ?? 'evn';
    final langId = langs.firstWhere((l) => l.code == studyCode, orElse: () => langs.first).id;
    final headers = await _authHeaders();
    if (!mounted) return;

    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) => DraggableScrollableSheet(
        expand: false,
        initialChildSize: 0.55,
        maxChildSize: 0.9,
        builder: (_, controller) => ListView(
          controller: controller,
          padding: const EdgeInsets.all(16),
          children: [
            Text(lexeme.lemma, style: Theme.of(ctx).textTheme.headlineSmall),
            Text(lexeme.translations.map((t) => t.text).join(', ')),
            const SizedBox(height: 16),
            if (lexeme.primaryImageUrl != null)
              ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: FutureBuilder<Map<String, String>>(
                  future: _authHeaders(),
                  builder: (context, snap) {
                    final h = snap.data ?? headers;
                    return CachedNetworkImage(
                      imageUrl: resolveApiMediaUrl(lexeme.primaryImageUrl),
                      httpHeaders: h,
                      height: 180,
                      width: double.infinity,
                      fit: BoxFit.cover,
                      errorWidget: (_, __, ___) => Container(
                        height: 180,
                        color: Colors.grey.shade200,
                        child: const Icon(Icons.image_not_supported_outlined),
                      ),
                    );
                  },
                ),
              ),
            if (lexeme.primaryAudioUrl != null) ...[
              const SizedBox(height: 12),
              _LexemeAudioPreview(url: resolveApiMediaUrl(lexeme.primaryAudioUrl!)),
            ],
            const SizedBox(height: 20),
            const Text('Прикрепить медиа', style: TextStyle(fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: [
                OutlinedButton.icon(
                  onPressed: () => _attachMedia(lexeme, langId, studyCode, 'image'),
                  icon: const Icon(Icons.image_outlined),
                  label: const Text('Картинка'),
                ),
                OutlinedButton.icon(
                  onPressed: () => _attachMedia(lexeme, langId, studyCode, 'audio'),
                  icon: const Icon(Icons.audiotrack_outlined),
                  label: const Text('Аудио'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _attachMedia(
    LexemeDto lexeme,
    String languageId,
    String languageCode,
    String kind,
  ) async {
    final mediaKind = kind == 'audio' ? 'audio' : 'image';
    final picked = await showMediaPicker(
      context: context,
      languageCode: languageCode,
      languageId: languageId,
      kind: mediaKind,
    );
    if (picked == null) return;
    final lexemeKind = kind == 'audio' ? 'audio_word' : 'image';
    if (_platformTab && _canEditPlatform) {
      await _lexicon.addPlatformLexemeMedia(
        lexemeId: lexeme.id,
        mediaAssetId: picked.id,
        kind: lexemeKind,
      );
    } else {
      await _lexicon.addTeacherLexemeMedia(
        lexemeId: lexeme.id,
        mediaAssetId: picked.id,
        kind: lexemeKind,
      );
    }
    if (mounted) Navigator.pop(context);
    await _load();
  }

  Future<void> _deleteLexeme(LexemeDto lexeme) async {
    if (_platformTab && !_canEditPlatform) return;
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить слово?'),
        content: Text(lexeme.lemma),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Удалить')),
        ],
      ),
    );
    if (ok != true) return;
    if (_platformTab && _canEditPlatform) {
      await _lexicon.deletePlatformLexeme(lexeme.id);
    } else {
      await _lexicon.deleteTeacherLexeme(lexeme.id);
    }
    await _load();
  }

  @override
  Widget build(BuildContext context) {
    final canEdit = _platformTab ? _canEditPlatform : true;
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('База знаний', style: Theme.of(context).textTheme.titleLarge),
            if (_canEditPlatform)
              TabBar(
                controller: _tabs,
                tabs: const [
                  Tab(text: 'Платформа'),
                  Tab(text: 'Личная база'),
                ],
              )
            else
              const SizedBox.shrink(),
            const SizedBox(height: 8),
          SearchFilterBar(
            controller: _search,
            hint: 'Поиск по лемме и переводу…',
            onChanged: _load,
            trailing: LanguageFilterChip(
              selectedCode: _languageCode,
              onChanged: (v) {
                setState(() => _languageCode = v ?? _languageCode);
                _load();
              },
            ),
          ),
          const SizedBox(height: 16),
          if (canEdit)
            Align(
              alignment: Alignment.centerRight,
              child: ElevatedButton.icon(
                onPressed: _addLexeme,
                icon: const Icon(Icons.add),
                label: const Text('Добавить слово'),
              ),
            ),
          const SizedBox(height: 8),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _items.isEmpty
                    ? const EmptyState(icon: Icons.menu_book_outlined, message: 'Слов не найдено')
                    : ListView.separated(
                        itemCount: _items.length,
                        separatorBuilder: (_, __) => const Divider(height: 1),
                        itemBuilder: (ctx, i) {
                          final lex = _items[i];
                          final trans = lex.translations.map((t) => t.text).join(', ');
                          return ListTile(
                            onTap: () => _showLexemeDetail(lex),
                            leading: _LexemeThumb(
                              imageUrl: lex.primaryImageUrl,
                              headersFuture: _authHeaders(),
                            ),
                            title: Text(
                              lex.lemma,
                              style: const TextStyle(fontWeight: FontWeight.w600),
                            ),
                            subtitle: Text(trans.isEmpty ? '—' : trans),
                            trailing: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                if (lex.primaryAudioUrl != null)
                                  const Icon(Icons.volume_up_outlined, size: 20),
                                if (canEdit)
                                  IconButton(
                                    icon: const Icon(Icons.delete_outline),
                                    onPressed: () => _deleteLexeme(lex),
                                  ),
                              ],
                            ),
                          );
                        },
                      ),
          ),
        ],
        ),
      ),
    );
  }
}

class _LexemeThumb extends StatelessWidget {
  final String? imageUrl;
  final Future<Map<String, String>> headersFuture;

  const _LexemeThumb({required this.imageUrl, required this.headersFuture});

  @override
  Widget build(BuildContext context) {
    if (imageUrl == null || imageUrl!.isEmpty) {
      return CircleAvatar(
        backgroundColor: Colors.grey.shade200,
        child: const Icon(Icons.menu_book_outlined, size: 18),
      );
    }
    return FutureBuilder<Map<String, String>>(
      future: headersFuture,
      builder: (context, snap) {
        return ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: CachedNetworkImage(
            imageUrl: resolveApiMediaUrl(imageUrl),
            httpHeaders: snap.data ?? const {},
            width: 48,
            height: 48,
            fit: BoxFit.cover,
            errorWidget: (_, __, ___) => CircleAvatar(
              backgroundColor: Colors.grey.shade200,
              child: const Icon(Icons.image_not_supported_outlined, size: 18),
            ),
          ),
        );
      },
    );
  }
}

class _LexemeAudioPreview extends StatefulWidget {
  final String url;
  const _LexemeAudioPreview({required this.url});

  @override
  State<_LexemeAudioPreview> createState() => _LexemeAudioPreviewState();
}

class _LexemeAudioPreviewState extends State<_LexemeAudioPreview> {
  VideoPlayerController? _controller;

  @override
  void initState() {
    super.initState();
    _controller = VideoPlayerController.networkUrl(Uri.parse(widget.url))
      ..initialize().then((_) {
        if (mounted) setState(() {});
      });
  }

  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final c = _controller;
    if (c == null || !c.value.isInitialized) {
      return const LinearProgressIndicator(minHeight: 2);
    }
    return Row(
      children: [
        IconButton(
          icon: Icon(c.value.isPlaying ? Icons.pause_circle : Icons.play_circle),
          iconSize: 40,
          onPressed: () {
            setState(() {
              c.value.isPlaying ? c.pause() : c.play();
            });
          },
        ),
        const Expanded(child: Text('Прослушать произношение')),
      ],
    );
  }
}
