package entities

type GroupInfo struct {
	Id            int64   `bson:"id"`
	Owner         int64   `bson:"owner"`
	Administrator []int64 `bson:"administrator"`
	Member        []int64 `bson:"member"`

	Name         string `bson:"name"`
	Introduction string `bson:"instruction"`
	Avatar       string `bson:"avatar"`
}

func NewGroupInfo(id, owner int64, name, introduction, avatar string) *GroupInfo {
	return &GroupInfo{
		Id:            id,
		Owner:         owner,
		Administrator: make([]int64, 0),
		Member:        []int64{owner},
		Name:          name,
		Introduction:  introduction,
		Avatar:        avatar,
	}
}

func NewEmptyGroupInfo() *GroupInfo {
	return &GroupInfo{
		Administrator: make([]int64, 0, 0),
		Member:        []int64{},
	}
}
