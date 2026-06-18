package service

import (
	"testing"

	"github.com/even-app/even-app/services/learning/internal/domain"
	"github.com/even-app/even-app/services/learning/internal/gen/query"
	"github.com/google/uuid"
)

func TestOutlineBlockFromSnap_gradableNoProgress(t *testing.T) {
	lessonID := uuid.New()
	blockID := uuid.New()
	label := "1.1"
	b := domain.BlockSnap{
		ID: blockID, SortOrder: 0, BlockType: "prompt_choose_word",
		DisplayLabel: &label, IsGradable: true,
	}
	got := outlineBlockFromSnap(lessonID, b, query.UserBlockProgress{})
	if got.Status != "not_started" || got.ProgressPercent != 0 {
		t.Fatalf("expected not_started, got %+v", got)
	}
	if got.LessonID != lessonID || got.ID != blockID {
		t.Fatalf("unexpected ids: %+v", got)
	}
}

func TestOutlineBlockFromSnap_gradableCompleted(t *testing.T) {
	lessonID := uuid.New()
	blockID := uuid.New()
	b := domain.BlockSnap{ID: blockID, BlockType: "prompt_choose_word", Title: strPtr("Exercise"), IsGradable: true}
	prog := query.UserBlockProgress{LessonBlockID: blockID, Status: "completed", Score: 1}
	got := outlineBlockFromSnap(lessonID, b, prog)
	if got.Status != "completed" || got.ProgressPercent != 100 {
		t.Fatalf("expected completed 100%%, got %+v", got)
	}
}

func TestOutlineBlockFromSnap_nonGradable(t *testing.T) {
	got := outlineBlockFromSnap(uuid.New(), domain.BlockSnap{BlockType: "text", IsGradable: false}, query.UserBlockProgress{})
	if got.Status != "completed" || got.ProgressPercent != 100 {
		t.Fatalf("non-gradable should be auto-complete, got %+v", got)
	}
}

func TestBlockProgressPercent(t *testing.T) {
	tests := []struct {
		status string
		score  float64
		want   float64
	}{
		{"not_started", 0, 0},
		{"in_progress", 0, 50},
		{"in_progress", 0.6, 60},
		{"completed", 0, 100},
	}
	for _, tc := range tests {
		if got := blockProgressPercent(tc.status, tc.score); got != tc.want {
			t.Errorf("blockProgressPercent(%q, %v) = %v, want %v", tc.status, tc.score, got, tc.want)
		}
	}
}

func TestSectionProgressPercent(t *testing.T) {
	blocks := []OutlineBlock{
		{IsGradable: true, ProgressPercent: 100},
		{IsGradable: true, ProgressPercent: 50},
		{IsGradable: false, ProgressPercent: 0},
	}
	got := sectionProgressPercent(blocks)
	want := 75.0
	if got != want {
		t.Fatalf("sectionProgressPercent = %v, want %v", got, want)
	}
}

func TestSectionProgressPercent_noGradable(t *testing.T) {
	got := sectionProgressPercent([]OutlineBlock{{IsGradable: false}})
	if got != 100 {
		t.Fatalf("empty gradable section should be 100%%, got %v", got)
	}
}

func TestPickCurrentOutlineBlock(t *testing.T) {
	done := uuid.New()
	current := uuid.New()
	lessonID := uuid.New()
	lessons := []OutlineLesson{{
		ID: lessonID,
		Sections: []OutlineSection{{
			Blocks: []OutlineBlock{
				{ID: done, IsGradable: true, Status: "completed"},
				{ID: current, IsGradable: true, Status: "in_progress"},
			},
		}},
	}}
	lid, bid := pickCurrentOutlineBlock(lessons)
	if lid == nil || bid == nil || *bid != current || *lid != lessonID {
		t.Fatalf("pickCurrentOutlineBlock = %v %v, want lesson=%s block=%s", lid, bid, lessonID, current)
	}
}

func TestPickCurrentOutlineBlock_skipsNonGradable(t *testing.T) {
	target := uuid.New()
	lessonID := uuid.New()
	lessons := []OutlineLesson{{
		ID: lessonID,
		Sections: []OutlineSection{{
			Blocks: []OutlineBlock{
				{ID: uuid.New(), IsGradable: false, Status: "completed"},
				{ID: target, IsGradable: true, Status: "not_started"},
			},
		}},
	}}
	_, bid := pickCurrentOutlineBlock(lessons)
	if bid == nil || *bid != target {
		t.Fatalf("expected first gradable not_started, got %v", bid)
	}
}

func strPtr(s string) *string { return &s }
