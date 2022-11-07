package entities

import (
	"go.mongodb.org/mongo-driver/bson"
	"liveChat/protocol"
	"strings"
)

type ContentType uint8

const (
	Text ContentType = iota
	Image
	Emoji
)

const protobufStringLengthLimit = 232

type Message struct {
	Id        uint64      `bson:"id"`
	Sender    int64       `bson:"sender"`
	Receiver  int64       `bson:"receiver"`
	Timestamp uint64      `bson:"timestamp"`
	Type      ContentType `bson:"type"`

	Content string `bson:"content"`
}

func NewMessage(id uint64, sender, receiver int64, timestamp uint64, contentType ContentType, content string) *Message {
	return &Message{
		Id:        id,
		Sender:    sender,
		Receiver:  receiver,
		Timestamp: timestamp,
		Type:      contentType,
		Content:   content,
	}
}

func NewMessageFromProtobufWithoutSeq(m *protocol.Message) *Message {
	return NewMessage(
		0,
		m.GetSender(),
		m.GetReceiver(),
		m.GetTimestamp(),
		ContentType(m.GetType()),
		buildStringFromProtobuf(m.Contents),
	)
}

func NewMessageFromProtobufWithSeq(m *protocol.Message) *Message {
	return NewMessage(
		m.Id,
		m.GetSender(),
		m.GetReceiver(),
		m.GetTimestamp(),
		ContentType(m.GetType()),
		buildStringFromProtobuf(m.Contents),
	)
}

func NewEmptyMessage() *Message {
	return &Message{}
}

func NewMessageFromChangeStreamBson(data bson.D) *Message {
	return &Message{
		Id:        uint64(data[1].Value.(int64)),
		Sender:    data[2].Value.(int64),
		Receiver:  data[3].Value.(int64),
		Timestamp: uint64(data[4].Value.(int64)),
		Type:      ContentType(data[5].Value.(int32)),
		Content:   data[6].Value.(string),
	}
}

func TransferMessageToProtoBuf(m *Message) *protocol.Message {
	message := protocol.Message{
		Id:        m.Id,
		Sender:    m.Sender,
		Receiver:  m.Receiver,
		Timestamp: m.Timestamp,
		Type:      protocol.MessageContentType(m.Type),
		Contents:  nil,
	}

	for i := 0; i < len(m.Content); i += protobufStringLengthLimit {
		ceil := i + protobufStringLengthLimit
		if ceil >= len(m.Content) {
			message.Contents = append(message.Contents, m.Content[i:])
		} else {
			message.Contents = append(message.Contents, m.Content[i:i+protobufStringLengthLimit])
		}
	}

	return &message
}

func buildStringFromProtobuf(slice []string) string {
	length := 0
	for i := 0; i < len(slice); i++ {
		length += len(slice)
	}

	builder := strings.Builder{}
	builder.Grow(length)
	for i := 0; i < len(slice); i++ {
		builder.WriteString(slice[i])
	}

	return builder.String()
}
