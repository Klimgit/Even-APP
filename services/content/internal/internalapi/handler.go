package internalapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/even-app/even-app/services/content/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	q   *query.Queries
	svc *service.ContentService
}

func New(q *query.Queries, svc *service.ContentService) *Handler {
	return &Handler{q: q, svc: svc}
}

func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /courses/by-invite", h.getCourseByInvite)
	mux.HandleFunc("GET /courses/published", h.listPublishedCourses)
	mux.HandleFunc("GET /courses/{courseId}", h.getCourseByID)
	mux.HandleFunc("GET /courses/{courseId}/published-lesson-snapshots", h.publishedLessonSnapshots)
	mux.HandleFunc("GET /courses/{courseId}/blocks", h.listBlocksByCourse)
	mux.HandleFunc("GET /lessons/{lessonId}/title", h.getLessonTitle)
	mux.HandleFunc("GET /lessons/{lessonId}/snapshot", h.getLessonSnapshot)
	mux.HandleFunc("GET /blocks/{blockId}", h.getBlock)
	mux.HandleFunc("GET /blocks/{blockId}/lesson-id", h.getBlockLessonID)
	mux.HandleFunc("GET /stats/published-courses", h.publishedCoursesCount)
	mux.HandleFunc("GET /stats/total-courses", h.totalCoursesCount)
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

func (h *Handler) getCourseByInvite(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(strings.ToUpper(r.URL.Query().Get("code")))
	if code == "" {
		writeErr(w, http.StatusBadRequest, "code required")
		return
	}
	row, err := h.q.GetCourseByInviteCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	invite := row.InviteCode
	writeJSON(w, http.StatusOK, map[string]any{
		"id": row.ID, "title": row.Title,
		"target_language_id": row.TargetLanguageID,
		"target_lang_code":   row.TargetLanguageCode,
		"target_lang_name":   row.TargetLanguageName,
		"ui_language_id":     row.UiLanguageID,
		"owner_id":           row.OwnerID,
		"is_published":       row.IsPublished,
		"visibility":         row.Visibility,
		"invite_code":        invite,
	})
}

func (h *Handler) getCourseByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("courseId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid course id")
		return
	}
	row, err := h.q.GetCourseByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": row.ID, "title": row.Title,
		"target_language_id": row.TargetLanguageID,
		"target_lang_code":   "",
		"target_lang_name":   "",
		"ui_language_id":     row.UiLanguageID,
		"owner_id":           row.OwnerID,
		"is_published":       row.IsPublished,
		"visibility":         row.Visibility,
	})
}

func (h *Handler) listPublishedCourses(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListPublishedCourses(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"id": row.ID, "title": row.Title,
			"target_language_id": row.TargetLanguageID,
			"target_lang_code":   row.TargetLanguageCode,
			"target_lang_name":   row.TargetLanguageName,
			"ui_language_id":     row.UiLanguageID,
			"owner_id":           row.OwnerID,
			"is_published":       row.IsPublished,
			"visibility":         row.Visibility,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) publishedLessonSnapshots(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(r.PathValue("courseId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid course id")
		return
	}
	snaps, err := h.svc.PublishedLessonSnapshots(r.Context(), courseID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snaps)
}

func (h *Handler) listBlocksByCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(r.PathValue("courseId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid course id")
		return
	}
	rows, err := h.q.ListBlocksByCourseID(r.Context(), courseID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"id": row.ID, "lesson_id": row.LessonID, "section_id": row.SectionID,
			"sort_order": row.SortOrder, "display_label": row.DisplayLabel, "title": row.Title,
			"block_type": row.BlockType, "config": json.RawMessage(row.Config), "is_homework": row.IsHomework,
			"lesson_title": row.LessonTitle, "lesson_sort_order": row.LessonSortOrder,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) getLessonTitle(w http.ResponseWriter, r *http.Request) {
	lessonID, err := uuid.Parse(r.PathValue("lessonId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid lesson id")
		return
	}
	title, err := h.q.GetLessonTitle(r.Context(), lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"title": title})
}

func (h *Handler) getLessonSnapshot(w http.ResponseWriter, r *http.Request) {
	lessonID, err := uuid.Parse(r.PathValue("lessonId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid lesson id")
		return
	}
	raw, err := h.svc.LessonSnapshotJSON(r.Context(), lessonID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]json.RawMessage{"snapshot": raw})
}

func (h *Handler) getBlock(w http.ResponseWriter, r *http.Request) {
	blockID, err := uuid.Parse(r.PathValue("blockId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid block id")
		return
	}
	view, err := h.svc.BlockView(r.Context(), blockID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *Handler) getBlockLessonID(w http.ResponseWriter, r *http.Request) {
	blockID, err := uuid.Parse(r.PathValue("blockId"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid block id")
		return
	}
	row, err := h.q.GetLessonBlockWithCourse(r.Context(), blockID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]uuid.UUID{"lesson_id": row.LessonID})
}

func (h *Handler) publishedCoursesCount(w http.ResponseWriter, r *http.Request) {
	n, err := h.q.CountPublishedCourses(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": int(n)})
}

func (h *Handler) totalCoursesCount(w http.ResponseWriter, r *http.Request) {
	n, err := h.q.CountTotalCourses(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": int(n)})
}
