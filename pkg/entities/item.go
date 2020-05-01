package entities

import "github.com/jinzhu/gorm"

type Item struct {
	gorm.Model
	RequestSender   uint   `json:"request_sender"`
	RequestReceiver uint   `json:"request_receiver"`
	Item            string `json:"item"`
	ReqDesc         string `json:"req_desc"`
}
