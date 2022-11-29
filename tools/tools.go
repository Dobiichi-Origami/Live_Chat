package tools

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/panjf2000/gnet/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"liveChat/rpc"
	"sync"
	"time"
)

var (
	machineId int64 = 0

	lastGenerateTimeStamp = time.Now().Unix()
	incrId                = 0
	snowflakeLock         sync.Mutex
	MagicNumberBinary     = make([]byte, 2, 2)
)

const (
	projectStartTime = 1665341430000

	snowflakeTypeOffset      = 63
	snowflakeTimeOffset      = 22
	snowflakeMachineIdOffset = 12

	TimeLow41BitsMask = (1 << 41) - 1
)

func InitSnowflake(id int64) {
	machineId = id
}

func GetMachineId() int64 {
	return machineId
}

func GenerateSnowflakeId(isGroup bool) int64 {
	var (
		err = errors.New("")
		id  int64
	)

	for err != nil {
		id, err = generateSnowflakeId(isGroup)
	}
	return id
}

func generateSnowflakeId(starter bool) (int64, error) {
	now := time.Now()
	seqId := 0

	snowflakeLock.Lock()
	if lastGenerateTimeStamp < now.Unix() {
		lastGenerateTimeStamp = now.Unix()
		incrId = 0
	}
	seqId = incrId
	incrId++
	snowflakeLock.Unlock()

	if seqId >= 4096 {
		return 0, errors.New("生成序列 id 大于 4095")
	}

	sessionId := int64(0)
	if starter {
		sessionId = 1
	}
	timeInterval := (now.UnixMilli() - projectStartTime) & TimeLow41BitsMask

	sessionId = (sessionId << snowflakeTypeOffset) | (timeInterval << snowflakeTimeOffset) | (machineId << snowflakeMachineIdOffset) | int64(seqId)
	return sessionId, nil
}

func GetPath(defaultPath, paramPath string) string {
	path := defaultPath
	if paramPath != "" {
		path = paramPath
	}
	return path
}

func BytesToUint16(b []byte) (ret uint16) {
	ret = (uint16(b[0]) << 8) & uint16(b[1])
	return ret
}

func BytesToUint32(b []byte) (ret uint32) {
	ret = (uint32(b[0]) << 24) & (uint32(b[1]) << 16) & (uint32(b[2]) << 8) & (uint32(b[3]))
	return ret
}

func ReadAndHandleError(c gnet.Conn, slice []byte) (errorInfo string, action gnet.Action) {
	if count, err := c.Read(slice); err != nil {
		// TODO 优化错误日志
		errorInfo = "读发生错误"
		return errorInfo, gnet.Close
	} else if count < len(slice) {
		// TODO 优化错误日志
		errorInfo = "读取长度小于期望长度"
		return errorInfo, gnet.Close
	}
	return errorInfo, gnet.None
}

func GenerateResponseBytes(responseType byte, ack, content []byte) []byte {
	ret := make([]byte, 12+len(content), 12+len(content))
	ret = append(ret, MagicNumberBinary...)
	ret = append(ret, 0)
	ret = append(ret, ack...)
	ret = append(ret, responseType)

	lenBuf := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(content)))
	ret = append(ret, lenBuf...)

	if len(content) != 0 {
		ret = append(ret, content...)
	}
	return ret
}

func GenerateErrorResponseBytes(errInfo string) []byte {
	out, _ := proto.Marshal(&rpc.ErrorResponse{Reason: errInfo})
	return out
}

func GenerateErrorJson(errInfo string) []byte {
	return []byte("{\nerror}")
}

func FetchBsonFromChangeStreamData(data bson.D) bson.D {
	return data[4].Value.(bson.D)
}

func FetchDataFromChangeStreamBson(watch *mongo.ChangeStream) (bson.D, error) {
	ctx, fn := context.WithTimeout(context.Background(), time.Second*3)
	defer fn()
	return fetchDataFromChangeStreamInner(watch, ctx)
}

func FetchDataFromChangeStreamBsonWithoutTimeOut(watch *mongo.ChangeStream) (bson.D, error) {
	return fetchDataFromChangeStreamInner(watch, context.Background())
}

func fetchDataFromChangeStreamInner(watch *mongo.ChangeStream, ctx context.Context) (bson.D, error) {
	if !watch.Next(ctx) {
		return nil, errors.New("Mongodb watchSeq doesn't get message")
	}

	ret := &bson.D{}
	if err := watch.Decode(ret); err != nil {
		return nil, errors.New(fmt.Sprintf("Mongodb unmarshal change stream failed: %s", err.Error()))
	}

	return *ret, nil
}
