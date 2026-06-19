import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/repositories/content_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';

class CourseAnalyticsScreen extends StatelessWidget {
  final String courseId;
  final String title;

  const CourseAnalyticsScreen({super.key, required this.courseId, required this.title});

  @override
  Widget build(BuildContext context) {
    final content = Get.find<ContentRepository>();
    return Scaffold(
      appBar: AppBar(title: Text('Статистика: $title')),
      body: FutureBuilder(
        future: content.getAnalytics(courseId),
        builder: (context, snap) {
          if (snap.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snap.hasError) return Center(child: Text('${snap.error}'));
          final a = snap.data!;
          return Padding(
            padding: const EdgeInsets.all(16),
            child: GridView.count(
              crossAxisCount: 2,
              crossAxisSpacing: 16,
              mainAxisSpacing: 16,
              childAspectRatio: 1.4,
              children: [
                StatCard(label: 'Учеников', value: '${a.studentCount}', icon: Icons.people),
                StatCard(label: 'Уроков', value: '${a.lessonCount}', icon: Icons.menu_book),
                StatCard(label: 'Модулей', value: '${a.moduleCount}', icon: Icons.view_module),
                StatCard(
                  label: 'Средний прогресс',
                  value: '${a.avgCompletionPercent.toStringAsFixed(0)}%',
                  icon: Icons.trending_up,
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}
