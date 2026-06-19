package content

import (
	"context"
	"encoding/json"
	"net/url"

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

func (c *Client) GetCourseByInviteCode(ctx context.Context, code string) (dto.CourseView, error) {
	var out dto.CourseView
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/courses/by-invite?code="+url.QueryEscape(code), nil, &out)
	return out, err
}

func (c *Client) GetCourseByID(ctx context.Context, id uuid.UUID) (dto.CourseView, error) {
	var out dto.CourseView
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/courses/"+id.String(), nil, &out)
	return out, err
}

func (c *Client) ListPublishedCourses(ctx context.Context) ([]dto.CourseView, error) {
	var out []dto.CourseView
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/courses/published", nil, &out)
	return out, err
}

func (c *Client) PublishedLessonSnapshots(ctx context.Context, courseID uuid.UUID) ([]dto.PublishedLessonSnapshot, error) {
	var out []dto.PublishedLessonSnapshot
	path := "/api/v1/internal/courses/" + courseID.String() + "/published-lesson-snapshots"
	err := c.http.DoJSON(ctx, "GET", path, nil, &out)
	return out, err
}

func (c *Client) GetBlock(ctx context.Context, blockID uuid.UUID) (dto.BlockView, error) {
	var out dto.BlockView
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/blocks/"+blockID.String(), nil, &out)
	return out, err
}

func (c *Client) GetLessonTitle(ctx context.Context, lessonID uuid.UUID) (string, error) {
	var out struct {
		Title string `json:"title"`
	}
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/lessons/"+lessonID.String()+"/title", nil, &out)
	return out.Title, err
}

func (c *Client) GetBlockLessonID(ctx context.Context, blockID uuid.UUID) (uuid.UUID, error) {
	var out struct {
		LessonID uuid.UUID `json:"lesson_id"`
	}
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/blocks/"+blockID.String()+"/lesson-id", nil, &out)
	return out.LessonID, err
}

func (c *Client) ListBlocksByCourseID(ctx context.Context, courseID uuid.UUID) ([]dto.CourseBlockRow, error) {
	var out []dto.CourseBlockRow
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/courses/"+courseID.String()+"/blocks", nil, &out)
	return out, err
}

func (c *Client) PublishedCoursesCount(ctx context.Context) (int, error) {
	var out dto.CountResponse
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/stats/published-courses", nil, &out)
	return out.Count, err
}

func (c *Client) TotalCoursesCount(ctx context.Context) (int, error) {
	var out dto.CountResponse
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/stats/total-courses", nil, &out)
	return out.Count, err
}

func (c *Client) LessonSnapshotJSON(ctx context.Context, lessonID uuid.UUID) (json.RawMessage, error) {
	var out struct {
		Snapshot json.RawMessage `json:"snapshot"`
	}
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/lessons/"+lessonID.String()+"/snapshot", nil, &out)
	return out.Snapshot, err
}
