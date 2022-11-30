package main

import (
	"flag"
	"fmt"
	"liveChat/config"
	"liveChat/controllers"
	"liveChat/db"
	"liveChat/http"
	"liveChat/rpc/rpc_implementation"
	"liveChat/tcp"
	"time"
)

const (
	configPathArgName = "path"
)

func main() {
	path := flag.String(configPathArgName, "", "启动配置文件路径")
	flag.Parse()

	generalConfig := config.NewGeneralConfig(*path)

	go http.InitHttpServer(generalConfig.HttpListenAddresses)
	time.Sleep(time.Second)
	http.InitNotificationQueue(generalConfig.NotificationQueueConfig.Urls, generalConfig.NotificationQueueConfig.Topics, generalConfig.NotificationQueueConfig.GroupsId)

	go tcp.InitiateTcpServer(generalConfig.TcpListenAddress)
	tcp.InitMessageQueue(generalConfig.MessageQueueConfig.Urls, generalConfig.MessageQueueConfig.Topics, generalConfig.MessageQueueConfig.GroupsId)

	db.InitRedisConnection(generalConfig.RedisConfig.Format())
	db.InitMongoDBConnection(generalConfig.MongoDBConfig.Format(), generalConfig.MongoDBConfig.Database)
	db.InitMysqlConnection(generalConfig.MysqlConfig.Format())

	fmt.Printf("准备初始化 RPC\n")
	go rpc_implementation.InitRpcServer(generalConfig.GrpcListenAddress)

	fmt.Printf("准备初始化 RPC链接\n")
	controllers.InitServerInterconnection(generalConfig.EtcdUrls, generalConfig.GrpcServeAddress)

	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case t := <-ticker.C:
			fmt.Printf("服务器正在运行：%v\n", t)
		}
	}
}
