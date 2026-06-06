package domain

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
}

func AuthOutcomeToResponse(outcome *AuthOutcome) AuthResponse {
	return AuthResponse{
		AccessToken:  outcome.Tokens.AccessToken,
		RefreshToken: outcome.Tokens.RefreshToken,
		User:         UserToDTO(outcome.User),
	}
}

func TokensToResponse(tokens *Tokens) AuthTokensResponse {
	return AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
}
