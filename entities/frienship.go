package entities

import (
	"database/sql"
	"errors"
)

type Friendship struct {
	SelfId   int64
	FriendId int64
	ChatId   int64
}

var (
	ScanFriendshipNoResult = errors.New("没有可供扫描的结果")
)

func NewFriendship(selfId, friendId, chatId int64) Friendship {
	return Friendship{
		SelfId:   selfId,
		FriendId: friendId,
		ChatId:   chatId,
	}
}

func ScanFriendshipsFromSqlResult(rows *sql.Rows) ([]Friendship, error) {
	var (
		retSlice  = make([]Friendship, 0)
		selfId    = int64(0)
		friendId  = int64(0)
		chatId    = int64(0)
		isDeleted = int8(0)
	)

	for rows.Next() {
		if err := rows.Scan(&selfId, &friendId, &chatId, &isDeleted); err != nil {
			return retSlice, err
		}
		retSlice = append(retSlice, NewFriendship(selfId, friendId, chatId))
	}

	return retSlice, nil
}
