package entities

type Chat struct {
	Receiver        uint      `json:"receiver" gorm:"primary_key"`
	Sender          uint      `json:"sender" gorm:"primary_key"`
	RequestReceiver uint      `json:"request_receiver"`
	RequestSender   uint      `json:"request_sender"`
	Title           string    `json:"title"`
	ReqDesc         string    `json:"req_desc"`
	SenderName      string    `json:"sender_name"`
	ReceiverName    string    `json:"receiver_name"`
	Messages        []Message `json:"messages" gorm:"foreignKey:RecieverRefer"`
	ReceiverDeleted bool      `json:"receiver_deleted"`
	SenderDeleted   bool      `json:"sender_deleted"`
	IsReported      bool      `json:"is_reported" gorm:"default:false"`
}
