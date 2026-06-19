import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/content_dto.dart';
import 'package:online_cource_app/api/repositories/content_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';

class StudentsScreen extends StatefulWidget {
  const StudentsScreen({super.key});

  @override
  State<StudentsScreen> createState() => _StudentsScreenState();
}

class _StudentsScreenState extends State<StudentsScreen> {
  final _content = Get.find<ContentRepository>();
  List<CourseDto> _courses = [];
  String? _selectedCourseId;
  List<StudentDto> _students = [];
  bool _loadingCourses = true;
  bool _loadingStudents = false;

  @override
  void initState() {
    super.initState();
    _loadCourses();
  }

  Future<void> _loadCourses() async {
    setState(() => _loadingCourses = true);
    try {
      _courses = await _content.listTeacherCourses();
      if (_courses.isNotEmpty) {
        _selectedCourseId ??= _courses.first.id;
        await _loadStudents();
      }
    } finally {
      if (mounted) setState(() => _loadingCourses = false);
    }
  }

  Future<void> _loadStudents() async {
    if (_selectedCourseId == null) return;
    setState(() => _loadingStudents = true);
    try {
      _students = await _content.listStudents(_selectedCourseId!);
    } finally {
      if (mounted) setState(() => _loadingStudents = false);
    }
  }

  Future<void> _enrollByEmail() async {
    if (_selectedCourseId == null) return;
    final ctrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Добавить ученика'),
        content: TextField(
          controller: ctrl,
          decoration: const InputDecoration(hintText: 'Email ученика'),
          keyboardType: TextInputType.emailAddress,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Добавить')),
        ],
      ),
    );
    if (ok == true && ctrl.text.trim().isNotEmpty) {
      await _content.enrollStudentByEmail(email: ctrl.text.trim(), courseId: _selectedCourseId!);
      await _loadStudents();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loadingCourses) return const Center(child: CircularProgressIndicator());
    if (_courses.isEmpty) {
      return const EmptyState(
        icon: Icons.people_outline,
        message: 'Сначала создайте курс, чтобы видеть учеников',
      );
    }

    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Ученики', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: DropdownButtonFormField<String>(
                  value: _selectedCourseId,
                  decoration: const InputDecoration(labelText: 'Курс'),
                  items: _courses
                      .map((c) => DropdownMenuItem(value: c.id, child: Text(c.title)))
                      .toList(),
                  onChanged: (v) async {
                    setState(() => _selectedCourseId = v);
                    await _loadStudents();
                  },
                ),
              ),
              const SizedBox(width: 16),
              ElevatedButton.icon(
                onPressed: _enrollByEmail,
                icon: const Icon(Icons.person_add),
                label: const Text('Добавить'),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Expanded(
            child: _loadingStudents
                ? const Center(child: CircularProgressIndicator())
                : _students.isEmpty
                    ? const EmptyState(icon: Icons.people_outline, message: 'Пока нет учеников')
                    : ListView.separated(
                        itemCount: _students.length,
                        separatorBuilder: (_, __) => const Divider(height: 1),
                        itemBuilder: (ctx, i) {
                          final s = _students[i];
                          return ListTile(
                            leading: CircleAvatar(child: Text(s.email[0].toUpperCase())),
                            title: Text(s.displayName ?? s.email),
                            subtitle: Text(s.email),
                            trailing: Text('${s.enrolledCourses.length} курсов'),
                          );
                        },
                      ),
          ),
        ],
        ),
      ),
    );
  }
}
