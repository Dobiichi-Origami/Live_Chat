package controllers

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"liveChat/constants"
	"liveChat/db"
	"liveChat/entities"
	"liveChat/log"
	"liveChat/pool"
	"liveChat/protocol"
	"liveChat/tools"
)

var workerPool *pool.WorkerPool

func init() {
	workerPool, _ = pool.NewWorkerPool(requestAsyncHandler)
}

func PushTask(arg interface{}) {
	workerPool.PushTask(arg)
}

func requestAsyncHandler(arg interface{}) {
	task := arg.(*pool.TCPRequestPackage)

	var retSlice []byte
	retType := constants.SuccessResponseLoad
	closeConnFlag := false

	defer func() {
		ci := GetConnection(task.Fd)
		if ci == nil {
			log.Error("链接已删除")
		} else if err := ci.conn.AsyncWrite(tools.GenerateResponseBytes(retType, task.Ack, retSlice), nil); err != nil {
			log.Error(fmt.Sprintf("异步处理中回包失败: %s", err.Error()))
		}
		if ci != nil && closeConnFlag {
			ci.conn.Close()
		}
		pool.PutRequestPackage(task)
	}()

	errorHandleHook := func(errInfo string, closeConn bool) {
		log.Error(errInfo)
		retSlice = tools.GenerateErrorResponseBytes(errInfo)
		retType = constants.ErrorResponseLoad
		if closeConn {
			closeConnFlag = true
		}
	}

	switch task.RequestType {
	case constants.ErrorResponseLoad:
	case constants.SuccessResponseLoad:
	case constants.MessageLoad:
		message := protocol.Message{}
		if err := proto.Unmarshal(task.Load, &message); err != nil {
			errorHandleHook(fmt.Sprintf("反序列化错误：%s", err.Error()), false)
			return
		}

		if err := db.AddMessage(&message); err != nil {
			errorHandleHook(fmt.Sprintf("消息入库失败: %s", err.Error()), false)
			return
		}

	case constants.RequestMessageLoad:
		request := protocol.RequestMessage{}
		if err := proto.Unmarshal(task.Load, &request); err != nil {
			errorHandleHook(fmt.Sprintf("反序列化错误：%s", err.Error()), false)
			return
		}

		message, err := db.FetchMessageCache(request.Receiver, request.Id)
		if err != nil || message == nil {
			message, err = db.GetMessageInSeq(request.Receiver, request.Id)
		}

		if err != nil {
			errorHandleHook(fmt.Sprintf("从数据库中获取消息失败: %s", err.Error()), false)
			return
		}

		ret, err := proto.Marshal(entities.TransferMessageToProtoBuf(message))
		if err != nil {
			errorHandleHook(fmt.Sprintf("返回消息体反序列化错误: %s", err.Error()), false)
			return
		}
		retSlice = ret
		retType = constants.MessageLoad

	case constants.RequestMultiMessageLoad:
		request := protocol.RequestMultiMessage{}
		if err := proto.Unmarshal(task.Load, &request); err != nil {
			errorHandleHook(fmt.Sprintf("反序列化错误：%s", err.Error()), false)
			return
		}

		messages, err := db.GetMessageInSeqRange(request.Receiver, request.BottomId, request.TopId)
		if err != nil {
			errorHandleHook(fmt.Sprintf("从数据库中批量获取消息失败: %s", err.Error()), false)
			return
		}

		protoMessageSlice := make([]*protocol.Message, len(messages), len(messages))
		for i := 0; i < len(messages); i++ {
			protoMessageSlice[i] = entities.TransferMessageToProtoBuf(&messages[i])
		}

		multiMessage := protocol.MultiMessage{Messages: protoMessageSlice}
		ret, err := proto.Marshal(&multiMessage)
		if err != nil {
			errorHandleHook(fmt.Sprintf("返回消息体反序列化错误: %s", err.Error()), false)
			return
		}
		retSlice = ret
		retType = constants.MultiMessageLoad

	case constants.RequestEstablishConnectionLoad:
		tokenMessage := protocol.RequestEstablishConnection{}
		if err := proto.Unmarshal(task.Load, &tokenMessage); err != nil {
			errorHandleHook(fmt.Sprintf("反序列化错误：%s", err.Error()), true)
			return
		}

		userId := GetUserIdByToken(tokenMessage.Token)
		if userId == -1 {
			errorHandleHook(fmt.Sprintf("token 对应用户 id 不存在. token: %s", tokenMessage.Token), true)
			return
		}
		userInfo, err := db.SearchUserInfo(userId)
		if err != nil {
			errorHandleHook(fmt.Sprintf("服务器内部错误，无法获取用户信息: %s", err.Error()), true)
			return
		}

		resp := protocol.ResponseEstablishConnection{}
		resp.PrivateChat = Subscribe(userInfo.PrivateChat, userId)
		resp.GroupChat = Subscribe(userInfo.Group, userId)

		ret, err := proto.Marshal(&resp)
		if err != nil {
			errorHandleHook(fmt.Sprintf("返回消息体反序列化错误: %s", err.Error()), false)
			return
		}
		SetTokenForConnection(tokenMessage.Token, task.Fd)
		retSlice = ret
		retType = constants.ResponseEstablishConnectionLoad

	case constants.HeartBeatLoad:
		retType = constants.HeartBeatResponse

	default:
		errInfo := fmt.Sprintf("未识别的消息类型 %d", task.RequestType)
		log.Error(errInfo)
		retSlice = tools.GenerateErrorResponseBytes(errInfo)
		retType = constants.ErrorResponseLoad
	}
}
