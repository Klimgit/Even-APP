import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';
import 'package:online_cource_app/SignUp/sign_up_scree.dart';
import 'package:online_cource_app/Utils/dialouge_utils.dart';
import 'package:online_cource_app/Utils/toast_messages.dart';
import 'package:online_cource_app/controllers/auth_controller.dart';
import 'package:online_cource_app/features/student/student_shell.dart';
import 'package:online_cource_app/features/teacher/api_teacher_dashboard.dart';
import 'package:online_cource_app/navigation/main_navigation.dart';
import 'package:online_cource_app/teacher/teacher_dashboard.dart';
import 'package:online_cource_app/theme/app_theme.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  _LoginPageState createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _animation;
  final TextEditingController _emailController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _isPasswordVisible = false;
  bool _isLoading = false;

  AuthController? get _firebaseAuth =>
      kUseApiAuth ? null : Get.find<AuthController>();

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1000),
    );

    _animation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(_animationController);

    _animationController.forward();
  }

  @override
  void dispose() {
    _animationController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _login() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isLoading = true;
    });

    String email = _emailController.text.trim();
    String password = _passwordController.text.trim();

    // REST-backed sign-in (migration off Firebase).
    if (kUseApiAuth) {
      await _loginViaApi(email, password);
      return;
    }

    try {
      showLoadingDialouge(context, 'Вход…');
      final user = await _firebaseAuth!.signInUsers(context, email, password);

      if (user == null) {
        Get.back();
        showErrorToast(context, 'Неверный email или пароль');
      } else {
        Get.back();
        // Ensure the Firestore user document is loaded before routing so the
        // role check below is accurate, then send to the right home screen.
        await _firebaseAuth!.fetchUserData();
        if (_firebaseAuth!.isTeacher) {
          Get.offAll(() => const TeacherDashboard());
        } else {
          Get.offAll(() => const MainNavigationScreen());
        }
      }
    } catch (e) {
      Get.back();
      showErrorToast(context, 'Ошибка: ${e.toString()}');
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> _loginViaApi(String email, String password) async {
    final apiAuth = Get.find<ApiAuthController>();
    try {
      showLoadingDialouge(context, 'Вход…');
      final error = await apiAuth.login(email, password);
      Get.back();
      if (error != null) {
        showErrorToast(context, error);
      } else if (apiAuth.isTeacher) {
        Get.offAll(() => const ApiTeacherDashboard());
      } else {
        Get.offAll(() => const StudentShell());
      }
    } catch (e) {
      Get.back();
      showErrorToast(context, 'Ошибка: ${e.toString()}');
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      body: SafeArea(
        child: SingleChildScrollView(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: FadeTransition(
              opacity: _animation,
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const SizedBox(height: 40),
                    Center(
                      child: Container(
                        width: 100,
                        height: 100,
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: [
                            BoxShadow(
                              color: AppTheme.primaryColor.withOpacity(0.2),
                              blurRadius: 20,
                              offset: const Offset(0, 10),
                            ),
                          ],
                        ),
                        child: Center(
                          child: Icon(
                            Icons.school,
                            size: 60,
                            color: AppTheme.primaryColor,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 40),
                    Text(
                      'С возвращением!',
                      style:
                          Theme.of(context).textTheme.displayMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                                color: AppTheme.textColor,
                              ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Войдите, чтобы продолжить обучение',
                      style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                            color: AppTheme.secondaryTextColor,
                          ),
                    ),
                    const SizedBox(height: 40),
                    TextFormField(
                      controller: _emailController,
                      decoration: InputDecoration(
                        labelText: 'Email',
                        hintText: 'Введите email',
                        prefixIcon: Icon(Icons.email,
                            color: AppTheme.secondaryTextColor),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide(color: AppTheme.dividerColor),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide(color: AppTheme.dividerColor),
                        ),
                        focusedBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide(color: AppTheme.primaryColor),
                        ),
                        filled: true,
                        fillColor: Colors.white,
                        contentPadding:
                            const EdgeInsets.symmetric(vertical: 16),
                      ),
                      keyboardType: TextInputType.emailAddress,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Введите email';
                        }
                        if (!GetUtils.isEmail(value)) {
                          return 'Некорректный email';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 20),
                    TextFormField(
                      controller: _passwordController,
                      obscureText: !_isPasswordVisible,
                      decoration: InputDecoration(
                        labelText: 'Пароль',
                        hintText: 'Введите пароль',
                        prefixIcon: Icon(Icons.lock,
                            color: AppTheme.secondaryTextColor),
                        suffixIcon: IconButton(
                          icon: Icon(
                            _isPasswordVisible
                                ? Icons.visibility
                                : Icons.visibility_off,
                            color: AppTheme.secondaryTextColor,
                          ),
                          onPressed: () {
                            setState(() {
                              _isPasswordVisible = !_isPasswordVisible;
                            });
                          },
                        ),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide(color: AppTheme.dividerColor),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide(color: AppTheme.dividerColor),
                        ),
                        focusedBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide(color: AppTheme.primaryColor),
                        ),
                        filled: true,
                        fillColor: Colors.white,
                        contentPadding:
                            const EdgeInsets.symmetric(vertical: 16),
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Введите пароль';
                        }
                        if (value.length < 8) {
                          return 'Минимум 8 символов';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 12),
                    Align(
                      alignment: Alignment.centerRight,
                      child: TextButton(
                        onPressed: () {
                          // Forgot password functionality
                        },
                        style: TextButton.styleFrom(
                          foregroundColor: AppTheme.primaryColor,
                        ),
                        child: const Text('Забыли пароль?'),
                      ),
                    ),
                    const SizedBox(height: 24),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppTheme.primaryColor,
                          foregroundColor: Colors.white,
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(30),
                          ),
                        ),
                        onPressed: _isLoading ? null : _login,
                        child: _isLoading
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                  color: Colors.white,
                                  strokeWidth: 3,
                                ),
                              )
                            : const Text(
                                'Войти',
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                      ),
                    ),
                    const SizedBox(height: 24),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          'Нет аккаунта? ',
                          style: TextStyle(
                            color: AppTheme.secondaryTextColor,
                          ),
                        ),
                        TextButton(
                          onPressed: () {
                            Get.to(() => const SignUpPage());
                          },
                          style: TextButton.styleFrom(
                            foregroundColor: AppTheme.primaryColor,
                            padding: EdgeInsets.zero,
                            minimumSize: const Size(0, 30),
                          ),
                          child: const Text(
                            'Регистрация',
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
