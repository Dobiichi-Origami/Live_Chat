package tcp

const (
	Web = iota
	Android
)

type TCPContext struct {
	UserId   int64
	Token    string
	Platform int
}
