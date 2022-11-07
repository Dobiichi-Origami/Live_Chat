package controllers

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"liveChat/constants"
	"liveChat/containers"
	"liveChat/db"
	"liveChat/entities"
	"liveChat/log"
	"liveChat/tools"
	"sync"
)

var watchMap *containers.ThreadSafeContainer

func init() {
	watchMap = containers.NewThreadSafeContainer()
}

func Subscribe(chatId []int64, userId int64) []int64 {
	failList := make([]int64, 0)
	for _, id := range chatId {
		ret := watchMap.Get(id)
		if ret == nil {
			watch, err := db.SubscribeChatSeq(id)
			if err != nil {
				log.Error(fmt.Sprintf("注册会话监听失败: %s. 会话 id: %d", err.Error(), id))
				failList = append(failList, id)
				continue
			}

			watchList := newWatchUserList(id, watch)
			ptr, ok := watchMap.LoadOrStore(id, watchList)
			if ptr == nil {
				log.Error(fmt.Sprintf("会话监听存储失败. 会话 id: %d", id))
				failList = append(failList, id)
				continue
			} else if !ok {
				go watchList.maintainer()
			}
			ret = ptr
		}
		ret.(*watchUserList).add(userId)
	}

	return failList
}

func DeleteSubscribe(chatId, userId int64) {
	if list := watchMap.Get(chatId).(*watchUserList); list != nil {
		list.delete(userId)
	}
}

type watchUserList struct {
	chatId  int64
	stream  *mongo.ChangeStream
	userSet map[int64]struct{}
	rwlock  sync.RWMutex
}

func newWatchUserList(chatId int64, s *mongo.ChangeStream) *watchUserList {
	return &watchUserList{
		chatId:  chatId,
		stream:  s,
		userSet: make(map[int64]struct{}),
		rwlock:  sync.RWMutex{},
	}
}

func (w *watchUserList) add(userId int64) {
	w.rwlock.Lock()
	w.userSet[userId] = struct{}{}
	w.rwlock.Unlock()
}

func (w *watchUserList) delete(userId int64) {
	w.rwlock.Lock()
	delete(w.userSet, userId)
	w.rwlock.Unlock()
}

func (w *watchUserList) size() int {
	return len(w.userSet)
}

func (w *watchUserList) maintainer() {
	for {
		ret, err := tools.FetchDataFromChangeStreamBsonWithoutTimeOut(w.stream)
		if err == nil {
			msg := entities.NewMessageFromChangeStreamBson(tools.FetchBsonFromChangeStreamData(ret))
			protoMarshal, err := proto.Marshal(entities.TransferMessageToProtoBuf(msg))
			if err != nil {
				log.Error("反序列化错误")
				continue
			}

			buf := tools.GenerateResponseBytes(constants.MessageLoad, tools.MagicNumberBinary, protoMarshal)
			w.rwlock.RLock()
			for k, _ := range w.userSet {
				if c := GetConnection(int(k)); !c.IsClosed() {
					c.conn.AsyncWrite(buf, nil)
				}
			}
			w.rwlock.RUnlock()
		} else if err = w.errorHandler(); err != nil {
			// TODO 补充错误信息
			log.Error("尝试恢复会话失败")
			return
		}
	}
}

func (w *watchUserList) errorHandler() (err error) {
	for i := 0; i < 3; i++ {
		watch, tmp := db.SubscribeChatMessage(w.chatId)
		if tmp == nil {
			w.stream = watch
			return nil
		}
		err = tmp
	}
	return err
}
