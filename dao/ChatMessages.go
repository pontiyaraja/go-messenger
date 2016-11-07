package dao

import (
	"time"
)

// This mongodb would be in shared collection model to support "Fan out"

type ChatMessage struct {
	Id          string    `bson:"_id" json:"chatMessageId"`
	ChatId      string    `bsong:"chatId" json:"chatId"`
	TimeStamp   time.Time `bsong:"timeStamp" json:"timeStamp"`
	FromUserId  string    `bson:"fromUserId" json:"fromUserId"`
	ToUserIds   []string  `bson:"toUserIds" json:"toUserIds"`
	RoomName    string    `bson:"roomName" json:"roomName"`
	MessageType string    `bson:"messageType" json:"messageType"`
	Message     string    `bson:"message" json:"message"`
	Attachment  string    `bson:"attachment" json:"attachment"`
	ROId        string    `bson:"roId,omitempty" json:"roId,omitempty"`
}
