import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/learning_dto.dart';
import 'package:online_cource_app/api/repositories/learning_repository.dart';
import 'package:online_cource_app/features/shared/profile_tab.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/features/student/courses/course_outline_screen.dart';
import 'package:online_cource_app/features/student/dictionary/student_dictionary_screen.dart';
import 'package:online_cource_app/features/student/stats/student_stats_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Student app shell with courses, stats, knowledge base, and profile.
class StudentShell extends StatefulWidget {
  const StudentShell({super.key});

  @override
  State<StudentShell> createState() => _StudentShellState();
}

class _StudentShellState extends State<StudentShell> {
  int _index = 0;

  @override
  void initState() {
    super.initState();
    if (!Get.isRegistered<LanguagesController>()) {
      Get.put(LanguagesController(), permanent: true);
    }
  }

  @override
  Widget build(BuildContext context) {
    final pages = const [
      StudentCoursesTab(),
      StudentStatsScreen(),
      StudentDictionaryScreen(),
      ProfileTab(),
    ];
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      body: pages[_index],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        indicatorColor: AppTheme.primaryColor.withValues(alpha: 0.12),
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.school_outlined),
            selectedIcon: Icon(Icons.school),
            label: 'Курсы',
          ),
          NavigationDestination(
            icon: Icon(Icons.bar_chart_outlined),
            selectedIcon: Icon(Icons.bar_chart),
            label: 'Статистика',
          ),
          NavigationDestination(
            icon: Icon(Icons.menu_book_outlined),
            selectedIcon: Icon(Icons.menu_book),
            label: 'База знаний',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline),
            selectedIcon: Icon(Icons.person),
            label: 'Профиль',
          ),
        ],
      ),
    );
  }
}

class StudentCoursesTab extends StatefulWidget {
  const StudentCoursesTab({super.key});

  @override
  State<StudentCoursesTab> createState() => _StudentCoursesTabState();
}

class _StudentCoursesTabState extends State<StudentCoursesTab>
    with SingleTickerProviderStateMixin {
  final _learning = Get.find<LearningRepository>();
  late TabController _tabs;
  final _search = TextEditingController();
  String? _languageCode;
  List<CourseListItemDto> _enrolled = [];
  List<PublicCourseListItemDto> _public = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 2, vsync: this);
    _tabs.addListener(() {
      if (!_tabs.indexIsChanging) _load();
    });
    _load();
  }

  @override
  void dispose() {
    _search.dispose();
    _tabs.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      _enrolled = await _learning.listEnrolledCourses();
      if (_tabs.index == 1) {
        _public = await _learning.listPublicCourses(query: _search.text.trim());
      } else {
        _public = await _learning.listPublicCourses();
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _applySearch() {
    if (_tabs.index == 1) {
      _load();
    } else {
      setState(() {});
    }
  }

  List<CourseListItemDto> get _filteredEnrolled {
    var list = _enrolled;
    final q = _search.text.trim().toLowerCase();
    if (q.isNotEmpty) {
      list = list.where((c) => c.title.toLowerCase().contains(q)).toList();
    }
    if (_languageCode != null) {
      list = list.where((c) => c.targetLanguage.code == _languageCode).toList();
    }
    return list;
  }

  List<PublicCourseListItemDto> get _filteredPublic {
    var list = _public;
    if (_languageCode != null) {
      list = list.where((c) => c.targetLanguage.code == _languageCode).toList();
    }
    return list;
  }

  Future<void> _joinByCode() async {
    final ctrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Код приглашения'),
        content: TextField(
          controller: ctrl,
          decoration: const InputDecoration(hintText: 'Введите код'),
          textCapitalization: TextCapitalization.characters,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Войти')),
        ],
      ),
    );
    if (ok == true && ctrl.text.trim().isNotEmpty) {
      await _learning.joinByInviteCode(ctrl.text.trim());
      _tabs.animateTo(0);
      await _load();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Вы записаны на курс')),
        );
      }
    }
    ctrl.dispose();
  }

  Future<void> _openPublicCourse(PublicCourseListItemDto course) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Записаться на курс?'),
        content: Text('«${course.title}»'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Записаться')),
        ],
      ),
    );
    if (ok != true) return;
    await _learning.enrollPublicCourse(course.id);
    _tabs.animateTo(0);
    await _load();
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Вы записаны на курс')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ShellHeader(
          title: 'Курсы',
          subtitle: 'Мои записи и каталог',
          trailing: IconButton.filled(
            style: IconButton.styleFrom(
              backgroundColor: Colors.white,
              foregroundColor: AppTheme.primaryColor,
            ),
            icon: const Icon(Icons.vpn_key_outlined),
            onPressed: _joinByCode,
            tooltip: 'Код приглашения',
          ),
        ),
        TabBar(
          controller: _tabs,
          tabs: const [
            Tab(text: 'Мои'),
            Tab(text: 'Каталог'),
          ],
        ),
        Padding(
          padding: const EdgeInsets.all(16),
          child: SearchFilterBar(
            controller: _search,
            hint: 'Поиск курсов…',
            onChanged: _applySearch,
            trailing: LanguageFilterChip(
              selectedCode: _languageCode,
              onChanged: (v) => setState(() => _languageCode = v),
            ),
          ),
        ),
        Expanded(
          child: _loading
              ? const Center(child: CircularProgressIndicator())
              : TabBarView(
                  controller: _tabs,
                  children: [
                    _enrolledList(),
                    _publicList(),
                  ],
                ),
        ),
      ],
    );
  }

  Widget _enrolledList() {
    final list = _filteredEnrolled;
    if (list.isEmpty) {
      return const EmptyState(
        icon: Icons.school_outlined,
        message: 'Вы ещё не записаны на курсы',
      );
    }
    return ListView.builder(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: list.length,
      itemBuilder: (ctx, i) {
        final c = list[i];
        return CourseCard(
          title: c.title,
          subtitle: '${c.targetLanguage.nativeName} · ${(c.progressPercent ?? 0).toStringAsFixed(0)}%',
          leadingColor: AppTheme.primaryColor,
          onTap: () => Get.to(() => CourseOutlineScreen(
                courseId: c.id,
                title: c.title,
                targetLanguageCode: c.targetLanguage.code,
              )),
        );
      },
    );
  }

  Widget _publicList() {
    final enrolledIds = _enrolled.map((e) => e.id).toSet();
    final list = _filteredPublic.where((c) => !enrolledIds.contains(c.id)).toList();
    if (list.isEmpty) {
      return const EmptyState(
        icon: Icons.public,
        message: 'Нет доступных публичных курсов',
      );
    }
    return ListView.builder(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: list.length,
      itemBuilder: (ctx, i) {
        final c = list[i];
        return CourseCard(
          title: c.title,
          subtitle: c.targetLanguage.nativeName,
          leadingIcon: Icons.public,
          leadingColor: AppTheme.accentColor,
          onTap: () => _openPublicCourse(c),
        );
      },
    );
  }
}
