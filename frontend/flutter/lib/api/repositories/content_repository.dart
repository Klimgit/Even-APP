import 'package:online_cource_app/api/base_repository.dart';
import 'package:online_cource_app/api/models/content_dto.dart';

class ContentRepository extends BaseRepository {
  ContentRepository(super.client);

  Future<List<CourseDto>> listTeacherCourses({String? query}) async {
    final list = await getList(
      '/teacher/courses',
      query: query != null && query.isNotEmpty ? {'q': query} : null,
    );
    return list.map((e) => CourseDto.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<CourseDto>> listPlatformCourses({String? query}) async {
    final data = await getJson(
      '/platform/courses',
      query: {
        if (query != null && query.isNotEmpty) 'q': query,
        'page': 1,
        'limit': 100,
      },
    );
    final items = data['items'] as List<dynamic>? ?? [];
    return items.map((e) => CourseDto.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<CourseDto> createCourse({
    required String title,
    required String targetLanguageId,
    required String uiLanguageId,
    String visibility = 'invite_only',
  }) async {
    final data = await postJson('/teacher/courses', {
      'title': title,
      'target_language_id': targetLanguageId,
      'ui_language_id': uiLanguageId,
      'visibility': visibility,
    });
    return CourseDto.fromJson(data);
  }

  Future<CourseDto> patchCourse(
    String courseId, {
    String? title,
    String? visibility,
  }) async {
    final body = <String, dynamic>{};
    if (title != null) body['title'] = title;
    if (visibility != null) body['visibility'] = visibility;
    return CourseDto.fromJson(await patchJson('/teacher/courses/$courseId', body));
  }

  Future<CourseDto> getCourse(String courseId) async {
    return CourseDto.fromJson(await getJson('/teacher/courses/$courseId'));
  }

  Future<CourseDto> publishCourse(String courseId) async {
    final data = await postJson('/teacher/courses/$courseId/publish', {});
    return CourseDto.fromJson(data);
  }

  Future<InviteCodeDto> getInviteCode(String courseId) async {
    return InviteCodeDto.fromJson(await getJson('/teacher/courses/$courseId/invite-code'));
  }

  Future<InviteCodeDto> regenerateInviteCode(String courseId) async {
    return InviteCodeDto.fromJson(
      await postJson('/teacher/courses/$courseId/invite-code/regenerate', {}),
    );
  }

  Future<CourseAnalyticsDto> getAnalytics(String courseId) async {
    return CourseAnalyticsDto.fromJson(
      await getJson('/teacher/courses/$courseId/analytics'),
    );
  }

  Future<List<StudentDto>> listStudents(String courseId) async {
    final list = await getList('/teacher/courses/$courseId/students');
    return list.map((e) => StudentDto.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<StudentDto> enrollStudentByEmail({
    required String email,
    required String courseId,
  }) async {
    return StudentDto.fromJson(await postJson('/teacher/students', {
      'email': email,
      'course_id': courseId,
    }));
  }

  Future<List<CourseModuleDto>> listModules(String courseId) async {
    final list = await getList('/teacher/courses/$courseId/modules');
    return list.map((e) => CourseModuleDto.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<CourseModuleDto> createModule(String courseId, String title) async {
    return CourseModuleDto.fromJson(
      await postJson('/teacher/courses/$courseId/modules', {'title': title}),
    );
  }

  Future<List<LessonSummaryDto>> listLessons(String courseId, {String? moduleId}) async {
    final list = await getList('/teacher/courses/$courseId/lessons');
    final lessons = list
        .map((e) => LessonSummaryDto.fromJson(e as Map<String, dynamic>))
        .toList();
    if (moduleId != null) {
      return lessons.where((l) => l.moduleId == moduleId).toList();
    }
    return lessons;
  }

  Future<LessonSummaryDto> createLesson({
    required String moduleId,
    required String title,
    int? sortOrder,
  }) async {
    return LessonSummaryDto.fromJson(
      await postJson('/teacher/modules/$moduleId/lessons', {
        'title': title,
        if (sortOrder != null) 'sort_order': sortOrder,
      }),
    );
  }

  Future<LessonDto> getLesson(String lessonId) async {
    return LessonDto.fromJson(await getJson('/teacher/lessons/$lessonId'));
  }

  Future<LessonDto> publishLesson(String lessonId) async {
    return LessonDto.fromJson(await postJson('/teacher/lessons/$lessonId/publish', {}));
  }

  Future<List<BlockTypeInfoDto>> listBlockTypes() async {
    final list = await getList('/teacher/block-types');
    final types = <BlockTypeInfoDto>[];
    for (final cat in list) {
      final map = cat as Map<String, dynamic>;
      for (final t in map['types'] as List<dynamic>? ?? []) {
        types.add(BlockTypeInfoDto.fromJson(t as Map<String, dynamic>));
      }
    }
    return types;
  }

  Future<LessonSectionDto> createSection({
    required String lessonId,
    required String title,
    String sectionKind = 'content',
  }) async {
    return LessonSectionDto.fromJson(
      await postJson('/teacher/lessons/$lessonId/sections', {
        'title': title,
        'section_kind': sectionKind,
      }),
    );
  }

  Future<LessonBlockDto> createBlock({
    required String lessonId,
    required String sectionId,
    required String blockType,
    required Map<String, dynamic> config,
    required int sortOrder,
    String? title,
  }) async {
    return LessonBlockDto.fromJson(
      await postJson('/teacher/lessons/$lessonId/blocks', {
        'section_id': sectionId,
        'sort_order': sortOrder,
        'block_type': blockType,
        'config': config,
        if (title != null) 'title': title,
      }),
    );
  }

  Future<LessonBlockDto> patchBlock({
    required String blockId,
    Map<String, dynamic>? config,
    String? title,
  }) async {
    final body = <String, dynamic>{};
    if (config != null) body['config'] = config;
    if (title != null) body['title'] = title;
    return LessonBlockDto.fromJson(await patchJson('/teacher/blocks/$blockId', body));
  }

  Future<void> deleteBlock(String blockId) => delete('/teacher/blocks/$blockId');
}
