import 'package:dio/dio.dart';

import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/api/base_repository.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/api/models/learning_dto.dart';

class LexiconRepository extends BaseRepository {
  LexiconRepository(super.client);

  Future<List<LanguageDto>> listPublicLanguages() async {
    try {
      // Public languages are mounted at gateway root `/languages`, not under `/api/v1`.
      final res = await client.dio.get<List<dynamic>>(
        '$kGatewayOrigin/languages',
      );
      return (res.data ?? [])
          .map((e) => LanguageDto.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<List<AlphabetLetterDto>> getLanguageAlphabet(String languageCode) async {
    try {
      final res = await client.dio.get<List<dynamic>>(
        '$kGatewayOrigin/languages/$languageCode/alphabet',
      );
      final letters = (res.data ?? [])
          .map((e) => AlphabetLetterDto.fromJson(e as Map<String, dynamic>))
          .toList();
      letters.sort((a, b) => a.sortOrder.compareTo(b.sortOrder));
      return letters;
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<List<PlatformLanguageDto>> listPlatformLanguages() async {
    final list = await getList('/platform/languages');
    return list
        .map((e) => PlatformLanguageDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<LexemeListResponseDto> listPlatformLexicon({
    required String languageCode,
    String? query,
    int page = 1,
    int limit = 50,
  }) async {
    return LexemeListResponseDto.fromJson(
      await getJson('/platform/languages/$languageCode/lexicon', query: {
        if (query != null && query.isNotEmpty) 'q': query,
        'page': page,
        'limit': limit,
      }),
    );
  }

  Future<LexemeDto> createPlatformLexeme({
    required String languageCode,
    required String lemma,
    String? partOfSpeech,
  }) async {
    return LexemeDto.fromJson(
      await postJson('/platform/languages/$languageCode/lexicon', {
        'lemma': lemma,
        if (partOfSpeech != null) 'part_of_speech': partOfSpeech,
      }),
    );
  }

  Future<LexemeDto> patchPlatformLexeme({
    required String lexemeId,
    String? lemma,
    String? notes,
  }) async {
    final body = <String, dynamic>{};
    if (lemma != null) body['lemma'] = lemma;
    if (notes != null) body['notes'] = notes;
    return LexemeDto.fromJson(await patchJson('/platform/lexemes/$lexemeId', body));
  }

  Future<void> deletePlatformLexeme(String lexemeId) =>
      delete('/platform/lexemes/$lexemeId');

  Future<LexemeTranslationDto> addPlatformTranslation({
    required String lexemeId,
    required String targetLanguageId,
    required String text,
  }) async {
    return LexemeTranslationDto.fromJson(
      await postJson('/platform/lexemes/$lexemeId/translations', {
        'target_language_id': targetLanguageId,
        'text': text,
      }),
    );
  }

  Future<LexemeListResponseDto> listTeacherLexicon({
    required String languageCode,
    String? query,
    int page = 1,
    int limit = 50,
  }) async {
    return LexemeListResponseDto.fromJson(
      await getJson('/teacher/languages/$languageCode/lexicon', query: {
        if (query != null && query.isNotEmpty) 'q': query,
        'page': page,
        'limit': limit,
      }),
    );
  }

  Future<LexemeDto> createTeacherLexeme({
    required String languageCode,
    required String lemma,
    String? partOfSpeech,
  }) async {
    return LexemeDto.fromJson(
      await postJson('/teacher/languages/$languageCode/lexicon', {
        'lemma': lemma,
        if (partOfSpeech != null) 'part_of_speech': partOfSpeech,
      }),
    );
  }

  Future<LexemeDto> patchTeacherLexeme({
    required String lexemeId,
    String? lemma,
    String? notes,
  }) async {
    final body = <String, dynamic>{};
    if (lemma != null) body['lemma'] = lemma;
    if (notes != null) body['notes'] = notes;
    return LexemeDto.fromJson(await patchJson('/teacher/lexemes/$lexemeId', body));
  }

  Future<void> deleteTeacherLexeme(String lexemeId) =>
      delete('/teacher/lexemes/$lexemeId');

  Future<LexemeTranslationDto> addTeacherTranslation({
    required String lexemeId,
    required String targetLanguageId,
    required String text,
  }) async {
    return LexemeTranslationDto.fromJson(
      await postJson('/teacher/lexemes/$lexemeId/translations', {
        'target_language_id': targetLanguageId,
        'text': text,
      }),
    );
  }
}
