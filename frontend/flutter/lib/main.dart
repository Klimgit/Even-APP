import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:get/get.dart';
import 'package:flutter_easyloading/flutter_easyloading.dart';

import 'package:online_cource_app/auth_gate.dart';
import 'package:online_cource_app/controllers/auth_controller.dart';
import 'package:online_cource_app/firebase_options.dart';
import 'package:online_cource_app/theme/app_theme.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Lock phones to portrait, but allow all orientations on tablets/desktop
  // (>= 600dp shortest side) so the teacher panel can run in landscape.
  final view = WidgetsBinding.instance.platformDispatcher.views.first;
  final shortestSide =
      (view.physicalSize / view.devicePixelRatio).shortestSide;
  await SystemChrome.setPreferredOrientations(
    shortestSide >= 600
        ? DeviceOrientation.values
        : [DeviceOrientation.portraitUp, DeviceOrientation.portraitDown],
  );

  // Initialize Firebase
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  // Register controllers
  Get.lazyPut(() => AuthController(), fenix: true);

  // Configure EasyLoading
  configureEasyLoading();

  runApp(const MyApp());
}

void configureEasyLoading() {
  EasyLoading.instance
    ..displayDuration = const Duration(milliseconds: 2000)
    ..indicatorType = EasyLoadingIndicatorType.fadingCircle
    ..loadingStyle = EasyLoadingStyle.custom
    ..indicatorSize = 45.0
    ..radius = 16.0
    ..progressColor = AppTheme.primaryColor
    ..backgroundColor = Colors.white
    ..indicatorColor = AppTheme.primaryColor
    ..textColor = AppTheme.textColor
    ..maskColor = Colors.black.withOpacity(0.5)
    ..userInteractions = false
    ..dismissOnTap = false;
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return GetMaterialApp(
      title: 'E-Learning App',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme(),
      home: const AuthGate(),
      builder: EasyLoading.init(),
      defaultTransition: Transition.fadeIn,
      transitionDuration: const Duration(milliseconds: 200),
    );
  }
}
