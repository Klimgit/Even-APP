import 'package:flutter/material.dart';
import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/Courses/course_details.dart';
import 'package:online_cource_app/Model/course_model.dart';
import 'package:online_cource_app/theme/app_theme.dart';

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
      body: StreamBuilder(
        stream: FirebaseFirestore.instance.collection('courses').snapshots(),
        builder: (context, AsyncSnapshot<QuerySnapshot> snapshot) {
          if (!snapshot.hasData) {
            return const Center(child: CircularProgressIndicator());
          }

          final courses = snapshot.data!.docs;
          if (courses.isEmpty) {
            return const Center(child: Text('No courses yet'));
          }

          // Same layout as Home → Popular Courses: 3 per row on desktop, 2 on
          // narrow screens; each card shows only the cover icon and the title.
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
                    final course = CourseModel.fromJson(
                        courses[index].data() as Map<String, dynamic>);
                    return CourseCard(course: course);
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
  final CourseModel course;
  const CourseCard({super.key, required this.course});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () => Get.to(() => CourseDetailsPage(course: course)),
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
                course.title,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.bold),
                maxLines: 3,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: Image.network(
                course.cover,
                width: 56,
                height: 56,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) => Container(
                  width: 56,
                  height: 56,
                  color: AppTheme.dividerColor,
                  child: Icon(Icons.image_not_supported,
                      color: AppTheme.secondaryTextColor, size: 24),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
