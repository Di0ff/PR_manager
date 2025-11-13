package dto

type Team struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	} `json:"members"`
}
