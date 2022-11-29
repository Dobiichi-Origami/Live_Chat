package controllers

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"liveChat/db"
	"liveChat/log"
	"liveChat/rpc"
	"strconv"
	"strings"
	"sync"
	"time"
)

type connectionEntry struct {
	id   int64
	conn grpc.ClientConnInterface
}

var (
	rpcConnections []connectionEntry
	opRpcLock      sync.RWMutex
)

func newRpcConnection(id int64, conn *grpc.ClientConn) connectionEntry {
	return connectionEntry{id, conn}
}

func InitServerInterconnection(etcdUrls []string, serverHost string) {
	db.InitEtcd(etcdUrls)
	db.RegisterService(serverHost)
	keys, values, err := db.GetAllKV(db.EtcdNodePrefix)
	if err != nil {
		panic(err)
	}

	rpcConnections = make([]connectionEntry, 0)
	opRpcLock = sync.RWMutex{}

	go registerServerWithRetry(keys, values)
	db.RegisterWatch(db.EtcdNodePrefix, watchHook)
}

func GetAllServerClients() (ret []rpc.ServerNodeClient) {
	opRpcLock.RLock()
	for _, entry := range rpcConnections {
		ret = append(ret, rpc.NewServerNodeClient(entry.conn))
	}
	opRpcLock.RUnlock()
	return
}

func watchHook(response clientv3.WatchResponse) {
	for _, event := range response.Events {
		switch event.Type {
		case clientv3.EventTypePut:
			go registerServerWithRetry([]string{string(event.Kv.Key)}, []string{string(event.Kv.Value)})
		case clientv3.EventTypeDelete:
			deleteConnectionFromList(stripNodeIdFromKey(string(event.Kv.Key)))
		default:
			continue
		}
	}
}

func stripNodeIdFromKey(key string) int64 {
	id, _ := strconv.Atoi(strings.Split(key, "/")[1])
	return int64(id)
}

func registerServerWithRetry(keys, values []string) {
	for len(keys) != 0 {
		failedKeys, failedValues := generateServerList(keys, values)
		keys, values = failedKeys, failedValues
		time.Sleep(time.Second)
	}
}

func generateServerList(keys, values []string) (failedKeys, failedValues []string) {
	for i := 0; i < len(keys); i++ {
		nodeId := stripNodeIdFromKey(keys[i])
		host := values[i]
		conn, err := grpc.Dial(host, grpc.WithInsecure())
		if err != nil {
			log.Error(err.Error())
			failedKeys = append(failedKeys, keys[i])
			failedValues = append(failedValues, values[i])
		} else {
			appendConnectionToList(nodeId, conn)
		}
	}
	return
}

func appendConnectionToList(id int64, conn *grpc.ClientConn) {
	opRpcLock.Lock()
	rpcConnections = append(rpcConnections, newRpcConnection(id, conn))
	opRpcLock.Unlock()
}

func deleteConnectionFromList(id int64) {
	opRpcLock.Lock()
	defer opRpcLock.Unlock()
	for i := 0; i < len(rpcConnections); i++ {
		if rpcConnections[i].id == id {
			rpcConnections[i].conn.(*grpc.ClientConn).Close()
			rpcConnections = append(rpcConnections[:i], rpcConnections[i+1:]...)
			return
		}
	}
}
