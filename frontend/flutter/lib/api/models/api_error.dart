/// Standard error envelope returned by the backend:
/// `{ error, message, details? }` (see `API.md` → "Общие соглашения").
class ApiError {
  /// Machine-readable error code, e.g. "validation", "credentials".
  final String error;

  /// Human-readable message.
  final String message;

  /// Optional structured details (e.g. per-field validation errors).
  final Map<String, dynamic>? details;

  const ApiError({
    required this.error,
    required this.message,
    this.details,
  });

  factory ApiError.fromJson(Map<String, dynamic> json) {
    return ApiError(
      error: (json['error'] as String?) ?? 'unknown',
      message: (json['message'] as String?) ?? 'Unknown error',
      details: json['details'] as Map<String, dynamic>?,
    );
  }

  @override
  String toString() => 'ApiError($error): $message';
}
