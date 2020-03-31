package entities

type Chat struct {
	Receiver        uint      `json:"receiver" gorm:"primary_key"`
	Sender          uint      `json:"sender" gorm:"primary_key"`
	Title           string    `json:"title"`
	RequestReceiver uint      `json:"request_receiver"`
	RequestSender   uint      `json:"request_sender"`
	Messages        []Message `json:"messages" gorm:"foreignKey:RecieverRefer"`
}
