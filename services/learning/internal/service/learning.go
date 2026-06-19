package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/even-app/even-app/services/learning/internal/domain"
	"github.com/even-app/even-app/services/learning/internal/gen/query"
	"github.com/even-app/even-app/services/learning/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LearningService struct {
	db      *pgxpool.Pool
	q       *query.Queries
	content *repository.ContentReader
	lexicon *repository.LexiconReader
	media   *repository.MediaReader
}

func NewLearningService(db *pgxpool.Pool, content *repository.ContentReader, lexicon *repository.LexiconReader, media *repository.MediaReader) *LearningService {
	return &LearningService{db: db, q: query.New(db), content: content, lexicon: lexicon, media: media}
}

type JoinResult struct {
	CourseID     uuid.UUID
	EnrollmentID uuid.UUID
}

func (s *LearningService) JoinCourse(ctx context.Context, userID uuid.UUID, inviteCode string) (*JoinResult, error) {
	code := strings.TrimSpace(strings.ToUpper(inviteCode))
	if code == "" {
		return nil, domain.ErrValidation
	}
	if !s.content.Available() {
		return nil, errors.New("content database not configured")
	}

	course, err := s.content.GetCourseByInviteCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if _, err := s.q.GetEnrollmentByUserAndCourse(ctx, query.GetEnrollmentByUserAndCourseParams{
		UserID: userID, CourseID: course.ID,
	}); err == nil {
		return nil, domain.ErrConflict
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	row, err := s.q.CreateEnrollment(ctx, query.CreateEnrollmentParams{
		UserID: userID, CourseID: course.ID, EnrolledBy: &userID,
	})
	if err != nil {
		return nil, err
	}

	if err := s.content.SyncPublishedLessonsForCourse(ctx, s.q, course.ID); err != nil {
		return nil, err
	}

	return &JoinResult{CourseID: course.ID, EnrollmentID: row.ID}, nil
}

func (s *LearningService) ListCourses(ctx context.Context, userID uuid.UUID) ([]CourseListItem, error) {
	enrolls, err := s.q.ListEnrollmentsByUser(ctx, query.ListEnrollmentsByUserParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	out := make([]CourseListItem, 0, len(enrolls))
	for _, e := range enrolls {
		item, err := s.courseListItem(ctx, userID, e.CourseID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

type CourseListItem struct {
	ID               uuid.UUID
	Title            string
	TargetLanguageID uuid.UUID
	TargetLangCode   string
	TargetLangName   string
	ProgressPercent  *float64
}

func (s *LearningService) GetCourse(ctx context.Context, userID, courseID uuid.UUID) (*domain.CourseView, error) {
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}
	return s.loadCourse(ctx, courseID)
}

func (s *LearningService) ListCourseLessons(ctx context.Context, userID, courseID uuid.UUID) ([]domain.LessonSummary, error) {
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}
	snaps, err := s.q.ListPublishedLessonSnapshotsByCourse(ctx, query.ListPublishedLessonSnapshotsByCourseParams{CourseID: courseID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.LessonSummary, 0, len(snaps))
	for _, snapRow := range snaps {
		snap, err := repository.SnapshotFromRow(snapRow.Snapshot)
		if err != nil {
			continue
		}
		pct := s.lessonCompletedPercent(ctx, userID, snap.ID)
		out = append(out, domain.LessonSummary{
			ID: snap.ID, Title: snap.Title, SortOrder: snap.SortOrder, CompletedPercent: pct,
		})
	}
	return out, nil
}

type CourseOutline struct {
	CourseID        uuid.UUID
	Title           string
	ProgressPercent float64
	CurrentLessonID *uuid.UUID
	CurrentBlockID  *uuid.UUID
	Lessons         []OutlineLesson
}

type OutlineLesson struct {
	ID              uuid.UUID
	Title           string
	SortOrder       int
	ProgressPercent float64
	Sections        []OutlineSection
}

type OutlineSection struct {
	ID              uuid.UUID
	Title           string
	SortOrder       int
	ProgressPercent float64
	Blocks          []OutlineBlock
}

type OutlineBlock struct {
	ID              uuid.UUID
	LessonID        uuid.UUID
	DisplayLabel    *string
	Title           string
	SortOrder       int
	BlockType       string
	IsGradable      bool
	Status          string
	ProgressPercent float64
	IsCurrent       bool
}

func (s *LearningService) GetCourseOutline(ctx context.Context, userID, courseID uuid.UUID) (*CourseOutline, error) {
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}
	course, err := s.loadCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	snaps, err := s.q.ListPublishedLessonSnapshotsByCourse(ctx, query.ListPublishedLessonSnapshotsByCourseParams{CourseID: courseID})
	if err != nil {
		return nil, err
	}
	counts, err := s.q.CountCompletedGradableBlocksForCourse(ctx, query.CountCompletedGradableBlocksForCourseParams{
		UserID: userID, CourseID: courseID,
	})
	coursePct := 0.0
	if err == nil && counts.Total > 0 {
		coursePct = float64(counts.Completed) / float64(counts.Total) * 100
	}

	outline := &CourseOutline{
		CourseID: courseID, Title: course.Title, ProgressPercent: coursePct,
		Lessons: make([]OutlineLesson, 0, len(snaps)),
	}
	var currentLessonID *uuid.UUID
	var currentBlockID *uuid.UUID

	for _, snapRow := range snaps {
		snap, err := repository.SnapshotFromRow(snapRow.Snapshot)
		if err != nil {
			continue
		}
		progRows, err := s.q.ListUserBlockProgressForLesson(ctx, query.ListUserBlockProgressForLessonParams{
			UserID: userID, LessonID: snap.ID,
		})
		if err != nil {
			return nil, err
		}
		byBlock := map[uuid.UUID]query.UserBlockProgress{}
		for _, r := range progRows {
			byBlock[r.LessonBlockID] = r
		}

		lesson := OutlineLesson{
			ID: snap.ID, Title: snap.Title, SortOrder: snap.SortOrder,
			ProgressPercent: s.lessonCompletedPercent(ctx, userID, snap.ID),
			Sections:        make([]OutlineSection, 0),
		}

		blocksBySection := map[uuid.UUID][]OutlineBlock{}
		orphanBlocks := make([]OutlineBlock, 0)

		for _, b := range snap.Blocks {
			ob := outlineBlockFromSnap(snap.ID, b, byBlock[b.ID])
			if b.SectionID != nil {
				blocksBySection[*b.SectionID] = append(blocksBySection[*b.SectionID], ob)
			} else {
				orphanBlocks = append(orphanBlocks, ob)
			}
		}

		for _, sec := range snap.Sections {
			secBlocks := blocksBySection[sec.ID]
			lesson.Sections = append(lesson.Sections, OutlineSection{
				ID: sec.ID, Title: sec.Title, SortOrder: sec.SortOrder,
				ProgressPercent: sectionProgressPercent(secBlocks),
				Blocks:          secBlocks,
			})
		}
		if len(orphanBlocks) > 0 {
			if len(lesson.Sections) == 0 {
				lesson.Sections = append(lesson.Sections, OutlineSection{
					ID: snap.ID, Title: "Content", SortOrder: 0,
					ProgressPercent: sectionProgressPercent(orphanBlocks),
					Blocks:          orphanBlocks,
				})
			} else {
				lesson.Sections[0].Blocks = append(orphanBlocks, lesson.Sections[0].Blocks...)
				lesson.Sections[0].ProgressPercent = sectionProgressPercent(lesson.Sections[0].Blocks)
			}
		}
		outline.Lessons = append(outline.Lessons, lesson)
	}

	currentLessonID, currentBlockID = pickCurrentOutlineBlock(outline.Lessons)

	if currentBlockID != nil {
		outline.CurrentLessonID = currentLessonID
		outline.CurrentBlockID = currentBlockID
		for li := range outline.Lessons {
			for si := range outline.Lessons[li].Sections {
				for bi := range outline.Lessons[li].Sections[si].Blocks {
					b := &outline.Lessons[li].Sections[si].Blocks[bi]
					b.IsCurrent = b.ID == *currentBlockID
				}
			}
		}
	}
	return outline, nil
}

func pickCurrentOutlineBlock(lessons []OutlineLesson) (lessonID, blockID *uuid.UUID) {
	for _, lesson := range lessons {
		for _, sec := range lesson.Sections {
			for _, b := range sec.Blocks {
				if !b.IsGradable {
					continue
				}
				if b.Status == "in_progress" || b.Status == "not_started" {
					id := b.ID
					lid := lesson.ID
					return &lid, &id
				}
			}
		}
	}
	return nil, nil
}

func outlineBlockFromSnap(lessonID uuid.UUID, b domain.BlockSnap, prog query.UserBlockProgress) OutlineBlock {
	title := b.BlockType
	if b.Title != nil && *b.Title != "" {
		title = *b.Title
	} else if b.DisplayLabel != nil && *b.DisplayLabel != "" {
		title = *b.DisplayLabel
	}
	ob := OutlineBlock{
		ID: b.ID, LessonID: lessonID, DisplayLabel: b.DisplayLabel, Title: title,
		SortOrder: b.SortOrder, BlockType: b.BlockType, IsGradable: b.IsGradable,
		Status: "not_started", ProgressPercent: 0, IsCurrent: false,
	}
	if !b.IsGradable {
		ob.Status = "completed"
		ob.ProgressPercent = 100
		return ob
	}
	if prog.LessonBlockID != uuid.Nil {
		ob.Status = prog.Status
		ob.ProgressPercent = blockProgressPercent(prog.Status, float64(prog.Score))
	}
	return ob
}

func blockProgressPercent(status string, score float64) float64 {
	switch status {
	case "completed":
		return 100
	case "in_progress":
		if score > 0 {
			return score * 100
		}
		return 50
	default:
		return 0
	}
}

func sectionProgressPercent(blocks []OutlineBlock) float64 {
	total := 0
	done := 0.0
	for _, b := range blocks {
		if !b.IsGradable {
			continue
		}
		total++
		done += b.ProgressPercent / 100
	}
	if total == 0 {
		return 100
	}
	return done / float64(total) * 100
}

func (s *LearningService) GetLesson(ctx context.Context, userID, lessonID uuid.UUID) (*domain.LessonSnapshot, error) {
	snap, courseID, err := s.loadLessonSnapshot(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *LearningService) GetLessonFlow(ctx context.Context, userID, lessonID uuid.UUID) (*LessonFlow, error) {
	snap, courseID, err := s.loadLessonSnapshot(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}

	flow := &LessonFlow{LessonID: snap.ID, Version: snap.Version, Items: []FlowItem{}}
	gradableCount := 0
	for _, block := range snap.Blocks {
		flow.Items = append(flow.Items, FlowItem{Kind: "lesson_block", Block: block})
		if block.IsGradable {
			gradableCount++
			if gradableCount%3 == 0 {
				if review, err := s.q.GetNextDueReviewItem(ctx, query.GetNextDueReviewItemParams{UserID: userID}); err == nil {
					revBlock, _, _, err := s.loadBlockForReview(ctx, review)
					if err == nil {
						title := ""
						if revBlock.Title != nil {
							title = *revBlock.Title
						}
						srcTitle, _ := s.lessonTitle(ctx, review.SourceLessonID)
						flow.Items = append(flow.Items, FlowItem{
							Kind: "review_injection",
							Review: &ReviewItemView{
								LessonBlockID: review.LessonBlockID, SubItemIndex: int(review.SubItemIndex),
								BlockType: revBlock.BlockType, Title: title,
								SourceLessonID: review.SourceLessonID, SourceLessonTitle: srcTitle,
								FailureCount: int(review.FailureCount), DueAt: review.DueAt,
								Status: review.Status, Block: revBlock,
							},
						})
					}
				}
			}
		}
	}
	return flow, nil
}

type LessonFlow struct {
	LessonID uuid.UUID
	Version  int
	Items    []FlowItem
}

type FlowItem struct {
	Kind   string
	Block  domain.BlockSnap
	Review *ReviewItemView
}

type ReviewItemView struct {
	LessonBlockID     uuid.UUID
	SubItemIndex      int
	BlockType         string
	Title             string
	SourceLessonID    uuid.UUID
	SourceLessonTitle string
	FailureCount      int
	DueAt             time.Time
	Status            string
	Block             domain.BlockSnap
}

type AttemptInput struct {
	SubItemIndex int
	Response     map[string]any
	Context      string
}

type AttemptResult struct {
	IsCorrect     bool
	Score         float64
	CorrectAnswer any
	BlockProgress BlockProgress
}

type BlockProgress struct {
	LessonBlockID uuid.UUID
	Status        string
	Score         float64
	Attempts      int
}

func (s *LearningService) SubmitBlockAttempt(ctx context.Context, userID, blockID uuid.UUID, in AttemptInput) (*AttemptResult, error) {
	block, courseID, _, err := s.loadBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}
	if !block.IsGradable {
		return nil, domain.ErrValidation
	}

	ctxVal := in.Context
	if ctxVal == "" {
		ctxVal = "lesson"
	}
	subIdx := in.SubItemIndex
	grade := domain.GradeBlock(block.BlockType, block.Config, subIdx, in.Response)

	respJSON, err := json.Marshal(in.Response)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	if err := qtx.InsertBlockAttempt(ctx, query.InsertBlockAttemptParams{
		UserID: userID, LessonBlockID: blockID, SubItemIndex: int32(subIdx),
		IsCorrect: grade.Correct, Response: respJSON, Context: ctxVal,
	}); err != nil {
		return nil, err
	}

	prev, prevErr := qtx.GetUserBlockProgress(ctx, query.GetUserBlockProgressParams{
		UserID: userID, LessonBlockID: blockID,
	})
	attempts := 1
	prevScore := grade.Score
	status := "in_progress"
	if grade.Correct {
		status = "completed"
	}
	if prevErr == nil {
		attempts = int(prev.Attempts) + 1
		if grade.Score > float64(prev.Score) {
			prevScore = grade.Score
		} else {
			prevScore = float64(prev.Score)
		}
		if grade.Correct || prev.Status == "completed" {
			status = "completed"
		}
	}

	prog, err := qtx.UpsertUserBlockProgress(ctx, query.UpsertUserBlockProgressParams{
		UserID: userID, LessonBlockID: blockID, Status: status,
		Score: float32(prevScore), Attempts: int32(attempts),
	})
	if err != nil {
		return nil, err
	}

	sourceLessonID := s.blockLessonID(ctx, blockID)
	if !grade.Correct {
		dueAt := domain.ReviewDueAt(1, time.Now())
		if err := qtx.UpsertReviewItemOnFailure(ctx, query.UpsertReviewItemOnFailureParams{
			UserID: userID, LessonBlockID: blockID, SubItemIndex: int32(subIdx),
			DueAt: dueAt, SourceLessonID: sourceLessonID, CourseID: courseID,
		}); err != nil {
			return nil, err
		}
	} else {
		_ = qtx.UpdateReviewItemOnSuccess(ctx, query.UpdateReviewItemOnSuccessParams{
			UserID: userID, LessonBlockID: blockID, SubItemIndex: int32(subIdx),
		})
		for _, ref := range block.LexemeRefs {
			if ref.Role == "introduced" || ref.Role == "prompt" || ref.Role == "answer" {
				_ = qtx.UpsertVocabularyOnSuccess(ctx, query.UpsertVocabularyOnSuccessParams{
					UserID: userID, LexemeID: ref.LexemeID, CourseID: &courseID, Mastery: float32(grade.Score),
				})
			}
		}
		if lexID := extractLexemeFromConfig(block.Config); lexID != uuid.Nil {
			_ = qtx.UpsertVocabularyOnSuccess(ctx, query.UpsertVocabularyOnSuccessParams{
				UserID: userID, LexemeID: lexID, CourseID: &courseID, Mastery: float32(grade.Score),
			})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	result := &AttemptResult{
		IsCorrect: grade.Correct, Score: grade.Score,
		BlockProgress: BlockProgress{
			LessonBlockID: blockID, Status: prog.Status,
			Score: float64(prog.Score), Attempts: int(prog.Attempts),
		},
	}
	if !grade.Correct {
		result.CorrectAnswer = grade.CorrectAnswer
	}
	return result, nil
}

func (s *LearningService) GetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID) (*LessonProgress, error) {
	snap, courseID, err := s.loadLessonSnapshot(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if err := s.requireEnrollment(ctx, userID, courseID); err != nil {
		return nil, err
	}
	rows, err := s.q.ListUserBlockProgressForLesson(ctx, query.ListUserBlockProgressForLessonParams{
		UserID: userID, LessonID: lessonID,
	})
	if err != nil {
		return nil, err
	}
	byBlock := map[uuid.UUID]query.UserBlockProgress{}
	for _, r := range rows {
		byBlock[r.LessonBlockID] = r
	}
	blocks := make([]BlockProgress, 0, len(snap.Blocks))
	for _, b := range snap.Blocks {
		if !b.IsGradable {
			continue
		}
		bp := BlockProgress{LessonBlockID: b.ID, Status: "not_started", Score: 0, Attempts: 0}
		if r, ok := byBlock[b.ID]; ok {
			bp.Status = r.Status
			bp.Score = float64(r.Score)
			bp.Attempts = int(r.Attempts)
		}
		blocks = append(blocks, bp)
	}
	return &LessonProgress{LessonID: lessonID, Blocks: blocks}, nil
}

type LessonProgress struct {
	LessonID uuid.UUID
	Blocks   []BlockProgress
}

type ReviewList struct {
	PendingCount int
	DueCount     int
	Items        []ReviewItemView
}

func (s *LearningService) ListReview(ctx context.Context, userID uuid.UUID, status string, dueOnly bool) (*ReviewList, error) {
	var st *string
	if status != "" {
		st = &status
	}
	counts, err := s.q.CountReviewItems(ctx, query.CountReviewItemsParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	rows, err := s.q.ListReviewItems(ctx, query.ListReviewItemsParams{
		UserID: userID, Status: st, DueOnly: dueOnly,
	})
	if err != nil {
		return nil, err
	}
	items := make([]ReviewItemView, 0, len(rows))
	for _, row := range rows {
		view, err := s.reviewRowToView(ctx, row)
		if err != nil {
			continue
		}
		items = append(items, view)
	}
	return &ReviewList{
		PendingCount: int(counts.PendingCount),
		DueCount:     int(counts.DueCount),
		Items:        items,
	}, nil
}

type VocabularyEntry struct {
	LexemeID   uuid.UUID
	Lemma      string
	LanguageID uuid.UUID
	FirstSeen  time.Time
	Mastery    float64
}

func (s *LearningService) ListDictionary(ctx context.Context, userID uuid.UUID, courseID *uuid.UUID) ([]VocabularyEntry, error) {
	rows, err := s.q.ListVocabulary(ctx, query.ListVocabularyParams{
		UserID: userID, CourseID: courseID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]VocabularyEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, VocabularyEntry{
			LexemeID: r.LexemeID, Lemma: r.LexemeID.String()[:8],
			LanguageID: uuid.Nil, FirstSeen: r.FirstSeenAt, Mastery: float64(r.Mastery),
		})
	}
	return out, nil
}

func (s *LearningService) requireEnrollment(ctx context.Context, userID, courseID uuid.UUID) error {
	ok, err := s.q.HasEnrollment(ctx, query.HasEnrollmentParams{UserID: userID, CourseID: courseID})
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *LearningService) loadCourse(ctx context.Context, courseID uuid.UUID) (*domain.CourseView, error) {
	if s.content.Available() {
		c, err := s.content.GetCourseByID(ctx, courseID)
		if err == nil {
			s.enrichCourseLanguage(ctx, &c)
			return &c, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	return nil, domain.ErrNotFound
}

func (s *LearningService) enrichCourseLanguage(ctx context.Context, c *domain.CourseView) {
	if c == nil || !s.lexicon.Available() || c.TargetLanguageID == uuid.Nil {
		return
	}
	lang, err := s.lexicon.GetLanguage(ctx, c.TargetLanguageID)
	if err != nil {
		return
	}
	c.TargetLangCode = lang.Code
	c.TargetLangName = lang.Name
}

type PublicCourseListItem struct {
	ID             uuid.UUID
	Title          string
	TargetLangCode string
	TargetLangName string
	LanguageID     uuid.UUID
	IsPublished    bool
	InviteCode     string
}

func (s *LearningService) ListPublicCourses(ctx context.Context) ([]PublicCourseListItem, error) {
	if !s.content.Available() {
		return nil, errors.New("content database not configured")
	}
	courses, err := s.content.ListPublishedCourses(ctx)
	if err != nil {
		return nil, err
	}
	langIDs := make([]uuid.UUID, 0, len(courses))
	for _, c := range courses {
		if c.TargetLanguageID != uuid.Nil {
			langIDs = append(langIDs, c.TargetLanguageID)
		}
	}
	langs, _ := s.lexiconLanguages(ctx, langIDs)
	out := make([]PublicCourseListItem, 0, len(courses))
	for _, c := range courses {
		item := PublicCourseListItem{
			ID: c.ID, Title: c.Title, LanguageID: c.TargetLanguageID,
			IsPublished: c.IsPublished, InviteCode: c.InviteCode,
		}
		if lang, ok := langs[c.TargetLanguageID]; ok {
			item.TargetLangCode = lang.Code
			item.TargetLangName = lang.Name
		}
		out = append(out, item)
	}
	return out, nil
}

type ProgressSummary struct {
	EnrolledCourses  int
	DictionaryWords  int
	ReviewDue        int
	CompletedBlocks  int
	CompletedLessons int
}

func (s *LearningService) GetProgressSummary(ctx context.Context, userID uuid.UUID) (ProgressSummary, error) {
	row, err := s.q.GetProgressSummary(ctx, query.GetProgressSummaryParams{UserID: userID})
	if err != nil {
		return ProgressSummary{}, err
	}
	return ProgressSummary{
		EnrolledCourses:  int(row.EnrolledCourses),
		DictionaryWords:  int(row.DictionaryWords),
		ReviewDue:        int(row.ReviewDue),
		CompletedBlocks:  int(row.CompletedBlocks),
		CompletedLessons: int(row.CompletedLessons),
	}, nil
}

type ReviewSession struct {
	Item         ReviewItemView
	PendingCount int
	DueCount     int
}

func (s *LearningService) StartReviewSession(ctx context.Context, userID uuid.UUID) (*ReviewSession, error) {
	counts, err := s.q.CountReviewItems(ctx, query.CountReviewItemsParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	row, err := s.q.GetNextDueReviewItem(ctx, query.GetNextDueReviewItemParams{UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	item, err := s.reviewRowToView(ctx, row)
	if err != nil {
		return nil, err
	}
	return &ReviewSession{
		Item:         item,
		PendingCount: int(counts.PendingCount),
		DueCount:     int(counts.DueCount),
	}, nil
}

func (s *LearningService) LexemeLookup(ctx context.Context, snap *domain.LessonSnapshot) map[uuid.UUID]repository.LexemeView {
	ids := collectLexemeIDs(snap)
	if len(ids) == 0 || !s.lexicon.Available() {
		return nil
	}
	out, err := s.lexicon.LexemesByIDs(ctx, ids)
	if err != nil {
		return nil
	}
	return out
}

func (s *LearningService) MediaLookup(ctx context.Context, snap *domain.LessonSnapshot) map[uuid.UUID]repository.MediaView {
	ids := collectMediaIDs(snap)
	if len(ids) == 0 || !s.media.Available() {
		return nil
	}
	out, err := s.media.MediaByIDs(ctx, ids)
	if err != nil {
		return nil
	}
	return out
}

func collectLexemeIDs(snap *domain.LessonSnapshot) []uuid.UUID {
	if snap == nil {
		return nil
	}
	seen := make(map[uuid.UUID]struct{})
	for _, b := range snap.Blocks {
		collectLexemeIDsFromConfig(b.Config, seen)
	}
	out := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}

func collectLexemeIDsFromConfig(cfg json.RawMessage, seen map[uuid.UUID]struct{}) {
	var raw any
	if err := json.Unmarshal(cfg, &raw); err != nil {
		return
	}
	walkLexemeIDs(raw, seen)
}

func walkLexemeIDs(v any, seen map[uuid.UUID]struct{}) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if k == "lexeme_id" || k == "lexemeId" {
				if s, ok := val.(string); ok {
					if id, err := uuid.Parse(s); err == nil {
						seen[id] = struct{}{}
					}
				}
			}
			if k == "lexeme_ids" || k == "lexemeIds" {
				if arr, ok := val.([]any); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							if id, err := uuid.Parse(s); err == nil {
								seen[id] = struct{}{}
							}
						}
					}
				}
			}
			walkLexemeIDs(val, seen)
		}
	case []any:
		for _, item := range t {
			walkLexemeIDs(item, seen)
		}
	}
}

func collectMediaIDs(snap *domain.LessonSnapshot) []uuid.UUID {
	if snap == nil {
		return nil
	}
	seen := make(map[uuid.UUID]struct{})
	for _, b := range snap.Blocks {
		collectMediaIDsFromConfig(b.Config, seen)
	}
	out := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}

func collectMediaIDsFromConfig(cfg json.RawMessage, seen map[uuid.UUID]struct{}) {
	var raw any
	if err := json.Unmarshal(cfg, &raw); err != nil {
		return
	}
	walkMediaIDs(raw, seen)
}

func walkMediaIDs(v any, seen map[uuid.UUID]struct{}) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if k == "media_asset_id" || k == "mediaAssetId" {
				if s, ok := val.(string); ok {
					if id, err := uuid.Parse(s); err == nil {
						seen[id] = struct{}{}
					}
				}
			}
			walkMediaIDs(val, seen)
		}
	case []any:
		for _, item := range t {
			walkMediaIDs(item, seen)
		}
	}
}

func (s *LearningService) lexiconLanguages(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]repository.LanguageView, error) {
	if !s.lexicon.Available() {
		return map[uuid.UUID]repository.LanguageView{}, nil
	}
	return s.lexicon.LanguagesByIDs(ctx, ids)
}

