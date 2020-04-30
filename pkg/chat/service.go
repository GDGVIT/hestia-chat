package chat

import (
	"github.com/ATechnoHazard/hestia-chat/pkg"
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/jinzhu/gorm"
	"sort"
)

type Service interface {
	SaveMessage(msg *entities.Message) error
	GetMessages(to, from uint) ([]entities.Message, []entities.Item, error)
	CreateChat(chat *entities.Chat) error
	GetChatsByID(userID uint) ([]entities.Chat, error)
	GetMyChats(userID uint) ([]entities.Chat, error)
	GetOtherChats(userID uint) ([]entities.Chat, error)
	DeleteChat(receiver, sender uint, whoDeleted string) error
	UpdateChat(chat *entities.Chat) error
	GetChat(chat *entities.Chat) (*entities.Chat, error)
	AddItem(item *entities.Item) error
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
	err := tx.Where("request_receiver = ?", chat.RequestReceiver).Where("request_sender = ?", chat.RequestSender).Find(&entities.Chat{}).Error
	if err == nil {
		tx.Rollback()
		return pkg.ErrAlreadyExists
	}

	err = tx.Create(chat).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return pkg.ErrNotFound

		default:
			return pkg.ErrDatabase
		}
	}

	err = tx.Create(entities.Item{
		RequestSender:   chat.RequestSender,
		RequestReceiver: chat.RequestReceiver,
		Item:            chat.Title,
	}).Error
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

func (c *chatSvc) GetMessages(to, from uint) ([]entities.Message, []entities.Item, error) {
	tx := c.db.Begin()
	msgs := make([]entities.Message, 0)

	err := tx.Where("receiver_refer = ?", to).Where("sender = ?", from).Order("created_at").Find(&msgs).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
		default:
			tx.Rollback()
			return nil, nil, pkg.ErrDatabase
		}
	}

	msgs2 := make([]entities.Message, 0)

	err = tx.Where("receiver_refer = ?", from).Where("sender = ?", to).Order("created_at").Find(&msgs2).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
		default:
			tx.Rollback()
			return nil, nil, pkg.ErrDatabase
		}
	}
	msgs = append(msgs, msgs2...)

	items := make([]entities.Item, 0)
	err = tx.Where("request_sender = ?", from).Where("request_receiver = ?", to).Find(&items).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
		default:
			tx.Rollback()
			return nil, nil, pkg.ErrDatabase
		}
	}

	items2 := make([]entities.Item, 0)
	err = tx.Where("request_sender = ?", to).Where("request_receiver = ?", from).Find(&items2).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
		default:
			tx.Rollback()
			return nil, nil, pkg.ErrDatabase
		}
	}

	items = append(items, items2...)

	sort.Sort(entities.MessageSlice(msgs))

	tx.Commit()
	return msgs, items, nil
}

func (c *chatSvc) GetMyChats(userID uint) ([]entities.Chat, error) {
	tx := c.db.Begin()
	chats := make([]entities.Chat, 0)
	if err := tx.Where("request_receiver = ?", userID).Where("receiver_deleted = ?", false).Find(&chats).Error; err != nil {
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
	if err := tx.Where("request_sender = ?", userID).Where("sender_deleted = ?", false).Find(&chats).Error; err != nil {
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

func (c *chatSvc) DeleteChat(receiver, sender uint, whoDeleted string) error {
	tx := c.db.Begin()
	chat := &entities.Chat{RequestReceiver: receiver, RequestSender: sender}
	err := tx.Where("request_receiver = ?", receiver).Where("request_sender = ?", sender).Find(chat).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return pkg.ErrNotFound
		default:
			return pkg.ErrDatabase
		}
	}

	switch whoDeleted {
	case "receiver":
		chat.ReceiverDeleted = true
	case "sender":
		chat.SenderDeleted = true
	default:
		tx.Rollback()
		return pkg.ErrInvalidSlug
	}

	err = tx.Save(chat).Error
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

func (c *chatSvc) UpdateChat(chat *entities.Chat) error {
	tx := c.db.Begin()
	err := tx.Where("request_receiver = ?", chat.RequestReceiver).Where("request_sender = ?", chat.RequestSender).Find(&entities.Chat{}).Error
	if err != nil {
		tx.Rollback()
		switch err {
		case gorm.ErrRecordNotFound:
			return pkg.ErrNotFound
		default:
			return pkg.ErrDatabase
		}
	}

	err = tx.Save(chat).Error
	if err != nil {
		tx.Rollback()
		return pkg.ErrDatabase
	}

	tx.Commit()
	return nil
}

func (c *chatSvc) GetChat(chat *entities.Chat) (*entities.Chat, error) {
	tx := c.db.Begin()
	err := tx.Where("request_receiver = ? AND request_sender = ?", chat.RequestReceiver, chat.RequestSender).Or("request_receiver = ? AND request_sender = ?", chat.RequestSender, chat.RequestReceiver).Find(&entities.Chat{}).Error
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
	return chat, err
}

func (c *chatSvc) AddItem(item *entities.Item) error {
	tx := c.db.Begin()
	err := tx.Save(item).Error
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
