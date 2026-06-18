package handler

import (
	"encoding/json"

	"github.com/even-app/even-app/services/content/internal/domain"
	http_v1 "github.com/even-app/even-app/services/content/internal/gen/http/v1"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/even-app/even-app/services/content/internal/service"
	"github.com/google/uuid"
)

func mapCourse(v service.CourseView) http_v1.Course {
	out := http_v1.Course{
		ID: v.Course.ID, Title: v.Course.Title,
		TargetLanguageID: v.Course.TargetLanguageID,
		UILanguageID:     v.Course.UiLanguageID,
		OwnerID:          v.Course.OwnerID,
		IsPublished:      v.Course.IsPublished,
	}
	if v.InviteCode != "" {
		out.InviteCode = http_v1.NewOptString(v.InviteCode)
	}
	return out
}

func mapCourses(rows []service.CourseView) []http_v1.Course {
	out := make([]http_v1.Course, len(rows))
	for i, r := range rows {
		out[i] = mapCourse(r)
	}
	return out
}

func mapLessonSummary(row query.Lesson) http_v1.LessonSummary {
	return http_v1.LessonSummary{
		ID: row.ID, CourseID: row.CourseID, Title: row.Title,
		SortOrder: int(row.SortOrder), Version: int(row.Version),
		Status: http_v1.LessonSummaryStatus(row.Status),
	}
}

func mapLessonSummaries(rows []query.Lesson) []http_v1.LessonSummary {
	out := make([]http_v1.LessonSummary, len(rows))
	for i, r := range rows {
		out[i] = mapLessonSummary(r)
	}
	return out
}

func mapLessonFull(full service.LessonFull) http_v1.Lesson {
	out := http_v1.Lesson{
		ID: full.Lesson.ID, CourseID: full.Lesson.CourseID, Title: full.Lesson.Title,
		SortOrder: int(full.Lesson.SortOrder), Version: int(full.Lesson.Version),
		Status:   http_v1.LessonStatus(full.Lesson.Status),
		Sections: make([]http_v1.LessonSection, 0, len(full.Sections)),
	}
	for _, sec := range full.Sections {
		out.Sections = append(out.Sections, mapSectionFull(sec))
	}
	return out
}

func mapSectionFull(sec service.SectionFull) http_v1.LessonSection {
	blocks := make([]http_v1.LessonBlock, 0, len(sec.Blocks))
	for _, b := range sec.Blocks {
		blocks = append(blocks, mapBlock(b))
	}
	return http_v1.LessonSection{
		ID: sec.Section.ID, Title: sec.Section.Title,
		SortOrder:   int(sec.Section.SortOrder),
		SectionKind: http_v1.LessonSectionSectionKind(sec.Section.SectionKind),
		Blocks:      blocks,
	}
}

func mapSection(row query.LessonSection) http_v1.LessonSection {
	return http_v1.LessonSection{
		ID: row.ID, Title: row.Title, SortOrder: int(row.SortOrder),
		SectionKind: http_v1.LessonSectionSectionKind(row.SectionKind),
		Blocks:      []http_v1.LessonBlock{},
	}
}

func mapBlock(row query.LessonBlock) http_v1.LessonBlock {
	out := http_v1.LessonBlock{
		ID: row.ID, SortOrder: int(row.SortOrder),
		BlockType: row.BlockType, Config: bytesToBlockConfig(row.Config),
		IsHomework: row.IsHomework,
		IsGradable: domain.IsGradableBlockType(row.BlockType),
	}
	if row.SectionID != nil {
		out.SectionID = http_v1.NewOptUUID(*row.SectionID)
	}
	if row.DisplayLabel != nil {
		out.DisplayLabel = http_v1.NewOptString(*row.DisplayLabel)
	}
	if row.Title != nil {
		out.Title = http_v1.NewOptString(*row.Title)
	}
	return out
}

