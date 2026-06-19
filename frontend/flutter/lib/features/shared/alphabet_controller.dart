import 'package:get/get.dart';
import 'package:online_cource_app/api/models/lexicon_dto.dart';
import 'package:online_cource_app/api/repositories/lexicon_repository.dart';

/// Cached alphabet letters per language code for the study keyboard.
class AlphabetController extends GetxController {
  final LexiconRepository _lexicon = Get.find<LexiconRepository>();
  final _cache = <String, List<AlphabetLetterDto>>{};

  final RxBool loading = false.obs;
  final RxnString error = RxnString();

  Future<List<AlphabetLetterDto>> lettersFor(String languageCode) async {
    final cached = _cache[languageCode];
    if (cached != null) return cached;

    loading.value = true;
    error.value = null;
    try {
      final list = await _lexicon.getLanguageAlphabet(languageCode);
      _cache[languageCode] = list;
      return list;
    } catch (e) {
      error.value = e.toString();
      return const [];
    } finally {
      loading.value = false;
    }
  }

  void invalidate(String languageCode) => _cache.remove(languageCode);
}
