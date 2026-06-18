import 'package:flutter_tts/flutter_tts.dart';

/// Thin wrapper around text-to-speech used by listening exercises.
///
/// Centralised here so the playback backend (TTS now, recorded audio files
/// later) can be swapped without touching the exercise widgets.
class AudioService {
  AudioService._();
  static final AudioService instance = AudioService._();

  final FlutterTts _tts = FlutterTts();
  bool _ready = false;

  Future<void> _ensureReady(String languageCode) async {
    if (!_ready) {
      await _tts.awaitSpeakCompletion(true);
      _ready = true;
    }
    await _tts.setLanguage(languageCode);
    await _tts.setSpeechRate(0.45);
    await _tts.setPitch(1.0);
  }

  /// Speaks [text] in the given [languageCode] (BCP-47, e.g. 'en-US').
  Future<void> speak(String text, {String languageCode = 'en-US'}) async {
    if (text.trim().isEmpty) return;
    await _ensureReady(languageCode);
    await _tts.stop();
    await _tts.speak(text);
  }

  Future<void> stop() => _tts.stop();
}
