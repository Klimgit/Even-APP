import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:flutter/material.dart';
import 'package:get/get.dart';

import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/exercise_catalog.dart';
import 'package:online_cource_app/exercises/lesson_runner.dart';
import 'package:online_cource_app/teacher/lessons/step_editor_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Lets a teacher build a lesson as an ordered list of exercise steps.
/// Persists to `classes/{classId}/lessons/{lessonId}` in Firestore.
class LessonEditorScreen extends StatefulWidget {
  final String classId;
  final String lessonId;
  final String initialTitle;

  const LessonEditorScreen({
    super.key,
    required this.classId,
    required this.lessonId,
    required this.initialTitle,
  });

  @override
  State<LessonEditorScreen> createState() => _LessonEditorScreenState();
}

class _LessonEditorScreenState extends State<LessonEditorScreen> {
  final FirebaseFirestore _db = FirebaseFirestore.instance;
  late final TextEditingController _title =
      TextEditingController(text: widget.initialTitle);
  final List<ExerciseData> _steps = [];
  bool _loading = true;
  bool _saving = false;

  DocumentReference<Map<String, dynamic>> get _doc => _db
      .collection('classes')
      .doc(widget.classId)
      .collection('lessons')
      .doc(widget.lessonId);

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _title.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final snap = await _doc.get();
    final data = snap.data();
    final rawSteps = (data?['steps'] as List?) ?? const [];
    setState(() {
      _steps
        ..clear()
        ..addAll(rawSteps
            .map((e) => exerciseFromJson(Map<String, dynamic>.from(e as Map))));
      _loading = false;
    });
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      await _doc.set({
        'title': _title.text.trim().isEmpty ? 'Untitled' : _title.text.trim(),
        'steps': _steps.map((e) => e.toJson()).toList(),
        'updatedAt': FieldValue.serverTimestamp(),
      }, SetOptions(merge: true));
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('Lesson saved')));
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  Future<void> _addStep() async {
    final type = await showModalBottomSheet<ExerciseType>(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text('Add a step',
                  style:
                      TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
            ),
            for (final t in exerciseTypes)
              ListTile(
                leading: Icon(t.icon, color: AppTheme.accentColor),
                title: Text(t.label),
                subtitle: Text(t.description),
                onTap: () => Navigator.of(context).pop(t),
              ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
    if (type == null) return;
    final built = await Navigator.of(context).push<ExerciseData>(
      MaterialPageRoute(
        builder: (_) => StepEditorScreen(typeId: type.id),
      ),
    );
    if (built != null) setState(() => _steps.add(built));
  }

  Future<void> _editStep(int index) async {
    final step = _steps[index];
    final typeId = step.toJson()['type'] as String;
    final updated = await Navigator.of(context).push<ExerciseData>(
      MaterialPageRoute(
        builder: (_) => StepEditorScreen(typeId: typeId, initial: step),
      ),
    );
    if (updated != null) setState(() => _steps[index] = updated);
  }

  void _preview() {
    if (_steps.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Add at least one step to preview')));
      return;
    }
    Get.to(() => LessonRunner(
          title: _title.text.trim().isEmpty ? 'Preview' : _title.text.trim(),
          exercises: List<ExerciseData>.from(_steps),
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
        title: const Text('Lesson'),
        actions: [
          IconButton(
            tooltip: 'Preview',
            onPressed: _preview,
            icon: const Icon(Icons.play_circle_outline),
          ),
          _saving
              ? const Padding(
                  padding: EdgeInsets.all(14),
                  child: SizedBox(
                      width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2)),
                )
              : TextButton(
                  onPressed: _save,
                  child: const Text('Save',
                      style: TextStyle(fontWeight: FontWeight.bold)),
                ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: TextField(
                    controller: _title,
                    decoration: InputDecoration(
                      labelText: 'Lesson title',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: AppTheme.dividerColor),
                      ),
                    ),
                  ),
                ),
                Expanded(child: _buildSteps()),
              ],
            ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _addStep,
        backgroundColor: AppTheme.accentColor,
        icon: const Icon(Icons.add),
        label: const Text('Add step'),
      ),
    );
  }

  Widget _buildSteps() {
    if (_steps.isEmpty) {
      return Center(
        child: Text('No steps yet. Tap "Add step".',
            style: TextStyle(color: AppTheme.secondaryTextColor)),
      );
    }
    return ReorderableListView.builder(
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 96),
      itemCount: _steps.length,
      onReorder: (oldIndex, newIndex) => setState(() {
        if (newIndex > oldIndex) newIndex -= 1;
        final item = _steps.removeAt(oldIndex);
        _steps.insert(newIndex, item);
      }),
      itemBuilder: (context, index) {
        final step = _steps[index];
        final typeId = step.toJson()['type'] as String;
        return Card(
          key: ValueKey(step),
          elevation: 0,
          margin: const EdgeInsets.only(bottom: 10),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
            side: BorderSide(color: AppTheme.dividerColor),
          ),
          child: ListTile(
            leading: CircleAvatar(
              backgroundColor: AppTheme.accentColor.withOpacity(0.12),
              child: Text('${index + 1}',
                  style: TextStyle(
                      color: AppTheme.accentColor,
                      fontWeight: FontWeight.bold)),
            ),
            title: Text(exerciseTypeLabel(typeId),
                style: const TextStyle(fontWeight: FontWeight.w600)),
            subtitle: Text(_stepSummary(step),
                maxLines: 1, overflow: TextOverflow.ellipsis),
            onTap: () => _editStep(index),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.delete_outline),
                  onPressed: () => setState(() => _steps.removeAt(index)),
                ),
                const Icon(Icons.drag_handle),
              ],
            ),
          ),
        );
      },
    );
  }

  String _stepSummary(ExerciseData step) {
    final json = step.toJson();
    switch (json['type']) {
      case 'listen_choice':
        return json['word'] as String? ?? '';
      case 'sentence_builder':
        return (json['correctWords'] as List?)?.join(' ') ?? '';
      case 'word_match':
        return '${(json['pairs'] as List?)?.length ?? 0} pairs';
      case 'material':
        final title = json['title'] as String? ?? '';
        final count = (json['elements'] as List?)?.length ?? 0;
        return title.isNotEmpty ? title : '$count element${count == 1 ? '' : 's'}';
      default:
        return '';
    }
  }
}
