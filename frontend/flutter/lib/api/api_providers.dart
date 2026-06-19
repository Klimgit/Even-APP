import 'package:get/get.dart';

import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/api/repositories/auth_repository.dart';
import 'package:online_cource_app/api/repositories/content_repository.dart';
import 'package:online_cource_app/api/repositories/learning_repository.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';
import 'package:online_cource_app/api/repositories/media_repository.dart';
import 'package:online_cource_app/api/token_storage.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';

/// Registers the REST API stack with GetX.
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

  final authRepo = Get.put(AuthRepository(client), permanent: true);
  Get.put(ApiAuthController(authRepo), permanent: true);
  Get.put(ContentRepository(client), permanent: true);
  Get.put(LearningRepository(client), permanent: true);
  Get.put(LexiconRepository(client), permanent: true);
  Get.put(MediaRepository(client), permanent: true);
}
