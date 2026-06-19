package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CourseView struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	TargetLanguageID uuid.UUID `json:"target_language_id"`
	TargetLangCode   string    `json:"target_lang_code,omitempty"`
	TargetLangName   string    `json:"target_lang_name,omitempty"`
	UILanguageID     uuid.UUID `json:"ui_language_id"`
	OwnerID          uuid.UUID `json:"owner_id"`
	IsPublished      bool      `json:"is_published"`
	Visibility       string    `json:"visibility,omitempty"`
	InviteCode       string    `json:"invite_code,omitempty"`
}

type LanguageView struct {
	ID         uuid.UUID `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	NativeName string    `json:"native_name"`
	Direction  string    `json:"direction"`
	IsActive   bool      `json:"is_active"`
}

type LexemeView struct {
	ID           uuid.UUID `json:"id"`
	Lemma        string    `json:"lemma"`
	PartOfSpeech *string   `json:"part_of_speech,omitempty"`
}

type MediaView struct {
	ID          uuid.UUID `json:"id"`
	Scope       string    `json:"scope"`
	DisplayName string    `json:"display_name"`
	MimeType    string    `json:"mime_type"`
	MediaKind   string    `json:"media_kind"`
	URL         string    `json:"url,omitempty"`
}

type UserView struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name,omitempty"`
	Role        string    `json:"role"`
	IsAdmin     bool      `json:"is_admin"`
}

type EnrollmentView struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	CourseID   uuid.UUID  `json:"course_id"`
	EnrolledBy *uuid.UUID `json:"enrolled_by,omitempty"`
	Status     string     `json:"status"`
}

type LessonProgressRow struct {
	LessonID    uuid.UUID `json:"lesson_id"`
	LessonTitle string    `json:"lesson_title"`
	Completed   bool      `json:"completed"`
	BlocksTotal int       `json:"blocks_total"`
	BlocksDone  int       `json:"blocks_done"`
	ScoreAvg    *float64  `json:"score_avg,omitempty"`
}

type PublishedLessonSnapshot struct {
	LessonID uuid.UUID       `json:"lesson_id"`
	CourseID uuid.UUID       `json:"course_id"`
	Version  int32           `json:"version"`
	Snapshot json.RawMessage `json:"snapshot"`
}

type BlockView struct {
	Block       json.RawMessage `json:"block"`
	CourseID    uuid.UUID       `json:"course_id"`
	LessonTitle string          `json:"lesson_title"`
}

type CourseBlockRow struct {
	ID              uuid.UUID       `json:"id"`
	LessonID        uuid.UUID       `json:"lesson_id"`
	SectionID       *uuid.UUID      `json:"section_id,omitempty"`
	SortOrder       int32           `json:"sort_order"`
	DisplayLabel    *string         `json:"display_label,omitempty"`
	Title           *string         `json:"title,omitempty"`
	BlockType       string          `json:"block_type"`
	Config          json.RawMessage `json:"config"`
	IsHomework      bool            `json:"is_homework"`
	LessonTitle     string          `json:"lesson_title"`
	LessonSortOrder int32           `json:"lesson_sort_order"`
}

type IDsRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

type FilterLexemeIDsRequest struct {
	IDs []uuid.UUID `json:"ids"`
	Q   string      `json:"q"`
}

type UpsertSnapshotRequest struct {
	LessonID uuid.UUID       `json:"lesson_id"`
	CourseID uuid.UUID       `json:"course_id"`
	Version  int32           `json:"version"`
	Snapshot json.RawMessage `json:"snapshot"`
}

type CreateEnrollmentRequest struct {
	UserID     uuid.UUID  `json:"user_id"`
	CourseID   uuid.UUID  `json:"course_id"`
	EnrolledBy *uuid.UUID `json:"enrolled_by,omitempty"`
}

type CountResponse struct {
	Count int `json:"count"`
}
