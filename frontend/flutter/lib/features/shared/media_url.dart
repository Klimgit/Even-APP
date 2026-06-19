import 'package:online_cource_app/api/api_client.dart';

/// Resolve lexicon/media API paths to absolute gateway URLs for widgets.
String resolveApiMediaUrl(String? url) {
  if (url == null || url.isEmpty) return '';
  if (url.startsWith('http://') || url.startsWith('https://')) return url;
  if (url.startsWith('/')) return '$kGatewayOrigin$url';
  return '$kGatewayOrigin/$url';
}
