import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/repositories/learning_repository.dart';
import 'package:online_cource_app/features/student/learning/lesson_player_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Stepik-like course outline with module/lesson tree and progress.
class CourseOutlineScreen extends StatefulWidget {
  final String courseId;
  final String title;
  final String targetLanguageCode;

  const CourseOutlineScreen({
    super.key,
    required this.courseId,
    required this.title,
    this.targetLanguageCode = 'evn',
  });

  @override
  State<CourseOutlineScreen> createState() => _CourseOutlineScreenState();
}

class _CourseOutlineScreenState extends State<CourseOutlineScreen> {
  final _learning = Get.find<LearningRepository>();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.title)),
      body: FutureBuilder(
        future: _learning.getOutline(widget.courseId),
        builder: (context, snap) {
          if (snap.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snap.hasError) return Center(child: Text('${snap.error}'));
          final outline = snap.data!;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              LinearProgressIndicator(value: outline.progressPercent / 100),
              const SizedBox(height: 8),
              Text('Прогресс: ${outline.progressPercent.toStringAsFixed(0)}%'),
              const SizedBox(height: 16),
              for (final module in outline.modules) ...[
                Text(module.title, style: Theme.of(context).textTheme.titleLarge),
                Text('${module.progressPercent.toStringAsFixed(0)}% модуля',
                    style: TextStyle(color: AppTheme.secondaryTextColor)),
                for (final lesson in module.lessons)
                  ListTile(
                    leading: CircularProgressIndicator(
                      value: lesson.progressPercent / 100,
                      strokeWidth: 3,
                    ),
                    title: Text(lesson.title),
                    subtitle: Text('${lesson.progressPercent.toStringAsFixed(0)}%'),
                    trailing: lesson.id == outline.currentLessonId
                        ? Chip(label: Text('Текущий', style: TextStyle(color: AppTheme.primaryColor)))
                        : const Icon(Icons.play_arrow),
                    onTap: () => Get.to(
                      () => LessonPlayerScreen(
                        lessonId: lesson.id,
                        title: lesson.title,
                        courseId: widget.courseId,
                        targetLanguageCode: widget.targetLanguageCode,
                      ),
                    ),
                  ),
                const Divider(),
              ],
            ],
          );
        },
      ),
    );
  }
}
