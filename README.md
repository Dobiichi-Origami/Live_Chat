# Live Chat
![GitHub](https://img.shields.io/github/license/Dobiichi-Origami/Live_Chat) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Dobiichi-Origami/Live_Chat)

#Live Chat：高可用、高性能即时通讯套件
+ 高度可水平拓展
+ 单机高性能
+ 分布式架构
+ 轻量级
+ **完全免费！！！**

# 部署方式
**方式2 与 方式3 需要自行部署中间件**

## 1. Docker-Compose（体验用）
```shell
git clone https://github.com/Dobiichi-Origami/Live_Chat.git
cd Live_Chat
mkdir -p "running/mysql" "running/mongodb"
docker-compose up
```

## 2. Docker
1. 复制仓库中文件 `default_config_files/config.json` 到本地目录中
2. 修改以下列表属性值为可使用的配置
    + `message_queue_config.Urls`: 消息队列中间件地址
    + `notification_queue_config.Urls`: 消息队列中间件地址
    + `grpc_serve_address`: 本机对外可供外部访问的地址（例如域名或集群 IP）
    + `mysql_config`: Mysql 配置
    + `mongo_db_config`: Mongodb 配置
    + `redis_config`: Redis 配置
3. docker run -p 1234:1234 -p 1345:1345 -p 5678:5678 -v path/to/config/folder:/appdata/config

## 3. 源码编译
需要 Golang 版本 1.17.2 及以上

1. 首先编译
```shell
git clone https://github.com/Dobiichi-Origami/Live_Chat.git
cd Live_Chat
go mod tidy
go build -o liveChat liveChat/main
```

2. 同方式2 修改 `default_config_files/config.json`
3. 执行 `./liveChat --path default_config_files/config.json`



