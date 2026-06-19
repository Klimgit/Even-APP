package domain

const (
	CourseVisibilityPublic     = "public"
	CourseVisibilityInviteOnly = "invite_only"
)

func ValidCourseVisibility(v string) bool {
	return v == CourseVisibilityPublic || v == CourseVisibilityInviteOnly
}
