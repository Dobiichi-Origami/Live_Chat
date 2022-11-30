package db

import (
	"context"
	"errors"
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
	"liveChat/rpc"
	"time"
)

const defaultMongoDBConfigPath = "./mongodb_config.json"

var MongoDBConfigPath = defaultMongoDBConfigPath

var (
	mongoDbDatabaseName string

	mongoConnection *mongo.Client

	// 存放所有消息的集合对象
	messageCollection *mongo.Collection
	// 存放所有会话最新消息序号的集合对象
	queueCollection *mongo.Collection
	// 存放所有需要处理的通知的集合对象
	notificationCollection *mongo.Collection
	// 存放某个对象最新通知序号的集合对象
	notiSeqCollection *mongo.Collection

	isMongodbInitiated bool = false
)

const (
	mongoQueueCollectionName   = "queue"
	mongoMessageCollectionName = "message"
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
	ChatId       = "id"
	ChatSequence = "sequence"

	MessageId       = "id"
	MessageReceiver = "receiver"

	NotificationId           = "receiver_id"
	NotificationSequence     = "sequence"
	NotificationHandleUserId = "handle_user_id"
	NotificationIsHandled    = "is_handled"
	NotificationIsAgree      = "is_agree"
)

const (
	filterChatSeq = "fullDocument." + ChatId
	filterMessage = "fullDocument." + MessageReceiver
)

const chatTypeMask = 1 << 63

var (
	MongoErrorNoNotification = errors.New("无匹配通知")
)

func InitMongoDBConnection(url, databaseName string) {
	if isMongodbInitiated {
		return
	}

	mongoDbDatabaseName = databaseName

	var err error
	mongoConnection, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(url), getDefaultMongoConcern())
	if err != nil {
		panic(err)
	}

	err = mongoConnection.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	if err = createCollectionsAndIndexes(); err != nil {
		panic(err)
	}

	initConnection()
	isMongodbInitiated = true
	return
}

