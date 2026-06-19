import 'package:flutter/material.dart';
import 'package:cloud_firestore/cloud_firestore.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:get/get.dart';
import 'package:online_cource_app/Courses/enhanced_course_details.dart';
import 'package:online_cource_app/Detail/course_detail.dart';
import 'package:online_cource_app/Model/course_model.dart';
import 'package:online_cource_app/Model/model.dart.dart';
import 'package:online_cource_app/Courses/course_progress_service.dart';
import 'package:online_cource_app/Courses/student_course_screen.dart';
import 'package:online_cource_app/theme/app_theme.dart';
import 'package:online_cource_app/exercises/demo_lesson.dart';
import 'package:online_cource_app/exercises/lesson_runner.dart';
import 'package:online_cource_app/Utils/custom_drawer.dart';
import 'package:online_cource_app/constants.dart';
import 'package:flutter_staggered_animations/flutter_staggered_animations.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:shimmer/shimmer.dart';

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key});

  @override
  _MyHomePageState createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage>
    with SingleTickerProviderStateMixin {
  GlobalKey<ScaffoldState> scaffoldKey = GlobalKey<ScaffoldState>();
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocus = FocusNode();
  bool _isSearching = false;
  bool _isLoading = true;
  late AnimationController _animationController;

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );

    // Simulate loading data
    Future.delayed(const Duration(seconds: 1), () {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    _searchFocus.dispose();
    _animationController.dispose();
    super.dispose();
  }

  String _getGreeting() {
    var hour = DateTime.now().hour;
    if (hour < 12) {
      return 'Good Morning';
    }
    if (hour < 17) {
      return 'Good Afternoon';
    }
    return 'Good Evening';
  }

  String _getUserName() {
    final user = FirebaseAuth.instance.currentUser;
    if (user != null) {
      if (user.displayName != null && user.displayName!.isNotEmpty) {
        return user.displayName!.split(' ')[0];
      } else {
        return 'Student';
      }
    }
    return 'Student';
  }

  Widget _buildSearchField() {
    return TextField(
      controller: _searchController,
      focusNode: _searchFocus,
      decoration: InputDecoration(
        hintText: 'Search for courses...',
        hintStyle: TextStyle(color: AppTheme.secondaryTextColor),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: AppTheme.dividerColor),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: AppTheme.dividerColor),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: AppTheme.textColor),
        ),
        filled: true,
        fillColor: Colors.white,
        prefixIcon: Icon(Icons.search, color: AppTheme.secondaryTextColor),
        contentPadding: const EdgeInsets.symmetric(vertical: 0),
      ),
      style: TextStyle(color: AppTheme.textColor),
      onSubmitted: (value) {
        // Implement search functionality
      },
    );
  }

  Widget _buildFeaturedBanner(BuildContext context) {
    return Container(
      height: 180,
      decoration: BoxDecoration(
        color: AppTheme.accentColor.withOpacity(0.12),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.dividerColor),
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Expanded(
              flex: 3,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    "New Collection",
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                          color: AppTheme.secondaryTextColor,
                        ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    "Start Learning Today",
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          color: AppTheme.textColor,
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () {},
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppTheme.accentColor,
                      foregroundColor: Colors.white,
                      elevation: 0,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 20, vertical: 12),
                    ),
                    child: const Text("Explore Now"),
                  ),
                ],
              ),
            ),
            Expanded(
              flex: 2,
              child: Image.asset(
                'assets/learning.png',
                errorBuilder: (context, error, stackTrace) {
                  return Icon(
                    Icons.school,
                    size: 80,
                    color: AppTheme.accentColor.withOpacity(0.7),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionHeader(String title, Function() onSeeAll) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          title,
          style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
        ),
        TextButton(
          onPressed: onSeeAll,
          child: Text(
            "See All",
            style: TextStyle(
              color: AppTheme.accentColor,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildContinueLearningSection() {
    return SizedBox(
      height: 170,
      child: StreamBuilder<List<CourseProgress>>(
        stream: CourseProgressService.startedCoursesStream(),
        builder: (context, snapshot) {
          if (!snapshot.hasData) {
            return _buildLoadingCourseCards();
          }
          final started = snapshot.data!;
          if (started.isEmpty) {
            return _notEnrolledCard();
          }
          return ListView.builder(
            scrollDirection: Axis.horizontal,
            itemCount: started.length,
            itemBuilder: (context, index) => _progressCard(started[index]),
          );
        },
      ),
    );
  }

  Widget _notEnrolledCard() {
    return Align(
      alignment: Alignment.centerLeft,
      child: Container(
        width: 280,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppTheme.dividerColor),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.school_outlined,
                size: 36, color: AppTheme.secondaryTextColor),
            const SizedBox(height: 12),
            Text(
              'Пока вы не записаны на курсы',
              textAlign: TextAlign.center,
              style: TextStyle(
                color: AppTheme.secondaryTextColor,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _progressCard(CourseProgress p) {
    final percent = (p.percent * 100).round();
    return GestureDetector(
      onTap: () => Get.to(() => StudentCourseScreen(
            courseId: p.courseId,
            courseName: p.name,
            coverColor: p.coverColor,
          )),
      child: Container(
        margin: const EdgeInsets.only(right: 16),
        width: 280,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppTheme.dividerColor),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                    color: Color(p.coverColor),
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    p.name,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: Theme.of(context)
                        .textTheme
                        .titleMedium
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
            const Spacer(),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: p.percent,
                minHeight: 6,
                backgroundColor: AppTheme.dividerColor,
                valueColor:
                    const AlwaysStoppedAnimation<Color>(AppTheme.accentColor),
              ),
            ),
            const SizedBox(height: 8),
            Text(
              '$percent% • ${p.completedLessons.length}/${p.totalLessons} lessons',
              style: TextStyle(color: AppTheme.secondaryTextColor),
            ),
          ],
        ),
      ),
    );
  }

  Widget _popularCourseCard(String id, Map<String, dynamic> data) {
    final name = data['name'] as String? ?? 'Untitled';
    final colorValue = data['coverColor'] as int? ?? 0xFF9BE8B4;
    return GestureDetector(
      onTap: () => Get.to(() => StudentCourseScreen(
            courseId: id,
            courseName: name,
            coverColor: colorValue,
          )),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppTheme.dividerColor),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Text(
                name,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.bold),
                maxLines: 3,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 12),
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: Color(colorValue),
                borderRadius: BorderRadius.circular(8),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLoadingCourseCards() {
    return SizedBox(
      height: 200,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        itemCount: 3,
        itemBuilder: (context, index) {
          return Container(
            margin: const EdgeInsets.only(right: 16),
            width: 280,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
            ),
            child: Shimmer.fromColors(
              baseColor: Colors.grey[300]!,
              highlightColor: Colors.grey[100]!,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 100,
                      height: 16,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    const SizedBox(height: 16),
                    Container(
                      width: 200,
                      height: 24,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    const SizedBox(height: 30),
                    Container(
                      width: double.infinity,
                      height: 8,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(4),
                      ),
                    ),
                    const SizedBox(height: 16),
                    Container(
                      width: 80,
                      height: 16,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    const Spacer(),
                    Container(
                      width: 100,
                      height: 36,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(18),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  Widget availableCourses(BuildContext context, int index) {
    return Container(
      margin: const EdgeInsets.only(bottom: 20),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.dividerColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: Image.network(
              onlineCourceOne[index]['img'],
              fit: BoxFit.cover,
              width: double.infinity,
              height: 150,
              errorBuilder: (context, error, stackTrace) => Container(
                height: 150,
                color: AppTheme.dividerColor,
                child: Icon(
                  Icons.image_not_supported,
                  color: AppTheme.secondaryTextColor,
                  size: 50,
                ),
              ),
            ),
          ),
          const SizedBox(height: 15),
          Text(
            onlineCourceOne[index]['title'],
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 15),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Icon(
                    Icons.play_circle_filled,
                    color: AppTheme.accentColor,
                    size: 16,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    "${onlineCourceOne[index]['session']} lessons",
                    style: TextStyle(
                      color: AppTheme.secondaryTextColor,
                      fontSize: 13,
                    ),
                  ),
                ],
              ),
              Text(
                "৳${onlineCourceOne[index]['price']}",
                style: TextStyle(
                  color: AppTheme.accentColor,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      key: scaffoldKey,
      drawer: CustomDrawer(),
      backgroundColor: AppTheme.backgroundColor,
      body: SafeArea(
        child: NestedScrollView(
          headerSliverBuilder: (context, innerBoxIsScrolled) {
            return [
              SliverAppBar(
                floating: true,
                snap: true,
                pinned: false,
                backgroundColor: AppTheme.backgroundColor,
                automaticallyImplyLeading: false,
                title: _isSearching ? _buildSearchField() : null,
                leading: IconButton(
                  icon: Icon(Icons.menu, color: AppTheme.textColor),
                  onPressed: () {
                    scaffoldKey.currentState?.openDrawer();
                  },
                ),
                actions: [
                  // TEMPORARY: preview the Duolingo-style exercise demo lesson.
                  if (!_isSearching)
                    IconButton(
                      icon: Icon(Icons.science_outlined,
                          color: AppTheme.accentColor),
                      tooltip: 'Demo lesson',
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) => LessonRunner(
                              title: 'Demo lesson',
                              exercises: demoExercises(),
                            ),
                          ),
                        );
                      },
                    ),
                  if (!_isSearching)
                    IconButton(
                      icon: Icon(Icons.search, color: AppTheme.textColor),
                      onPressed: () {
                        setState(() {
                          _isSearching = true;
                        });
                      },
                    ),
                  if (_isSearching)
                    IconButton(
                      icon: Icon(Icons.close, color: AppTheme.textColor),
                      onPressed: () {
                        setState(() {
                          _isSearching = false;
                          _searchController.clear();
                        });
                      },
                    ),
                ],
              )
            ];
          },
          body: RefreshIndicator(
            color: AppTheme.primaryColor,
            onRefresh: () async {
              // Refresh data
              setState(() {
                _isLoading = true;
              });
              await Future.delayed(const Duration(seconds: 1));
              if (mounted) {
                setState(() {
                  _isLoading = false;
                });
              }
            },
            child: ListView(
              padding: const EdgeInsets.all(20),
              children: AnimationConfiguration.toStaggeredList(
                duration: const Duration(milliseconds: 300),
                childAnimationBuilder: (widget) => SlideAnimation(
                  horizontalOffset: 50.0,
                  child: FadeInAnimation(
                    child: widget,
                  ),
                ),
                children: [
                  // Welcome section
                  if (!_isSearching) ...[
                    const SizedBox(height: 10),
                    Text(
                      "${_getGreeting()},",
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                            color: AppTheme.secondaryTextColor,
                          ),
                    ),
                    Text(
                      _getUserName(),
                      style:
                          Theme.of(context).textTheme.displayMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                    ),
                    const SizedBox(height: 16),
                  ],

                  // Featured banner
                  if (!_isSearching) _buildFeaturedBanner(context),

                  // Continue learning section
                  if (!_isSearching) ...[
                    const SizedBox(height: 30),
                    _buildSectionHeader("Continue Learning", () {
                      // Navigate to enrolled courses
                    }),
                    const SizedBox(height: 16),
                    _buildContinueLearningSection(),
                  ],

                  // Categories section
                  const SizedBox(height: 30),
                  _buildSectionHeader("Categories", () {
                    // Navigate to categories
                  }),
                  const SizedBox(height: 16),

                  // Popular courses section
                  const SizedBox(height: 30),
                  _buildSectionHeader("Popular Courses", () {
                    // Navigate to all courses
                  }),
                  const SizedBox(height: 16),

                  // Recommended courses — only real teacher-created courses.
                  StreamBuilder<QuerySnapshot<Map<String, dynamic>>>(
                    stream: FirebaseFirestore.instance
                        .collection('classes')
                        .snapshots(),
                    builder: (context, snapshot) {
                      if (!snapshot.hasData) {
                        return const Padding(
                          padding: EdgeInsets.all(20),
                          child: Center(child: CircularProgressIndicator()),
                        );
                      }
                      final docs = snapshot.data!.docs;
                      if (docs.isEmpty) {
                        return Padding(
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          child: Text('No courses yet',
                              style: TextStyle(
                                  color: AppTheme.secondaryTextColor)),
                        );
                      }
                      final shown =
                          docs.length > 6 ? docs.sublist(0, 6) : docs;
                      return LayoutBuilder(
                        builder: (context, constraints) {
                          final crossAxisCount =
                              constraints.maxWidth >= 900 ? 3 : 2;
                          return GridView.count(
                            physics: const NeverScrollableScrollPhysics(),
                            shrinkWrap: true,
                            crossAxisCount: crossAxisCount,
                            childAspectRatio: 2.6,
                            crossAxisSpacing: 16,
                            mainAxisSpacing: 16,
                            children: [
                              for (final doc in shown)
                                _popularCourseCard(doc.id, doc.data()),
                            ],
                          );
                        },
                      );
                    },
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
