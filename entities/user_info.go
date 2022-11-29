package entities

import (
	"errors"
)

const (
	DefaultUserName         = "默认用户名"
	DefaultUserAvatar       = "url"
	DefaultUserIntroduction = ""
)

var (
	ScanUserInfoNoResult = errors.New("没有可供扫描的结果")
)

type UserInfo struct {
	Id               int64  `gorm:"primaryKey" json:"id"`
	Username         string `json:"username"`
	UserAvatar       string `json:"avatar"`
	UserIntroduction string `json:"introduction"`

	Friendships []Friendship  `gorm:"foreignKey:SelfId" json:"friendships"`
	Groups      []GroupMember `gorm:"foreignKey:MemberId" json:"groupList"`
}

func NewUserInfo(id int64, userName string) *UserInfo {
	return &UserInfo{
		Id:       id,
		Username: userName,
	}
}

func NewUserInfoWithDefaultValue(id int64) *UserInfo {
	return &UserInfo{
		Id:               id,
		Username:         DefaultUserName,
		UserAvatar:       DefaultUserAvatar,
		UserIntroduction: DefaultUserIntroduction,

		Friendships: make([]Friendship, 0, 0),
		Groups:      make([]GroupMember, 0, 0),
	}
}

func NewEmptyUserInfo() *UserInfo {
	return &UserInfo{
		Friendships: make([]Friendship, 0, 0),
		Groups:      make([]GroupMember, 0, 0),
	}
}
