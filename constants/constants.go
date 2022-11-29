package constants

import (
	"math"
	"time"
)

const MaxUInt64 = math.MaxUint64

const MagicNumber = uint16(10086)
const (
	ErrorResponseLoad byte = iota
	SuccessResponseLoad

	MessageLoad
	RequestMessageLoad

	MultiMessageLoad
	RequestMultiMessageLoad
	RequestEstablishConnectionLoad
	ResponseEstablishConnectionLoad
	NotificationRequestLoad

	HeartBeatLoad
	HeartBeatResponse
)

const HeartBeatMaxInterval = 180
const KeepAlivePeriod = time.Second * 15
