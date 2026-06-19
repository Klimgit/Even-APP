package repository

import (
	"context"
	"strings"

	lexiconclient "github.com/even-app/even-app/libs/clients/lexicon"
	"github.com/google/uuid"
)

type LexiconRemote struct {
	client *lexiconclient.Client
}

func NewLexiconRemote(baseURL, token string) *LexiconRemote {
	return &LexiconRemote{client: lexiconclient.New(baseURL, token)}
}

func (r *LexiconRemote) Available() bool {
	return r != nil && r.client.Available()
}

func (r *LexiconRemote) GetLanguage(ctx context.Context, id uuid.UUID) (LanguageView, error) {
	row, err := r.client.GetLanguage(ctx, id)
	if err != nil {
		return LanguageView{}, err
	}
	return LanguageView{
		ID: row.ID, Code: row.Code, Name: row.Name, NativeName: row.NativeName,
		Direction: row.Direction, IsActive: row.IsActive,
	}, nil
}

func (r *LexiconRemote) LanguagesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]LanguageView, error) {
	rows, err := r.client.LanguagesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]LanguageView, len(rows))
	for id, row := range rows {
		out[id] = LanguageView{
			ID: row.ID, Code: row.Code, Name: row.Name, NativeName: row.NativeName,
			Direction: row.Direction, IsActive: row.IsActive,
		}
	}
	return out, nil
}

func (r *LexiconRemote) LexemesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]LexemeView, error) {
	rows, err := r.client.LexemesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]LexemeView, len(rows))
	for id, row := range rows {
		out[id] = LexemeView{ID: row.ID, Lemma: row.Lemma, PartOfSpeech: row.PartOfSpeech}
	}
	return out, nil
}

func (r *LexiconRemote) FilterLexemeIDsBySearch(ctx context.Context, ids []uuid.UUID, search string) ([]uuid.UUID, error) {
	search = strings.TrimSpace(search)
	if search == "" || len(ids) == 0 {
		return ids, nil
	}
	return r.client.FilterLexemeIDsBySearch(ctx, ids, search)
}

var _ LexiconSource = (*LexiconRemote)(nil)