func (s *LearningService) courseListItem(ctx context.Context, userID, courseID uuid.UUID) (CourseListItem, error) {
	c, err := s.loadCourse(ctx, courseID)
	if err != nil {
		return CourseListItem{}, err
	}
	counts, err := s.q.CountCompletedGradableBlocksForCourse(ctx, query.CountCompletedGradableBlocksForCourseParams{
		UserID: userID, CourseID: courseID,
	})
	var pct *float64
	if err == nil && counts.Total > 0 {
		v := float64(counts.Completed) / float64(counts.Total)
		pct = &v
	}
	return CourseListItem{
		ID: c.ID, Title: c.Title,
		TargetLanguageID: c.TargetLanguageID,
		TargetLangCode:   c.TargetLangCode, TargetLangName: c.TargetLangName,
		ProgressPercent: pct,
	}, nil
}

func (s *LearningService) loadLessonSnapshot(ctx context.Context, lessonID uuid.UUID) (*domain.LessonSnapshot, uuid.UUID, error) {
	row, err := s.q.GetPublishedLessonSnapshot(ctx, query.GetPublishedLessonSnapshotParams{LessonID: lessonID})
	if err == nil {
		snap, err := repository.SnapshotFromRow(row.Snapshot)
		if err != nil {
			return nil, uuid.Nil, err
		}
		return &snap, row.CourseID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, uuid.Nil, err
	}
	if s.content.Available() {
		snap, err := s.content.BuildLessonSnapshot(ctx, lessonID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, uuid.Nil, domain.ErrNotFound
			}
			return nil, uuid.Nil, err
		}
		raw, _ := snap.MarshalJSONBlob()
		_ = s.q.UpsertPublishedLessonSnapshot(ctx, query.UpsertPublishedLessonSnapshotParams{
			LessonID: lessonID, CourseID: snap.CourseID, Version: int32(snap.Version), Snapshot: raw,
		})
		return &snap, snap.CourseID, nil
	}
	return nil, uuid.Nil, domain.ErrNotFound
}

