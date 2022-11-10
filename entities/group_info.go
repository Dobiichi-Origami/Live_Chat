package entities

import "database/sql"

type GroupInfo struct {
	Id    int64 `bson:"id"`
	Owner int64 `bson:"owner"`

	Name         string `bson:"name"`
	Introduction string `bson:"instruction"`
	Avatar       string `bson:"avatar"`
}

func NewGroupInfo(id, owner int64, name, introduction, avatar string) *GroupInfo {
	return &GroupInfo{
		Id:           id,
		Owner:        owner,
		Name:         name,
		Introduction: introduction,
		Avatar:       avatar,
	}
}

func NewEmptyGroupInfo() *GroupInfo {
	return &GroupInfo{}
}

func ScanGroupInfoFromSqlResult(rows *sql.Rows) (*GroupInfo, error) {
	var (
		groupId           = int64(0)
		groupOwner        = int64(0)
		groupName         = ""
		groupAvatar       = ""
		groupIntroduction = ""
		isDeleted         = int8(0)
	)

	if err := rows.Scan(&groupId, &groupName, &groupOwner, &groupIntroduction, &groupAvatar, &isDeleted); err != nil {
		return nil, err
	}

	return &GroupInfo{
		Id:           groupId,
		Owner:        groupOwner,
		Name:         groupName,
		Introduction: groupIntroduction,
		Avatar:       groupAvatar,
	}, nil
}
