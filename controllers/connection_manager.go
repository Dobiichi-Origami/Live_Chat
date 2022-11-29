package controllers

import (
	"github.com/panjf2000/gnet/v2"
	"liveChat/containers"
	"sync"
)

var connectionMap *containers.ThreadSafeContainer

func init() {
	connectionMap = containers.NewThreadSafeContainer()
}

func AddConnection(c gnet.Conn, userId int64, platform int) bool {
	ret, _ := connectionMap.LoadOrStore(userId, newConnectionForUser())
	conns := ret.(*connectionForUser)
	return conns.addConnection(c, platform)
}

func DeleteConnection(userId int64, platform int) bool {
	conn, ok := connectionMap.Get(userId)
	if !ok || conn == nil {
		return false
	}
	flag, ok := conn.(*connectionForUser).deleteConnection(platform)
	if flag {
		connectionMap.Delete(userId)
	}

	return ok
}

func GetConnection(userId int64) []gnet.Conn {
	if ret, ok := connectionMap.Get(userId); ok && ret != nil {
		return ret.(*connectionForUser).getConnections()
	}
	return nil
}

type connectionForUser struct {
	rwLock   *sync.RWMutex
	conns    []gnet.Conn
	size     int
	isClosed bool
}

func newConnectionForUser() *connectionForUser {
	return &connectionForUser{
		rwLock:   &sync.RWMutex{},
		conns:    make([]gnet.Conn, 2, 2),
		size:     0,
		isClosed: false,
	}
}

func (c *connectionForUser) addConnection(conn gnet.Conn, platform int) bool {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	if c.isClosed {
		return false
	}

	if c.conns[platform] != nil {
		c.conns[platform].Close()
	} else {
		c.size++
	}
	c.conns[platform] = conn
	return true
}

func (c *connectionForUser) deleteConnection(platform int) (bool, bool) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	if c.isClosed || c.conns[platform] == nil {
		return false, false
	}

	c.conns[platform].Close()
	c.conns[platform] = nil
	c.size--

	if c.size == 0 {
		c.isClosed = true
	}

	return c.isClosed, true
}

func (c *connectionForUser) getConnections() []gnet.Conn {
	c.rwLock.RLock()
	defer c.rwLock.RUnlock()

	if c.isClosed {
		return nil
	}

	ret := make([]gnet.Conn, 0)
	for _, conn := range c.conns {
		if conn == nil {
			continue
		}
		ret = append(ret, conn)
	}
	return ret
}
