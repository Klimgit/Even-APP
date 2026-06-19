class MediaAssetDto {
  final String id;
  final String scope;
  final String languageId;
  final String displayName;
  final String mimeType;
  final String mediaKind;
  final String url;
  final int sizeBytes;

  const MediaAssetDto({
    required this.id,
    required this.scope,
    required this.languageId,
    required this.displayName,
    required this.mimeType,
    required this.mediaKind,
    required this.url,
    required this.sizeBytes,
  });

  factory MediaAssetDto.fromJson(Map<String, dynamic> json) => MediaAssetDto(
        id: json['id'] as String,
        scope: json['scope'] as String? ?? 'teacher',
        languageId: json['language_id'] as String,
        displayName: json['display_name'] as String,
        mimeType: json['mime_type'] as String,
        mediaKind: json['media_kind'] as String,
        url: json['url'] as String,
        sizeBytes: (json['size_bytes'] as num?)?.toInt() ?? 0,
      );
}

class MediaListResponseDto {
  final List<MediaAssetDto> items;
  final int total;

  const MediaListResponseDto({required this.items, required this.total});

  factory MediaListResponseDto.fromJson(Map<String, dynamic> json) =>
      MediaListResponseDto(
        items: (json['items'] as List<dynamic>? ?? [])
            .map((e) => MediaAssetDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        total: json['total'] as int? ?? 0,
      );
}

class PresignResponseDto {
  final String uploadUrl;
  final String objectKey;
  final String mediaAssetId;

  const PresignResponseDto({
    required this.uploadUrl,
    required this.objectKey,
    required this.mediaAssetId,
  });

  factory PresignResponseDto.fromJson(Map<String, dynamic> json) =>
      PresignResponseDto(
        uploadUrl: json['upload_url'] as String,
        objectKey: json['object_key'] as String,
        mediaAssetId: json['media_asset_id'] as String,
      );
}
