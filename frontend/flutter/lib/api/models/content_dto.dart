class CourseDto {
  final String id;
  final String title;
  final String targetLanguageId;
  final String uiLanguageId;
  final String ownerId;
  final bool isPublished;
  final String visibility;
  final String? inviteCode;

  const CourseDto({
    required this.id,
    required this.title,
    required this.targetLanguageId,
    required this.uiLanguageId,
    required this.ownerId,
    required this.isPublished,
    required this.visibility,
    this.inviteCode,
  });

  bool get isPublic => visibility == 'public';

  factory CourseDto.fromJson(Map<String, dynamic> json) => CourseDto(
        id: json['id'] as String,
        title: json['title'] as String,
        targetLanguageId: json['target_language_id'] as String,
        uiLanguageId: json['ui_language_id'] as String,
        ownerId: json['owner_id'] as String,
        isPublished: json['is_published'] as bool? ?? false,
        visibility: json['visibility'] as String? ?? 'invite_only',
        inviteCode: json['invite_code'] as String?,
      );
}

class CourseModuleDto {
  final String id;
  final String courseId;
  final String title;
  final int sortOrder;

  const CourseModuleDto({
    required this.id,
    required this.courseId,
    required this.title,
    required this.sortOrder,
  });

  factory CourseModuleDto.fromJson(Map<String, dynamic> json) => CourseModuleDto(
        id: json['id'] as String,
        courseId: json['course_id'] as String,
        title: json['title'] as String,
        sortOrder: json['sort_order'] as int? ?? 0,
      );
}

class LessonSummaryDto {
  final String id;
  final String courseId;
  final String moduleId;
  final String title;
  final int sortOrder;
  final int version;
  final String status;

  const LessonSummaryDto({
    required this.id,
    required this.courseId,
    required this.moduleId,
    required this.title,
    required this.sortOrder,
    required this.version,
    required this.status,
  });

  factory LessonSummaryDto.fromJson(Map<String, dynamic> json) => LessonSummaryDto(
        id: json['id'] as String,
        courseId: json['course_id'] as String,
        moduleId: json['module_id'] as String,
        title: json['title'] as String,
        sortOrder: json['sort_order'] as int? ?? 0,
        version: json['version'] as int? ?? 1,
        status: json['status'] as String? ?? 'draft',
      );
}

class LessonBlockDto {
  final String id;
  final String? sectionId;
  final int sortOrder;
  final String? displayLabel;
  final String? title;
  final String blockType;
  final Map<String, dynamic> config;
  final bool isHomework;
  final bool isGradable;

  const LessonBlockDto({
    required this.id,
    this.sectionId,
    required this.sortOrder,
    this.displayLabel,
    this.title,
    required this.blockType,
    required this.config,
    required this.isHomework,
    required this.isGradable,
  });

  factory LessonBlockDto.fromJson(Map<String, dynamic> json) => LessonBlockDto(
        id: json['id'] as String,
        sectionId: json['section_id'] as String?,
        sortOrder: json['sort_order'] as int? ?? 0,
        displayLabel: json['display_label'] as String?,
        title: json['title'] as String?,
        blockType: json['block_type'] as String,
        config: Map<String, dynamic>.from(json['config'] as Map? ?? {}),
        isHomework: json['is_homework'] as bool? ?? false,
        isGradable: json['is_gradable'] as bool? ?? false,
      );
}

class LessonSectionDto {
  final String id;
  final String title;
  final int sortOrder;
  final String sectionKind;
  final List<LessonBlockDto> blocks;

  const LessonSectionDto({
    required this.id,
    required this.title,
    required this.sortOrder,
    required this.sectionKind,
    required this.blocks,
  });

  factory LessonSectionDto.fromJson(Map<String, dynamic> json) => LessonSectionDto(
        id: json['id'] as String,
        title: json['title'] as String,
        sortOrder: json['sort_order'] as int? ?? 0,
        sectionKind: json['section_kind'] as String? ?? 'content',
        blocks: (json['blocks'] as List<dynamic>? ?? [])
            .map((e) => LessonBlockDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class LessonDto {
  final String id;
  final String courseId;
  final String title;
  final int sortOrder;
  final int version;
  final String status;
  final List<LessonSectionDto> sections;

  const LessonDto({
    required this.id,
    required this.courseId,
    required this.title,
    required this.sortOrder,
    required this.version,
    required this.status,
    required this.sections,
  });

  factory LessonDto.fromJson(Map<String, dynamic> json) => LessonDto(
        id: json['id'] as String,
        courseId: json['course_id'] as String,
        title: json['title'] as String,
        sortOrder: json['sort_order'] as int? ?? 0,
        version: json['version'] as int? ?? 1,
        status: json['status'] as String? ?? 'draft',
        sections: (json['sections'] as List<dynamic>? ?? [])
            .map((e) => LessonSectionDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class BlockTypeInfoDto {
  final String blockType;
  final String title;
  final String? description;
  final bool isGradable;
  final bool isFavorite;

  const BlockTypeInfoDto({
    required this.blockType,
    required this.title,
    this.description,
    required this.isGradable,
    required this.isFavorite,
  });

  factory BlockTypeInfoDto.fromJson(Map<String, dynamic> json) => BlockTypeInfoDto(
        blockType: json['block_type'] as String,
        title: json['title'] as String,
        description: json['description'] as String?,
        isGradable: json['is_gradable'] as bool? ?? false,
        isFavorite: json['is_favorite'] as bool? ?? false,
      );
}

class CourseAnalyticsDto {
  final int studentCount;
  final int lessonCount;
  final int moduleCount;
  final double avgCompletionPercent;

  const CourseAnalyticsDto({
    required this.studentCount,
    required this.lessonCount,
    required this.moduleCount,
    required this.avgCompletionPercent,
  });

  factory CourseAnalyticsDto.fromJson(Map<String, dynamic> json) => CourseAnalyticsDto(
        studentCount: json['student_count'] as int? ?? 0,
        lessonCount: json['lesson_count'] as int? ?? 0,
        moduleCount: json['module_count'] as int? ?? 0,
        avgCompletionPercent: (json['avg_completion_percent'] as num?)?.toDouble() ?? 0,
      );
}

class StudentDto {
  final String id;
  final String email;
  final String? displayName;
  final List<StudentEnrolledCourseDto> enrolledCourses;

  const StudentDto({
    required this.id,
    required this.email,
    this.displayName,
    required this.enrolledCourses,
  });

  factory StudentDto.fromJson(Map<String, dynamic> json) => StudentDto(
        id: json['id'] as String,
        email: json['email'] as String,
        displayName: json['display_name'] as String?,
        enrolledCourses: (json['enrolled_courses'] as List<dynamic>? ?? [])
            .map((e) => StudentEnrolledCourseDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class StudentEnrolledCourseDto {
  final String courseId;
  final String title;
  final double progressPercent;

  const StudentEnrolledCourseDto({
    required this.courseId,
    required this.title,
    required this.progressPercent,
  });

  factory StudentEnrolledCourseDto.fromJson(Map<String, dynamic> json) =>
      StudentEnrolledCourseDto(
        courseId: json['course_id'] as String,
        title: json['title'] as String,
        progressPercent: (json['progress_percent'] as num?)?.toDouble() ?? 0,
      );
}

class InviteCodeDto {
  final String inviteCode;

  const InviteCodeDto({required this.inviteCode});

  factory InviteCodeDto.fromJson(Map<String, dynamic> json) =>
      InviteCodeDto(inviteCode: json['invite_code'] as String);
}
