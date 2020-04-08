package entities

type DelReq struct {
	Receiver   uint   `json:"receiver"`
	Sender     uint   `json:"sender"`
	WhoDeleted string `json:"who_deleted"`
}
