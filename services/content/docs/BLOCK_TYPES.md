# Block types — MVP config contract

Even content service MVP exposes **17 block types** (6 content + 11 gradable) via `GET /api/v1/teacher/block-types`. Each lesson block stores `block_type` and a JSON `config` object. The content service validates config on create/patch (`ValidateBlockConfig`: required keys, array shapes, index bounds).

## Categories

| Category ID | Title | Types |
|-------------|-------|-------|
| `content` | Контент | `text`, `images_stacked`, `audio`, `video`, `vocabulary_set`, `note` |
| `basic_exercises` | Базовые задания | 7 gradable prompt/gap types |
| `listening` | Аудирование | 4 gradable listen types |

## Content blocks (not gradable)

### `text`

```json
{
  "items": [
    { "kind": "text", "text": "Привет!", "language_id": "uuid-optional" }
  ]
}
```

Uses shared **PromptConfig** shape (`items` array).

### `images_stacked`

```json
{
  "items": [
    { "kind": "image", "media_asset_id": "uuid" },
    { "kind": "text", "text": "Подпись" }
  ]
}
```

### `audio`

```json
{
  "items": [
    { "kind": "audio", "media_asset_id": "uuid", "lexeme_id": "uuid-optional" }
  ]
}
```

### `video`

```json
{
  "media_asset_id": "uuid",
  "caption": "optional text"
}
```

### `vocabulary_set`

Introduces lexemes (drives coverage `introduced`).

```json
{
  "lexeme_ids": ["uuid", "uuid"],
  "show_images": true,
  "show_audio": true
}
```

### `note`

```json
{
  "body": "Teacher note or tip"
}
```

## Gradable blocks

All gradable types set `is_gradable: true` in API responses. Shared **PromptConfig** for stimulus:

```json
{
  "items": [
    { "kind": "lexeme", "lexeme_id": "uuid" },
    { "kind": "image", "media_asset_id": "uuid" },
    { "kind": "audio", "media_asset_id": "uuid" },
    { "kind": "text", "text": "..." }
  ]
}
```

### Basic exercises

| `block_type` | Config highlights |
|--------------|-------------------|
| `prompt_choose_word` | `{ "prompt": PromptConfig, "choices": [{ "lexeme_id": "uuid" }] , "correct_index": 0 }` |
| `prompt_type_word` | `{ "prompt": PromptConfig, "correct_lexeme_id": "uuid", "correct_form_id": "uuid-optional" }` |
| `word_choose_image` | `{ "prompt": PromptConfig, "choices": [{ "media_asset_id": "uuid" }], "correct_index": 0 }` |
| `word_choose_translation` | `{ "prompt": PromptConfig, "choices": [{ "text": "..." }], "correct_index": 0 }` |
| `gap_sentence_choose_word` | `{ "body": "… {{gap:0}} …", "gaps": [{ "id": 0, "correct_lexeme_id": "uuid" }], "choices": ["uuid"] }` |
| `prompt_sentence_word_order` | `{ "sub_items": [{ "prompt": PromptConfig, "tokens": ["uuid-or-text"], "correct_order": [0,1,2] }] }` |
| `prompt_sentence_type` | `{ "prompt": PromptConfig, "correct_text": "…" }` |

### Listening

| `block_type` | Config highlights |
|--------------|-------------------|
| `listen_choose_word` | Same as `prompt_choose_word` with audio-only prompt |
| `listen_type_word` | Same as `prompt_type_word` |
| `listen_sentence_word_order` | Same as `prompt_sentence_word_order` |
| `listen_sentence_type` | Same as `prompt_sentence_type` |

## Lexeme references for coverage

The content service extracts lexeme usage from config for coverage endpoints:

- `lexeme_id`, `lexeme_ids[]`, `correct_lexeme_id`, `word_bank_lexeme_ids[]`
- `gaps[].correct_lexeme_id`, `gaps[].correct_form_id`
- Nested `items` / `sub_items` / `prompt.items`

Usage kinds:

- **introduced** — content blocks (`vocabulary_set`, media/text introducing words)
- **exercised** — gradable blocks
- **referenced** — word bank / distractor lexeme IDs

## Fields on every block (API, not in config)

| Field | Description |
|-------|-------------|
| `section_id` | Optional section container |
| `sort_order` | Order within lesson |
| `display_label` | e.g. `1.3` |
| `title` | Editor title |
| `is_homework` | Homework flag |

## Phase 2 (not in MVP catalog)

Grammar (`grammar_table`, `grammar_exercise`), reading blocks, gap drag/type variants, tests, favorites API — see full BlockType enum in [DTO.md](../../DTO.md).