func (s *LearningService) loadBlock(ctx context.Context, blockID uuid.UUID) (domain.BlockSnap, uuid.UUID, string, error) {
	if s.content.Available() {
		return s.content.GetBlock(ctx, blockID)
	}
	hit, err := repository.FindBlockInSnapshots(ctx, s.db, blockID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.BlockSnap{}, uuid.Nil, "", domain.ErrNotFound
		}
		return domain.BlockSnap{}, uuid.Nil, "", err
	}
	block, err := repository.BlockFromSnapshotHit(hit, blockID)
	if err != nil {
		return domain.BlockSnap{}, uuid.Nil, "", domain.ErrNotFound
	}
	return block, hit.CourseID, "", nil
}

func (s *LearningService) loadBlockForReview(ctx context.Context, review query.UserReviewItem) (domain.BlockSnap, uuid.UUID, string, error) {
	if s.content.Available() {
		return s.content.GetBlock(ctx, review.LessonBlockID)
	}
	snapRow, err := s.q.GetPublishedLessonSnapshot(ctx, query.GetPublishedLessonSnapshotParams{LessonID: review.SourceLessonID})
	if err != nil {
		return domain.BlockSnap{}, uuid.Nil, "", err
	}
	snap, err := repository.SnapshotFromRow(snapRow.Snapshot)
	if err != nil {
		return domain.BlockSnap{}, uuid.Nil, "", err
	}
	for _, b := range snap.Blocks {
		if b.ID == review.LessonBlockID {
			return b, snapRow.CourseID, snap.Title, nil
		}
	}
	return domain.BlockSnap{}, uuid.Nil, "", domain.ErrNotFound
}

