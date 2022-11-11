package entities

import (
	"database/sql"
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
	Id               int64 `gorm:"primaryKey"`
	Username         string
	UserAvatar       string
	UserIntroduction string

	Friendships []Friendship  `gorm:"foreignKey:SelfId"`
	Groups      []GroupMember `gorm:"foreignKey:MemberId"`
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

func ScanUserInfoFromSqlResult(rows *sql.Rows) (*UserInfo, error) {
	if !rows.Next() {
		return nil, ScanUserInfoNoResult
	}

	var (
		userId           = int64(0)
		userName         = ""
		userAvatar       = ""
		userIntroduction = ""
	)

	if err := rows.Scan(&userId, &userName, &userAvatar, &userIntroduction); err != nil {
		return nil, err
	}

	return &UserInfo{
		Id:               userId,
		Username:         userName,
		UserAvatar:       userAvatar,
		UserIntroduction: userIntroduction,
	}, nil
}
