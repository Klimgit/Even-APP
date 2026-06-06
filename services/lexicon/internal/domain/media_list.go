package domain

type MediaListResponse struct {
	Items []MediaAssetDTO `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}
