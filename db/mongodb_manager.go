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
