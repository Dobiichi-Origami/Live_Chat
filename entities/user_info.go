package entities

type UserInfo struct {
	Id       int64  `bson:"id"`
	Username string `bson:"username"`

	Friendship  []int64 `bson:"friendship"`
	PrivateChat []int64 `bson:"private_chat"`
	Group       []int64 `bson:"group"`
}

func NewUserInfo(id int64, userName string) *UserInfo {
	return &UserInfo{
		Id:          id,
		Username:    userName,
		Friendship:  make([]int64, 0),
		PrivateChat: make([]int64, 0),
		Group:       make([]int64, 0),
	}
}

func NewEmptyUserInfo() *UserInfo {
	return &UserInfo{
		Friendship:  make([]int64, 0),
		PrivateChat: make([]int64, 0),
		Group:       make([]int64, 0),
	}
}
