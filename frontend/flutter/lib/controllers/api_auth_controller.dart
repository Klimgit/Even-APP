import 'package:get/get.dart';

import 'package:online_cource_app/api/models/models.dart';
import 'package:online_cource_app/api/repositories/auth_repository.dart';

/// Auth state backed by the REST backend (used when `kUseApiAuth` is on).
///
/// Mirrors the role-routing role of the Firebase [AuthController] but sources
/// the user from JWT/`/auth/me` instead of Firestore.
class ApiAuthController extends GetxController {
  final AuthRepository _repo;

  ApiAuthController(this._repo);

  /// Current user, or null when signed out.
  final Rxn<UserDto> user = Rxn<UserDto>();

  /// A login/register request is in flight.
  final RxBool isLoading = false.obs;

  /// Initial "do we have a valid session?" check is running.
  final RxBool isRestoring = true.obs;

  bool get isLoggedIn => user.value != null;
  bool get isAdmin => user.value?.isAdmin ?? false;
  bool get isTeacher => (user.value?.isTeacher ?? false) || isAdmin;

  @override
  void onInit() {
    super.onInit();
    restoreSession();
  }

  /// Restores the session on startup: if tokens exist, fetch the user; on any
  /// failure (including a failed refresh) drop to signed-out.
  Future<void> restoreSession() async {
    isRestoring.value = true;
    try {
      if (await _repo.isLoggedIn()) {
        user.value = await _repo.me();
      }
    } catch (_) {
      await _repo.logout();
      user.value = null;
    } finally {
      isRestoring.value = false;
    }
  }

  /// Returns null on success, or a user-facing error message on failure.
  Future<String?> login(String email, String password) async {
    return _run(() async {
      final auth = await _repo.login(
        LoginRequest(email: email, password: password),
      );
      user.value = auth.user;
    });
  }

  /// Returns null on success, or a user-facing error message on failure.
  Future<String?> register({
    required String email,
    required String password,
    String? displayName,
    String? role,
  }) async {
    return _run(() async {
      final auth = await _repo.register(
        RegisterRequest(
          email: email,
          password: password,
          displayName: displayName,
          role: role,
        ),
      );
      user.value = auth.user;
    });
  }

  Future<void> logout() async {
    await _repo.logout();
    user.value = null;
  }

  /// Invoked by [ApiClient] when a token refresh fails — the session is gone.
  void handleSessionExpired() {
    user.value = null;
  }

  Future<String?> _run(Future<void> Function() action) async {
    isLoading.value = true;
    try {
      await action();
      return null;
    } on ApiException catch (e) {
      return e.message;
    } catch (_) {
      return 'Network error. Please try again.';
    } finally {
      isLoading.value = false;
    }
  }
}
