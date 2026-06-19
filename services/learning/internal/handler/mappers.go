package handler

import (
	"encoding/json"

	"github.com/even-app/even-app/services/learning/internal/domain"
	http_v1 "github.com/even-app/even-app/services/learning/internal/gen/http/v1"
	"github.com/even-app/even-app/services/learning/internal/repository"
	"github.com/even-app/even-app/services/learning/internal/service"
	"github.com/google/uuid"
)

func mapCourse(c domain.CourseView) http_v1.Course {
	return http_v1.Course{
		ID: c.ID, Title: c.Title,
		TargetLanguageID: c.TargetLanguageID, UILanguageID: c.UILanguageID,
		OwnerID: c.OwnerID, IsPublished: c.IsPublished,
	}
}
func mapCourseListItem(c service.CourseListItem) http_v1.CourseListItem {
	out := http_v1.CourseListItem{
		ID: c.ID, Title: c.Title,
		TargetLanguage: http_v1.Language{
			ID: c.TargetLanguageID, Code: c.TargetLangCode, Name: c.TargetLangName,
			NativeName: c.TargetLangName, Direction: http_v1.LanguageDirectionLtr, IsActive: true,
		},
	}
	if c.ProgressPercent != nil {
		out.ProgressPercent = http_v1.NewOptFloat64(*c.ProgressPercent)
	}
	return out
}

func mapCourseOutline(o service.CourseOutline) http_v1.CourseOutlineResponse {
	lessons := make([]http_v1.CourseOutlineLesson, 0, len(o.Lessons))
	for _, l := range o.Lessons {
		sections := make([]http_v1.CourseOutlineSection, 0, len(l.Sections))
		for _, s := range l.Sections {
			blocks := make([]http_v1.CourseOutlineBlock, 0, len(s.Blocks))
			for _, b := range s.Blocks {
				blocks = append(blocks, mapOutlineBlock(b))
			}
			sections = append(sections, http_v1.CourseOutlineSection{
				ID: s.ID, Title: s.Title, SortOrder: s.SortOrder,
				ProgressPercent: s.ProgressPercent, Blocks: blocks,
			})
		}
		lessons = append(lessons, http_v1.CourseOutlineLesson{
			ID: l.ID, Title: l.Title, SortOrder: l.SortOrder,
			ProgressPercent: l.ProgressPercent, Sections: sections,
		})
	}
	out := http_v1.CourseOutlineResponse{
		CourseID: o.CourseID, Title: o.Title, ProgressPercent: o.ProgressPercent,
		Lessons: lessons,
	}
	if o.CurrentLessonID != nil {
		out.CurrentLessonID = http_v1.NewOptUUID(*o.CurrentLessonID)
	}
	if o.CurrentBlockID != nil {
		out.CurrentBlockID = http_v1.NewOptUUID(*o.CurrentBlockID)
	}
	return out
}

func mapOutlineBlock(b service.OutlineBlock) http_v1.CourseOutlineBlock {
	out := http_v1.CourseOutlineBlock{
		ID: b.ID, LessonID: b.LessonID, Title: b.Title, SortOrder: b.SortOrder,
		BlockType: b.BlockType, IsGradable: b.IsGradable,
		Status:          http_v1.CourseOutlineBlockStatus(b.Status),
		ProgressPercent: b.ProgressPercent, IsCurrent: b.IsCurrent,
	}
	if b.DisplayLabel != nil {
		out.DisplayLabel = http_v1.NewOptString(*b.DisplayLabel)
	}
	return out
}

func mapLesson(s domain.LessonSnapshot, lexemes map[uuid.UUID]repository.LexemeView, media map[uuid.UUID]repository.MediaView) http_v1.Lesson {
	sections := make([]http_v1.LessonSection, 0, len(s.Sections))
	blocksBySection := map[string][]http_v1.LessonBlock{}
	orphanBlocks := make([]http_v1.LessonBlock, 0)
	for _, b := range s.Blocks {
		lb := mapBlock(b)
		if b.SectionID != nil {
			blocksBySection[b.SectionID.String()] = append(blocksBySection[b.SectionID.String()], lb)
		} else {
			orphanBlocks = append(orphanBlocks, lb)
		}
	}
	for _, sec := range s.Sections {
		sections = append(sections, http_v1.LessonSection{
			ID: sec.ID, Title: sec.Title, SortOrder: sec.SortOrder,
			SectionKind: http_v1.LessonSectionSectionKind(sec.SectionKind),
			Blocks:      blocksBySection[sec.ID.String()],
		})
	}
	if len(orphanBlocks) > 0 && len(sections) == 0 {
		sections = append(sections, http_v1.LessonSection{
			ID: s.ID, Title: "Content", SortOrder: 0,
			SectionKind: http_v1.LessonSectionSectionKindContent,
			Blocks:      orphanBlocks,
		})
	}
	out := http_v1.Lesson{
		ID: s.ID, CourseID: s.CourseID, Title: s.Title,
		SortOrder: s.SortOrder, Version: s.Version,
		Status: http_v1.LessonStatus(s.Status), Sections: sections,
	}
	if resolved := mapResolvedLexemes(lexemes); resolved != nil {
		out.ResolvedLexemes = http_v1.NewOptLessonResolvedLexemes(resolved)
	}
	if resolvedMedia := mapResolvedMedia(media); resolvedMedia != nil {
		out.ResolvedMedia = http_v1.NewOptLessonResolvedMedia(resolvedMedia)
	}
	return out
}

