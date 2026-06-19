package learning

import (
	"context"

	"github.com/even-app/even-app/libs/clients"
	"github.com/even-app/even-app/libs/clients/dto"
	"github.com/google/uuid"
)

type Client struct {
	http *clients.HTTP
}

func New(baseURL, token string) *Client {
	return &Client{http: clients.NewHTTP(baseURL, token)}
}

func (c *Client) Available() bool {
	return c != nil && c.http.Available()
}

func (c *Client) ListEnrollmentsByCourse(ctx context.Context, courseID uuid.UUID) ([]dto.EnrollmentView, error) {
	var out []dto.EnrollmentView
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/courses/"+courseID.String()+"/enrollments", nil, &out)
	return out, err
}

func (c *Client) CreateEnrollment(ctx context.Context, req dto.CreateEnrollmentRequest) (dto.EnrollmentView, error) {
	var out dto.EnrollmentView
	err := c.http.DoJSON(ctx, "POST", "/api/v1/internal/enrollments", req, &out)
	return out, err
}

func (c *Client) GetEnrollmentByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (dto.EnrollmentView, error) {
	var out dto.EnrollmentView
	path := "/api/v1/internal/enrollments?user_id=" + userID.String() + "&course_id=" + courseID.String()
	err := c.http.DoJSON(ctx, "GET", path, nil, &out)
	return out, err
}

func (c *Client) GetStudentLessonProgress(ctx context.Context, userID, courseID uuid.UUID) ([]dto.LessonProgressRow, error) {
	var out []dto.LessonProgressRow
	path := "/api/v1/internal/users/" + userID.String() + "/courses/" + courseID.String() + "/progress"
	err := c.http.DoJSON(ctx, "GET", path, nil, &out)
	return out, err
}

func (c *Client) UpsertPublishedLessonSnapshot(ctx context.Context, req dto.UpsertSnapshotRequest) error {
	return c.http.DoJSON(ctx, "PUT", "/api/v1/internal/lesson-snapshots", req, nil)
}

func (c *Client) ActiveEnrollmentsCount(ctx context.Context) (int, error) {
	var out dto.CountResponse
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/stats/active-enrollments", nil, &out)
	return out.Count, err
}
