import 'content_dto.dart';

class LanguageDto {
  final String id;
  final String code;
  final String name;
  final String nativeName;

  const LanguageDto({
    required this.id,
    required this.code,
    required this.name,
    required this.nativeName,
  });

  factory LanguageDto.fromJson(Map<String, dynamic> json) => LanguageDto(
        id: json['id'] as String,
        code: json['code'] as String,
        name: json['name'] as String,
        nativeName: json['native_name'] as String? ?? json['name'] as String,
      );
}

class CourseListItemDto {
  final String id;
  final String title;
  final LanguageDto targetLanguage;
  final double? progressPercent;

  const CourseListItemDto({
    required this.id,
    required this.title,
    required this.targetLanguage,
    this.progressPercent,
  });

  factory CourseListItemDto.fromJson(Map<String, dynamic> json) => CourseListItemDto(
        id: json['id'] as String,
        title: json['title'] as String,
        targetLanguage:
            LanguageDto.fromJson(json['target_language'] as Map<String, dynamic>),
        progressPercent: (json['progress_percent'] as num?)?.toDouble(),
      );
}

class PublicCourseListItemDto {
  final String id;
  final String title;
  final LanguageDto targetLanguage;
  final bool isPublished;

  const PublicCourseListItemDto({
    required this.id,
    required this.title,
    required this.targetLanguage,
    required this.isPublished,
  });

  factory PublicCourseListItemDto.fromJson(Map<String, dynamic> json) =>
      PublicCourseListItemDto(
        id: json['id'] as String,
        title: json['title'] as String,
        targetLanguage:
            LanguageDto.fromJson(json['target_language'] as Map<String, dynamic>),
        isPublished: json['is_published'] as bool? ?? false,
      );
}

class ProgressSummaryDto {
  final int enrolledCourses;
  final int dictionaryWords;
  final int reviewDue;
  final int completedBlocks;
  final int completedLessons;
  final int inProgressLessons;
  final double averageScore;
  final int timeSpentSeconds;

  const ProgressSummaryDto({
    required this.enrolledCourses,
    required this.dictionaryWords,
    required this.reviewDue,
    required this.completedBlocks,
    required this.completedLessons,
    required this.inProgressLessons,
    required this.averageScore,
    required this.timeSpentSeconds,
  });

  factory ProgressSummaryDto.fromJson(Map<String, dynamic> json) => ProgressSummaryDto(
        enrolledCourses: json['enrolled_courses'] as int? ?? 0,
        dictionaryWords: json['dictionary_words'] as int? ?? 0,
        reviewDue: json['review_due'] as int? ?? 0,
        completedBlocks: json['completed_blocks'] as int? ?? 0,
        completedLessons: json['completed_lessons'] as int? ?? 0,
        inProgressLessons: json['in_progress_lessons'] as int? ?? 0,
        averageScore: (json['average_score'] as num?)?.toDouble() ?? 0,
        timeSpentSeconds: json['time_spent_seconds'] as int? ?? 0,
      );
}

class ResolvedLexemeDto {
  final String id;
  final String lemma;
  final String? partOfSpeech;

  const ResolvedLexemeDto({
    required this.id,
    required this.lemma,
    this.partOfSpeech,
  });

  factory ResolvedLexemeDto.fromJson(Map<String, dynamic> json) => ResolvedLexemeDto(
        id: json['id'] as String,
        lemma: json['lemma'] as String,
        partOfSpeech: json['part_of_speech'] as String?,
      );
}

class ResolvedMediaDto {
  final String id;
  final String url;
  final String displayName;
  final String mimeType;
  final String mediaKind;

  const ResolvedMediaDto({
    required this.id,
    required this.url,
    required this.displayName,
    required this.mimeType,
    required this.mediaKind,
  });

  factory ResolvedMediaDto.fromJson(Map<String, dynamic> json) => ResolvedMediaDto(
        id: json['id'] as String,
        url: json['url'] as String,
        displayName: json['display_name'] as String,
        mimeType: json['mime_type'] as String,
        mediaKind: json['media_kind'] as String,
      );
}

