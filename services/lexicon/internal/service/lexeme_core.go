package service

import (
	"time"

	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/google/uuid"
)

const (
	LexemeScopePlatform = "platform"
	LexemeScopeTeacher  = "teacher"
)

type LexemeCore struct {
	ID           uuid.UUID
	LanguageID   uuid.UUID
	Lemma        string
	PartOfSpeech *string
	Notes        *string
	Scope        string
	OwnerID      *uuid.UUID
	CreatedBy    *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func lexemeFromList(row query.ListLexemesByLanguageRow) LexemeCore {
	return LexemeCore{
		ID: row.ID, LanguageID: row.LanguageID, Lemma: row.Lemma,
		PartOfSpeech: row.PartOfSpeech, Notes: row.Notes, Scope: row.Scope,
		OwnerID: row.OwnerID, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func lexemeFromPicker(row query.ListPickerLexemesRow) LexemeCore {
	return LexemeCore{
		ID: row.ID, LanguageID: row.LanguageID, Lemma: row.Lemma,
		PartOfSpeech: row.PartOfSpeech, Notes: row.Notes, Scope: row.Scope,
		OwnerID: row.OwnerID, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func lexemeFromGet(row query.GetLexemeRow) LexemeCore {
	return LexemeCore{
		ID: row.ID, LanguageID: row.LanguageID, Lemma: row.Lemma,
		PartOfSpeech: row.PartOfSpeech, Notes: row.Notes, Scope: row.Scope,
		OwnerID: row.OwnerID, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func lexemeFromCreate(row query.CreateLexemeRow) LexemeCore {
	return LexemeCore{
		ID: row.ID, LanguageID: row.LanguageID, Lemma: row.Lemma,
		PartOfSpeech: row.PartOfSpeech, Notes: row.Notes, Scope: row.Scope,
		OwnerID: row.OwnerID, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func lexemeFromUpdate(row query.UpdateLexemeRow) LexemeCore {
	return LexemeCore{
		ID: row.ID, LanguageID: row.LanguageID, Lemma: row.Lemma,
		PartOfSpeech: row.PartOfSpeech, Notes: row.Notes, Scope: row.Scope,
		OwnerID: row.OwnerID, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
