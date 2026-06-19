package repository

import (
	"context"

	contentclient "github.com/even-app/even-app/libs/clients/content"
	learningclient "github.com/even-app/even-app/libs/clients/learning"
)

type StatsSource interface {
	PublishedCourses(ctx context.Context) (int, error)
	ActiveEnrollments(ctx context.Context) (int, error)
}

var (
	_ StatsSource = (*PlatformStatsReader)(nil)
	_ StatsSource = (*PlatformStatsRemote)(nil)
)

type PlatformStatsRemote struct {
	content  *contentclient.Client
	learning *learningclient.Client
}

func NewPlatformStatsRemote(contentURL, learningURL, token string) *PlatformStatsRemote {
	return &PlatformStatsRemote{
		content:  contentclient.New(contentURL, token),
		learning: learningclient.New(learningURL, token),
	}
}

func (r *PlatformStatsRemote) PublishedCourses(ctx context.Context) (int, error) {
	if r.content == nil || !r.content.Available() {
		return 0, nil
	}
	return r.content.PublishedCoursesCount(ctx)
}

func (r *PlatformStatsRemote) ActiveEnrollments(ctx context.Context) (int, error) {
	if r.learning == nil || !r.learning.Available() {
		return 0, nil
	}
	return r.learning.ActiveEnrollmentsCount(ctx)
}
