import 'package:flutter/material.dart';
import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/Courses/student_course_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Student courses list. Shows courses created by teachers (the `classes`
/// collection), not the old template catalogue. Tapping a course opens its
/// lessons, which the student plays through the shared lesson runner.
class CourseListPage extends StatelessWidget {
  const CourseListPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        title: const Text('Courses'),
        backgroundColor: AppTheme.backgroundColor,
        elevation: 0,
        foregroundColor: AppTheme.textColor,
      ),
      body: StreamBuilder<QuerySnapshot<Map<String, dynamic>>>(
        stream: FirebaseFirestore.instance.collection('classes').snapshots(),
        builder: (context, snapshot) {
          if (!snapshot.hasData) {
            return const Center(child: CircularProgressIndicator());
          }

          final courses = snapshot.data!.docs;
          if (courses.isEmpty) {
            return Center(
              child: Text('No courses available yet.',
                  style: TextStyle(color: AppTheme.secondaryTextColor)),
            );
          }

          // Same minimalist card grid as Home → Popular Courses.
          return Padding(
            padding: const EdgeInsets.all(20),
            child: LayoutBuilder(
              builder: (context, constraints) {
                final crossAxisCount = constraints.maxWidth >= 900 ? 3 : 2;
                return GridView.builder(
                  gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: crossAxisCount,
                    childAspectRatio: 2.6,
                    crossAxisSpacing: 16,
                    mainAxisSpacing: 16,
                  ),
                  itemCount: courses.length,
                  itemBuilder: (context, index) {
                    final doc = courses[index];
                    return CourseCard(id: doc.id, data: doc.data());
                  },
                );
              },
            ),
          );
        },
      ),
    );
  }
}

class CourseCard extends StatelessWidget {
  final String id;
  final Map<String, dynamic> data;

  const CourseCard({super.key, required this.id, required this.data});

  @override
  Widget build(BuildContext context) {
    final name = (data['name'] as String?) ?? 'Untitled';
    final colorValue = (data['coverColor'] as int?) ?? 0xFF9BE8B4;
    return GestureDetector(
      onTap: () => Get.to(
          () => StudentCourseScreen(courseId: id, courseName: name)),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppTheme.dividerColor),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Text(
                name,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.bold),
                maxLines: 3,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 12),
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: Color(colorValue),
                borderRadius: BorderRadius.circular(8),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
