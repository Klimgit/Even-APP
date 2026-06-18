import 'package:flutter/material.dart';
import 'package:online_cource_app/theme/app_theme.dart';
import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/exercise_common.dart';

class WordPair {
  /// Word in the known/native language (left column).
  final String known;

  /// Word in the language being learned (right column).
  final String target;

  const WordPair({required this.known, required this.target});
}

/// Exercise type 3: two columns of words (known on the left, target on the
/// right). Tap one in each column: a matching meaning locks the pair as solved;
/// a mismatch counts as an error and the words stay selectable. Done when all
/// pairs are matched.
class WordMatchExercise extends ExerciseData {
  final String prompt;
  final List<WordPair> pairs;

  const WordMatchExercise({
    required this.pairs,
    this.prompt = 'Match the pairs',
  });

  @override
  Widget build({required ValueChanged<bool> onResult}) =>
      _WordMatchWidget(data: this, onResult: onResult);
}

class _WordMatchWidget extends StatefulWidget {
  final WordMatchExercise data;
  final ValueChanged<bool> onResult;

  const _WordMatchWidget({required this.data, required this.onResult});

  @override
  State<_WordMatchWidget> createState() => _WordMatchWidgetState();
}

class _WordMatchWidgetState extends State<_WordMatchWidget> {
  late List<int> _leftOrder; // pair indices, shuffled
  late List<int> _rightOrder; // pair indices, shuffled
  final Set<int> _matched = {};
  int? _selectedLeft;
  int? _selectedRight;
  int? _wrongLeft;
  int? _wrongRight;
  int _mistakes = 0;
  bool _checked = false;

  @override
  void initState() {
    super.initState();
    final indices = List.generate(widget.data.pairs.length, (i) => i);
    _leftOrder = [...indices]..shuffle();
    _rightOrder = [...indices]..shuffle();
  }

  bool get _allMatched => _matched.length == widget.data.pairs.length;

  void _tapLeft(int pairIndex) {
    if (_matched.contains(pairIndex)) return;
    setState(() => _selectedLeft = pairIndex);
    _evaluate();
  }

  void _tapRight(int pairIndex) {
    if (_matched.contains(pairIndex)) return;
    setState(() => _selectedRight = pairIndex);
    _evaluate();
  }

  void _evaluate() {
    if (_selectedLeft == null || _selectedRight == null) return;
    final left = _selectedLeft!;
    final right = _selectedRight!;
    if (left == right) {
      setState(() {
        _matched.add(left);
        _selectedLeft = null;
        _selectedRight = null;
      });
    } else {
      _mistakes++;
      setState(() {
        _wrongLeft = left;
        _wrongRight = right;
        _selectedLeft = null;
        _selectedRight = null;
      });
      Future.delayed(const Duration(milliseconds: 600), () {
        if (mounted) {
          setState(() {
            _wrongLeft = null;
            _wrongRight = null;
          });
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          widget.data.prompt,
          style: Theme.of(context)
              .textTheme
              .titleLarge
              ?.copyWith(fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 24),
        Expanded(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(child: _column(_leftOrder, isLeft: true)),
              const SizedBox(width: 16),
              Expanded(child: _column(_rightOrder, isLeft: false)),
            ],
          ),
        ),
        ExerciseBottomBar(
          status: _checked
              ? (_mistakes == 0
                  ? ExerciseStatus.correct
                  : ExerciseStatus.wrong)
              : ExerciseStatus.answering,
          canCheck: _allMatched && !_checked,
          onCheck: () => setState(() => _checked = true),
          onContinue: () => widget.onResult(_mistakes == 0),
        ),
      ],
    );
  }

  Widget _column(List<int> order, {required bool isLeft}) {
    return Column(
      children: [
        for (final pairIndex in order) ...[
          _MatchButton(
            label: isLeft
                ? widget.data.pairs[pairIndex].known
                : widget.data.pairs[pairIndex].target,
            matched: _matched.contains(pairIndex),
            selected: isLeft
                ? _selectedLeft == pairIndex
                : _selectedRight == pairIndex,
            wrong: isLeft
                ? _wrongLeft == pairIndex
                : _wrongRight == pairIndex,
            onTap: () => isLeft ? _tapLeft(pairIndex) : _tapRight(pairIndex),
          ),
          const SizedBox(height: 12),
        ],
      ],
    );
  }
}

class _MatchButton extends StatelessWidget {
  final String label;
  final bool matched;
  final bool selected;
  final bool wrong;
  final VoidCallback onTap;

  const _MatchButton({
    required this.label,
    required this.matched,
    required this.selected,
    required this.wrong,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    Color borderColor = AppTheme.dividerColor;
    Color bgColor = Colors.white;
    Color textColor = AppTheme.textColor;

    if (matched) {
      borderColor = AppTheme.accentColor;
      bgColor = AppTheme.accentColor.withOpacity(0.12);
      textColor = AppTheme.accentColor;
    } else if (wrong) {
      borderColor = AppTheme.errorColor;
      bgColor = AppTheme.errorColor.withOpacity(0.1);
    } else if (selected) {
      borderColor = AppTheme.accentColor;
      bgColor = AppTheme.accentColor.withOpacity(0.08);
    }

    return Opacity(
      opacity: matched ? 0.6 : 1,
      child: InkWell(
        onTap: matched ? null : onTap,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          width: double.infinity,
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 16),
          decoration: BoxDecoration(
            color: bgColor,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: borderColor,
              width: selected || wrong || matched ? 1.8 : 1,
            ),
          ),
          child: Text(
            label,
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: textColor,
            ),
          ),
        ),
      ),
    );
  }
}
