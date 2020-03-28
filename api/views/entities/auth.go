package entities

type AuthResponse struct {
	Valid  bool `json:"valid"`
	UserID uint `json:"userID"`
}
