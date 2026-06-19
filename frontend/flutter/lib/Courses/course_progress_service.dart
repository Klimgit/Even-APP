import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:get/get.dart';

import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';

/// A student's progress in one teacher course.
class CourseProgress {
  final String courseId;
  final String name;
  final int coverColor;
  final int totalLessons;
  final List<String> completedLessons;

  const CourseProgress({
    required this.courseId,
    required this.name,
    required this.coverColor,
    required this.totalLessons,
    required this.completedLessons,
  });

  double get percent =>
      totalLessons == 0 ? 0 : (completedLessons.length / totalLessons).clamp(0, 1);

  factory CourseProgress.fromDoc(
      QueryDocumentSnapshot<Map<String, dynamic>> doc) {
    final data = doc.data();
    return CourseProgress(
      courseId: data['courseId'] as String? ?? doc.id,
      name: data['name'] as String? ?? 'Course',
      coverColor: data['coverColor'] as int? ?? 0xFF9BE8B4,
      totalLessons: data['totalLessons'] as int? ?? 0,
      completedLessons: ((data['completedLessons'] as List?) ?? const [])
          .map((e) => e as String)
          .toList(),
    );
  }
}

/// Tracks which teacher courses a student has started and their lesson
/// completion. Stored in Firestore under `users/{uid}/courseProgress`.
class CourseProgressService {
  static final FirebaseFirestore _db = FirebaseFirestore.instance;

  /// Current user id — from the REST session when [kUseApiAuth], else Firebase.
  static String? get _userId {
    if (kUseApiAuth) {
      return Get.isRegistered<ApiAuthController>()
          ? Get.find<ApiAuthController>().user.value?.id
          : null;
    }
    return FirebaseAuth.instance.currentUser?.uid;
  }

  static CollectionReference<Map<String, dynamic>>? _collection() {
    final uid = _userId;
    if (uid == null) return null;
    return _db.collection('users').doc(uid).collection('courseProgress');
  }

  /// Marks a course as started (records metadata; idempotent).
  static Future<void> markStarted({
    required String courseId,
    required String name,
    required int coverColor,
    required int totalLessons,
  }) async {
    final col = _collection();
    if (col == null) return;
    await col.doc(courseId).set({
      'courseId': courseId,
      'name': name,
      'coverColor': coverColor,
      'totalLessons': totalLessons,
      'lastAccessed': FieldValue.serverTimestamp(),
    }, SetOptions(merge: true));
  }

  static Future<void> markLessonCompleted({
    required String courseId,
    required String lessonId,
  }) async {
    final col = _collection();
    if (col == null) return;
    await col.doc(courseId).set({
      'completedLessons': FieldValue.arrayUnion([lessonId]),
      'lastAccessed': FieldValue.serverTimestamp(),
    }, SetOptions(merge: true));
  }

  /// Stream of started courses, most recently accessed first.
  static Stream<List<CourseProgress>> startedCoursesStream() {
    final col = _collection();
    if (col == null) return Stream.value(const []);
    return col
        .orderBy('lastAccessed', descending: true)
        .snapshots()
        .map((s) => s.docs.map(CourseProgress.fromDoc).toList());
  }
}
