package db

import (
	"context"
	"fmt"
	"liveChat/entities"
	"liveChat/protocol"
	"liveChat/tools"
	"reflect"
	"testing"
)

const (
	mockEmail    = "test@out.com"
	mockPassword = "password"

	mockBadAccount = "test"
)

const testMysqlConfigFilePath = "../default_config_files/default_mysql_config.json"

const (
	mockUserName1 = "testUser"
	mockUserName2 = "testUser2"

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

var (
	mockUserId1 = int64(0)
	mockUserId2 = int64(0)
)

const testMongodbConfigFilePath = "../default_config_files/default_mongodb_config.json"

func TestMain(m *testing.M) {
	if err := InTestInitMysqlConnection(); err != nil {
		panic(err)
	}
	if err := InTestCreateCollectionsAndIndexesAndConnection(); err != nil {
		panic(err)
	}
	m.Run()
	dropDatabase(mongoDbCfg)

	deleteStrs := make([]string, 0)
	mysqlDb.Raw(fmt.Sprintf("SELECT concat('DROP TABLE IF EXISTS ', table_name, ';') FROM information_schema.tables WHERE table_schema = '%s';", mysqlCfg.Db)).Scan(&deleteStrs)
	mysqlDb.Exec("SET FOREIGN_KEY_CHECKS=0;")
	for _, str := range deleteStrs {
		mysqlDb.Exec(str)
	}
	mysqlDb.Exec("SET FOREIGN_KEY_CHECKS=1;")
}

func InTestInitMysqlConnection() error {
	return InitMysqlConnection(testMysqlConfigFilePath)
}

func InTestCreateCollectionsAndIndexesAndConnection() error {
	return InitMongoDBConnection(testMongodbConfigFilePath, true)
}

func TestRegister(t *testing.T) {
	if _, err := Register(mockUserName1, mockEmail, mockPassword); err != nil {
		t.Fatalf("Mysql register user failed: %s", err.Error())
	}

	if _, err := Register(mockUserName1, mockEmail, mockPassword); err == nil {
		t.Fatalf("Mysql register same user unexpected succeed")
	}

	if _, err := Register(mockUserName2, mockEmail, mockPassword); err != nil {
		t.Fatalf("Mysql register user failed: %s", err.Error())
	}
}

func TestLogin(t *testing.T) {
	var err = error(nil)
	mockUserId1, err = Login(mockUserName1, mockPassword)
	if err != nil {
		t.Fatalf("Mysql login failed: %s", err.Error())
	} else if mockUserId1 == -1 {
		t.Fatalf("Mysql login user not found")
	}

	if _, err = Login(mockBadAccount, mockPassword); err == nil {
		t.Fatalf("Mysql login wrong user found")
	}

	mockUserId2, err = Login(mockUserName2, mockPassword)
	if err != nil {
		t.Fatalf("Mysql login failed: %s", err.Error())
	} else if mockUserId1 == -1 {
		t.Fatalf("Mysql login user not found")
	}
}

func TestUser(t *testing.T) {
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

func addFriends(id1, id2 int64, t *testing.T) {
	chatId, err := AgreeFriendShip(id1, id2)
	if err != nil {
		t.Fatalf("Mysql adds friendship failed: %s", err.Error())
	}

	chatId1 := friendshipCheck(id1, id2, chatId, t)
	chatId2 := friendshipCheck(id2, id1, chatId, t)

	if chatId1 != chatId2 {
		t.Fatalf("Mysql chat id dosen't equal. chat id1: %d, chat id2: %d", chatId1, chatId2)
	}

	checkChatInfo(chatId, t)
}

func friendshipCheck(id1, id2, chatId int64, t *testing.T) int64 {
	friendships, err := SelectFriendShip(id1)
	if err != nil {
		t.Fatalf("Mysql find user failed: %s", err.Error())
	} else if len(friendships) != 1 || friendships[0].FriendId != id2 {
		t.Fatalf("Mysql search friendship failed. friend number: %d", len(friendships))
	} else if friendships[0].ChatId != chatId {
		t.Fatalf("Mysql private chat number mismatched. expected: %d, received: %d", chatId, friendships[0].ChatId)
	}

	return friendships[0].ChatId
}

func checkChatInfo(id int64, t *testing.T) {
	chatInfo := entities.NewEmptyChat()
	if err := findDocumentOne(context.Background(), getBson(ChatId, id), queueCollection, chatInfo); err != nil {
		t.Fatalf("Mysql finds chatInfo failed: %s", err.Error())
	}

	if chatInfo.Sequence != 0 {
		t.Fatalf("Mysql chatInfo sequence is %d instead of zero", chatInfo.Sequence)
	}
}

func modifyUsername(id int64, name string, t *testing.T) {
	if err := UpdateUserName(id, name); err != nil {
		t.Fatalf("Mysql modify username failed: %s", err.Error())
	}

	if userInfo, err := SearchUserInfo(id); err != nil {
		t.Fatalf("Mysql search user failed: %s", err.Error())
	} else if userInfo.Username != name {
		t.Fatalf("Mysql modified username mismatched. expected: %s, received: %s", name, userInfo.Username)
	}
}

func deleteFriendship(id1, id2 int64, t *testing.T) {
	friendships1, err := SelectFriendShip(id1)
	if err != nil {
		t.Fatalf("Mysql get friendships failed: %s", err.Error())
	}

	chatId1 := friendships1[0].ChatId

	if err := DeleteFriendShip(id1, id2); err != nil {
		t.Fatalf("Mysql deletes friendship failed: %s", err.Error())
	}

	friendships1, err = SelectFriendShip(id1)
	if err != nil {
		t.Fatalf("Mysql finds user failed: %s", err.Error())
	}

	friendships2, err := SelectFriendShip(id2)
	if err != nil {
		t.Fatalf("Mysql finds user failed: %s", err.Error())
	}

	if len(friendships1) != 0 || len(friendships2) != 0 {
		t.Fatalf("Mysql deletes friendship list entry failed. list1 length: %d, list2 length: %d", len(friendships1), len(friendships2))
	}

	chatId2, err := AgreeFriendShip(id1, id2)
	if err != nil {
		t.Fatalf("Mysql agree friendships again failed: %s", err.Error())
	}

	if chatId1 != chatId2 {
		t.Fatalf("Mysql repair friendship failed")
	}
}

func createGroup(owner int64, name, instruction, avatar string, t *testing.T) int64 {
	groupId, err := AddGroupInfo(owner, name, instruction, avatar)
	if err != nil {
		t.Fatalf("Mysql creats new group failed: %s", err.Error())
	}

	_, err = SearchGroupInfo(groupId)
	if err != nil {
		t.Fatalf("Mysql finds group by id failed: %s", err.Error())
	}

	return groupId
}

func joinAuthAndQuitGroup(id, groupId int64, t *testing.T) {
	if err := AddAdministrator(id, groupId); err == nil {
		t.Fatalf("Mysql group sets administrator successfully unexpectedly")
	}

	if err := DeleteFromGroup(id, groupId); err == nil {
		t.Fatalf("Mysql deletes user from group which it's not in it failed: %s", err.Error())
	}

	if err := AgreeJoinGroup(id, groupId); err != nil {
		t.Fatalf("Mysql join to group failed: %s", err.Error())
	}

	if err := AddAdministrator(id, groupId); err != nil {
		t.Fatalf("Mysql group sets administrator failed: %s", err.Error())
	}

	groupMember, err := SelectGroupMemberList(groupId)
	if err != nil {
		t.Fatalf("Mysql finds group by id failed: %s", err.Error())
	}

	if len(groupMember) != 2 {
		t.Fatalf("Mysql group member number mismatched. received: %+v", groupMember)
	}

	for _, e := range groupMember {
		if e.MemberId == id && !e.IsAdministrator {
			t.Fatalf("Mysql group administrator info mismatched.")
		}
	}

	if err = DeleteAdministrator(id, groupId); err != nil {
		t.Fatalf("Mysql deletes administator failed: %s", err.Error())
	}

	if err = DeleteFromGroup(id, groupId); err != nil {
		t.Fatalf("Mysql deletes user from group failed: %s", err.Error())
	}

	groupMember, err = SelectGroupMemberList(groupId)
	if err != nil {
		t.Fatalf("Mysql finds group by id failed: %s", err.Error())
	}

	if len(groupMember) != 2 {
		t.Fatalf("Mysql group member info mismatched. member list: %+v", groupMember)
	}
	for _, e := range groupMember {
		if e.MemberId == id && !e.IsDeleted {
			t.Fatalf("Mysql group delete member failed.")
		}
	}
}

func updateGroupInfoAndDeleteGroup(id int64, name, introduction, avatar string, t *testing.T) {
	groupInfo, err := SearchGroupInfo(id)
	if err != nil {
		t.Fatalf("Mysql finds group by id failed: %s", err.Error())
	}

	groupInfo.Name = name
	groupInfo.Introduction = introduction
	groupInfo.Avatar = avatar
	if err = UpdateGroupName(id, name); err != nil {
		t.Fatalf("Mysql group modifies group name failed: %s", err.Error())
	}

	if err = UpdateGroupIntroduction(id, introduction); err != nil {
		t.Fatalf("Mysql group modifies group isntruction failed: %s", err.Error())
	}

	if err = UpdateGroupAvatar(id, avatar); err != nil {
		t.Fatalf("Mysql group modifies group avatar failed: %s", err.Error())
	}

	modifiedInfo, err := SearchGroupInfo(id)
	if err != nil {
		t.Fatalf("Mysql finds group by id failed: %s", err.Error())
	}

	if !reflect.DeepEqual(modifiedInfo, groupInfo) {
		t.Fatalf("Mysql group info mismatched between modifications. expected: %+v, receiced: %+v", groupInfo, modifiedInfo)
	}

	if err = DeleteGroupInfo(id); err != nil {
		t.Fatalf("Mysql deletes group failed: %s", err.Error())
	}

	modifiedInfo, err = SearchGroupInfo(id)
	if err != nil {
		t.Fatalf("Mysql finds group by accientally")
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
