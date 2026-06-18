package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestParseLessonSnapshot(t *testing.T) {
	id := uuid.New()
	raw, err := json.Marshal(LessonSnapshot{
		ID: id, Title: "Test", SortOrder: 1, Version: 2, Status: "published",
		Sections: []SectionSnap{{ID: uuid.New(), Title: "S1", SortOrder: 0, SectionKind: "content"}},
		Blocks: []BlockSnap{
			{ID: uuid.New(), SortOrder: 0, BlockType: "text", Config: json.RawMessage(`{}`), IsGradable: false},
			{ID: uuid.New(), SortOrder: 1, BlockType: "prompt_choose_word", Config: json.RawMessage(`{}`), IsGradable: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseLessonSnapshot(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id || got.Title != "Test" {
		t.Fatalf("unexpected snapshot: %+v", got)
	}
}

func TestParseLessonSnapshot_invalidJSON(t *testing.T) {
	_, err := ParseLessonSnapshot([]byte(`{`))
	if err == nil {
		t.Fatal("expected error on invalid json")
	}
}

func TestGradableBlocksInOrder(t *testing.T) {
	b1 := BlockSnap{ID: uuid.New(), BlockType: "text", IsGradable: false}
	b2 := BlockSnap{ID: uuid.New(), BlockType: "prompt_choose_word", IsGradable: true}
	b3 := BlockSnap{ID: uuid.New(), BlockType: "prompt_type_word", IsGradable: true}
	snap := LessonSnapshot{Blocks: []BlockSnap{b1, b2, b3}}
	got := GradableBlocksInOrder(snap)
	if len(got) != 2 || got[0].ID != b2.ID || got[1].ID != b3.ID {
		t.Fatalf("unexpected gradable blocks: %+v", got)
	}
}

func TestReviewDueAt(t *testing.T) {
	from := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	got := ReviewDueAt(2, from)
	want := from.Add(24 * time.Hour)
	if !got.Equal(want) {
		t.Fatalf("ReviewDueAt(2) = %v, want %v", got, want)
	}
}

func TestLessonSnapshotMarshalRoundTrip(t *testing.T) {
	s := LessonSnapshot{
		ID: uuid.New(), CourseID: uuid.New(), Title: "R", SortOrder: 1, Version: 1, Status: "draft",
	}
	raw, err := s.MarshalJSONBlob()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseLessonSnapshot(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Title != s.Title {
		t.Fatalf("round trip failed: %+v", parsed)
	}
}
