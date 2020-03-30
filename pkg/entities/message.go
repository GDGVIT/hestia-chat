package entities

import "github.com/jinzhu/gorm"

type Message struct {
	gorm.Model
	ReceiverRefer uint   `json:"receiver" gorm:"not null"`
	Sender        uint   `json:"sender" gorm:"not null"`
	Text          string `json:"text" gorm:"not null"`
}
