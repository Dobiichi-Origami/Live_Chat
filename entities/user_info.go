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
	Id               int64  `bson:"id"`
	Username         string `bson:"username"`
	UserAvatar       string
	UserIntroduction string
}

func NewUserInfo(id int64, userName string) *UserInfo {
	return &UserInfo{
		Id:       id,
		Username: userName,
	}
}

func NewEmptyUserInfo() *UserInfo {
	return &UserInfo{}
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
