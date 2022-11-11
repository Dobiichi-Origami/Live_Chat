package entities

import (
	"errors"
	"gorm.io/gorm"
)

type Friendship struct {
	gorm.Model
	SelfId    int64 `gorm:"uniqueIndex:friend_index;index:self_reverse_index"`
	FriendId  int64 `gorm:"uniqueIndex:friend_index"`
	IsDeleted bool  `gorm:"index:self_reverse_index"`
	ChatId    int64 `gorm:"index:chat_id_index"`
}

var (
	ScanFriendshipNoResult = errors.New("没有可供扫描的结果")
)

func NewFriendship(selfId, friendId, chatId int64) *Friendship {
	return &Friendship{
		SelfId:    selfId,
		FriendId:  friendId,
		ChatId:    chatId,
		IsDeleted: false,
	}
}
