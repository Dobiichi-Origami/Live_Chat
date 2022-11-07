package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"liveChat/config"
	"liveChat/constants"
	"liveChat/entities"
	"liveChat/protocol"
	"liveChat/tools"
)

const defaultMongoDBConfigPath = "./mongodb_config.json"

var MongoDBConfigPath = defaultMongoDBConfigPath

var (
	mongoDbCfg *config.MongoDBConfig

	mongoConnection *mongo.Client = nil

	// 存放所有消息的集合对象
	messageCollection *mongo.Collection = nil
	// 存放所有会话最新消息序号的集合对象
	queueCollection *mongo.Collection = nil
	// 存放所有用户信息的集合对象
	userInfoCollection *mongo.Collection = nil
	// 存放所有聊天信息的集合对象
	chatInfoCollection *mongo.Collection = nil

	isMongodbInitiated bool = false
)

const (
	mongoDbSet         = "$set"
	mongoDbIncr        = "$inc"
	mongoDbPush        = "$push"
	mongoDbPull        = "$pull"
	mongoDbSetInInsert = "$setOnInsert"

	mongoDbIn           = "$in"
	mongoDbMatch        = "$match"
	mongoDbAnd          = "$and"
	mongoDbOr           = "$or"
	mongoDbGreater      = "$gt"
	mongoDbGreaterEqual = "$gte"
	mongoDbLess         = "$lt"
	mongoDbLessEqual    = "$lte"
)

const (
	UserId          = "id"
	UserName        = "username"
	UserFriendShip  = "friendship"
	UserPrivateChat = "private_chat"
	UserGroup       = "group"

	GroupId            = "id"
	GroupAdministrator = "administrator"
	GroupMember        = "member"
	GroupName          = "name"
	GroupInstruction   = "instruction"
	GroupAvatar        = "avatar"

	PrivateChatId      = "id"
	PrivateChatMembers = "members"

	ChatId       = "id"
	ChatSequence = "sequence"

	MessageId       = "id"
	MessageReceiver = "receiver"
)

const (
	filterChatSeq = "fullDocument." + ChatId
	filterMessage = "fullDocument." + MessageReceiver
)

const chatTypeMask = 1 << 63

func InitMongoDBConnection(configPath string, needToInitDb bool) error {
	if isMongodbInitiated {
		return nil
	}

	path := tools.GetPath(MongoDBConfigPath, configPath)
	mongoDbCfg = config.NewMongoDBConfig(path)
	url := mongoDbCfg.Format()

	var err error
	mongoConnection, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(url), getDefaultMongoConcern())
	if err != nil {
		return err
	}

	err = mongoConnection.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}

	if needToInitDb {
		if err = createCollectionsAndIndexes(mongoDbCfg); err != nil {
			return err
		}
	}

	initConnection(mongoDbCfg)
	isMongodbInitiated = true
	return nil
}

func AddUserInfo(id int64, userName string) error {
	return insertDocumentOne(context.Background(), entities.NewUserInfo(id, userName), userInfoCollection)
}

func SearchUserInfo(id int64) (*entities.UserInfo, error) {
	ptr := entities.NewEmptyUserInfo()
	if err := findDocumentOne(context.Background(), getBson(UserId, id), userInfoCollection, ptr); err != nil {
		return nil, err
	}

	return ptr, nil
}

func UpdateUserName(userId int64, userName string) error {
	filter := getBson(UserId, userId)
	update := getOpBson(mongoDbSet, UserName, userName)
	return updateDocumentOne(context.Background(), filter, update, userInfoCollection)
}

func AddFriendShip(userId1, userId2 int64) error {
	sess, err := mongoConnection.StartSession()
	if err != nil {
		return err
	}

	hook := getFriendOpHook(mongoDbPush, userId1, userId2)
	if err = execTransaction(sess, hook); err != nil {
		return err
	}
	return nil
}

