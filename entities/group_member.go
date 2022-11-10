package entities

import "database/sql"

type GroupMember struct {
	GroupId         int64
	MemberId        int64
	IsAdministrator bool
}

func NewGroupMember(groupId, userId int64, isAdministrator bool) GroupMember {
	return GroupMember{
		GroupId:         groupId,
		MemberId:        userId,
		IsAdministrator: isAdministrator,
	}
}

func ScanGroupMemberFromSqlResult(rows *sql.Rows) ([]GroupMember, error) {
	var (
		retSlice        = make([]GroupMember, 0)
		id              = int32(0)
		groupId         = int64(0)
		userId          = int64(0)
		isAdministrator = int8(0)
		isDeleted       = int8(0)
		boolFlag        = false
	)

	for rows.Next() {
		if err := rows.Scan(&id, &groupId, &userId, &isAdministrator, &isDeleted); err != nil {
			return retSlice, err
		}

		boolFlag = isAdministrator == 1
		retSlice = append(retSlice, NewGroupMember(groupId, userId, boolFlag))
	}

	return retSlice, nil
}
