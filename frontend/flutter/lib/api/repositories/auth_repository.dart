import 'package:dio/dio.dart';

import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/api/models/models.dart';

/// Thrown for backend error responses, wrapping the `{error,message,details}`
/// envelope so callers can show [message] and branch on [statusCode].
class ApiException implements Exception {
  final int? statusCode;
  final ApiError error;

  ApiException(this.statusCode, this.error);

  String get message => error.message;

  @override
  String toString() => 'ApiException($statusCode): ${error.message}';
}

/// Auth endpoints (`/auth/*`) — the only fully-implemented backend chain.
///
/// On successful register/login the token pair is persisted via the client's
/// [TokenStorage]; subsequent authenticated calls are signed automatically by
/// the [ApiClient] interceptor.
class AuthRepository {
  final ApiClient _client;

  AuthRepository(this._client);

  Future<AuthResponse> register(RegisterRequest request) async {
    final res = await _post('/auth/register', request.toJson());
    final auth = AuthResponse.fromJson(res);
    await _saveSession(auth);
    return auth;
  }

  Future<AuthResponse> login(LoginRequest request) async {
    final res = await _post('/auth/login', request.toJson());
    final auth = AuthResponse.fromJson(res);
    await _saveSession(auth);
    return auth;
  }

  /// Current user from the stored access token. Throws [ApiException] (401)
  /// if the session is invalid and could not be refreshed.
  Future<UserDto> me() async {
    try {
      final res = await _client.dio.get<Map<String, dynamic>>('/auth/me');
      return UserDto.fromJson(res.data!);
    } on DioException catch (e) {
      throw _toApiException(e);
    }
  }

  Future<void> logout() => _client.tokens.clear();

  Future<bool> isLoggedIn() => _client.tokens.hasTokens();

  // ---------------------------------------------------------------------------

  Future<void> _saveSession(AuthResponse auth) {
    return _client.tokens.saveTokens(
      accessToken: auth.accessToken,
      refreshToken: auth.refreshToken,
    );
  }

  Future<Map<String, dynamic>> _post(
      String path, Map<String, dynamic> body) async {
    try {
      final res =
          await _client.dio.post<Map<String, dynamic>>(path, data: body);
      return res.data!;
    } on DioException catch (e) {
      throw _toApiException(e);
    }
  }

  ApiException _toApiException(DioException e) {
    final data = e.response?.data;
    final apiError = data is Map<String, dynamic>
        ? ApiError.fromJson(data)
        : ApiError(
            error: 'network',
            message: e.message ?? 'Network error',
          );
    return ApiException(e.response?.statusCode, apiError);
  }
}
