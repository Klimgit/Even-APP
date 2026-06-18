package domain

// BlockTypeInfo describes one MVP block type in the editor catalog.
type BlockTypeInfo struct {
	BlockType   string
	Title       string
	Description string
	IsGradable  bool
}

// BlockTypeCategory groups block types for the editor palette.
type BlockTypeCategory struct {
	ID    string
	Title string
	Types []BlockTypeInfo
}

var gradableBlockTypes = map[string]struct{}{
	"prompt_choose_word":         {},
	"prompt_type_word":           {},
	"word_choose_image":          {},
	"word_choose_translation":    {},
	"gap_sentence_choose_word":   {},
	"prompt_sentence_word_order": {},
	"prompt_sentence_type":       {},
	"listen_choose_word":         {},
	"listen_type_word":           {},
	"listen_sentence_word_order": {},
	"listen_sentence_type":       {},
}

// IsGradableBlockType reports whether a block type is scored in MVP.
func IsGradableBlockType(blockType string) bool {
	_, ok := gradableBlockTypes[blockType]
	return ok
}

// MVPBlockTypeCatalog returns the static 6 content + 11 gradable block types.
func MVPBlockTypeCatalog() []BlockTypeCategory {
	return []BlockTypeCategory{
		{
			ID:    "content",
			Title: "Контент",
			Types: []BlockTypeInfo{
				{BlockType: "text", Title: "Текст", Description: "Произвольный текст", IsGradable: false},
				{BlockType: "images_stacked", Title: "Изображения", Description: "Стопка изображений", IsGradable: false},
				{BlockType: "audio", Title: "Аудио", Description: "Аудиозапись", IsGradable: false},
				{BlockType: "video", Title: "Видео", Description: "Видеозапись", IsGradable: false},
				{BlockType: "vocabulary_set", Title: "Набор слов", Description: "Ввод новой лексики", IsGradable: false},
				{BlockType: "note", Title: "Заметка", Description: "Подсказка или памятка", IsGradable: false},
			},
		},
		{
			ID:    "basic_exercises",
			Title: "Базовые задания",
			Types: []BlockTypeInfo{
				{BlockType: "prompt_choose_word", Title: "Выбрать слово", IsGradable: true},
				{BlockType: "prompt_type_word", Title: "Напечатать слово", IsGradable: true},
				{BlockType: "word_choose_image", Title: "Слово → картинка", IsGradable: true},
				{BlockType: "word_choose_translation", Title: "Слово → перевод", IsGradable: true},
				{BlockType: "gap_sentence_choose_word", Title: "Пропуск в предложении", IsGradable: true},
				{BlockType: "prompt_sentence_word_order", Title: "Составить предложение", IsGradable: true},
				{BlockType: "prompt_sentence_type", Title: "Напечатать предложение", IsGradable: true},
			},
		},
		{
			ID:    "listening",
			Title: "Аудирование",
			Types: []BlockTypeInfo{
				{BlockType: "listen_choose_word", Title: "Послушать → выбрать слово", IsGradable: true},
				{BlockType: "listen_type_word", Title: "Послушать → напечатать слово", IsGradable: true},
				{BlockType: "listen_sentence_word_order", Title: "Послушать → составить", IsGradable: true},
				{BlockType: "listen_sentence_type", Title: "Послушать → напечатать", IsGradable: true},
			},
		},
	}
}
