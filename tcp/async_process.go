package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/panjf2000/gnet/v2"
	"liveChat/constants"
	"liveChat/controllers"
	"liveChat/db"
	"liveChat/entities"
	"liveChat/log"
	"liveChat/pool"
	"liveChat/rpc"
	"liveChat/tools"
	"time"
)

var workerPool *pool.WorkerPool

func init() {
	workerPool, _ = pool.NewWorkerPool(requestAsyncHandler)
}

func PushTask(arg interface{}) {
	workerPool.PushTask(arg)
}

func requestAsyncHandler(arg interface{}) {
	var (
		task          = arg.(pool.TCPRequestPackage)
		ctx           = task.Conn.Context().(pool.TCPContext)
		err           error
		retSlice      []byte
		retType       = constants.SuccessResponseLoad
		closeConnFlag = false
	)

	if task.RequestType != constants.RequestEstablishConnectionLoad && ctx.Token == "" {
		err = errors.New("非法请求")
		closeConnFlag = true
		return
	}

	defer func() {
		if err != nil {
			log.Error(err.Error())
			retSlice = errorHandleHook(err.Error())
			retType = constants.ErrorResponseLoad
		}

		if err = task.Conn.AsyncWrite(tools.GenerateResponseBytes(retType, task.Ack, retSlice), nil); err != nil {
			log.Error(fmt.Sprintf("异步处理中回包失败: %s", err.Error()))
		}

		if closeConnFlag {
			task.Conn.Close()
		}
		pool.PutRequestPackage(task)
	}()

	switch task.RequestType {
	case constants.ErrorResponseLoad:
	case constants.SuccessResponseLoad:
	case constants.MessageLoad:
		message := rpc.Message{}
		if err = proto.Unmarshal(task.Load, &message); err != nil {
			err = errors.New(fmt.Sprintf("反序列化错误：%s", err.Error()))
			return
		}

		if message.Sender != ctx.UserId {
			err = errors.New("非法消息：用户 id 不一致")
			return
		}

		if err = checkAuthForRelationships(ctx.UserId, message.Receiver); err != nil {
			return
		}

		if err = db.AddMessage(context.Background(), &message); err != nil {
			err = errors.New(fmt.Sprintf("消息入库失败: %s", err.Error()))
			return
		}

		SendMessage(&message)
		err = db.CacheMessageWithTimeOut(entities.NewMessageFromProtobufWithSeq(&message))
		if err != nil {
			log.Error(fmt.Sprintf("消息缓存 Redis 失败: %s", err.Error()))
			err = nil
		}

		retType = constants.SuccessResponseLoad

	case constants.RequestMessageLoad:
		request := rpc.RequestMessage{}
		if err = proto.Unmarshal(task.Load, &request); err != nil {
			err = errors.New(fmt.Sprintf("反序列化错误：%s", err.Error()))
			return
		}

		if err = checkAuthForRelationships(ctx.UserId, request.Receiver); err != nil {
			return
		}

		var message *entities.Message
		message, err = db.FetchMessageCache(request.Receiver, request.Id)
		if err != nil || message == nil {
			message, err = db.GetMessageInSeq(context.Background(), request.Receiver, request.Id)
		}

		if err != nil {
			err = errors.New(fmt.Sprintf("从数据库中获取消息失败: %s", err.Error()))
			return
		}

		retSlice, err = proto.Marshal(entities.TransferMessageToProtoBuf(message))
		if err != nil {
			err = errors.New(fmt.Sprintf("返回消息体反序列化错误: %s", err.Error()))
			return
		}
		retType = constants.MessageLoad

	case constants.RequestMultiMessageLoad:
		request := rpc.RequestMultiMessage{}
		if err = proto.Unmarshal(task.Load, &request); err != nil {
			err = errors.New(fmt.Sprintf("反序列化错误：%s", err.Error()))
			return
		}

		if err = checkAuthForRelationships(ctx.UserId, request.Receiver); err != nil {
			return
		}

		var messages []entities.Message
		messages, err = db.GetMessageInSeqRange(context.Background(), request.Receiver, request.BottomId, request.TopId)
		if err != nil {
			err = errors.New(fmt.Sprintf("从数据库中批量获取消息失败: %s", err.Error()))
			return
		}

		protoMessageSlice := make([]*rpc.Message, len(messages), len(messages))
		for i := 0; i < len(messages); i++ {
			protoMessageSlice[i] = entities.TransferMessageToProtoBuf(&messages[i])
		}

		multiMessage := rpc.MultiMessage{Messages: protoMessageSlice}

		retSlice, err = proto.Marshal(&multiMessage)
		if err != nil {
			err = errors.New(fmt.Sprintf("返回消息体反序列化错误: %s", err.Error()))
			return
		}
		retType = constants.MultiMessageLoad

	case constants.RequestEstablishConnectionLoad:
		establishMessage := rpc.RequestEstablishConnection{}
		if err = proto.Unmarshal(task.Load, &establishMessage); err != nil {
			err = errors.New(fmt.Sprintf("反序列化错误：%s", err.Error()))
			return
		}

		var userId int64
		userId, err = controllers.GetUserIdByToken(establishMessage.Token)
		if err != nil {
			err = errors.New(fmt.Sprintf("服务器内部错误，无法获取用户信息: %s", err.Error()))
			return
		} else if userId == -1 {
			err = errors.New(fmt.Sprintf("token 对应用户 id 不存在. token: %s", establishMessage.Token))
			return
		}

		var (
			platform = int(establishMessage.Platform)
			token    = establishMessage.Token
		)

		if !addConnection(task.Conn, userId, platform) {
			err = errors.New("注册链接失败，请重试")
			return
		}

		ctx.Token = token
		ctx.Platform = platform
		ctx.UserId = userId

		retType = constants.SuccessResponseLoad

	case constants.HeartBeatLoad:
		retType = constants.HeartBeatResponse

	default:
		err = errors.New(fmt.Sprintf("未识别的消息类型 %d", task.RequestType))
	}
}

func checkAuthForRelationships(sender, receiver int64) error {
	if receiver < 0 {
		flag, err := controllers.CheckIsUserInGroup(sender, receiver, false)
		if err != nil {
			return errors.New(fmt.Sprintf("无法鉴别用户信息: %s", err.Error()))
		} else if !flag {
			return errors.New("用户不在群组中")
		}
	} else {
		flag, err := controllers.CheckAreUsersFriend(sender, receiver, false)
		if err != nil {
			return errors.New(fmt.Sprintf("无法鉴别用户信息: %s", err.Error()))
		} else if !flag {
			return errors.New("与目标用户不是好友关系")
		}
	}

	return nil
}

func errorHandleHook(errInfo string) (retSlice []byte) {
	log.Error(errInfo)
	retSlice = tools.GenerateErrorResponseBytes(errInfo)
	return
}

func addConnection(conn gnet.Conn, userId int64, platform int) bool {
	for _, c := range controllers.GetAllServerClients() {
		ctx, cfn := context.WithTimeout(context.Background(), time.Second*3)
		_, err := c.KickUserOffOnSpecificPlatform(ctx, &rpc.KickOffRequest{
			RequestId: 0,
			UserId:    userId,
			Platform:  rpc.KickOffRequest_PlatformType(platform),
		}, nil)
		if err != nil {
			log.Error(err.Error())
		}
		cfn()
	}

	flag := false
	for i := 0; i < 3; i++ {
		flag = controllers.AddConnection(conn, userId, platform)
		if flag {
			break

		}
	}

	return flag
}
