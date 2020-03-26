package entities

import "github.com/jinzhu/gorm"

type Chat struct {
	gorm.Model
	Title    string    `json:"title"`
	Messages []Message `json:"-"`
}
