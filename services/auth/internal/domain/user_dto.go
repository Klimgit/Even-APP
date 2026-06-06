package domain

import "time"

type UserDTO struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	IsAdmin     bool    `json:"is_admin"`
	CreatedAt   string  `json:"created_at"`
	DisplayName *string `json:"display_name,omitempty"`
}

func UserToDTO(u *User) UserDTO {
	dto := UserDTO{
		ID: u.ID.String(), Email: u.Email, Role: u.Role,
		IsAdmin: u.IsAdmin, CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
	}
	if u.DisplayName != nil {
		dto.DisplayName = u.DisplayName
	}
	return dto
}
