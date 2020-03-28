package entities

type Chat struct {
	Receiver uint      `json:"receiver" gorm:"primary_key"`
	Title    string    `json:"title"`
	Messages []Message `json:"messages" gorm:"foreignKey:RecieverRefer"`
}
