package entities

import (
	"gorm.io/gorm"
)

type GroupMember struct {
	GormModel       gorm.Model `gorm:"embedded" json:"-"`
	GroupId         int64      `gorm:"uniqueIndex:group_info_index" json:"groupId"`
	MemberId        int64      `gorm:"uniqueIndex:group_info_index;index:reverse_select_index" json:"memberId"`
	IsAdministrator bool       `json:"isAdministrator"`
	IsDeleted       bool       `gorm:"uniqueIndex:group_info_index;index:reverse_select_index" json:"-"`
}

func NewGroupMember(groupId, userId int64, isAdministrator bool) *GroupMember {
	return &GroupMember{
		GroupId:         groupId,
		MemberId:        userId,
		IsAdministrator: isAdministrator,
		IsDeleted:       false,
	}
}
