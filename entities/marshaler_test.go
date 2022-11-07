package entities

import (
	"github.com/golang/protobuf/proto"
	"liveChat/protocol"
	"testing"
)

var protobufMessage = protocol.Message{
	Id:        1,
	Sender:    1,
	Receiver:  1,
	Timestamp: 1,
	Type:      1,
	Contents:  []string{"sdadbusiadguisagdiugsaiudas", "dsaduisagidugasiudiua"},
}
var protobufBinary, _ = proto.Marshal(&protobufMessage)

var jsonMessage = NewMessage(1, 1, 1, 1, 1, "sdadbusiadguisagdiugsaiudasdsaduisagidugasiudiua")

var jsonBinary, _ = jsonMessage.MarshalJSON()

func BenchmarkSerializeMessageWithProtobuf(b *testing.B) {
	proto.Marshal(&protobufMessage)
}

func BenchmarkDeserializeMessageWithProtobuf(b *testing.B) {
	proto.Unmarshal(protobufBinary, &protocol.Message{})
}

func BenchmarkSerializeMessageWithJson(b *testing.B) {
	jsonMessage.MarshalJSON()
}

func BenchmarkDeserializeMessageWithJson(b *testing.B) {
	NewEmptyMessage().UnmarshalJSON(jsonBinary)
}
