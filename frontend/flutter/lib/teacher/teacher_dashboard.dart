import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:flutter/material.dart';
import 'package:get/get.dart';

import 'package:online_cource_app/controllers/auth_controller.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Teacher / admin workspace.
///
/// Layout mirrors the provided mock-up: a top bar with the product name and a
/// circular avatar, a left navigation rail (My classes / Students / Knowledge
/// base) and a content area. Classes live in the top-level `classes`
/// collection and are shared with the web admin panel.
class TeacherDashboard extends StatefulWidget {
  const TeacherDashboard({super.key});

  @override
  State<TeacherDashboard> createState() => _TeacherDashboardState();
}

enum _TeacherSection { classes, students, knowledgeBase }

class _TeacherDashboardState extends State<TeacherDashboard> {
  final AuthController _auth = Get.find<AuthController>();
  _TeacherSection _section = _TeacherSection.classes;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Column(
          children: [
            _buildTopBar(),
            const Divider(height: 1),
            Expanded(
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  _buildSideMenu(),
                  const VerticalDivider(width: 1),
                  Expanded(child: _buildContent()),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Top bar
  // ---------------------------------------------------------------------------
  Widget _buildTopBar() {
    final name = (_auth.userData['name'] as String?) ??
        _auth.currentUser?.displayName ??
        _auth.currentUser?.email ??
        '?';
    final initial = name.isNotEmpty ? name[0].toUpperCase() : '?';

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
      child: Row(
        children: [
          Text(
            'Even App',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.bold,
              color: AppTheme.textColor,
            ),
          ),
          const Spacer(),
          PopupMenuButton<String>(
            offset: const Offset(0, 48),
            onSelected: (value) {
              if (value == 'logout') _auth.signOutUsers();
            },
            itemBuilder: (context) => const [
              PopupMenuItem(value: 'logout', child: Text('Выйти')),
            ],
            child: CircleAvatar(
              radius: 20,
              backgroundColor: AppTheme.accentColor,
              child: Text(
                initial,
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Side menu
  // ---------------------------------------------------------------------------
  Widget _buildSideMenu() {
    return SizedBox(
      width: 200,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _sideItem('Мои классы', _TeacherSection.classes),
            const SizedBox(height: 8),
            _sideItem('Ученики', _TeacherSection.students),
            const SizedBox(height: 8),
            _sideItem('База знаний', _TeacherSection.knowledgeBase),
          ],
        ),
      ),
    );
  }

  Widget _sideItem(String label, _TeacherSection section) {
    final bool selected = _section == section;
    return InkWell(
      onTap: () => setState(() => _section = section),
      borderRadius: BorderRadius.circular(8),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              label,
              style: TextStyle(
                fontSize: 17,
                fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
                color: AppTheme.textColor,
              ),
            ),
            if (selected)
              Container(
                margin: const EdgeInsets.only(top: 4),
                height: 2,
                width: label.length * 9.0,
                color: AppTheme.textColor,
              ),
          ],
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Content
  // ---------------------------------------------------------------------------
  Widget _buildContent() {
    switch (_section) {
      case _TeacherSection.classes:
        return const _ClassesSection();
      case _TeacherSection.students:
        return const _PlaceholderSection(
          icon: Icons.people_outline,
          title: 'Ученики',
          message: 'Здесь появится список ваших учеников.',
        );
      case _TeacherSection.knowledgeBase:
        return const _PlaceholderSection(
          icon: Icons.menu_book_outlined,
          title: 'База знаний',
          message: 'Здесь будут материалы для уроков.',
        );
    }
  }
}

// =============================================================================
// Classes section: "Мои классы" / "Публичные классы" tabs.
// =============================================================================
class _ClassesSection extends StatefulWidget {
  const _ClassesSection();

  @override
  State<_ClassesSection> createState() => _ClassesSectionState();
}

class _ClassesSectionState extends State<_ClassesSection> {
  final AuthController _auth = Get.find<AuthController>();
  final FirebaseFirestore _db = FirebaseFirestore.instance;
  bool _showPublic = false;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              _tab('Мои классы', !_showPublic,
                  () => setState(() => _showPublic = false)),
              const SizedBox(width: 24),
              _tab('Публичные классы', _showPublic,
                  () => setState(() => _showPublic = true)),
            ],
          ),
          const SizedBox(height: 24),
          Expanded(child: _buildList()),
        ],
      ),
    );
  }

  Widget _tab(String label, bool selected, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      child: Column(
        children: [
          Text(
            label,
            style: TextStyle(
              fontSize: 16,
              fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
              color: AppTheme.textColor,
            ),
          ),
          const SizedBox(height: 4),
          Container(
            height: 2,
            width: label.length * 8.0,
            color: selected ? AppTheme.textColor : Colors.transparent,
          ),
        ],
      ),
    );
  }

  Query<Map<String, dynamic>> get _query {
    final classes = _db.collection('classes');
    if (_showPublic) {
      return classes.where('isPublic', isEqualTo: true);
    }
    return classes.where('ownerId', isEqualTo: _auth.currentUser?.uid);
  }

  Widget _buildList() {
    return StreamBuilder<QuerySnapshot<Map<String, dynamic>>>(
      stream: _query.snapshots(),
      builder: (context, snapshot) {
        if (snapshot.hasError) {
          return Center(child: Text('Ошибка: ${snapshot.error}'));
        }
        if (!snapshot.hasData) {
          return const Center(child: CircularProgressIndicator());
        }
        final docs = snapshot.data!.docs;
        return SingleChildScrollView(
          child: Wrap(
            spacing: 16,
            runSpacing: 16,
            crossAxisAlignment: WrapCrossAlignment.start,
            children: [
              for (final doc in docs) _classCard(doc.data()),
              if (!_showPublic) _createCard(),
            ],
          ),
        );
      },
    );
  }

  Widget _classCard(Map<String, dynamic> data) {
    final name = (data['name'] as String?) ?? 'Без названия';
    final colorValue = (data['coverColor'] as int?) ?? 0xFF9BE8B4;
    return SizedBox(
      width: 150,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            height: 80,
            decoration: BoxDecoration(
              color: Color(colorValue),
              borderRadius: BorderRadius.circular(6),
            ),
          ),
          const SizedBox(height: 6),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
            decoration: BoxDecoration(
              border: Border.all(color: AppTheme.dividerColor),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              name,
              style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
            ),
          ),
        ],
      ),
    );
  }

  Widget _createCard() {
    return InkWell(
      onTap: _showCreateClassDialog,
      child: Container(
        width: 320,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 16),
        decoration: BoxDecoration(
          border: Border.all(color: AppTheme.dividerColor),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Row(
          children: [
            const Text(
              'Создать новый класс',
              style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
            ),
            const Spacer(),
            Icon(Icons.add_circle_outline, color: AppTheme.primaryColor),
          ],
        ),
      ),
    );
  }

  Future<void> _showCreateClassDialog() async {
    final controller = TextEditingController();
    bool isPublic = false;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Новый класс'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: controller,
                autofocus: true,
                decoration: const InputDecoration(
                  hintText: 'Название класса',
                ),
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Checkbox(
                    value: isPublic,
                    onChanged: (v) =>
                        setDialogState(() => isPublic = v ?? false),
                  ),
                  const Text('Публичный'),
                ],
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Отмена'),
            ),
            ElevatedButton(
              onPressed: () async {
                final name = controller.text.trim();
                if (name.isEmpty) return;
                await _createClass(name, isPublic);
                if (context.mounted) Navigator.pop(context);
              },
              child: const Text('Создать'),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _createClass(String name, bool isPublic) async {
    await _db.collection('classes').add({
      'name': name,
      'isPublic': isPublic,
      'ownerId': _auth.currentUser?.uid,
      'ownerName': _auth.userData['name'] ?? _auth.currentUser?.displayName,
      'coverColor': 0xFF9BE8B4,
      'studentIds': <String>[],
      'createdAt': FieldValue.serverTimestamp(),
    });
  }
}

// =============================================================================
// Generic placeholder for not-yet-built sections.
// =============================================================================
class _PlaceholderSection extends StatelessWidget {
  final IconData icon;
  final String title;
  final String message;

  const _PlaceholderSection({
    required this.icon,
    required this.title,
    required this.message,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, size: 56, color: AppTheme.secondaryTextColor),
          const SizedBox(height: 16),
          Text(
            title,
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.w600,
              color: AppTheme.textColor,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            message,
            style: TextStyle(color: AppTheme.secondaryTextColor),
          ),
        ],
      ),
    );
  }
}
