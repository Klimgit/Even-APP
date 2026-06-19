import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/controllers/api_auth_controller.dart';
import 'package:online_cource_app/Login/login_page.dart';
import 'package:online_cource_app/features/student/student_shell.dart';
import 'package:online_cource_app/features/teacher/api_teacher_dashboard.dart';
import 'package:online_cource_app/Utils/dialouge_utils.dart';
import 'package:online_cource_app/Utils/toast_messages.dart';
import 'package:online_cource_app/controllers/auth_controller.dart';
import 'package:online_cource_app/theme/app_theme.dart';

class SignUpPage extends StatefulWidget {
  const SignUpPage({super.key});

  @override
  _SignUpPageState createState() => _SignUpPageState();
}

class _SignUpPageState extends State<SignUpPage>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _animation;
  final TextEditingController _nameController = TextEditingController();
  final TextEditingController _emailController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _isPasswordVisible = false;
  bool _isLoading = false;
  String _role = 'student';

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
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _onSignUp() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isLoading = true;
    });

    // REST-backed sign-up (migration off Firebase).
    if (kUseApiAuth) {
      await _signUpViaApi();
      return;
    }

    try {
      showLoadingDialouge(context, 'Регистрация…');
      await _firebaseAuth!.signUpNewUsers(context, _emailController.text.trim(),
          _passwordController.text.trim(), _nameController.text.trim());
      Get.back();
      showSuccessToast(context, 'Регистрация успешна');
      Get.to(() => const LoginPage());
    } catch (e) {
      Get.back();
      showErrorToast(context, 'Ошибка: ${e.toString()}');
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> _signUpViaApi() async {
    final apiAuth = Get.find<ApiAuthController>();
    try {
      showLoadingDialouge(context, 'Регистрация…');
      // Role selected by user on the form.
      final error = await apiAuth.register(
        email: _emailController.text.trim(),
        password: _passwordController.text.trim(),
        displayName: _nameController.text.trim(),
        role: _role,
      );
      Get.back();
      if (error != null) {
        showErrorToast(context, error);
      } else {
        showSuccessToast(context, 'Регистрация успешна');
        if (apiAuth.isTeacher) {
          Get.offAll(() => const ApiTeacherDashboard());
        } else {
          Get.offAll(() => const StudentShell());
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
                            Icons.person_add,
                            size: 60,
                            color: AppTheme.primaryColor,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 40),
                    Text(
                      'Регистрация',
                      style:
                          Theme.of(context).textTheme.displayMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                                color: AppTheme.textColor,
                              ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Создайте аккаунт для начала обучения',
                      style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                            color: AppTheme.secondaryTextColor,
                          ),
                    ),
                    const SizedBox(height: 40),
                    TextFormField(
                      controller: _nameController,
                      decoration: InputDecoration(
                        labelText: 'Имя',
                        hintText: 'Введите имя',
                        prefixIcon: Icon(Icons.person,
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
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Введите имя';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 20),
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
                    const SizedBox(height: 20),
                    Text(
                      'Я регистрируюсь как',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 12),
                    SegmentedButton<String>(
                      segments: const [
                        ButtonSegment(
                          value: 'student',
                          label: Text('Ученик'),
                          icon: Icon(Icons.school_outlined),
                        ),
                        ButtonSegment(
                          value: 'teacher',
                          label: Text('Учитель'),
                          icon: Icon(Icons.person_outline),
                        ),
                      ],
                      selected: {_role},
                      onSelectionChanged: (values) {
                        setState(() => _role = values.first);
                      },
                    ),
                    const SizedBox(height: 32),
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
                        onPressed: _isLoading ? null : _onSignUp,
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
                                'Зарегистрироваться',
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
                          'Уже есть аккаунт? ',
                          style: TextStyle(
                            color: AppTheme.secondaryTextColor,
                          ),
                        ),
                        TextButton(
                          onPressed: () {
                            Get.to(() => const LoginPage());
                          },
                          style: TextButton.styleFrom(
                            foregroundColor: AppTheme.primaryColor,
                            padding: EdgeInsets.zero,
                            minimumSize: const Size(0, 30),
                          ),
                          child: const Text(
                            'Войти',
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