func DeleteFriendShip(userId1, userId2 int64) error {
	sess, err := mongoConnection.StartSession()
	if err != nil {
		return err
	}

	hook := getFriendOpHook(mongoDbPull, userId1, userId2)
	if err = execTransaction(sess, hook); err != nil {
		return err
	}
	return nil
}

func AddGroupInfo(owner int64, name, introduction, avatar string) (int64, error) {
	groupId := tools.GenerateSnowflakeId(true)
	hook := func(sessCtx mongo.SessionContext) error {
		info := entities.NewGroupInfo(groupId, owner, name, introduction, avatar)
		info.Member = []int64{owner}

		if err := insertDocumentOne(sessCtx, info, chatInfoCollection); err != nil {
			return err
		}

		if err := addChat(sessCtx, groupId); err != nil {
			return nil
		}

		return updateDocumentOne(sessCtx, getBson(UserId, owner), getOpBson(mongoDbPush, UserGroup, groupId), userInfoCollection)
	}

	sess, err := mongoConnection.StartSession()
	if err != nil {
		return -1, err
	}

	if err = execTransaction(sess, hook); err != nil {
		return -1, err
	}

	return groupId, nil
}

func SearchGroupInfo(id int64) (*entities.GroupInfo, error) {
	ptr := entities.NewEmptyGroupInfo()
	if err := findDocumentOne(context.Background(), getBson(GroupId, id), chatInfoCollection, ptr); err != nil {
		return nil, err
	}

	return ptr, nil
}

func DeleteGroupInfo(groupId int64) error {
	info, err := SearchGroupInfo(groupId)
	if err != nil {
		return err
	}

	hook := func(sessCtx mongo.SessionContext) error {
		for _, userId := range info.Member {
			filter := getBson(UserId, userId)
			update := getOpBson(mongoDbPull, UserGroup, groupId)
			if err = updateDocumentOne(sessCtx, filter, update, userInfoCollection); err != nil {
				return err
			}
		}

		// 保留聊天关系，后期方便查聊天记录
		// 这个功能有待商榷
		//_, err = chatInfoCollection.DeleteOne(sessCtx, getBson(GroupId, groupId))
		//if err != nil {
		//	return err
		//}

		return nil
	}

	sess, err := mongoConnection.StartSession(nil)
	if err != nil {
		return err
	}

	return execTransaction(sess, hook)
}

func JoinToGroup(userId, groupId int64) error {
	return joinOrQuitGroup(mongoDbPush, userId, groupId)
}

func DeleteFromGroup(userId, groupId int64) error {
	return joinOrQuitGroup(mongoDbPull, userId, groupId)
}

func AddAdministrator(userId, groupId int64) error {
	return authOrDeAuthAdmin(mongoDbPush, userId, groupId)
}

func DeleteAdministrator(userId, groupId int64) error {
	return authOrDeAuthAdmin(mongoDbPull, userId, groupId)
}

func UpdateGroupName(groupId int64, name string) error {
	filter := getBson(GroupId, groupId)
	update := getOpBson(mongoDbSet, GroupName, name)
	return updateDocumentOne(context.Background(), filter, update, chatInfoCollection)
}

func UpdateGroupIntroduction(groupId int64, introdution string) error {
	filter := getBson(GroupId, groupId)
	update := getOpBson(mongoDbSet, GroupInstruction, introdution)
	return updateDocumentOne(context.Background(), filter, update, chatInfoCollection)
}

func UpdateGroupAvatar(groupId int64, avatar string) error {
	filter := getBson(GroupId, groupId)
	update := getOpBson(mongoDbSet, GroupAvatar, avatar)
	return updateDocumentOne(context.Background(), filter, update, chatInfoCollection)
}

func SearchPrivateChatInfo(chatId int64) (*entities.PrivateChatInfo, error) {
	privateChat := entities.NewEmptyPrivateChatInfo()
	if err := findDocumentOne(context.Background(), getBson(PrivateChatId, chatId), chatInfoCollection, privateChat); err != nil {
		return nil, err
	}
	return privateChat, nil
}

