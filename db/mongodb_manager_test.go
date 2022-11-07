package db

import (
	"context"
	"liveChat/entities"
	"liveChat/protocol"
	"liveChat/tools"
	"reflect"
	"testing"
)

const (
	mockUserId1   = 1234
	mockUserName1 = "testUser"

	mockUserId2   = 5678
	mockUserName2 = "testUser2"

	mockUserID3   = 9012
	mockUserName3 = "testUser3"

	mockModifiedUserName = "modified_username"

	mockGroupName         = "testGroup"
	mockGroupIntroduction = "this is a test group"
	mockGroupAvatar       = "https://localhost:12345/this/is/a/path/to/pic"

	mockGroupModifiedName         = "modifiedGroup"
	mockGroupModifiedIntroduction = "this is a modified group"
	mockGroupModifiedAvatar       = "https://modified.com"

	mockMessageSegment1 = "这是一条测试消息"
	mockMessageSegment2 = ",这是一条测试消息2"
)

const testMongodbConfigFilePath = "../default_config_files/default_mongodb_config.json"

func TestCreateCollectionsAndIndexesAndConnection(t *testing.T) {
	if err := InitMongoDBConnection(testMongodbConfigFilePath, true); err != nil {
		t.Fatalf("init mongodb failed. reason: %s", err.Error())
	}
}

func TestUser(t *testing.T) {
	addAndCheckUserInfo(mockUserId1, mockUserName1, t)
	addAndCheckUserInfo(mockUserId2, mockUserName2, t)
	addFriends(mockUserId1, mockUserId2, t)
	modifyUsername(mockUserId1, mockModifiedUserName, t)
	deleteFriendship(mockUserId1, mockUserId2, t)
}

func TestGroup(t *testing.T) {
	groupId := createGroup(mockUserId1, mockGroupName, mockGroupIntroduction, mockGroupAvatar, t)
	joinAuthAndQuitGroup(mockUserId2, groupId, t)
	updateGroupInfoAndDeleteGroup(groupId, mockGroupModifiedName, mockGroupModifiedIntroduction, mockGroupModifiedAvatar, t)
	checkChatSeq(groupId, t)
	checkWatchAndRetrieve(groupId, t)
}

func TestDropDatabase(t *testing.T) {
	if err := dropDatabase(mongoDbCfg); err != nil {
		t.Fatalf("drop database failed. reason: %s", err.Error())
	}
}

func addAndCheckUserInfo(id int64, name string, t *testing.T) {
	if err := AddUserInfo(id, name); err != nil {
		t.Fatalf("Mongodb creates new user failed: %s", err.Error())
	}

	if userInfo, err := SearchUserInfo(id); err != nil {
		t.Fatalf("Mongodb finds user failed: %s", err.Error())
	} else if userInfo.Username != name {
		t.Fatalf("Mongodb username mismatched. expected: %s, receiveed: %s", name, userInfo.Username)
	}
}

func addFriends(id1, id2 int64, t *testing.T) {
	if err := AddFriendShip(id1, id2); err != nil {
		t.Fatalf("Mongodb adds friendship failed: %s", err.Error())
	}

	chatId1 := friendshipCheck(id1, id2, t)
	chatId2 := friendshipCheck(id2, id1, t)

	if chatId1 != chatId2 {
		t.Fatalf("Mongodb chat id dosen't equal. chat id1: %d, chat id2: %d", chatId1, chatId2)
	}

	if privateChat, err := SearchPrivateChatInfo(chatId1); err != nil {
		t.Fatalf("Mongodb finds private chat failed: %s", err.Error())
	} else if (privateChat.Members[0] != id1 && privateChat.Members[1] != id1) ||
		(privateChat.Members[0] != id2 && privateChat.Members[1] != id2) {
		t.Fatalf("Mongodb private chat members mismatched. slice: %v, id1: %d, id2: %d", privateChat.Members, id1, id2)
	}

	checkChatInfo(chatId1, t)
}

func friendshipCheck(id1, id2 int64, t *testing.T) int64 {
	userInfo, err := SearchUserInfo(id1)
	if err != nil {
		t.Fatalf("Mongodb finds user failed: %s", err.Error())
	} else if len(userInfo.Friendship) != 1 || userInfo.Friendship[0] != id2 {
		t.Fatalf("Mongodb search friendship failed. friend number: %d", len(userInfo.Friendship))
	} else if len(userInfo.PrivateChat) != 1 {
		t.Fatalf("Mongodb private chat number mismatched. expected: 1, received: %d", len(userInfo.PrivateChat))
	}

	return userInfo.PrivateChat[0]
}

