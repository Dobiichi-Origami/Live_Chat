package entities

import (
	"gorm.io/gorm"
)

type Friendship struct {
	gorm.Model
	SelfId    int64 `gorm:"uniqueIndex:friend_index;index:self_reverse_index"`
	FriendId  int64 `gorm:"uniqueIndex:friend_index"`
	IsDeleted bool  `gorm:"index:self_reverse_index"`
	ChatId    int64 `gorm:"index:chat_id_index"`
}

func NewFriendship(selfId, friendId, chatId int64) *Friendship {
	return &Friendship{
		SelfId:    selfId,
		FriendId:  friendId,
		ChatId:    chatId,
		IsDeleted: false,
	}
}
