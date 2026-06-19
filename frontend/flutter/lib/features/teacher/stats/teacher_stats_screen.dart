import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/content_dto.dart';
import 'package:online_cource_app/api/repositories/content_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/features/teacher/stats/course_analytics_screen.dart';

class TeacherStatsScreen extends StatefulWidget {
  const TeacherStatsScreen({super.key});

  @override
  State<TeacherStatsScreen> createState() => _TeacherStatsScreenState();
}

class _TeacherStatsScreenState extends State<TeacherStatsScreen> {
  final _content = Get.find<ContentRepository>();
  List<CourseDto> _courses = [];

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    _courses = await _content.listTeacherCourses();
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    if (_courses.isEmpty) {
      return const EmptyState(icon: Icons.bar_chart, message: 'Нет курсов для статистики');
    }
    return ListView.separated(
      padding: const EdgeInsets.all(24),
      itemCount: _courses.length,
      separatorBuilder: (_, __) => const Divider(),
      itemBuilder: (ctx, i) {
        final c = _courses[i];
        return ListTile(
          title: Text(c.title),
          subtitle: Text(c.isPublished ? 'Опубликован' : 'Черновик'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Get.to(() => CourseAnalyticsScreen(courseId: c.id, title: c.title)),
        );
      },
    );
  }
}