func checkChatInfo(id int64, t *testing.T) {
	chatInfo := entities.NewEmptyChat()
	if err := findDocumentOne(context.Background(), getBson(ChatId, id), queueCollection, chatInfo); err != nil {
		t.Fatalf("Mongodb finds chatInfo failed: %s", err.Error())
	}

	if chatInfo.Sequence != 0 {
		t.Fatalf("Mongodb chatInfo sequence is %d instead of zero", chatInfo.Sequence)
	}
}

func modifyUsername(id int64, name string, t *testing.T) {
	if err := UpdateUserName(id, name); err != nil {
		t.Fatalf("Mongodb modify username failed: %s", err.Error())
	}

	if userInfo, err := SearchUserInfo(id); err != nil {
		t.Fatalf("Mongodb search user failed: %s", err.Error())
	} else if userInfo.Username != name {
		t.Fatalf("Mongodb modified username mismatched. expected: %s, received: %s", name, userInfo.Username)
	}
}

func deleteFriendship(id1, id2 int64, t *testing.T) {
	if err := DeleteFriendShip(id1, id2); err != nil {
		t.Fatalf("Mongodb deletes friendship failed: %s", err.Error())
	}

	userinfo1, err := SearchUserInfo(id1)
	if err != nil {
		t.Fatalf("Mongodb finds user failed: %s", err.Error())
	}

	userinfo2, err := SearchUserInfo(id2)
	if err != nil {
		t.Fatalf("Mongodb finds user failed: %s", err.Error())
	}

	if len(userinfo1.Friendship) != 0 || len(userinfo2.Friendship) != 0 {
		t.Fatalf("Mongodb deletes friendship list entry failed. list1 length: %d, list2 length: %d", len(userinfo1.Friendship), len(userinfo2.Friendship))
	} else if len(userinfo1.PrivateChat) != 1 || len(userinfo2.PrivateChat) != 1 {
		t.Fatalf("Mongodb deletes private chat list entry uncautionally. list1 length: %d, list2 length: %d", len(userinfo1.PrivateChat), len(userinfo2.PrivateChat))
	}

	if err = DeleteFriendShip(id1, id2); err == nil {
		t.Fatalf("Mongodb deletes a friendship doesn't exist")
	}
}

func createGroup(owner int64, name, instruction, avatar string, t *testing.T) int64 {
	groupId, err := AddGroupInfo(owner, name, instruction, avatar)
	if err != nil {
		t.Fatalf("Mongodb creats new group failed: %s", err.Error())
	}

	groupInfo, err := SearchGroupInfo(groupId)
	if err != nil {
		t.Fatalf("Mongodb finds group by id failed: %s", err.Error())
	}

	testGroupInfo := entities.NewGroupInfo(groupId, owner, name, instruction, avatar)

	if !reflect.DeepEqual(testGroupInfo, groupInfo) {
		t.Fatalf("Mongodb group info mismatched: expected: %+v, received: %+v", testGroupInfo, groupInfo)
	}

	return groupId
}

func joinAuthAndQuitGroup(id, groupId int64, t *testing.T) {
	if err := AddAdministrator(id, groupId); err == nil {
		t.Fatalf("Mongodb group sets administrator successfully unexpectedly")
	}

	if err := DeleteFromGroup(id, groupId); err == nil {
		t.Fatalf("Mongodb deletes user from group which it's not in it failed: %s", err.Error())
	}

	if err := JoinToGroup(id, groupId); err != nil {
		t.Fatalf("Mongodb join to group failed: %s", err.Error())
	}

	if err := AddAdministrator(id, groupId); err != nil {
		t.Fatalf("Mongodb group sets administrator failed: %s", err.Error())
	}

	groupInfo, err := SearchGroupInfo(groupId)
	if err != nil {
		t.Fatalf("Mongodb finds group by id failed: %s", err.Error())
	}

	if len(groupInfo.Member) != 2 {
		t.Fatalf("Mongodb group member number mismatched. received: %+v", groupInfo.Member)
	} else if len(groupInfo.Administrator) != 1 || groupInfo.Administrator[0] != id {
		t.Fatalf("Mongodb group administrator info mismatched. expected: %d received: %+v", id, groupInfo.Administrator)
	}

	if err = DeleteAdministrator(id, groupId); err != nil {
		t.Fatalf("Mongodb deletes administator failed: %s", err.Error())
	}

	if err = DeleteFromGroup(id, groupId); err != nil {
		t.Fatalf("Mongodb deletes user from group failed: %s", err.Error())
	}

	groupInfo, err = SearchGroupInfo(groupId)
	if err != nil {
		t.Fatalf("Mongodb finds group by id failed: %s", err.Error())
	}

	if len(groupInfo.Member) != 1 || len(groupInfo.Administrator) != 0 {
		t.Fatalf("Mongodb group info mismatched. member list: %+v, admin list: %+v", groupInfo.Member, groupInfo.Administrator)
	}
}

