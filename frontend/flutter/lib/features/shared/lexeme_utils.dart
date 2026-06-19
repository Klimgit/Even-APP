import 'package:online_cource_app/api/models/learning_dto.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';

/// Collect lexeme UUIDs referenced anywhere in a block config JSON tree.
Set<String> collectLexemeIdsFromConfig(Map<String, dynamic> config) {
  final ids = <String>{};
  void walk(dynamic value) {
    if (value is Map) {
      for (final entry in value.entries) {
        final key = entry.key.toString();
        final v = entry.value;
        if (key == 'lexeme_id' || key == 'correct_lexeme_id' || key == 'answer_lexeme_id') {
          final id = v?.toString() ?? '';
          if (_looksLikeUuid(id)) ids.add(id);
        } else if (key == 'lexeme_ids' && v is List) {
          for (final item in v) {
            final id = item.toString();
            if (_looksLikeUuid(id)) ids.add(id);
          }
        } else if (key == 'choices' && v is List) {
          for (final item in v) {
            if (item is String && _looksLikeUuid(item)) {
              ids.add(item);
            } else {
              walk(item);
            }
          }
        } else {
          walk(v);
        }
      }
    } else if (value is List) {
      for (final item in value) {
        walk(item);
      }
    }
  }

  walk(config);
  return ids;
}

Set<String> collectLexemeIdsFromBlocks(Iterable<dynamic> blocks) {
  final ids = <String>{};
  for (final block in blocks) {
    Map<String, dynamic>? config;
    if (block is Map) {
      config = Map<String, dynamic>.from(block['config'] as Map? ?? {});
    } else {
      try {
        config = (block as dynamic).config as Map<String, dynamic>?;
      } catch (_) {
        config = null;
      }
    }
    if (config != null) ids.addAll(collectLexemeIdsFromConfig(config));
  }
  return ids;
}

bool _looksLikeUuid(String value) {
  if (value.length != 36) return false;
  return RegExp(
    r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$',
  ).hasMatch(value);
}

/// Merge API-resolved lexemes with labels stored in block config and lexicon lookups.
Map<String, ResolvedLexemeDto> buildLexemeDisplayMap({
  required Map<String, ResolvedLexemeDto> resolved,
  required Iterable<Map<String, dynamic>> blockConfigs,
  Map<String, String>? extraLabels,
}) {
  final out = Map<String, ResolvedLexemeDto>.from(resolved);

  void addLabel(String id, String lemma) {
    if (lemma.isEmpty || !_looksLikeUuid(id)) return;
    out[id] = ResolvedLexemeDto(id: id, lemma: lemma);
  }

  for (final labels in extraLabels?.entries ?? const Iterable.empty()) {
    addLabel(labels.key, labels.value);
  }

  for (final config in blockConfigs) {
    _extractEmbeddedLabels(config, addLabel);
  }

  return out;
}

void _extractEmbeddedLabels(
  Map<String, dynamic> config,
  void Function(String id, String lemma) add,
) {
  void walk(dynamic value) {
    if (value is Map) {
      final id = value['lexeme_id']?.toString();
      final lemma = value['lemma']?.toString();
      if (id != null && lemma != null && lemma.isNotEmpty) {
        add(id, lemma);
      }
      final labels = value['lexeme_labels'];
      if (labels is Map) {
        for (final entry in labels.entries) {
          add(entry.key.toString(), entry.value.toString());
        }
      }
      for (final v in value.values) {
        walk(v);
      }
    } else if (value is List) {
      for (final item in value) {
        walk(item);
      }
    }
  }

  walk(config);
}

/// Load lemma strings for the given IDs into [cache] using platform then teacher lexicon.
Future<void> preloadLemmaCache({
  required LexiconRepository lexicon,
  required String languageCode,
  required Set<String> ids,
  required Map<String, String> cache,
  bool includeTeacherLexicon = true,
}) async {
  if (ids.isEmpty) return;
  final missing = ids.where((id) => !cache.containsKey(id)).toSet();
  if (missing.isEmpty) return;

  Future<void> scan(List<LexemeDto> items) async {
    for (final item in items) {
      if (missing.contains(item.id)) {
        cache[item.id] = item.lemma;
      }
    }
  }

  var page = 1;
  while (missing.any((id) => !cache.containsKey(id)) && page <= 5) {
    final resp = await lexicon.listPlatformLexicon(
      languageCode: languageCode,
      page: page,
      limit: 100,
    );
    await scan(resp.items);
    if (resp.items.length < 100) break;
    page++;
  }

  if (!includeTeacherLexicon) return;
  page = 1;
  while (missing.any((id) => !cache.containsKey(id)) && page <= 5) {
    final resp = await lexicon.listTeacherLexicon(
      languageCode: languageCode,
      page: page,
      limit: 100,
    );
    await scan(resp.items);
    if (resp.items.length < 100) break;
    page++;
  }
}

String lexemeDisplayLabel({
  required String? lexemeId,
  required Map<String, ResolvedLexemeDto> lexemes,
  Map<String, String>? lemmaCache,
  String? embeddedLemma,
}) {
  if (embeddedLemma != null && embeddedLemma.isNotEmpty) return embeddedLemma;
  if (lexemeId == null || lexemeId.isEmpty) return 'Выбрать…';
  final fromResolved = lexemes[lexemeId]?.lemma;
  if (fromResolved != null && fromResolved.isNotEmpty) return fromResolved;
  final fromCache = lemmaCache?[lexemeId];
  if (fromCache != null && fromCache.isNotEmpty) return fromCache;
  return 'Слово';
}
