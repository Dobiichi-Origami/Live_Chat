package http

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"liveChat/controllers"
	"liveChat/db"
	"liveChat/entities"
	"liveChat/log"
	"liveChat/rpc"
	"time"
)

var (
	notificationAsyncProducer *db.KafkaAsyncProducer
	notificationConsumerGroup db.MQConsumerGroup
)

func InitNotificationQueue(urls, topics []string, groupsId string) {
	var err error
	notificationAsyncProducer, err = db.NewKafkaAsyncProducer(urls)
	if err != nil {
		panic(err)
	}

	notificationConsumerGroup, err = db.NewKafkaGroupConsumer(urls, topics, groupsId, 12, consumeNotificationFunc)
	if err != nil {
		panic(err)
	}
	notificationConsumerGroup.StartConsume()
}

func SendNotification(notification *entities.Notification) {
	protoNot := rpc.NotificationRequest{
		RequestId:       0,
		Id:              notification.Seq,
		Sender:          notification.SenderId,
		Receiver:        notification.ReceiverId,
		Timestamp:       uint64(notification.Timestamp),
		Op:              rpc.NotificationRequest_OpType(notification.OpType),
		ReceiveType:     rpc.NotificationRequest_ReceiveType(notification.ReceiveType),
		IsHandledByAuth: notification.IsHandled,
		IsAgree:         notification.IsAgree,
	}

	data, _ := proto.Marshal(&protoNot)
	notificationAsyncProducer.AsyncSendMessage(notification.ReceiverId, data)
}

func consumeNotificationFunc(m *sarama.ConsumerMessage) {
	noti := &rpc.NotificationRequest{}
	err := proto.Unmarshal(m.Value, noti)
	if err != nil {
		log.Error(fmt.Sprintf("反序列化通知 proto 错误: %s", err.Error()))
		return
	}

	clients := controllers.GetAllServerClients()
	for _, client := range clients {
		ctx, cfn := context.WithTimeout(context.Background(), time.Second*3)
		_, err := client.BroadcastNotification(ctx, noti, nil)
		cfn()
		if err != nil {
			log.Error(err.Error())
			continue
		}
	}
}
