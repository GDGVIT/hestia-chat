package entities

type User struct {
	ID uint `json:"user_id"`
}

type UserDetails struct {
	Name string `json:"name"`
	ID   uint    `json:"id"`
}
