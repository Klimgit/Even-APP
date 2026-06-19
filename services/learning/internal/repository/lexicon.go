package repository

import (
	"context"
	"strings"

	"github.com/even-app/even-app/services/learning/internal/gen/lexiconquery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LexiconReader struct {
	queries *lexiconquery.Queries
}

func NewLexiconReader(pool *pgxpool.Pool) *LexiconReader {
	if pool == nil {
		return nil
	}
	return &LexiconReader{queries: lexiconquery.New(pool)}
}

func (r *LexiconReader) Available() bool {
	return r != nil && r.queries != nil
}

type LanguageView struct {
	ID         uuid.UUID
	Code       string
	Name       string
	NativeName string
	Direction  string
	IsActive   bool
}

func (r *LexiconReader) GetLanguage(ctx context.Context, id uuid.UUID) (LanguageView, error) {
	row, err := r.queries.GetLanguageByID(ctx, lexiconquery.GetLanguageByIDParams{ID: id})
	if err != nil {
		return LanguageView{}, err
	}
	return LanguageView{
		ID: row.ID, Code: row.Code, Name: row.Name, NativeName: row.NativeName,
		Direction: row.Direction, IsActive: row.IsActive,
	}, nil
}

func (r *LexiconReader) LanguagesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]LanguageView, error) {
	out := make(map[uuid.UUID]LanguageView)
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.queries.ListLanguagesByIDs(ctx, lexiconquery.ListLanguagesByIDsParams{Column1: ids})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = LanguageView{
			ID: row.ID, Code: row.Code, Name: row.Name, NativeName: row.NativeName,
			Direction: row.Direction, IsActive: row.IsActive,
		}
	}
	return out, nil
}

type LexemeView struct {
	ID           uuid.UUID
	Lemma        string
	PartOfSpeech *string
}

func (r *LexiconReader) LexemesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]LexemeView, error) {
	out := make(map[uuid.UUID]LexemeView)
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.queries.ListLexemesByIDs(ctx, lexiconquery.ListLexemesByIDsParams{Column1: ids})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = LexemeView{ID: row.ID, Lemma: row.Lemma, PartOfSpeech: row.PartOfSpeech}
	}
	return out, nil
}

func (r *LexiconReader) FilterLexemeIDsBySearch(ctx context.Context, ids []uuid.UUID, search string) ([]uuid.UUID, error) {
	search = strings.TrimSpace(search)
	if search == "" || len(ids) == 0 {
		return ids, nil
	}
	return r.queries.FilterLexemeIDsBySearch(ctx, lexiconquery.FilterLexemeIDsBySearchParams{
		Column1: ids, Column2: &search,
	})
}
