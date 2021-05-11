package chattingroom

type Message struct {
	data []byte
	room string
}

func NewMessage(room string, data []byte) *Message {
	return &Message{
		room: room,
		data: data,
	}
}
