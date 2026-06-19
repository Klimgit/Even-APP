class LexemeDto {
  final String id;
  final String languageId;
  final String lemma;
  final String? partOfSpeech;
  final String? notes;
  final List<LexemeTranslationDto> translations;
  final List<LexemeFormDto> forms;
  final List<LexemeMediaDto> media;
  final String? primaryImageUrl;
  final String? primaryAudioUrl;

  const LexemeDto({
    required this.id,
    required this.languageId,
    required this.lemma,
    this.partOfSpeech,
    this.notes,
    required this.translations,
    required this.forms,
    required this.media,
    this.primaryImageUrl,
    this.primaryAudioUrl,
  });

  factory LexemeDto.fromJson(Map<String, dynamic> json) => LexemeDto(
        id: json['id'] as String,
        languageId: json['language_id'] as String,
        lemma: json['lemma'] as String,
        partOfSpeech: json['part_of_speech'] as String?,
        notes: json['notes'] as String?,
        translations: (json['translations'] as List<dynamic>? ?? [])
            .map((e) => LexemeTranslationDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        forms: (json['forms'] as List<dynamic>? ?? [])
            .map((e) => LexemeFormDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        media: (json['media'] as List<dynamic>? ?? [])
            .map((e) => LexemeMediaDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        primaryImageUrl: json['primary_image_url'] as String?,
        primaryAudioUrl: json['primary_audio_url'] as String?,
      );

  String? translationText(String? uiLanguageId) {
    if (translations.isEmpty) return null;
    if (uiLanguageId != null) {
      for (final t in translations) {
        if (t.targetLanguageId == uiLanguageId) return t.text;
      }
    }
    return translations.first.text;
  }
}

class LexemeTranslationDto {
  final String id;
  final String targetLanguageId;
  final String text;

  const LexemeTranslationDto({
    required this.id,
    required this.targetLanguageId,
    required this.text,
  });

  factory LexemeTranslationDto.fromJson(Map<String, dynamic> json) =>
      LexemeTranslationDto(
        id: json['id'] as String,
        targetLanguageId: json['target_language_id'] as String,
        text: json['text'] as String,
      );
}

class LexemeFormDto {
  final String id;
  final String lexemeId;
  final String form;

  const LexemeFormDto({
    required this.id,
    required this.lexemeId,
    required this.form,
  });

  factory LexemeFormDto.fromJson(Map<String, dynamic> json) => LexemeFormDto(
        id: json['id'] as String,
        lexemeId: json['lexeme_id'] as String,
        form: json['form'] as String,
      );
}

class LexemeMediaDto {
  final String id;
  final String kind;
  final String url;
  final bool isPrimary;

  const LexemeMediaDto({
    required this.id,
    required this.kind,
    required this.url,
    required this.isPrimary,
  });

  factory LexemeMediaDto.fromJson(Map<String, dynamic> json) => LexemeMediaDto(
        id: json['id'] as String,
        kind: json['kind'] as String,
        url: json['url'] as String,
        isPrimary: json['is_primary'] as bool? ?? false,
      );
}

class LexemeListResponseDto {
  final List<LexemeDto> items;
  final int total;
  final int page;
  final int limit;

  const LexemeListResponseDto({
    required this.items,
    required this.total,
    required this.page,
    required this.limit,
  });

  factory LexemeListResponseDto.fromJson(Map<String, dynamic> json) =>
      LexemeListResponseDto(
        items: (json['items'] as List<dynamic>? ?? [])
            .map((e) => LexemeDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        total: json['total'] as int? ?? 0,
        page: json['page'] as int? ?? 1,
        limit: json['limit'] as int? ?? 20,
      );
}

class AlphabetLetterDto {
  final String id;
  final String languageId;
  final String character;
  final String? upperChar;
  final int sortOrder;
  final String? label;
  final String? transcription;

  const AlphabetLetterDto({
    required this.id,
    required this.languageId,
    required this.character,
    this.upperChar,
    required this.sortOrder,
    this.label,
    this.transcription,
  });

  factory AlphabetLetterDto.fromJson(Map<String, dynamic> json) => AlphabetLetterDto(
        id: json['id'] as String,
        languageId: json['language_id'] as String,
        character: json['character'] as String,
        upperChar: json['upper_char'] as String?,
        sortOrder: json['sort_order'] as int? ?? 0,
        label: json['label'] as String?,
        transcription: json['transcription'] as String?,
      );

  String displayChar({required bool shift}) {
    if (shift && upperChar != null && upperChar!.isNotEmpty) {
      return upperChar!;
    }
    return character;
  }
}

class PlatformLanguageDto {
  final String id;
  final String code;
  final String name;
  final String nativeName;
  final bool isActive;

  const PlatformLanguageDto({
    required this.id,
    required this.code,
    required this.name,
    required this.nativeName,
    required this.isActive,
  });

  factory PlatformLanguageDto.fromJson(Map<String, dynamic> json) =>
      PlatformLanguageDto(
        id: json['id'] as String,
        code: json['code'] as String,
        name: json['name'] as String,
        nativeName: json['native_name'] as String? ?? json['name'] as String,
        isActive: json['is_active'] as bool? ?? true,
      );
}
