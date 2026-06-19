import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:flutter/material.dart';
import 'package:get/get.dart';

import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/exercise_catalog.dart';
import 'package:online_cource_app/exercises/lesson_runner.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Student view of a teacher-created course: its published lessons, each of
/// which can be played through the shared [LessonRunner].
class StudentCourseScreen extends StatelessWidget {
  final String courseId;
  final String courseName;

  const StudentCourseScreen({
    super.key,
    required this.courseId,
    required this.courseName,
  });

  CollectionReference<Map<String, dynamic>> get _lessons => FirebaseFirestore
      .instance
      .collection('classes')
      .doc(courseId)
      .collection('lessons');

  void _openLesson(Map<String, dynamic> data) {
    final rawSteps = (data['steps'] as List?) ?? const [];
    final steps = <ExerciseData>[];
    for (final e in rawSteps) {
      try {
        steps.add(exerciseFromJson(Map<String, dynamic>.from(e as Map)));
      } catch (_) {
        // Skip unknown/broken steps rather than failing the whole lesson.
      }
    }
    if (steps.isEmpty) {
      Get.snackbar('Empty lesson', 'This lesson has no exercises yet.',
          snackPosition: SnackPosition.BOTTOM);
      return;
    }
    Get.to(() => LessonRunner(
          title: data['title'] as String? ?? 'Lesson',
          exercises: steps,
        ));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        backgroundColor: AppTheme.backgroundColor,
        elevation: 0,
        foregroundColor: AppTheme.textColor,
        title: Text(courseName),
      ),
      body: StreamBuilder<QuerySnapshot<Map<String, dynamic>>>(
        stream: _lessons.orderBy('createdAt').snapshots(),
        builder: (context, snapshot) {
          if (!snapshot.hasData) {
            return const Center(child: CircularProgressIndicator());
          }
          final docs = snapshot.data!.docs;
          if (docs.isEmpty) {
            return Center(
              child: Text('No lessons in this course yet.',
                  style: TextStyle(color: AppTheme.secondaryTextColor)),
            );
          }
          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: docs.length,
            itemBuilder: (context, index) {
              final data = docs[index].data();
              final title = data['title'] as String? ?? 'Untitled';
              final stepCount = (data['steps'] as List?)?.length ?? 0;
              return Card(
                elevation: 0,
                margin: const EdgeInsets.only(bottom: 10),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: BorderSide(color: AppTheme.dividerColor),
                ),
                child: ListTile(
                  leading: CircleAvatar(
                    backgroundColor: AppTheme.accentColor.withOpacity(0.12),
                    child: Text('${index + 1}',
                        style: TextStyle(
                            color: AppTheme.accentColor,
                            fontWeight: FontWeight.bold)),
                  ),
                  title: Text(title,
                      style: const TextStyle(fontWeight: FontWeight.w600)),
                  subtitle:
                      Text('$stepCount step${stepCount == 1 ? '' : 's'}'),
                  trailing: Icon(Icons.play_circle_outline,
                      color: AppTheme.accentColor),
                  onTap: stepCount == 0 ? null : () => _openLesson(data),
                ),
              );
            },
          );
        },
      ),
    );
  }
}