func updateGroupInfoAndDeleteGroup(id int64, name, introduction, avatar string, t *testing.T) {
	groupInfo, err := SearchGroupInfo(id)
	if err != nil {
		t.Fatalf("Mongodb finds group by id failed: %s", err.Error())
	}

	groupInfo.Name = name
	groupInfo.Introduction = introduction
	groupInfo.Avatar = avatar
	if err = UpdateGroupName(id, name); err != nil {
		t.Fatalf("Mongodb group modifies group name failed: %s", err.Error())
	}

	if err = UpdateGroupIntroduction(id, introduction); err != nil {
		t.Fatalf("Mongodb group modifies group isntruction failed: %s", err.Error())
	}

	if err = UpdateGroupAvatar(id, avatar); err != nil {
		t.Fatalf("Mongodb group modifies group avatar failed: %s", err.Error())
	}

	modifiedInfo, err := SearchGroupInfo(id)
	if err != nil {
		t.Fatalf("Mongodb finds group by id failed: %s", err.Error())
	}

	if !reflect.DeepEqual(modifiedInfo, groupInfo) {
		t.Fatalf("Mongodb group info mismatched between modifications. expected: %+v, receiced: %+v", groupInfo, modifiedInfo)
	}

	if err = DeleteGroupInfo(id); err != nil {
		t.Fatalf("Mongodb deletes group failed: %s", err.Error())
	}

	modifiedInfo, err = SearchGroupInfo(id)
	if err != nil {
		t.Fatalf("Mongodb finds group by id failed: %s", err.Error())
	}

	if userInfo, err := SearchUserInfo(modifiedInfo.Owner); err != nil {
		t.Fatalf("Mongodb finds user failed: %s", err.Error())
	} else if len(userInfo.Group) != 0 {
		t.Fatalf("Mongodb delete group from user failed")
	}
}

func checkChatSeq(chatId int64, t *testing.T) {
	if idSlice, err := GetChatSeqInSlice([]int64{chatId}); err != nil {
		t.Fatalf("Mongodb get chat sequence failed: %s", err.Error())
	} else if len(idSlice) != 1 || idSlice[0].Sequence != 0 {
		t.Fatalf("Mongodb caht sequence info mismatched: %+v", idSlice)
	}
}

func checkWatchAndRetrieve(chatId int64, t *testing.T) {
	watchSeq, err := SubscribeChatSeq(chatId)
	defer watchSeq.Close(context.Background())
	if err != nil {
		t.Fatalf("Mongodb watchSeq chat change failed: %s", err.Error())
	}

	watchMsg, err := SubscribeChatMessage(chatId)
	defer watchMsg.Close(context.Background())
	if err != nil {
		t.Fatalf("Mongodb watchMsg chat change failed: %s", err.Error())
	}

	protoMsg := &protocol.Message{
		Id:        0,
		Sender:    mockUserId1,
		Receiver:  chatId,
		Timestamp: 0,
		Type:      protocol.Message_Text,
		Contents:  []string{"这是一条测试消息"},
	}
	if err = AddMessage(protoMsg); err != nil {
		t.Fatalf("Mongodb add message failed: %s", err.Error())
	}

	if seq, err := GetChatSequence(chatId); err != nil {
		t.Fatalf("Mongodb get chat seq failed: %s", err.Error())
	} else if seq != 1 {
		t.Fatalf("Mongodb retrieved seq mismatched. target: 1, received: %d", seq)
	}

	ret, err := tools.FetchDataFromChangeStreamBson(watchSeq)
	if err != nil {
		t.Fatal(err.Error())
	}

	chat := entities.NewChatFromChangeStreamBson(tools.FetchBsonFromChangeStreamData(ret))
	if chat.Id != chatId {
		t.Fatalf("Mongodb chat id from change stream mismatched. target: %d, received: %d", chatId, chat.Id)
	} else if chat.Sequence != 1 {
		t.Fatalf("Mongodb chat seq from change stream mismatched. target: %d, received: %d", 1, chat.Sequence)
	}

	ret, err = tools.FetchDataFromChangeStreamBson(watchMsg)
	if err != nil {
		t.Fatal(err.Error())
	}

	msg := entities.NewMessageFromChangeStreamBson(tools.FetchBsonFromChangeStreamData(ret))
	targetMessage := entities.NewMessageFromProtobufWithoutSeq(protoMsg)
	targetMessage.Id = 0
	if !reflect.DeepEqual(targetMessage, msg) {
		t.Fatalf("Mongodb retrieved message mismatched. target: %v. received: %v", targetMessage, msg)
	}
}
