import 'package:flutter/material.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/theme/app_theme.dart';

/// On-screen keyboard for typing answers in the study language (e.g. Even Cyrillic).
class StudyLanguageKeyboard extends StatefulWidget {
  final List<AlphabetLetterDto> letters;
  final ValueChanged<String> onCharacter;
  final VoidCallback onBackspace;
  final VoidCallback onSpace;
  final String? title;

  const StudyLanguageKeyboard({
    super.key,
    required this.letters,
    required this.onCharacter,
    required this.onBackspace,
    required this.onSpace,
    this.title,
  });

  @override
  State<StudyLanguageKeyboard> createState() => _StudyLanguageKeyboardState();
}

class _StudyLanguageKeyboardState extends State<StudyLanguageKeyboard> {
  bool _shift = false;

  static const _keysPerRow = 9;

  @override
  Widget build(BuildContext context) {
    if (widget.letters.isEmpty) return const SizedBox.shrink();

    final rows = <List<AlphabetLetterDto>>[];
    for (var i = 0; i < widget.letters.length; i += _keysPerRow) {
      rows.add(widget.letters.sublist(
        i,
        i + _keysPerRow > widget.letters.length ? widget.letters.length : i + _keysPerRow,
      ));
    }

    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFE8ECF4),
        borderRadius: BorderRadius.circular(AppTheme.borderRadius),
        border: Border.all(color: AppTheme.dividerColor),
      ),
      padding: const EdgeInsets.all(8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            widget.title ?? 'Клавиатура языка',
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppTheme.secondaryTextColor,
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 8),
          for (final row in rows) ...[
            Row(
              children: [
                for (final letter in row)
                  Expanded(
                    child: Padding(
                      padding: const EdgeInsets.all(2),
                      child: _LetterKey(
                        letter: letter,
                        shift: _shift,
                        onTap: () => widget.onCharacter(letter.displayChar(shift: _shift)),
                      ),
                    ),
                  ),
                for (var i = row.length; i < _keysPerRow; i++)
                  const Expanded(child: SizedBox()),
              ],
            ),
          ],
          const SizedBox(height: 4),
          Row(
            children: [
              _ActionKey(
                label: _shift ? '⇧' : '⇧',
                active: _shift,
                flex: 2,
                onTap: () => setState(() => _shift = !_shift),
              ),
              _ActionKey(
                label: 'пробел',
                flex: 5,
                onTap: widget.onSpace,
              ),
              _ActionKey(
                icon: Icons.backspace_outlined,
                flex: 2,
                onTap: widget.onBackspace,
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _LetterKey extends StatelessWidget {
  final AlphabetLetterDto letter;
  final bool shift;
  final VoidCallback onTap;

  const _LetterKey({
    required this.letter,
    required this.shift,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final char = letter.displayChar(shift: shift);
    return Material(
      color: Colors.white,
      elevation: 1,
      shadowColor: Colors.black26,
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: AspectRatio(
          aspectRatio: 1.1,
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                char,
                style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
              ),
              if (letter.transcription != null && letter.transcription!.isNotEmpty)
                Text(
                  letter.transcription!,
                  style: TextStyle(
                    fontSize: 9,
                    color: AppTheme.secondaryTextColor.withValues(alpha: 0.8),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ActionKey extends StatelessWidget {
  final String? label;
  final IconData? icon;
  final VoidCallback onTap;
  final int flex;
  final bool active;

  const _ActionKey({
    this.label,
    this.icon,
    required this.onTap,
    this.flex = 1,
    this.active = false,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      flex: flex,
      child: Padding(
        padding: const EdgeInsets.all(2),
        child: Material(
          color: active ? AppTheme.primaryColor.withValues(alpha: 0.15) : Colors.white,
          elevation: 1,
          borderRadius: BorderRadius.circular(8),
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(8),
            child: SizedBox(
              height: 44,
              child: Center(
                child: icon != null
                    ? Icon(icon, size: 20, color: AppTheme.textColor)
                    : Text(
                        label ?? '',
                        style: TextStyle(
                          fontSize: label == '⇧' ? 18 : 13,
                          fontWeight: FontWeight.w600,
                          color: AppTheme.textColor,
                        ),
                      ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
