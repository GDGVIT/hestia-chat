package entities

import "github.com/jinzhu/gorm"

type Message struct {
	gorm.Model
	ReceiverRefer uint   `json:"receiver" gorm:"not null"`
	Sender        uint   `json:"sender" gorm:"not null"`
	Text          string `json:"text" gorm:"not null"`
}


type MessageSlice []Message

func (m MessageSlice) Len() int {
	return len(m)
}

func (m MessageSlice) Less(i, j int) bool {
	return m[i].CreatedAt.Before(m[j].CreatedAt)
}

func (m MessageSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
