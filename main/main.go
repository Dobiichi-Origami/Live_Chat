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
	"os"
	"time"
)

const (
	configPathArgName = "path"

	mongoAddressENV         = "MONGO_ADDRESS"
	mysqlAddressENV         = "MYSQL_ADDRESS"
	redisAddressENV         = "REDIS_ADDRESS"
	NotificationQueueUrlENV = "NOTIFICATION_QUEUE_URL"
	MessageQueueUrlENV      = "MESSAGE_QUEUE_URL"
	EtcdAddressENV          = "ETCD_ADDRESS"
)

var (
	mongoAddressVar         string
	mysqlAddressVar         string
	redisAddressVar         string
	notificationQueueUrlVar string
	messageQueueUrlVar      string
	etcdAddressVar          string
)

func main() {
	path := flag.String(configPathArgName, "", "启动配置文件路径")
	flag.Parse()
	parseENV()
	generalConfig := config.NewGeneralConfig(*path)

	go http.InitHttpServer(generalConfig.HttpListenAddresses)
	go tcp.InitiateTcpServer(generalConfig.TcpListenAddress)
	go rpc_implementation.InitRpcServer(generalConfig.GrpcListenAddress)

	initMessageQueue(generalConfig)
	initNotificationQueue(generalConfig)
	initMysql(generalConfig)
	initRedis(generalConfig)
	initMongoDb(generalConfig)
	initEtcd(generalConfig)

	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case t := <-ticker.C:
			fmt.Printf("服务器正在运行：%v\n", t)
		}
	}
}

func parseENV() {
	mongoAddressVar = os.Getenv(mongoAddressENV)
	mysqlAddressVar = os.Getenv(mysqlAddressENV)
	redisAddressVar = os.Getenv(redisAddressENV)
	notificationQueueUrlVar = os.Getenv(NotificationQueueUrlENV)
	messageQueueUrlVar = os.Getenv(MessageQueueUrlENV)
	etcdAddressVar = os.Getenv(EtcdAddressENV)
}

func initNotificationQueue(cfg *config.GeneralConfig) {
	urlSlice := make([]string, 0)
	if notificationQueueUrlVar != "" {
		urlSlice = []string{notificationQueueUrlVar}
	} else {
		urlSlice = cfg.NotificationQueueConfig.Urls
	}
	http.InitNotificationQueue(urlSlice, cfg.NotificationQueueConfig.Topics, cfg.NotificationQueueConfig.GroupsId)
}

func initMessageQueue(cfg *config.GeneralConfig) {
	urlSlice := make([]string, 0)
	if messageQueueUrlVar != "" {
		urlSlice = []string{messageQueueUrlVar}
	} else {
		urlSlice = cfg.MessageQueueConfig.Urls
	}
	tcp.InitMessageQueue(urlSlice, cfg.MessageQueueConfig.Topics, cfg.MessageQueueConfig.GroupsId)
}

func initRedis(cfg *config.GeneralConfig) {
	if redisAddressVar != "" {
		db.InitRedisConnection(redisAddressVar)
	} else {
		db.InitRedisConnection(cfg.RedisConfig.Format())
	}
}

func initMongoDb(cfg *config.GeneralConfig) {
	url := ""
	if mongoAddressVar != "" {
		url = mongoAddressVar
	} else {
		url = cfg.MongoDBConfig.Format()
	}
	db.InitMongoDBConnection(url, cfg.MongoDBConfig.Database)
}

func initMysql(cfg *config.GeneralConfig) {
	url := ""
	if mysqlAddressVar != "" {
		url = mysqlAddressVar
	} else {
		url = cfg.MysqlConfig.Format()
	}
	db.InitMysqlConnection(url)
}

func initEtcd(cfg *config.GeneralConfig) {
	var url []string
	if etcdAddressVar != "" {
		url = []string{etcdAddressVar}
	} else {
		url = cfg.EtcdUrls
	}
	controllers.InitServerInterconnection(url, cfg.GrpcServeAddress)
}
