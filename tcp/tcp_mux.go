package tcp

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/gnet/v2"
	"liveChat/constants"
	"liveChat/controllers"
	"liveChat/log"
	"liveChat/pool"
)

var upgrade = ws.Upgrader{
	OnHost: func(host []byte) error {
		// TODO 加入 CORS 判断
		return nil
	},
}

var engine *engineImplementation

type engineImplementation struct {
	gnet.BuiltinEventEngine
}

func (engine *engineImplementation) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if err := c.SetKeepAlivePeriod(constants.KeepAlivePeriod); err != nil {
		log.Error(err.Error())
		return nil, gnet.Close
	}

	return nil, gnet.None
}

func (engine *engineImplementation) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if c.Context() != nil {
		ctx := c.Context().(pool.TCPContext)
		controllers.DeleteConnection(ctx.UserId, ctx.Platform)
		pool.PutTCPContext(ctx)
	}

	if err != nil {
		log.Error(fmt.Sprintf("关闭参数中错误不为空: %s", err.Error()))
	}
	return gnet.Close
}

func (engine *engineImplementation) OnTraffic(c gnet.Conn) (action gnet.Action) {
	if c.Context() == nil {
		_, err := upgrade.Upgrade(c)
		if err != nil {
			return gnet.Close
		}

		c.SetContext(pool.GetTCPContext())
		return gnet.None
	}

	messages := make([]wsutil.Message, 0)
	messages, err := wsutil.ReadClientMessage(c, messages)
	if err != nil {
		log.Error(err.Error())
		return gnet.Close
	}

	for _, m := range messages {
		switch m.OpCode {
		case ws.OpPing:
			wsutil.WriteServerMessage(c, ws.OpPong, nil)
		case ws.OpClose:
			c.Close()
		case ws.OpBinary:
			request := pool.GetRequestPackage()
			if err = request.SetPackageUsingPayload(m.Payload, c); err != nil {
				pool.PutRequestPackage(request)
				log.Error(fmt.Sprintf("收到客户端非法请求: %s", err.Error()))
				return gnet.Close
			} else {
				PushTask(request)
			}
		}
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
