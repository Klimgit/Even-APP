import 'package:online_cource_app/exercises/exercise.dart';
import 'package:online_cource_app/exercises/listen_choice_exercise.dart';
import 'package:online_cource_app/exercises/sentence_builder_exercise.dart';
import 'package:online_cource_app/exercises/word_match_exercise.dart';

/// Sample lesson used to preview the exercise types end-to-end.
List<ExerciseData> demoExercises() => const [
      // Type 1a: listen and choose from a 2x2 grid.
      ListenChoiceExercise(
        word: 'apple',
        prompt: 'Which word did you hear?',
        options: ['apple', 'orange', 'banana', 'grape'],
        correctAnswer: 'apple',
      ),
      // Type 1b: listen and type the answer.
      ListenChoiceExercise(
        word: 'hello',
        prompt: 'Type what you hear',
        correctAnswer: 'hello',
      ),
      // Type 2: rebuild the sentence from shuffled words.
      SentenceBuilderExercise(
        promptText: 'Я студент',
        speakText: 'I am a student',
        correctWords: ['I', 'am', 'a', 'student'],
      ),
      // Type 3: match known words to their translations.
      WordMatchExercise(
        pairs: [
          WordPair(known: 'кошка', target: 'cat'),
          WordPair(known: 'собака', target: 'dog'),
          WordPair(known: 'дом', target: 'house'),
          WordPair(known: 'вода', target: 'water'),
          WordPair(known: 'книга', target: 'book'),
        ],
      ),
    ];
