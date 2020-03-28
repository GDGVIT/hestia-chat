package chat

import (
	"github.com/ATechnoHazard/hestia-chat/pkg"
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/jinzhu/gorm"
)

type Service interface {
	SaveMessage(msg *entities.Message) error
	GetMessages(chatID uint) ([]entities.Message, error)
	CreateChat(chat *entities.Chat) error
}

type chatSvc struct {
	db *gorm.DB
}

func NewChatService(db *gorm.DB) Service {
	return &chatSvc{db: db}
}

func (c *chatSvc) SaveMessage(msg *entities.Message) error {
	tx := c.db.Begin()
	err := tx.Create(msg).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return pkg.ErrNotFound
		default:
			return pkg.ErrDatabase
		}
	}
	tx.Commit()
	return nil
}

func (c *chatSvc) CreateChat(chat *entities.Chat) error {
	tx := c.db.Begin()
	err := tx.Create(chat).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return pkg.ErrNotFound
		default:
			return pkg.ErrDatabase
		}
	}
	tx.Commit()
	return nil
}

func (c *chatSvc) GetMessages(chatID uint) ([]entities.Message, error) {
	tx := c.db.Begin()
	chat := &entities.Chat{}
	err := tx.Where("receiver = ?", chatID).Find(chat).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, pkg.ErrNotFound
		default:
			return nil, pkg.ErrDatabase
		}
	}
	err = tx.Model(chat).Related(&chat.Messages, "ReceiverRefer").Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, pkg.ErrNotFound
		default:
			return nil, pkg.ErrDatabase
		}
	}

	tx.Commit()
	return chat.Messages, nil
}
