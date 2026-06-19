import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/auth_gate.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';
import 'package:online_cource_app/theme/app_theme.dart';

class ProfileTab extends StatelessWidget {
  const ProfileTab({super.key});

  @override
  Widget build(BuildContext context) {
    final auth = Get.find<ApiAuthController>();
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Профиль', style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 24),
            Obx(() {
              final u = auth.user.value;
              final roleLabel = auth.isTeacher ? 'Учитель' : 'Ученик';
              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    u?.displayName ?? u?.email ?? '',
                    style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 4),
                  Text(u?.email ?? '', style: TextStyle(color: AppTheme.secondaryTextColor)),
                  const SizedBox(height: 8),
                  Chip(label: Text(roleLabel)),
                ],
              );
            }),
            const Spacer(),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                onPressed: () async {
                  await auth.logout();
                  Get.offAll(() => const AuthGate());
                },
                child: const Text('Выйти'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
