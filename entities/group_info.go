package entities

import "time"

type GroupInfo struct {
	Id    int64 `gorm:"primaryKey" json:"id"`
	Owner int64 `json:"ownerId"`

	Name         string `json:"name"`
	Introduction string `json:"introduction"`
	Avatar       string `json:"avatar"`
	IsDeleted    bool   `json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Members []GroupMember `gorm:"foreignKey:GroupId" json:"members"`
}

func NewGroupInfo(id, owner int64, name, introduction, avatar string) *GroupInfo {
	return &GroupInfo{
		Id:           id,
		Owner:        owner,
		Name:         name,
		Introduction: introduction,
		Avatar:       avatar,
		Members:      make([]GroupMember, 0),
	}
}

func NewEmptyGroupInfo() *GroupInfo {
	return &GroupInfo{
		Members: make([]GroupMember, 0),
	}
}
