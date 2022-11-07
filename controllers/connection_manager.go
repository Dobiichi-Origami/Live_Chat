package controllers

import (
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/atomic"
	"liveChat/containers"
)

var connectionMap *containers.ThreadSafeContainer

func init() {
	connectionMap = containers.NewThreadSafeContainer()
}

func AddConnection(c gnet.Conn, cc chan struct{}) {
	connectionMap.Set(int64(c.Fd()), &Connection{token: "", conn: c, closeChan: cc, closeLock: atomic.NewBool(false)})
}

func SetTokenForConnection(token string, fd int) {
	connectionMap.Get(int64(fd)).(*Connection).token = token
}

func DeleteConnection(c gnet.Conn) {
	connectionMap.Delete(int64(c.Fd()))
}

func GetConnection(fd int) *Connection {
	if ret := connectionMap.Get(int64(fd)).(*Connection); ret != nil && !ret.closeLock.Load() {
		return ret
	}
	return nil
}

type Connection struct {
	token     string
	conn      gnet.Conn
	closeChan chan struct{}
	closeLock *atomic.Bool
}

func (c *Connection) GetToken() string {
	return c.token
}

func (c *Connection) CloseConnection() {
	c.closeChan <- struct{}{}
}

func (c *Connection) CloseLock() bool {
	return c.closeLock.CAS(false, true)
}

func (c *Connection) IsClosed() bool {
	return c.closeLock.Load()
}