func GetChatSeqInSlice(ctx context.Context, chatId []int64) ([]entities.Chat, error) {
	filter := bson.D{{ChatId, bson.D{{mongoDbIn, chatId}}}}
	cursor, err := queueCollection.Find(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	seqSlice := make([]entities.Chat, 0)
	if err = decodeDataInCursor(cursor, &seqSlice); err != nil {
		return nil, err
	}
	return seqSlice, nil
}

func GetChatSequence(ctx context.Context, chatId int64) (uint64, error) {
	chat := entities.NewEmptyChat()
	if err := findDocumentOne(ctx, getBson(ChatId, chatId), queueCollection, chat); err != nil {
		return constants.MaxUInt64, err
	}

	return chat.Sequence, nil
}

func GetAndAddChatSequence(ctx context.Context, chatId int64) (uint64, error) {
	chat := entities.NewEmptyChat()
	result := queueCollection.FindOneAndUpdate(
		ctx,
		getBson(ChatId, chatId),
		bson.D{{mongoDbIncr, bson.D{{ChatSequence, 1}}}, {mongoDbSetInInsert, bson.D{{ChatId, chatId}}}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	if result.Err() != nil {
		return constants.MaxUInt64, result.Err()
	}

	if err := result.Decode(&chat); err != nil {
		return constants.MaxUInt64, err
	}

	return chat.Sequence, nil
}

func GetMessageInSeqRange(ctx context.Context, chatId int64, bottom, top uint64) ([]entities.Message, error) {
	cursor, err := messageCollection.Find(ctx,
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

func GetMessageInSeq(ctx context.Context, chatId int64, seq uint64) (*entities.Message, error) {
	message := entities.NewEmptyMessage()
	if err := findDocumentOne(ctx, bson.D{{MessageReceiver, chatId}, {MessageId, seq}}, messageCollection, message); err != nil {
		return nil, err
	}
	return message, nil
}

func CreateMessageWithSeq(ctx context.Context, m *rpc.Message) (*entities.Message, error) {
	seq, err := GetAndAddChatSequence(ctx, m.Receiver)
	if err != nil {
		return nil, err
	}

	message := entities.NewMessageFromProtobufWithoutSeq(m)
	message.Id = seq
	return message, nil
}

func AddMessage(ctx context.Context, m *rpc.Message) error {
	message, err := CreateMessageWithSeq(ctx, m)
	if err != nil {
		return err
	}
	if err = insertDocumentOne(ctx, message, messageCollection); err != nil {
		return err
	}
	return nil
}

func GetNotificationSequence(ctx context.Context, receiverId int64) (uint64, error) {
	noti := &entities.Notification{}
	if err := findDocumentOne(ctx, getBson(NotificationId, receiverId), notiSeqCollection, noti); err != nil {
		return constants.MaxUInt64, err
	}

	return noti.Seq, nil
}

func GetNotificationInSeqRange(ctx context.Context, receiverId int64, bottom, top uint64) ([]entities.Notification, error) {
	cursor, err := notificationCollection.Find(ctx,
		bson.D{{NotificationId, receiverId}, {NotificationSequence, bson.D{{mongoDbGreaterEqual, bottom}, {mongoDbLessEqual, top}}}},
		nil,
	)

	if err != nil {
		return nil, err
	}

	notiSlice := make([]entities.Notification, 0)
	if err = decodeDataInCursor(cursor, &notiSlice); err != nil {
		return nil, err
	}
	return notiSlice, nil
}

func GetNotificationInSeq(ctx context.Context, receiverId int64, seq uint64) (*entities.Notification, error) {
	noti := &entities.Notification{}
	if err := findDocumentOne(ctx, bson.D{{NotificationId, receiverId}, {NotificationSequence, seq}}, notificationCollection, noti); err != nil {
		return nil, err
	}
	return noti, nil
}

func HandleNotification(ctx context.Context, receiverId, handleUserId int64, seq uint64, isAgree bool) (*entities.Notification, error) {
	result := notificationCollection.FindOneAndUpdate(
		ctx,
		bson.D{{mongoDbAnd, bson.A{bson.D{{NotificationId, receiverId}}, bson.D{{NotificationSequence, seq}}}}},
		bson.D{{NotificationIsHandled, true}, {NotificationIsAgree, isAgree}, {NotificationHandleUserId, handleUserId}},
		nil,
	)

	noti := &entities.Notification{}
	if err := result.Decode(noti); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, MongoErrorNoNotification
		}
		return nil, err
	}

	noti.HandleUserId = handleUserId
	noti.IsHandled = true
	noti.IsAgree = isAgree
	return noti, nil
}

func getAndAddNotificationSequence(ctx context.Context, receiverId int64) (uint64, error) {
	noti := entities.Notification{}
	result := notiSeqCollection.FindOneAndUpdate(
		ctx,
		getBson(NotificationId, receiverId),
		bson.D{{mongoDbIncr, bson.D{{NotificationSequence, 1}}}, {mongoDbSetInInsert, bson.D{{NotificationId, receiverId}}}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	if result.Err() != nil {
		return constants.MaxUInt64, result.Err()
	}

	if err := result.Decode(&noti); err != nil {
		return constants.MaxUInt64, err
	}

	return noti.Seq, nil
}

func AddAndReturnNotification(ctx context.Context, n *entities.Notification) (*entities.Notification, error) {
	sequence, err := getAndAddNotificationSequence(ctx, n.ReceiverId)
	if err != nil {
		return nil, err
	}
	n.Seq = sequence
	n.Timestamp = time.Now().Unix()

	if err = insertDocumentOne(ctx, n, notificationCollection); err != nil {
		return nil, err
	}
	return n, nil
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

func addChat(ctx context.Context, chatId int64) error {
	_, err := queueCollection.UpdateOne(ctx, getBson("id", chatId), getOpBson(mongoDbSetInInsert, "sequence", 0), (&options.UpdateOptions{}).SetUpsert(true))
	return err
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

func initConnection() {
	db := mongoConnection.Database(mongoDbDatabaseName)
	messageCollection = db.Collection(mongoMessageCollectionName)
	queueCollection = db.Collection(mongoQueueCollectionName)

}

func createCollectionsAndIndexes() error {
	db := mongoConnection.Database(mongoDbDatabaseName)
	lists, err := db.ListCollectionNames(context.Background(), bson.D{}, nil)
	if err != nil {
		return err
	}

	flag1, flag2 := false, false
	for _, name := range lists {
		if name == mongoMessageCollectionName {
			flag1 = true
		} else if name == mongoQueueCollectionName {
			flag2 = true
		}
	}

	if !flag1 {
		if err = createCollectionAndIndexOne(db, mongoMessageCollectionName, MessageReceiver, MessageId); err != nil {
			return err
		}
	}

	if !flag2 {
		if err = createCollectionAndIndexOne(db, mongoQueueCollectionName, ChatId); err != nil {
			return err
		}
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
