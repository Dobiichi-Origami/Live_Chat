package entities

import "liveChat/rpc"

const (
	Add byte = iota
	Delete
)

const (
	Friend byte = iota
	Group
	Administrator
)

type Notification struct {
	SenderId     int64  `bson:"sender_id"`
	ReceiverId   int64  `bson:"receiver_id"`
	Seq          uint64 `bson:"sequence"`
	Timestamp    int64  `bson:"timestamp"`
	OpType       byte   `bson:"op_type"`
	ReceiveType  byte   `bson:"receive_type"`
	IsHandled    bool   `bson:"is_handled"`
	IsAgree      bool   `bson:"is_agree"`
	HandleUserId int64  `bson:"handle_user_id"`
}

func NewNotification(senderId int64, receiverId int64, opType, receiveType byte, isHandled, isAgree bool) *Notification {
	return &Notification{
		SenderId:    senderId,
		ReceiverId:  receiverId,
		Seq:         0,
		Timestamp:   0,
		OpType:      opType,
		ReceiveType: receiveType,
		IsHandled:   isHandled,
		IsAgree:     isAgree,
	}
}

func NewNotificationFromRpc(request *rpc.NotificationRequest) *Notification {
	return &Notification{
		SenderId:    request.Sender,
		ReceiverId:  request.Receiver,
		Seq:         request.Id,
		Timestamp:   int64(request.Timestamp),
		OpType:      byte(request.Op),
		ReceiveType: byte(request.ReceiveType),
		IsHandled:   request.IsHandledByAuth,
		IsAgree:     request.IsAgree,
	}
}
