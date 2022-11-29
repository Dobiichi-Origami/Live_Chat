package pool

import (
	"encoding/binary"
	"errors"
	"github.com/panjf2000/gnet/v2"
	"liveChat/constants"
	"liveChat/tools"
	"sync"
)

// 负载格式：
// 首 2 个字节为一个魔数，为大端存储的 10086
// 其次接一个 1 字节的版本号以标识消息版本
// 后接 4 字节 uint32 大端序存储的消息序号
// 后接 1 字节的消息体类型
//  0 为 rpc.ErrorResponse,
//  1 为 SuccessResponse, 无对应 Proto，即 Load 中 content 为空
//	2 为 rpc.Message,
//	3 为 rpc.RequestMessage,
//	4 为 rpc.MultiMessage,
//	5 为 rpc.RequestMultiMessage,
//  6 为 rpc.RequestEstablishConnection,
//  7 为 rpc.ResponseEstablishConnection,
//  8 为 rpc.NotificationRequest,
//  9 为 HeartBeatLoad, 无对应 Proto
//  10 为 HeartBeatResponse, 无对应 Proto
// 后再接 4 字节 uint32 大端序存储的消息长度
// 随后是经过 protobuf 序列化后的 Message 字节流

type TCPRequestPackage struct {
	Version     byte
	Ack         []byte
	RequestType byte
	Load        []byte
	Conn        gnet.Conn
}

func (p *TCPRequestPackage) SetPackageUsingConn(c gnet.Conn) error {
	bytesSlice := Get12BytesSlice()
	defer Put12BytesSlice(bytesSlice)

	if errString, status := tools.ReadAndHandleError(c, bytesSlice); status != gnet.None {
		return errors.New(errString)
	}

	if binary.BigEndian.Uint16(bytesSlice[0:2]) != constants.MagicNumber {
		return errors.New("魔数不匹配")
	}

	if bytesSlice[2] != 0 {
		return errors.New("版本号不存在")
	}

	requestType := bytesSlice[7]
	length := tools.BytesToUint32(bytesSlice[8:])
	contentSlice := make([]byte, length, length)

	if errString, status := tools.ReadAndHandleError(c, contentSlice); status != gnet.None {
		return errors.New(errString)
	}

	p.Version = bytesSlice[2]
	p.Ack = make([]byte, 4, 4)
	copy(p.Ack, bytesSlice[3:7])
	p.RequestType = requestType
	p.Load = contentSlice
	p.Conn = c
	return nil
}

var packagePool sync.Pool

func init() {
	binary.BigEndian.PutUint16(tools.MagicNumberBinary, constants.MagicNumber)
	packagePool = sync.Pool{New: func() interface{} {
		return TCPRequestPackage{}
	}}
}

func GetRequestPackage() TCPRequestPackage {
	return packagePool.Get().(TCPRequestPackage)
}

func PutRequestPackage(ptr interface{}) {
	packagePool.Put(ptr)
}
