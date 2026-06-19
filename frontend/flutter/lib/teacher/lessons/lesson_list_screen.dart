import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:flutter/material.dart';
import 'package:get/get.dart';

import 'package:online_cource_app/teacher/lessons/lesson_editor_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Lessons of a single class. Teacher creates lessons and opens the editor to
/// build them from exercise steps.
class LessonListScreen extends StatelessWidget {
  final String classId;
  final String className;

  const LessonListScreen({
    super.key,
    required this.classId,
    required this.className,
  });

  CollectionReference<Map<String, dynamic>> get _lessons => FirebaseFirestore
      .instance
      .collection('classes')
      .doc(classId)
      .collection('lessons');

  Future<void> _createLesson(BuildContext context) async {
    final controller = TextEditingController();
    final title = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('New lesson'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(hintText: 'Lesson title'),
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel')),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, controller.text.trim()),
            child: const Text('Create'),
          ),
        ],
      ),
    );
    if (title == null || title.isEmpty) return;
    final doc = await _lessons.add({
      'title': title,
      'steps': <Map<String, dynamic>>[],
      'createdAt': FieldValue.serverTimestamp(),
    });
    Get.to(() => LessonEditorScreen(
          classId: classId,
          lessonId: doc.id,
          initialTitle: title,
        ));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        backgroundColor: AppTheme.backgroundColor,
        elevation: 0,
        foregroundColor: AppTheme.textColor,
        title: Text(className),
      ),
      body: StreamBuilder<QuerySnapshot<Map<String, dynamic>>>(
        stream: _lessons.orderBy('createdAt').snapshots(),
        builder: (context, snapshot) {
          if (!snapshot.hasData) {
            return const Center(child: CircularProgressIndicator());
          }
          final docs = snapshot.data!.docs;
          if (docs.isEmpty) {
            return Center(
              child: Text('No lessons yet. Tap + to create one.',
                  style: TextStyle(color: AppTheme.secondaryTextColor)),
            );
          }
          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: docs.length,
            itemBuilder: (context, index) {
              final data = docs[index].data();
              final title = data['title'] as String? ?? 'Untitled';
              final stepCount = (data['steps'] as List?)?.length ?? 0;
              return Card(
                elevation: 0,
                margin: const EdgeInsets.only(bottom: 10),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: BorderSide(color: AppTheme.dividerColor),
                ),
                child: ListTile(
                  leading: Icon(Icons.menu_book_outlined,
                      color: AppTheme.accentColor),
                  title: Text(title,
                      style: const TextStyle(fontWeight: FontWeight.w600)),
                  subtitle: Text('$stepCount step${stepCount == 1 ? '' : 's'}'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => Get.to(() => LessonEditorScreen(
                        classId: classId,
                        lessonId: docs[index].id,
                        initialTitle: title,
                      )),
                ),
              );
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _createLesson(context),
        backgroundColor: AppTheme.accentColor,
        child: const Icon(Icons.add),
      ),
    );
  }
}
