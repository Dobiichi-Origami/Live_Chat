package entities

type GroupInfo struct {
	Id    int64 `gorm:"primaryKey"`
	Owner int64

	Name         string
	Introduction string
	Avatar       string
	IsDeleted    bool

	Members []GroupMember `gorm:"foreignKey:GroupId"`
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
