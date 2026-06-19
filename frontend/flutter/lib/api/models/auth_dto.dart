import 'package:online_cource_app/api/models/user_dto.dart';

/// Request/response DTOs for the `/auth/*` endpoints (see `DTO.md`, `API.md`).
/// Manual JSON mapping (no codegen). Field names match the wire format
/// (snake_case) exactly.

/// `POST /auth/register` body.
class RegisterRequest {
  final String email;
  final String password;
  final String? displayName;

  /// "student" | "teacher" (optional; backend defaults to "student").
  final String? role;

  const RegisterRequest({
    required this.email,
    required this.password,
    this.displayName,
    this.role,
  });

  Map<String, dynamic> toJson() => {
        'email': email,
        'password': password,
        if (displayName != null) 'display_name': displayName,
        if (role != null) 'role': role,
      };
}

/// `POST /auth/login` body.
class LoginRequest {
  final String email;
  final String password;

  const LoginRequest({required this.email, required this.password});

  Map<String, dynamic> toJson() => {
        'email': email,
        'password': password,
      };
}

/// `POST /auth/refresh` body.
class RefreshRequest {
  final String refreshToken;

  const RefreshRequest({required this.refreshToken});

  Map<String, dynamic> toJson() => {
        'refresh_token': refreshToken,
      };
}

/// Response of `POST /auth/register` (201) and `POST /auth/login` (200).
class AuthResponse {
  final String accessToken;
  final String refreshToken;
  final UserDto user;

  const AuthResponse({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      accessToken: json['access_token'] as String,
      refreshToken: json['refresh_token'] as String,
      user: UserDto.fromJson(json['user'] as Map<String, dynamic>),
    );
  }
}

/// Response of `POST /auth/refresh` (200) — a fresh token pair, no user.
class TokenPair {
  final String accessToken;
  final String refreshToken;

  const TokenPair({required this.accessToken, required this.refreshToken});

  factory TokenPair.fromJson(Map<String, dynamic> json) {
    return TokenPair(
      accessToken: json['access_token'] as String,
      refreshToken: json['refresh_token'] as String,
    );
  }
}
