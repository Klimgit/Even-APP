/// Mirrors `UserDTO` from the backend (`DTO.md`).
///
/// ```
/// UserDTO {
///   id: string
///   email: string
///   display_name?: string
///   role: "student" | "teacher"
///   is_admin: boolean
///   created_at: string  // ISO8601
/// }
/// ```
class UserDto {
  final String id;
  final String email;
  final String? displayName;

  /// "student" | "teacher". Kept as a raw string to tolerate unknown values;
  /// use [isTeacher] / [isStudent] for checks.
  final String role;
  final bool isAdmin;
  final DateTime createdAt;

  const UserDto({
    required this.id,
    required this.email,
    required this.displayName,
    required this.role,
    required this.isAdmin,
    required this.createdAt,
  });

  bool get isStudent => role == 'student';
  bool get isTeacher => role == 'teacher';

  factory UserDto.fromJson(Map<String, dynamic> json) {
    return UserDto(
      id: json['id'] as String,
      email: json['email'] as String,
      displayName: json['display_name'] as String?,
      role: (json['role'] as String?) ?? 'student',
      isAdmin: (json['is_admin'] as bool?) ?? false,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'email': email,
        'display_name': displayName,
        'role': role,
        'is_admin': isAdmin,
        'created_at': createdAt.toIso8601String(),
      };
}
