package rpc_implementation

import (
	"context"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"liveChat/constants"
	"liveChat/controllers"
	"liveChat/log"
	"liveChat/rpc"
	"liveChat/tools"
	"net"
)

var grpcServer *grpc.Server

func InitRpcServer(listenAddress string) {
	acp, err := net.Listen("tpc", listenAddress)
	if err != nil {
		panic(err)
	}

	grpcServer = grpc.NewServer()
	rpc.RegisterServerNodeServer(grpcServer, &RpcServer{})
	if err = grpcServer.Serve(acp); err != nil {
		panic(err)
	}
}

type RpcServer struct {
	rpc.UnimplementedServerNodeServer
}

func (s RpcServer) KickUserOffOnSpecificPlatform(ctx context.Context, request *rpc.KickOffRequest) (*rpc.Response, error) {
	ok := controllers.DeleteConnection(request.UserId, int(request.Platform))
	return generateRpcResponse(request.RequestId, ok, true, ""), nil
}

func (s RpcServer) BroadcastNotification(ctx context.Context, request *rpc.NotificationRequest) (*rpc.Response, error) {
	data, err := proto.Marshal(request)
	if err != nil {
		return generateRpcResponse(request.RequestId, false, false, err.Error()), nil
	}

	userList := make([]int64, 0)
	if request.Receiver > 0 {
		userList = append(userList, request.Receiver)
	} else if userList, err = getUserListInGroup(request.Receiver, true); err != nil {
		return generateRpcResponse(request.RequestId, false, false, err.Error()), nil
	}

	if sendToUser(constants.NotificationRequestLoad, userList, data) == 0 {
		return generateRpcResponse(request.RequestId, false, true, ""), nil
	}
	return generateRpcResponse(request.RequestId, true, true, ""), nil
}

func (s RpcServer) BroadcastMessage(ctx context.Context, request *rpc.MessageRequest) (*rpc.Response, error) {
	data, err := proto.Marshal(request)
	if err != nil {
		return generateRpcResponse(request.RequestId, false, false, err.Error()), nil
	}

	userList := make([]int64, 0)
	if request.Message.Receiver > 0 {
		userList = append(userList, request.Message.Receiver)
	} else if userList, err = getUserListInGroup(request.Message.Receiver, false); err != nil {
		return generateRpcResponse(request.RequestId, false, false, err.Error()), nil
	}

	if sendToUser(constants.MessageLoad, userList, data) == 0 {
		return generateRpcResponse(request.RequestId, false, true, ""), nil
	}
	return generateRpcResponse(request.RequestId, true, true, ""), nil
}

func generateRpcResponse(requestId uint64, isProcessed, isSucceeded bool, failureReason string) *rpc.Response {
	return &rpc.Response{
		RequestId:            requestId,
		IsProcessedByOneSelf: isProcessed,
		IsSucceeded:          isSucceeded,
		FailureReason:        failureReason,
	}
}

func sendToUser(msgType byte, userList []int64, data []byte) (counter int) {
	data = tools.GenerateResponseBytes(msgType, []byte{0, 0, 0, 0}, data)
	for _, userId := range userList {
		conns := controllers.GetConnection(userId)
		if conns == nil {
			continue
		}

		counter++
		for _, conn := range conns {
			err := conn.AsyncWrite(data, nil)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
	return
}

func getUserListInGroup(groupId int64, onlyAdministrator bool) (userList []int64, err error) {
	info, err := controllers.GetUserListInGroup(groupId, false)
	if err != nil {
		return nil, err
	}

	for _, member := range info {
		if member.IsDeleted {
			continue
		}

		if onlyAdministrator {
			if member.IsAdministrator {
				userList = append(userList, member.MemberId)
			}
		} else {
			userList = append(userList, member.MemberId)
		}
	}

	return userList, nil
}
