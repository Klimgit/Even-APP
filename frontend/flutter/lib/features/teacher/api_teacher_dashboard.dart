import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/features/shared/alphabet_controller.dart';
import 'package:online_cource_app/features/shared/profile_tab.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/features/teacher/courses/course_admin_screen.dart';
import 'package:online_cource_app/features/teacher/knowledge/knowledge_base_screen.dart';
import 'package:online_cource_app/features/teacher/students/students_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Teacher shell — same bottom navigation pattern as the student app.
class ApiTeacherDashboard extends StatefulWidget {
  const ApiTeacherDashboard({super.key});

  @override
  State<ApiTeacherDashboard> createState() => _ApiTeacherDashboardState();
}

class _ApiTeacherDashboardState extends State<ApiTeacherDashboard> {
  int _index = 0;

  @override
  void initState() {
    super.initState();
    if (!Get.isRegistered<LanguagesController>()) {
      Get.put(LanguagesController(), permanent: true);
    }
    if (!Get.isRegistered<AlphabetController>()) {
      Get.put(AlphabetController(), permanent: true);
    }
  }

  @override
  Widget build(BuildContext context) {
    final pages = const [
      CourseAdminScreen(),
      StudentsScreen(),
      KnowledgeBaseScreen(),
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
            icon: Icon(Icons.people_outline),
            selectedIcon: Icon(Icons.people),
            label: 'Ученики',
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
