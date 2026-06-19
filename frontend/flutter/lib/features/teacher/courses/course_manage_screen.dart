import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/content_dto.dart';
import 'package:online_cource_app/api/repositories/content_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/features/teacher/courses/lesson_editor_screen.dart';
import 'package:online_cource_app/features/teacher/stats/course_analytics_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

class CourseManageScreen extends StatefulWidget {
  final String courseId;
  final String title;

  const CourseManageScreen({super.key, required this.courseId, required this.title});

  @override
  State<CourseManageScreen> createState() => _CourseManageScreenState();
}

class _CourseManageScreenState extends State<CourseManageScreen> {
  final _content = Get.find<ContentRepository>();
  CourseDto? _course;
  List<CourseModuleDto> _modules = [];
  final Map<String, List<LessonSummaryDto>> _lessons = {};
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      _course = await _content.getCourse(widget.courseId);
      _modules = await _content.listModules(widget.courseId);
      _lessons.clear();
      for (final m in _modules) {
        _lessons[m.id] = await _content.listLessons(widget.courseId, moduleId: m.id);
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _showInvite() async {
    final code = await _content.getInviteCode(widget.courseId);
    if (!mounted) return;
    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Код приглашения'),
        content: SelectableText(code.inviteCode, style: const TextStyle(fontSize: 24, letterSpacing: 2)),
        actions: [
          TextButton(
            onPressed: () {
              Clipboard.setData(ClipboardData(text: code.inviteCode));
              ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Скопировано')));
            },
            child: const Text('Копировать'),
          ),
          TextButton(
            onPressed: () async {
              final regen = await _content.regenerateInviteCode(widget.courseId);
              Navigator.pop(ctx);
              Clipboard.setData(ClipboardData(text: regen.inviteCode));
              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('Новый код: ${regen.inviteCode}')),
                );
              }
            },
            child: const Text('Перегенерировать'),
          ),
        ],
      ),
    );
  }

  Future<void> _addModule() async {
    final ctrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Новый модуль'),
        content: TextField(controller: ctrl, decoration: const InputDecoration(hintText: 'Название')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Создать')),
        ],
      ),
    );
    if (ok == true && ctrl.text.trim().isNotEmpty) {
      await _content.createModule(widget.courseId, ctrl.text.trim());
      await _load();
    }
  }

  Future<void> _addLesson(String moduleId) async {
    final ctrl = TextEditingController(text: 'Новый урок');
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Новый урок'),
        content: TextField(controller: ctrl),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Создать')),
        ],
      ),
    );
    if (ok == true) {
      final existing = _lessons[moduleId] ?? [];
      await _content.createLesson(
        moduleId: moduleId,
        title: ctrl.text.trim(),
        sortOrder: existing.length + 1,
      );
      await _load();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        actions: [
          IconButton(
            icon: const Icon(Icons.analytics_outlined),
            tooltip: 'Статистика',
            onPressed: () => Get.to(() => CourseAnalyticsScreen(courseId: widget.courseId, title: widget.title)),
          ),
          IconButton(icon: const Icon(Icons.vpn_key_outlined), tooltip: 'Инвайт', onPressed: _showInvite),
          if (_course != null && !_course!.isPublished)
            TextButton(
              onPressed: () async {
                await _content.publishCourse(widget.courseId);
                await _load();
              },
              child: const Text('Опубликовать'),
            ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (_course != null)
                  Wrap(
                    spacing: 8,
                    children: [
                      Chip(
                        label: Text(_course!.isPublic ? 'Публичный' : 'По инвайту'),
                        avatar: Icon(_course!.isPublic ? Icons.public : Icons.lock, size: 18),
                      ),
                      Chip(
                        label: Text(_course!.isPublished ? 'Опубликован' : 'Черновик'),
                        backgroundColor: _course!.isPublished
                            ? AppTheme.successColor.withOpacity(0.15)
                            : AppTheme.secondaryTextColor.withOpacity(0.15),
                      ),
                    ],
                  ),
                const SizedBox(height: 16),
                Row(
                  children: [
                    Text('Модули', style: Theme.of(context).textTheme.titleLarge),
                    const Spacer(),
                    TextButton.icon(onPressed: _addModule, icon: const Icon(Icons.add), label: const Text('Модуль')),
                  ],
                ),
                for (final module in _modules) ...[
                  ListTile(
                    title: Text(module.title, style: const TextStyle(fontWeight: FontWeight.w600)),
                    trailing: IconButton(
                      icon: const Icon(Icons.add),
                      onPressed: () => _addLesson(module.id),
                    ),
                  ),
                  ...(_lessons[module.id] ?? []).map(
                    (lesson) => ListTile(
                      contentPadding: const EdgeInsets.only(left: 32),
                      leading: Icon(
                        lesson.status == 'published' ? Icons.check_circle : Icons.edit_outlined,
                        color: lesson.status == 'published' ? AppTheme.successColor : null,
                      ),
                      title: Text(lesson.title),
                      subtitle: Text(lesson.status),
                      onTap: () async {
                        final langs = Get.find<LanguagesController>();
                        final lang = langs.byId(_course?.targetLanguageId ?? '');
                        await Get.to(
                          () => LessonEditorScreen(
                            lessonId: lesson.id,
                            title: lesson.title,
                            languageCode: lang?.code ?? 'evn',
                            languageId: lang?.id ?? _course!.targetLanguageId,
                          ),
                        );
                        await _load();
                      },
                    ),
                  ),
                ],
              ],
            ),
    );
  }
}
