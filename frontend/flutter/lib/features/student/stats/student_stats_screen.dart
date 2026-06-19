import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/repositories/learning_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/theme/app_theme.dart';

class StudentStatsScreen extends StatelessWidget {
  const StudentStatsScreen({super.key});

  String _formatDuration(int seconds) {
    final h = seconds ~/ 3600;
    final m = (seconds % 3600) ~/ 60;
    if (h > 0) return '${h}ч ${m}м';
    return '${m} мин';
  }

  @override
  Widget build(BuildContext context) {
    final learning = Get.find<LearningRepository>();
    return SafeArea(
      child: FutureBuilder(
        future: learning.getProgressSummary(),
        builder: (context, snap) {
          if (snap.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snap.hasError) return Center(child: Text('${snap.error}'));
          final s = snap.data!;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text('Моя статистика', style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height: 16),
              GridView.count(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                crossAxisCount: 2,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 1.3,
                children: [
                  StatCard(
                    label: 'Курсов',
                    value: '${s.enrolledCourses}',
                    icon: Icons.school,
                  ),
                  StatCard(
                    label: 'Уроков завершено',
                    value: '${s.completedLessons}',
                    icon: Icons.check_circle_outline,
                  ),
                  StatCard(
                    label: 'Средний балл',
                    value: '${(s.averageScore * 100).toStringAsFixed(0)}%',
                    icon: Icons.star_outline,
                  ),
                  StatCard(
                    label: 'Время обучения',
                    value: _formatDuration(s.timeSpentSeconds),
                    icon: Icons.schedule,
                  ),
                  StatCard(
                    label: 'Слов в словаре',
                    value: '${s.dictionaryWords}',
                    icon: Icons.menu_book,
                  ),
                  StatCard(
                    label: 'На повторение',
                    value: '${s.reviewDue}',
                    icon: Icons.replay,
                  ),
                ],
              ),
              const SizedBox(height: 24),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Прогресс', style: Theme.of(context).textTheme.titleMedium),
                      const SizedBox(height: 12),
                      _progressRow('Блоков выполнено', s.completedBlocks, s.completedBlocks + s.inProgressLessons * 3),
                      const SizedBox(height: 8),
                      _progressRow('Уроков в процессе', s.inProgressLessons, s.completedLessons + s.inProgressLessons),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _progressRow(String label, int value, int total) {
    final pct = total > 0 ? value / total : 0.0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(label),
            const Spacer(),
            Text('$value', style: TextStyle(color: AppTheme.secondaryTextColor)),
          ],
        ),
        const SizedBox(height: 4),
        LinearProgressIndicator(value: pct.clamp(0.0, 1.0)),
      ],
    );
  }
}
