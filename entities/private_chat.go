package entities

type PrivateChatInfo struct {
	Id      int64    `bson:"id"`
	Members [2]int64 `bson:"members"`
}

func NewPrivateChatInfo(id, member0, member1 int64) *PrivateChatInfo {
	return &PrivateChatInfo{
		Id:      id,
		Members: [2]int64{member0, member1},
	}
}

func NewEmptyPrivateChatInfo() *PrivateChatInfo {
	return &PrivateChatInfo{
		Members: [2]int64{},
	}
}
