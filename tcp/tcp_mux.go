package tcp

import (
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"liveChat/constants"
	"liveChat/controllers"
	"liveChat/log"
	"liveChat/pool"
	"liveChat/tools"
)

var engine *engineImplementation

type engineImplementation struct {
	gnet.BuiltinEventEngine
}

func (engine *engineImplementation) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if err := c.SetKeepAlivePeriod(constants.KeepAlivePeriod); err != nil {
		log.Error(err.Error())
		return nil, gnet.Close
	}

	c.SetContext(pool.GetTCPContext())
	return nil, gnet.None
}

func (engine *engineImplementation) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	ctx := c.Context().(pool.TCPContext)
	controllers.DeleteConnection(ctx.UserId, ctx.Platform)
	pool.PutTCPContext(ctx)

	if err != nil {
		log.Error(fmt.Sprintf("关闭参数中错误不为空: %s", err.Error()))
	}
	return gnet.Close
}

func (engine *engineImplementation) OnTraffic(c gnet.Conn) (action gnet.Action) {
	request := pool.GetRequestPackage()
	if err := request.SetPackageUsingConn(c); err != nil {
		pool.PutRequestPackage(request)
		errInfo := fmt.Sprintf("收到客户端非法请求: %s", err.Error())
		log.Error(errInfo)
		err = c.AsyncWrite(tools.GenerateErrorResponseBytes(errInfo), nil)
		if err != nil {
			log.Error(fmt.Sprintf("错误处理中异步发送回包失败: %s", err.Error()))
		}
	} else {
		PushTask(request)
	}
	return gnet.None
}

func InitiateTcpServer(address string) {
	engine = &engineImplementation{}
	if err := gnet.Run(engine, address, gnet.WithMulticore(true)); err != nil {
		panic(err)
	}

	return
}
