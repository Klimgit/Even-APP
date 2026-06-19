import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/content_dto.dart';
import 'package:online_cource_app/api/repositories/content_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/features/teacher/courses/course_manage_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Teacher course admin panel — full course management.
class CourseAdminScreen extends StatefulWidget {
  const CourseAdminScreen({super.key});

  @override
  State<CourseAdminScreen> createState() => _CourseAdminScreenState();
}

class _CourseAdminScreenState extends State<CourseAdminScreen> {
  final _content = Get.find<ContentRepository>();
  final _search = TextEditingController();
  String? _languageCode;
  String _filter = 'all';
  List<CourseDto> _courses = [];
  bool _loading = true;
  String? _error;

  static const _filterOptions = [
    ('Все', 'all'),
    ('Опубликованные', 'published'),
    ('Черновики', 'draft'),
    ('Публичные', 'public'),
    ('По инвайту', 'invite'),
  ];

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      _courses = await _content.listTeacherCourses(query: _search.text.trim());
    } catch (e) {
      _error = e.toString();
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  List<CourseDto> get _filtered {
    var list = _courses;
    switch (_filter) {
      case 'published':
        list = list.where((c) => c.isPublished).toList();
        break;
      case 'draft':
        list = list.where((c) => !c.isPublished).toList();
        break;
      case 'public':
        list = list.where((c) => c.isPublic).toList();
        break;
      case 'invite':
        list = list.where((c) => !c.isPublic).toList();
        break;
    }
    if (_languageCode != null) {
      final lang = Get.find<LanguagesController>().byCode(_languageCode!);
      if (lang != null) {
        list = list.where((c) => c.targetLanguageId == lang.id).toList();
      }
    }
    return list;
  }

  Future<void> _createCourse() async {
    final langsCtrl = Get.find<LanguagesController>();
    if (langsCtrl.languages.isEmpty) {
      await langsCtrl.load();
    }
    final langs = langsCtrl.languages.toList();
    if (langs.isEmpty) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              langsCtrl.error.value != null
                  ? 'Не удалось загрузить языки. Проверьте, что сервер запущен.'
                  : 'Список языков пуст.',
            ),
          ),
        );
      }
      return;
    }

    final titleCtrl = TextEditingController();
    var targetId = langs.first.id;
    var uiId = langs.first.id;
    var visibility = 'invite_only';

    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDlg) => AlertDialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          title: const Text('Новый курс'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: titleCtrl,
                  autofocus: true,
                  decoration: const InputDecoration(hintText: 'Название курса'),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  value: targetId,
                  decoration: const InputDecoration(labelText: 'Язык курса'),
                  items: langs
                      .map((l) => DropdownMenuItem(value: l.id, child: Text(l.nativeName)))
                      .toList(),
                  onChanged: (v) => setDlg(() => targetId = v ?? targetId),
                ),
                DropdownButtonFormField<String>(
                  value: uiId,
                  decoration: const InputDecoration(labelText: 'Язык интерфейса'),
                  items: langs
                      .map((l) => DropdownMenuItem(value: l.id, child: Text(l.nativeName)))
                      .toList(),
                  onChanged: (v) => setDlg(() => uiId = v ?? uiId),
                ),
                DropdownButtonFormField<String>(
                  value: visibility,
                  decoration: const InputDecoration(labelText: 'Видимость'),
                  items: const [
                    DropdownMenuItem(value: 'invite_only', child: Text('По инвайту')),
                    DropdownMenuItem(value: 'public', child: Text('Публичный')),
                  ],
                  onChanged: (v) => setDlg(() => visibility = v ?? visibility),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
            ElevatedButton(
              onPressed: () => Navigator.pop(ctx, titleCtrl.text.trim().isNotEmpty),
              child: const Text('Создать'),
            ),
          ],
        ),
      ),
    );
    final title = titleCtrl.text.trim();
    titleCtrl.dispose();
    if (ok != true || !mounted || title.isEmpty) return;

    try {
      await _content.createCourse(
        title: title,
        targetLanguageId: targetId,
        uiLanguageId: uiId,
        visibility: visibility,
      );
      await _load();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Курс создан')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка: $e')),
        );
      }
    }
  }

  Future<void> _showInvite(CourseDto course) async {
    final code = await _content.getInviteCode(course.id);
    if (!mounted) return;
    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Инвайт: ${course.title}'),
        content: SelectableText(code.inviteCode, style: const TextStyle(fontSize: 22, letterSpacing: 2)),
        actions: [
          TextButton(
            onPressed: () {
              Clipboard.setData(ClipboardData(text: code.inviteCode));
              ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Скопировано')));
            },
            child: const Text('Копировать'),
          ),
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Закрыть')),
        ],
      ),
    );
  }

  Future<void> _toggleVisibility(CourseDto course) async {
    await _content.patchCourse(
      course.id,
      visibility: course.isPublic ? 'invite_only' : 'public',
    );
    await _load();
  }

  Future<void> _publish(CourseDto course) async {
    await _content.publishCourse(course.id);
    await _load();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ShellHeader(
          title: 'Мои курсы',
          subtitle: 'Создавайте и управляйте учебными материалами',
          trailing: FilledButton.icon(
            onPressed: _createCourse,
            style: FilledButton.styleFrom(
              backgroundColor: Colors.white,
              foregroundColor: AppTheme.primaryColor,
            ),
            icon: const Icon(Icons.add),
            label: const Text('Создать'),
          ),
        ),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                SearchFilterBar(
                  controller: _search,
                  hint: 'Поиск курсов…',
                  onChanged: _load,
                  trailing: LanguageFilterChip(
                    selectedCode: _languageCode,
                    onChanged: (v) => setState(() => _languageCode = v),
                  ),
                ),
                const SizedBox(height: 12),
                FilterChipsRow(
                  options: _filterOptions,
                  selected: _filter,
                  onSelected: (v) => setState(() => _filter = v),
                ),
                const SizedBox(height: 16),
                Expanded(child: _body()),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _body() {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      return EmptyState(
        icon: Icons.error_outline,
        message: 'Не удалось загрузить курсы',
        actionLabel: 'Повторить',
        onAction: _load,
      );
    }
    final list = _filtered;
    if (list.isEmpty) {
      return EmptyState(
        icon: Icons.school_outlined,
        message: _courses.isEmpty ? 'Создайте первый курс' : 'Нет курсов по фильтру',
        actionLabel: _courses.isEmpty ? 'Создать курс' : null,
        onAction: _courses.isEmpty ? _createCourse : null,
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.only(bottom: 16),
      itemCount: list.length,
      itemBuilder: (ctx, i) {
        final c = list[i];
        final langs = Get.find<LanguagesController>();
        final langName = langs.byId(c.targetLanguageId)?.nativeName ?? '—';
        final status = c.isPublished ? 'Опубликован' : 'Черновик';
        final access = c.isPublic ? 'Публичный' : 'По инвайту';
        return CourseCard(
          title: c.title,
          subtitle: '$langName · $status · $access',
          leadingIcon: c.isPublic ? Icons.public : Icons.lock_outline,
          leadingColor: c.isPublished ? AppTheme.accentColor : AppTheme.secondaryTextColor,
          onTap: () => Get.to(() => CourseManageScreen(courseId: c.id, title: c.title)),
          trailing: PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert),
            onSelected: (action) async {
              switch (action) {
                case 'edit':
                  Get.to(() => CourseManageScreen(courseId: c.id, title: c.title));
                case 'invite':
                  await _showInvite(c);
                case 'visibility':
                  await _toggleVisibility(c);
                case 'publish':
                  await _publish(c);
              }
            },
            itemBuilder: (ctx) => [
              const PopupMenuItem(value: 'edit', child: Text('Редактировать')),
              if (!c.isPublic)
                const PopupMenuItem(value: 'invite', child: Text('Инвайт-код')),
              PopupMenuItem(
                value: 'visibility',
                child: Text(c.isPublic ? 'Сделать по инвайту' : 'Сделать публичным'),
              ),
              if (!c.isPublished)
                const PopupMenuItem(value: 'publish', child: Text('Опубликовать')),
            ],
          ),
        );
      },
    );
  }
}
