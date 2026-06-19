import 'package:get/get.dart';

import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/api/repositories/auth_repository.dart';
import 'package:online_cource_app/api/token_storage.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';

/// Registers the REST API stack with GetX.
///
/// Called from `main()` only when [kUseApiAuth] is enabled, so the Firebase
/// path incurs no network calls. Order matters: storage → client → repo →
/// controller (which restores the session on init).
void setupApiDependencies() {
  final tokens = Get.put(TokenStorage(), permanent: true);

  final client = Get.put(
    ApiClient(
      tokens: tokens,
      onUnauthorized: () {
        if (Get.isRegistered<ApiAuthController>()) {
          Get.find<ApiAuthController>().handleSessionExpired();
        }
      },
    ),
    permanent: true,
  );

  final repo = Get.put(AuthRepository(client), permanent: true);
  Get.put(ApiAuthController(repo), permanent: true);
}
