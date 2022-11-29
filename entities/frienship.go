package entities

import (
	"gorm.io/gorm"
)

type Friendship struct {
	GormModel gorm.Model `gorm:"embedded" json:"-"`
	SelfId    int64      `gorm:"uniqueIndex:friend_index;index:self_reverse_index" json:"selfId"`
	FriendId  int64      `gorm:"uniqueIndex:friend_index" json:"friendId"`
	IsDeleted bool       `gorm:"index:self_reverse_index" json:"-"`
	ChatId    int64      `gorm:"index:chat_id_index" json:"chatId"`
}

func NewFriendship(selfId, friendId, chatId int64) *Friendship {
	return &Friendship{
		SelfId:    selfId,
		FriendId:  friendId,
		ChatId:    chatId,
		IsDeleted: false,
	}
}
