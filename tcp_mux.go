package liveChat

import (
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"liveChat/constants"
	"liveChat/controllers"
	"liveChat/db"
	"liveChat/log"
	"liveChat/pool"
	"liveChat/tools"
	"time"
)

type engineImplementation struct {
	gnet.BuiltinEventEngine
}

func (engine *engineImplementation) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	closeChan := make(chan struct{}, 1)
	controllers.AddConnection(c, closeChan)
	controllers.NewTimer(c.Fd())
	go func() {
		ticker := time.NewTicker(constants.HeartBeatMaxInterval)
		for {
			select {
			case <-closeChan:
				return
			case <-ticker.C:
				if !controllers.CheckTimer(c.Fd()) {
					c.Close()
				}
			}
		}
	}()

	if err := c.SetKeepAlivePeriod(constants.KeepAlivePeriod); err != nil {
		log.Error(err.Error())
	}
	return nil, gnet.None
}

func (engine *engineImplementation) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	connection := controllers.GetConnection(c.Fd())
	if connection == nil || !connection.CloseLock() {
		return gnet.Close
	}
	controllers.DeleteConnection(c)
	connection.CloseConnection()

	if token := connection.GetToken(); token != "" {
		userId := controllers.GetUserIdByToken(token)
		if info, err := db.SearchUserInfo(userId); err != nil {
			log.Error(fmt.Sprintf("尝试删除消息监听失败: %s", err.Error()))
		} else {
			for _, friendship := range info.Friendships {
				controllers.DeleteSubscribe(friendship.ChatId, userId)
			}
			for _, group := range info.Groups {
				controllers.DeleteSubscribe(group.GroupId, userId)
			}
		}
	}

	controllers.DeleteTimer(c.Fd())
	if err != nil {
		log.Error(fmt.Sprintf("链接意外关闭: %s", err.Error()))
	}
	return gnet.Close
}

func (engine *engineImplementation) OnTraffic(c gnet.Conn) (action gnet.Action) {
	controllers.ResetTimer(c.Fd())
	request := pool.GetRequestPackage()
	if err := request.SetPackage(c); err != nil {
		pool.PutRequestPackage(request)
		errInfo := fmt.Sprintf("收到客户端非法请求: %s", err.Error())
		log.Error(errInfo)
		err := c.AsyncWrite(tools.GenerateErrorResponseBytes(errInfo), nil)
		if err != nil {
			log.Error(fmt.Sprintf("错误处理中异步发送回包失败: %s", err.Error()))
		}
	} else {
		controllers.PushTask(request)
	}
	return gnet.None
}

func InitiateTcpServer() error {
	engine := new(engineImplementation)
	if err := gnet.Run(engine, "tcp://0.0.0.0:1234", gnet.WithMulticore(true)); err != nil {
		return err
	}

	return nil
}
