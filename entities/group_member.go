package entities

import (
	"gorm.io/gorm"
)

type GroupMember struct {
	gorm.Model
	GroupId         int64 `gorm:"uniqueIndex:group_info_index"`
	MemberId        int64 `gorm:"uniqueIndex:group_info_index;index:reverse_select_index"`
	IsAdministrator bool
	IsDeleted       bool `gorm:"uniqueIndex:group_info_index;index:reverse_select_index"`
}

func NewGroupMember(groupId, userId int64, isAdministrator bool) *GroupMember {
	return &GroupMember{
		GroupId:         groupId,
		MemberId:        userId,
		IsAdministrator: isAdministrator,
		IsDeleted:       false,
	}
}
