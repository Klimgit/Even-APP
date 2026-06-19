import 'package:online_cource_app/api/base_repository.dart';
import 'package:online_cource_app/api/models/learning_dto.dart';

class LearningRepository extends BaseRepository {
  LearningRepository(super.client);

  Future<List<CourseListItemDto>> listEnrolledCourses() async {
    final list = await getList('/courses');
    return list.map((e) => CourseListItemDto.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<PublicCourseListItemDto>> listPublicCourses({String? query}) async {
    final list = await getList(
      '/courses/public',
      query: query != null && query.isNotEmpty ? {'q': query} : null,
    );
    return list
        .map((e) => PublicCourseListItemDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> enrollPublicCourse(String courseId) async {
    await postJson('/courses/$courseId/enroll', {});
  }

  Future<void> joinByInviteCode(String inviteCode) async {
    await postJson('/courses/join', {'invite_code': inviteCode});
  }

  Future<CourseOutlineDto> getOutline(String courseId) async {
    return CourseOutlineDto.fromJson(await getJson('/courses/$courseId/outline'));
  }

  Future<StudentLessonDto> getLesson(String lessonId) async {
    return StudentLessonDto.fromJson(await getJson('/lessons/$lessonId'));
  }

  Future<LessonFlowDto> getLessonFlow(String lessonId) async {
    return LessonFlowDto.fromJson(await getJson('/lessons/$lessonId/flow'));
  }

  Future<BlockAttemptResultDto> submitBlockAttempt({
    required String blockId,
    required Map<String, dynamic> response,
    int subItemIndex = 0,
    int timeSpentSeconds = 0,
  }) async {
    return BlockAttemptResultDto.fromJson(
      await postJson('/progress/blocks/$blockId/attempt', {
        'response': response,
        'sub_item_index': subItemIndex,
        'time_spent_seconds': timeSpentSeconds,
      }),
    );
  }

  Future<ProgressSummaryDto> getProgressSummary() async {
    return ProgressSummaryDto.fromJson(await getJson('/progress/summary'));
  }

  Future<List<VocabularyEntryDto>> listDictionary({String? query}) async {
    final list = await getList(
      '/dictionary',
      query: query != null && query.isNotEmpty ? {'q': query} : null,
    );
    return list
        .map((e) => VocabularyEntryDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
