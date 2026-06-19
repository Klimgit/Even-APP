import 'package:cached_network_image/cached_network_image.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/models/media_dto.dart';
import 'package:online_cource_app/api/repositories/media_repository.dart';
import 'package:online_cource_app/features/shared/widgets.dart';
import 'package:online_cource_app/theme/app_theme.dart';

Future<MediaAssetDto?> showMediaPicker({
  required BuildContext context,
  required String languageCode,
  required String languageId,
  String? kind,
}) async {
  return showModalBottomSheet<MediaAssetDto>(
    context: context,
    isScrollControlled: true,
    builder: (ctx) => _MediaPickerSheet(
      languageCode: languageCode,
      languageId: languageId,
      kind: kind,
    ),
  );
}

class _MediaPickerSheet extends StatefulWidget {
  final String languageCode;
  final String languageId;
  final String? kind;

  const _MediaPickerSheet({
    required this.languageCode,
    required this.languageId,
    this.kind,
  });

  @override
  State<_MediaPickerSheet> createState() => _MediaPickerSheetState();
}

class _MediaPickerSheetState extends State<_MediaPickerSheet>
    with SingleTickerProviderStateMixin {
  final _media = Get.find<MediaRepository>();
  late TabController _tabs;
  List<MediaAssetDto> _teacherItems = [];
  List<MediaAssetDto> _platformItems = [];
  bool _loading = false;
  bool _uploading = false;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 3, vsync: this);
    _load();
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final teacher = await _media.listTeacherMedia(
        languageCode: widget.languageCode,
        kind: widget.kind,
      );
      final platform = await _media.listPlatformMediaPicker(
        languageCode: widget.languageCode,
        kind: widget.kind,
      );
      setState(() {
        _teacherItems = teacher.items;
        _platformItems = platform.items;
      });
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _upload() async {
    final kind = widget.kind;
    FileType fileType = FileType.any;
    if (kind == 'image') fileType = FileType.image;
    if (kind == 'audio') fileType = FileType.audio;
    if (kind == 'video') fileType = FileType.video;

    final result = await FilePicker.platform.pickFiles(
      type: fileType,
      withData: true,
    );
    if (result == null || result.files.isEmpty) return;

    setState(() => _uploading = true);
    try {
      final asset = await _media.uploadTeacherMedia(
        file: result.files.first,
        languageId: widget.languageId,
      );
      if (mounted) Navigator.pop(context, asset);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return DraggableScrollableSheet(
      expand: false,
      initialChildSize: 0.75,
      maxChildSize: 0.95,
      builder: (ctx, scrollController) => Column(
        children: [
          TabBar(
            controller: _tabs,
            tabs: const [
              Tab(text: 'Мои'),
              Tab(text: 'Платформа'),
              Tab(text: 'Загрузить'),
            ],
          ),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : TabBarView(
                    controller: _tabs,
                    children: [
                      _mediaGrid(_teacherItems, scrollController),
                      _mediaGrid(_platformItems, scrollController),
                      _uploadTab(),
                    ],
                  ),
          ),
        ],
      ),
    );
  }

  Widget _uploadTab() {
    return Center(
      child: _uploading
          ? const CircularProgressIndicator()
          : Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.cloud_upload_outlined, size: 64, color: AppTheme.primaryColor),
                const SizedBox(height: 16),
                const Text('Загрузить файл в медиатеку'),
                const SizedBox(height: 24),
                ElevatedButton.icon(
                  onPressed: _upload,
                  icon: const Icon(Icons.upload_file),
                  label: const Text('Выбрать файл'),
                ),
              ],
            ),
    );
  }

  Widget _mediaGrid(List<MediaAssetDto> items, ScrollController scrollController) {
    if (items.isEmpty) {
      return const EmptyState(icon: Icons.perm_media_outlined, message: 'Медиа не найдено');
    }
    return GridView.builder(
      controller: scrollController,
      padding: const EdgeInsets.all(12),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        crossAxisSpacing: 8,
        mainAxisSpacing: 8,
        childAspectRatio: 1,
      ),
      itemCount: items.length,
      itemBuilder: (ctx, i) {
        final m = items[i];
        return InkWell(
          onTap: () => Navigator.pop(context, m),
          child: Card(
            clipBehavior: Clip.antiAlias,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Expanded(
                  child: m.mediaKind == 'image'
                      ? CachedNetworkImage(imageUrl: m.url, fit: BoxFit.cover)
                      : Center(
                          child: Icon(
                            m.mediaKind == 'audio' ? Icons.audiotrack : Icons.videocam,
                            size: 32,
                          ),
                        ),
                ),
                Padding(
                  padding: const EdgeInsets.all(4),
                  child: Text(
                    m.displayName,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontSize: 11),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class MediaPickerField extends StatelessWidget {
  final String? mediaId;
  final String? displayName;
  final String label;
  final VoidCallback onPick;

  const MediaPickerField({
    super.key,
    required this.mediaId,
    required this.displayName,
    required this.label,
    required this.onPick,
  });

  @override
  Widget build(BuildContext context) {
    return InputDecorator(
      decoration: InputDecoration(labelText: label),
      child: InkWell(
        onTap: onPick,
        child: Row(
          children: [
            Expanded(
              child: Text(
                displayName ?? (mediaId != null && mediaId!.isNotEmpty ? mediaId! : 'Выбрать…'),
                style: TextStyle(
                  color: displayName != null ? AppTheme.textColor : AppTheme.secondaryTextColor,
                ),
              ),
            ),
            const Icon(Icons.perm_media_outlined, size: 20),
          ],
        ),
      ),
    );
  }
}
