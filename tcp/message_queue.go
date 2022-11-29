package tcp

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"liveChat/controllers"
	"liveChat/db"
	"liveChat/log"
	"liveChat/rpc"
	"time"
)

var (
	messageAsyncProducer *db.KafkaAsyncProducer
	messageConsumerGroup db.MQConsumerGroup
)

func InitMessageQueue(urls, topics []string, groupsId string) {
	var err error
	messageAsyncProducer, err = db.NewKafkaAsyncProducer(urls)
	if err != nil {
		panic(err)
	}

	messageConsumerGroup, err = db.NewKafkaGroupConsumer(urls, topics, groupsId, 12, consumeMessageFunc)
	messageConsumerGroup.StartConsume()
}

func SendMessage(message *rpc.Message) {
	data, err := proto.Marshal(message)
	if err != nil {
		log.Error(err.Error())
		return
	}

	messageAsyncProducer.AsyncSendMessage(message.Receiver, data)
}

func consumeMessageFunc(m *sarama.ConsumerMessage) {
	message := rpc.Message{}
	if err := proto.Unmarshal(m.Value, &message); err != nil {
		log.Error(fmt.Sprintf("反序列化消息 proto 错误: %s", err.Error()))
		return
	}

	messageRequest := rpc.MessageRequest{
		RequestId: 0,
		Message:   &message,
	}

	clients := controllers.GetAllServerClients()
	for _, client := range clients {
		ctx, cfn := context.WithTimeout(context.Background(), time.Second*3)
		_, err := client.BroadcastMessage(ctx, &messageRequest, nil)
		cfn()
		if err != nil {
			log.Error(err.Error())
			continue
		}
	}
}
