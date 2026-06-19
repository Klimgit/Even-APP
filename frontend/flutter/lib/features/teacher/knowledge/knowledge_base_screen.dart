import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/features/shared/alphabet_controller.dart';
import 'package:online_cource_app/features/shared/study_text_field.dart';

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
                            title: Text(lex.lemma, style: const TextStyle(fontWeight: FontWeight.w600)),
                            subtitle: Text(trans.isEmpty ? '—' : trans),
                            trailing: canEdit
                                ? IconButton(
                                    icon: const Icon(Icons.delete_outline),
                                    onPressed: () => _deleteLexeme(lex),
                                  )
                                : null,
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
