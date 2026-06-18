package handler

import (
	"testing"

	"github.com/even-app/even-app/services/learning/internal/service"
	"github.com/google/uuid"
)

func TestMapCourseOutline(t *testing.T) {
	courseID := uuid.New()
	lessonID := uuid.New()
	sectionID := uuid.New()
	blockID := uuid.New()
	currentLesson := lessonID
	currentBlock := blockID
	label := "1.1"

	out := mapCourseOutline(service.CourseOutline{
		CourseID: courseID, Title: "Course", ProgressPercent: 33.3,
		CurrentLessonID: &currentLesson, CurrentBlockID: &currentBlock,
		Lessons: []service.OutlineLesson{{
			ID: lessonID, Title: "L1", SortOrder: 1, ProgressPercent: 50,
			Sections: []service.OutlineSection{{
				ID: sectionID, Title: "Section", SortOrder: 0, ProgressPercent: 50,
				Blocks: []service.OutlineBlock{{
					ID: blockID, LessonID: lessonID, DisplayLabel: &label, Title: "Task",
					SortOrder: 0, BlockType: "prompt_choose_word", IsGradable: true,
					Status: "in_progress", ProgressPercent: 50, IsCurrent: true,
				}},
			}},
		}},
	})

	if out.CourseID != courseID || out.Title != "Course" {
		t.Fatalf("unexpected course fields: %+v", out)
	}
	if !out.CurrentLessonID.IsSet() || out.CurrentLessonID.Value != currentLesson {
		t.Fatalf("current lesson not mapped: %+v", out.CurrentLessonID)
	}
	if len(out.Lessons) != 1 || len(out.Lessons[0].Sections) != 1 || len(out.Lessons[0].Sections[0].Blocks) != 1 {
		t.Fatalf("tree not mapped: %+v", out.Lessons)
	}
	block := out.Lessons[0].Sections[0].Blocks[0]
	if !block.DisplayLabel.IsSet() || block.DisplayLabel.Value != label {
		t.Fatalf("display label not mapped: %+v", block.DisplayLabel)
	}
	if block.Status != "in_progress" || !block.IsCurrent {
		t.Fatalf("block status/current wrong: %+v", block)
	}
}

func TestMapBlockProgress(t *testing.T) {
	blockID := uuid.New()
	got := mapBlockProgress(service.BlockProgress{
		LessonBlockID: blockID, Status: "completed", Score: 1, Attempts: 2,
	})
	if got.LessonBlockID != blockID || got.Status != "completed" || got.Attempts != 2 {
		t.Fatalf("unexpected progress map: %+v", got)
	}
}

func TestMapAttemptResponse_withCorrectAnswer(t *testing.T) {
	blockID := uuid.New()
	got := mapAttemptResponse(&service.AttemptResult{
		IsCorrect: false, Score: 0,
		BlockProgress: service.BlockProgress{LessonBlockID: blockID, Status: "in_progress", Score: 0, Attempts: 1},
		CorrectAnswer: map[string]any{"selected_index": float64(0)},
	})
	if got.IsCorrect || !got.CorrectAnswer.IsSet() {
		t.Fatalf("expected incorrect with correct_answer, got %+v", got)
	}
}
