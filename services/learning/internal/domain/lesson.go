package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type LessonSnapshot struct {
	ID        uuid.UUID     `json:"id"`
	CourseID  uuid.UUID     `json:"course_id"`
	Title     string        `json:"title"`
	SortOrder int           `json:"sort_order"`
	Version   int           `json:"version"`
	Status    string        `json:"status"`
	Sections  []SectionSnap `json:"sections"`
	Blocks    []BlockSnap   `json:"blocks"`
}

type SectionSnap struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	SortOrder   int       `json:"sort_order"`
	SectionKind string    `json:"section_kind"`
}

type BlockSnap struct {
	ID           uuid.UUID       `json:"id"`
	SectionID    *uuid.UUID      `json:"section_id,omitempty"`
	SortOrder    int             `json:"sort_order"`
	DisplayLabel *string         `json:"display_label,omitempty"`
	Title        *string         `json:"title,omitempty"`
	BlockType    string          `json:"block_type"`
	Config       json.RawMessage `json:"config"`
	IsHomework   bool            `json:"is_homework"`
	IsGradable   bool            `json:"is_gradable"`
	LexemeRefs   []LexemeRefSnap `json:"lexeme_refs,omitempty"`
}

type LexemeRefSnap struct {
	LexemeID uuid.UUID  `json:"lexeme_id"`
	FormID   *uuid.UUID `json:"form_id,omitempty"`
	Role     string     `json:"role"`
}

type CourseView struct {
	ID               uuid.UUID
	Title            string
	TargetLanguageID uuid.UUID
	TargetLangCode   string
	TargetLangName   string
	UILanguageID     uuid.UUID
	OwnerID          uuid.UUID
	IsPublished      bool
	InviteCode       string
}

type LessonSummary struct {
	ID               uuid.UUID
	Title            string
	SortOrder        int
	CompletedPercent float64
}

func (s LessonSnapshot) MarshalJSONBlob() ([]byte, error) {
	return json.Marshal(s)
}

func ParseLessonSnapshot(raw []byte) (LessonSnapshot, error) {
	var s LessonSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		return LessonSnapshot{}, err
	}
	return s, nil
}

func GradableBlocksInOrder(s LessonSnapshot) []BlockSnap {
	out := make([]BlockSnap, 0)
	for _, b := range s.Blocks {
		if b.IsGradable {
			out = append(out, b)
		}
	}
	return out
}

func ReviewDueAt(failureCount int, from time.Time) time.Time {
	return from.Add(time.Duration(ReviewDueInterval(failureCount)) * time.Hour)
}
