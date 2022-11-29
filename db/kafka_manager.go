package db

import (
	"context"
	"github.com/Shopify/sarama"
	"liveChat/log"
	"strconv"
	"time"
)

type ConsumeFunc func(message *sarama.ConsumerMessage)

type MQConsumerGroup interface {
	StartConsume()
}

type kafkaGroupConsumer struct {
	fn                 ConsumeFunc
	group              sarama.ConsumerGroup
	topics             []string
	asyncHandleChan    []chan *sarama.ConsumerMessage
	asyncHandlerNumber int64
}

func NewKafkaGroupConsumer(addresses, topics []string, groupIds string, asyncHandlerNumber int, fn ConsumeFunc) (MQConsumerGroup, error) {
	cfg := &sarama.Config{}
	cfg.Version = sarama.V3_2_3_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Consumer.Return.Errors = false

	group, err := sarama.NewConsumerGroup(addresses, groupIds, cfg)
	if err != nil {
		return nil, err
	}

	asyncHandlerChan := make([]chan *sarama.ConsumerMessage, asyncHandlerNumber, asyncHandlerNumber)
	for i := 0; i < asyncHandlerNumber; i++ {
		index := i
		asyncHandlerChan[index] = make(chan *sarama.ConsumerMessage, 5000)
		go func() {
			for {
				msg := <-asyncHandlerChan[index]
				fn(msg)
			}
		}()
	}

	return kafkaGroupConsumer{
		fn:                 fn,
		topics:             topics,
		group:              group,
		asyncHandleChan:    asyncHandlerChan,
		asyncHandlerNumber: int64(asyncHandlerNumber),
	}, nil
}

func (kgc kafkaGroupConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (kgc kafkaGroupConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (kgc kafkaGroupConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		kgc.asyncHandleChan[msg.Offset%kgc.asyncHandlerNumber] <- msg
		session.MarkMessage(msg, "")
	}
	return nil
}

func (kgc kafkaGroupConsumer) StartConsume() {
	go func() {
		for {
			err := kgc.group.Consume(context.Background(), kgc.topics, kgc)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}()
}

type KafkaAsyncProducer struct {
	producer sarama.AsyncProducer
	topic    string
}

func NewKafkaAsyncProducer(addresses []string) (*KafkaAsyncProducer, error) {
	cfg := sarama.Config{}
	cfg.Producer.Flush.Frequency = time.Second
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Flush.Messages = 1000
	cfg.Producer.Idempotent = true
	cfg.Producer.Partitioner = sarama.NewHashPartitioner
	cfg.Producer.Return.Successes = true

	p, err := sarama.NewAsyncProducer(addresses, nil)
	if err != nil {
		return nil, err
	}

	return &KafkaAsyncProducer{
		producer: p,
	}, nil
}

func (ap *KafkaAsyncProducer) AsyncSendMessage(userId int64, bytes []byte) {
	key := strconv.FormatInt(userId, 10)
	ap.producer.Input() <- &sarama.ProducerMessage{Topic: ap.topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(bytes)}
	<-ap.producer.Successes()
}
