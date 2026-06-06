package domain

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type AuthOutcome struct {
	Tokens Tokens
	User   *User
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
