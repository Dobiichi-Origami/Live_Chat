package db

import (
	"context"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"go.etcd.io/etcd/client/v3"
	"liveChat/log"
	"strconv"
	"time"
)

const EtcdNodePrefix = "node_prefix/"

const (
	etcdDefaultNodeId    = -1
	etcdDefaultLeaseTime = 9
	etcdDefaultOutTime   = 3
)

var client *clientv3.Client

func InitEtcd(urls []string) {
	var err error
	client, err = clientv3.New(clientv3.Config{
		Endpoints:            urls,
		AutoSyncInterval:     0,
		DialTimeout:          time.Second * 5,
		DialKeepAliveTime:    time.Second * 10,
		DialKeepAliveTimeout: time.Second * 30,
	})
	if err != nil {
		panic(err)
	}
}

func RegisterService(listenHost string) {
	go func() {
		lease := clientv3.NewLease(client)
		kv := clientv3.NewKV(client)
		curLeaseId := clientv3.LeaseID(etcdDefaultNodeId)

		for {
			if curLeaseId == etcdDefaultNodeId {
				grant, err := lease.Grant(context.Background(), etcdDefaultLeaseTime)
				for err != nil {
					log.Error(err.Error())
					time.Sleep(time.Second)
					grant, err = lease.Grant(context.Background(), etcdDefaultLeaseTime)
				}

				curLeaseId = grant.ID
				key := EtcdNodePrefix + strconv.FormatInt(int64(curLeaseId), 10)

				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*etcdDefaultOutTime)
				_, err = kv.Put(ctx, key, listenHost, clientv3.WithLease(curLeaseId))
				for err != nil {
					cancelFunc()
					log.Error(err.Error())
					time.Sleep(time.Second)
					ctx, cancelFunc = context.WithTimeout(context.Background(), time.Second*etcdDefaultOutTime)
					_, err = kv.Put(ctx, key, listenHost, clientv3.WithLease(curLeaseId))
				}
				cancelFunc()
			} else if _, err := lease.KeepAliveOnce(context.TODO(), curLeaseId); err == rpctypes.ErrLeaseNotFound {
				log.Error(err.Error())
				curLeaseId = etcdDefaultNodeId
			}
			time.Sleep(time.Second)
		}
	}()
}

func RegisterWatch(key string, hook func(response clientv3.WatchResponse)) {
	watcher := clientv3.NewWatcher(client)
	go func() {
		for {
			respChan := watcher.Watch(context.Background(), key, clientv3.WithPrefix())
			for resp := range respChan {
				if resp.Canceled {
					log.Error(resp.Err().Error())
					break
				}
				hook(resp)
			}
			time.Sleep(time.Second)
		}
	}()
}

func GetAllKV(key string) (keys, values []string, err error) {
	kv := clientv3.NewKV(client)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*etcdDefaultOutTime)
	resps, err := kv.Get(ctx, key, clientv3.WithPrefix())
	defer cancelFunc()
	if err != nil {
		return nil, nil, err
	}

	for _, ikv := range resps.Kvs {
		keys = append(keys, string(ikv.Key))
		values = append(values, string(ikv.Value))
	}
	return
}
