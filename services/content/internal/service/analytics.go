package service

import (
	"context"

	"github.com/google/uuid"
)

type CourseAnalytics struct {
	StudentCount         int
	LessonCount          int
	ModuleCount          int
	AvgCompletionPercent float64
}

func (s *ContentService) GetCourseAnalytics(ctx context.Context, courseID, ownerID uuid.UUID, isAdmin bool) (CourseAnalytics, error) {
	if err := s.assertCourseOwner(ctx, courseID, ownerID, isAdmin); err != nil {
		return CourseAnalytics{}, err
	}
	modules, err := s.q.ListModulesByCourseID(ctx, courseID)
	if err != nil {
		return CourseAnalytics{}, err
	}
	lessons, err := s.q.ListLessonsByCourseID(ctx, courseID)
	if err != nil {
		return CourseAnalytics{}, err
	}
	out := CourseAnalytics{
		LessonCount: len(lessons),
		ModuleCount: len(modules),
	}
	if s.learning == nil || !s.learning.Available() {
		return out, nil
	}
	enrolls, err := s.learning.ListEnrollmentsByCourse(ctx, courseID)
	if err != nil {
		return CourseAnalytics{}, err
	}
	out.StudentCount = len(enrolls)
	if out.StudentCount == 0 {
		return out, nil
	}
	var totalPct float64
	for _, e := range enrolls {
		prog, err := s.GetStudentProgress(ctx, courseID, e.UserID, ownerID, isAdmin)
		if err != nil {
			continue
		}
		if len(prog.Lessons) == 0 {
			continue
		}
		var sum float64
		for _, l := range prog.Lessons {
			if l.TotalBlocks > 0 {
				sum += float64(l.CompletedBlocks) / float64(l.TotalBlocks) * 100
			}
		}
		totalPct += sum / float64(len(prog.Lessons))
	}
	out.AvgCompletionPercent = totalPct / float64(out.StudentCount)
	return out, nil
}
