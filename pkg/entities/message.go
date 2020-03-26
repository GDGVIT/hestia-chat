package entities

import "github.com/jinzhu/gorm"

type Message struct {
	gorm.Model
	ChatID uint   `json:"chat_id"`
	Text   string `json:"text"`
}
