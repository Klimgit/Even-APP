import 'package:audioplayers/audioplayers.dart';
import 'package:chewie/chewie.dart';
import 'package:flutter/material.dart';
import 'package:video_player/video_player.dart';

import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// Kinds of content a teacher can stack into a study-material step.
enum MaterialElementKind { text, image, audio, video }

MaterialElementKind _kindFromString(String? s) {
  return MaterialElementKind.values.firstWhere(
    (k) => k.name == s,
    orElse: () => MaterialElementKind.text,
  );
}

/// One element of a study-material step: a paragraph of text, or a URL to an
/// image / audio / video.
class MaterialElement {
  final MaterialElementKind kind;

  /// Text content (for [MaterialElementKind.text]) or a media URL otherwise.
  final String value;

  const MaterialElement({required this.kind, required this.value});

  factory MaterialElement.fromJson(Map<String, dynamic> json) =>
      MaterialElement(
        kind: _kindFromString(json['kind'] as String?),
        value: json['value'] as String? ?? '',
      );

  Map<String, dynamic> toJson() => {'kind': kind.name, 'value': value};
}

/// Study-material step (non-gradable). Shows stacked text/image/audio/video
/// elements the learner scrolls through; audio and video can be played/paused.
/// Always reports success when the learner continues.
class MaterialExercise extends ExerciseData {
  final String title;
  final List<MaterialElement> elements;

  const MaterialExercise({this.title = '', required this.elements});

  static const String typeId = 'material';

  factory MaterialExercise.fromJson(Map<String, dynamic> json) {
    final raw = json['elements'] as List?;
    return MaterialExercise(
      title: json['title'] as String? ?? '',
      elements: raw
              ?.map((e) =>
                  MaterialElement.fromJson(Map<String, dynamic>.from(e as Map)))
              .toList() ??
          const [],
    );
  }

  @override
  Map<String, dynamic> toJson() => {
        'type': typeId,
        'title': title,
        'elements': elements.map((e) => e.toJson()).toList(),
      };

  @override
  Widget build({required ValueChanged<bool> onResult}) =>
      _MaterialWidget(data: this, onResult: onResult);
}

class _MaterialWidget extends StatelessWidget {
  final MaterialExercise data;
  final ValueChanged<bool> onResult;

  const _MaterialWidget({required this.data, required this.onResult});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (data.title.isNotEmpty) ...[
          Text(
            data.title,
            style: Theme.of(context)
                .textTheme
                .titleLarge
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 16),
        ],
        Expanded(
          child: ListView.separated(
            itemCount: data.elements.length,
            separatorBuilder: (_, __) => const SizedBox(height: 16),
            itemBuilder: (context, index) =>
                _MaterialElementView(element: data.elements[index]),
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 50,
          child: ElevatedButton(
            onPressed: () => onResult(true),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppTheme.accentColor,
              foregroundColor: Colors.white,
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: const Text('Continue',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          ),
        ),
      ],
    );
  }
}

class _MaterialElementView extends StatelessWidget {
  final MaterialElement element;
  const _MaterialElementView({required this.element});

  @override
  Widget build(BuildContext context) {
    switch (element.kind) {
      case MaterialElementKind.text:
        return Text(element.value,
            style: const TextStyle(fontSize: 16, height: 1.4));
      case MaterialElementKind.image:
        return ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: Image.network(
            element.value,
            fit: BoxFit.cover,
            errorBuilder: (_, __, ___) => _broken('Image unavailable'),
          ),
        );
      case MaterialElementKind.audio:
        return _AudioElement(url: element.value);
      case MaterialElementKind.video:
        return _VideoElement(url: element.value);
    }
  }

  Widget _broken(String label) => Container(
        height: 80,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: AppTheme.dividerColor.withOpacity(0.3),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Text(label, style: TextStyle(color: AppTheme.secondaryTextColor)),
      );
}

/// Audio element with a single play/pause control.
class _AudioElement extends StatefulWidget {
  final String url;
  const _AudioElement({required this.url});

  @override
  State<_AudioElement> createState() => _AudioElementState();
}

class _AudioElementState extends State<_AudioElement> {
  final AudioPlayer _player = AudioPlayer();
  bool _playing = false;

  @override
  void initState() {
    super.initState();
    _player.onPlayerStateChanged.listen((state) {
      if (mounted) setState(() => _playing = state == PlayerState.playing);
    });
  }

  @override
  void dispose() {
    _player.dispose();
    super.dispose();
  }

  Future<void> _toggle() async {
    if (_playing) {
      await _player.pause();
    } else {
      await _player.play(UrlSource(widget.url));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.dividerColor),
      ),
      child: Row(
        children: [
          IconButton(
            onPressed: _toggle,
            icon: Icon(
              _playing ? Icons.pause_circle_filled : Icons.play_circle_fill,
              color: AppTheme.accentColor,
              size: 36,
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text('Audio',
                style: TextStyle(color: AppTheme.secondaryTextColor)),
          ),
        ],
      ),
    );
  }
}

/// Video element with standard play/pause controls (Chewie).
class _VideoElement extends StatefulWidget {
  final String url;
  const _VideoElement({required this.url});

  @override
  State<_VideoElement> createState() => _VideoElementState();
}

class _VideoElementState extends State<_VideoElement> {
  VideoPlayerController? _video;
  ChewieController? _chewie;
  bool _error = false;

  @override
  void initState() {
    super.initState();
    _init();
  }

  Future<void> _init() async {
    try {
      final controller =
          VideoPlayerController.networkUrl(Uri.parse(widget.url));
      await controller.initialize();
      setState(() {
        _video = controller;
        _chewie = ChewieController(
          videoPlayerController: controller,
          autoPlay: false,
          looping: false,
          aspectRatio: controller.value.aspectRatio,
        );
      });
    } catch (_) {
      if (mounted) setState(() => _error = true);
    }
  }

  @override
  void dispose() {
    _chewie?.dispose();
    _video?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_error) {
      return Container(
        height: 80,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: AppTheme.dividerColor.withOpacity(0.3),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Text('Video unavailable',
            style: TextStyle(color: AppTheme.secondaryTextColor)),
      );
    }
    if (_chewie == null) {
      return const SizedBox(
        height: 120,
        child: Center(child: CircularProgressIndicator()),
      );
    }
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: AspectRatio(
        aspectRatio: _video!.value.aspectRatio,
        child: Chewie(controller: _chewie!),
      ),
    );
  }
}