func GetChatSeqInSlice(chatId []int64) ([]entities.Chat, error) {
	filter := bson.D{{ChatId, bson.D{{mongoDbIn, chatId}}}}
	cursor, err := queueCollection.Find(context.Background(), filter, nil)
	if err != nil {
		return nil, err
	}

	seqSlice := make([]entities.Chat, 0)
	if err = decodeDataInCursor(cursor, &seqSlice); err != nil {
		return nil, err
	}
	return seqSlice, nil
}

func GetChatSequence(chatId int64) (uint64, error) {
	chat := entities.NewEmptyChat()
	if err := findDocumentOne(context.Background(), getBson(ChatId, chatId), queueCollection, chat); err != nil {
		return constants.MaxUInt64, err
	}

	return chat.Sequence, nil
}

func GetAndAddChatSequence(chatId int64) (uint64, error) {
	chat := entities.NewEmptyChat()
	result := queueCollection.FindOneAndUpdate(
		context.Background(),
		getBson(ChatId, chatId),
		bson.D{{mongoDbIncr, bson.D{{ChatSequence, 1}}}, {mongoDbSetInInsert, bson.D{{ChatId, chatId}}}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	if result.Err() == nil {
		if err := result.Decode(&chat); err == nil {
			return chat.Sequence, nil
		} else {
			return constants.MaxUInt64, err
		}
	}

	if result.Err() != mongo.ErrNoDocuments {
		return constants.MaxUInt64, result.Err()
	}

	return 0, nil
}

func GetMessageInSeqRange(chatId int64, bottom, top uint64) ([]entities.Message, error) {
	cursor, err := messageCollection.Find(context.Background(),
		bson.D{{MessageReceiver, chatId}, {MessageId, bson.D{{mongoDbGreaterEqual, bottom}, {mongoDbLessEqual, top}}}},
		nil,
	)

	if err != nil {
		return nil, err
	}

	messageSlice := make([]entities.Message, 0)
	if err = decodeDataInCursor(cursor, &messageSlice); err != nil {
		return nil, err
	}
	return messageSlice, nil
}

func GetMessageInSeq(chatId int64, seq uint64) (*entities.Message, error) {
	message := entities.NewEmptyMessage()
	if err := findDocumentOne(context.Background(), bson.D{{MessageReceiver, chatId}, {MessageId, seq}}, messageCollection, message); err != nil {
		return nil, err
	}
	return message, nil
}

func CreateMessageWithSeq(m *protocol.Message) (*entities.Message, error) {
	seq, err := GetAndAddChatSequence(m.Receiver)
	if err != nil {
		return nil, err
	}

	message := entities.NewMessageFromProtobufWithoutSeq(m)
	message.Id = seq
	return message, nil
}

func AddMessage(m *protocol.Message) error {
	message, err := CreateMessageWithSeq(m)
	if err != nil {
		return err
	}
	if err = insertDocumentOne(context.Background(), message, messageCollection); err != nil {
		return err
	}
	// 转移到其他业务文件中处理 CacheMessageWithTimeOut(message)
	return nil
}

func SubscribeChatSeq(chatId int64) (*mongo.ChangeStream, error) {
	return queueCollection.Watch(context.Background(),
		mongo.Pipeline{bson.D{{mongoDbMatch, bson.D{{mongoDbAnd, bson.A{bson.D{{"operationType", bson.D{{mongoDbIn, bson.A{"insert", "update"}}}}},
			bson.D{{filterChatSeq, chatId}}}}}}}},
		options.ChangeStream().SetFullDocument(options.UpdateLookup))
}

func SubscribeChatMessage(chatId int64) (*mongo.ChangeStream, error) {
	return messageCollection.Watch(context.Background(),
		mongo.Pipeline{bson.D{{mongoDbMatch, bson.D{{mongoDbAnd, bson.A{bson.D{{"operationType", "insert"}},
			bson.D{{filterMessage, chatId}}}}}}}},
		options.ChangeStream().SetFullDocument(options.UpdateLookup))
}

func insertDocumentOne(ctx context.Context, obj interface{}, collection *mongo.Collection) error {
	_, err := collection.InsertOne(ctx, obj, nil)
	if err != nil {
		return err
	}

	return nil
}

func updateDocumentOne(ctx context.Context, filter, update bson.D, collection *mongo.Collection) error {
	result, err := collection.UpdateOne(ctx, filter, update, nil)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		err = mongo.ErrNoDocuments
		return err
	}

	return nil
}

func findDocumentOne(ctx context.Context, filter bson.D, collection *mongo.Collection, container interface{}) error {
	result := collection.FindOne(ctx, filter, nil)
	if result.Err() != nil {
		return result.Err()
	}

	if err := result.Decode(container); err != nil {
		return err
	}

	return nil
}

func addChat(sessCtx mongo.SessionContext, chatId int64) error {
	return insertDocumentOne(sessCtx, entities.NewChat(chatId, 0), queueCollection)
}

func getFriendOpHook(op string, userId1, userId2 int64) func(sessCtx mongo.SessionContext) error {
	return func(sessCtx mongo.SessionContext) error {
		// BUGFIX 不能正确删除会话
		id := tools.GenerateSnowflakeId(false)
		privateChat := entities.NewPrivateChatInfo(id, userId1, userId2)

		if err := addOrDeleteFriend(op, id, userId1, userId2, sessCtx); err != nil {
			return err
		}

		if err := addOrDeleteFriend(op, id, userId2, userId1, sessCtx); err != nil {
			return err
		}

		// 同样值得商榷，是否需要保留聊天关系以查证聊天记录
		if op == mongoDbPush {
			if err := addChat(sessCtx, id); err != nil {
				return nil
			}
			return insertDocumentOne(sessCtx, privateChat, chatInfoCollection)
		}

		return nil
	}
}

func addOrDeleteFriend(op string, chatId, masterId, friendId int64, ctx context.Context) error {
	filter := getBson(UserId, masterId)
	update := bson.D{getOpBsonEWithMultiKV(op, []string{UserFriendShip, UserPrivateChat}, []interface{}{friendId, chatId})}
	if err := updateDocumentOne(ctx, filter, update, userInfoCollection); err != nil {
		return err
	}

	return nil
}

func joinOrQuitGroup(op string, userId, groupId int64) error {
	sess, err := mongoConnection.StartSession()
	if err != nil {
		return err
	}

	userInfoFilter := getBson(UserId, userId)
	groupInfoFilter := getBson(GroupId, groupId)

	userInfoUpdate := getOpBson(op, UserGroup, groupId)
	groupInfoUpdate := getOpBson(op, GroupMember, userId)

	hook := func(sessCtx mongo.SessionContext) error {
		if err := updateDocumentOne(sessCtx, userInfoFilter, userInfoUpdate, userInfoCollection); err != nil {
			return err
		}

		if err := updateDocumentOne(sessCtx, groupInfoFilter, groupInfoUpdate, chatInfoCollection); err != nil {
			return err
		}

		return err
	}

	if err = execTransaction(sess, hook); err != nil {
		return err
	}

	return nil
}

func authOrDeAuthAdmin(op string, userId, groupId int64) error {
	filter := getBson(mongoDbAnd, bson.A{bson.D{{GroupId, groupId}}, bson.D{{GroupMember, getBson(mongoDbIn, []int64{userId})}}})
	update := getOpBson(op, GroupAdministrator, userId)
	if err := updateDocumentOne(context.Background(), filter, update, chatInfoCollection); err != nil {
		return err
	}

	return nil
}

func execTransaction(sess mongo.Session, hook func(sessCtx mongo.SessionContext) error) error {
	sessCtx := mongo.NewSessionContext(context.Background(), sess)
	if err := sess.StartTransaction(); err != nil {
		return err
	}

	if err := hook(sessCtx); err != nil {
		return errorWhenTransactionGoing(sess, sessCtx, err)
	}

	if err := sess.CommitTransaction(sessCtx); err != nil {
		return errorWhenTransactionGoing(sess, sessCtx, err)
	}

	return nil
}

func execTransactionWithRetry(sess mongo.Session, hook func(sessCtx mongo.SessionContext) error) error {
	err := execTransaction(sess, hook)
	for i := 1; i < 3 && err != nil; i++ {
		err = execTransaction(sess, hook)
	}
	if err != nil {
		return err
	}
	return nil
}

func errorWhenTransactionGoing(sess mongo.Session, sessCtx mongo.SessionContext, err error) error {
	if errAbort := sess.AbortTransaction(sessCtx); errAbort != nil {
		return fmt.Errorf("2 errors in transaction.\n1: %s\n2: %s\n", err.Error(), errAbort.Error())
	}
	return err
}

func getDefaultMongoConcern() *options.ClientOptions {
	opt := options.ClientOptions{}
	opt.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	opt.SetReadConcern(readconcern.Majority())
	opt.SetReadPreference(readpref.SecondaryPreferred())
	return &opt
}

func initConnection(cfg *config.MongoDBConfig) {
	db := mongoConnection.Database(cfg.Database)
	messageCollection = db.Collection(cfg.MessageCollection)
	queueCollection = db.Collection(cfg.QueueCollection)
	userInfoCollection = db.Collection(cfg.UserInfoCollection)
	chatInfoCollection = db.Collection(cfg.GroupInfoCollection)

}

func createCollectionsAndIndexes(cfg *config.MongoDBConfig) error {
	db := mongoConnection.Database(cfg.Database)

	if err := createCollectionAndIndexOne(db, cfg.MessageCollection, MessageReceiver, MessageId); err != nil {
		return err
	}

	if err := createCollectionAndIndexOne(db, cfg.QueueCollection, ChatId); err != nil {
		return err
	}

	if err := createCollectionAndIndexOne(db, cfg.UserInfoCollection, UserId); err != nil {
		return err
	}

	if err := createCollectionAndIndexOne(db, cfg.GroupInfoCollection, GroupId); err != nil {
		return err
	}

	return nil
}

func createCollectionAndIndexOne(db *mongo.Database, collection string, key ...string) error {
	if err := db.CreateCollection(context.Background(), collection); err != nil {
		return err
	}
	if _, err := db.Collection(collection).Indexes().CreateOne(context.Background(), generateKeysIndex(key...)); err != nil {
		return err
	}

	return nil
}

func dropDatabase(cfg *config.MongoDBConfig) error {
	if err := mongoConnection.Database(cfg.Database).Drop(context.Background()); err != nil {
		return err
	}

	return nil
}

func generateKeysIndex(keys ...string) mongo.IndexModel {
	keyDoc := bson.D{}
	for _, key := range keys {
		keyDoc = append(keyDoc, bson.E{Key: key, Value: 1})
	}

	return mongo.IndexModel{
		Keys:    keyDoc,
		Options: (&options.IndexOptions{}).SetUnique(true),
	}
}

func decodeDataInCursor(cursor *mongo.Cursor, slicePtr interface{}) error {
	return cursor.All(context.Background(), slicePtr)
}

func getBson(key string, val interface{}) bson.D {
	return bson.D{{key, val}}
}

func getOpBson(op, key string, val interface{}) bson.D {
	return bson.D{{op, bson.D{{key, val}}}}
}

func getOpBsonEWithMultiKV(op string, key []string, val []interface{}) bson.E {
	valSlice := make(bson.D, 0, len(key))
	for i, k := range key {
		valSlice = append(valSlice, bson.E{Key: k, Value: val[i]})
	}
	return bson.E{Key: op, Value: valSlice}
}
