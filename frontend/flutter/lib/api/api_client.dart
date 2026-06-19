import 'package:dio/dio.dart';

import 'package:online_cource_app/api/models/auth_dto.dart';
import 'package:online_cource_app/api/token_storage.dart';

/// Base URL of the API gateway. Override per environment with:
/// `--dart-define=API_BASE_URL=https://api.example.com/api/v1`.
const String kApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080/api/v1',
);

/// Gateway origin without the `/api/v1` suffix (public routes like `/languages`).
String get kGatewayOrigin {
  final trimmed = kApiBaseUrl.endsWith('/')
      ? kApiBaseUrl.substring(0, kApiBaseUrl.length - 1)
      : kApiBaseUrl;
  const suffix = '/api/v1';
  if (trimmed.endsWith(suffix)) {
    return trimmed.substring(0, trimmed.length - suffix.length);
  }
  return trimmed;
}

/// Feature flag for the migration off Firebase: when true, auth (login,
/// register, session, role routing) goes through the REST backend instead of
/// Firebase. Off by default so the Firebase-backed screens keep working until
/// they are migrated. Enable with `--dart-define=USE_API_AUTH=true`.
const bool kUseApiAuth = bool.fromEnvironment(
  'USE_API_AUTH',
  defaultValue: false,
);

/// Central HTTP client for the Go backend.
///
/// Responsibilities:
/// - attach `Authorization: Bearer <access>` to every request;
/// - on 401, refresh the token pair once (shared across concurrent failures)
///   via `POST /auth/refresh` and retry the original request;
/// - on refresh failure, clear tokens and invoke [onUnauthorized] so the app
///   can route back to login.
class ApiClient {
  final Dio dio;
  final TokenStorage tokens;

  /// Called when the session is unrecoverable (refresh failed/absent).
  final void Function()? onUnauthorized;

  /// Bare client (no interceptors) used to perform the refresh call itself,
  /// avoiding recursion through the 401 handler.
  final Dio _refreshDio;

  /// Shared in-flight refresh so parallel 401s trigger a single refresh.
  Future<bool>? _refreshing;

  ApiClient({
    required this.tokens,
    String baseUrl = kApiBaseUrl,
    this.onUnauthorized,
  })  : dio = Dio(BaseOptions(
          baseUrl: baseUrl,
          contentType: Headers.jsonContentType,
        )),
        _refreshDio = Dio(BaseOptions(baseUrl: baseUrl)) {
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: _onRequest,
      onError: _onError,
    ));
  }

  Future<void> _onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await tokens.readAccessToken();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  Future<void> _onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final response = err.response;
    final isAuthCall = err.requestOptions.path.contains('/auth/');
    // Only handle a genuine 401 that we haven't already retried.
    if (response?.statusCode != 401 ||
        isAuthCall ||
        err.requestOptions.extra['retried'] == true) {
      return handler.next(err);
    }

    final refreshed = await _refreshTokens();
    if (!refreshed) {
      await tokens.clear();
      onUnauthorized?.call();
      return handler.next(err);
    }

    // Retry the original request once with the new access token.
    try {
      final newToken = await tokens.readAccessToken();
      final options = err.requestOptions
        ..extra['retried'] = true
        ..headers['Authorization'] = 'Bearer $newToken';
      final retried = await dio.fetch<dynamic>(options);
      return handler.resolve(retried);
    } on DioException catch (e) {
      return handler.next(e);
    }
  }

  /// Refreshes the token pair, deduplicating concurrent callers.
  Future<bool> _refreshTokens() {
    return _refreshing ??= _doRefresh().whenComplete(() => _refreshing = null);
  }

  Future<bool> _doRefresh() async {
    final refreshToken = await tokens.readRefreshToken();
    if (refreshToken == null) return false;
    try {
      final res = await _refreshDio.post<Map<String, dynamic>>(
        '/auth/refresh',
        data: RefreshRequest(refreshToken: refreshToken).toJson(),
      );
      final pair = TokenPair.fromJson(res.data!);
      await tokens.saveTokens(
        accessToken: pair.accessToken,
        refreshToken: pair.refreshToken,
      );
      return true;
    } on DioException {
      return false;
    }
  }
}
