import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/learning_dto.dart';
import 'package:online_cource_app/api/repositories/learning_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Student personal vocabulary (GET /dictionary).
class StudentDictionaryScreen extends StatefulWidget {
  const StudentDictionaryScreen({super.key});

  @override
  State<StudentDictionaryScreen> createState() => _StudentDictionaryScreenState();
}

class _StudentDictionaryScreenState extends State<StudentDictionaryScreen> {
  final _learning = Get.find<LearningRepository>();
  final _search = TextEditingController();
  List<VocabularyEntryDto> _items = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
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
      _items = await _learning.listDictionary(query: _search.text.trim());
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: Text('База знаний', style: Theme.of(context).textTheme.titleLarge),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: SearchFilterBar(
              controller: _search,
              hint: 'Поиск по словам…',
              onChanged: _load,
            ),
          ),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _items.isEmpty
                    ? const EmptyState(
                        icon: Icons.menu_book_outlined,
                        message: 'Слов пока нет — проходите уроки',
                      )
                    : ListView.separated(
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        itemCount: _items.length,
                        separatorBuilder: (_, __) => const Divider(height: 1),
                        itemBuilder: (ctx, i) {
                          final w = _items[i];
                          return ListTile(
                            title: Text(w.lemma, style: const TextStyle(fontWeight: FontWeight.w600)),
                            subtitle: Text(w.partOfSpeech ?? '—'),
                            trailing: Text(
                              '${(w.mastery * 100).toStringAsFixed(0)}%',
                              style: TextStyle(color: AppTheme.primaryColor),
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }
}