func (s *LearningService) reviewRowToView(ctx context.Context, row query.UserReviewItem) (ReviewItemView, error) {
	block, _, _, err := s.loadBlockForReview(ctx, row)
	if err != nil {
		return ReviewItemView{}, err
	}
	title := ""
	if block.Title != nil {
		title = *block.Title
	}
	srcTitle, _ := s.lessonTitle(ctx, row.SourceLessonID)
	return ReviewItemView{
		LessonBlockID: row.LessonBlockID, SubItemIndex: int(row.SubItemIndex),
		BlockType: block.BlockType, Title: title,
		SourceLessonID: row.SourceLessonID, SourceLessonTitle: srcTitle,
		FailureCount: int(row.FailureCount), DueAt: row.DueAt,
		Status: row.Status, Block: block,
	}, nil
}

func (s *LearningService) lessonTitle(ctx context.Context, lessonID uuid.UUID) (string, error) {
	if title, err := s.q.GetLessonTitleFromSnapshot(ctx, query.GetLessonTitleFromSnapshotParams{LessonID: lessonID}); err == nil {
		return title, nil
	}
	if s.content.Available() {
		return s.content.GetLessonTitle(ctx, lessonID)
	}
	return "", nil
}

func (s *LearningService) lessonCompletedPercent(ctx context.Context, userID, lessonID uuid.UUID) float64 {
	counts, err := s.q.CountCompletedGradableBlocksForLesson(ctx, query.CountCompletedGradableBlocksForLessonParams{
		UserID: userID, LessonID: lessonID,
	})
	if err != nil || counts.Total == 0 {
		return 0
	}
	return float64(counts.Completed) / float64(counts.Total)
}

func (s *LearningService) blockLessonID(ctx context.Context, blockID uuid.UUID) uuid.UUID {
	if s.content.Available() {
		if id, err := s.content.GetBlockLessonID(ctx, blockID); err == nil {
			return id
		}
	}
	if row, err := repository.FindBlockInSnapshots(ctx, s.db, blockID); err == nil {
		return row.LessonID
	}
	return uuid.Nil
}

func extractLexemeFromConfig(raw json.RawMessage) uuid.UUID {
	var cfg map[string]any
	if json.Unmarshal(raw, &cfg) != nil {
		return uuid.Nil
	}
	for _, k := range []string{"lexeme_id", "prompt_lexeme_id", "correct_lexeme_id"} {
		if v, ok := cfg[k]; ok {
			if s, ok := v.(string); ok {
				if id, err := uuid.Parse(s); err == nil {
					return id
				}
			}
		}
	}
	return uuid.Nil
}