func mapResolvedMedia(items map[uuid.UUID]repository.MediaView) http_v1.LessonResolvedMedia {
	if len(items) == 0 {
		return nil
	}
	out := make(http_v1.LessonResolvedMedia, len(items))
	for id, m := range items {
		item := http_v1.ResolvedMedia{
			ID: m.ID, URL: repository.MediaRefURL(m.Scope, m.ID),
			DisplayName: m.DisplayName, MimeType: m.MimeType,
			MediaKind: http_v1.ResolvedMediaMediaKind(m.MediaKind),
			Scope:     http_v1.ResolvedMediaScope(m.Scope),
		}
		out[id.String()] = item
	}
	return out
}

func mapResolvedLexemes(lexemes map[uuid.UUID]repository.LexemeView) http_v1.LessonResolvedLexemes {
	if len(lexemes) == 0 {
		return nil
	}
	out := make(http_v1.LessonResolvedLexemes, len(lexemes))
	for id, lx := range lexemes {
		item := http_v1.ResolvedLexeme{ID: lx.ID, Lemma: lx.Lemma}
		if lx.PartOfSpeech != nil {
			item.PartOfSpeech = http_v1.NewOptString(*lx.PartOfSpeech)
		}
		out[id.String()] = item
	}
	return out
}

func mapBlock(b domain.BlockSnap) http_v1.LessonBlock {
	cfg := mapBlockConfig(b.Config)
	out := http_v1.LessonBlock{
		ID: b.ID, SortOrder: b.SortOrder, BlockType: b.BlockType,
		Config: cfg, IsHomework: b.IsHomework, IsGradable: b.IsGradable,
	}
	if b.SectionID != nil {
		out.SectionID = http_v1.NewOptUUID(*b.SectionID)
	}
	if b.DisplayLabel != nil {
		out.DisplayLabel = http_v1.NewOptString(*b.DisplayLabel)
	}
	if b.Title != nil {
		out.Title = http_v1.NewOptString(*b.Title)
	}
	return out
}

func mapBlockConfig(raw json.RawMessage) http_v1.LessonBlockConfig {
	out := http_v1.LessonBlockConfig{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func mapBlockProgress(p service.BlockProgress) http_v1.UserBlockProgress {
	return http_v1.UserBlockProgress{
		LessonBlockID: p.LessonBlockID,
		Status:        http_v1.UserBlockProgressStatus(p.Status),
		Score:         p.Score,
		Attempts:      p.Attempts,
	}
}

func mapReviewItem(r service.ReviewItemView) http_v1.ReviewItem {
	out := http_v1.ReviewItem{
		LessonBlockID: r.LessonBlockID, SubItemIndex: r.SubItemIndex,
		BlockType: r.BlockType, SourceLessonID: r.SourceLessonID,
		SourceLessonTitle: r.SourceLessonTitle, FailureCount: r.FailureCount,
		DueAt: r.DueAt, Status: http_v1.ReviewItemStatus(r.Status),
		Block: mapBlock(r.Block),
	}
	if r.Title != "" {
		out.Title = http_v1.NewOptString(r.Title)
	}
	return out
}

func mapFlowItem(item service.FlowItem) http_v1.LessonFlowItem {
	if item.Kind == "review_injection" && item.Review != nil {
		return http_v1.NewLessonFlowReviewItemLessonFlowItem(http_v1.LessonFlowReviewItem{
			Kind:   http_v1.LessonFlowReviewItemKindReviewInjection,
			Review: mapReviewItem(*item.Review),
		})
	}
	return http_v1.NewLessonFlowBlockItemLessonFlowItem(http_v1.LessonFlowBlockItem{
		Kind:  http_v1.LessonFlowBlockItemKindLessonBlock,
		Block: mapBlock(item.Block),
	})
}

func mapAttemptResponse(r *service.AttemptResult) http_v1.BlockAttemptResponse {
	out := http_v1.BlockAttemptResponse{
		IsCorrect: r.IsCorrect, Score: r.Score,
		BlockProgress: mapBlockProgress(r.BlockProgress),
	}
	if r.CorrectAnswer != nil {
		if raw, err := json.Marshal(r.CorrectAnswer); err == nil {
			var cfg http_v1.BlockAttemptResponseCorrectAnswer
			if json.Unmarshal(raw, &cfg) == nil {
				out.CorrectAnswer = http_v1.NewOptBlockAttemptResponseCorrectAnswer(cfg)
			}
		}
	}
	return out
}

func mapVocabularyEntry(v service.VocabularyEntry) http_v1.VocabularyEntry {
	return http_v1.VocabularyEntry{
		Lexeme: http_v1.Lexeme{
			ID: v.LexemeID, LanguageID: v.LanguageID, Lemma: v.Lemma,
		},
		FirstSeenAt: v.FirstSeen,
		Mastery:     v.Mastery,
	}
}
