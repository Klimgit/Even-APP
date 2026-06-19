package internalapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/even-app/even-app/libs/clients/dto"
	"github.com/even-app/even-app/services/learning/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	q *query.Queries
}

func New(q *query.Queries) *Handler {
	return &Handler{q: q}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /courses/{courseId}/enrollments", h.listEnrollments)
	mux.HandleFunc("GET /enrollments", h.getEnrollment)
	mux.HandleFunc("POST /enrollments", h.createEnrollment)
	mux.HandleFunc("GET /users/{userId}/courses/{courseId}/progress", h.studentProgress)
	mux.HandleFunc("PUT /lesson-snapshots", h.upsertSnapshot)
	mux.HandleFunc("GET /stats/active-enrollments", h.activeEnrollmentsCount)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inner := http.NewServeMux()
	h.Mount(inner)
	inner.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg, "error": msg})
}

func (h *Handler) listEnrollments(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(r.PathValue("courseId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid course id")
		return
	}
	rows, err := h.q.ListEnrollmentsByCourse(r.Context(), query.ListEnrollmentsByCourseParams{CourseID: courseID})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]dto.EnrollmentView, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.EnrollmentView{
			UserID: row.UserID, CourseID: row.CourseID,
			EnrolledBy: row.EnrolledBy, Status: row.Status,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) getEnrollment(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user_id required")
		return
	}
	courseID, err := uuid.Parse(r.URL.Query().Get("course_id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "course_id required")
		return
	}
	row, err := h.q.GetEnrollmentByUserAndCourse(r.Context(), query.GetEnrollmentByUserAndCourseParams{
		UserID: userID, CourseID: courseID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto.EnrollmentView{
		ID: row.ID, UserID: row.UserID, CourseID: row.CourseID,
		EnrolledBy: row.EnrolledBy, Status: row.Status,
	})
}

func (h *Handler) createEnrollment(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	row, err := h.q.CreateEnrollment(r.Context(), query.CreateEnrollmentParams{
		UserID: req.UserID, CourseID: req.CourseID, EnrolledBy: req.EnrolledBy,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, dto.EnrollmentView{
		ID: row.ID, UserID: row.UserID, CourseID: row.CourseID,
		EnrolledBy: row.EnrolledBy, Status: row.Status,
	})
}

func (h *Handler) studentProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid user id")
		return
	}
	courseID, err := uuid.Parse(r.PathValue("courseId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid course id")
		return
	}
	rows, err := h.q.GetStudentLessonProgress(r.Context(), query.GetStudentLessonProgressParams{
		UserID: userID, CourseID: courseID,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]dto.LessonProgressRow, 0, len(rows))
	for _, row := range rows {
		score := float64(row.ScoreAvg)
		out = append(out, dto.LessonProgressRow{
			LessonID: row.LessonID, LessonTitle: row.Title,
			Completed:   int(row.CompletedBlocks) >= int(row.TotalBlocks) && row.TotalBlocks > 0,
			BlocksTotal: int(row.TotalBlocks),
			BlocksDone:  int(row.CompletedBlocks),
			ScoreAvg:    &score,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) upsertSnapshot(w http.ResponseWriter, r *http.Request) {
	var req dto.UpsertSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	err := h.q.UpsertPublishedLessonSnapshot(r.Context(), query.UpsertPublishedLessonSnapshotParams{
		LessonID: req.LessonID, CourseID: req.CourseID,
		Version: req.Version, Snapshot: req.Snapshot,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) activeEnrollmentsCount(w http.ResponseWriter, r *http.Request) {
	n, err := h.q.CountActiveEnrollments(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto.CountResponse{Count: int(n)})
}