class LessonFlowDto {
  final String lessonId;
  final int version;
  final List<LessonFlowItemDto> items;

  const LessonFlowDto({
    required this.lessonId,
    required this.version,
    required this.items,
  });

  factory LessonFlowDto.fromJson(Map<String, dynamic> json) => LessonFlowDto(
        lessonId: json['lesson_id'] as String,
        version: json['version'] as int? ?? 1,
        items: (json['items'] as List<dynamic>? ?? [])
            .map((e) => LessonFlowItemDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class LessonFlowItemDto {
  final String kind;
  final FlowBlockDto? block;
  final Map<String, dynamic>? review;

  const LessonFlowItemDto({
    required this.kind,
    this.block,
    this.review,
  });

  factory LessonFlowItemDto.fromJson(Map<String, dynamic> json) => LessonFlowItemDto(
        kind: json['kind'] as String,
        block: json['block'] != null
            ? FlowBlockDto.fromJson(json['block'] as Map<String, dynamic>)
            : null,
        review: json['review'] as Map<String, dynamic>?,
      );
}

class FlowBlockDto {
  final String id;
  final String blockType;
  final Map<String, dynamic> config;
  final bool isGradable;
  final String? title;
  final String? displayLabel;

  const FlowBlockDto({
    required this.id,
    required this.blockType,
    required this.config,
    required this.isGradable,
    this.title,
    this.displayLabel,
  });

  factory FlowBlockDto.fromJson(Map<String, dynamic> json) => FlowBlockDto(
        id: json['id'] as String,
        blockType: json['block_type'] as String,
        config: Map<String, dynamic>.from(json['config'] as Map? ?? {}),
        isGradable: json['is_gradable'] as bool? ?? false,
        title: json['title'] as String?,
        displayLabel: json['display_label'] as String?,
      );
}

class BlockAttemptResultDto {
  final bool isCorrect;
  final double score;
  final Map<String, dynamic>? correctAnswer;
  final UserBlockProgressDto blockProgress;

  const BlockAttemptResultDto({
    required this.isCorrect,
    required this.score,
    this.correctAnswer,
    required this.blockProgress,
  });

  factory BlockAttemptResultDto.fromJson(Map<String, dynamic> json) =>
      BlockAttemptResultDto(
        isCorrect: json['is_correct'] as bool? ?? false,
        score: (json['score'] as num?)?.toDouble() ?? 0,
        correctAnswer: json['correct_answer'] as Map<String, dynamic>?,
        blockProgress: UserBlockProgressDto.fromJson(
          json['block_progress'] as Map<String, dynamic>,
        ),
      );
}

class UserBlockProgressDto {
  final String lessonBlockId;
  final String status;
  final double score;
  final int attempts;

  const UserBlockProgressDto({
    required this.lessonBlockId,
    required this.status,
    required this.score,
    required this.attempts,
  });

  factory UserBlockProgressDto.fromJson(Map<String, dynamic> json) =>
      UserBlockProgressDto(
        lessonBlockId: json['lesson_block_id'] as String,
        status: json['status'] as String? ?? 'not_started',
        score: (json['score'] as num?)?.toDouble() ?? 0,
        attempts: json['attempts'] as int? ?? 0,
      );
}

class CourseOutlineDto {
  final String courseId;
  final String title;
  final double progressPercent;
  final String? currentLessonId;
  final String? currentBlockId;
  final List<CourseOutlineModuleDto> modules;

  const CourseOutlineDto({
    required this.courseId,
    required this.title,
    required this.progressPercent,
    this.currentLessonId,
    this.currentBlockId,
    required this.modules,
  });

  factory CourseOutlineDto.fromJson(Map<String, dynamic> json) => CourseOutlineDto(
        courseId: json['course_id'] as String,
        title: json['title'] as String,
        progressPercent: (json['progress_percent'] as num?)?.toDouble() ?? 0,
        currentLessonId: json['current_lesson_id'] as String?,
        currentBlockId: json['current_block_id'] as String?,
        modules: (json['modules'] as List<dynamic>? ?? [])
            .map((e) => CourseOutlineModuleDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class CourseOutlineModuleDto {
  final String id;
  final String title;
  final int sortOrder;
  final double progressPercent;
  final List<CourseOutlineLessonDto> lessons;

  const CourseOutlineModuleDto({
    required this.id,
    required this.title,
    required this.sortOrder,
    required this.progressPercent,
    required this.lessons,
  });

  factory CourseOutlineModuleDto.fromJson(Map<String, dynamic> json) =>
      CourseOutlineModuleDto(
        id: json['id'] as String,
        title: json['title'] as String,
        sortOrder: json['sort_order'] as int? ?? 0,
        progressPercent: (json['progress_percent'] as num?)?.toDouble() ?? 0,
        lessons: (json['lessons'] as List<dynamic>? ?? [])
            .map((e) => CourseOutlineLessonDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class CourseOutlineLessonDto {
  final String id;
  final String title;
  final int sortOrder;
  final double progressPercent;

  const CourseOutlineLessonDto({
    required this.id,
    required this.title,
    required this.sortOrder,
    required this.progressPercent,
  });

  factory CourseOutlineLessonDto.fromJson(Map<String, dynamic> json) =>
      CourseOutlineLessonDto(
        id: json['id'] as String,
        title: json['title'] as String,
        sortOrder: json['sort_order'] as int? ?? 0,
        progressPercent: (json['progress_percent'] as num?)?.toDouble() ?? 0,
      );
}

class StudentLessonDto {
  final String id;
  final String courseId;
  final String title;
  final int sortOrder;
  final int version;
  final String status;
  final Map<String, ResolvedLexemeDto> resolvedLexemes;
  final Map<String, ResolvedMediaDto> resolvedMedia;
  final List<LessonSectionDto> sections;

  const StudentLessonDto({
    required this.id,
    required this.courseId,
    required this.title,
    required this.sortOrder,
    required this.version,
    required this.status,
    required this.resolvedLexemes,
    required this.resolvedMedia,
    required this.sections,
  });

  factory StudentLessonDto.fromJson(Map<String, dynamic> json) {
    final lexRaw = json['resolved_lexemes'] as Map<String, dynamic>? ?? {};
    final mediaRaw = json['resolved_media'] as Map<String, dynamic>? ?? {};
    return StudentLessonDto(
      id: json['id'] as String,
      courseId: json['course_id'] as String,
      title: json['title'] as String,
      sortOrder: json['sort_order'] as int? ?? 0,
      version: json['version'] as int? ?? 1,
      status: json['status'] as String? ?? 'published',
      resolvedLexemes: lexRaw.map(
        (k, v) => MapEntry(k, ResolvedLexemeDto.fromJson(v as Map<String, dynamic>)),
      ),
      resolvedMedia: mediaRaw.map(
        (k, v) => MapEntry(k, ResolvedMediaDto.fromJson(v as Map<String, dynamic>)),
      ),
      sections: (json['sections'] as List<dynamic>? ?? [])
          .map((e) => LessonSectionDto.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

class VocabularyEntryDto {
  final String id;
  final String lemma;
  final String? partOfSpeech;
  final double mastery;

  const VocabularyEntryDto({
    required this.id,
    required this.lemma,
    this.partOfSpeech,
    required this.mastery,
  });

  factory VocabularyEntryDto.fromJson(Map<String, dynamic> json) {
    final lex = json['lexeme'] as Map<String, dynamic>;
    return VocabularyEntryDto(
      id: lex['id'] as String,
      lemma: lex['lemma'] as String,
      partOfSpeech: lex['part_of_speech'] as String?,
      mastery: (json['mastery'] as num?)?.toDouble() ?? 0,
    );
  }
}
