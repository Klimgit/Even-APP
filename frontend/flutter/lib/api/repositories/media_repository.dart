import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';

import 'package:online_cource_app/api/base_repository.dart';
import 'package:online_cource_app/api/models/media_dto.dart';

class MediaRepository extends BaseRepository {
  MediaRepository(super.client);

  Future<MediaListResponseDto> listTeacherMedia({
    String? query,
    String? kind,
    String? languageCode,
    int page = 1,
    int limit = 50,
  }) async {
    return MediaListResponseDto.fromJson(
      await getJson('/teacher/media', query: {
        if (query != null && query.isNotEmpty) 'q': query,
        if (kind != null) 'kind': kind,
        if (languageCode != null) 'language_code': languageCode,
        'page': page,
        'limit': limit,
      }),
    );
  }

  Future<MediaListResponseDto> listPlatformMediaPicker({
    required String languageCode,
    String? query,
    String? kind,
    int page = 1,
    int limit = 50,
  }) async {
    return MediaListResponseDto.fromJson(
      await getJson('/teacher/languages/$languageCode/media/platform', query: {
        if (query != null && query.isNotEmpty) 'q': query,
        if (kind != null) 'kind': kind,
        'page': page,
        'limit': limit,
      }),
    );
  }

  Future<MediaAssetDto> uploadTeacherMedia({
    required PlatformFile file,
    required String languageId,
    String? displayName,
    String? linkedLexemeId,
  }) async {
    final bytes = file.bytes;
    if (bytes == null) {
      throw Exception('Не удалось прочитать файл');
    }
    final mime = file.extension != null ? _guessMime(file.extension!) : 'application/octet-stream';
    final presign = PresignResponseDto.fromJson(
      await postJson('/teacher/media/presign', {
        'filename': file.name,
        'mime_type': mime,
        'size_bytes': bytes.length,
        'language_id': languageId,
      }),
    );

    final uploadDio = Dio();
    await uploadDio.put<void>(
      presign.uploadUrl,
      data: bytes,
      options: Options(headers: {'Content-Type': mime}),
    );

    return MediaAssetDto.fromJson(
      await postJson('/teacher/media/confirm', {
        'object_key': presign.objectKey,
        'mime_type': mime,
        'size_bytes': bytes.length,
        'language_id': languageId,
        'display_name': displayName ?? file.name,
        if (linkedLexemeId != null) 'linked_lexeme_id': linkedLexemeId,
      }),
    );
  }

  String _guessMime(String ext) {
    switch (ext.toLowerCase()) {
      case 'png':
        return 'image/png';
      case 'jpg':
      case 'jpeg':
        return 'image/jpeg';
      case 'gif':
        return 'image/gif';
      case 'webp':
        return 'image/webp';
      case 'mp3':
        return 'audio/mpeg';
      case 'wav':
        return 'audio/wav';
      case 'mp4':
        return 'video/mp4';
      case 'webm':
        return 'video/webm';
      default:
        return 'application/octet-stream';
    }
  }
}
