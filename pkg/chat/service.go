package chat

import (
	"github.com/ATechnoHazard/hestia-chat/pkg"
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/jinzhu/gorm"
)

type Service interface {
	SaveMessage(msg *entities.Message) error
	GetMessages(to, from uint) ([]entities.Message, error)
	CreateChat(chat *entities.Chat) error
	GetChatsByID(userID uint) ([]entities.Chat, error)
	GetMyChats(userID uint) ([]entities.Chat, error)
	GetOtherChats(userID uint) ([]entities.Chat, error)
}

type chatSvc struct {
	db *gorm.DB
}

func (c *chatSvc) GetChatsByID(from uint) ([]entities.Chat, error) {
	tx := c.db.Begin()
	chats := make([]entities.Chat, 0)
	err := tx.Where("sender = ? OR receiver = ?", from, from).Find(&chats).Error
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
	return chats, nil
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

func (c *chatSvc) GetMessages(to, from uint) ([]entities.Message, error) {
	tx := c.db.Begin()
	msgs := make([]entities.Message, 0)

	err := tx.Where("receiver_refer = ?", to).Where("sender = ?", from).Find(&msgs).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, pkg.ErrNotFound
		default:
			return nil, pkg.ErrDatabase
		}
	}

	msgs2 := make([]entities.Message, 0)

	err = tx.Where("receiver_refer = ?", from).Where("sender = ?", to).Find(&msgs2).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			tx.Commit()
			return msgs, nil
		default:
			tx.Rollback()
			return nil, pkg.ErrDatabase
		}
	}
	msgs = append(msgs, msgs2...)

	tx.Commit()
	return msgs, nil
}

func (c *chatSvc) GetMyChats(userID uint) ([]entities.Chat, error) {
	tx := c.db.Begin()
	chats := make([]entities.Chat, 0)
	if err := tx.Where("request_receiver = ?", userID).Find(&chats).Error; err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, pkg.ErrNotFound
		default:
			return nil, pkg.ErrDatabase
		}
	}
	tx.Commit()
	return chats, nil
}

func (c *chatSvc) GetOtherChats(userID uint) ([]entities.Chat, error) {
	tx := c.db.Begin()
	chats := make([]entities.Chat, 0)
	if err := tx.Where("request_sender = ?", userID).Find(&chats).Error; err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, pkg.ErrNotFound
		default:
			return nil, pkg.ErrDatabase
		}
	}
	tx.Commit()
	return chats, nil
}