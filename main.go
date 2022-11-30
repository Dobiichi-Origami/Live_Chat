package liveChat

import (
	"flag"
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
	generalConfig := config.NewGeneralConfig(*path)

	http.InitHttpServer(generalConfig.HttpListenAddresses)
	http.InitNotificationQueue(generalConfig.NotificationQueueConfig.Urls, generalConfig.NotificationQueueConfig.Topics, generalConfig.NotificationQueueConfig.GroupsId)

	tcp.InitiateTcpServer(generalConfig.TcpListenAddress)
	tcp.InitMessageQueue(generalConfig.MessageQueueConfig.Urls, generalConfig.MessageQueueConfig.Topics, generalConfig.MessageQueueConfig.GroupsId)

	db.InitRedisConnection(generalConfig.RedisConfig.Format())
	db.InitMongoDBConnection(generalConfig.MongoDBConfig.Format())
	db.InitMysqlConnection(generalConfig.MysqlConfig.Format())

	rpc_implementation.InitRpcServer(generalConfig.GrpcListenAddress)
	controllers.InitServerInterconnection(generalConfig.EtcdUrls, generalConfig.GrpcServeAddress)

	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ticker.C:
			// TODO SOMETHING
		}
	}
}