func mapBlockTypes(cats []domain.BlockTypeCategory) []http_v1.BlockTypeCategory {
	out := make([]http_v1.BlockTypeCategory, len(cats))
	for i, cat := range cats {
		types := make([]http_v1.BlockTypeInfo, len(cat.Types))
		for j, t := range cat.Types {
			info := http_v1.BlockTypeInfo{
				BlockType: t.BlockType, Title: t.Title,
				IsGradable: t.IsGradable, IsFavorite: false,
			}
			if t.Description != "" {
				info.Description = http_v1.NewOptString(t.Description)
			}
			types[j] = info
		}
		out[i] = http_v1.BlockTypeCategory{ID: cat.ID, Title: cat.Title, Types: types}
	}
	return out
}

func mapCoverageSummary(s service.CoverageSummary) http_v1.CourseLexiconCoverage {
	return http_v1.CourseLexiconCoverage{
		IntroducedCount: s.IntroducedCount,
		ExercisedCount:  s.ExercisedCount,
		LexemeCount:     s.LexemeCount,
	}
}

func mapCoverageByLesson(rows []service.CoverageByLesson) []http_v1.CourseLexiconByLesson {
	out := make([]http_v1.CourseLexiconByLesson, len(rows))
	for i, row := range rows {
		items := make([]http_v1.CourseLexiconByLessonItem, len(row.Items))
		for j, item := range row.Items {
			entry := http_v1.CourseLexiconByLessonItem{
				LexemeID:  mustParseUUID(item.LexemeID),
				Lemma:     item.LexemeID,
				UsageKind: http_v1.CourseLexiconByLessonItemUsageKind(item.UsageKind),
			}
			if item.FormID != "" {
				entry.Form = http_v1.NewOptString(item.FormID)
			}
			if item.BlockDisplayLabel != "" {
				entry.BlockDisplayLabel = http_v1.NewOptString(item.BlockDisplayLabel)
			}
			items[j] = entry
		}
		out[i] = http_v1.CourseLexiconByLesson{
			LessonID: row.LessonID, LessonTitle: row.LessonTitle, Items: items,
		}
	}
	return out
}

func mapFormsCoverage(rows []service.FormCoverage) []http_v1.FormsCoverage {
	out := make([]http_v1.FormsCoverage, len(rows))
	for i, row := range rows {
		forms := make([]http_v1.FormCoverageEntry, len(row.Forms))
		for j, f := range row.Forms {
			forms[j] = http_v1.FormCoverageEntry{
				FormID: mustParseUUID(f.FormID), Form: f.Form,
				Introduced: f.Introduced, Exercised: f.Exercised,
			}
		}
		out[i] = http_v1.FormsCoverage{
			LexemeID: mustParseUUID(row.LexemeID), Lemma: row.Lemma, Forms: forms,
		}
	}
	return out
}

func mapStudentProgress(v service.StudentProgressView) http_v1.StudentProgress {
	lessons := make([]http_v1.StudentProgressLesson, len(v.Lessons))
	for i, l := range v.Lessons {
		lessons[i] = http_v1.StudentProgressLesson{
			LessonID: l.LessonID, Title: l.Title,
			CompletedBlocks: l.CompletedBlocks, TotalBlocks: l.TotalBlocks,
			ScoreAvg: l.ScoreAvg,
		}
	}
	return http_v1.StudentProgress{
		UserID: v.UserID, CourseID: v.CourseID, Lessons: lessons,
	}
}

func bytesToBlockConfig(raw []byte) http_v1.BlockConfig {
	if len(raw) == 0 {
		return http_v1.BlockConfig{}
	}
	var cfg http_v1.BlockConfig
	if err := json.Unmarshal(raw, &cfg); err != nil || cfg == nil {
		return http_v1.BlockConfig{}
	}
	return cfg
}

func blockConfigToBytes(cfg http_v1.BlockConfig) ([]byte, error) {
	if len(cfg) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(cfg)
}

func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func errBody(msg string) http_v1.ErrorResponse {
	return http_v1.ErrorResponse{Error: "not found", Message: msg}
}
