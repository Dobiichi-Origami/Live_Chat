package entities

import "go.mongodb.org/mongo-driver/bson"

type Chat struct {
	Id       int64  `bson:"id"` // Id 必须是唯一索引
	Sequence uint64 `bson:"sequence"`
}

func NewChat(chatId int64, sequence uint64) *Chat {
	return &Chat{
		Id:       chatId,
		Sequence: sequence,
	}
}

func NewEmptyChat() *Chat {
	return &Chat{}
}

func NewChatFromChangeStreamBson(data bson.D) *Chat {
	return &Chat{
		Id:       data[1].Value.(int64),
		Sequence: uint64(data[2].Value.(int32)),
	}
}
