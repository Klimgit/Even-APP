import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/learning_dto.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Cached language list for pickers and filters.
class LanguagesController extends GetxController {
  final LexiconRepository _lexicon = Get.find<LexiconRepository>();

  final RxList<LanguageDto> languages = <LanguageDto>[].obs;
  final RxBool loading = false.obs;
  final RxnString error = RxnString();

  @override
  void onInit() {
    super.onInit();
    load();
  }

  Future<void> load() async {
    loading.value = true;
    error.value = null;
    try {
      languages.assignAll(await _lexicon.listPublicLanguages());
    } catch (e) {
      error.value = e.toString();
      languages.clear();
    } finally {
      loading.value = false;
    }
  }

  LanguageDto? byCode(String code) {
    for (final l in languages) {
      if (l.code == code) return l;
    }
    return null;
  }

  LanguageDto? byId(String id) {
    for (final l in languages) {
      if (l.id == id) return l;
    }
    return null;
  }
}

/// Gradient page header used across student and teacher shells.
class ShellHeader extends StatelessWidget {
  final String title;
  final String? subtitle;
  final Widget? trailing;

  const ShellHeader({
    super.key,
    required this.title,
    this.subtitle,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 20),
      decoration: const BoxDecoration(
        gradient: AppTheme.primaryGradient,
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(24)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
                      ),
                ),
                if (subtitle != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    subtitle!,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Colors.white.withValues(alpha: 0.85),
                        ),
                  ),
                ],
              ],
            ),
          ),
          if (trailing != null) trailing!,
        ],
      ),
    );
  }
}

class SearchFilterBar extends StatelessWidget {
  final TextEditingController controller;
  final String hint;
  final VoidCallback onChanged;
  final Widget? trailing;

  const SearchFilterBar({
    super.key,
    required this.controller,
    required this.hint,
    required this.onChanged,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      elevation: 2,
      shadowColor: AppTheme.primaryColor.withValues(alpha: 0.08),
      borderRadius: BorderRadius.circular(AppTheme.borderRadius),
      color: Colors.white,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 4),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: controller,
                decoration: InputDecoration(
                  hintText: hint,
                  prefixIcon: Icon(Icons.search, color: AppTheme.primaryColor.withValues(alpha: 0.7)),
                  border: InputBorder.none,
                  enabledBorder: InputBorder.none,
                  focusedBorder: InputBorder.none,
                  contentPadding: const EdgeInsets.symmetric(vertical: 12),
                  isDense: true,
                ),
                onChanged: (_) => onChanged(),
                onSubmitted: (_) => onChanged(),
              ),
            ),
            if (trailing != null) ...[
              Container(
                height: 32,
                width: 1,
                color: AppTheme.dividerColor,
              ),
              const SizedBox(width: 8),
              trailing!,
              const SizedBox(width: 8),
            ],
          ],
        ),
      ),
    );
  }
}

class LanguageFilterChip extends StatelessWidget {
  final String? selectedCode;
  final ValueChanged<String?> onChanged;

  const LanguageFilterChip({
    super.key,
    required this.selectedCode,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final langs = Get.find<LanguagesController>().languages;
    return Obx(() {
      if (langs.isEmpty) return const SizedBox.shrink();
      return DropdownButton<String?>(
        value: selectedCode,
        underline: const SizedBox.shrink(),
        hint: Text('Язык', style: TextStyle(color: AppTheme.secondaryTextColor)),
        items: [
          const DropdownMenuItem(value: null, child: Text('Все языки')),
          ...langs.map(
            (l) => DropdownMenuItem(value: l.code, child: Text(l.nativeName)),
          ),
        ],
        onChanged: onChanged,
      );
    });
  }
}

class FilterChipsRow extends StatelessWidget {
  final List<(String label, String value)> options;
  final String selected;
  final ValueChanged<String> onSelected;

  const FilterChipsRow({
    super.key,
    required this.options,
    required this.selected,
    required this.onSelected,
  });

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          for (final (label, value) in options)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: FilterChip(
                label: Text(label),
                selected: selected == value,
                showCheckmark: false,
                selectedColor: AppTheme.primaryColor,
                labelStyle: TextStyle(
                  color: selected == value ? Colors.white : AppTheme.textColor,
                  fontWeight: selected == value ? FontWeight.w600 : FontWeight.w500,
                ),
                side: BorderSide(
                  color: selected == value ? AppTheme.primaryColor : AppTheme.dividerColor,
                ),
                onSelected: (_) => onSelected(value),
              ),
            ),
        ],
      ),
    );
  }
}

class StatCard extends StatelessWidget {
  final String label;
  final String value;
  final IconData icon;

  const StatCard({
    super.key,
    required this.label,
    required this.value,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: AppTheme.primaryColor),
            const SizedBox(height: 8),
            Text(value, style: Theme.of(context).textTheme.headlineMedium),
            Text(label, style: Theme.of(context).textTheme.bodySmall),
          ],
        ),
      ),
    );
  }
}

class EmptyState extends StatelessWidget {
  final IconData icon;
  final String message;
  final String? actionLabel;
  final VoidCallback? onAction;

  const EmptyState({
    super.key,
    required this.icon,
    required this.message,
    this.actionLabel,
    this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: AppTheme.primaryColor.withValues(alpha: 0.08),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, size: 40, color: AppTheme.primaryColor.withValues(alpha: 0.7)),
            ),
            const SizedBox(height: 16),
            Text(
              message,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                    color: AppTheme.secondaryTextColor,
                  ),
            ),
            if (actionLabel != null && onAction != null) ...[
              const SizedBox(height: 16),
              OutlinedButton(onPressed: onAction, child: Text(actionLabel!)),
            ],
          ],
        ),
      ),
    );
  }
}

class CourseCard extends StatelessWidget {
  final String title;
  final String subtitle;
  final IconData leadingIcon;
  final Color leadingColor;
  final VoidCallback onTap;
  final Widget? trailing;

  const CourseCard({
    super.key,
    required this.title,
    required this.subtitle,
    required this.onTap,
    this.leadingIcon = Icons.school_outlined,
    this.leadingColor = AppTheme.primaryColor,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Material(
        color: Colors.white,
        elevation: 2,
        shadowColor: AppTheme.primaryColor.withValues(alpha: 0.06),
        borderRadius: BorderRadius.circular(AppTheme.borderRadius),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(AppTheme.borderRadius),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        leadingColor,
                        leadingColor.withValues(alpha: 0.75),
                      ],
                    ),
                    borderRadius: BorderRadius.circular(14),
                  ),
                  child: Icon(leadingIcon, color: Colors.white, size: 22),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style: Theme.of(context).textTheme.titleLarge,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        subtitle,
                        style: Theme.of(context).textTheme.bodySmall,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                if (trailing != null) trailing! else Icon(Icons.chevron_right, color: AppTheme.secondaryTextColor),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
